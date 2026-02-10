from src.clients.k8s import K8SClient
from src.clients.config import ConfigClient
from src.clients.cms_data.cr import CRInfo

import yaml
from urllib.parse import urlparse


class ProtonCliClient(object):
    _instance = None
    @classmethod
    def instance(cls) -> "ProtonCliClient":
        if cls._instance is None:
            cls._instance = cls()
        return cls._instance

    def __init__(self):
        k8s_client = K8SClient.instance()
        cfg = ConfigClient.load_config()
        if not cfg.use_protoncli():
            raise Exception("protoncli not enabled, cannot get cr info from protoncli")
        secret_name, secret_namespace, secret_key = cfg.proton_cli_secret_info()
        content = k8s_client.sdk_client.get_secret(secret_name, secret_namespace)[secret_key]
        self.config = yaml.safe_load(content.decode("utf-8"))

    def cr_info(self) -> CRInfo:
        if self.config["cr"].get("local", None) is not None:
            ipFamilies = self.config["cs"].get("ipFamilies", [])
            def crNodes() -> list[str]:
                result = []
                for host in self.config["cr"]["local"]["hosts"]:
                    for node in self.config["nodes"]:
                        if host == node["name"]:
                            ip = node.get("ip4", "")
                            if "ip6" in node and "IPv6" in ipFamilies:
                                ip = node["ip6"]
                            if "ip4" in node and "IPv4" in ipFamilies:
                                ip = node["ip4"]
                            result.append(ip)
                            break
                return result
            return CRInfo(
                chart_repository="chartmuseum",
                image_repository="registry",
                chartmuseum=CRInfo.ChartmuseumInfo(
                    push=True,
                    port=self.config["cr"]["local"]["ports"]["chartmuseum"],
                    auth_user="",
                    auth_passwd="",
                    projects=["helm_repos"],
                    repo_url=f"http://chartmuseum.aishu.cn:{self.config['cr']['local']['ha_ports']['chartmuseum']}",
                    hosts=crNodes(),
                ),
                registry=CRInfo.RegistryInfo(
                    push=True,
                    port=self.config["cr"]["local"]["ports"]["registry"],
                    protocol="http",
                    server=f"registry.aishu.cn:{self.config['cr']['local']['ha_ports']['registry']}",
                    image_pull_secret="",
                    hosts=crNodes(),
                ),
                oci=None,
            )
        elif self.config["cr"].get("external", None) is not None:
            chartmuseum_info = None
            registry_info = None
            oci_info = None
            if "registry" in self.config["cr"]["external"]:
                registryAddress: str = self.config["cr"]["external"]["registry"]["host"]
                if not registryAddress.startswith("http://") and not registryAddress.startswith("https://"):
                    registryAddress = f"https://{registryAddress}"
                imagePullSecretName = ""
                if self.config["cr"]["external"]["registry"].get("username", "") != "":
                    imagePullSecretName = "external-registry"
                    k8s_client = K8SClient.instance()
                    k8s_client.sdk_client.create_secret_docker_registry(
                        ns=k8s_client.rest_client.self_namespace,
                        name=imagePullSecretName,
                        host=self.config["cr"]["external"]["registry"]["host"],
                        username=self.config["cr"]["external"]["registry"]["username"],
                        password=self.config["cr"]["external"]["registry"]["password"],
                    )
                registry = urlparse(registryAddress)
                registry_info = CRInfo.RegistryInfo(
                    push=True,
                    port=registry.port,
                    protocol=registry.scheme,
                    server=self.config["cr"]["external"]["registry"]["host"],
                    image_pull_secret=imagePullSecretName,
                    hosts=[],
                )
            if "chartmuseum" in self.config["cr"]["external"]:
                chartmuseum = urlparse(self.config["cr"]["external"]["chartmuseum"]["host"])
                chartmuseum_info = CRInfo.ChartmuseumInfo(
                    push=True,
                    port=chartmuseum.port,
                    auth_user=self.config["cr"]["external"]["chartmuseum"].get("username", ""),
                    auth_passwd=self.config["cr"]["external"]["chartmuseum"].get("password", ""),
                    projects=["helm_repos"],
                    repo_url=self.config["cr"]["external"]["chartmuseum"]["host"],
                    hosts=[],
                )
            if "oci" in self.config["cr"]["external"]:
                oci_info = CRInfo.OCI(**(self.config["cr"]["external"]["oci"]))
            return CRInfo(
                chart_repository=self.config["cr"]["external"].get("chart_repository", "chartmuseum"),
                image_repository=self.config["cr"]["external"].get("image_repository", "registry"),
                registry=registry_info,
                chartmuseum=chartmuseum_info,
                oci=oci_info,
            )
        raise Exception("cannot found cr info in proton configuration")
    
    def nodes(self) -> list[dict]:
        return self.config.get("nodes", [])
    def deploy_mode(self) -> str:
        return self.config.get("deploy", {}).get("mode", "")

    def devicespce_ini(self) -> str:
        return f"""
[DeviceSpec]
HardwareType = {self.config.get("deploy", {}).get("devicespec", "")}
"""