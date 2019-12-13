package workshop

import (
	"context"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	nexus "github.com/redhat/openshift-workshop-operator/pkg/deployment/nexus"
	"github.com/redhat/openshift-workshop-operator/pkg/util"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciling Nexus
func (r *ReconcileWorkshop) reconcileNexus(instance *openshiftv1alpha1.Workshop) (reconcile.Result, error) {
	enabledNexus := instance.Spec.Infrastructure.Nexus.Enabled

	if enabledNexus {

		if result, err := r.addNexus(instance); err != nil {
			return result, err
		}

		// Installed
		if instance.Status.Nexus != util.OperatorStatus.Installed {
			instance.Status.Nexus = util.OperatorStatus.Installed
			if err := r.client.Status().Update(context.TODO(), instance); err != nil {
				logrus.Errorf("Failed to update Workshop status: %s", err)
				return reconcile.Result{}, err
			}
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addNexus(instance *openshiftv1alpha1.Workshop) (reconcile.Result, error) {
	nexusNamespace := deployment.NewNamespace(instance, "opentlc-shared")
	if err := r.client.Create(context.TODO(), nexusNamespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Project", nexusNamespace.Name)
	}

	nexusCustomResourceDefinition := deployment.NewCustomResourceDefinition(instance, "nexus.gpte.opentlc.com", "gpte.opentlc.com", "Nexus", "NexusList", "nexus", "nexus", "v1alpha1", nil, nil)
	if err := r.client.Create(context.TODO(), nexusCustomResourceDefinition); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Custom Resource Definition", nexusCustomResourceDefinition.Name)
	}

	nexusServiceAccount := deployment.NewServiceAccount(instance, "nexus-operator", nexusNamespace.Name)
	if err := r.client.Create(context.TODO(), nexusServiceAccount); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Service Account", nexusServiceAccount.Name)
	}

	nexusClusterRole := deployment.NewClusterRole(instance, "nexus-operator", nexusNamespace.Name, nexus.NewRules())
	if err := r.client.Create(context.TODO(), nexusClusterRole); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Cluster Role", nexusClusterRole.Name)
	}

	nexusClusterRoleBinding := deployment.NewClusterRoleBindingForServiceAccount(instance, "nexus-operator", nexusNamespace.Name, "nexus-operator", "nexus-operator", "ClusterRole")
	if err := r.client.Create(context.TODO(), nexusClusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Cluster Role Binding", nexusClusterRoleBinding.Name)
	}

	nexusOperator := deployment.NewAnsibleOperatorDeployment(instance, "nexus-operator", nexusNamespace.Name, "quay.io/mcouliba/nexus-operator:v0.10", "nexus-operator")
	if err := r.client.Create(context.TODO(), nexusOperator); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Operator", nexusOperator.Name)
	}

	nexusCustomResource := nexus.NewCustomResource(instance, "nexus", nexusNamespace.Name)
	if err := r.client.Create(context.TODO(), nexusCustomResource); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Custom Resource", nexusCustomResource.Name)
	}

	//Success
	return reconcile.Result{}, nil
}
