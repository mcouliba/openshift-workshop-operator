package customresource

import (
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewGogsCustomResource(cr *openshiftv1alpha1.Workshop, name string, namespace string) *Gogs {
	return &Gogs{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Gogs",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: GogsSpec{
			GogsVolumeSize:       "4Gi",
			GogsSsl:              false,
			PostgresqlVolumeSize: "4Gi",
		},
	}
}
