module github.com/crowdstrike/falcon-operator

go 1.15

require (
	github.com/containers/image/v5 v5.10.1
	github.com/crowdstrike/gofalcon v0.2.1
	github.com/go-logr/logr v0.3.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/openshift/api v0.0.0-20201120165435-072a4cd8ca42
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	sigs.k8s.io/controller-runtime v0.7.0
)