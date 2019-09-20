package workshop

import (
	"context"
	"fmt"
	"time"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	smcp "github.com/redhat/openshift-workshop-operator/pkg/deployment/maistra/servicemeshcontrolplane"
	smmr "github.com/redhat/openshift-workshop-operator/pkg/deployment/maistra/servicemeshmemberroll"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciling ServiceMesh
func (r *ReconcileWorkshop) reconcileServiceMesh(instance *openshiftv1alpha1.Workshop, users int) (reconcile.Result, error) {
	enabledServiceMesh := instance.Spec.ServiceMesh.Enabled

	if enabledServiceMesh {
		if result, err := r.addServiceMesh(instance, users); err != nil {
			return result, err
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addServiceMesh(instance *openshiftv1alpha1.Workshop, users int) (reconcile.Result, error) {
	reqLogger := log.WithName("ServiceMesh")

	// ElasticSearch from OperatorHub
	serviceMeshCatalogSourceConfig := deployment.NewCatalogSourceConfig(instance, "installed-service-mesh",
		"openshift-operators", "elasticsearch-operator,jaeger-product,kiali-ossm,servicemeshoperator")
	if err := r.client.Create(context.TODO(), serviceMeshCatalogSourceConfig); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s CatalogSourceConfig", serviceMeshCatalogSourceConfig.Name)
	}

	elasticSubscription := deployment.NewSubscription(instance, "elasticsearch-operator", "openshift-operators", serviceMeshCatalogSourceConfig.Name,
		"preview", "elasticsearch-operator.4.1.15-201909041605")
	if err := r.client.Create(context.TODO(), elasticSubscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Subscription", elasticSubscription.Name)
	}

	jaegerSubscription := deployment.NewSubscription(instance, "jaeger-product", "openshift-operators", serviceMeshCatalogSourceConfig.Name,
		"stable", "jaeger-operator.v1.13.1")
	if err := r.client.Create(context.TODO(), jaegerSubscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Subscription", jaegerSubscription.Name)
	}

	kialiSubscription := deployment.NewSubscription(instance, "kiali-ossm", "openshift-operators", serviceMeshCatalogSourceConfig.Name,
		"stable", "kiali-operator.v1.0.5")
	if err := r.client.Create(context.TODO(), kialiSubscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Subscription", kialiSubscription.Name)
	}

	servicemeshSubscription := deployment.NewSubscription(instance, "servicemeshoperator", "openshift-operators", serviceMeshCatalogSourceConfig.Name,
		"1.0", "servicemeshoperator.v1.0.0")
	if err := r.client.Create(context.TODO(), servicemeshSubscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Subscription", servicemeshSubscription.Name)
	}

	// ISTIO-SYSTEM
	istioSystemNamespace := deployment.NewNamespace(instance, "istio-system")
	if err := r.client.Create(context.TODO(), istioSystemNamespace); err != nil && !errors.IsAlreadyExists(err) {
		reqLogger.Error(err, "Failed to create Namespace", "Resource.name", istioSystemNamespace.Name)
		return reconcile.Result{}, err
	} else if err == nil {
		reqLogger.Info("Created Istio Project (istio-system)")
	}

	serviceMeshControlPlaneCR := smcp.NewServiceMeshControlPlaneCR(smcp.NewServiceMeshControlPlaneCRParameters{
		Name:      "full-install",
		Namespace: istioSystemNamespace.Name,
	})
	if err := r.client.Create(context.TODO(), serviceMeshControlPlaneCR); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 30}, err
	} else if err == nil {
		reqLogger.Info("Created ServiceMeshControlPlane Custom Resource")
	}

	istioMembers := []string{}
	for id := 1; id <= users; id++ {
		istioMembers = append(istioMembers, fmt.Sprintf("coolstore%d", id))
	}

	serviceMeshMemberRollCR := smmr.NewServiceMeshMemberRollCR(smmr.NewServiceMeshMemberRollCRParameters{
		Name:      "default",
		Namespace: istioSystemNamespace.Name,
		Members:   istioMembers,
	})
	if err := r.client.Create(context.TODO(), serviceMeshMemberRollCR); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 30}, err
	} else if err == nil {
		reqLogger.Info("Created ServiceMeshControlPlane Custom Resource")
	}

	//Success
	return reconcile.Result{}, nil
}
