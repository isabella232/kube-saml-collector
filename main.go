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
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

const (
	annotSAMLMetadataEndpoint      = "k8s.unicon.net/saml-metadata-endpoint"
	annotationSAMLMetadataPort     = "k8s.unicon.net/saml-metadata-port"
	annotationSAMLMetadataProtocol = "k8s.unicon.net/saml-metadata-protocol"

	defaultPort  = "8080"
	defaultProto = "http"

	envRootCA = "ROOT_CA"
)

var (
	host      = flag.String("host", "", "Set a custom kubernetes host. If unset, defaults to in-cluster config")
	printOnly = flag.Bool("print-only", false, fmt.Sprintf("Set %s to only print out pod URLS that would have been collected", os.Args[0]))
	interval  = flag.Duration("interval", 10*time.Second, "The polling interval for querying kubernetes pods")

	file           = flag.String("file", "", "Write all metadata out to the provided file name")
	serveAggregate = flag.Bool("serve-aggregate", false, "If true, the container will itself serve its aggregated metadata at /saml/metadata on the http listener addr/port")
	metricsAddr    = flag.String("metrics-listen", ":8000", "The $IP:$PORT address to listen on for metrics requests")
)

func init() {
	flag.Parse()
}

func main() {
	maybeServeMetadata()
	registerAndServeMetrics()

	var cfg *rest.Config

	switch *host {
	case "":
		var err error
		cfg, err = rest.InClusterConfig()
		if err != nil {
			glog.Fatal(err)
		}
	default:
		cfg = &rest.Config{
			Host: *host,
		}
	}

	ca := os.Getenv(envRootCA)
	if ca != "" {
		glog.Info("Root CA pulled from environment: " + ca)
		cfg.CAData = []byte(ca)
	}

	cli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatal(err)
	}

	for {
		t := prometheus.NewTimer(collections)

		pods, err := cli.Pods("").List(v1.ListOptions{})

		if err != nil {
			glog.Fatal(err)
		}

		agg := aggregator{}

		for i := range pods.Items {

			if pods.Items[i].Annotations[annotSAMLMetadataEndpoint] == "" {
				continue
			}

			url := getURL(&pods.Items[i])

			fmt.Fprintln(os.Stderr, "Retrieving URL", url)
			if *printOnly {
				continue
			}

			res, err := http.Get(url)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			err = agg.add(res.Body)
			res.Body.Close()

			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
		}

		if *file != "" {
			w := os.Stdout
			if *file != "" {
				w, err = os.OpenFile(*file, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
				if err != nil {
					glog.Fatalf("Error opening file %s for write:\n\t%s", *file, err)
				}
			}

			agg.WriteTo(w)
			err = w.Close()
			if err != nil {
				glog.Error("Error closing file", err)
			}
		}

		t.ObserveDuration()

		time.Sleep(*interval)
	}
}

func getURL(p *v1.Pod) string {
	port := p.Annotations[annotationSAMLMetadataPort]
	if port == "" {
		port = defaultPort
	}

	proto := p.Annotations[annotationSAMLMetadataProtocol]
	if proto == "" {
		proto = defaultProto
	}

	endpoint := strings.TrimLeft(p.Annotations[annotSAMLMetadataEndpoint], "/")

	return fmt.Sprintf("%s://%s:%s/%s", proto, p.Status.PodIP, port, endpoint)
}
