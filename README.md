## OpenShift Workshop Operator

### Build and Push into quay.io

```sh
# Build and push the app-operator image to a public registry such as quay.io
$ operator-sdk build quay.io/mcouliba/openshift-workshop-operator:rhte
$ docker push quay.io/mcouliba/openshift-workshop-operator:rhte
```