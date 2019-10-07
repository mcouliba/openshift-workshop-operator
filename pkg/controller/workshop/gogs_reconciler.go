package workshop

import (
	"context"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	gogscustomresource "github.com/redhat/openshift-workshop-operator/pkg/customresource/gogs"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciling Gogs
func (r *ReconcileWorkshop) reconcileGogs(instance *openshiftv1alpha1.Workshop) (reconcile.Result, error) {
	enabledGogs := instance.Spec.Infrastructure.Gogs.Enabled

	if enabledGogs {
		if result, err := r.addGogs(instance); err != nil {
			return result, err
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addGogs(instance *openshiftv1alpha1.Workshop) (reconcile.Result, error) {

	imageName := instance.Spec.Infrastructure.Gogs.Image.Name
	imageTag := instance.Spec.Infrastructure.Gogs.Image.Tag

	gogsCustomResourceDefinition := deployment.NewCustomResourceDefinition(instance, "gogs.gpte.opentlc.com", "gpte.opentlc.com", "Gogs", "GogsList", "gogs", "gogs", "v1alpha1", nil, nil)
	if err := r.client.Create(context.TODO(), gogsCustomResourceDefinition); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Custom Resource Definition", gogsCustomResourceDefinition.Name)
	}

	gogsServiceAccount := deployment.NewServiceAccount(instance, "gogs-operator", instance.Namespace)
	if err := r.client.Create(context.TODO(), gogsServiceAccount); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Service Account", gogsServiceAccount.Name)
	}

	gogsClusterRole := deployment.NewClusterRole(instance, "gogs-operator", instance.Namespace, deployment.GogsRules())
	if err := r.client.Create(context.TODO(), gogsClusterRole); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Cluster Role", gogsClusterRole.Name)
	}

	gogsClusterRoleBinding := deployment.NewClusterRoleBindingForServiceAccount(instance, "gogs-operator", instance.Namespace, "gogs-operator", "gogs-operator", "ClusterRole")
	if err := r.client.Create(context.TODO(), gogsClusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Cluster Role Binding", gogsClusterRoleBinding.Name)
	}

	gogsOperator := deployment.NewOperatorDeployment(instance, "gogs-operator", instance.Namespace, imageName+":"+imageTag, "gogs-operator", 60000, nil, nil, nil, nil)
	if err := r.client.Create(context.TODO(), gogsOperator); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Operator", gogsOperator.Name)
	}

	gogsCustomResource := gogscustomresource.NewGogsCustomResource(instance, "gogs-server", instance.Namespace)
	if err := r.client.Create(context.TODO(), gogsCustomResource); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Custom Resource", gogsCustomResource.Name)
	}

	//Success
	return reconcile.Result{}, nil
}
