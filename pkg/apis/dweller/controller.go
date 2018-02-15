package dweller

import (
	"fmt"
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/fukt/dweller/pkg/client/clientset/versioned"
	"github.com/fukt/dweller/pkg/client/informers/externalversions"
)

const maxRetries = 5

// Controller is a main dweller controller structure.
type Controller struct {
	logger    Logger
	clientset kubernetes.Interface
	queue     workqueue.RateLimitingInterface
	informer  cache.SharedIndexInformer
}

// ControllerOption is a function option for dweller controller.
type ControllerOption func(*Controller)

// Logger sets specified logger as a default one.
func WithLogger(lg Logger) ControllerOption {
	return func(c *Controller) {
		c.logger = lg
	}
}

// New returns newly created dweller controller or nil on error.
func New(k8sConfig *rest.Config, client kubernetes.Interface, options ...ControllerOption) (*Controller, error) {
	defaultResync := 0 * time.Millisecond

	clientset, err := versioned.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating clientset: %v", err)
	}

	informerFactory := externalversions.NewSharedInformerFactory(clientset, defaultResync)
	informer := informerFactory.Dweller().V1alpha1().VaultSecretClaims().Informer()
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	result := &Controller{
		logger:    &dummyLogger{},
		clientset: client,
		informer:  informer,
		queue:     queue,
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
	defer c.queue.ShutDown()

	c.logger.Infof("starting dweller controller")

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for caches to populate"))
		return
	}

	c.logger.Infof("dweller controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
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
	c.logger.Infof("processing change to VaultSecretClaim %s", key)

	_, _, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("error fetching object with key %s from store: %v", key, err)
	}
	return nil
}
