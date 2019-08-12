package deployment

import (
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NewRoleParameters struct {
	Name      string
	Namespace string
	Rules     []rbac.PolicyRule
}

func NewRole(param NewRoleParameters) *rbac.Role {
	return &rbac.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: rbac.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      param.Name,
			Namespace: param.Namespace,
		},
		Rules: param.Rules,
	}
}
