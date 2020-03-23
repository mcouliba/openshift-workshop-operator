package deployment

import (
	admissionregistration "k8s.io/api/admissionregistration/v1beta1"
)

func VaultAgentInjectorWebHook(namespace string) []admissionregistration.Webhook {
	path := "/mutate"

	return []admissionregistration.Webhook{
		{
			Name: "vault.hashicorp.com",
			ClientConfig: admissionregistration.WebhookClientConfig{
				CABundle: []byte{},
				Service: &admissionregistration.ServiceReference{
					Name:      "vault-agent-injector",
					Namespace: namespace,
					Path:      &path,
				},
			},
			Rules: []admissionregistration.RuleWithOperations{
				{
					Operations: []admissionregistration.OperationType{
						"CREATE",
						"UPDATE",
					},
					Rule: admissionregistration.Rule{
						APIGroups: []string{
							"",
						},
						APIVersions: []string{
							"v1",
						},
						Resources: []string{
							"pods",
						},
					},
				},
			},
		},
	}
}
