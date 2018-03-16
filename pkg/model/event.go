package model

import (
	"os"
	"time"

	core_v1 "k8s.io/api/core/v1"
)

type Event struct {
	Time              time.Time         `json:"time"`
	Name              string            `json:"name,omitempty"`
	Namespace         string            `json:"namespace,omitempty"`
	CreationTimestamp time.Time         `json:"creationTimestamp,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	Kind              string            `json:"kind,omitempty"`
	Reason            string            `json:"reason,omitempty"`
	Message           string            `json:"message,omitempty"`
	FirstTimestamp    time.Time         `json:"firstTimestamp,omitempty"`
	LastTimestamp     time.Time         `json:"lastTimestamp,omitempty"`
	Count             int32             `json:"count,omitempty"`
	Type              string            `json:"type,omitempty"`
	Action            string            `json:"action,omitempty"`
	EventTime         time.Time         `json:"eventTime"`
	Env               string            `json:"env"`
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
