package deployment

import (
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewCatalogSource(cr *openshiftv1alpha1.Workshop, name string, image string,
	displayName string, publisher string) *olmv1alpha1.CatalogSource {
	return &olmv1alpha1.CatalogSource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CatalogSource",
			APIVersion: "operators.coreos.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "openshift-marketplace",
		},
		Spec: olmv1alpha1.CatalogSourceSpec{
			SourceType:  olmv1alpha1.SourceTypeGrpc,
			Image:       image,
			DisplayName: displayName,
			Publisher:   publisher,
		},
	}
}
