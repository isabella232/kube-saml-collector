# kube-saml-collector
## Kubernetes-native automatic saml metadata aggregation

Usage:

1. Run kube-saml-collector as a sidecar container to your IdP pod, and use
   `-file` to output metadata to a directory (probably `emptyDir`) mounted in
   both containers.
2. Make sure your IdP is configured to read the metadata from the shared file.
3. For all SP pods that have valid saml metadata endpoints (usually
   `$APP/saml/metadata`), annotate them with the following options:
   - `k8s.unicon.net/saml-metadata-endpoint: /your/endpoint/here` *Required
	 - `k8s.unicon.net/saml-metadata-port: "8080"` *Optional
	 - `k8s.unicon.net/saml-metadata-protocol: "http"` *Optional; http by default
       for inside the k8s network

```
Usage of ./kube-saml-collector:
  -file string
    	Write all metadata out to the provided file name
  -host string
    	Set a custom kubernetes host. If unset, defaults to in-cluster config
  -interval duration
    	The polling interval for querying kubernetes pods (default 10s)
  -print-only
    	Set ./kube-saml-collector to only print out pod URLS that would have been collected
```
