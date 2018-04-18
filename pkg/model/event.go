package model

import (
	"os"
	"time"

	core_v1 "k8s.io/api/core/v1"
)

const (
	ContainerStatusWaiting    = "Waiting"
	ContainerStatusTerminated = "Terminated"
	ContainerStatusRunning    = "Running"
)

type Event struct {
	Time              time.Time                  `json:"time"`
	Name              string                     `json:"name,omitempty"`
	Namespace         string                     `json:"namespace,omitempty"`
	CreationTimestamp time.Time                  `json:"creationTimestamp,omitempty"`
	Labels            map[string]string          `json:"labels,omitempty"`
	Annotations       map[string]string          `json:"annotations,omitempty"`
	Kind              string                     `json:"kind,omitempty"`
	Reason            string                     `json:"reason,omitempty"`
	Message           string                     `json:"message,omitempty"`
	FirstTimestamp    time.Time                  `json:"firstTimestamp,omitempty"`
	LastTimestamp     time.Time                  `json:"lastTimestamp,omitempty"`
	Count             int32                      `json:"count,omitempty"`
	Type              string                     `json:"type,omitempty"`
	Action            string                     `json:"action,omitempty"`
	EventTime         time.Time                  `json:"eventTime"`
	Env               string                     `json:"env"`
	PodCondition      PodCondition               `json:"podCondition"`
	ContainerStatus   map[string]ContainerStatus `json:"containerStatus"`
}

type PodCondition struct {
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
	Type    string `json:"type,omitempty"`
	Status  string `json:"status,omitempty"`
}

type ContainerStatus struct {
	Name     string `json:"name,omitempty"`
	State    string `json:"state,omitempty"`
	ExitCode int32  `json:"exitCode,omitempty"`
	Signal   int32  `json:"signal,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Message  string `json:"message,omitempty"`
}

func ConvertEvent(ev *core_v1.Event) *Event {
	env := os.Getenv("KUBEENV")
	if "" == env {
		env = "unknown"
	}
	return &Event{
		Time:              time.Now(),
		Name:              ev.ObjectMeta.Name,
		Namespace:         ev.ObjectMeta.Namespace,
		CreationTimestamp: ev.ObjectMeta.CreationTimestamp.Time,
		Labels:            ev.ObjectMeta.Labels,
		Annotations:       ev.ObjectMeta.Annotations,
		Kind:              ev.InvolvedObject.Kind,
		Reason:            ev.Reason,
		Message:           ev.Message,
		FirstTimestamp:    ev.FirstTimestamp.Time,
		LastTimestamp:     ev.LastTimestamp.Time,
		Count:             ev.Count,
		Type:              ev.Type,
		Action:            ev.Action,
		EventTime:         ev.EventTime.Time,
		Env:               env,
	}
}

func ConvertPodEvent(po *core_v1.Pod) *Event {
	env := os.Getenv("KUBEENV")
	if "" == env {
		env = "unknown"
	}

	ev := &Event{
		Time:              time.Now(),
		Name:              po.ObjectMeta.Name,
		Namespace:         po.ObjectMeta.Namespace,
		CreationTimestamp: po.ObjectMeta.CreationTimestamp.Time,
		Labels:            po.ObjectMeta.Labels,
		Annotations:       po.ObjectMeta.Annotations,
		Kind:              "Pod",
		Env:               env,
		ContainerStatus:   make(map[string]ContainerStatus),
	}

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

	for _, s := range po.Status.ContainerStatuses {
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
		ev.ContainerStatus[s.Name] = cs
	}

	return ev
}

func ConvertPodDeleteEvent(po *core_v1.Pod) *Event {
	env := os.Getenv("KUBEENV")
	if "" == env {
		env = "unknown"
	}

	return &Event{
		Time:              time.Now(),
		Name:              po.ObjectMeta.Name,
		Namespace:         po.ObjectMeta.Namespace,
		CreationTimestamp: po.ObjectMeta.CreationTimestamp.Time,
		Labels:            po.ObjectMeta.Labels,
		Annotations:       po.ObjectMeta.Annotations,
		Kind:              "Pod",
		Env:               env,
		Action:            "Delete",
	}
}
