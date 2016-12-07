package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

const (
	LabelSAMLMetadataEndpoint = "k8s.unicon.net/saml-metadata-endpoint"
	LabelSAMLMetadataPort     = "k8s.unicon.net/saml-metadata-port"
	LabelSAMLMetadataProtocol = "k8s.unicon.net/saml-metadata-protocol"

	DefaultPort     = "8080"
	DefaultProtocol = "http"
)

var (
	host = flag.String("host", "", "Set a custom kubernetes host. If unset, defaults to in-cluster config")
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

	pods, err := cli.Pods("").List(v1.ListOptions{
		FieldSelector: "metadata.annotations." + LabelSAMLMetadataEndpoint,
	})

	if err != nil {
		log.Fatal(err)
	}

	for i := range pods.Items {
		res, err := http.Get(getURL(&pods.Items[i]))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		io.Copy(os.Stdout, res.Body)
		res.Body.Close()
	}
}

func getURL(p *v1.Pod) string {
	port := p.Labels[LabelSAMLMetadataPort]
	if port == "" {
		port = DefaultPort
	}

	proto := p.Labels[LabelSAMLMetadataProtocol]
	if proto == "" {
		proto = DefaultProtocol
	}

	endpoint := strings.TrimLeft(p.Labels[LabelSAMLMetadataEndpoint], "/")

	return fmt.Sprintf("%s://%s:%s/%s", proto, p.Status.PodIP, port, endpoint)
}
