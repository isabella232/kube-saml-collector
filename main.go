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
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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
)

var (
	host      = flag.String("host", "", "Set a custom kubernetes host. If unset, defaults to in-cluster config")
	file      = flag.String("file", "", "Write all metadata out to the provided file name")
	printOnly = flag.Bool("print-only", false, fmt.Sprintf("Set %s to only print out pod URLS that would have been collected", os.Args[0]))
	interval  = flag.Duration("interval", 10*time.Second, "The polling interval for querying kubernetes pods")

	metricsAddr = flag.String("metrics-listen", ":8080", "The $IP:$PORT address to listen on for metrics requests")
)

func init() {
	flag.Parse()
}

func main() {
	registerAndServeMetrics()

	var cfg *rest.Config

	switch *host {
	case "":
		var err error
		cfg, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal(err)
		}
	default:
		cfg = &rest.Config{
			Host: *host,
		}
	}

	cli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	for {
		t := prometheus.NewTimer(collections)

		pods, err := cli.Pods("").List(v1.ListOptions{})

		if err != nil {
			log.Fatal(err)
		}

		var w *os.File
		switch *file {
		case "":
			w = os.Stdout
		default:
			w, err = os.OpenFile(*file, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
			if err != nil {
				log.Fatalf("Error opening file %s for write:\n\t%s", *file, err)
			}
		}

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
			}
			_, err = io.Copy(w, res.Body)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			res.Body.Close()
		}

		if *file != "" {
			err = w.Close()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error closing file", err)
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
