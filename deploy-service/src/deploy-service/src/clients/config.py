import pathlib
import yaml
import os
import typing
class ConfigClient:

    def __init__(self, config: dict) -> None:
        self.config: dict = config

    _instance = {}

    @classmethod
    def load_config(cls, path: str = "/app/conf/config/deploy-service-config.yaml", renew: bool = False) -> "ConfigClient":
        if path in cls._instance and not renew:
            return cls._instance[path]
        cfg = yaml.safe_load(pathlib.Path(path).open())
        cls._instance[path] = cls(cfg)
        return cls._instance[path]
    
    def rds_type(self) -> str:
        return self.config["depServices"]["rds"]["type"]
    
    def rds_info(self) -> dict:
        rdsconf = {
            "system_id": ""
        }
        conf = self.config["depServices"]["rds"]
        if conf["type"].lower() == "dm8" and "," in conf["host"]:
            conf["host"] = "DM"
        os.environ["DB_TYPE"] = conf["type"].upper()
        rdsconf.update(conf)
        return rdsconf

    def ossgateway_version(self) -> str:
        return self.config["depServices"]["ossgateway"].get("deployTrait", {}).get("version", "unknown")


    def get_dep_service_info(self, service_name: str) -> dict:
        return self.config["depServices"].get(service_name, {})

    def use_protoncli(self) -> bool:
        return self.config.get("useProtonCli", True)

    def init_access_address(self) -> dict:
        return self.config.get("initAccessAddress", {})

    def proton_cli_secret_info(self) -> typing.Tuple[str, str, str]:
        secret_name: str = self.config.get("protonCliConfig", {})["name"]
        secret_namespace: str = self.config.get("protonCliConfig", {})["namespace"]
        secret_key: str = self.config.get("protonCliConfig", {})["key"]
        return secret_name, secret_namespace, secret_key
    


class TiduConfig:
    def __init__(self, config: dict) -> None:
        self.config: dict = config

    def access_addr(self) -> str:
        return self.config.get("accessAddr", "")

    def access_type(self) -> str:
        return self.config.get("accessType", "external")