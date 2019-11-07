#!/bin/bash

IMAGE=$1

operator-sdk build ${IMAGE}
docker push ${IMAGE}