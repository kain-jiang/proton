from src.handlers.ssl_handler import (
    CertDownloadFeatureHandler,
    DownloadCertHandler,
    GetCertInfoHandler,
    SetGlobalCertHandler,
    UploadCertHandler,
)
from src.handlers.communication_handler import CommunicationHandler
from src.handlers.access_addr_handler import AccessAddrHandler,RedirectHTTPSHandler
from src.handlers.timestamp_handler import TimestampHandler

url_prefix = "/api/deploy-manager"

urls = [
    # 通信框架
    (rf"{url_prefix}/v1/communication/all", CommunicationHandler),  # POST: 初始化通信配置, PUT: 更新通信配置
    # ssl 证书
    (rf"{url_prefix}/cert/set-global-cert/app/(?P<ip>.*)", SetGlobalCertHandler),  # 生成证书
    (rf"{url_prefix}/cert/cert-info", GetCertInfoHandler),  # 获取证书信息
    (rf"{url_prefix}/cert/download-cert/app", DownloadCertHandler),  # 下载证书
    (rf"{url_prefix}/cert/upload-cert/app", UploadCertHandler),  # 上传自己的证书
    (rf"{url_prefix}/cert/feature/download", CertDownloadFeatureHandler),  # DELETE 禁用证书下载, GET 查询状态， POST/PUT 启用证书下载
    # 设置访问地址
    (rf"{url_prefix}/v1/access-addr/app", AccessAddrHandler),  # 设置访问地址
    (rf"{url_prefix}/v1/access-addr/redirect/(?P<uri>.*)", RedirectHTTPSHandler),  # 迁移端口到访问地址重定向

    (rf"/api/proton-openapi/v1alpha1/timestamp", TimestampHandler) # 迁移proton-openapi接口，获取当前AS服务器时间戳，所有使用接口的方案都应该考虑重构
]
