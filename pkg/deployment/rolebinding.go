package deployment

import (
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NewRoleBindingSAParameters struct {
	Name               string
	Namespace          string
	labels             map[string]string
	ServiceAccountName string
	RoleName           string
	RoleKind           string
}

type NewRoleBindingUserParameters struct {
	Name      string
	Namespace string
	labels    map[string]string
	Username  string
	RoleName  string
	RoleKind  string
}

func NewRoleBindingSA(param NewRoleBindingSAParameters) *rbac.RoleBinding {
	return &rbac.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: rbac.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      param.Name,
			Namespace: param.Namespace,
			Labels:    param.labels,
		},
		Subjects: []rbac.Subject{
			{
				Kind:      rbac.ServiceAccountKind,
				Name:      param.ServiceAccountName,
				Namespace: param.Namespace,
			},
		},
		RoleRef: rbac.RoleRef{
			Name:     param.RoleName,
			Kind:     param.RoleKind,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

func NewRoleBindingUser(param NewRoleBindingUserParameters) *rbac.RoleBinding {
	return &rbac.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: rbac.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      param.Name,
			Namespace: param.Namespace,
			Labels:    param.labels,
		},
		Subjects: []rbac.Subject{
			{
				Kind: rbac.UserKind,
				Name: param.Username,
			},
		},
		RoleRef: rbac.RoleRef{
			Name:     param.RoleName,
			Kind:     param.RoleKind,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}
