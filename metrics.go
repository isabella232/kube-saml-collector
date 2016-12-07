package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var collections = prometheus.NewSummary(prometheus.SummaryOpts{
	Name: "kube-saml-collector_collection_seconds",
	Help: "A summary of the metadata aggregation performed by kube-saml-collector",
})

func registerAndServe() {
	prometheus.MustRegister(collections)
	go http.ListenAndServe(":8080", promhttp.Handler())
}

func init() {
	registerAndServe()
}
