package workshop

import (
	"context"
	"fmt"
	"time"

	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	smcp "github.com/redhat/openshift-workshop-operator/pkg/deployment/maistra/servicemeshcontrolplane"
	smmr "github.com/redhat/openshift-workshop-operator/pkg/deployment/maistra/servicemeshmemberroll"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Reconciling ServiceMesh
func (r *ReconcileWorkshop) reconcileServiceMesh(instance *openshiftv1alpha1.Workshop, users int) error {
	enabledServiceMesh := instance.Spec.ServiceMesh.Enabled

	if enabledServiceMesh {
		if err := r.addServiceMesh(instance, users); err != nil {
			return err
		}
	}

	//Success
	return nil
}

func (r *ReconcileWorkshop) addServiceMesh(instance *openshiftv1alpha1.Workshop, users int) error {
	reqLogger := log.WithName("ServiceMesh")

	jaegerOperatorImage := instance.Spec.ServiceMesh.JaegerOperatorImage
	kialiOperatorImage := instance.Spec.ServiceMesh.KialiOperatorImage
	istioOperatorImage := instance.Spec.ServiceMesh.IstioOperatorImage

	// JAEGER OPERATOR
	jaegerOperatorNamespace := deployment.NewNamespace(instance, "observability")
	if err := r.client.Create(context.TODO(), jaegerOperatorNamespace); err != nil && !errors.IsAlreadyExists(err) {
		reqLogger.Error(err, "Failed to create Namespace", "Resource.name", jaegerOperatorNamespace.Name)
		return err
	} else if err == nil {
		reqLogger.Info("Created Jaeger Operator Project")
	}

	jaegerOperatorCustomResourceDefinition := deployment.NewCustomResourceDefinition(instance, "jaegers.jaegertracing.io", "jaegertracing.io", "Jaeger", "JaegerList", "jaegers", "jaeger", "v1", nil, nil)
	if err := r.client.Create(context.TODO(), jaegerOperatorCustomResourceDefinition); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created  Jaeger Operator Custom Resource Definition")
	}

	jaegerOperatorServiceAccount := deployment.NewServiceAccount(instance, "jaeger-operator", jaegerOperatorNamespace.Name)
	if err := r.client.Create(context.TODO(), jaegerOperatorServiceAccount); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Jaeger Operator Service Account")
	}

	jaegerOperatorClusterRole := deployment.NewClusterRole(instance, "jaeger-operator", jaegerOperatorNamespace.Name, deployment.JaegerRules())
	if err := r.client.Create(context.TODO(), jaegerOperatorClusterRole); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Jaeger Operator Cluster Role")
	}

	jaegerOperatorClusterRoleBinding := deployment.NewClusterRoleBindingForServiceAccount(instance, "jaeger-operator", jaegerOperatorNamespace.Name, "jaeger-operator", "jaeger-operator", "ClusterRole")
	if err := r.client.Create(context.TODO(), jaegerOperatorClusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Jaeger Operator Cluster Role Binding")
	}

	args := []string{
		"start",
	}
	jaegerOperator := deployment.NewOperatorDeployment(instance, "jaeger-operator", jaegerOperatorNamespace.Name, jaegerOperatorImage, "jaeger-operator", 8383, nil, args, nil, nil)
	if err := r.client.Create(context.TODO(), jaegerOperator); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Jaeger Operator")
	}

	// KIALI-OPERATOR
	kialiOperatorNamespace := deployment.NewNamespace(instance, "kiali-operator")
	if err := r.client.Create(context.TODO(), kialiOperatorNamespace); err != nil && !errors.IsAlreadyExists(err) {
		reqLogger.Error(err, "Failed to create Namespace", "Resource.name", kialiOperatorNamespace.Name)
		return err
	} else if err == nil {
		reqLogger.Info("Created Kiali Operator Project (kiali-operator)")
	}

	kialiOperatorCustomResourceDefinition := deployment.NewCustomResourceDefinition(instance, "kialis.kiali.io", "kiali.io", "Kiali", "KialiList", "kialis", "kiali", "v1alpha1", nil, nil)
	if err := r.client.Create(context.TODO(), kialiOperatorCustomResourceDefinition); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Kiali Operator Custom Resource Definition (kialis.kiali.io)")
	}

	kialiOperatorCustomResourceDefinition2 := deployment.NewCustomResourceDefinition(instance, "monitoringdashboards.monitoring.kiali.io", "monitoring.kiali.io", "MonitoringDashboard", "MonitoringDashboardList", "monitoringdashboards", "monitoringdashboard", "v1alpha1", nil, nil)
	if err := r.client.Create(context.TODO(), kialiOperatorCustomResourceDefinition2); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Kiali Operator Custom Resource Definition (monitoringdashboards.monitoring.kiali.io)")
	}

	kialiOperatorServiceAccount := deployment.NewServiceAccount(instance, "kiali-operator", kialiOperatorNamespace.Name)
	if err := r.client.Create(context.TODO(), kialiOperatorServiceAccount); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Kiali Operator Service Account")
	}

	kialiOperatorClusterRole := deployment.NewClusterRole(instance, "kiali-operator", kialiOperatorNamespace.Name, deployment.KialiRules())
	if err := r.client.Create(context.TODO(), kialiOperatorClusterRole); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Kiali Operator Cluster Role")
	}

	kialiOperatorClusterRoleBinding := deployment.NewClusterRoleBindingForServiceAccount(instance, "kiali-operator", kialiOperatorNamespace.Name, "kiali-operator", "kiali-operator", "ClusterRole")
	if err := r.client.Create(context.TODO(), kialiOperatorClusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Kiali Operator Cluster Role Binding")
	}

	kialiOperator := deployment.NewAnsibleOperatorDeployment(instance, "kiali-operator", kialiOperatorNamespace.Name, kialiOperatorImage, "kiali-operator")
	if err := r.client.Create(context.TODO(), kialiOperator); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Kiali Operator")
	}

	// ISTIO-OPERATOR
	istioOperatorNamespace := deployment.NewNamespace(instance, "istio-operator")
	if err := r.client.Create(context.TODO(), istioOperatorNamespace); err != nil && !errors.IsAlreadyExists(err) {
		reqLogger.Error(err, "Failed to create Namespace", "Resource.name", istioOperatorNamespace.Name)
		return err
	} else if err == nil {
		reqLogger.Info("Created Istio Project (istio-operator)")
	}

	istioOperatorCustomResourceDefinition := deployment.NewCustomResourceDefinition(instance, "controlplanes.istio.openshift.com", "istio.openshift.com", "ControlPlane", "ControlPlaneList", "controlplanes", "controlplane", "v1alpha3", nil, nil)
	if err := r.client.Create(context.TODO(), istioOperatorCustomResourceDefinition); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Istio Custom Resource Definition (controlplanes.istio.openshift.com)")
	}

	istioOperatorCustomResourceDefinition2 := deployment.NewCustomResourceDefinition(instance, "servicemeshcontrolplanes.maistra.io", "maistra.io", "ServiceMeshControlPlane", "ServiceMeshControlPlaneList", "servicemeshcontrolplanes", "servicemeshcontrolplane", "v1", []string{"smcp"}, nil)
	if err := r.client.Create(context.TODO(), istioOperatorCustomResourceDefinition2); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Istio Custom Resource Definition (servicemeshcontrolplanes.maistra.io)")
	}

	additionalPrinterColumns := []apiextensionsv1beta1.CustomResourceColumnDefinition{
		{
			JSONPath:    ".spec.members",
			Description: "Namespaces that are members of this Control Plane",
			Name:        "Members",
			Type:        "string",
		},
	}
	istioOperatorCustomResourceDefinition3 := deployment.NewCustomResourceDefinition(instance, "servicemeshmemberrolls.maistra.io", "maistra.io", "ServiceMeshMemberRoll", "ServiceMeshMemberRollList", "servicemeshmemberrolls", "servicemeshmemberroll", "v1", []string{"smmr"}, additionalPrinterColumns)
	if err := r.client.Create(context.TODO(), istioOperatorCustomResourceDefinition3); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Istio Custom Resource Definition (servicemeshmemberrolls.maistra.io)")
	}

	istioOperatorServiceAccount := deployment.NewServiceAccount(instance, "istio-operator", istioOperatorNamespace.Name)
	if err := r.client.Create(context.TODO(), istioOperatorServiceAccount); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Istio Operator Service Account")
	}

	istioOperatorClusterRole := deployment.NewClusterRole(instance, "maistra-admin", istioOperatorNamespace.Name, deployment.MaistraAdminRules())
	if err := r.client.Create(context.TODO(), istioOperatorClusterRole); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Istio Operator Cluster Role (maistra-admin)")
	}

	istioOperatorClusterRole2 := deployment.NewClusterRole(instance, "istio-admin", istioOperatorNamespace.Name, deployment.IstioAdminRules())
	if err := r.client.Create(context.TODO(), istioOperatorClusterRole2); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Istio Operator Cluster Role (istio-admin)")
	}

	istioOperatorClusterRole3 := deployment.NewClusterRole(instance, "istio-operator", istioOperatorNamespace.Name, deployment.IstioRules())
	if err := r.client.Create(context.TODO(), istioOperatorClusterRole3); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("CreatedIstio Operator Cluster Role (istio-operator)")
	}

	istioOperatorClusterRoleBinding := deployment.NewClusterRoleBinding(instance, "maistra-admin", istioOperatorNamespace.Name, "maistra-admin", "ClusterRole")
	if err := r.client.Create(context.TODO(), istioOperatorClusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Istio Operator Cluster Role Binding (maistra-admin)")
	}

	istioOperatorClusterRoleBinding2 := deployment.NewClusterRoleBinding(instance, "istio-admin", istioOperatorNamespace.Name, "istio-admin", "ClusterRole")
	if err := r.client.Create(context.TODO(), istioOperatorClusterRoleBinding2); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Istio Operator Cluster Role Binding (istio-admin)")
	}

	istioOperatorClusterRoleBinding3 := deployment.NewClusterRoleBindingForServiceAccount(instance, "istio-operator", istioOperatorNamespace.Name, "istio-operator", "istio-operator", "ClusterRole")
	if err := r.client.Create(context.TODO(), istioOperatorClusterRoleBinding3); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Istio Operator Cluster Role Binding (istio-operator)")
	}

	istioValidatingWebhookConfiguration := deployment.NewValidatingWebhookConfiguration(instance, "istio-operator.servicemesh-resources.maistra.io", deployment.IstioWebHook())
	if err := r.client.Create(context.TODO(), istioValidatingWebhookConfiguration); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Istio Operator Validating Webhook Configuration")
	}

	targetPort := intstr.IntOrString{
		IntVal: 11999,
	}
	adminControllerService := deployment.NewCustomService(instance, "admission-controller", instance.Namespace, []string{"admin"}, []int32{443}, []intstr.IntOrString{targetPort})
	if err := r.client.Create(context.TODO(), adminControllerService); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Admin Controller Service")
	}

	commands := []string{
		"istio-operator",
		"--discoveryCacheDir",
		"/home/istio-operator/.kube/cache/discovery",
	}
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "discovery-cache",
			MountPath: "/home/istio-operator/.kube/cache/discovery",
		},
	}
	volumes := []corev1.Volume{
		{
			Name: "discovery-cache",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium: corev1.StorageMediumMemory,
				},
			},
		},
	}
	istioOperator := deployment.NewOperatorDeployment(instance, "istio-operator", istioOperatorNamespace.Name, istioOperatorImage, "istio-operator", 60000, commands, nil, volumeMounts, volumes)
	if err := r.client.Create(context.TODO(), istioOperator); err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if err == nil {
		reqLogger.Info("Created Istio Operator")
	}

	// Wait for Istio Operator to be running
	time.Sleep(30 * time.Second)

	// ISTIO-SYSTEM
	istioSystemNamespace := deployment.NewNamespace(instance, "istio-system")
	if err := r.client.Create(context.TODO(), istioSystemNamespace); err != nil && !errors.IsAlreadyExists(err) {
		reqLogger.Error(err, "Failed to create Namespace", "Resource.name", istioSystemNamespace.Name)
		return err
	} else if err == nil {
		reqLogger.Info("Created Istio Project (istio-system)")
	}

	serviceMeshControlPlaneCR := smcp.NewServiceMeshControlPlaneCR(smcp.NewServiceMeshControlPlaneCRParameters{
		Name:      "full-install",
		Namespace: istioSystemNamespace.Name,
	})
	if err := r.client.Create(context.TODO(), serviceMeshControlPlaneCR); err != nil && !errors.IsAlreadyExists(err) {
		return err
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
		return err
	} else if err == nil {
		reqLogger.Info("Created ServiceMeshControlPlane Custom Resource")
	}

	//Success
	return nil
}
