# CTR Prediction Service

English | [简体中文](../../zh_CN/examples/criteo_ctr.md)

This example is based on the ctr prediction service trained on the criteo dataset. For more detail information please refer to [Paddle Serving](https://github.com/PaddlePaddle/Serving/blob/develop/python/examples/criteo_ctr/README.md)

## Build Image for Service (Optional)

The test service is build on `registry.baidubce.com/paddlepaddle/serving:0.6.0-devel`, and have published in public image register `registry.baidubce.com/paddleflow-public/criteoctr-serving:latest`.
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

WORKDIR /home/Serving/python/examples/criteo_ctr

RUN wget https://paddle-serving.bj.bcebos.com/criteo_ctr_example/criteo_ctr_demo_model.tar.gz && \
    tar xzf criteo_ctr_demo_model.tar.gz && rm -rf criteo_ctr_demo_model.tar.gz && \
    mv models/ctr_client_conf . && mv models/ctr_serving_model .

ENTRYPOINT ["python3", "-m", "paddle_serving_server.serve", "--model", "ctr_serving_model/", "--port", "9292"]
```

3. Build Image

```bash
docker build . -t registry.baidubce.com/paddleflow-public/criteoctr-serving:latest
```

## Create PaddleService

1. Prepare YAML File

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

2. Create PaddleService

```bash
$ kubectl apply -f criteoctr.yaml
paddleservice.elasticserving.paddlepaddle.org/paddleservice-criteoctr created
```

## Check The Status of Service

1. check the status of criteo ctr model service

```bash
# Check service in namespace paddleservice-system
kubectl get svc -n paddleservice-system | grep paddleservice-criteoctr

# Check knative service in namespace paddleservice-system
kubectl get ksvc paddleservice-criteoctr -n paddleservice-system

# Check pods in namespace paddleservice-system
kubectl get pods -n paddleservice-system
```

2. Obtain ClusterIP
```bash
$ kubectl get svc paddleservice-criteoctr-default-private -n paddleservice-system
```

## Test Model Service

The model service supports HTTP / BRPC / GRPC client, refer to [criteo ctr service](https://github.com/PaddlePaddle/Serving/tree/develop/python/examples/criteo_ctr)
to obtain the code of client. It should be noted that you need to replace the service IP address and port in the client code with the cluster-ip and port from the paddleservice-criteoctr-default-private service mentioned above.

For example, modify the code of `test_client.py` as follows
```python
 client.load_client_config(sys.argv[1])
- client.connect(["127.0.0.1:9292"])
+ client.connect(["172.16.183.200:80"])
 reader = CriteoReader(1000001)
```
