package maistra

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewServiceMeshMemberRollCR(param NewServiceMeshMemberRollCRParameters) *ServiceMeshMemberRoll {
	return &ServiceMeshMemberRoll{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceMeshMemberRoll",
			APIVersion: "maistra.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      param.Name,
			Namespace: param.Namespace,
		},
		Spec: ServiceMeshMemberRollSpec{
			Members: param.Members,
		},
	}
}
