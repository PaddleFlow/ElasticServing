#!/usr/bin/bash

set -ex

export KNATIVE_VERSION=v0.22.0

# Install knative
kubectl apply --filename https://github.com/knative/serving/releases/download/${KNATIVE_VERSION}/serving-crds.yaml
kubectl apply --filename https://github.com/knative/serving/releases/download/${KNATIVE_VERSION}/serving-core.yaml

# Setup kourier
kubectl apply --filename https://github.com/knative/net-kourier/releases/download/${KNATIVE_VERSION}/kourier.yaml
kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress.class":"kourier.ingress.networking.knative.dev"}}'
