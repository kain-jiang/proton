from src.handlers.containerized_service_manager_handler import (
    MultiContainerizedManagerHandler,
    UPMultiContainerizedManagerHandler,
    ObsInstancesHandler,
)


from src.handlers.oss_handler import OSSHandler


url_prefix = "/api/deploy-manager"

urls = [

    # 安裝oss网关
    (rf"{url_prefix}/v1/containerized/multi-instance-service/(?P<service>.*)", MultiContainerizedManagerHandler),
    (rf"{url_prefix}/v1/containerized/upmulti-instance-service/(?P<service>.*)/(?P<nodes>.*)", UPMultiContainerizedManagerHandler),
    # CloudHub通知DPS添加对象存储
    ("/api/instances/v1/obs", ObsInstancesHandler),
    # 添加/更新OSS存储
    (rf"{url_prefix}/v1/oss", OSSHandler),
]
