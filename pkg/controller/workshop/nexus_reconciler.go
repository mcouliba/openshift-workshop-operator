package workshop

import (
	"context"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	nexus "github.com/redhat/openshift-workshop-operator/pkg/deployment/nexus"
	"k8s.io/apimachinery/pkg/api/errors"
)

// Reconciling Nexus
func (r *ReconcileWorkshop) reconcileNexus(instance *openshiftv1alpha1.Workshop) error {
	enabledNexus := instance.Spec.Nexus.Enabled

	if enabledNexus {
		if err := r.addNexus(instance); err != nil {
			return err
		}
	}

	//Success
	return nil
}

func (r *ReconcileWorkshop) addNexus(instance *openshiftv1alpha1.Workshop) error {
	reqLogger := log.WithName("Nexus")

	nexusNamespace := deployment.NewNamespace(instance, "opentlc-shared")
	if err := r.client.Create(context.TODO(), nexusNamespace); err != nil && !errors.IsAlreadyExists(err) {
		reqLogger.Error(err, "Failed to create Namespace", "Resource.name", nexusNamespace.Name)
		return err
	} else if err == nil {
		reqLogger.Info("Created Nexus Project")
	}

	nexusCustomResourceDefinition := deployment.NewCustomResourceDefinition(instance, "nexus.gpte.opentlc.com", "gpte.opentlc.com", "Nexus", "NexusList", "nexus", "nexus", "v1alpha1", nil, nil)
	if err := r.client.Create(context.TODO(), nexusCustomResourceDefinition); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created  Nexus Custom Resource Definition")
	}

	nexusServiceAccount := deployment.NewServiceAccount(instance, "nexus-operator", nexusNamespace.Name)
	if err := r.client.Create(context.TODO(), nexusServiceAccount); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created  Nexus Service Account")
	}

	nexusClusterRole := deployment.NewClusterRole(instance, "nexus-operator", nexusNamespace.Name, nexus.NewRules())
	if err := r.client.Create(context.TODO(), nexusClusterRole); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created  Nexus Cluster Role")
	}

	nexusClusterRoleBinding := deployment.NewClusterRoleBindingForServiceAccount(instance, "nexus-operator", nexusNamespace.Name, "nexus-operator", "nexus-operator", "ClusterRole")
	if err := r.client.Create(context.TODO(), nexusClusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Nexus Cluster Role Binding")
	}

	nexusOperator := deployment.NewAnsibleOperatorDeployment(instance, "nexus-operator", nexusNamespace.Name, "quay.io/mcouliba/nexus-operator:v0.10", "nexus-operator")
	if err := r.client.Create(context.TODO(), nexusOperator); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Nexus Operator")
	}

	nexusCustomResource := nexus.NewCustomResource(instance, "nexus", nexusNamespace.Name)
	if err := r.client.Create(context.TODO(), nexusCustomResource); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Nexus Custom Resource")
	}

	//Success
	return nil
}
