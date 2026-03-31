
# 应用包定义文件：./kweaver-core.yaml

需要实现以下命令行内容：

## 1. 从清单文件导出离线包
```bash
proton-cli offline_package app -f/--manifest ./kweaver-core.yaml export --output/-o offline-app-package.tar \
    --platform linux/amd64
```

- --platform: 支持 linux/amd64 和 linux/arm64 两个架构，同时接受 x86_64 aarch64 amd64 arm64
- --manifest/-f: 清单文件地址，接受 url
- --output/-o: 文件导出路径

需要保存其中两种资源类型：镜像和Chart

### Chart：取自清单文件 `source` `releases` 两个字段
### 镜像：需要针对每个chart进行提取，chart中拥有以下values.yaml的定义
- 1. 单镜像结构
    image.registry
    image.repository
    image.tag
- 2. 多镜像结构
    image.registry
    image.<name>.repository
    image.<name>.tag
需要通过这个方式取得所有镜像，镜像使用 oci-layout 进行保存

offline-app-package.tar
- manifest.yaml # 原样保存清单文件，补充 platform 信息
- images # 镜像oci-layout文件，所有镜像全部保存，镜像名存储为 <repository>:<tag>，不保留 registry
- charts # charts列表，所有chart平铺保存

## 2. 从清单文件导入离线包本地仓库
```bash
proton-cli offline_package app import --input -i offline-app-package.tar \
    --registry registry.hello.com \
    --registry-username username \
    --registry-password password \
    --registry-plain-http true/false \
    --chartmuseum-url https://chartmuseum.hello.com \
    --chartmuseum-username username \
    --chartmuseum-password password \
```


- --input/-i: 离线包本地路径，不接受url
- --registry-xxxx: 导入得目标镜像仓库地址
- --chartmuseum-xxxx：导入的镜像仓库地址