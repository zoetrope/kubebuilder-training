package controller

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	metricsNamespace = "markdownview"
)

var (
	NotReadyVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Name:      "notready",
		Help:      "The cluster status about not ready condition",
	}, []string{"name", "namespace"})

	AvailableVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Name:      "available",
		Help:      "The cluster status about available condition",
	}, []string{"name", "namespace"})

	HealthyVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Name:      "healthy",
		Help:      "The cluster status about healthy condition",
	}, []string{"name", "namespace"})
)

func init() {
	metrics.Registry.MustRegister(NotReadyVec, AvailableVec, HealthyVec)
}
