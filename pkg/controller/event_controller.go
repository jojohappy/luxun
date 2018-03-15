package controller

import (
	"fmt"
	"time"

	events_v1beta1 "k8s.io/api/events/v1beta1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type EventController struct {
	informer cache.SharedIndexInformer
	queue    workqueue.RateLimitingInterface
	client   kubernetes.Interface
}

func init() {
	RegisterController("events", New)
}

func New(client kubernetes.Interface) cache.Controller {
	return newEventController(client)
}

func newEventController(client kubernetes.Interface) *EventController {
	f := informers.NewSharedInformerFactory(client, DefaultResyncPeriod)
	ec := &EventController{
		informer: f.Events().V1beta1().Events().Informer(),
		queue:    workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		client:   client,
	}

	ec.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{ec.OnAdd, ec.OnUpdate, ec.OnDelete})
	return ec
}

func (ec *EventController) Run(stopCh <-chan struct{}) {
	defer ec.queue.ShutDown()

	fmt.Println("start event controller")
	go ec.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, ec.HasSynced) {
		fmt.Println("timed out waiting for caches to sync")
		return
	}

	fmt.Println("event controller synced and ready")

	wait.Until(ec.worker, time.Second, stopCh)
}

func (ec *EventController) HasSynced() bool {
	return ec.informer.HasSynced()
}

func (ec *EventController) LastSyncResourceVersion() string {
	return ec.informer.LastSyncResourceVersion()
}

func (ec *EventController) worker() {
	for ec.nextWork() {
	}
}

func (ec *EventController) nextWork() bool {
	key, quit := ec.queue.Get()
	if quit {
		fmt.Println("unexpected quit of queue")
		return false
	}
	defer ec.queue.Done(key)
	fmt.Printf("handle event by key: %s\n", key)
	err := ec.processItem(key.(string))
	if err == nil {
		ec.queue.Forget(key)
	} else if ec.queue.NumRequeues(key) < MaxRetries {
		fmt.Printf("error processing %s (will retry): %v\n", key, err)
		ec.queue.AddRateLimited(key)
	} else {
		fmt.Printf("error processing %s (giving up): %v\n", key, err)
		ec.queue.Forget(key)
	}
	return true
}

func (ec *EventController) processItem(key string) error {
	obj, _, err := ec.informer.GetIndexer().GetByKey(key)
	if nil != err {
		return fmt.Errorf("error fetching object with key %s from store: %v", key, err)
	}
	ev, ok := obj.(*events_v1beta1.Event)
	if ok {
		fmt.Printf("event is %v\n", ev)
	}
	return nil
}

// OnAdd calls AddFunc if it's not nil.
func (ec *EventController) OnAdd(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	fmt.Printf("Processing add to event: %s\n", key)
	if err == nil {
		ec.queue.Add(key)
	}
}

// OnUpdate calls UpdateFunc if it's not nil.
func (ec *EventController) OnUpdate(oldObj, newObj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(newObj)
	fmt.Printf("Processing update to event: %s\n", key)
	if err == nil {
		ec.queue.Add(key)
	}
}

// OnDelete calls DeleteFunc if it's not nil.
func (ec *EventController) OnDelete(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	fmt.Printf("Processing delete to event: %s\n", key)
	if err == nil {
		ec.queue.Add(key)
	}
}
