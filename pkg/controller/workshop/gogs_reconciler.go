package workshop

import (
	"context"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	gogscustomresource "github.com/redhat/openshift-workshop-operator/pkg/customresource/gogs"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	"k8s.io/apimachinery/pkg/api/errors"
)

// Reconciling Gogs
func (r *ReconcileWorkshop) reconcileGogs(instance *openshiftv1alpha1.Workshop) error {
	enabledGogs := instance.Spec.Gogs.Enabled

	if enabledGogs {
		if err := r.addGogs(instance); err != nil {
			return err
		}
	}

	//Success
	return nil
}

func (r *ReconcileWorkshop) addGogs(instance *openshiftv1alpha1.Workshop) error {
	reqLogger := log.WithName("Gogs")

	gogsCustomResourceDefinition := deployment.NewCustomResourceDefinition(instance, "gogs.gpte.opentlc.com", "gpte.opentlc.com", "Gogs", "GogsList", "gogs", "gogs", "v1alpha1", nil, nil)
	if err := r.client.Create(context.TODO(), gogsCustomResourceDefinition); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Gogs Custom Resource Definition")
	}

	gogsServiceAccount := deployment.NewServiceAccount(instance, "gogs-operator", instance.Namespace)
	if err := r.client.Create(context.TODO(), gogsServiceAccount); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Gogs Service Account")
	}

	gogsClusterRole := deployment.NewClusterRole(instance, "gogs-operator", instance.Namespace, deployment.GogsRules())
	if err := r.client.Create(context.TODO(), gogsClusterRole); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Gogs ClusterRole")
	}

	gogsClusterRoleBinding := deployment.NewClusterRoleBindingForServiceAccount(instance, "gogs-operator", instance.Namespace, "gogs-operator", "gogs-operator", "ClusterRole")
	if err := r.client.Create(context.TODO(), gogsClusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Gogs Cluster Role Binding")
	}

	gogsOperator := deployment.NewOperatorDeployment(instance, "gogs-operator", instance.Namespace, "quay.io/wkulhanek/gogs-operator:v0.0.6", "gogs-operator", 60000, nil, nil, nil, nil)
	if err := r.client.Create(context.TODO(), gogsOperator); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Gogs Operator")
	}

	gogsCustomResource := gogscustomresource.NewGogsCustomResource(instance, "gogs-server", instance.Namespace)
	if err := r.client.Create(context.TODO(), gogsCustomResource); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Gogs Custom Resource")
	}

	//Success
	return nil
}
