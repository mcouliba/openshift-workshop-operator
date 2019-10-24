package deployment

import (
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewNetworkPolicyAllowFromOpenShfitIngress(name string, namespace string) *networking.NetworkPolicy {
	return &networking.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NetworkPolicy",
			APIVersion: networking.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: networking.NetworkPolicySpec{
			Ingress: []networking.NetworkPolicyIngressRule{
				{
					From: []networking.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"network.openshift.io/policy-group": "ingress",
								},
							},
						},
					},
				},
			},
			PolicyTypes: []networking.PolicyType{
				networking.PolicyTypeIngress,
			},
		},
	}
}
