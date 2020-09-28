package workshop

import (
	"context"
	"fmt"
	"time"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	smcp "github.com/redhat/openshift-workshop-operator/pkg/deployment/maistra/servicemeshcontrolplane"
	smmr "github.com/redhat/openshift-workshop-operator/pkg/deployment/maistra/servicemeshmemberroll"
	"github.com/redhat/openshift-workshop-operator/pkg/util"
	"github.com/sirupsen/logrus"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciling ServiceMesh
func (r *ReconcileWorkshop) reconcileServiceMesh(instance *openshiftv1alpha1.Workshop, users int) (reconcile.Result, error) {
	enabledServiceMesh := instance.Spec.Infrastructure.ServiceMesh.Enabled
	enabledServerless := instance.Spec.Infrastructure.Serverless.Enabled

	if enabledServiceMesh || enabledServerless {

		if result, err := r.addElasticSearchOperator(instance); err != nil {
			return result, err
		}

		if result, err := r.addJaegerOperator(instance); err != nil {
			return result, err
		}

		if result, err := r.addKialiOperator(instance); err != nil {
			return result, err
		}

		if result, err := r.addServiceMesh(instance, users); err != nil {
			return result, err
		}

		// Installed
		if instance.Status.ServiceMesh != util.OperatorStatus.Installed {
			instance.Status.ServiceMesh = util.OperatorStatus.Installed
			if err := r.client.Status().Update(context.TODO(), instance); err != nil {
				logrus.Errorf("Failed to update Workshop status: %s", err)
				return reconcile.Result{}, nil
			}
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addServiceMesh(instance *openshiftv1alpha1.Workshop, users int) (reconcile.Result, error) {

	// Service Mesh Operator
	channel := instance.Spec.Infrastructure.ServiceMesh.ServiceMeshOperatorHub.Channel
	clusterserviceversion := instance.Spec.Infrastructure.ServiceMesh.ServiceMeshOperatorHub.ClusterServiceVersion

	subscription := deployment.NewRedHatSubscription(instance, "servicemeshoperator", "openshift-operators",
		"servicemeshoperator", channel, clusterserviceversion)
	if err := r.client.Create(context.TODO(), subscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Subscription", subscription.Name)
	}

	if err := r.ApproveInstallPlan(clusterserviceversion, "servicemeshoperator", "openshift-operators"); err != nil {
		logrus.Infof("Waiting for Subscription to create InstallPlan for %s", subscription.Name)
		return reconcile.Result{}, err
	}

	// Deploy Service Mesh
	istioSystemNamespace := deployment.NewNamespace(instance, "istio-system")
	if err := r.client.Create(context.TODO(), istioSystemNamespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Namespace", istioSystemNamespace.Name)
	}

	istioMembers := []string{}
	istioUsers := []rbac.Subject{}

	if instance.Spec.Infrastructure.ArgoCD.Enabled {
		argocdSubject := rbac.Subject{
			Kind: rbac.UserKind,
			Name: "system:serviceaccount:argocd:argocd-application-controller",
		}
		istioUsers = append(istioUsers, argocdSubject)
	}

	for id := 1; id <= users; id++ {
		username := fmt.Sprintf("user%d", id)
		stagingProjectName := fmt.Sprintf("%s%d", instance.Spec.Infrastructure.Project.StagingName, id)
		userSubject := rbac.Subject{
			Kind: rbac.UserKind,
			Name: username,
		}

		istioMembers = append(istioMembers, stagingProjectName)
		istioUsers = append(istioUsers, userSubject)

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

	meshUserRoleBinding := deployment.NewRoleBindingUsers(deployment.NewRoleBindingUsersParameters{
		Name:      "mesh-users",
		Namespace: "istio-system",
		Subject:   istioUsers,
		RoleName:  "mesh-user",
		RoleKind:  "Role",
	})

	if err := r.client.Create(context.TODO(), meshUserRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Role Binding", meshUserRoleBinding.Name)
	}

	serviceMeshControlPlaneCR := smcp.NewServiceMeshControlPlaneCR(smcp.NewServiceMeshControlPlaneCRParameters{
		Name:      "full-install",
		Namespace: istioSystemNamespace.Name,
	})
	if err := r.client.Create(context.TODO(), serviceMeshControlPlaneCR); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 1}, err
	} else if err == nil {
		logrus.Infof("Created %s Custom Resource", serviceMeshControlPlaneCR.Name)
	}

	serviceMeshMemberRollCR := smmr.NewServiceMeshMemberRollCR(smmr.NewServiceMeshMemberRollCRParameters{
		Name:      "default",
		Namespace: istioSystemNamespace.Name,
		Members:   istioMembers,
	})
	if err := r.client.Create(context.TODO(), serviceMeshMemberRollCR); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Custom Resource", serviceMeshMemberRollCR.Name)
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addElasticSearchOperator(instance *openshiftv1alpha1.Workshop) (reconcile.Result, error) {

	channel := instance.Spec.Infrastructure.ServiceMesh.ElasticSearchOperatorHub.Channel
	clusterserviceversion := instance.Spec.Infrastructure.ServiceMesh.ElasticSearchOperatorHub.ClusterServiceVersion

	redhatOperatorsNamespace := deployment.NewNamespace(instance, "openshift-operators-redhat")
	if err := r.client.Create(context.TODO(), redhatOperatorsNamespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Namespace", redhatOperatorsNamespace.Name)
	}

	subscription := deployment.NewRedHatSubscription(instance, "elasticsearch-operator", "openshift-operators-redhat",
		"elasticsearch-operator", channel, clusterserviceversion)
	if err := r.client.Create(context.TODO(), subscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Subscription", subscription.Name)
	}

	if err := r.ApproveInstallPlan(clusterserviceversion, "elasticsearch-operator", "openshift-operators-redhat"); err != nil {
		logrus.Infof("Waiting for Subscription to create InstallPlan for %s", subscription.Name)
		return reconcile.Result{}, err
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addJaegerOperator(instance *openshiftv1alpha1.Workshop) (reconcile.Result, error) {

	channel := instance.Spec.Infrastructure.ServiceMesh.JaegerOperatorHub.Channel
	clusterserviceversion := instance.Spec.Infrastructure.ServiceMesh.JaegerOperatorHub.ClusterServiceVersion

	subscription := deployment.NewRedHatSubscription(instance, "jaeger-product", "openshift-operators",
		"jaeger-product", channel, clusterserviceversion)
	if err := r.client.Create(context.TODO(), subscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Subscription", subscription.Name)
	}

	if err := r.ApproveInstallPlan(clusterserviceversion, "jaeger-product", "openshift-operators"); err != nil {
		logrus.Infof("Waiting for Subscription to create InstallPlan for %s", subscription.Name)
		return reconcile.Result{}, err
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addKialiOperator(instance *openshiftv1alpha1.Workshop) (reconcile.Result, error) {

	channel := instance.Spec.Infrastructure.ServiceMesh.KialiOperatorHub.Channel
	clusterserviceversion := instance.Spec.Infrastructure.ServiceMesh.KialiOperatorHub.ClusterServiceVersion

	subscription := deployment.NewRedHatSubscription(instance, "kiali-ossm", "openshift-operators",
		"kiali-ossm", channel, clusterserviceversion)
	if err := r.client.Create(context.TODO(), subscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Subscription", subscription.Name)
	}

	if err := r.ApproveInstallPlan(clusterserviceversion, "kiali-ossm", "openshift-operators"); err != nil {
		logrus.Infof("Waiting for Subscription to create InstallPlan for %s", subscription.Name)
		return reconcile.Result{}, err
	}

	//Success
	return reconcile.Result{}, nil
}
