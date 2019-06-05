package deployment

import (
	cloudnativev1alpha1 "github.com/redhat/cloud-native-workshop-operator/pkg/apis/cloudnative/v1alpha1"
)

func GetLabels(cr *cloudnativev1alpha1.Workshop, component string) (labels map[string]string) {
	labels = map[string]string{"app": "cloud-native-workshop", "component": component}
	return labels
}
