# Chinese Word Segmentation

English | [简体中文](../../zh_CN/examples/lac.md)

This example uses the Chinese word segmentation model for service deployment. For more detail information please refer to [Paddle Serving](https://github.com/PaddlePaddle/Serving/blob/develop/python/examples/lac/README.md)

## Build Image for Service (Optional)

The test service is build on `registry.baidubce.com/paddlepaddle/serving:0.6.0-devel`, and have published in public image register `registry.baidubce.com/paddleflow-public/lac-serving:latest`.
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

WORKDIR /home/Serving/python/examples/lac

RUN python3 -m paddle_serving_app.package --get_model lac && \
    tar xzf lac.tar.gz && rm -rf lac.tar.gz

ENTRYPOINT ["python3", "-m", "paddle_serving_server.serve", "--model", "lac_model/", "--port", "9292"]
```

3. Build Image

```bash
docker build . -t registry.baidubce.com/paddleflow-public/lac-serving:latest
```

## Create PaddleService

1. Prepare YAML File

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

2. Create PaddleService

```bash
$ kubectl apply -f lac.yaml
paddleservice.elasticserving.paddlepaddle.org/paddleservice-lac created
```

## Check The Status of Service

1. check the status of model service

```bash
# Check service in namespace paddleservice-system
kubectl get svc -n paddleservice-system | grep paddleservice-lac

# Check knative service in namespace paddleservice-system
kubectl get ksvc paddleservice-lac -n paddleservice-system

# Check pods in namespace paddleservice-system
kubectl get pods -n paddleservice-system
```

2. Obtain ClusterIP
```bash
$ kubectl get svc paddleservice-lac-default-private -n paddleservice-system
```

## Test Model Service

The model service supports HTTP / BRPC / GRPC client, refer to [lac service](https://github.com/PaddlePaddle/Serving/blob/develop/python/examples/lac/README.md)

You can simply test the service through the following command
```bash
# Note: the IP-address and Port should be replaced with the cluster-ip and port of the paddleservice-criteoctr-default-private service mentioned above.
curl -H "Host: paddleservice-sample.paddleservice-system.example.com" -H "Content-Type:application/json" -X POST -d '{"feed":[{"words": "我爱北京天安门"}], "fetch":["word_seg"]}' http://<IP-address>:<Port>/lac/prediction
```

Expected result
```bash
{"result":[{"word_seg":"\u6211|\u7231|\u5317\u4eac|\u5929\u5b89\u95e8"}]}
```
