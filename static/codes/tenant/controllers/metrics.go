package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	addedNamespaces = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "added_namespaces",
			Help: "Number of added namespaces",
		},
	)
	removedNamespaces = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "removed_namespaces",
			Help: "Number of removed namespaces",
		},
	)
)

func init() {
	metrics.Registry.MustRegister(addedNamespaces, removedNamespaces)
}
