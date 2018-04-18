package model

const (
	ContainerStatusWaiting    = "Waiting"
	ContainerStatusTerminated = "Terminated"
	ContainerStatusRunning    = "Running"
)

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
