# ElasticServing

ElasticServing provides a Kubernetes custom resource definition (CRD) for serving machine learning (ML) models on mainstream framework such as tensorflow, onnx, paddle. It encapsulates the complexity of auto scaling, fault tolerant, health check and use kustomize for configuration reconcile. It also natively support heterogeneous hardware like nvidia GPU or KunLun chip. With ElasticServing it’s easy to scaling to zero and do the canary launch for ML deployment.

## Quick Start

Image used here is [Paddle Serving Image for CPU](https://github.com/PaddlePaddle/Serving#installation). This can be modified in ```config/configmap/configmap.yaml```

The sample used here is [Chinese Word Segmentation](https://github.com/PaddlePaddle/Serving#-pre-built-services-with-paddle-serving). The preparation work is done making used of the entrypoint of docker. This can be modified in ```args``` column in ```config/samples/elasticserving_v1_paddle.yaml``` 

### Prerequisites
- Kubernetes cluster
- Knative Serving with Istio Installed.
You can refer to the [Installation guide](https://knative.dev/docs/install/any-kubernetes-cluster/#installing-the-serving-component) or run `hack/install_knative.sh`.

### Installation

``` bash
# Download ElasticServing
git clone https://github.com/PaddleFlow/ElasticServing.git
cd ElasticServing

# Install elastic serving CRD and controller manager
kubectl apply -f assets/elasticserving_operator.yaml

# Deploy paddle service
kubectl apply -f assets/sample_service.yaml
```

### Installation Test

``` bash
# Check service in namespace paddleservice-system
kubectl get svc -n paddleservice-system

# Check knative service in namespace paddleservice-system
kubectl get ksvc -n paddleservice-system

# Check pods in namespace paddleservice-system
kubectl get pods -n paddleservice-system

# Check if the preparation work has been finished
kubectl logs <pod-name> -n paddleservice-system -c paddleserving -f

```

### Run Sample

``` bash
# Find the public IP address of the gateway (make a note of the EXTERNAL-IP field in the output)
kubectl get service istio-ingressgateway --namespace=istio-system
# If you are using minikube, the public IP address of the gateway will be listed once you execute the following command (There will exist four URLs and maybe choose the second one)
minikube service --url istio-ingressgateway -n istio-system

# Find the URL of the application. The expected result may be http://paddle-sample-service.paddleservice-system.example.com
kubectl get ksvc paddle-sample-service -n paddleservice-system

# Start to send data to the server. <IP-address> is what has been got in the first or the second command.
curl -H "Host: paddleservice-sample.paddleservice-system.example.com" -H "Content-Type:application/json" -X POST -d '{"feed":[{"words": "我爱北京天安门"}], "fetch":["word_seg"]}' http://<IP-address>/lac/prediction

```

#### Expected Result

``` bash
# The expected output should be 

default: 
{"result":[{"word_seg":"\u6211|\u7231|\u5317\u4eac|\u5929\u5b89\u95e8"}]}

canary:
{"result":[{"word_seg":"\u6211-\u7231-\u5317\u4eac-\u5929\u5b89\u95e8"}]}
```

## Installation

``` bash
# Download ElasticServing
git clone https://github.com/PaddleFlow/ElasticServing.git
cd ElasticServing

# Install elastic serving CRD
kubectl apply -f assets/crd.yaml

# Install elastic serving controller manager
kubectl apply -f assets/elasticserving_operator.yaml

# Deploy paddle service
kubectl apply -f assets/sample_service.yaml
```

### Create your own PaddleService

Imitate ```config/samples/elasticserving_v1_paddle.yaml``` to create your own PaddleService.  Please follow the following format.

example.yaml

``` yaml
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
  deploymentName: paddleservice
  runtimeVersion: paddleserving
  service:
    minScale: 1
```

Execute the following command:

``` bash
kubectl apply -f /dir/to/this/yaml/example.yaml
```

## License

This project is under the [Apache-2.0 license](https://github.com/PaddleFlow/ElasticServing/blob/main/LICENSE).
