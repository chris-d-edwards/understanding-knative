module lazy-service-controller

go 1.16

require (
	github.com/go-logr/logr v0.4.0
	github.com/gogo/protobuf v1.3.2
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.3
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/common v0.19.0
	istio.io/api v0.0.0-20210218044411-561dc276d04d // indirect
	istio.io/client-go v1.9.1
	k8s.io/api v0.20.1
	k8s.io/apimachinery v0.20.1
	k8s.io/client-go v0.20.1
	knative.dev/pkg v0.0.0-20210318052054-dfeeb1817679
	knative.dev/serving v0.21.0
	sigs.k8s.io/controller-runtime v0.7.0
)
