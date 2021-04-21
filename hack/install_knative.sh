#!/usr/bin/bash

set -ex

# NOTE: some resources doesn't exist for KNATIVE_VERSION < v0.21.0
export KNATIVE_VERSION=v0.21.0

# Install knative
kubectl apply --filename https://github.com/knative/serving/releases/download/${KNATIVE_VERSION}/serving-crds.yaml
kubectl apply --filename https://github.com/knative/serving/releases/download/${KNATIVE_VERSION}/serving-core.yaml

# Setup Istio
kubectl apply --filename https://github.com/knative/net-istio/releases/download/${KNATIVE_VERSION}/istio.yaml
kubectl apply --filename https://github.com/knative/net-istio/releases/download/${KNATIVE_VERSION}/net-istio.yaml
