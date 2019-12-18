package workshop

import (
	"context"
	"reflect"
	"time"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	"github.com/redhat/openshift-workshop-operator/pkg/util"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciling ArgoCD
func (r *ReconcileWorkshop) reconcileArgoCD(instance *openshiftv1alpha1.Workshop, users int,
	appsHostnameSuffix string, openshiftConsoleURL string, openshiftAPIURL string) (reconcile.Result, error) {
	enabledArgoCD := instance.Spec.Infrastructure.ArgoCD.Enabled

	if enabledArgoCD {

		if result, err := r.addArgoCD(instance, users, appsHostnameSuffix, openshiftConsoleURL, openshiftAPIURL); err != nil {
			return result, err
		}

		// Installed
		if instance.Status.ArgoCD != util.OperatorStatus.Installed {
			instance.Status.ArgoCD = util.OperatorStatus.Installed
			if err := r.client.Status().Update(context.TODO(), instance); err != nil {
				logrus.Errorf("Failed to update Workshop status: %s", err)
				return reconcile.Result{}, err
			}
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addArgoCD(instance *openshiftv1alpha1.Workshop, users int,
	appsHostnameSuffix string, openshiftConsoleURL string, openshiftAPIURL string) (reconcile.Result, error) {

	namespace := deployment.NewNamespace(instance, "argocd")
	if err := r.client.Create(context.TODO(), namespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Project", namespace.Name)
	}

	catalogSource := deployment.NewCatalogSource(instance, "argocd-catalog", "quay.io/mcouliba/argocd-operator-registry:latest", "Argo CD Operator", "Argo CD")
	if err := r.client.Create(context.TODO(), catalogSource); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Catalog Source", catalogSource.Name)
	}

	operatorGroup := deployment.NewOperatorGroup(instance, "argocd-operator", namespace.Name)
	if err := r.client.Create(context.TODO(), operatorGroup); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s OperatorGroup", operatorGroup.Name)
	}

	subscription := deployment.NewCustomSubscription(instance, "argocd-operator", namespace.Name, "argocd-operator",
		"alpha", catalogSource.Name)
	if err := r.client.Create(context.TODO(), subscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Subscription", subscription.Name)
	}

	// Wait for ArgoCD Operator to be running
	time.Sleep(time.Duration(1) * time.Second)
	operatorDeployment, err := r.GetEffectiveDeployment(instance, "argocd-operator", namespace.Name)
	if err != nil {
		logrus.Warnf("Waiting for OLM to create %s deployment", "argocd-operator")
		return reconcile.Result{}, err
	}

	if operatorDeployment.Status.AvailableReplicas != 1 {
		scaled := k8sclient.GetDeploymentStatus("argocd-operator", namespace.Name)
		if !scaled {
			return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 1}, err
		}
	}

	argocdRoute := "https://argocd-server-argocd." + appsHostnameSuffix
	redirectURIs := []string{argocdRoute + "/api/dex/callback"}
	oauthClient := deployment.NewOAuthClient(instance, "argocd-dex", redirectURIs)
	if err := r.client.Create(context.TODO(), oauthClient); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s OAuth Client", oauthClient.Name)
	}

	data := map[string]string{
		"dex.config": `connectors:
  - type: openshift
    id: openshift
    name: OpenShift
    config:
      issuer: ` + openshiftAPIURL + `:6443
      clientID: ` + oauthClient.Name + `
      clientSecret: openshift
      insecureCA: true
      redirectURI: ` + argocdRoute + `/api/dex/callback`,
		"url": argocdRoute,
	}

	labels := map[string]string{
		"app.kubernetes.io/name":    "argocd-cm",
		"app.kubernetes.io/part-of": "argocd",
	}

	configmap := deployment.NewConfigMap(instance, "argocd-cm", namespace.Name, labels, data)
	if err := r.client.Create(context.TODO(), configmap); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s ConfigMap", configmap.Name)
	} else if errors.IsAlreadyExists(err) {
		configMapFound := &corev1.ConfigMap{}
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: configmap.Name, Namespace: namespace.Name}, configMapFound); err != nil {
			return reconcile.Result{}, err
		} else if err == nil {
			if !reflect.DeepEqual(configMapFound.Data, data) {
				// Update
				configMapFound.Data = data
				if err := r.client.Update(context.TODO(), configMapFound); err != nil {
					return reconcile.Result{}, err
				}
				logrus.Infof("Updated %s ConfigMap", configMapFound.Name)
			}
		}
	}

	customResource := deployment.NewArgoCDCustomResource(instance, "argocd", namespace.Name)
	if err := r.client.Create(context.TODO(), customResource); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Custom Resource", customResource.Name)
	}

	//Success
	return reconcile.Result{}, nil
}
