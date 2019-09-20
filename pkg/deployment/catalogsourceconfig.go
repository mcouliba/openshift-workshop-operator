package deployment

import (
	ompv1 "github.com/operator-framework/operator-marketplace/pkg/apis/operators/v1"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewCatalogSourceConfig(cr *openshiftv1alpha1.Workshop, name string, namespace string, packages string) *ompv1.CatalogSourceConfig {
	return &ompv1.CatalogSourceConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CatalogSourceConfig",
			APIVersion: "operators.coreos.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "openshift-marketplace",
		},
		Spec: ompv1.CatalogSourceConfigSpec{
			Packages:        packages,
			TargetNamespace: namespace,
		},
	}
}
