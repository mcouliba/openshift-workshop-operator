package deployment

import (
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewSubscription(cr *openshiftv1alpha1.Workshop, name string, namespace string, startingCSV string) *olmv1alpha1.Subscription {
	return &olmv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "operators.coreos.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string {
				"csc-owner-name": "installed-"+name,
    			"csc-owner-namespace": "openshift-marketplace",
			},
		},
		Spec: &olmv1alpha1.SubscriptionSpec{
			Channel: "stable",
			CatalogSource:  "installed-"+name,
			CatalogSourceNamespace: namespace,
			StartingCSV: startingCSV,
			InstallPlanApproval: olmv1alpha1.ApprovalAutomatic,
			Package: name,
		},
	}
}