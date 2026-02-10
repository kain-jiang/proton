import os
from dataclasses import dataclass


@dataclass
class CRInfo(object):
    @dataclass
    class ChartmuseumInfo(object):
        auth_user: str
        auth_passwd: str
        hosts: list[str]
        port: int
        projects: list[str]
        push: bool
        repo_url: str

        @classmethod
        def from_dict(cls, data: dict) -> "CRInfo.ChartmuseumInfo":
            return cls(**data)
        
        def get_repo_url(self) -> str:
            if "chartmuseum.aishu.cn" in self.repo_url:
                env_host_ip = os.getenv("HOST_IP")
                if env_host_ip is not None:
                    return self.repo_url.replace("chartmuseum.aishu.cn", env_host_ip)
            return self.repo_url

    @dataclass
    class RegistryInfo(object):
        hosts: list[str]
        port: int
        image_pull_secret: str
        protocol: str
        push: bool
        server: str

        @classmethod
        def from_dict(cls, data: dict) -> "CRInfo.RegistryInfo":
            return cls(**data)
        
    @dataclass
    class OCI(object):
        registry: str = ""
        plain_http: bool = False
        username: str = ""
        password: str = ""

        def get_registry(self) -> str:
            if "registry.aishu.cn" in self.registry:
                env_host_ip = os.getenv("HOST_IP")
                if env_host_ip is not None:
                    return self.registry.replace("registry.aishu.cn", env_host_ip)
            return self.registry

    chartmuseum: ChartmuseumInfo
    registry: RegistryInfo
    oci: OCI
    chart_repository: str
    image_repository: str

    @classmethod
    def from_dict(cls, data: dict) -> "CRInfo":
        return cls(
            chartmuseum=CRInfo.ChartmuseumInfo.from_dict(data["chartmuseum"]),
            registry=CRInfo.RegistryInfo.from_dict(data["registry"]),
            oci=CRInfo.OCI(**(data.get("oci", {}))),
            chart_repository=data["chart_repository"],
            image_repository=data["image_repository"],
        )
    
    def registry_address(self) -> str:
        if self.image_repository == "oci":
            return self.oci.registry
        return self.registry.server