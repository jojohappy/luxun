package model

import (
	"fmt"
	"time"

	core_v1 "k8s.io/api/core/v1"
)

type KVObject struct {
	Key   string
	Value string
}

type Event struct {
	Time              time.Time               `json:"time"`
	Name              string                  `json:"name,omitempty"`
	Namespace         string                  `json:"namespace,omitempty"`
	CreationTimestamp time.Time               `json:"creationTimestamp,omitempty"`
	Labels            map[int]KVObject        `json:"labels,omitempty"`
	Annotations       map[int]KVObject        `json:"annotations,omitempty"`
	Kind              string                  `json:"kind,omitempty"`
	Reason            string                  `json:"reason,omitempty"`
	Message           string                  `json:"message,omitempty"`
	FirstTimestamp    time.Time               `json:"firstTimestamp,omitempty"`
	LastTimestamp     time.Time               `json:"lastTimestamp,omitempty"`
	Count             int32                   `json:"count,omitempty"`
	Type              string                  `json:"type,omitempty"`
	Action            string                  `json:"action,omitempty"`
	EventTime         time.Time               `json:"eventTime"`
	Env               string                  `json:"env"`
	PodCondition      PodCondition            `json:"podCondition,omitempty"`
	ContainerStatus   map[int]ContainerStatus `json:"containerStatus,omitempty"`
	PodStatus         string                  `json:"podStatus,omitempty`
}

func ConvertEvent(ev *core_v1.Event) *Event {
	return &Event{
		Time:              time.Now(),
		Name:              ev.ObjectMeta.Name,
		Namespace:         ev.ObjectMeta.Namespace,
		CreationTimestamp: ev.ObjectMeta.CreationTimestamp.Time,
		Kind:              ev.InvolvedObject.Kind,
		Reason:            ev.Reason,
		Message:           ev.Message,
		FirstTimestamp:    ev.FirstTimestamp.Time,
		LastTimestamp:     ev.LastTimestamp.Time,
		Count:             ev.Count,
		Type:              ev.Type,
		Action:            ev.Action,
		EventTime:         ev.EventTime.Time,
		Env:               GetEnv(),
	}
}

func ConvertPodEvent(po *core_v1.Pod) *Event {
	ev := ConvertPodBasicEvent(po)
	ev.ContainerStatus = make(map[int]ContainerStatus)

	for _, condition := range po.Status.Conditions {
		if condition.Type == core_v1.PodReady {
			ev.PodCondition = PodCondition{
				Status:  string(condition.Status),
				Type:    string(condition.Type),
				Reason:  condition.Reason,
				Message: condition.Message,
			}
			break
		}
	}

	reason := string(po.Status.Phase)
	if po.Status.Reason != "" {
		reason = po.Status.Reason
	}

	initializing := false
	for i := range po.Status.InitContainerStatuses {
		container := po.Status.InitContainerStatuses[i]
		switch {
		case container.State.Terminated != nil && container.State.Terminated.ExitCode == 0:
			continue
		case container.State.Terminated != nil:
			// initialization is failed
			if len(container.State.Terminated.Reason) == 0 {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Init:Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("Init:ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else {
				reason = "Init:" + container.State.Terminated.Reason
			}
			initializing = true
		case container.State.Waiting != nil && len(container.State.Waiting.Reason) > 0 && container.State.Waiting.Reason != "PodInitializing":
			reason = "Init:" + container.State.Waiting.Reason
			initializing = true
		default:
			reason = fmt.Sprintf("Init:%d/%d", i, len(po.Spec.InitContainers))
			initializing = true
		}
		break
	}
	if !initializing {
		for i := len(po.Status.ContainerStatuses) - 1; i >= 0; i-- {
			container := po.Status.ContainerStatuses[i]
			cs := ContainerStatus{
				Name: container.Name,
			}
			if nil != container.State.Waiting {
				cs.State = ContainerStatusWaiting
				cs.Reason = container.State.Waiting.Reason
				cs.Message = container.State.Waiting.Message
			} else if nil != container.State.Running {
				cs.State = ContainerStatusRunning
			} else if nil != container.State.Terminated {
				cs.State = ContainerStatusTerminated
				cs.ExitCode = container.State.Terminated.ExitCode
				cs.Signal = container.State.Terminated.Signal
				cs.Reason = container.State.Terminated.Reason
				cs.Message = container.State.Terminated.Message
			}
			ev.ContainerStatus[i] = cs
			if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
				reason = container.State.Waiting.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason != "" {
				reason = container.State.Terminated.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason == "" {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("ExitCode:%d", container.State.Terminated.ExitCode)
				}
			}
		}
	}

	if po.DeletionTimestamp != nil && po.Status.Reason == "NodeLost" {
		reason = "Unknown"
	} else if po.DeletionTimestamp != nil {
		reason = "Terminating"
	}

	ev.PodStatus = reason
	return ev
}

func ConvertPodDeleteEvent(po *core_v1.Pod) *Event {
	ev := ConvertPodBasicEvent(po)
	ev.Action = "Delete"
	return ev
}

func ConvertPodBasicEvent(po *core_v1.Pod) *Event {
	ev := &Event{
		Time:              time.Now(),
		Name:              po.ObjectMeta.Name,
		Namespace:         po.ObjectMeta.Namespace,
		CreationTimestamp: po.ObjectMeta.CreationTimestamp.Time,
		Labels:            make(map[int]KVObject),
		Annotations:       make(map[int]KVObject),
		Kind:              "Pod",
		Env:               GetEnv(),
		ContainerStatus:   make(map[int]ContainerStatus),
	}

	i := 0
	for k, v := range po.ObjectMeta.Labels {
		ev.Labels[i] = KVObject{k, v}
		i++
	}

	i = 0
	for k, v := range po.ObjectMeta.Annotations {
		ev.Annotations[i] = KVObject{k, v}
		i++
	}
	return ev
}
