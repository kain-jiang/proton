from dataclasses import asdict, dataclass
from typing import Dict, Optional
from base64 import b64encode

import yaml
from kubernetes import client as kube_client

from src.clients.k8s import K8SClient

@dataclass
class CMSObject(object):
    name: str
    use: str
    data: Dict[str, dict]
    encrypt_field: list

    def to_dict(self) -> dict:
        return asdict(self)

    @property
    def real_data(self) -> dict:
        if self.use in self.data:
            return self.data[self.use]
        if "default" in self.data:
            return self.data["default"]
        for _, v in self.data.items():
            return v
        raise Exception("cms real data not found")

    @real_data.setter
    def real_data(self, data: dict):
        if self.use in self.data:
            self.data[self.use] = data
            return
        if "default" in self.data:
            self.data["default"] = data
            return
        for k, v in self.data.items():
            self.data[k] = data
            return
        raise Exception("cms real data not found")

    @classmethod
    def from_dict(cls, data: dict) -> "CMSObject":
        return cls(**data)

    def save(self, cli: "CMSClient"):
        cli.save_cms_data(self)

    def delete(self, cli: "CMSClient"):
        cli.delete_cms_data(self.name)

    @classmethod
    def create(cls, name: str, data: Optional[dict] = None) -> "CMSObject":
        return cls(name=name, use="default", data={"default": data or {}}, encrypt_field=[])
    
    def secret_data_b64encoded(self) -> dict[str, str]:
        data = {
            "name": b64encode(self.name.encode("utf-8")).decode("utf-8"),
            "use": b64encode(f"{self.use}.yaml".encode("utf-8")).decode("utf-8"),
            "encrypt_field": b64encode(yaml.safe_dump(self.encrypt_field).encode("utf-8")).decode("utf-8"),
        }
        for k, v in self.data.items():
            data[f"{k}.yaml"] = b64encode(yaml.safe_dump(v).encode("utf-8")).decode("utf-8")
        return data

class CMSClient(object):
    def __init__(self, namespace = "", host = "", port = ""):
        pass

    _instance = None

    @classmethod
    def instance(cls) -> "CMSClient":
        if cls._instance is None:
            cls._instance = cls()
        return cls._instance

    def _data_from_secret(self, data:" dict[str, bytes]") -> dict:
        for secret_key, secret_val in data.items():
            if secret_key.endswith(".yaml"):
                return yaml.safe_load(secret_val.decode("utf-8"))
        raise Exception("cms data not found")

    async def async_get_cms_data(self, name: str) -> CMSObject:
        return self.get_cms_data(name)

    def get_cms_data(self, name: str) -> CMSObject:
        info = self._head_from_secret(name)
        if info:
            return CMSObject.create(name, info)
        raise Exception(f"cms_{name} not found")
    
    def head_cms_data(self, name: str) -> Optional[CMSObject]:
        info = self._head_from_secret(name)
        if info:
            return CMSObject.create(name, info)
        return None

    def _head_from_secret(self, name: str) -> dict:
        try:
            k8s_cli = K8SClient.instance()
            secret = k8s_cli.sdk_client.get_secret(f"cms-release-config-{name}", k8s_cli.rest_client.self_namespace)
            return self._data_from_secret(secret)
        except kube_client.ApiException as api_exp:
            if api_exp.status == 404:
                return {}
            raise api_exp
    
    def save_cms_data(self, data: CMSObject, name: str = "") -> None:
        if not name:
            name = data.name
        k8s_cli = K8SClient.instance()
        k8s_cli.sdk_client.save_secret(
            name=f"cms-release-config-{name}",
            namespace=k8s_cli.rest_client.self_namespace,
            data=data.secret_data_b64encoded(),
        )

    def delete_cms_data(self, name: str) -> None:
        k8s_cli = K8SClient.instance()
        k8s_cli.sdk_client.delete_secret(f"cms-release-config-{name}", k8s_cli.rest_client.self_namespace)
