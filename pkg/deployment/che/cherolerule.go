package che

import (
	rbac "k8s.io/api/rbac/v1"
)

func CheRules() []rbac.PolicyRule {
	return []rbac.PolicyRule{
		{
			APIGroups: []string{
				"extensions/v1beta1",
			},
			Resources: []string{
				"ingresses",
			},
			Verbs: []string{
				"*",
			},
		},
		{
			APIGroups: []string{
				"route.openshift.io",
			},
			Resources: []string{
				"routes",
			},
			Verbs: []string{
				"*",
			},
		},
		{
			APIGroups: []string{
				"rbac.authorization.k8s.io",
			},
			Resources: []string{
				"roles",
				"rolebindings",
			},
			Verbs: []string{
				"*",
			},
		},
		{
			APIGroups: []string{
				"rbac.authorization.k8s.io",
			},
			Resources: []string{
				"clusterroles",
				"clusterrolebindings",
			},
			Verbs: []string{
				"*",
			},
		},
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"pods",
				"services",
				"serviceaccounts",
				"endpoints",
				"persistentvolumeclaims",
				"events",
				"configmaps",
				"secrets",
				"pods/exec",
				"pods/log",
			},
			Verbs: []string{
				"*",
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
			},
			Verbs: []string{
				"*",
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
				"org.eclipse.che",
			},
			Resources: []string{
				"*",
			},
			Verbs: []string{
				"*",
			},
		},
		{
			APIGroups: []string{
				"oauth.openshift.io",
			},
			Resources: []string{
				"oauthclients",
			},
			Verbs: []string{
				"get",
				"create",
				"delete",
				"update",
				"list",
				"watch",
			},
		},
	}
}
