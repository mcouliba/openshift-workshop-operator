package deployment

import (
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
)

func GetLabels(cr *openshiftv1alpha1.Workshop, component string) (labels map[string]string) {
	labels = map[string]string{"app": "openshift-workshop", "component": component}
	return labels
}
