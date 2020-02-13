//
// Copyright (c) 2012-2019 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//
package workshop

import (
	"context"

	oauth "github.com/openshift/api/oauth/v1"
	routev1 "github.com/openshift/api/route/v1"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileWorkshop) GetEffectiveDeployment(instance *openshiftv1alpha1.Workshop, name string, namespace string) (deployment *appsv1.Deployment, err error) {
	deployment = &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, deployment)
	if err != nil {
		logrus.Errorf("Failed to get %s deployment: %s", name, err)
		return nil, err
	}
	return deployment, nil
}

func (r *ReconcileWorkshop) GetEffectiveIngress(instance *openshiftv1alpha1.Workshop, name string, namespace string) (ingress *v1beta1.Ingress) {
	ingress = &v1beta1.Ingress{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, ingress)
	if err != nil {
		logrus.Errorf("Failed to get %s ingress: %s", name, err)
		return nil
	}
	return ingress
}

func (r *ReconcileWorkshop) GetEffectiveRoute(instance *openshiftv1alpha1.Workshop, name string, namespace string) (route *routev1.Route) {
	route = &routev1.Route{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, route)
	if err != nil {
		logrus.Errorf("Failed to get %s route: %s", name, err)
		return nil
	}
	return route
}

func (r *ReconcileWorkshop) GetEffectiveConfigMap(instance *openshiftv1alpha1.Workshop, name string, namespace string) (configMap *corev1.ConfigMap, err error) {
	configMap = &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, configMap)
	if err != nil {
		logrus.Errorf("Failed to get %s config map: %s", name, err)
		return nil, err
	}
	return configMap, nil

}

func (r *ReconcileWorkshop) GetEffectiveSecretResourceVersion(instance *openshiftv1alpha1.Workshop, name string, namespace string) string {
	secret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, secret)
	if err != nil {
		if !errors.IsNotFound(err) {
			logrus.Errorf("Failed to get %s secret: %s", name, err)
		}
		return ""
	}
	return secret.ResourceVersion
}

func (r *ReconcileWorkshop) GetEffectiveCSV(instance *openshiftv1alpha1.Workshop, name string, namespace string) (clusterServiceVersion *olmv1alpha1.ClusterServiceVersion, err error) {
	clusterServiceVersion = &olmv1alpha1.ClusterServiceVersion{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, clusterServiceVersion)
	if err != nil {
		logrus.Errorf("Failed to get %s cluster service version: %s", name, err)
		return nil, err
	}
	return clusterServiceVersion, nil

}

func (r *ReconcileWorkshop) GetCR(request reconcile.Request) (instance *openshiftv1alpha1.Workshop, err error) {
	err = r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		logrus.Errorf("Failed to get %s CR: %s", instance.Name, err)
		return nil, err
	}
	return instance, nil
}

func (r *ReconcileWorkshop) GetOAuthClient(oAuthClientName string) (oAuthClient *oauth.OAuthClient, err error) {
	oAuthClient = &oauth.OAuthClient{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: oAuthClientName, Namespace: ""}, oAuthClient); err != nil {
		logrus.Errorf("Failed to Get oAuthClient %s: %s", oAuthClientName, err)
		return nil, err
	}
	return oAuthClient, nil
}

func (r *ReconcileWorkshop) GetDeploymentEnv(deployment *appsv1.Deployment, key string) (value string) {
	env := deployment.Spec.Template.Spec.Containers[0].Env
	for i := range env {
		name := env[i].Name
		if name == key {
			value = env[i].Value
			break
		}
	}
	return value
}

func (r *ReconcileWorkshop) GetDeploymentEnvVarSource(deployment *appsv1.Deployment, key string) (valueFrom *corev1.EnvVarSource) {
	env := deployment.Spec.Template.Spec.Containers[0].Env
	for i := range env {
		name := env[i].Name
		if name == key {
			valueFrom = env[i].ValueFrom
			break
		}
	}
	return valueFrom
}

func (r *ReconcileWorkshop) GetInstallPlan(name string, namespace string) (installPlan *olmv1alpha1.InstallPlan, err error) {
	installPlan = &olmv1alpha1.InstallPlan{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, installPlan); err != nil {
		logrus.Errorf("Failed to Get InstallPlan %s: %s", installPlan, err)
		return nil, err
	}
	return installPlan, nil
}

func (r *ReconcileWorkshop) GetSubscription(name string, namespace string) (subscription *olmv1alpha1.Subscription, err error) {
	subscription = &olmv1alpha1.Subscription{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, subscription); err != nil {
		logrus.Errorf("Failed to Get Subscription %s: %s", subscription, err)
		return nil, err
	}
	return subscription, nil
}
