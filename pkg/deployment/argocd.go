package deployment

import (
	argocd "github.com/argoproj-labs/argocd-operator/pkg/apis/argoproj/v1alpha1"
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewArgoCDCustomResource(cr *openshiftv1alpha1.Workshop, name string, namespace string, argocdPolicy string) *argocd.ArgoCD {

	scopes := "[preferred_username]"

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
			ApplicationInstanceLabelKey: "argocd.argoproj.io/instance",
			Dex: argocd.ArgoCDDexSpec{
				OpenShiftOAuth: true,
			},
			Server: argocd.ArgoCDServerSpec{
				Insecure: true,
				Route: argocd.ArgoCDRouteSpec{
					Enabled: true,
				},
			},
			RBAC: argocd.ArgoCDRBACSpec{
				Policy: &argocdPolicy,
				Scopes: &scopes,
			},
		},
	}
}
