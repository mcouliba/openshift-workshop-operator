package deployment

import (
	argocd "github.com/jmckind/argocd-operator/pkg/apis/argoproj/v1alpha1"
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewArgoCDCustomResource(cr *openshiftv1alpha1.Workshop, name string, namespace string) *argocd.ArgoCD {
	return &argocd.ArgoCD{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ArgoCD",
			APIVersion: "argoproj.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: argocd.ArgoCDSpec{
			Image:   "argoproj/argocd",
			Version: "v1.4.2",
			Grafana: argocd.ArgoCDGrafanaSpec{
				Enabled: false,
			},
			Ingress: argocd.ArgoCDIngressSpec{
				Enabled: false,
			},
			Prometheus: argocd.ArgoCDPrometheusSpec{
				Enabled: false,
			},
			Server: argocd.ArgoCDServerSpec{
				Insecure: true,
			},
			Dex: argocd.ArgoCDDexSpec{
				Image:   "quay.io/ablock/dex",
				Version: "openshift-connector",
			},
		},
	}
}
