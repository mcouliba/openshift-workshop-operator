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

// Reconciling Pipeline
func (r *ReconcileWorkshop) reconcilePipeline(instance *openshiftv1alpha1.Workshop) (reconcile.Result, error) {
	enabledPipeline := instance.Spec.Infrastructure.Pipeline.Enabled

	if enabledPipeline {
		if result, err := r.addPipeline(instance); err != nil {
			return result, err
		}

		// Installed
		if instance.Status.Pipeline != util.OperatorStatus.Installed {
			instance.Status.Pipeline = util.OperatorStatus.Installed
			if err := r.client.Status().Update(context.TODO(), instance); err != nil {
				logrus.Errorf("Failed to update Workshop status: %s", err)
				return reconcile.Result{}, err
			}
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addPipeline(instance *openshiftv1alpha1.Workshop) (reconcile.Result, error) {

	// pipelineCatalogSourceConfig := deployment.NewCatalogSourceConfig(instance, "installed-pipeline",
	// 	"openshift-operators", "openshift-pipelines-operator")
	// if err := r.client.Create(context.TODO(), pipelineCatalogSourceConfig); err != nil && !errors.IsAlreadyExists(err) {
	// 	return reconcile.Result{}, err
	// } else if err == nil {
	// 	logrus.Infof("Created %s CatalogSourceConfig", pipelineCatalogSourceConfig.Name)
	// }

	pipelineSubscription := deployment.NewCommunitySubscription(instance, "openshift-pipelines-operator", "openshift-operators", "openshift-pipelines-operator",
		instance.Spec.Infrastructure.Pipeline.OperatorHub.Channel,
		instance.Spec.Infrastructure.Pipeline.OperatorHub.ClusterServiceVersion)
	if err := r.client.Create(context.TODO(), pipelineSubscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Subscription", pipelineSubscription.Name)
	}

	//Success
	return reconcile.Result{}, nil
}
