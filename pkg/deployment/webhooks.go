package deployment

import (
	admissionregistration "k8s.io/api/admissionregistration/v1beta1"
)

func IstioWebHook() []admissionregistration.Webhook {
	var smcp = new(string)
	*smcp = "/validate-smcp"

	var smmr = new(string)
	*smmr = "/validate-smmr"

	var failurePolicy = new(admissionregistration.FailurePolicyType)
	*failurePolicy = admissionregistration.Fail

	return []admissionregistration.Webhook{
		{
			Name:          "smcp.validation.maistra.io",
			FailurePolicy: failurePolicy,
			ClientConfig: admissionregistration.WebhookClientConfig{
				CABundle: []byte{},
				Service: &admissionregistration.ServiceReference{
					Name:      "admission-controller",
					Namespace: "istio-operator",
					Path:      smcp,
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
							"maistra.io",
						},
						APIVersions: []string{
							"v1",
						},
						Resources: []string{
							"servicemeshcontrolplanes",
						},
					},
				},
			},
		},
		{
			Name:          "smmr.validation.maistra.io",
			FailurePolicy: failurePolicy,
			ClientConfig: admissionregistration.WebhookClientConfig{
				CABundle: []byte{},
				Service: &admissionregistration.ServiceReference{
					Name:      "admission-controller",
					Namespace: "istio-operator",
					Path:      smmr,
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
							"maistra.io",
						},
						APIVersions: []string{
							"v1",
						},
						Resources: []string{
							"servicemeshmemberrolls",
						},
					},
				},
			},
		},
	}
}
