#!/usr/bin/env python3
# -*- coding: utf-8 -*-

# ----------------------类型注解--------------------
ServiceName = str

OSSGATEWAY_Chart = "ossgateway"

# 服务访问配置模板文件（此路径已失效）
# ACCESS_TEM_FILE_PATH = "/app/conf/deploy_access/service_access.yaml"


ERROR_RESULT_TEM = {"code": "", "message": "", "cause": ""}

OAUTH_CLIENT_ID = "oauthClientID"
OAUTH_CLIENT_SECRET = "oauthClientSecret"


INGRESS_SECRET_NAME = "anyshare-ingress-tls"


CMS_ACCESS_INFO = {
    "host": "configuration-management-service.resource",
    "port": "8080",
}


# ----------------------文件参数--------------------
# APP_INGRESS_FILE = "/app/conf/deploy_access/app_ingress.yaml"

# 自身的 namespace
NAMESPACE_INFO_FILE_PATH = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
SERVICE_HOST_ENV_NAME = "KUBERNETES_SERVICE_HOST"  # service(kubernetes) host
SERVICE_PORT_ENV_NAME = "KUBERNETES_SERVICE_PORT"  # service(kubernetes) host
SERVICE_TOKEN_FILENAME = "/var/run/secrets/kubernetes.io/serviceaccount/token"  # kubernetes api bearer token
SERVICE_CERT_FILENAME = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"  # kubernetes api ca cert



# service_access 和 ingress端口信息存放的 secret 的（未带配置管理服务的前缀）名字
# 服务感知不到,可认为这就是 secret 的名字
SERVICE_ACCESS_CONFIG = "service-access"

OAUTH_REGISTRY_CONFIG = "oauth-registry-info"

# service_access 和 ingress端口信息存放在 secret.data 里 key
SERVICE_ACCESS_FILE_NAME = "service_access"
OAUTH_REGISTRY_FILE_NAME = "oauth_registry_info"

# ----------------url 参数---------------------

# secret
# url 中的 config_name 参数并不代表 Secret 的 name，配置管理服务会加上前缀，组合成 Secret 的名字
# get: 读取，post：新增，patch：更新
CONFIG_MGNT_URL = "http://{host}:{port}/api/cms/v1/configuration/service/{config_name}?namespace={namespace}"

# ----------------主模块相关---------------------
DEPLOYMENT_CONSOLE = "DeploymentConsole"
MANAGEMENT_CONSOLE = "ManagementConsole"
# AnyShare 主模块沿用旧的模块名 ManagementConsole
ANYSHARE_MAIN_MODULE: ServiceName = MANAGEMENT_CONSOLE

# ----------------可观测性高级服务相关------------
OBSERVABILITY: ServiceName = "Observability"

# ----------------知识中心相关--------------------
KNOWLEDGE_CENTER: ServiceName = "KnowledgeCenter"

# ----------------AnyDATA相关--------------------
ANYDATA: ServiceName = "AnyDATA"
# 不进行扩缩容的服务
SKIP_ADD_REMOVE_NODE_SERVICES = [
    "ingress-manager",
    "nginx-ingress-controller",
    "deploy-service",
    "deploy-web",
]


FORBIDDEN_UNINSTALL_SERVICES = [
    DEPLOYMENT_CONSOLE,
]

AISHU_HARBORS = ["acr.aishu.cn", "acr-arm.aishu.cn"]
