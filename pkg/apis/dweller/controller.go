package dweller

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	dweller_v1a1 "github.com/fukt/dweller/pkg/apis/dweller/v1alpha1"
)

const maxRetries = 5

type Controller struct {
	logger    *logrus.Entry
	clientset kubernetes.Interface
	queue     workqueue.RateLimitingInterface
	informer  cache.SharedIndexInformer
}

func Start() {
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("error creating k8s client: %s", err.Error())
		return
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("error creating k8s client: %s", err.Error())
		return
	}

	c := newController(config, kubeClient)
	stopCh := make(chan struct{})
	defer close(stopCh)

	go c.Run(stopCh)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)
	<-sigterm
}

func newController(k8sConfig *rest.Config, client kubernetes.Interface) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	config := *k8sConfig
	gv := dweller_v1a1.SchemeGroupVersion
	config.APIPath = "/apis"
	config.GroupVersion = &gv
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: dweller_v1a1.CodecFactory}
	c, err := rest.RESTClientFor(&config)
	if nil != err {
		fmt.Printf("error creating client: %v", err)
		return nil
	}

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (result runtime.Object, err error) {
				result = &dweller_v1a1.SecretClaimList{}
				err = c.Get().
					Namespace(meta_v1.NamespaceAll).
					Resource("secretclaims").
					VersionedParams(&options, dweller_v1a1.ParameterCodec).
					Do().
					Into(result)
				return
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				options.Watch = true
				watcher, err := c.Get().
					Namespace(meta_v1.NamespaceAll).
					Resource("secretclaims").
					VersionedParams(&options, dweller_v1a1.ParameterCodec).
					Watch()
				return watcher, err
			},
		},
		&dweller_v1a1.SecretClaim{},
		0, //Skip resync
		cache.Indexers{},
	)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	})

	return &Controller{
		logger:    logrus.WithField("pkg", "dweller"),
		clientset: client,
		informer:  informer,
		queue:     queue,
	}
}

// Run starts the kubewatch controller
func (c *Controller) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("Starting dweller controller")

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	c.logger.Info("Dweller controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced is required for the cache.Controller interface.
func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

// LastSyncResourceVersion is required for the cache.Controller interface.
func (c *Controller) LastSyncResourceVersion() string {
	return c.informer.LastSyncResourceVersion()
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
		c.logger.Errorf("Error processing %s (will retry): %v", key, err)
		c.queue.AddRateLimited(key)
	} else {
		// err != nil and too many retries
		c.logger.Errorf("Error processing %s (giving up): %v", key, err)
		c.queue.Forget(key)
		utilruntime.HandleError(err)
	}

	return true
}

// processItem is a main processing function
func (c *Controller) processItem(key string) error {
	c.logger.Infof("Processing change to SecretClaim %s", key)

	_, _, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("Error fetching object with key %s from store: %v", key, err)
	}
	return nil
}
