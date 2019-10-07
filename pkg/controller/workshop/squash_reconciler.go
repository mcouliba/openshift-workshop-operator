package workshop

import (
	"context"

	securityv1 "github.com/openshift/api/security/v1"
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	"github.com/redhat/openshift-workshop-operator/pkg/deployment/squash"
	"github.com/redhat/openshift-workshop-operator/pkg/util"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciling Squash
func (r *ReconcileWorkshop) reconcileSquash(instance *openshiftv1alpha1.Workshop, users int) (reconcile.Result, error) {
	enabledSquash := instance.Spec.Infrastructure.Squash.Enabled

	if enabledSquash {
		if result, err := r.addSquash(instance); err != nil {
			return result, err
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addSquash(instance *openshiftv1alpha1.Workshop) (reconcile.Result, error) {
	squashNamespace := deployment.NewNamespace(instance, "squash-debugger")
	if err := r.client.Create(context.TODO(), squashNamespace); err != nil && !errors.IsAlreadyExists(err) {
		logrus.Errorf("Failed to created %s Project: %v", squashNamespace.Name, err)
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Project", squashNamespace.Name)
	}

	squashCustomResourceDefinition := deployment.NewCustomResourceDefinition(instance, "debugattachments.squash.solo.io", "squash.solo.io", "DebugAttachment", "DebugAttachmentList", "debugattachments", "debugattachment", "v1", []string{"debatt"}, nil)
	if err := r.client.Create(context.TODO(), squashCustomResourceDefinition); err != nil && !errors.IsAlreadyExists(err) {
		logrus.Errorf("Failed to created %s Custom Resource Definition: %v", squashCustomResourceDefinition.Name, err)
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Custom Resource Definition", squashCustomResourceDefinition.Name)
	}

	squashServiceAccount := deployment.NewServiceAccount(instance, "squash", squashNamespace.Name)
	if err := r.client.Create(context.TODO(), squashServiceAccount); err != nil && !errors.IsAlreadyExists(err) {
		logrus.Errorf("Failed to created %s Service Account: %v", squashServiceAccount.Name, err)
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Service Account", squashServiceAccount.Name)
	}

	squashClusterRole := deployment.NewClusterRole(instance, "squash-cr-pods", squashNamespace.Name, squash.NewRules())
	if err := r.client.Create(context.TODO(), squashClusterRole); err != nil && !errors.IsAlreadyExists(err) {
		logrus.Errorf("Failed to created %s Cluster Role: %v", squashClusterRole.Name, err)
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Cluster Role", squashClusterRole.Name)
	}

	squashClusterRoleBinding := deployment.NewClusterRoleBindingForServiceAccount(instance, "squash-crb-pods", squashNamespace.Name, "squash", "squash-cr-pods", "ClusterRole")
	if err := r.client.Create(context.TODO(), squashClusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		logrus.Errorf("Failed to created %s Cluster Role Binding: %v", squashClusterRoleBinding.Name, err)
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Cluster Role Binding", squashClusterRoleBinding.Name)
	}

	serviceaccount := "system:serviceaccount:" + squashNamespace.Name + ":" + squashServiceAccount.Name
	privilegedSCCFound := &securityv1.SecurityContextConstraints{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: "privileged"}, privilegedSCCFound); err != nil {
		logrus.Errorf("Failed to get Privileged SCC: %v", err)
		return reconcile.Result{}, err
	}

	if !util.StringInSlice(serviceaccount, privilegedSCCFound.Users) {
		privilegedSCCFound.Users = append(privilegedSCCFound.Users, serviceaccount)
		if err := r.client.Update(context.TODO(), privilegedSCCFound); err != nil {
			logrus.Errorf("Failed to update the Privileged SCC: %v", err)
			return reconcile.Result{}, err
		} else if err == nil {
			logrus.Infof("Updated the Privileged SCC")
		}
	}

	squashDeployment := squash.NewDeployment(instance, "squash", squashNamespace.Name)
	if err := r.client.Create(context.TODO(), squashDeployment); err != nil && !errors.IsAlreadyExists(err) {
		logrus.Errorf("Failed to created %s Deployment: %v", squashDeployment.Name, err)
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Deployment", squashDeployment.Name)
	}

	//Success
	return reconcile.Result{}, nil
}
