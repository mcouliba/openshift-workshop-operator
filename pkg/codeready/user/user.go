package codeready

import (
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
)

func NewCodeReadyUser(cr *openshiftv1alpha1.Workshop, username string, password string) *User {
	return &User{
		Username: username,
		Enabled:  true,
		Email:    username + "@none.com",
		Credentials: []Credential{
			{
				Type:  "password",
				Value: password,
			},
		},
		ClientRoles: ClientRoles{
			RealmManagement: []string{
				"user",
			},
		},
	}
}
