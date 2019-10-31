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
	enabledServiceMesh := instance.Spec.Infrastructure.ServiceMesh.Enabled

	if enabledServiceMesh {
		if result, err := r.addServiceMesh(instance, users); err != nil {
			return result, err
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addServiceMesh(instance *openshiftv1alpha1.Workshop, users int) (reconcile.Result, error) {
	// ElasticSearch from OperatorHub
	// serviceMeshCatalogSourceConfig := deployment.NewCatalogSourceConfig(instance, "installed-service-mesh",
	// 	"openshift-operators", "elasticsearch-operator,jaeger-product,kiali-ossm,servicemeshoperator")
	// if err := r.client.Create(context.TODO(), serviceMeshCatalogSourceConfig); err != nil && !errors.IsAlreadyExists(err) {
	// 	return reconcile.Result{}, err
	// } else if err == nil {
	// 	logrus.Infof("Created %s CatalogSourceConfig", serviceMeshCatalogSourceConfig.Name)
	// }

	// elasticSubscription := deployment.NewCertifiedSubscription(instance, "elasticsearch-operator", "openshift-operators",
	// 	instance.Spec.Infrastructure.ServiceMesh.ElasticSearchOperatorHub.Channel,
	// 	instance.Spec.Infrastructure.ServiceMesh.ElasticSearchOperatorHub.ClusterServiceVersion)
	// if err := r.client.Create(context.TODO(), elasticSubscription); err != nil && !errors.IsAlreadyExists(err) {
	// 	return reconcile.Result{}, err
	// } else if err == nil {
	// 	logrus.Infof("Created %s Subscription", elasticSubscription.Name)
	// }

	// jaegerSubscription := deployment.NewCommunitySubscription(instance, "jaeger-workshop", "openshift-operators", "jaeger",
	// 	instance.Spec.Infrastructure.ServiceMesh.JaegerOperatorHub.Channel,
	// 	instance.Spec.Infrastructure.ServiceMesh.JaegerOperatorHub.ClusterServiceVersion)
	// if err := r.client.Create(context.TODO(), jaegerSubscription); err != nil && !errors.IsAlreadyExists(err) {
	// 	return reconcile.Result{}, err
	// } else if err == nil {
	// 	logrus.Infof("Created %s Subscription", jaegerSubscription.Name)
	// }

	// kialiSubscription := deployment.NewCommunitySubscription(instance, "kiali-workshop", "openshift-operators", "kiali",
	// 	instance.Spec.Infrastructure.ServiceMesh.KialiOperatorHub.Channel,
	// 	instance.Spec.Infrastructure.ServiceMesh.KialiOperatorHub.ClusterServiceVersion)
	// if err := r.client.Create(context.TODO(), kialiSubscription); err != nil && !errors.IsAlreadyExists(err) {
	// 	return reconcile.Result{}, err
	// } else if err == nil {
	// 	logrus.Infof("Created %s Subscription", kialiSubscription.Name)
	// }

	servicemeshSubscription := deployment.NewRedHatSubscription(instance, "servicemeshoperator", "openshift-operators", "servicemeshoperator",
		instance.Spec.Infrastructure.ServiceMesh.ServiceMeshOperatorHub.Channel,
		instance.Spec.Infrastructure.ServiceMesh.ServiceMeshOperatorHub.ClusterServiceVersion)
	if err := r.client.Create(context.TODO(), servicemeshSubscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Subscription", servicemeshSubscription.Name)
	}

	// ISTIO-SYSTEM
	istioSystemNamespace := deployment.NewNamespace(instance, "istio-system")
	if err := r.client.Create(context.TODO(), istioSystemNamespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Namespace", istioSystemNamespace.Name)
	}

	serviceMeshControlPlaneCR := smcp.NewServiceMeshControlPlaneCR(smcp.NewServiceMeshControlPlaneCRParameters{
		Name:      "full-install",
		Namespace: istioSystemNamespace.Name,
	})
	if err := r.client.Create(context.TODO(), serviceMeshControlPlaneCR); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 30}, err
	} else if err == nil {
		logrus.Infof("Created %s Custom Resource", serviceMeshControlPlaneCR.Name)
	}

	istioMembers := []string{}
	for id := 1; id <= users; id++ {
		username := fmt.Sprintf("user%d", id)
		projectName := fmt.Sprintf("%s%d", instance.Spec.Infrastructure.Project.Name, id)

		istioMembers = append(istioMembers, projectName)

		jaegerRole := deployment.NewRole(deployment.NewRoleParameters{
			Name:      username + "-jaeger",
			Namespace: "istio-system",
			Rules:     deployment.JaegerUserRules(),
		})
		if err := r.client.Create(context.TODO(), jaegerRole); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			logrus.Infof("Created %s Role", jaegerRole.Name)
		}

		JaegerRoleBinding := deployment.NewRoleBindingUser(deployment.NewRoleBindingUserParameters{
			Name:      username + "-jaeger",
			Namespace: "istio-system",
			Username:  username,
			RoleName:  jaegerRole.Name,
			RoleKind:  "Role",
		})
		if err := r.client.Create(context.TODO(), JaegerRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			logrus.Infof("Created %s Role Binding", JaegerRoleBinding.Name)
		}
	}

	serviceMeshMemberRollCR := smmr.NewServiceMeshMemberRollCR(smmr.NewServiceMeshMemberRollCRParameters{
		Name:      "default",
		Namespace: istioSystemNamespace.Name,
		Members:   istioMembers,
	})
	if err := r.client.Create(context.TODO(), serviceMeshMemberRollCR); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 30}, err
	} else if err == nil {
		logrus.Infof("Created %s Custom Resource", serviceMeshMemberRollCR.Name)
	}

	//Success
	return reconcile.Result{}, nil
}
