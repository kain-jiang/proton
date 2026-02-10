from src.handlers.client_package_handler import (
    CheckStorageHandler,
    GetDownloadUrlHandler,
    OsConfigHandler,
    PackageHandler,
    PackageVersionHandler,
    SetDownloadInfoHandler,
    SetVersionDescriptionHandler,
    UpadteTypeHandler, ClientPackageUploadHandler,
)

url_prefix = "/api/deploy-manager"

urls = [
    # 客户端更新机制
    (rf"{url_prefix}/client/package-info", PackageHandler),  # 获取客户端升级包信息
    (rf"{url_prefix}/client/package", PackageHandler),  # 获取客户端升级包信息(客户端)
    (rf"{url_prefix}/client/delete-package/(?P<ostype>.*)", PackageHandler),  # 删除客户端升级包信息
    (rf"{url_prefix}/client/get-config", OsConfigHandler),  # 获取所有客户端是否开放下载配置/或指定系统
    (rf"{url_prefix}/client/downloadable", OsConfigHandler),  # 获取所有客户端是否开放下载配置/或指定系统==（客户端）
    (rf"{url_prefix}/client/check-storage", CheckStorageHandler),  # 检查是否有swift或oss存储
    (rf"{url_prefix}/client/set-download-info", SetDownloadInfoHandler),  # 设置包下载信息
    (rf"{url_prefix}/client/get-download-url", GetDownloadUrlHandler),  # 获取包下载地址
    (rf"{url_prefix}/client/download-url", GetDownloadUrlHandler),  # 获取包下载地址===（客户端）
    (rf"{url_prefix}/client/version/(?P<ostype>.*)/(?P<version>.*)", PackageVersionHandler),  # 检查是否有升级包更新（客户端）
    (rf"{url_prefix}/client/set-package-config", SetVersionDescriptionHandler),  # 设置安装包描述信息，是否开放下载
    (rf"{url_prefix}/client/package-config", SetVersionDescriptionHandler),  # 设置安装包描述信息，是否开放下载
    (rf"{url_prefix}/client/set-update-type", UpadteTypeHandler),  # 设置当前使用的升级包类型
    (rf"{url_prefix}/client/update-type", UpadteTypeHandler),  # 设置当前使用的升级包类型
    (rf"{url_prefix}/client/get-update-type", UpadteTypeHandler),  # 获取当前使用的升级包类型
    (rf"{url_prefix}/v1/client/package/upload", ClientPackageUploadHandler),  # 上传客户端包
]