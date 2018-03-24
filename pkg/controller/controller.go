package controller

import (
	"fmt"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/fukt/dweller/pkg/apis/dweller/v1alpha1"
	"github.com/fukt/dweller/pkg/client/clientset/versioned"
	"github.com/fukt/dweller/pkg/client/informers/externalversions"
	"github.com/fukt/dweller/pkg/log"
	"github.com/fukt/dweller/pkg/secret"
)

const maxRetries = 5

// Controller is a main dweller controller structure.
type Controller struct {
	logger    log.Logger
	clientset kubernetes.Interface
	queue     workqueue.RateLimitingInterface
	informer  cache.SharedIndexInformer

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
	defaultResync := 0 * time.Millisecond

	clientset, err := versioned.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating clientset: %v", err)
	}

	informerFactory := externalversions.NewSharedInformerFactory(clientset, defaultResync)
	informer := informerFactory.Dweller().V1alpha1().VaultSecretClaims().Informer()
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	result := &Controller{
		logger:    &log.Dummy{},
		clientset: client,
		informer:  informer,
		queue:     queue,
		asm:       asm,
	}

	for _, option := range options {
		option(result)
	}

	result.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				result.queue.Add(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				result.queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				result.queue.Add(key)
			}
		},
	})

	return result, nil
}

// Run starts the kubewatch controller
func (c *Controller) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()

	c.logger.Infof("starting dweller controller")

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		c.informer.Run(stopCh)
		c.queue.ShutDown()
	}()

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for caches to populate"))
		return
	}

	c.logger.Infof("dweller controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
	wg.Wait()
}

// HasSynced is required for the cache.Controller interface.
func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
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

	err := c.processItem(key.(string))
	if err == nil {
		// No error, reset the ratelimit counters
		c.queue.Forget(key)
	} else if c.queue.NumRequeues(key) < maxRetries {
		c.logger.Errorf("error processing %s (will retry): %v", key, err)
		c.queue.AddRateLimited(key)
	} else {
		// err != nil and too many retries
		c.logger.Errorf("error processing %s (giving up): %v", key, err)
		c.queue.Forget(key)
		utilruntime.HandleError(err)
	}

	return true
}

// processItem is a main processing function
func (c *Controller) processItem(key string) error {
	c.logger.Debugf("process VaultSecretClaim %s", key)

	item, _, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("fetch object by key %s from store: %v", key, err)
	}

	if item == nil {
		// The secret will be automatically garbage collected, so no need
		// to delete it manually.
		return nil
	}

	vsc, ok := item.(*v1alpha1.VaultSecretClaim)
	if !ok {
		return fmt.Errorf("item is not of type VaultSecretClaim but %T", item)
	}

	namespace := "default"
	if vsc.Namespace != "" {
		namespace = vsc.Namespace
	}

	// We need to fetch existing secretList and check if any is already owned by
	// the vault secret claim. If no such secret exists, we need to create one.
	// Otherwise, we need to update existing secret.

	labelSelector := labels.SelectorFromSet(vsc.Spec.Secret.Metadata.GetLabels())
	secretList, err := c.clientset.CoreV1().
		Secrets(namespace).
		List(metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		})
	if err != nil {
		return err
	}

	var existingSecret *corev1.Secret
	for _, sec := range secretList.Items {
		if metav1.IsControlledBy(&sec, vsc) {
			existingSecret = &sec
			break
		}
	}

	if existingSecret != nil {
		return c.updateSecret(vsc, *existingSecret)
	}

	return c.createSecret(vsc, namespace)
}

func (c *Controller) createSecret(vsc *v1alpha1.VaultSecretClaim, namespace string) error {
	sec, err := c.asm.Assemble(vsc)
	if err != nil {
		return err
	}

	_, err = c.clientset.CoreV1().Secrets(namespace).Create(&sec)
	if err != nil {
		return fmt.Errorf("create kubernetes secret: %v", err)
	}

	return nil
}

func (c *Controller) updateSecret(vsc *v1alpha1.VaultSecretClaim, secret corev1.Secret) error {
	newSecret, err := c.asm.Assemble(vsc)
	if err != nil {
		return err
	}

	// In meta, we need to update only labels and annotations.
	secret.ObjectMeta.Labels = newSecret.Labels
	secret.ObjectMeta.Annotations = newSecret.Annotations

	// In data, we update it as a whole.
	secret.Data = nil
	secret.StringData = newSecret.StringData

	_, err = c.clientset.CoreV1().Secrets(secret.Namespace).Update(&secret)
	if err != nil {
		return fmt.Errorf("update kubernetes secret: %v", err)
	}

	return nil
}
