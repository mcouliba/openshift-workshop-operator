#!/bin/bash

IMAGE_TAG=$1

operator-sdk build quay.io/mcouliba/openshift-workshop-operator:${IMAGE_TAG}
docker push quay.io/mcouliba/openshift-workshop-operator:${IMAGE_TAG}