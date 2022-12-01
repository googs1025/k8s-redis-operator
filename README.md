## 自定义资源对象与控制器 RedisCluster

### 项目思路与功能
项目背景：自定义部署Redis资源，结合Deployment与Service(可以自定义)。
```bigquery
apiVersion: redis.jiang.operator/v1
kind: RedisCluster
metadata:
  name: rediscluster-sample
spec:
  size: 2 # pod副本
  image: redis:5-alpine # 镜像
  service: true   # 自定义是否要配置Service true false
  service_type: NodePort # 目前支持NodePort或ClusterIP
  ports: #端口
    - port: 80
      targetPort: 80 # 容器端口
      nodePort: 30002 #service端口 # 使用ClusterIP，需要把这个字段删去
```
思路：

自定义CRD启动时，会启动自己的Controller，并同时关联和拉起Deployment、Service(自定义)。

![](https://github.com/googs1025/Kubernetes-operator-AppDeployer/blob/main/images/%E6%B5%81%E7%A8%8B%E5%9B%BE.jpg?raw=true)

### 附注
1. 本项目依赖 kubebuilder kustomize k8s集群(kubeadm安装)，请先安装这些依赖
```bash
[root@VM-0-16-centos samples]# kubebuilder version
Version: main.version{KubeBuilderVersion:"3.4.0", KubernetesVendor:"1.23.5", GitCommit:"75241ab9ff9457de77e902645792cee41ba29fed", BuildDate:"2022-04-28T17:09:31Z", GoOs:"linux", GoArch:"amd64"}
[root@VM-0-16-centos samples]# kustomize version
{Version:kustomize/v3.5.2 GitCommit:79a891f4881cfc780e77789a1d240d8f4bfa2598 BuildDate:2019-12-17T03:48:17Z GoOs:linux GoArch:amd64}
[root@VM-0-16-centos samples]# kubectl version
Client Version: version.Info{Major:"1", Minor:"22", GitVersion:"v1.22.3", GitCommit:"c92036820499fedefec0f847e2054d824aea6cd1", GitTreeState:"clean", BuildDate:"2021-10-27T18:41:28Z", GoVersion:"go1.16.9", Compiler:"gc", Platform:"linux/amd64"}
Server Version: version.Info{Major:"1", Minor:"22", GitVersion:"v1.22.3", GitCommit:"c92036820499fedefec0f847e2054d824aea6cd1", GitTreeState:"clean", BuildDate:"2021-10-27T18:35:25Z", GoVersion:"go1.16.9", Compiler:"gc", Platform:"linux/amd64"}
```
2. 本项目在Dockerfile Makefile中有稍作修改，如运行项目出现报错，可自行适配

### 项目测试
1. make 执行一下
```bash
[root@VM-0-16-centos k8s-operator-redis]# make
/root/k8s-operator-redis/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
go vet ./...
go build -o bin/manager main.go
```
2. make install 安装CRD
```bash
[root@VM-0-16-centos k8s-operator-redis]# make install
/root/k8s-operator-redis/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
kustomize build config/crd | kubectl apply -f -
customresourcedefinition.apiextensions.k8s.io/redisclusters.redis.jiang.operator configured
```
3. make run 启动控制器
```bash
[root@VM-0-16-centos k8s-operator-redis]# make run
/root/k8s-operator-redis/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
/root/k8s-operator-redis/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
go vet ./...
go run ./main.go
1.6699024307412884e+09  INFO    controller-runtime.metrics      Metrics server is starting to listen    {"addr": ":8080"}
1.6699024307415726e+09  INFO    setup   starting manager
1.6699024307418582e+09  INFO    Starting server {"path": "/metrics", "kind": "metrics", "addr": "[::]:8080"}
1.669902430741884e+09   INFO    Starting server {"kind": "health probe", "addr": "[::]:8081"}
1.6699024307419376e+09  INFO    controller.rediscluster Starting EventSource    {"reconciler group": "redis.jiang.operator", "reconciler kind": "RedisCluster", "source": "kind source: *v1.RedisCluster"}
1.6699024307419868e+09  INFO    controller.rediscluster Starting EventSource    {"reconciler group": "redis.jiang.operator", "reconciler kind": "RedisCluster", "source": "kind source: *v1.Deployment"}
1.6699024307419994e+09  INFO    controller.rediscluster Starting EventSource    {"reconciler group": "redis.jiang.operator", "reconciler kind": "RedisCluster", "source": "kind source: *v1.Service"}
1.6699024307420106e+09  INFO    controller.rediscluster Starting EventSource    {"reconciler group": "redis.jiang.operator", "reconciler kind": "RedisCluster", "source": "kind source: *v1.Deployment"}
1.6699024307420244e+09  INFO    controller.rediscluster Starting EventSource    {"reconciler group": "redis.jiang.operator", "reconciler kind": "RedisCluster", "source": "kind source: *v1.Service"}
1.6699024307420516e+09  INFO    controller.rediscluster Starting Controller     {"reconciler group": "redis.jiang.operator", "reconciler kind": "RedisCluster"}
1.6699024308431652e+09  INFO    controller.rediscluster Starting workers        {"reconciler group": "redis.jiang.operator", "reconciler kind": "RedisCluster", "worker count": 1}
1.6699024308432896e+09  INFO    controller.rediscluster Start Reconcile Loop    {"reconciler group": "redis.jiang.operator", "reconciler kind": "RedisCluster", "name": "rediscluster-sample", "namespace": "default"}
1.669902430846875e+09   INFO    controller.rediscluster CreateOrUpdate  {"reconciler group": "redis.jiang.operator", "reconciler kind": "RedisCluster", "name": "rediscluster-sample", "namespace": "default", "RedisDeployment": "updated"}
1.6699024308511002e+09  INFO    controller.rediscluster CreateOrUpdate  {"reconciler group": "redis.jiang.operator", "reconciler kind": "RedisCluster", "name": "rediscluster-sample", "namespace": "default", "RedisService": "updated"}
```
4. 创建对象
```bash
[root@VM-0-16-centos k8s-operator-appDeployer]# cd config/samples/
[root@VM-0-16-centos samples]# kubectl apply -f .
rediscluster.redis.jiang.operator/rediscluster-sample configured
```
### 项目部署
1. 部署
```bigquery
# 打包controller docker 镜像
make docker-build
# 部署controller 项目
make deploy
```
2. 查看
```bash
[root@VM-0-16-centos k8s-operator-redis]# kubectl get ns | grep operator-redis-system
operator-redis-system     Active   12m
[root@VM-0-16-centos k8s-operator-redis]# kubectl get pods -noperator-redis-system
NAME                                                 READY   STATUS    RESTARTS   AGE
operator-redis-controller-manager-6f599f74f4-n4d2t   2/2     Running   0          12m
```

