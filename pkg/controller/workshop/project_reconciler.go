package workshop

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	routev1 "github.com/openshift/api/route/v1"
	securityv1 "github.com/openshift/api/security/v1"
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	"github.com/redhat/openshift-workshop-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// Reconciling Project
func (r *ReconcileWorkshop) reconcileProject(instance *openshiftv1alpha1.Workshop, users int,
	userEndpointStr *strings.Builder, appsHostnameSuffix string, openshiftConsoleURL string, openshiftAPIURL string) error {
	enabledGuide := instance.Spec.Guide.Enabled

	id := 1
	for {
		username := fmt.Sprintf("user%d", id)
		coolstoreProjectName := fmt.Sprintf("coolstore%d", id)
		infraProjectName := fmt.Sprintf("infra%d", id)

		if id <= users {
			// Add user endpoints in Etherpad
			userEndpointStr.WriteString(fmt.Sprintf("You are user%d\t|\thttp://guide-infra%d.%s\t|\t<INSERT_YOUR_NAME>\n", id, id, appsHostnameSuffix))

			// Coolstore Project
			if err := r.addCoolstoreProject(instance, coolstoreProjectName, username); err != nil {
				return err
			}

			// Infra Project
			if err := r.addInfraProject(instance, infraProjectName, username); err != nil {
				return err
			}

			// Guide
			if enabledGuide {
				if err := r.addGuide(instance, coolstoreProjectName, infraProjectName, username,
					appsHostnameSuffix, openshiftConsoleURL, openshiftAPIURL); err != nil {
					return err
				}
			} else {
				if err := r.deleteGuide(instance, infraProjectName); err != nil {
					return err
				}
			}

		} else {

			coolstoreNamespace := deployment.NewNamespace(instance, coolstoreProjectName)
			coolstoreNamespaceFound := &corev1.Namespace{}
			coolstoreNamespaceErr := r.client.Get(context.TODO(), types.NamespacedName{Name: coolstoreNamespace.Name}, coolstoreNamespaceFound)

			infraNamespace := deployment.NewNamespace(instance, infraProjectName)
			infraNamespaceFound := &corev1.Namespace{}
			infraNamespaceErr := r.client.Get(context.TODO(), types.NamespacedName{Name: infraNamespace.Name}, infraNamespaceFound)

			if coolstoreNamespaceErr != nil && errors.IsNotFound(coolstoreNamespaceErr) && infraNamespaceErr != nil && errors.IsNotFound(infraNamespaceErr) {
				break
			}

			if err := r.deleteProject(coolstoreNamespace); err != nil {
				return err
			}

			if err := r.deleteProject(infraNamespace); err != nil {
				return err
			}
		}

		id++
	}

	//Success
	return nil
}

func (r *ReconcileWorkshop) addCoolstoreProject(instance *openshiftv1alpha1.Workshop, coolstoreProjectName string, username string) error {
	reqLogger := log.WithName("Coolstore Project")

	// CoolStoreProject
	coolstoreNamespace := deployment.NewNamespace(instance, coolstoreProjectName)
	if err := r.client.Create(context.TODO(), coolstoreNamespace); err != nil && !errors.IsAlreadyExists(err) {
		reqLogger.Error(err, "Failed to create Namespace", "Resource.name", coolstoreNamespace.Name)
		return err
	} else if err == nil {
		reqLogger.Info("Created " + coolstoreNamespace.Name + " project")
	}

	istioRole := deployment.NewRole(deployment.NewRoleParameters{
		Name:      username + "-istio",
		Namespace: coolstoreNamespace.Name,
		Rules:     deployment.IstioUserRules(),
	})
	if err := r.client.Create(context.TODO(), istioRole); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created " + username + "-istio" + " Role")
	}

	istioRoleBinding := deployment.NewRoleBindingUser(deployment.NewRoleBindingUserParameters{
		Name:      username + "-istio",
		Namespace: coolstoreNamespace.Name,
		Username:  username,
		RoleName:  username + "-istio",
		RoleKind:  "Role",
	})
	if err := r.client.Create(context.TODO(), istioRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created " + username + "-istio" + " Role Binding")
	}

	userRoleBinding := deployment.NewRoleBindingUser(deployment.NewRoleBindingUserParameters{
		Name:      username + "-admin",
		Namespace: coolstoreNamespace.Name,
		Username:  username,
		RoleName:  "admin",
		RoleKind:  "ClusterRole",
	})
	if err := r.client.Create(context.TODO(), userRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created " + username + "-admin Role Binding for Coolstore Project")
	}

	defaultRoleBinding := deployment.NewRoleBindingSA(deployment.NewRoleBindingSAParameters{
		Name:               "view",
		Namespace:          coolstoreNamespace.Name,
		ServiceAccountName: "default",
		RoleName:           "view",
		RoleKind:           "ClusterRole",
	})
	if err := r.client.Create(context.TODO(), defaultRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created View Role Binding")
	}

	//SCC
	serviceaccount := "system:serviceaccount:" + coolstoreProjectName + ":default"

	privilegedSCCFound := &securityv1.SecurityContextConstraints{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: "privileged"}, privilegedSCCFound); err != nil {
		reqLogger.Error(err, "Failed to get Privileged SCC", "Resource.name", "privileged")
		return err
	}

	if !util.StringInSlice(serviceaccount, privilegedSCCFound.Users) {
		privilegedSCCFound.Users = append(privilegedSCCFound.Users, serviceaccount)
		if err := r.client.Update(context.TODO(), privilegedSCCFound); err != nil {
			reqLogger.Error(err, "Failed to update the Privileged SCC")
			return err
		} else if err == nil {
			reqLogger.Info("Updated the Privileged SCC")
		}
	}

	anyuidSCCFound := &securityv1.SecurityContextConstraints{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: "anyuid"}, anyuidSCCFound); err != nil {
		reqLogger.Error(err, "Failed to get Anyuid SCC", "Resource.name", "anyuid")
		return err
	}

	if !util.StringInSlice(serviceaccount, anyuidSCCFound.Users) {
		anyuidSCCFound.Users = append(anyuidSCCFound.Users, serviceaccount)
		if err := r.client.Update(context.TODO(), anyuidSCCFound); err != nil {
			reqLogger.Error(err, "Failed to update the Anyuid SCC")
			return err
		} else if err == nil {
			reqLogger.Info("Updated the Anyuid SCC")
		}
	}

	//Success
	return nil
}

func (r *ReconcileWorkshop) addInfraProject(instance *openshiftv1alpha1.Workshop, infraProjectName string, username string) error {
	reqLogger := log.WithName("Infra Project")

	// InfraProject
	infraNamespace := deployment.NewNamespace(instance, infraProjectName)
	if err := r.client.Create(context.TODO(), infraNamespace); err != nil && !errors.IsAlreadyExists(err) {
		reqLogger.Error(err, "Failed to create Namespace", "Resource.name", infraNamespace.Name)
		return err
	} else if err == nil {
		reqLogger.Info("Created " + infraProjectName + " project")
	}

	userRoleBinding := deployment.NewRoleBindingUser(deployment.NewRoleBindingUserParameters{
		Name:      username + "-admin",
		Namespace: infraNamespace.Name,
		Username:  username,
		RoleName:  "admin",
		RoleKind:  "ClusterRole",
	})
	if err := r.client.Create(context.TODO(), userRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created " + username + "-admin Role Binding for Infra Project")
	}

	//Success
	return nil
}

func (r *ReconcileWorkshop) deleteProject(namespaces *corev1.Namespace) error {
	reqLogger := log.WithName("Project")

	if err := r.client.Delete(context.TODO(), namespaces); err != nil && !errors.IsNotFound(err) {
		reqLogger.Error(err, "Failed to delete Namespace", "Resource.name", namespaces.Name)
		return err
	} else if err == nil {
		reqLogger.Info("Deleted Project '" + namespaces.Name + "'")
	}

	//Success
	return nil
}

func (r *ReconcileWorkshop) addGuide(instance *openshiftv1alpha1.Workshop, coolstoreProjectName string, infraProjectName string, username string,
	appsHostnameSuffix string, openshiftConsoleURL string, openshiftAPIURL string) error {
	reqLogger := log.WithName("Guide")

	// Deploy/Update Guide
	guideDeployment := deployment.NewWorkshopperDeployment(instance, "guide", infraProjectName, coolstoreProjectName,
		infraProjectName, username, appsHostnameSuffix, openshiftConsoleURL, openshiftAPIURL)
	if err := r.client.Create(context.TODO(), guideDeployment); err != nil && !errors.IsAlreadyExists(err) {
		reqLogger.Error(err, "Failed to deploy Guide", "Resource.name", infraProjectName)
		return err
	} else if err == nil {
		reqLogger.Info("Created Guide Service for " + infraProjectName)
	} else if errors.IsAlreadyExists(err) {
		if !reflect.DeepEqual(guideDeployment.Spec.Template.Spec.Containers[0].Env, guideDeployment.Spec.Template.Spec.Containers[0].Env) {
			// Update Guide
			reqLogger.Info("Updating the Guide", "Namespace.name", infraProjectName)
			if err := r.client.Update(context.TODO(), guideDeployment); err != nil {
				reqLogger.Error(err, "Failed to update Guide", "Resource.name", infraProjectName)
				return err
			}
		}
	}

	// Create Service
	guideService := deployment.NewService(instance, "guide", infraProjectName, []string{"http"}, []int32{8080})
	if err := r.client.Create(context.TODO(), guideService); err != nil && !errors.IsAlreadyExists(err) {
		reqLogger.Error(err, "Failed to create Guide Service", "Resource.name", infraProjectName)
		return err
	} else if err == nil {
		reqLogger.Info("Created Guide Service for " + infraProjectName)
	}

	// Create Route
	guideRoute := deployment.NewRoute(instance, "guide", infraProjectName, "guide", 8080)
	if err := r.client.Create(context.TODO(), guideRoute); err != nil && !errors.IsAlreadyExists(err) {
		reqLogger.Error(err, "Failed to create Guide Route", "Resource.name", infraProjectName)
		return err
	} else if err == nil {
		reqLogger.Info("Created Guide Route for " + infraProjectName)
	}

	//Success
	return nil
}

func (r *ReconcileWorkshop) deleteGuide(instance *openshiftv1alpha1.Workshop, infraProjectName string) error {
	reqLogger := log.WithName("Guide")

	guideRouteFound := &routev1.Route{}
	guideRouteErr := r.client.Get(context.TODO(), types.NamespacedName{Name: "guide", Namespace: infraProjectName}, guideRouteFound)
	if guideRouteErr == nil {
		// Delete Route
		reqLogger.Info("Deleting a Guide Route", "Namespace.name", infraProjectName)
		if err := r.client.Delete(context.TODO(), guideRouteFound); err != nil {
			reqLogger.Error(err, "Failed to delete Guide Route", "Resource.name", infraProjectName)
			return err
		}
	}
	guideServiceFound := &corev1.Service{}
	guideServiceErr := r.client.Get(context.TODO(), types.NamespacedName{Name: "guide", Namespace: infraProjectName}, guideServiceFound)
	if guideServiceErr == nil {
		// Delete Service
		reqLogger.Info("Deleting a Guide Service", "Namespace.name", infraProjectName)
		if err := r.client.Delete(context.TODO(), guideServiceFound); err != nil {
			reqLogger.Error(err, "Failed to delete Guide Service", "Resource.name", infraProjectName)
			return err
		}
	}

	guideDeploymentFound := &appsv1.Deployment{}
	guideDeploymentErr := r.client.Get(context.TODO(), types.NamespacedName{Name: "guide", Namespace: infraProjectName}, guideDeploymentFound)
	if guideDeploymentErr == nil {
		// Undeploy Guide
		reqLogger.Info("Undeploying Guide")
		if err := r.client.Delete(context.TODO(), guideDeploymentFound); err != nil {
			reqLogger.Error(err, "Failed to undeploy Guide")
			return err
		}
	}

	//Success
	return nil
}
