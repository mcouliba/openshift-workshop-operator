package deployment

import (
	rbac "k8s.io/api/rbac/v1"
)

func GiteaRules() []rbac.PolicyRule {
	return []rbac.PolicyRule{
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"pods",
				"services",
				"endpoints",
				"persistentvolumeclaims",
				"events",
				"configmaps",
				"secrets",
				"serviceaccounts",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"apps",
			},
			Resources: []string{
				"deployments",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"apps",
			},
			Resources: []string{
				"replicasets",
			},
			Verbs: []string{
				"get",
			},
		},
		{
			APIGroups: []string{
				"route.openshift.io",
			},
			Resources: []string{
				"routes",
				"routes/custom-host",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"monitoring.coreos.com",
			},
			Resources: []string{
				"servicemonitors",
			},
			Verbs: []string{
				"get",
				"create",
			},
		},
		{
			APIGroups: []string{
				"gpte.opentlc.com",
			},
			Resources: []string{
				"*",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"apps",
			},
			ResourceNames: []string{
				"gitea-operator",
			},
			Resources: []string{
				"deployments/finalizers",
			},
			Verbs: []string{
				"update",
			},
		},
	}
}

func IstioUserRules() []rbac.PolicyRule {
	return []rbac.PolicyRule{
		{
			APIGroups: []string{
				"networking.istio.io",
			},
			Resources: []string{
				"destinationrules",
				"gateways",
				"bypasses",
				"serviceentries",
				"sidecars",
				"virtualservices",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"authentication.istio.io",
			},
			Resources: []string{
				"policies",
				"meshpolicies",
			},
			Verbs: []string{
				"list",
			},
		},
		{
			APIGroups: []string{
				"config.istio.io",
			},
			Resources: []string{
				"rules",
				"quotaspecs",
				"quotaspecbindings",
			},
			Verbs: []string{
				"list",
			},
		},
		{
			APIGroups: []string{
				"rbac.istio.io",
			},
			Resources: []string{
				"serviceroles",
				"rbacconfigs",
				"servicerolebindings",
			},
			Verbs: []string{
				"list",
			},
		},
	}
}

func IstioArgoCDRules() []rbac.PolicyRule {
	return []rbac.PolicyRule{
		{
			APIGroups: []string{
				"networking.istio.io",
			},
			Resources: []string{
				"destinationrules",
				"gateways",
				"virtualservices",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
	}
}

func JaegerUserRules() []rbac.PolicyRule {
	return []rbac.PolicyRule{
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"pods",
			},
			Verbs: []string{
				"get",
			},
		},
	}
}

func IstioWorkspaceRules() []rbac.PolicyRule {
	return []rbac.PolicyRule{
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"pods",
				"services",
				"endpoints",
				"persistentvolumeclaims",
				"events",
				"configmaps",
				"secrets",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"namespaces",
			},
			Verbs: []string{
				"get",
			},
		},
		{
			APIGroups: []string{
				"apps",
			},
			Resources: []string{
				"deployments",
				"daemonsets",
				"replicasets",
				"statefulsets",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"apps.openshift.io",
			},
			Resources: []string{
				"deploymentconfigs",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"monitoring.coreos.com",
			},
			Resources: []string{
				"servicemonitors",
			},
			Verbs: []string{
				"get",
				"create",
			},
		},
		{
			APIGroups: []string{
				"istio.openshift.com",
				"networking.istio.io",
				"maistra.io",
			},
			Resources: []string{
				"*",
			},
			Verbs: []string{
				"*",
			},
		},
	}
}

func IstioWorkspaceUserRules() []rbac.PolicyRule {
	return []rbac.PolicyRule{
		{
			APIGroups: []string{
				"maistra.io",
			},
			Resources: []string{
				"sessions",
			},
			Verbs: []string{
				"create",
				"update",
				"delete",
				"get",
				"list",
				"watch",
				"patch",
			},
		},
	}
}

func VaultAgentInjectorRules() []rbac.PolicyRule {
	return []rbac.PolicyRule{
		{
			APIGroups: []string{
				"admissionregistration.k8s.io",
			},
			Resources: []string{
				"mutatingwebhookconfigurations",
			},
			Verbs: []string{
				"get",
				"list",
				"watch",
				"patch",
			},
		},
	}
}

func CheRules() []rbac.PolicyRule {
	return []rbac.PolicyRule{
		{
			APIGroups: []string{
				"project.openshift.io",
			},
			Resources: []string{
				"projectrequests",
			},
			Verbs: []string{
				"create",
			},
		},
	}
}
