# ElasticServing

ElasticServing provides a Kubernetes custom resource definition (CRD) for serving machine learning (ML) models on mainstream framework such as tensorflow, onnx, paddle. It encapsulates the complexity of auto scaling, fault tolerant, health check and use kustomize for configuration reconcile. It also natively support heterogeneous hardware like nvidia GPU or KunLun chip. With ElasticServing it’s easy to scaling to zero and do the canary launch for ML deployment.

## Installation

``` cd ElasticServing```

```make manifests```

```kubectl create -f config/crd/bases/elasticserving.paddlepaddle.org_paddles.yaml```

```kubectl create -f config/samples/elasticserving_v1_paddle.yaml```

```make run```

