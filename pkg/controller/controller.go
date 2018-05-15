package controller

import (
	"flag"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const DefaultResyncPeriod = 0
const MaxRetries = 5

var controllerBuilders = make(map[string]ControllerBuilder)
var controllStopCh = make(map[string]chan struct{})

type ControllerBuilder func(client kubernetes.Interface) cache.Controller

func RegisterController(name string, fn ControllerBuilder) {
	controllerBuilders[name] = fn
}

func Execute() {
	client, err := initKubeClient()
	if nil != err {
		panic(err.Error())
	}
	for name, builder := range controllerBuilders {
		fmt.Println("starting init controller: ", name)
		c := builder(client)
		stopChC := make(chan struct{})
		go c.Run(stopChC)
		controllStopCh[name] = stopChC
		fmt.Printf("controller %s started!\n", name)
	}
}

func Stop() {
	for name, stopChC := range controllStopCh {
		close(stopChC)
		fmt.Printf("controller %s stopped!\n", name)
	}
}

func initKubeClient() (kubernetes.Interface, error) {
	var kubeconfig *string
	var err error
	var config *rest.Config
	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()

	if *kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			return nil, err
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
