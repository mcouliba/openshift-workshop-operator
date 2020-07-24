package workshop

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	routev1 "github.com/openshift/api/route/v1"
	openshiftv1alpha1 "github.com/redhat/openshift-workshop-operator/pkg/apis/openshift/v1alpha1"
	deployment "github.com/redhat/openshift-workshop-operator/pkg/deployment"
	giteacustomresource "github.com/redhat/openshift-workshop-operator/pkg/deployment/gitea"
	"github.com/redhat/openshift-workshop-operator/pkg/util"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciling Gitea
func (r *ReconcileWorkshop) reconcileGitea(instance *openshiftv1alpha1.Workshop, users int) (reconcile.Result, error) {
	enabledGitea := instance.Spec.Infrastructure.Gitea.Enabled

	if enabledGitea {

		if result, err := r.addGitea(instance, users); err != nil {
			return result, err
		}

		// Installed
		if instance.Status.Gitea != util.OperatorStatus.Installed {
			instance.Status.Gitea = util.OperatorStatus.Installed
			if err := r.client.Status().Update(context.TODO(), instance); err != nil {
				logrus.Errorf("Failed to update Workshop status: %s", err)
				return reconcile.Result{}, err
			}
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *ReconcileWorkshop) addGitea(instance *openshiftv1alpha1.Workshop, users int) (reconcile.Result, error) {

	imageName := instance.Spec.Infrastructure.Gitea.Image.Name
	imageTag := instance.Spec.Infrastructure.Gitea.Image.Tag

	giteaNamespace := deployment.NewNamespace(instance, "gitea")
	if err := r.client.Create(context.TODO(), giteaNamespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Project", giteaNamespace.Name)
	}

	giteaCustomResourceDefinition := deployment.NewCustomResourceDefinition(instance, "giteas.gpte.opentlc.com", "gpte.opentlc.com", "Gitea", "GiteaList", "giteas", "gitea", "v1alpha1", nil, nil)
	if err := r.client.Create(context.TODO(), giteaCustomResourceDefinition); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Custom Resource Definition", giteaCustomResourceDefinition.Name)
	}

	giteaServiceAccount := deployment.NewServiceAccount(instance, "gitea-operator", giteaNamespace.Name)
	if err := r.client.Create(context.TODO(), giteaServiceAccount); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Service Account", giteaServiceAccount.Name)
	}

	giteaClusterRole := deployment.NewClusterRole(instance, "gitea-operator", giteaNamespace.Name, deployment.GiteaRules())
	if err := r.client.Create(context.TODO(), giteaClusterRole); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Cluster Role", giteaClusterRole.Name)
	}

	giteaClusterRoleBinding := deployment.NewClusterRoleBindingForServiceAccount(instance, "gitea-operator", giteaNamespace.Name, "gitea-operator", "gitea-operator", "ClusterRole")
	if err := r.client.Create(context.TODO(), giteaClusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Cluster Role Binding", giteaClusterRoleBinding.Name)
	}

	giteaOperator := deployment.NewAnsibleOperatorDeployment(instance, "gitea-operator", giteaNamespace.Name, imageName+":"+imageTag, "gitea-operator")

	if err := r.client.Create(context.TODO(), giteaOperator); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Operator", giteaOperator.Name)
	}

	giteaCustomResource := giteacustomresource.NewGiteaCustomResource(instance, "gitea-server", giteaNamespace.Name)
	if err := r.client.Create(context.TODO(), giteaCustomResource); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		logrus.Infof("Created %s Custom Resource", giteaCustomResource.Name)
	}

	// Wait for Server to be running
	if !k8sclient.GetDeploymentStatus("gitea-server", giteaNamespace.Name) {
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 1}, nil
	}

	// extract app route suffix from openshift-console
	giteaRouteFound := &routev1.Route{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: "gitea-server", Namespace: giteaNamespace.Name}, giteaRouteFound); err != nil {
		logrus.Errorf("Failed to find %s route", "gitea-server")
		return reconcile.Result{}, err
	}

	giteaURL := "https://" + giteaRouteFound.Spec.Host

	for id := 1; id <= users; id++ {
		username := fmt.Sprintf("user%d", id)

		if result, err := createGitUser(instance, username, giteaURL); err != nil {
			return result, err
		}
	}
	//Success
	return reconcile.Result{}, nil
}

func createGitUser(instance *openshiftv1alpha1.Workshop, username string, giteaURL string) (reconcile.Result, error) {

	var (
		openshiftUserPassword = instance.Spec.User.Password
		err                   error
		httpResponse          *http.Response
		httpRequest           *http.Request
		requestURL            = giteaURL + "/user/sign_up"
		body                  = url.Values{}
		client                = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			// Do not follow Redirect
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	)

	body.Set("user_name", username)
	body.Set("email", username+"@none.com")
	body.Set("password", openshiftUserPassword)
	body.Set("retype", openshiftUserPassword)

	httpRequest, err = http.NewRequest("POST", requestURL, strings.NewReader(body.Encode()))
	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpRequest.Header.Set("Accept", "application/json")
	httpRequest.Header.Set("Content-Length", strconv.Itoa(len(body.Encode())))

	httpResponse, err = client.Do(httpRequest)
	if err != nil {
		return reconcile.Result{}, err
	}
	if httpResponse.StatusCode == http.StatusCreated {
		logrus.Infof("Created %s user in Gitea", username)
	}

	defer httpResponse.Body.Close()

	//Success
	return reconcile.Result{}, nil
}
