module github.com/redhat/openshift-workshop-operator

go 1.14

require (
	contrib.go.opencensus.io/exporter/ocagent v0.4.9 // indirect
	github.com/Azure/go-autorest v11.5.2+incompatible // indirect
	github.com/appscode/jsonpatch v0.0.0-20190108182946-7c0e3b262f30 // indirect
	github.com/argoproj-labs/argocd-operator v0.0.12
	github.com/eclipse/che-operator v0.0.0-20191211154745-df0be398efea
	github.com/evanphx/json-patch v4.1.0+incompatible // indirect
	github.com/go-logr/logr v0.1.0 // indirect
	github.com/go-logr/zapr v0.1.0 // indirect
	github.com/go-openapi/spec v0.19.7
	github.com/go-openapi/validate v0.18.0 // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gophercloud/gophercloud v0.0.0-20190318015731-ff9851476e98 // indirect
	github.com/kr/pty v1.1.5 // indirect
	github.com/openshift/api v3.9.1-0.20190916204813-cdbe64fb0c91+incompatible
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-20190522155654-7b2b397ad0cc
	github.com/operator-framework/operator-marketplace v0.0.0-20190820221623-6881ba35b740
	github.com/operator-framework/operator-sdk v0.18.0
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/sirupsen/logrus v1.4.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/objx v0.2.0 // indirect
	go.opencensus.io v0.19.2 // indirect
	go.uber.org/atomic v1.3.2 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	golang.org/x/crypto v0.0.0-20190611184440-5c40567a22f8 // indirect
	golang.org/x/sys v0.0.0-20190616124812-15dcb6c0061f // indirect
	golang.org/x/tools v0.0.0-20190614205625-5aca471b1d59 // indirect
	k8s.io/api v0.18.3
	k8s.io/apiextensions-apiserver v0.0.0-20190228180357-d002e88f6236
	k8s.io/apimachinery v0.18.3
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.0.0-20181203235156-f8cba74510f3
	k8s.io/gengo v0.0.0-20190327210449-e17681d19d3a
	k8s.io/klog v0.3.0 // indirect
	k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	sigs.k8s.io/controller-runtime v0.6.0
	sigs.k8s.io/controller-tools v0.1.10
	sigs.k8s.io/testing_frameworks v0.1.0 // indirect
	sigs.k8s.io/yaml v1.1.0
)

// Pinned to kubernetes-1.13.1
replace (
	k8s.io/api => k8s.io/api v0.0.0-20181213150558-05914d821849
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20181213153335-0fe22c71c476
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20181127025237-2b1284ed4c93
	k8s.io/client-go => k8s.io/client-go v0.0.0-20181213151034-8d9ed539ba31
)

replace (
	github.com/coreos/prometheus-operator => github.com/coreos/prometheus-operator v0.29.0
	github.com/operator-framework/operator-sdk => github.com/operator-framework/operator-sdk v0.8.0
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20181117043124-c2090bec4d9b
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20180711000925-0cf8f7e6ed1d
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.1.10
	sigs.k8s.io/controller-tools => sigs.k8s.io/controller-tools v0.1.11-0.20190411181648-9d55346c2bde
)
