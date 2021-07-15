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
curl -H "Host: paddleservice-sample-service.paddleservice-system.example.com" -H "Content-Type:application/json" -X POST -d '{"feed":[{"words": "我爱北京天安门"}], "fetch":["word_seg"]}' http://<IP-address>/lac/prediction

```

#### Expected Result

``` bash
# The expected output should be 
{"result":[{"word_seg":"\u6211|\u7231|\u5317\u4eac|\u5929\u5b89\u95e8"}]}
```

## Installation

``` bash
# Download ElasticServing
git clone https://github.com/PaddleFlow/ElasticServing.git
cd ElasticServing

# Install elastic serving CRD and controller manager
kubectl apply -f assets/elasticserving_operator.yaml

# Deploy paddle service
kubectl apply -f assets/sample_service.yaml
```

### Change Paddle Serving Image 

``` yaml
paddleService: |-
{
"containerImage": "hub.baidubce.com/paddlepaddle/serving",
"version": "latest",
"port": 9292
}
```

The sample ```config/configmap/configmap.yaml``` uses TAG ``` latest``` in row ```version```. It is used for CPU runtime version of paddle serving image. If you want to use other version like GPU version, please check out [the Image description part](https://github.com/PaddlePaddle/Serving/blob/v0.4.0/doc/DOCKER_IMAGES.md#image-description).

Then run

``` bash
kubectl create -f config/configmap/configmap.yaml
```

### Create your own PaddleService

Imitate ```config/samples/elasticserving_v1_paddle.yaml``` to create your own PaddleService.  Please follow the following format.

example.yaml

``` yaml
apiVersion: elasticserving.paddlepaddle.org/v1
kind: PaddleService
metadata:
  name: <new-PS-name>
  namespace: <new-PS-namespace>
spec:
  # Add fields here
  deploymentName: <deployment-name>
  runtimeVersion: <runtime-version>
  arg: <argument>
  # (All the values below are default if omitted)
  service: 
    autoscalar: autoscaling.KPA
    metric: "concurrency"
    window: "60s"
    panicWindow: "10"
    panicThreshold: "200"
    minScale: 1 
    maxScale: 10
    target: 100
  resources:
    cpu: "0.2"
    memory: "512Mi"
```

Execute the following command:

``` bash
kubectl create ns <new-PS-namespace>
kubectl apply -f /dir/to/this/yaml/example.yaml
```

## License

This project is under the [Apache-2.0 license](https://github.com/PaddleFlow/ElasticServing/blob/main/LICENSE).
