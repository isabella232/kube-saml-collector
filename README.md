# kube-saml-collector
## Kubernetes-native automatic saml metadata aggregation

Usage:

1. Run kube-saml-collector as a [sidecar
   container](http://blog.kubernetes.io/2015/06/the-distributed-system-toolkit-patterns.html)
   to your IdP pod, and use
   `-file` to output metadata to a directory (probably `emptyDir`) mounted in
   both containers.
2. Make sure your IdP is configured to read the metadata from the shared file.
3. If you need to use a custom certificate authority for your cluster masters,
   include the full PEM-encoded certificate as the environment variable
   `ROOT_CA`.
4. For all SP pods that have valid saml metadata endpoints (usually
   `$APP/saml/metadata`), annotate them with the following options:
 - `k8s.unicon.net/saml-metadata-endpoint: /your/endpoint/here` *Required
 - `k8s.unicon.net/saml-metadata-port: "8080"` *Optional
 - `k8s.unicon.net/saml-metadata-protocol: "http"` *Optional; http by default
       for inside the k8s network

```
Usage of ./kube-saml-collector:
  -alsologtostderr
    	log to standard error as well as files
  -file string
    	Write all metadata out to the provided file name
  -host string
    	Set a custom kubernetes host. If unset, defaults to in-cluster config
  -interval duration
    	The polling interval for querying kubernetes pods (default 10s)
  -log_backtrace_at value
    	when logging hits line file:N, emit a stack trace
  -log_dir string
    	If non-empty, write log files in this directory
  -logtostderr
    	log to standard error instead of files
  -metrics-listen string
    	The $IP:$PORT address to listen on for metrics requests (default ":8080")
  -namespace string
    	The kubernetes namespace to scan for metadata. Defaults to all namespaces.
  -print-only
    	Set ./kube-saml-collector to only print out pod URLS that would have been collected
  -serve-aggregate
    	If true, the container will itself serve its aggregated metadata at /saml/metadata on the http listener addr/port
  -stderrthreshold value
    	logs at or above this threshold go to stderr
  -v value
    	log level for V logs
  -vmodule value
    	comma-separated list of pattern=N settings for file-filtered logging
```
