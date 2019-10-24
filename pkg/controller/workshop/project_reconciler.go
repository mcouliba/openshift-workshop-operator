package workshop

import (
	"context"
	"fmt"

	securityv1 "github.com/openshift/api/security/v1"
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
		projectName := fmt.Sprintf("%s%d", instance.Spec.Infrastructure.Project.Name, id)

		if id <= users && enabledProject {
			// Project
			if result, err := r.addProject(instance, projectName, username); err != nil {
				return result, err
			}

		} else {

			projectNamespace := deployment.NewNamespace(instance, projectName)
			projectNamespaceFound := &corev1.Namespace{}
			projectNamespaceErr := r.client.Get(context.TODO(), types.NamespacedName{Name: projectNamespace.Name}, projectNamespaceFound)

			if projectNamespaceErr != nil && errors.IsNotFound(projectNamespaceErr) {
				break
			}

			if result, err := r.deleteProject(projectNamespace); err != nil {
				return result, err
			}
		}

		id++
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

	istioRole := deployment.NewRole(deployment.NewRoleParameters{
		Name:      username + "-istio",
		Namespace: projectNamespace.Name,
		Rules:     deployment.IstioUserRules(),
	})
	if err := r.client.Create(context.TODO(), istioRole); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Role", istioRole.Name)
	}

	istioRoleBinding := deployment.NewRoleBindingUser(deployment.NewRoleBindingUserParameters{
		Name:      username + "-istio",
		Namespace: projectNamespace.Name,
		Username:  username,
		RoleName:  username + "-istio",
		RoleKind:  "Role",
	})
	if err := r.client.Create(context.TODO(), istioRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Role Binding", istioRole.Name)
	}

	userRoleBinding := deployment.NewRoleBindingUser(deployment.NewRoleBindingUserParameters{
		Name:      username + "-admin",
		Namespace: projectNamespace.Name,
		Username:  username,
		RoleName:  "admin",
		RoleKind:  "ClusterRole",
	})
	if err := r.client.Create(context.TODO(), userRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Role Binding", userRoleBinding.Name)
	}

	defaultRoleBinding := deployment.NewRoleBindingSA(deployment.NewRoleBindingSAParameters{
		Name:               "view",
		Namespace:          projectNamespace.Name,
		ServiceAccountName: "default",
		RoleName:           "view",
		RoleKind:           "ClusterRole",
	})
	if err := r.client.Create(context.TODO(), defaultRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Role Binding", defaultRoleBinding.Name)
	}

	//SCC
	serviceaccount := "system:serviceaccount:" + projectName + ":default"

	privilegedSCCFound := &securityv1.SecurityContextConstraints{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: "privileged"}, privilegedSCCFound); err != nil {
		return reconcile.Result{}, err
	}

	if !util.StringInSlice(serviceaccount, privilegedSCCFound.Users) {
		privilegedSCCFound.Users = append(privilegedSCCFound.Users, serviceaccount)
		if err := r.client.Update(context.TODO(), privilegedSCCFound); err != nil {
			return reconcile.Result{}, err
		} else if err == nil {
			logrus.Infof("Updated %s Privileged SCC", privilegedSCCFound.Name)
		}
	}

	anyuidSCCFound := &securityv1.SecurityContextConstraints{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: "anyuid"}, anyuidSCCFound); err != nil {
		return reconcile.Result{}, err
	}

	if !util.StringInSlice(serviceaccount, anyuidSCCFound.Users) {
		anyuidSCCFound.Users = append(anyuidSCCFound.Users, serviceaccount)
		if err := r.client.Update(context.TODO(), anyuidSCCFound); err != nil {
			return reconcile.Result{}, err
		} else if err == nil {
			logrus.Infof("Updated %s Anyuid SCC", anyuidSCCFound.Name)
		}
	}

	// Explicitly allows traffic from the OpenShift ingress to the project
	networkPolicy := deployment.NewNetworkPolicyAllowFromOpenShfitIngress("allow-from-openshift-ingress", projectNamespace.Name)
	if err := r.client.Create(context.TODO(), networkPolicy); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Network Policy for %s", networkPolicy.Name, projectNamespace.Name)
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
