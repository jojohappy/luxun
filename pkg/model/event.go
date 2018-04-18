package model

import (
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
	PodCondition      PodCondition            `json:"podCondition"`
	ContainerStatus   map[int]ContainerStatus `json:"containerStatus"`
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

	for i, s := range po.Status.ContainerStatuses {
		cs := ContainerStatus{
			Name: s.Name,
		}
		if nil != s.State.Waiting {
			cs.State = ContainerStatusWaiting
			cs.Reason = s.State.Waiting.Reason
			cs.Message = s.State.Waiting.Message
		} else if nil != s.State.Running {
			cs.State = ContainerStatusRunning
		} else if nil != s.State.Terminated {
			cs.State = ContainerStatusTerminated
			cs.ExitCode = s.State.Terminated.ExitCode
			cs.Signal = s.State.Terminated.Signal
			cs.Reason = s.State.Terminated.Reason
			cs.Message = s.State.Terminated.Message
		}
		ev.ContainerStatus[i] = cs
	}

	return ev
}

func ConvertPodDeleteEvent(po *core_v1.Pod) *Event {
	ev := ConvertPodBasicEvent(po)
	ev.Action = "Pod"
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
