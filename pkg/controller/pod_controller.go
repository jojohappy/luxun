package controller

import (
	"fmt"

	"github.com/luxun/pkg/model"
	"github.com/luxun/pkg/stream"

	core_v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type PodController struct {
	informer cache.SharedIndexInformer
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
		client:   client,
	}

	pc.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{pc.OnAdd, pc.OnUpdate, pc.OnDelete})
	return pc
}

func (pc *PodController) Run(stopCh <-chan struct{}) {
	fmt.Println("start pod controller")
	go pc.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, pc.HasSynced) {
		fmt.Println("timed out waiting for caches to sync")
		return
	}

	fmt.Println("pod controller synced and ready")

	<-stopCh
}

func (pc *PodController) HasSynced() bool {
	return pc.informer.HasSynced()
}

func (pc *PodController) LastSyncResourceVersion() string {
	return pc.informer.LastSyncResourceVersion()
}

func (pc *PodController) OnAdd(obj interface{}) {}

func (pc *PodController) OnUpdate(_, newObj interface{}) {
	pod, err := convertToPod(newObj)
	if err != nil {
		fmt.Println("converting to Pod object failed in OnUpdate", "err", err)
		return
	}
	stream.Process(model.ConvertPodEvent(pod))
}

func (pc *PodController) OnDelete(obj interface{}) {
	pod, err := convertToPod(obj)
	if err != nil {
		fmt.Println("converting to Pod object failed in OnDelete", "err", err)
		return
	}
	stream.Process(model.ConvertPodDeleteEvent(pod))
}

func convertToPod(o interface{}) (*core_v1.Pod, error) {
	pod, ok := o.(*core_v1.Pod)
	if ok {
		return pod, nil
	}

	deletedState, ok := o.(cache.DeletedFinalStateUnknown)
	if !ok {
		return nil, fmt.Errorf("Received unexpected object: %v", o)
	}
	pod, ok = deletedState.Obj.(*core_v1.Pod)
	if !ok {
		return nil, fmt.Errorf("DeletedFinalStateUnknown contained non-Pod object: %v", deletedState.Obj)
	}
	return pod, nil
}
