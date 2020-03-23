package deployment

import (
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	admissionregistration "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewMutatingWebhookConfiguration(cr *openshiftv1alpha1.Workshop, name string, labels map[string]string,
	webhooks []admissionregistration.Webhook) *admissionregistration.MutatingWebhookConfiguration {
	return &admissionregistration.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Webhooks: webhooks,
	}
}
