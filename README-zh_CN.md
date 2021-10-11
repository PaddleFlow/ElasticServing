# ElasticServing

[English](./README.md) | 简体中文

ElasticServing 通过提供自定义资源 PaddleService，支持用户在 Kubernetes 集群上使用 TensorFlow、onnx、PaddlePaddle 等主流框架部署模型服务。
ElasticServing 构建在 [Knative Serving](https://github.com/knative/serving) 之上，其提供了自动扩缩容、容错、健康检查等功能，并且支持在异构硬件上部署服务，如 Nvidia GPU 或 昆仑芯片。
ElasticServing 采用的是 serverless 架构，当没有预估请求时，服务规模可以缩容到零，以节约集群资源，同时它还支持并蓝绿发版等功能。

## 快速开始

本示例使用的模型服务镜像基于 [Paddle Serving CPU 版](https://github.com/PaddlePaddle/Serving/blob/v0.6.0/README_CN.md) 构建而成.

跟多详情信息请查看 [Resnet50](https://github.com/PaddlePaddle/Serving/tree/v0.6.0/python/examples/imagenet) 和 [中文分词模型](https://github.com/PaddlePaddle/Serving#-pre-built-services-with-paddle-serving).

### 前提条件
- Kubernetes >= 1.18
- 安装 Knative Serving 依赖的网络插件
  请查考 [安装指南](https://knative.dev/v0.21-docs/install/any-kubernetes-cluster/#installing-the-serving-component) 或者执行脚本： `hack/install_knative.sh`(knative serving v0.21 with istio) / `hack/install_knative_kourier.sh`(knative serving v0.22 with kourier).

### 安装

```bash
# 下载 ElasticServing
git clone https://github.com/PaddleFlow/ElasticServing.git
cd ElasticServing

# 安装 CRD
kubectl apply -f assets/crd.yaml

# 安装自定义 Controller
kubectl apply -f assets/elasticserving_operator.yaml
```

### 使用示例

```bash
# 部署 paddle service
kubectl apply -f assets/sample_service.yaml
```

#### 检查服务状态

```bash
# 查看命名空间 paddleservice-system 下的 Service
kubectl get svc -n paddleservice-system

# 查看命名空间 paddleservice-system 下的 knative service
kubectl get ksvc -n paddleservice-system

# 查看命名空间 paddleservice-system 下的 pod
kubectl get pods -n paddleservice-system

# 查看 Paddle Service Pod 的日志信息
kubectl logs <pod-name> -n paddleservice-system -c paddleserving

```

本示例使用 Istio 插件作为 Knative Serving 的网络方案，您也可以使用其他的网络插件比如：Kourier 和 Ambassador。

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

#### Resnet_50_vd 示例
编写 `sample_service.yaml` 如下:

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

##### 输出的结果
```
# 期望的输出结果如下

default:
{"result":{"label":["daisy"],"prob":[0.9341399073600769]}}

canary:
{"result":{"isCanary":["true"],"label":["daisy"],"prob":[0.9341399073600769]}}
```

### 创建你自己的 PaddleService

安装好 CRD ```kubectl apply -f assets/crd.yaml``` 和 Controller ```kubectl apply -f assets/elasticserving_operator.yaml``` 后, 您可以通过编写如下所示的 Yaml 文件来创建 PaddleService。

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

注意：上述 Yaml 文件 Spec 部分只有 `default` 是必填的字段，其他字段可以是为空。如果您自己的 paddleservice 不需要字段 `canary` 和 `canaryTrafficPercent`，可以不填。


执行如下命令来创建 PaddleService

```bash
kubectl apply -f /dir/to/this/yaml/example.yaml
```

## 更多示例

- [BERT](./docs/zh_CN/examples/bert.md)： 语义理解预测服务
- [LAC](./docs/zh_CN/examples/lac.md)： 中文分词模型
- [Criteo Ctr](./docs/zh_CN/examples/criteo_ctr.md)：CTR预估服务
- [Wide & Deep](./docs/zh_CN/examples/wide_deep.md)： Wide & Deep Pipeline

## 更多信息

关于更多自定义资源 PaddleService 的信息，请查看 [API docs](./docs/en/api_doc.md) 文档。

## License

该开源项目遵循 [Apache-2.0 license](https://github.com/PaddleFlow/ElasticServing/blob/main/LICENSE) 协议.
