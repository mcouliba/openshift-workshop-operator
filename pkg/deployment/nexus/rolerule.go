package nexus

import (
	rbac "k8s.io/api/rbac/v1"
)

func NewRules() []rbac.PolicyRule {
	return []rbac.PolicyRule{
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"pods/exec",
			},
			Verbs: []string{
				"create",
			},
		},
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
				"route.openshift.io",
			},
			Resources: []string{
				"routes",
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
				"nexus",
				"nexus/status",
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
				"nexus-operator",
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
