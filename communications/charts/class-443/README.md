以proton的nginx-ingress-controller魔改版为基础组件，生成新的应用内组件。该文件将会生成一个新的chart，并重命名chart名称，合并基础组件chart的values.yaml值与本文件内设置的值作为新chart的默认值。

新chart的使用，应该在对应的模块化安装包中引用。

```yaml
service:
    network: "hostnetwork"
    ingressClass: class-443
    healthPort: 100222
    httpPort: 80
    httpsPort: 443
```
* instance-name: nginx-ingress-controller服务组件实例名称，一个实例名称也会时一个chart名称
* service：nginx-controller网络相关配置
  * network：目前支持"hostport"、"hostnetwork"和空
    * hostPort： 服务以容器网络运行，但httpPort与httpsPort将会占用相同值的主机端口，并利用该端口对外提供服务
    * hostnetwork：服务将会以主机网络运行，healthPort、httpPort和httpsPort将监听主机端口
    * 空值: 容器网络，不在主机上暴露端口，未来应该演进为各种环境下的loadbalance
  * ingressClass：nginx-ingress-controller在集群内唯一标识，建议与instance-name保持一致
  * healthPort：容器健康检查端口，主要用于为监控平台提供服务运行指标
  * httpPort: nginx提供http服务时的服务端口，当不设置或值小于1时，不启动http
  * httpsPort：nginx提供httpss服务时的服务端口，当不设置或值小于1时，不启动https