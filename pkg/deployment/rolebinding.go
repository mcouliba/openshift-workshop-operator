package deployment

import (
	cloudnativev1alpha1 "github.com/redhat/cloud-native-workshop-operator/pkg/apis/cloudnative/v1alpha1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewRoleBindingForServiceAccount(cr *cloudnativev1alpha1.Workshop, name string, namespace string, serviceAccountName string, roleName string, roleKind string) *rbac.RoleBinding {
	labels := GetLabels(cr, name)
	return &rbac.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: rbac.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Subjects: []rbac.Subject{
			{
				Kind:      rbac.ServiceAccountKind,
				Name:      serviceAccountName,
				Namespace: namespace,
			},
		},
		RoleRef: rbac.RoleRef{
			Name:     roleName,
			Kind:     roleKind,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

func NewRoleBindingForUser(cr *cloudnativev1alpha1.Workshop, name string, namespace string, username string, roleName string, roleKind string) *rbac.RoleBinding {
	labels := GetLabels(cr, name)
	return &rbac.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: rbac.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Subjects: []rbac.Subject{
			{
				Kind: rbac.UserKind,
				Name: username,
			},
		},
		RoleRef: rbac.RoleRef{
			Name:     roleName,
			Kind:     roleKind,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}
