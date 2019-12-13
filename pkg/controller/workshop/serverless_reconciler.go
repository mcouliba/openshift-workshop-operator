package workshop

import (
	"context"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	"github.com/redhat/openshift-workshop-operator/pkg/util"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciling Serverless
func (r *ReconcileWorkshop) reconcileServerless(instance *openshiftv1alpha1.Workshop) (reconcile.Result, error) {
	enabledServerless := instance.Spec.Infrastructure.Serverless.Enabled

	if enabledServerless {

		if result, err := r.addServerless(instance); err != nil {
			return result, err
		}

		// Installed
		if instance.Status.Serverless != util.OperatorStatus.Installed {
			instance.Status.Serverless = util.OperatorStatus.Installed
			if err := r.client.Status().Update(context.TODO(), instance); err != nil {
				logrus.Errorf("Failed to update Workshop status: %s", err)
				return reconcile.Result{}, nil
			}
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addServerless(instance *openshiftv1alpha1.Workshop) (reconcile.Result, error) {

	serverlessSubscription := deployment.NewRedHatSubscription(instance, "serverless-operator", "openshift-operators", "serverless-operator",
		instance.Spec.Infrastructure.Serverless.OperatorHub.Channel,
		instance.Spec.Infrastructure.Serverless.OperatorHub.ClusterServiceVersion)
	if err := r.client.Create(context.TODO(), serverlessSubscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Subscription", serverlessSubscription.Name)
	}

	knativeServingNamespace := deployment.NewNamespace(instance, "knative-serving")
	if err := r.client.Create(context.TODO(), knativeServingNamespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Namespace", knativeServingNamespace.Name)
	}

	// TODO
	// Add  knativeServingNamespace to ServiceMeshMember
	// Create CR

	//Success
	return reconcile.Result{}, nil
}
