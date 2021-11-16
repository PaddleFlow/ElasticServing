<p>Packages:</p>
<ul>
<li>
<a href="#elasticserving.paddlepaddle.org%2fv1">elasticserving.paddlepaddle.org/v1</a>
</li>
</ul>
<h2 id="elasticserving.paddlepaddle.org/v1">elasticserving.paddlepaddle.org/v1</h2>
<p>
<p>Package v1 contains PaddleService</p>
</p>
Resource Types:
<ul></ul>
<h3 id="elasticserving.paddlepaddle.org/v1.Autoscaler">Autoscaler
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#elasticserving.paddlepaddle.org/v1.ServiceSpec">ServiceSpec</a>)
</p>
<p>
<p>Autoscaler defines the autoscaler class</p>
</p>
<h3 id="elasticserving.paddlepaddle.org/v1.AutoscalerMetric">AutoscalerMetric
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#elasticserving.paddlepaddle.org/v1.ServiceSpec">ServiceSpec</a>)
</p>
<p>
<p>AutoscalerMetric defines the metric for the autoscaler</p>
</p>
<h3 id="elasticserving.paddlepaddle.org/v1.EndpointSpec">EndpointSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#elasticserving.paddlepaddle.org/v1.PaddleServiceSpec">PaddleServiceSpec</a>)
</p>
<p>
<p>EndpointSpec defines the running containers</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>containerImage</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>tag</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>port</code></br>
<em>
int32
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>arg</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</tbody>
</table>
<h3 id="elasticserving.paddlepaddle.org/v1.PaddleService">PaddleService
</h3>
<p>
<p>PaddleService is the Schema for the paddles API</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#elasticserving.paddlepaddle.org/v1.PaddleServiceSpec">
PaddleServiceSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>runtimeVersion</code></br>
<em>
string
</em>
</td>
<td>
<p>Version of the service</p>
</td>
</tr>
<tr>
<td>
<code>resources</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<p>Defaults to requests and limits of 1CPU, 2Gb MEM.</p>
</td>
</tr>
<tr>
<td>
<code>default</code></br>
<em>
<a href="#elasticserving.paddlepaddle.org/v1.EndpointSpec">
EndpointSpec
</a>
</em>
</td>
<td>
<p>DefaultTag defines default PaddleService endpoints</p>
</td>
</tr>
<tr>
<td>
<code>canary</code></br>
<em>
<a href="#elasticserving.paddlepaddle.org/v1.EndpointSpec">
EndpointSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CanaryTag defines an alternative PaddleService endpoints</p>
</td>
</tr>
<tr>
<td>
<code>canaryTrafficPercent</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>CanaryTrafficPercent defines the percentage of traffic going to canary PaddleService endpoints</p>
</td>
</tr>
<tr>
<td>
<code>service</code></br>
<em>
<a href="#elasticserving.paddlepaddle.org/v1.ServiceSpec">
ServiceSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>workingDir</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Container&rsquo;s working directory.
If not specified, the container runtime&rsquo;s default will be used, which
might be configured in the container image.
Cannot be updated.</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Pod volumes to mount into the container&rsquo;s filesystem.
Cannot be updated.</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of volumes that can be mounted by containers belonging to the pod.
More info: <a href="https://kubernetes.io/docs/concepts/storage/volumes">https://kubernetes.io/docs/concepts/storage/volumes</a></p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#elasticserving.paddlepaddle.org/v1.PaddleServiceStatus">
PaddleServiceStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="elasticserving.paddlepaddle.org/v1.PaddleServiceSpec">PaddleServiceSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#elasticserving.paddlepaddle.org/v1.PaddleService">PaddleService</a>)
</p>
<p>
<p>PaddleServiceSpec defines the desired state of PaddleService</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>runtimeVersion</code></br>
<em>
string
</em>
</td>
<td>
<p>Version of the service</p>
</td>
</tr>
<tr>
<td>
<code>resources</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<p>Defaults to requests and limits of 1CPU, 2Gb MEM.</p>
</td>
</tr>
<tr>
<td>
<code>default</code></br>
<em>
<a href="#elasticserving.paddlepaddle.org/v1.EndpointSpec">
EndpointSpec
</a>
</em>
</td>
<td>
<p>DefaultTag defines default PaddleService endpoints</p>
</td>
</tr>
<tr>
<td>
<code>canary</code></br>
<em>
<a href="#elasticserving.paddlepaddle.org/v1.EndpointSpec">
EndpointSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CanaryTag defines an alternative PaddleService endpoints</p>
</td>
</tr>
<tr>
<td>
<code>canaryTrafficPercent</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
<p>CanaryTrafficPercent defines the percentage of traffic going to canary PaddleService endpoints</p>
</td>
</tr>
<tr>
<td>
<code>service</code></br>
<em>
<a href="#elasticserving.paddlepaddle.org/v1.ServiceSpec">
ServiceSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>workingDir</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Container&rsquo;s working directory.
If not specified, the container runtime&rsquo;s default will be used, which
might be configured in the container image.
Cannot be updated.</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Pod volumes to mount into the container&rsquo;s filesystem.
Cannot be updated.</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of volumes that can be mounted by containers belonging to the pod.
More info: <a href="https://kubernetes.io/docs/concepts/storage/volumes">https://kubernetes.io/docs/concepts/storage/volumes</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="elasticserving.paddlepaddle.org/v1.PaddleServiceStatus">PaddleServiceStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#elasticserving.paddlepaddle.org/v1.PaddleService">PaddleService</a>)
</p>
<p>
<p>PaddleServiceStatus defines the observed state of PaddleService</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>Status</code></br>
<em>
knative.dev/pkg/apis/duck/v1.Status
</em>
</td>
<td>
<p>
(Members of <code>Status</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>url</code></br>
<em>
string
</em>
</td>
<td>
<p>URL of the PaddleService</p>
</td>
</tr>
<tr>
<td>
<code>default</code></br>
<em>
<a href="#elasticserving.paddlepaddle.org/v1.StatusConfigurationSpec">
StatusConfigurationSpec
</a>
</em>
</td>
<td>
<p>Statuses for the default endpoints of the PaddleService</p>
</td>
</tr>
<tr>
<td>
<code>canary</code></br>
<em>
<a href="#elasticserving.paddlepaddle.org/v1.StatusConfigurationSpec">
StatusConfigurationSpec
</a>
</em>
</td>
<td>
<p>Statuses for the canary endpoints of the PaddleService</p>
</td>
</tr>
<tr>
<td>
<code>address</code></br>
<em>
knative.dev/pkg/apis/duck/v1.Addressable
</em>
</td>
<td>
<p>Addressable URL for eventing</p>
</td>
</tr>
<tr>
<td>
<code>replicas</code></br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</tbody>
</table>
<h3 id="elasticserving.paddlepaddle.org/v1.ServiceSpec">ServiceSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#elasticserving.paddlepaddle.org/v1.PaddleServiceSpec">PaddleServiceSpec</a>)
</p>
<p>
<p>ServiceSpec defines the configuration for Knative Service.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>autoscaler</code></br>
<em>
<a href="#elasticserving.paddlepaddle.org/v1.Autoscaler">
Autoscaler
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>metric</code></br>
<em>
<a href="#elasticserving.paddlepaddle.org/v1.AutoscalerMetric">
AutoscalerMetric
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>window</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>panicWindow</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>panicThreshold</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>minScale</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>maxScale</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>target</code></br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>targetUtilization</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</tbody>
</table>
<h3 id="elasticserving.paddlepaddle.org/v1.StatusConfigurationSpec">StatusConfigurationSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#elasticserving.paddlepaddle.org/v1.PaddleServiceStatus">PaddleServiceStatus</a>)
</p>
<p>
<p>StatusConfigurationSpec describes the state of the configuration receiving traffic.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<p>Latest revision name that is in ready state</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>97fc986</code>.
</em></p>
