// Copyright 2016 Andrew Stuart

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var collections = prometheus.NewSummary(prometheus.SummaryOpts{
	Name: "kube_saml_collector_collection_seconds",
	Help: "A summary of the metadata aggregation performed by kube-saml-collector",
})

func registerAndServeMetrics() {
	prometheus.MustRegister(collections)
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(*metricsAddr, nil)
}
