package collector

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/jojohappy/luxun/pkg/storage"
)

var podStatusDesc = prometheus.NewDesc("kube_pod_status_reason_count", "Count of the aggregate status of the containers in the pod.", []string{"reason"}, nil)

type podStatusCollector struct{}

func NewCollector() *podStatusCollector {
	return &podStatusCollector{}
}

func (p *podStatusCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- podStatusDesc
}

func (p *podStatusCollector) Collect(ch chan<- prometheus.Metric) {
	status := storage.StorageInst().GetAll()
	for s, count := range status {
		ch <- prometheus.MustNewConstMetric(podStatusDesc, prometheus.CounterValue, count, s)
	}
}
