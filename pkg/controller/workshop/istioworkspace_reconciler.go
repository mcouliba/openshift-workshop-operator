package workshop

import (
	"context"
	"fmt"

	securityv1 "github.com/openshift/api/security/v1"
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	"github.com/redhat/openshift-workshop-operator/pkg/util"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciling IstioWorkspace
func (r *ReconcileWorkshop) reconcileIstioWorkspace(instance *openshiftv1alpha1.Workshop, users int) (reconcile.Result, error) {
	enabled := instance.Spec.Infrastructure.IstioWorkspace.Enabled

	if enabled {

		if result, err := r.addIstioWorkspace(instance, users); err != nil {
			return result, err
		}

		// Installed
		if instance.Status.IstioWorkspace != util.OperatorStatus.Installed {
			instance.Status.IstioWorkspace = util.OperatorStatus.Installed
			if err := r.client.Status().Update(context.TODO(), instance); err != nil {
				logrus.Errorf("Failed to update Workshop status: %s", err)
				return reconcile.Result{}, err
			}
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addIstioWorkspace(instance *openshiftv1alpha1.Workshop, users int) (reconcile.Result, error) {

	imageName := instance.Spec.Infrastructure.IstioWorkspace.Image.Name
	imageTag := instance.Spec.Infrastructure.IstioWorkspace.Image.Tag

	customResourceDefinition := deployment.NewCustomResourceDefinition(instance, "sessions.maistra.io", "maistra.io", "Session", "SessionList", "sessions", "session", "v1alpha1", nil, nil)
	if err := r.client.Create(context.TODO(), customResourceDefinition); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Custom Resource Definition", customResourceDefinition.Name)
	}

	serviceAccount := deployment.NewServiceAccount(instance, "istio-workspace", instance.Namespace)
	if err := r.client.Create(context.TODO(), serviceAccount); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Service Account", serviceAccount.Name)
	}

	clusterRole := deployment.NewClusterRole(instance, "istio-workspace", instance.Namespace, deployment.IstioWorkspaceRules())
	if err := r.client.Create(context.TODO(), clusterRole); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Cluster Role", clusterRole.Name)
	}

	clusterRoleBinding := deployment.NewClusterRoleBindingForServiceAccount(instance, "istio-workspace", instance.Namespace, "istio-workspace", "istio-workspace", "ClusterRole")
	if err := r.client.Create(context.TODO(), clusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Cluster Role Binding", clusterRoleBinding.Name)
	}

	operator := deployment.NewOperatorDeployment(instance, "istio-workspace", instance.Namespace, imageName+":"+imageTag, "istio-workspace", 8383, []string{"ike"}, []string{"serve"}, nil, nil)
	if err := r.client.Create(context.TODO(), operator); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Operator", operator.Name)
	}

	for id := 1; id <= users; id++ {
		username := fmt.Sprintf("user%d", id)
		stagingProjectName := fmt.Sprintf("%s%d", instance.Spec.Infrastructure.Project.StagingName, id)

		role := deployment.NewRole(deployment.NewRoleParameters{
			Name:      username + "-istio-workspace",
			Namespace: stagingProjectName,
			Rules:     deployment.IstioWorkspaceUserRules(),
		})
		if err := r.client.Create(context.TODO(), role); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			logrus.Infof("Created %s Role", role.Name)
		}

		roleBinding := deployment.NewRoleBindingUser(deployment.NewRoleBindingUserParameters{
			Name:      username + "-istio-workspace",
			Namespace: stagingProjectName,
			Username:  username,
			RoleName:  username + "-istio-workspace",
			RoleKind:  "Role",
		})
		if err := r.client.Create(context.TODO(), roleBinding); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			logrus.Infof("Created %s Role Binding", roleBinding.Name)
		}

		// Create SCC
		serviceAccountUser := "system:serviceaccount:" + stagingProjectName + ":default"

		privilegedSCCFound := &securityv1.SecurityContextConstraints{}
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: "privileged"}, privilegedSCCFound); err != nil {
			return reconcile.Result{}, err
		}

		if !util.StringInSlice(serviceAccountUser, privilegedSCCFound.Users) {
			privilegedSCCFound.Users = append(privilegedSCCFound.Users, serviceAccountUser)
			if err := r.client.Update(context.TODO(), privilegedSCCFound); err != nil {
				return reconcile.Result{}, err
			} else if err == nil {
				logrus.Infof("Updated %s SCC", privilegedSCCFound.Name)
			}
		}

		anyuidSCCFound := &securityv1.SecurityContextConstraints{}
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: "anyuid"}, anyuidSCCFound); err != nil {
			return reconcile.Result{}, err
		}

		if !util.StringInSlice(serviceAccountUser, anyuidSCCFound.Users) {
			anyuidSCCFound.Users = append(anyuidSCCFound.Users, serviceAccountUser)
			if err := r.client.Update(context.TODO(), anyuidSCCFound); err != nil {
				return reconcile.Result{}, err
			} else if err == nil {
				logrus.Infof("Updated %s SCC", anyuidSCCFound.Name)
			}
		}

	}

	//Success
	return reconcile.Result{}, nil
}
