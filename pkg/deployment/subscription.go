package deployment

import (
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewCertifiedSubscription(cr *openshiftv1alpha1.Workshop, name string, namespace string, packageName string, channel string, startingCSV string) *olmv1alpha1.Subscription {
	return &olmv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "operators.coreos.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"csc-owner-name":      "certified-operators",
				"csc-owner-namespace": "openshift-marketplace",
			},
		},
		Spec: &olmv1alpha1.SubscriptionSpec{
			Channel:                channel,
			CatalogSource:          "certified-operators",
			CatalogSourceNamespace: "openshift-marketplace",
			StartingCSV:            startingCSV,
			InstallPlanApproval:    olmv1alpha1.ApprovalManual,
			Package:                packageName,
		},
	}
}

func NewCommunitySubscription(cr *openshiftv1alpha1.Workshop, name string, namespace string, packageName string, channel string, startingCSV string) *olmv1alpha1.Subscription {
	return &olmv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "operators.coreos.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"csc-owner-name":      "community-operators",
				"csc-owner-namespace": "openshift-marketplace",
			},
		},
		Spec: &olmv1alpha1.SubscriptionSpec{
			Channel:                channel,
			CatalogSource:          "community-operators",
			CatalogSourceNamespace: "openshift-marketplace",
			StartingCSV:            startingCSV,
			InstallPlanApproval:    olmv1alpha1.ApprovalManual,
			Package:                packageName,
		},
	}
}

func NewRedHatSubscription(cr *openshiftv1alpha1.Workshop, name string, namespace string, packageName string, channel string, startingCSV string) *olmv1alpha1.Subscription {
	return &olmv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "operators.coreos.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"csc-owner-name":      "redhat-operators",
				"csc-owner-namespace": "openshift-marketplace",
			},
		},
		Spec: &olmv1alpha1.SubscriptionSpec{
			Channel:                channel,
			CatalogSource:          "redhat-operators",
			CatalogSourceNamespace: "openshift-marketplace",
			StartingCSV:            startingCSV,
			InstallPlanApproval:    olmv1alpha1.ApprovalManual,
			Package:                packageName,
		},
	}
}

func NewCustomSubscription(cr *openshiftv1alpha1.Workshop, name string, namespace string, packageName string,
	channel string, catalogSource string) *olmv1alpha1.Subscription {
	return &olmv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "operators.coreos.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"csc-owner-name":      "custom-operators",
				"csc-owner-namespace": "openshift-marketplace",
			},
		},
		Spec: &olmv1alpha1.SubscriptionSpec{
			Channel:                channel,
			CatalogSource:          catalogSource,
			CatalogSourceNamespace: "openshift-marketplace",
			Package:                packageName,
		},
	}
}
