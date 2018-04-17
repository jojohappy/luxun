package controller

import (
	"fmt"
	"time"

	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type PodController struct {
	informer cache.SharedIndexInformer
	queue    workqueue.RateLimitingInterface
	client   kubernetes.Interface
}

func init() {
	RegisterController("pods", NewPodController)
}

func NewPodController(client kubernetes.Interface) cache.Controller {
	return newPodController(client)
}

func newPodController(client kubernetes.Interface) *PodController {
	f := informers.NewSharedInformerFactory(client, DefaultResyncPeriod)
	pc := &PodController{
		informer: f.Core().V1().Pods().Informer(),
		queue:    workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		client:   client,
	}

	pc.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{pc.OnAdd, pc.OnUpdate, pc.OnDelete})
	return pc
}

func (pc *PodController) Run(stopCh <-chan struct{}) {
	defer pc.queue.ShutDown()

	fmt.Println("start event controller")
	go pc.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, pc.HasSynced) {
		fmt.Println("timed out waiting for caches to sync")
		return
	}

	fmt.Println("event controller synced and ready")

	wait.Until(pc.worker, time.Second, stopCh)
}

func (pc *PodController) HasSynced() bool {
	return pc.informer.HasSynced()
}

func (pc *PodController) LastSyncResourceVersion() string {
	return pc.informer.LastSyncResourceVersion()
}

func (pc *PodController) worker() {
	for pc.nextWork() {
	}
}

func (pc *PodController) nextWork() bool {
	key, quit := pc.queue.Get()
	if quit {
		fmt.Println("unexpected quit of queue")
		return false
	}
	defer pc.queue.Done(key)
	err := pc.processItem(key.(string))
	if err == nil {
		pc.queue.Forget(key)
	} else if pc.queue.NumRequeues(key) < MaxRetries {
		fmt.Printf("error processing %s (will retry): %v\n", key, err)
		pc.queue.AddRateLimited(key)
	} else {
		fmt.Printf("error processing %s (giving up): %v\n", key, err)
		pc.queue.Forget(key)
	}
	return true
}

func (pc *PodController) processItem(key string) error {
	obj, _, err := pc.informer.GetIndexer().GetByKey(key)
	if nil != err {
		return fmt.Errorf("error fetching object with key %s from store: %v", key, err)
	}
	p, ok := obj.(*core_v1.Pod)
	if !ok {
		fmt.Println("failed to convert object to core.v1.Pod ")
		return nil
	}
	fmt.Printf("%#v\n", p)
	return nil
}

func (pc *PodController) OnAdd(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err == nil {
		pc.queue.Add(key)
	}
}

func (pc *PodController) OnUpdate(oldObj, newObj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(newObj)
	if err == nil {
		pc.queue.Add(key)
	}
}

func (pc *PodController) OnDelete(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err == nil {
		pc.queue.Add(key)
	}
}
