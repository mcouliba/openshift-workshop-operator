package workshop

import (
	"context"
	"fmt"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	"k8s.io/apimachinery/pkg/api/errors"
)

// Reconciling Squash
func (r *ReconcileWorkshop) reconcileSquash(instance *openshiftv1alpha1.Workshop, users int) error {
	enabledSquash := instance.Spec.Squash.Enabled

	if enabledSquash {
		if err := r.grantSquash(instance, users); err != nil {
			return err
		}
	}

	//Success
	return nil
}

func (r *ReconcileWorkshop) grantSquash(instance *openshiftv1alpha1.Workshop, users int) error {
	reqLogger := log.WithName("Squash")

	cheWorkspaceClusterRoleBinding := deployment.NewClusterRoleBindingForServiceAccount(instance, "cluster-admin-che-workspace",
		"workspaces", "che-workspace", "cluster-admin", "ClusterRole")
	if err := r.client.Create(context.TODO(), cheWorkspaceClusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created 'che-workspace' Role Binding for Squash")
	}

	id := 1
	for {
		infraProjectName := fmt.Sprintf("infra%d", id)
		squashPlankClusterRoleBinding := deployment.NewClusterRoleBindingForServiceAccount(instance,
			"cluster-admin-squash-plank-"+infraProjectName, infraProjectName, "squash-plank", "cluster-admin", "ClusterRole")

		if id <= users {
			if err := r.client.Create(context.TODO(), squashPlankClusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
				return err
			} else if err == nil {
				reqLogger.Info("Created Squash Plank Role Binding for Squash for project '" + infraProjectName + "'")
			}
		} else {
			if err := r.client.Delete(context.TODO(), squashPlankClusterRoleBinding); err != nil {
				break
			}
		}
		id++
	}

	//Success
	return nil
}
