package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/fukt/dweller/pkg/apis/dweller"
)

// Specification represents application environment variables configuration.
type Specification struct {
	KubernetesConfig string `required:"false" desc:"Absolute path to the kubeconfig file"`
}

func main() {
	var s Specification
	err := envconfig.Process("", &s)
	if err != nil {
		panic(err.Error())
	}

	config := mustConfig(s.KubernetesConfig)
	kubeClient := mustInitKubernetesClient(config)

	c, err := dweller.New(config, kubeClient, dweller.WithLogger(logrus.New()))
	if err != nil {
		panic(err.Error())
	}

	stopCh := make(chan struct{})
	defer close(stopCh)
	go c.Run(stopCh)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)
	signal.Notify(signals, syscall.SIGINT)
	<-signals
	close(signals)
}

func mustConfig(kubeconfig string) *rest.Config {
	if kubeconfig == "" {
		// If there is no out-cluster config, create in-cluster
		return mustInClusterConfig()
	}

	// Use provided out-cluster config otherwise
	return mustOutClusterConfig(kubeconfig)
}

func mustOutClusterConfig(kubeconfig string) *rest.Config {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic("error creating out-cluster k8s client: " + err.Error())
	}
	return config
}

func mustInClusterConfig() *rest.Config {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic("error creating in-cluster k8s client: " + err.Error())
	}
	return config
}

func mustInitKubernetesClient(config *rest.Config) *kubernetes.Clientset {
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic("error creating k8s client: " + err.Error())
	}
	return kubeClient
}
