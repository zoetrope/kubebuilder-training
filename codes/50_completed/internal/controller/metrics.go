package controller

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	metricsNamespace = "markdownview"
)

var (
	AvailableVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Name:      "available",
		Help:      "The cluster status about available condition",
	}, []string{"name", "namespace"})
)

func init() {
	metrics.Registry.MustRegister(AvailableVec)
}
