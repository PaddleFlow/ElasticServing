# Wide & Deep Pipeline

本文档主要讲述了如何使用 Paddle Operator 和 ElasticServing 等各个组件，完成 Wide & Deep 模型的 Pipeline 流程，该流程包括`数据准备-模型训练-模型服务部署`等步骤。
Wide & Deep 模型是 Google 2016 年发布的推荐框架，本示例中使用的模型模型代码由 [PaddleRec 项目](https://github.com/PaddlePaddle/PaddleRec/blob/release/2.1.0/models/rank/wide_deep/README.md) 提供。

## 一、数据准备

本示例使用的是 [Display Advertising Challenge](https://www.kaggle.com/c/criteo-display-ad-challenge/) 提供的 Criteo 数据集，
我们已经将数据存放在[百度对象存储（BOS）](http://baidu.netnic.com.cn/doc/BOS/BOSCLI.html#BOS.20CMD) 公开课访问的 Bucket 中 `bos://paddleflow-public.hkg.bcebos.com/criteo`。
使用 [Paddle Operator](https://github.com/PaddleFlow/paddle-operator) 的样本缓存组件可以将样本数据缓存到集群本地，加速模型训练效率。
如何安装 Paddle Operator 样本缓存组件组件可以参考[快速上手文档](https://github.com/xiaolao/paddle-operator/blob/sampleset/docs/zh_CN/ext-get-start.md)。
在本示例中我们使用 [JuiceFS CSI](https://github.com/juicedata/juicefs-csi-driver) 插件为数据和模型提供存储功能，同时它也是 Paddle Operator 样本缓存组件的一部分。

### 1. 创建对象存储所需的 Secret 对象

本示例使用 BOS 作为存储后端，您也可以使用其他 [JuiceFS 支持的对象存储](https://github.com/juicedata/juicefs/blob/main/docs/zh_cn/databases_for_metadata.md) 。
创建如下的 Secret 对象，样本缓存组件需要其提供的 access-key / secret-key 来访问对象存储, metaurl 是元数据存储引擎的访问链接。

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

注意：data 中的各个字段信息需要用 base64 进行编码

### 2. 创建 SampleSet

Paddle Operator 样本缓存组件通过自定义 CRD SampleSet 提供了样本数据集的抽象，方便用户管理样本数据。
创建如下的 SampleSet 并等待数据完成同步，配置中的 nodeAffinity 可以来用指定数据需要缓存到那些节点， 比如可以指定将数据缓存到 GPU 节点。

```yaml
# criteo-sampleset.yaml
apiVersion: batch.paddlepaddle.org/v1alpha1
kind: SampleSet
metadata:
  name: criteo
  namespace: paddle-system
spec:
  # 缓存分区，一台宿主机代表一个分区
  partitions: 1
  source:
    # 样本数据源
    uri: bos://paddleflow-public.hkg.bcebos.com/criteo
    secretRef:
      # 这里填写上诉创建的 Secret
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

由于样本数据有22G，创建好 SampleSet 后可能需要等待一段时间，等其状态变更为 Ready 后即可开始进行模型训练的步骤了。

```bash
$ kubectl apply -f criteo-sampleset.yaml
sampleset.batch.paddlepaddle.org/imagenet created

$ kubectl get sampleset criteo -n paddle-system
NAME     TOTAL SIZE   CACHED SIZE   AVAIL SPACE   RUNTIME   PHASE   AGE
criteo   22 GiB       22 GiB        9.4 GiB       1/1       Ready   2d6h
```

### 3. 准备用于存储模型的 Volume (可选)

这个步骤主要是创建 PV 和 PVC 资源对象用于存储后续步骤产出的模型，本示例采用 [JuiceFS CSI 静态绑定的方式](https://github.com/juicedata/juicefs-csi-driver/tree/master/examples/static-provisioning) 
来创建 PV 和 PVC，存储后端依然是 BOS。 你也可以使用其他的存储引擎的 CSI 插件， 比如 Ceph / Glusterfs 等。

创建 PV 
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
    # 访问远程对象存储所需的 Secret
    nodePublishSecretRef:
      name: criteo
      namespace: paddle-system
    # 对象存储的 Bucket 路径，用于存放模型文件  
    volumeHandle: model-center
  persistentVolumeReclaimPolicy: Retain
  storageClassName: model-center
  volumeMode: Filesystem
```

创建 PVC
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

## 二、模型训练

Wide & Deep 模型的实现代码来自项目 [PaddleRec](https://github.com/PaddlePaddle/PaddleRec/blob/release/2.1.0/models/rank/wide_deep/README.md) ，
我们提供了该项目的镜像并放置在公开的镜像仓库中：`registry.baidubce.com/paddleflow-public/paddlerec:2.1.0-gpu-cuda10.2-cudnn7`。
本示例使用 [Collective](https://fleet-x.readthedocs.io/en/latest/paddle_fleet_rst/collective/index.html) 的模式进行训练，所以要求要有 GPU 设备。
你也可以在采用 [Parameter Server](https://fleet-x.readthedocs.io/en/latest/paddle_fleet_rst/paddle_on_k8s.html#wide-and-deep) 模式来进行模型训练。
PaddleRec 的 cpu 版本的镜像地址为：`registry.baidubce.com/paddleflow-public/paddlerec:2.1.0`。

### 1. 创建 ConfigMap

PaddleRec 项目各个模型可以通过配置文件来指定模型的超参和模型训练配置，故通过 ConfigMap 的方式将配置文件挂载进训练容器，可以比较方便的修改配置文件。
关于配置文件中各个字段含义可以参考文档：[PaddleRec config.yaml 配置说明](https://github.com/PaddlePaddle/PaddleRec/blob/release/2.1.0/doc/yaml.md)

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

创建 wide-deep-config ConfigMap

```bash
kubectl create configmap wide-deep-config -n paddle-system --from-file=wide_deep_config.yaml 
```

### 2. 创建 PaddleJob

PaddleJob 是 Paddle Operator 项目中的自定义资源，用于定义 Paddle 模型训练作业。

```yaml
# wide-deep.yaml
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: wide-deep
  namespace: paddle-system
spec:
  cleanPodPolicy: Never
  # 申明要使用的数据集
  sampleSetRef:
    name: criteo
    # 数据集在容器内的挂盘路径
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
            # 将宿主机内存挂载进容器内，防止容器内OOM出错
            - mountPath: /dev/shm
              name: dshm
            # 将 ConfigMap 挂载进容器内
            - mountPath: /mnt/config
              name: config-volume
            # 用于存储模型
            - mountPath: /mnt/model
              name: model-volume
            resources:
              limits:
                # 每个宿主机上使用两块GPU设备
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

创建 PaddleJob 并查看任务状态

```bash
$ kubectl create -f wide-deep.yaml

$ kubectl get paddlejob wide-deep -n paddle-system
NAME        STATUS   MODE        WORKER   AGE
wide-deep   Running  Collective  2/2      2m
```

## 三、模型服务

等模型训练完后，模型存储路径中就可以看到数字编号的目录，如 `0/`。这是因为模型训练过程中，每个 epoch 都会保存一份模型快照，
故用数字最大的文件夹中的模型文件来部署服务即可。在使用 Paddle Serving 部署模型前，需要对训练文件完成转换工作方可进行服务部署。
我们将模型文件存放在公开可访问的 bucket 中：`https://paddleflow-public.hkg.bcebos.com/models/wide-deep/wide-deep.tar.gz` 。

解压后模型文件目录结构如下：
```
.
├── rec_inference.pdiparams
├── rec_inference.pdmodel
├── rec_static.pdmodel
├── rec_static.pdopt
└── rec_static.pdparams
```

### 1. 创建 PaddleService

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

### 2. 查看服务状态

```bash
# 查看命名空间 paddleservice-system 下的 Service
kubectl get svc -n paddleservice-system | grep wide-deep-serving

# 查看命名空间 paddleservice-system 下 knative service 的状态
kubectl get ksvc wide-deep-serving -n paddleservice-system

# 查看命名空空间 paddleservice-system 下所有的 pod
kubectl get pods -n paddleservice-system

# 运行以下命令获取 ClusterIP
kubectl get svc wide-deep-serving-default-private -n paddleservice-system
```

模型服务支持 HTTP / BRPC / GRPC 三种客户端访问，客户端代码和环境配置详情请查看文档: [在线Serving部署](https://github.com/PaddlePaddle/PaddleRec/blob/release/2.1.0/doc/serving.md) 。
需要注意的是，您需要在客户端代码中将服务 IP 地址和端口替换成上述 wide-deep-serving-default-private service 的 cluster-ip 和端口。