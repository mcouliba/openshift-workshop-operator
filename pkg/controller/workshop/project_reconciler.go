package workshop

import (
	"context"
	"fmt"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	"github.com/redhat/openshift-workshop-operator/pkg/util"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciling Project
func (r *ReconcileWorkshop) reconcileProject(instance *openshiftv1alpha1.Workshop, users int) (reconcile.Result, error) {
	enabledProject := instance.Spec.Infrastructure.Project.Enabled

	id := 1
	for {
		username := fmt.Sprintf("user%d", id)
		stagingProjectName := fmt.Sprintf("%s%d", instance.Spec.Infrastructure.Project.StagingName, id)

		if id <= users && enabledProject {
			// Project
			if instance.Spec.Infrastructure.Project.StagingName != "" {
				if result, err := r.addProject(instance, stagingProjectName, username); err != nil {
					return result, err
				}
			}

		} else {
			stagingProjectNamespace := deployment.NewNamespace(instance, stagingProjectName)
			stagingProjectNamespaceFound := &corev1.Namespace{}
			stagingProjectNamespaceErr := r.client.Get(context.TODO(), types.NamespacedName{Name: stagingProjectNamespace.Name}, stagingProjectNamespaceFound)

			if stagingProjectNamespaceErr != nil && errors.IsNotFound(stagingProjectNamespaceErr) {
				break
			}

			if !(stagingProjectNamespaceErr != nil && errors.IsNotFound(stagingProjectNamespaceErr)) {
				if result, err := r.deleteProject(stagingProjectNamespace); err != nil {
					return result, err
				}
			}
		}

		id++
	}

	// Installed
	if instance.Status.Project != util.OperatorStatus.Installed {
		instance.Status.Project = util.OperatorStatus.Installed
		if err := r.client.Status().Update(context.TODO(), instance); err != nil {
			logrus.Errorf("Failed to update Workshop status: %s", err)
			return reconcile.Result{}, err
		}
	}
	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addProject(instance *openshiftv1alpha1.Workshop, projectName string, username string) (reconcile.Result, error) {

	projectNamespace := deployment.NewNamespace(instance, projectName)
	if err := r.client.Create(context.TODO(), projectNamespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Namespace", projectNamespace.Name)
	}

	if result, err := r.manageRoles(projectNamespace.Name, username); err != nil {
		return result, err
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) deleteProject(namespaces *corev1.Namespace) (reconcile.Result, error) {

	if err := r.client.Delete(context.TODO(), namespaces); err != nil && !errors.IsNotFound(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Deleted %s Namespace", namespaces.Name)
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) manageRoles(projectName string, username string) (reconcile.Result, error) {

	// Istio
	istioRole := deployment.NewRole(deployment.NewRoleParameters{
		Name:      username + "-istio",
		Namespace: projectName,
		Rules:     deployment.IstioUserRules(),
	})
	if err := r.client.Create(context.TODO(), istioRole); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Role", istioRole.Name)
	}

	istioRoleBinding := deployment.NewRoleBindingUser(deployment.NewRoleBindingUserParameters{
		Name:      username + "-istio",
		Namespace: projectName,
		Username:  username,
		RoleName:  istioRole.Name,
		RoleKind:  "Role",
	})
	if err := r.client.Create(context.TODO(), istioRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Role Binding", istioRole.Name)
	}

	istioArgocdRole := deployment.NewRole(deployment.NewRoleParameters{
		Name:      username + "-istio-argocd",
		Namespace: projectName,
		Rules:     deployment.IstioArgoCDRules(),
	})
	if err := r.client.Create(context.TODO(), istioArgocdRole); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Role", istioArgocdRole.Name)
	}

	istioArgocdRoleBinding := deployment.NewRoleBindingUser(deployment.NewRoleBindingUserParameters{
		Name:      username + "-istio-argocd",
		Namespace: projectName,
		Username:  "system:serviceaccount:argocd:argocd-application-controller",
		RoleName:  istioArgocdRole.Name,
		RoleKind:  "Role",
	})
	if err := r.client.Create(context.TODO(), istioArgocdRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Role Binding", istioArgocdRoleBinding.Name)
	}

	// User
	userRoleBinding := deployment.NewRoleBindingUser(deployment.NewRoleBindingUserParameters{
		Name:      username + "-admin",
		Namespace: projectName,
		Username:  username,
		RoleName:  "admin",
		RoleKind:  "ClusterRole",
	})
	if err := r.client.Create(context.TODO(), userRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Role Binding", userRoleBinding.Name)
	}

	// Default
	defaultRoleBinding := deployment.NewRoleBindingSA(deployment.NewRoleBindingSAParameters{
		Name:               "view",
		Namespace:          projectName,
		ServiceAccountName: "default",
		RoleName:           "view",
		RoleKind:           "ClusterRole",
	})
	if err := r.client.Create(context.TODO(), defaultRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Role Binding", defaultRoleBinding.Name)
	}

	//Argo CD
	argocdEditRoleBinding := deployment.NewRoleBindingUser(deployment.NewRoleBindingUserParameters{
		Name:      username + "-argocd",
		Namespace: projectName,
		Username:  "system:serviceaccount:argocd:argocd-application-controller",
		RoleName:  "edit",
		RoleKind:  "ClusterRole",
	})
	if err := r.client.Create(context.TODO(), argocdEditRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Role Binding", argocdEditRoleBinding.Name)
	}

	//Success
	return reconcile.Result{}, nil
}
