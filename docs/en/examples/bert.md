# BERT AS Service

English | [简体中文](../../zh_CN/examples/bert.md)

This example uses the BERT pre-training model to deploy text analysis and prediction services. For more detail information please refer to [Paddle Serving](https://github.com/PaddlePaddle/Serving/blob/develop/python/examples/bert/README.md)

## Build Image for Service (Optional)

The test service is build on `registry.baidubce.com/paddlepaddle/serving:0.6.0-devel`, and have published in public image register `registry.baidubce.com/paddleflow-public/bert-serving:latest`.
If you need run the service in GPU or other device, please refer to [Docker Images](https://github.com/PaddlePaddle/Serving/blob/v0.6.0/doc/DOCKER_IMAGES_CN.md) and build the model server image as below.

1. Download `Paddle Serving`

```bash
$ wget https://github.com/PaddlePaddle/Serving/archive/refs/tags/v0.6.0.tar.gz
$ tar xzvf Serving-0.6.0.tar.gz
$ mv Serving-0.6.0 Serving
$ cd Serving
```

2. Write Dockerfile

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

3. Build Image

```bash
docker build . -t registry.baidubce.com/paddleflow-public/bert-serving:latest
```

## Create PaddleService

1. Prepare YAML File

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

2. Create PaddleService

```bash
$ kubectl apply -f bert.yaml
paddleservice.elasticserving.paddlepaddle.org/paddleservice-bert created
```

## Check The Status of Service

1. check the status of bert model service

```bash
# Check service in namespace paddleservice-system
kubectl get svc -n paddleservice-system | grep paddleservice-bert

# Check knative service in namespace paddleservice-system
kubectl get ksvc paddleservice-bert -n paddleservice-system

# Check pods in namespace paddleservice-system
kubectl get pods -n paddleservice-system
```

2. Obtain ClusterIP
```bash
$ kubectl get svc paddleservice-bert-default-private -n paddleservice-system
```

## Test The BERT Model Service

The model service supports HTTP / BRPC / GRPC client, refer to [bert service](https://github.com/PaddlePaddle/Serving/blob/develop/python/examples/bert/README_CN.md)
to obtain the code of client. It should be noted that you need to replace the service IP address and port in the client code with the cluster-ip and port from the paddleservice-bert-default-private service mentioned above.

For example, modify the code of `bert_client.py` as follows
```python
  fetch = ["pooled_output"]
- endpoint_list = ['127.0.0.1:9292']
+ endpoint_list = ['172.16.237.0:80']
  client = Client()
```
