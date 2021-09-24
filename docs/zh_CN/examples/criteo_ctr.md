# CTR预测服务

[English](../../en/examples/criteo_ctr.md) | 简体中文

本示例是基于 criteo 数据集训练的 ctr 预估服务，更多模型和代码的详情信息可以查看 [Paddle Serving](https://github.com/PaddlePaddle/Serving/blob/develop/python/examples/criteo_ctr/README_CN.md).

## 构建服务镜像（可选）

本示例模型服务镜像基于 `registry.baidubce.com/paddlepaddle/serving:0.6.0-devel` 构建而成，并上传到公开可访问的镜像仓库 `registry.baidubce.com/paddleflow-public/criteoctr-serving:latest` 。
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

WORKDIR /home/Serving/python/examples/criteo_ctr

RUN wget https://paddle-serving.bj.bcebos.com/criteo_ctr_example/criteo_ctr_demo_model.tar.gz && \
    tar xzf criteo_ctr_demo_model.tar.gz && rm -rf criteo_ctr_demo_model.tar.gz && \
    mv models/ctr_client_conf . && mv models/ctr_serving_model .

ENTRYPOINT ["python3", "-m", "paddle_serving_server.serve", "--model", "ctr_serving_model/", "--port", "9292"]
```

3. 构建镜像

```bash
docker build . -t registry.baidubce.com/paddleflow-public/criteoctr-serving:latest
```

## 创建 PaddleService

1. 编写 YAML 文件

```yaml
# criteoctr.yaml
apiVersion: elasticserving.paddlepaddle.org/v1
kind: PaddleService
metadata:
  name: paddleservice-criteoctr
  namespace: paddleservice-system
spec:
  default:
    arg: python3 -m paddle_serving_server.serve --model ctr_serving_model/ --port 9292
    containerImage: registry.baidubce.com/paddleflow-public/criteoctr-serving
    port: 9292
    tag: latest
  runtimeVersion: paddleserving
  service:
    minScale: 1
```

2. 创建 PaddleService

```bash
$ kubectl apply -f criteoctr.yaml
paddleservice.elasticserving.paddlepaddle.org/paddleservice-criteoctr created
```

## 查看服务状态

1. 您可以使用以下命令查看服务状态

```bash
# Check service in namespace paddleservice-system
kubectl get svc -n paddleservice-system | grep paddleservice-criteoctr

# Check knative service in namespace paddleservice-system
kubectl get ksvc paddleservice-criteoctr -n paddleservice-system

# Check pods in namespace paddleservice-system
kubectl get pods -n paddleservice-system
```

2. 运行以下命令获取 ClusterIP
```bash
$ kubectl get svc paddleservice-criteoctr-default-private -n paddleservice-system
```

## 测试模型服务

模型服务支持 HTTP / BRPC / GRPC 三种客户端访问，客户端代码和环境配置详情请查看文档 [语义理解预测服务](https://github.com/PaddlePaddle/Serving/blob/develop/python/examples/criteo_ctr/README_CN.md) 。
需要注意的是，您需要在客户端代码中将服务 IP 地址和端口替换成上述 paddleservice-criteoctr-default-private service 的 cluster-ip 和端口。

例如修改 `test_client.py` 的代码

```python
 client.load_client_config(sys.argv[1])
- client.connect(["127.0.0.1:9292"])
+ client.connect(["172.16.183.200:80"])
 reader = CriteoReader(1000001)
```
