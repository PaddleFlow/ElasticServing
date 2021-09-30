# Wide & Deep Pipeline

This document mainly describes how to use components such as Paddle Operator and ElasticServing to complete the Pipeline of the Wide & Deep model, which includes the steps of `data preparation, model training and model serving` .
The Wide & Deep model is a recommended framework published by Google in 2016. The model code used in this demo is from [PaddleRec Project](https://github.com/PaddlePaddle/PaddleRec/blob/release/2.1.0/models/rank/wide_deep/README.md) .

## 1. Data Preparation

This demo uses the Criteo dataset provided by [Display Advertising Challenge](https://www.kaggle.com/c/criteo-display-ad-challenge/), 
And We have stored the data in the public bucket of [Baidu Object Storage (BOS)](http://baidu.netnic.com.cn/doc/BOS/BOSCLI.html#BOS.20CMD): `bos:// paddleflow-public.hkg.bcebos.com/criteo`.
Use the cache component of [Paddle Operator](https://github.com/PaddleFlow/paddle-operator) to cache sample data locally can speed up model training jobs.
Please refer to [quick start document](https://github.com/xiaolao/paddle-operator/blob/sampleset/docs/zh_CN/ext-get-start.md) to install Paddle Operator.
In this example, we use the [JuiceFS CSI](https://github.com/juicedata/juicefs-csi-driver) plugin to store data and models, and it is also part of the Paddle Operator.

### Create Secret for Access Object Storage

This example uses BOS as the storage backend. You can also use other [JuiceFS supported object storage](https://github.com/juicedata/juicefs/blob/main/docs/zh_cn/databases_for_metadata.md).
Create Secret as following. The cache component of Paddle Operator needs the access-key / secret-key to access the object storage. And the metaurl is the access link of the metadata storage engine.

```yaml
# criteo-secret.yaml
apiVersion: v1
data:
  access-key: xxx
  bucket: xxx
  metaurl: xxx
  name: Y3JpdGVv
  secret-key: xxx
  storage: Ym9z
kind: Secret
metadata:
  name: criteo
  namespace: paddle
type: Opaque
```

Note: Each field in `data` needs to be encoded with base64

### Create SampleSet

The cache component of Paddle Operator provides an abstraction of sample data sets through CRD: SampleSet. Create the following SampleSet and wait for the data to be synchronized. 
The nodeAffinity in the configuration can be used to specify which nodes the data should be cached to, for example, you can specify to cache the data to nodes with GPU device.

```yaml
# criteo-sampleset.yaml
apiVersion: batch.paddlepaddle.org/v1alpha1
kind: SampleSet
metadata:
  name: criteo
  namespace: paddle-system
spec:
  # Partitions of cache data
  partitions: 1
  source:
    # Uri of sample data source
    uri: bos://paddleflow-public.hkg.bcebos.com/criteo
    secretRef:
      # Secret to access data source
      name: criteo-source
  secretRef:
    name: criteo
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
        - matchExpressions:
            - key: beta.kubernetes.io/instance-gpu
              operator: In
              values:
                - "true"
```

Since the sample data has a size of 22G, it may take a while after the SampleSet is created, 
and the model training can be started after SampleSet's status is changed to Ready.

```bash
$ kubectl apply -f criteo-sampleset.yaml
sampleset.batch.paddlepaddle.org/imagenet created

$ kubectl get sampleset criteo -n paddle-system
NAME     TOTAL SIZE   CACHED SIZE   AVAIL SPACE   RUNTIME   PHASE   AGE
criteo   22 GiB       22 GiB        9.4 GiB       1/1       Ready   2d6h
```

### Prepare Volume for Storing Model (Optional)

This step is mainly to create PV and PVC resource to store the models. We create PV and PVC by [JuiceFS CSI](https://github.com/juicedata/juicefs-csi-driver/tree/master/examples/static-provisioning) ,
and the storage backend is still BOS. You can also use other CSI plugins, such as Ceph / Glusterfs and so on.

Create PV
```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: model-center
spec:
  accessModes:
    - ReadWriteMany
  capacity:
    storage: 10Pi
  csi:
    driver: csi.juicefs.com
    fsType: juicefs
    # Secret to access remote object storage
    nodePublishSecretRef:
      name: criteo
      namespace: paddle-system
    # Bucket of object storage, the model files is store in this path
    volumeHandle: model-center
  persistentVolumeReclaimPolicy: Retain
  storageClassName: model-center
  volumeMode: Filesystem
```

Create PVC
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: model-center
  namespace: paddle-system
spec:
  accessModes:
  - ReadWriteMany
  resources:
    requests:
      storage: 10Pi
  storageClassName: model-center
  volumeMode: Filesystem
```

## 2. Training Model

The model training script of this example is from [PaddleRec project](https://github.com/PaddlePaddle/PaddleRec/blob/release/2.1.0/models/rank/wide_deep/README.md),
and image is placed in: `registry.baidubce.com/paddleflow-public/paddlerec:2.1.0-gpu-cuda10.2-cudnn7`.
This example uses the [Collective](https://fleet-x.readthedocs.io/en/latest/paddle_fleet_rst/collective/index.html) mode for training, so a GPU device is required.
You can also use [Parameter Server](https://fleet-x.readthedocs.io/en/latest/paddle_fleet_rst/paddle_on_k8s.html#wide-and-deep) mode for model training.
And the cpu version image of PaddleRec is placed: `registry.baidubce.com/paddleflow-public/paddlerec:2.1.0`.

### 1. Create ConfigMap

Each model of the PaddleRec project has a config file. In this demo we use configmap to store configurations and mount it into containers of PaddleJob,
so we modify the configurations more convenient. Please refer to the document: [PaddleRec config.yaml Configuration Instructions](https://github.com/PaddlePaddle/PaddleRec/blob/release/2.1.0/doc/yaml.md).

```yaml
# wide_deep_config.yaml
# global settings
runner:
  train_data_dir: "/mnt/criteo/slot_train_data_full"
  train_reader_path: "criteo_reader" # importlib format
  use_gpu: True
  use_auc: True
  train_batch_size: 4096
  epochs: 4
  print_interval: 10
  model_save_path: "/mnt/model"
  test_data_dir: "/mnt/criteo/slot_test_data_full"
  infer_reader_path: "criteo_reader" # importlib format
  infer_batch_size: 4096
  infer_load_path: "/mnt/model"
  infer_start_epoch: 0
  infer_end_epoch: 4
  use_inference: True
  save_inference_feed_varnames: ["C1","C2","C3","C4","C5","C6","C7","C8","C9","C10","C11","C12","C13","C14","C15","C16","C17","C18","C19","C20","C21","C22","C23","C24","C25","C26","dense_input"]
  save_inference_fetch_varnames: ["sigmoid_0.tmp_0"]
  #use fleet
  use_fleet: True

# hyper parameters of user-defined network
hyper_parameters:
  # optimizer config
  optimizer:
    class: Adam
    learning_rate: 0.001
    strategy: async
  # user-defined <key, value> pairs
  sparse_inputs_slots: 27
  sparse_feature_number: 1000001
  sparse_feature_dim: 9
  dense_input_dim: 13
  fc_sizes: [512, 256, 128, 32]
  distributed_embedding: 0
```

Create ConfigMap named as wide-deep-config

```bash
kubectl create configmap wide-deep-config -n paddle-system --from-file=wide_deep_config.yaml 
```

### 2. Create PaddleJob

PaddleJob is a custom resource in the Paddle Operator project, used to define Paddle training jobs.

```yaml
# wide-deep.yaml
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: wide-deep
  namespace: paddle-system
spec:
  cleanPodPolicy: Never
  sampleSetRef:
    name: criteo
    mountPath: /mnt/criteo
  worker:
    replicas: 1
    template:
      spec:
        containers:
          - name: paddlerec
            image: registry.baidubce.com/paddleflow-public/paddlerec:2.1.0-gpu-cuda10.2-cudnn7
            workingDir: /home/PaddleRec/models/rank/wide_deep
            command: ["/bin/bash", "-c", "cp /mnt/config/wide_deep_config.yaml . && mkdir -p /mnt/model/wide-deep && python -m paddle.distributed.launch --log_dir /mnt/model/log --gpus '0,1' ../../../tools/trainer.py -m wide_deep_config.yaml"]
            volumeMounts:
            - mountPath: /dev/shm
              name: dshm
            - mountPath: /mnt/config
              name: config-volume
            - mountPath: /mnt/model
              name: model-volume
            resources:
              limits:
                nvidia.com/gpu: 2
        volumes:
        - name: dshm
          emptyDir:
            medium: Memory
        - name: config-volume
          configMap:
            name: wide-deep-config
        - name: model-volume
          persistentVolumeClaim:
            claimName: model-center
```

Create PaddleJob and check it status

```bash
$ kubectl create -f wide-deep.yaml

$ kubectl get paddlejob wide-deep -n paddle-system
NAME        STATUS   MODE        WORKER   AGE
wide-deep   Running  Collective  2/2      2m
```

## 3. Model Serving

After PaddleJob is finished, you may notice that the model file stored in numeric dir, such as `0/`. During the model training process, Paddle will save a checkpoint after each epoch is done,
so you use the model files in the folder with the largest number to deploy service. Before deploy model serving, you need to convert the model file the format that Paddle Serving can use.
And we have already placed model files in bucket: `https://paddleflow-public.hkg.bcebos.com/models/wide-deep/wide-deep.tar.gz`.

The directory structure:
```
.
├── rec_inference.pdiparams
├── rec_inference.pdmodel
├── rec_static.pdmodel
├── rec_static.pdopt
└── rec_static.pdparams
```

### 1. Create PaddleService

```yaml
apiVersion: elasticserving.paddlepaddle.org/v1
kind: PaddleService
metadata:
  name: wide-deep-serving
  namespace: paddleservice-system
spec:
  default:
    arg: wget https://paddleflow-public.hkg.bcebos.com/models/wide-deep/wide-deep.tar.gz &&
      tar xzf wide-deep.tar.gz && rm -rf wide-deep.tar.gz &&
      python3 -m paddle_serving_client.convert --dirname wide-deep/ --model_filename rec_inference.pdmodel --params_filename rec_inference.pdiparams &&
      python3 -m paddle_serving_server.serve --model serving_server --port 9292
    containerImage: registry.baidubce.com/paddleflow-public/serving
    port: 9292
    tag: v0.6.2
  runtimeVersion: paddleserving
  service:
    minScale: 1
```

### 2. Check the Service Status

```bash
# Check service in namespace paddleservice-system
kubectl get svc -n paddleservice-system | grep paddleservice-criteoctr

# Check knative service in namespace paddleservice-system
kubectl get ksvc paddleservice-criteoctr -n paddleservice-system

# Check pods in namespace paddleservice-system
kubectl get pods -n paddleservice-system

# Obtain ClusterIP
kubectl get svc wide-deep-serving-default-private -n paddleservice-system
```

The model service supports HTTP / BRPC / GRPC client, refer to [Serving](https://github.com/PaddlePaddle/PaddleRec/blob/release/2.1.0/doc/serving.md)
to obtain the code of client. It should be noted that you need to replace the service IP address and port in the client code with the cluster-ip and port from the `wide-deep-serving-default-private` service mentioned above.
