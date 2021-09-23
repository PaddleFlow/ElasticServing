#### Lac sample
You can also change the `image` and `arg` like the following to serve different models.
``` yaml
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
  default:
    arg: python3 Serving/python/examples/lac/lac_web_service.py lac_model/ lac_workdir
      9292
    containerImage: registry.baidubce.com/paddleflow-public/lac-serving
    port: 9292
    tag: latest
  runtimeVersion: paddleserving
  service:
    minScale: 1
```

``` bash
# Start to send data to the server. <IP-address> is what has been got in the first or the second command.
curl -H "Host: paddleservice-sample.paddleservice-system.example.com" -H "Content-Type:application/json" -X POST -d '{"feed":[{"words": "我爱北京天安门"}], "fetch":["word_seg"]}' http://<IP-address>:<Port>/lac/prediction
```

##### Expected Result

``` bash
# The expected output should be 

default: 
{"result":[{"word_seg":"\u6211|\u7231|\u5317\u4eac|\u5929\u5b89\u95e8"}]}

canary:
{"result":[{"word_seg":"\u6211-\u7231-\u5317\u4eac-\u5929\u5b89\u95e8"}]}
```