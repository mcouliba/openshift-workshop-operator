package workshop

import (
	"context"
	"time"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	certmanager "github.com/redhat/openshift-workshop-operator/pkg/deployment/certmanager"
	"github.com/redhat/openshift-workshop-operator/pkg/util"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciling CertManager
func (r *ReconcileWorkshop) reconcileCertManager(instance *openshiftv1alpha1.Workshop, users int) (reconcile.Result, error) {
	enabledCertManager := instance.Spec.Infrastructure.CertManager.Enabled

	if enabledCertManager {

		if result, err := r.addCertManager(instance, users); err != nil {
			return result, err
		}

		// Installed
		if instance.Status.CertManager != util.OperatorStatus.Installed {
			instance.Status.CertManager = util.OperatorStatus.Installed
			if err := r.client.Status().Update(context.TODO(), instance); err != nil {
				logrus.Errorf("Failed to update Workshop status: %s", err)
				return reconcile.Result{}, nil
			}
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addCertManager(instance *openshiftv1alpha1.Workshop, users int) (reconcile.Result, error) {

	channel := instance.Spec.Infrastructure.CertManager.OperatorHub.Channel
	clusterServiceVersion := instance.Spec.Infrastructure.CertManager.OperatorHub.ClusterServiceVersion

	CertManagerSubscription := deployment.NewCertifiedSubscription(instance, "cert-manager-operator", "openshift-operators",
		"cert-manager-operator", channel, clusterServiceVersion)
	if err := r.client.Create(context.TODO(), CertManagerSubscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Subscription", CertManagerSubscription.Name)
	}

	// Approve the installation
	if err := r.ApproveInstallPlan(clusterServiceVersion, "cert-manager-operator", "openshift-operators"); err != nil {
		logrus.Infof("Waiting for Subscription to create InstallPlan for %s", "CertManageroperator")
		return reconcile.Result{}, err
	}

	namespace := deployment.NewNamespace(instance, "cert-manager")
	if err := r.client.Create(context.TODO(), namespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Namespace", namespace.Name)
	}

	customresource := certmanager.NewCertManagerCustomResource("cert-manager", namespace.Name)
	if err := r.client.Create(context.TODO(), customresource); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 1}, err
	} else if err == nil {
		logrus.Infof("Created %s Custom Resource", customresource.Name)
	}

	//Success
	return reconcile.Result{}, nil
}
