package squash

import (
	rbac "k8s.io/api/rbac/v1"
)

func NewRules() []rbac.PolicyRule {
	return []rbac.PolicyRule{
		{
			Verbs:     []string{"get", "list", "watch", "create", "delete"},
			Resources: []string{"pods"},
			APIGroups: []string{""},
		},
		{
			Verbs:     []string{"list"},
			Resources: []string{"namespaces"},
			APIGroups: []string{""},
		},
		{
			Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
			Resources: []string{"debugattachments"},
			APIGroups: []string{"squash.solo.io"},
		},
		{
			Verbs:     []string{"create"},
			Resources: []string{"clusterrolebindings"},
			APIGroups: []string{"rbac.authorization.k8s.io"},
		},
		{
			Verbs:     []string{"create"},
			Resources: []string{"clusterrole"},
			APIGroups: []string{"rbac.authorization.k8s.io"},
		},
		{
			Verbs:     []string{"create"},
			Resources: []string{"serviceaccount"},
			APIGroups: []string{""},
		},
		{
			Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
			Resources: []string{"customresourcedefinitions"},
			APIGroups: []string{"apiextensions.k8s.io"},
		},
	}
}
