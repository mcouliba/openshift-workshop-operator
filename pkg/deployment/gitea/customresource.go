package customresource

import (
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewGiteaCustomResource(cr *openshiftv1alpha1.Workshop, name string, namespace string) *Gitea {
	return &Gitea{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Gitea",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: GiteaSpec{
			GiteaVolumeSize:      "4Gi",
			GiteaSsl:             true,
			PostgresqlVolumeSize: "4Gi",
		},
	}
}
