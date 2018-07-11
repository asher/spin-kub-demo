# Sample webserver that emits a custom stackdriver metric for a Kayenta workshop

Source to prod Kubernetes Spinnaker artifact code is staged here. Simple webserver that emits a stackdriver metric
called custom.googleapis.com/workshop/canary/request/errors tagged with spinnaker cluster and server group names.
The metric is randomly generated and will be high if the cluster contains canary in the name, or low if the baseline.
