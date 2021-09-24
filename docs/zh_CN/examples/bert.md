# BERT 语义理解服务

[English](../../en/examples/bert.md) | 简体中文

本示例采用 BERT 预训练模型进行文本分析与预测服务的部署，更多模型和代码的详情信息可以查看 [Paddle Serving](https://github.com/PaddlePaddle/Serving/blob/develop/python/examples/bert/README_CN.md).

## 构建服务镜像（可选）

本示例模型服务镜像基于 `registry.baidubce.com/paddlepaddle/serving:0.6.0-devel` 构建而成，并上传到公开可访问的镜像仓库 `registry.baidubce.com/paddleflow-public/bert-serving:latest` 。
如您需要 GPU 或其他版本的基础镜像，可以查看文档 [Docker 镜像](https://github.com/PaddlePaddle/Serving/blob/v0.6.0/doc/DOCKER_IMAGES_CN.md), 并按照如下步骤构建镜像。

1. 下载 Paddle Serving 代码

```bash
$ wget https://github.com/PaddlePaddle/Serving/archive/refs/tags/v0.6.0.tar.gz
$ tar xzvf Serving-0.6.0.tar.gz
$ mv Serving-0.6.0 Serving
$ cd Serving
```

2. 编写如下 Dockerfile

```Dockerfile
FROM registry.baidubce.com/paddlepaddle/serving:0.6.0-devel

WORKDIR /home

COPY . /home/Serving

WORKDIR /home/Serving

# install depandences
RUN pip install -r python/requirements.txt -i https://pypi.tuna.tsinghua.edu.cn/simple && \
    pip install paddle-serving-server==0.6.0 -i https://pypi.tuna.tsinghua.edu.cn/simple && \
    pip install paddle-serving-client==0.6.0 -i https://pypi.tuna.tsinghua.edu.cn/simple

WORKDIR /home/Serving/python/examples/bert

# download pre-trained BERT model
RUN wget https://paddle-serving.bj.bcebos.com/paddle_hub_models/text/SemanticModel/bert_chinese_L-12_H-768_A-12.tar.gz && \
    tar -xzf bert_chinese_L-12_H-768_A-12.tar.gz && rm -rf bert_chinese_L-12_H-768_A-12.tar.gz && \
    mv bert_chinese_L-12_H-768_A-12_model bert_seq128_model && mv bert_chinese_L-12_H-768_A-12_client bert_seq128_client

ENTRYPOINT ["python3", "-m", "paddle_serving_server.serve", "--model", "bert_seq128_model/", "--port", "9292"]
```

3. 构建镜像

```bash
docker build . -t registry.baidubce.com/paddleflow-public/bert-serving:latest
```

## 创建 PaddleService

1. 编写 YAML 文件

```yaml
# bert.yaml
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
  name: paddleservice-bert
  namespace: paddleservice-system
spec:
  default:
    arg: python3 -m paddle_serving_server.serve --model bert_seq128_model/ --port 9292
    containerImage: registry.baidubce.com/paddleflow-public/bert-serving
    port: 9292
    tag: latest
  runtimeVersion: paddleserving
  service:
    minScale: 1
```

2. 创建 PaddleService

```bash
$ kubectl apply -f bert.yaml
paddleservice.elasticserving.paddlepaddle.org/paddleservice-bert created
```

## 查看服务状态

1. 您可以使用以下命令查看服务状态

```bash
# Check service in namespace paddleservice-system
kubectl get svc -n paddleservice-system | grep paddleservice-bert

# Check knative service in namespace paddleservice-system
kubectl get ksvc paddleservice-bert -n paddleservice-system

# Check pods in namespace paddleservice-system
kubectl get pods -n paddleservice-system
```

2. 运行以下命令获取 ClusterIP
```bash
$ kubectl get svc paddleservice-bert-default-private -n paddleservice-system
```

## 测试 BERT 模型服务

模型服务支持 HTTP / BRPC / GRPC 三种客户端访问，客户端代码和环境配置详情请查看文档 [语义理解预测服务](https://github.com/PaddlePaddle/Serving/blob/develop/python/examples/bert/README_CN.md) 。
需要注意的是，您需要在客户端代码中将服务 IP 地址和端口替换成上述 paddleservice-bert-default-private service 的 cluster-ip 和端口。

例如修改 `bert_client.py` 的代码

```python
  fetch = ["pooled_output"]
- endpoint_list = ['127.0.0.1:9292']
+ endpoint_list = ['172.16.237.0:80']
  client = Client()
```
