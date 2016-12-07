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
	printOnly = flag.Bool("print-only", false, fmt.Sprintf("Set %s to only print out pod URLS that would have been collected", os.Args[0]))
	interval  = flag.Duration("interval", 10*time.Second, "The polling interval for querying kubernetes pods")
)

func init() {
	flag.Parse()
}

func main() {
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
		pods, err := cli.Pods("").List(v1.ListOptions{})

		if err != nil {
			log.Fatal(err)
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
			io.Copy(os.Stdout, res.Body)
			res.Body.Close()
		}

		time.Sleep(10 * time.Second)
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
