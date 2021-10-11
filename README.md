# ElasticServing

English | [简体中文](./README-zh_CN.md)

ElasticServing provides a Kubernetes custom resource definition (CRD) for serving machine learning (ML) models on mainstream framework such as tensorflow, onnx, paddle. It encapsulates the complexity of auto scaling, fault tolerant, health check and use kustomize for configuration reconcile. It also natively support heterogeneous hardware like nvidia GPU or KunLun chip. With ElasticServing it's easy to scaling to zero and do the canary launch for ML deployment.

## Quick Start

The image used in our sample service is based on [Paddle Serving Image for CPU](https://github.com/PaddlePaddle/Serving#installation).

The sample used here is [Resnet50 in ImageNet](https://github.com/PaddlePaddle/Serving/tree/v0.6.0/python/examples/imagenet) and [Chinese Word Segmentation](https://github.com/PaddlePaddle/Serving#-pre-built-services-with-paddle-serving). The preparation work is done based on the entrypoint of docker. This can be modified in `arg`.

### Prerequisites
- Kubernetes >= 1.18
- Knative Serving with networking layer Installed.
You can refer to the [installation guide](https://knative.dev/v0.21-docs/install/any-kubernetes-cluster/#installing-the-serving-component) or run `hack/install_knative.sh`(knative serving v0.21 with istio) / `hack/install_knative_kourier.sh`(knative serving v0.22 with kourier).

### Installation

```bash
# Download ElasticServing
git clone https://github.com/PaddleFlow/ElasticServing.git
cd ElasticServing

# Install elastic serving CRD
kubectl apply -f assets/crd.yaml

# Install elastic serving controller manager
kubectl apply -f assets/elasticserving_operator.yaml
```

### Run Sample

```bash
# Deploy paddle service
kubectl apply -f assets/sample_service.yaml
```

#### Sample Service Test

```bash
# Check service in namespace paddleservice-system
kubectl get svc -n paddleservice-system

# Check knative service in namespace paddleservice-system
kubectl get ksvc -n paddleservice-system

# Check pods in namespace paddleservice-system
kubectl get pods -n paddleservice-system

# Check if the preparation work has been finished
kubectl logs <pod-name> -n paddleservice-system -c paddleserving
```

We use Istio as the networking layer for Knative serving. It's also fine for users to use others, i.e, Kourier, Contour and Ambassador.

```bash
# Find the public IP address of the gateway (make a note of the EXTERNAL-IP field in the output)
kubectl get service istio-ingressgateway --namespace=istio-system
# If the EXTERNAL-IP is pending, get the ip with the following command
kubectl get po -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].status.hostIP}'
# If you are using minikube, the public IP address of the gateway will be listed once you execute the following command (There will exist four URLs and maybe choose the second one)
minikube service --url istio-ingressgateway -n istio-system

# Get the port of the gateway
kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="http2")].nodePort}'

# Find the URL of the application. The expected result may be http://paddleservice-sample.paddleservice-system.example.com
kubectl get ksvc paddle-sample-service -n paddleservice-system
```

#### Resnet_50_vd sample
The related `sample_service.yaml` is as follows:
```yaml
apiVersion: v1
kind: Namespace
metadata:
  labels:
    istio-injection: enabled
  name: paddleservice-system
---
apiVersion: elasticserving.paddlepaddle.org/v1
kind: PaddleService
metadata:
  name: paddleservice-sample
  namespace: paddleservice-system
spec:
  canary:
    arg: cd Serving/python/examples/imagenet && python3 resnet50_web_service_canary.py
      ResNet50_vd_model cpu 9292
    containerImage: jinmionhaobaidu/resnetcanary
    port: 9292
    tag: latest
  canaryTrafficPercent: 50
  default:
    arg: cd Serving/python/examples/imagenet && python3 resnet50_web_service.py ResNet50_vd_model
      cpu 9292
    containerImage: jinmionhaobaidu/resnet
    port: 9292
    tag: latest
  runtimeVersion: paddleserving
  service:
    minScale: 0
    window: 10s
```
```bash
# Start to send data to the server. <IP-address> is what has been got in the first or the second command.
curl -H "host:paddleservice-sample.paddleservice-system.example.com" -H "Content-Type:application/json" -X POST -d '{"feed":[{"image": "https://paddle-serving.bj.bcebos.com/imagenet-example/daisy.jpg"}], "fetch": ["score"]}' http://<IP-address>:<Port>/image/prediction
```

##### Expected Result
```
# The expected output should be

default:
{"result":{"label":["daisy"],"prob":[0.9341399073600769]}}

canary:
{"result":{"isCanary":["true"],"label":["daisy"],"prob":[0.9341399073600769]}}
```

### Create your own PaddleService

After insntalling CRD ```kubectl apply -f assets/crd.yaml``` and controller manager ```kubectl apply -f assets/elasticserving_operator.yaml```, you can build your own PaddleService by applying your yaml which looks like the following one.

example.yaml

```yaml
apiVersion: v1
kind: Namespace
metadata:
  labels:
    istio-injection: enabled
  name: paddleservice-system
---
apiVersion: elasticserving.paddlepaddle.org/v1
kind: PaddleService
metadata:
  name: paddleservice-sample
  namespace: paddleservice-system
spec:
  canary:
    arg: python3 Serving/python/examples/lac/lac_web_service_canary.py lac_model/
      lac_workdir 9292
    containerImage: jinmionhaobaidu/pdservinglaccanary
    port: 9292
    tag: latest
  canaryTrafficPercent: 50
  default:
    arg: python3 Serving/python/examples/lac/lac_web_service.py lac_model/ lac_workdir
      9292
    containerImage: jinmionhaobaidu/pdservinglac
    port: 9292
    tag: latest
  runtimeVersion: paddleserving
  service:
    minScale: 0
    maxScale: 0
    autoscaler: "kpa"
    metric: "concurrency" # scaling metric
    window: "60s"
    panicWindow: 10 # percentage of stable window
    target: 100
    targetUtilization: 70
```

Please note that only the field `default` is required. Other fields can be empty and default value will be set. Field `canary` and `canaryTrafficPercent` are not required if your own paddleservice doesn't need them.

Execute the following command:

```bash
kubectl apply -f /dir/to/this/yaml/example.yaml
```

## More Examples

- [BERT](./docs/en/examples/bert.md): Semantic Understanding Prediction
- [LAC](./docs/en/examples/lac.md): Chinese Word Segmentation
- [Criteo Ctr](./docs/en/examples/criteo_ctr.md): CTR Prediction Service
- [Wide & Deep](./docs/en/examples/wide_deep.md)： Wide & Deep Pipeline

## More Information

Please refer to the [API docs](./docs/en/api_doc.md) for more information about custom resource definition.

## License

This project is under the [Apache-2.0 license](https://github.com/PaddleFlow/ElasticServing/blob/main/LICENSE).
