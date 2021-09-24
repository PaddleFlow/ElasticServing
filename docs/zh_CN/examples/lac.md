# 中文分词模型服务

[English](../../en/examples/lac.md) | 简体中文

本示例采用 lac 中文分词模型来做服务部署，更多模型和代码的详情信息可以查看 [Paddle Serving](https://github.com/PaddlePaddle/Serving/blob/develop/python/examples/lac/README_CN.md).

## 构建服务镜像（可选）

本示例模型服务镜像基于 `registry.baidubce.com/paddlepaddle/serving:0.6.0-devel` 构建而成，并上传到公开可访问的镜像仓库 `registry.baidubce.com/paddleflow-public/lac-serving:latest` 。
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

WORKDIR /home/Serving/python/examples/lac

RUN python3 -m paddle_serving_app.package --get_model lac && \
    tar xzf lac.tar.gz && rm -rf lac.tar.gz

ENTRYPOINT ["python3", "-m", "paddle_serving_server.serve", "--model", "lac_model/", "--port", "9292"]
```

3. 构建镜像

```bash
docker build . -t registry.baidubce.com/paddleflow-public/lac-serving:latest
```

## 创建 PaddleService

1. 编写 YAML 文件

```yaml
# lac.yaml
apiVersion: elasticserving.paddlepaddle.org/v1
kind: PaddleService
metadata:
  name: paddleservice-sample
  namespace: paddleservice-system
spec:
  default:
    arg: python3 -m paddle_serving_server.serve --model lac_model/ --port 9292
    containerImage: registry.baidubce.com/paddleflow-public/lac-serving
    port: 9292
    tag: latest
  runtimeVersion: paddleserving
  service:
    minScale: 1
```

2. 创建 PaddleService

```bash
$ kubectl apply -f lac.yaml
paddleservice.elasticserving.paddlepaddle.org/paddleservice-lac created
```

## 查看服务状态

1. 您可以使用以下命令查看服务状态

```bash
# Check service in namespace paddleservice-system
kubectl get svc -n paddleservice-system | grep paddleservice-lac

# Check knative service in namespace paddleservice-system
kubectl get ksvc paddleservice-lac -n paddleservice-system

# Check pods in namespace paddleservice-system
kubectl get pods -n paddleservice-system
```

2. 运行以下命令获取 ClusterIP
```bash
$ kubectl get svc paddleservice-lac-default-private -n paddleservice-system
```

## 测试 BERT 模型服务

模型服务支持 HTTP / BRPC / GRPC 三种客户端访问，客户端代码和环境配置详情请查看文档 [中文分词模型服务](https://github.com/PaddlePaddle/Serving/blob/develop/python/examples/lac/README_CN.md) 。

通过以下命令可以简单测试下服务是否正常
```bash
# 注意将 IP-address 和 Port 替换成上述 paddleservice-criteoctr-default-private service 的 cluster-ip 和端口。
curl -H "Host: paddleservice-sample.paddleservice-system.example.com" -H "Content-Type:application/json" -X POST -d '{"feed":[{"words": "我爱北京天安门"}], "fetch":["word_seg"]}' http://<IP-address>:<Port>/lac/prediction
```

预期结果
```bash
{"result":[{"word_seg":"\u6211|\u7231|\u5317\u4eac|\u5929\u5b89\u95e8"}]}
```
