package controller

import (
	"fmt"
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/fukt/dweller/pkg/apis/dweller/v1alpha1"
	"github.com/fukt/dweller/pkg/client/clientset/versioned"
	"github.com/fukt/dweller/pkg/client/informers/externalversions"
	dwellerlisters "github.com/fukt/dweller/pkg/client/listers/dweller/v1alpha1"
	"github.com/fukt/dweller/pkg/log"
	"github.com/fukt/dweller/pkg/secret"
)

const (
	maxRetries = 5

	// builtinResync is resync period for built-it kubernetes objects, in our
	// case secrets.
	builtinResync = time.Minute * 5

	// customResync is resync period for custom resource definitions introduced
	// by the controller, in our case vault secret claims.
	customResync = time.Minute * 1
)

// Controller is a main dweller controller structure.
type Controller struct {
	builtinFactory informers.SharedInformerFactory
	customFactory  externalversions.SharedInformerFactory

	client    kubernetes.Interface
	clientset versioned.Interface

	logger log.Logger

	queue workqueue.RateLimitingInterface

	// saLister can list/get service accounts the shared informer's store.
	saLister corelisters.ServiceAccountLister

	// secretLister can list/get secrets from the shared informer's store.
	secretLister corelisters.SecretLister

	// vscLister can list/get vault secret claims from the shared informer's store.
	vscLister dwellerlisters.VaultSecretClaimLister

	// asm is a secret assembler that is used to create kubernetes secrets
	// based on vault secret claim.
	asm secret.Assembler
}

// Option is a function option for dweller controller.
type Option func(*Controller)

// WithLogger sets specified logger as a default one.
func WithLogger(lg log.Logger) Option {
	return func(c *Controller) {
		c.logger = lg
	}
}

// New returns newly created dweller controller or nil on error.
func New(k8sConfig *rest.Config, client kubernetes.Interface, asm secret.Assembler, options ...Option) (*Controller, error) {
	clientset, err := versioned.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating client: %v", err)
	}

	ctrl := &Controller{
		logger:    &log.Dummy{},
		client:    client,
		clientset: clientset,
		asm:       asm,
	}

	for _, option := range options {
		option(ctrl)
	}

	ctrl.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	builtinFactory := informers.NewSharedInformerFactory(client, builtinResync)
	customFactory := externalversions.NewSharedInformerFactory(clientset, customResync)

	ctrl.builtinFactory = builtinFactory
	ctrl.customFactory = customFactory

	ctrl.saLister = builtinFactory.Core().V1().ServiceAccounts().Lister()
	ctrl.secretLister = builtinFactory.Core().V1().Secrets().Lister()
	ctrl.vscLister = customFactory.Dweller().V1alpha1().VaultSecretClaims().Lister()

	vscInformer := customFactory.Dweller().V1alpha1().VaultSecretClaims().Informer()
	vscInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ctrl.addVaultSecretClaim,
		UpdateFunc: ctrl.updateVaultSecretClaim,
		DeleteFunc: ctrl.deleteVaultSecretClaim,
	})

	// TODO: we should also watch for secrets to:
	// 1) enforce vsc definitions;
	// 2) watch if conflicting secrets were deleted and we can safely create new.

	// Hacky stuff.
	utilruntime.ErrorHandlers[0] = func(err error) {
		ctrl.logger.Errorf(err.Error())
	}

	return ctrl, nil
}

// Run starts the controller.
func (c *Controller) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	go func() {
		<-stopCh
		c.queue.ShutDown()
	}()

	c.logger.Infof("Starting dweller controller")
	defer c.logger.Infof("Shutting down dweller controller")

	c.builtinFactory.Start(stopCh)
	c.customFactory.Start(stopCh)

	if err := c.syncInformersCache(stopCh); err != nil {
		utilruntime.HandleError(err)
		return
	}

	c.logger.Infof("Dweller controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
}

func (c *Controller) syncInformersCache(stopCh <-chan struct{}) error {
	var syncedInformers map[reflect.Type]bool
	var errs []error

	// Sync all built-in informers that we are using in the controller.
	syncedInformers = c.builtinFactory.WaitForCacheSync(stopCh)
	for inf, synced := range syncedInformers {
		if !synced {
			errs = append(errs, fmt.Errorf("Couldn't sync cache for %v", inf))
		} else {
			c.logger.Debugf("Synced cache for %s", inf)
		}
	}

	// Sync custom informers.
	syncedInformers = c.customFactory.WaitForCacheSync(stopCh)
	for inf, synced := range syncedInformers {
		if !synced {
			errs = append(errs, fmt.Errorf("couldn't sync cache for %v", inf))
		} else {
			c.logger.Debugf("Synced cache for %s", inf)
		}
	}

	return utilerrors.NewAggregate(errs)
}

func (c *Controller) addVaultSecretClaim(obj interface{}) {
	vsc := obj.(*v1alpha1.VaultSecretClaim)
	c.logger.Infof("Adding VaultSecretClaim \"%s/%s\"", vsc.Namespace, vsc.Name)
	c.enqueue(vsc)
}

func (c *Controller) updateVaultSecretClaim(old, new interface{}) {
	oldVsc := old.(*v1alpha1.VaultSecretClaim)
	newVsc := new.(*v1alpha1.VaultSecretClaim)
	c.logger.Infof("Updating VaultSecretClaim \"%s/%s\"", oldVsc.Namespace, oldVsc.Name)
	c.enqueue(newVsc)
}

func (c *Controller) deleteVaultSecretClaim(obj interface{}) {
	vsc, ok := obj.(*v1alpha1.VaultSecretClaim)
	if ok {
		c.logger.Infof("Deleting VaultSecretClaim \"%s/%s\"", vsc.Namespace, vsc.Name)
		c.enqueue(vsc)
		return
	}

	tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
		return
	}
	vsc, ok = tombstone.Obj.(*v1alpha1.VaultSecretClaim)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("Tombstone contained object that is not a VaultSecretClaim %#v", obj))
		return
	}
}

// enqueue adds vault secret claim to the queue.
func (c *Controller) enqueue(vsc *v1alpha1.VaultSecretClaim) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(vsc)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", vsc, err))
		return
	}

	c.queue.Add(key)
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *Controller) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.syncVaultSecretClaim(key.(string))
	c.handleProcessingError(err, key)

	return true
}

func (c *Controller) handleProcessingError(err error, key interface{}) {
	if err == nil {
		// No error, reset the ratelimit counters
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < maxRetries {
		utilruntime.HandleError(fmt.Errorf("Error processing %s (will retry): %v", key, err))
		c.queue.AddRateLimited(key)
		return
	}

	// too many retries
	utilruntime.HandleError(fmt.Errorf("Error processing %s (giving up): %v", key, err))
	c.queue.Forget(key)
}

// syncVaultSecretClaim will sync the vault secret claim with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (c *Controller) syncVaultSecretClaim(key string) error {
	startTime := time.Now()
	c.logger.Infof("Started syncing VaultSecretClaim %q (%v)", key, startTime.Format(time.RFC3339Nano))
	defer func() {
		c.logger.Infof("Finished syncing VaultSecretClaim %q (%v)", key, time.Since(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	c.logger.Debugf("Looking for VaultSecretClaim \"%s/%s\"", namespace, name)

	vaultSecretClaim, err := c.vscLister.VaultSecretClaims(namespace).Get(name)
	if apierrors.IsNotFound(err) {
		// VaultSecretClaim's child secret will be automatically garbage
		// collected, so no need to delete it manually.
		c.logger.Infof("VaultSecretClaim %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	// Deep-copy otherwise we are mutating our cache.
	vsc := vaultSecretClaim.DeepCopy()

	// TODO: check if spec.serviceAccountName exists.
	// TODO: check if spec.vaultRole exists.

	relatedServiceAccount, err := c.saLister.ServiceAccounts(vsc.Namespace).Get(vsc.Spec.ServiceAccountName)
	if err != nil {
		return err
	}
	if len(relatedServiceAccount.Secrets) > 1 {
		return fmt.Errorf("service account \"%s/%s\" has no secrets associated with it", relatedServiceAccount.Namespace, relatedServiceAccount.Name)
	}

	secretName := relatedServiceAccount.Secrets[0].Name
	saSecret, err := c.secretLister.Secrets(vsc.Namespace).Get(secretName)
	if err != nil {
		return err
	}

	token := string(saSecret.Data["token"])
	if token == "" {
		return fmt.Errorf("token not found in service account secret \"%s/%s\"", saSecret.Namespace, saSecret.Name)
	}

	creds := &secret.Credentials{
		Token: token,
		Role:  vsc.Spec.VaultRole,
	}

	c.logger.Debugf("Looking for Secret \"%s/%s\"", vsc.Namespace, vsc.Name)
	relatedSecret, err := c.secretLister.Secrets(vsc.Namespace).Get(vsc.Name)
	if apierrors.IsNotFound(err) {
		c.logger.Debugf("Secret \"%s/%s\" was not found - VaultSecretClaim will create one", vsc.Namespace, vsc.Name)
		if err := c.createSecret(vsc, creds); err != nil {
			return err
		}
		c.logger.Infof("Secret for VaultSecretClaim %v has been created", key)
		return nil
	}

	if err != nil {
		return err
	}

	// Deep-copy otherwise we are mutating our cache.
	sec := relatedSecret.DeepCopy()

	if !metav1.IsControlledBy(sec, vsc) {
		err := fmt.Errorf("Conflict: found secret \"%s/%s\" that is not owned by vault secret claim. This must be resolved manually.", sec.Namespace, sec.Name)
		utilruntime.HandleError(err)
		return nil // Don't need to retry.
	}

	// TODO: It might be worthwhile to revisit creation/updating
	// logic to handle all the fields properly.

	// Found a secret, and it is controlled by vault secret claim.
	// Just sync it.
	if err := c.updateSecret(vsc, sec, creds); err != nil {
		return err
	}
	c.logger.Infof("Secret for VaultSecretClaim %v has been updated", key)
	return nil
}

func (c *Controller) createSecret(vsc *v1alpha1.VaultSecretClaim, creds *secret.Credentials) error {
	sec, err := c.asm.Assemble(vsc, creds)
	if err != nil {
		return err
	}

	_, err = c.client.CoreV1().Secrets(vsc.Namespace).Create(&sec)
	if err != nil {
		return fmt.Errorf("create kubernetes secret: %v", err)
	}

	return nil
}

func (c *Controller) updateSecret(vsc *v1alpha1.VaultSecretClaim, secret *corev1.Secret, creds *secret.Credentials) error {
	// TODO: compute and compare hashes to not to do worthless updates if vault
	// secret claim is not actually changed (for example in case of resync).
	// Implement something like "pod-template-hash" in ReplicaSet.

	newSecret, err := c.asm.Assemble(vsc, creds)
	if err != nil {
		return err
	}

	// In meta, we need to update only labels and annotations.
	secret.ObjectMeta.Labels = newSecret.Labels
	secret.ObjectMeta.Annotations = newSecret.Annotations

	// In data, we update it as a whole.
	secret.Data = nil
	secret.StringData = newSecret.StringData

	_, err = c.client.CoreV1().Secrets(secret.Namespace).Update(secret)
	if err != nil {
		return fmt.Errorf("update kubernetes secret: %v", err)
	}

	return nil
}
