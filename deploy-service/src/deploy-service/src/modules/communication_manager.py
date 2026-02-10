
from typing import Optional
from urllib.parse import urlparse

import socket
import copy
import json
from tornado import httpclient

from src.clients.cms import CMSClient, CMSObject
from src.clients.helm import HelmClient
from src.clients.hydra import HydraClient
from src.clients.k8s import K8SClient
from src.common.error import MyHttpError, MyHttpCode
from src.common.log_util import logger
from src.modules.ssl_manager import SSLManager
from src.utils.net import get_host_for_url


def init_communication(data:dict):
    # 判断是否使用HTTP-Ingress
    cli = CommunicationManager()
    access_host = urlparse(cli.get_as_info()["access_addr"]).hostname
    SSLManager().init_global_https(access_host)  # 可能失败
    cli.patch_service_access(data["backend-service"])
    cli.patch_oauth_registry_info(data=data["oauth2"])


def update_communication(data:dict):
    cli = CommunicationManager()
    cli.patch_service_access(data["backend-service"])
    cli.patch_oauth_registry_info(data=data["oauth2"])

class CommunicationManager:
    def __init__(self, cms_client: Optional[CMSClient] = None) -> None:
        # 依赖的客户端
        self.__k8s_client: K8SClient = K8SClient()
        self.__cms_client: CMSClient = cms_client or CMSClient(self.__k8s_client.rest_client.self_namespace)  # anyshare
        self.__helm_client: HelmClient = HelmClient(self.__k8s_client.rest_client.self_namespace)  # anyshare

    def patch_service_access(self, data:dict=dict()):
        # TODO remove service access obj when uninstall
        # TODO process isntance thread safe != == !! :: := 1::2
        logger.info("[start] init or upgrade service access")
        cms_service_access = self.__cms_client.head_cms_data("service-access") or CMSObject.create("service-access")
        service_access_data = cms_service_access.real_data
        # 从 application_communication.conf 读取
        for k8s_service, info in data.items():
            k8s_service: str
            info: dict
            service_name = info.get("server-name", "")
            if service_name == "":
                continue
            if service_name not in service_access_data:
                service_access_data[service_name] = {}
            host_key: str = info.get("host-name", "host")
            port_key: str = info.get("port-name", "port")
            port: int = info.get("port", 0)
            service_access_data[service_name][host_key] = k8s_service
            service_access_data[service_name][port_key] = port
        # TODO 从其他cms读取
        cms_service_access.real_data = service_access_data
        cms_service_access.save(self.__cms_client)
        logger.info("[end] init or upgrade service access")

    def _update_redirect_uri(self, uri:str) -> str:
        access_addr = self.get_as_info()["access_addr"]
        access_addr_parse = urlparse(access_addr)
        port = access_addr_parse.port or 443 if access_addr_parse.scheme == "https" else 80
        uri = uri.replace('${access_scheme}', access_addr_parse.scheme)
        uri = uri.replace('${access_host}', get_host_for_url(str(access_addr_parse.hostname)))
        uri = uri.replace('${access_port}', str(port))
        uri = uri.replace('${access_path}', access_addr_parse.path.removesuffix("/"))
        return uri

    def refresh_oauth_client_uri(self, hydra_client: Optional[HydraClient] = None):
        logger.info("[start] init or upgrade registry info")
    
        hydra_client = hydra_client or HydraClient(self.__cms_client)
        cms_oauth_registry_info = self.__cms_client.head_cms_data("oauth-registry-info") or CMSObject.create(
            "oauth-registry-info")  # fmt: skip
        oauth_registry_info_data = cms_oauth_registry_info.real_data
        # refresh
        # """
        # oauth_client = {
        #     "deployOauthRawRequestBody": "raw request body in data["oauth2], type is a dict",
        #     "oauthClientID":  "str",
        #     "oauthClientSecret": "str",
        # }
        # """
        for client_name, oauth_client in oauth_registry_info_data.items():

            client_id:str = oauth_client.get("oauthClientID", "")
            if client_id=="":
                continue

            registry_params = oauth_client.get("deployOauthRawRequestBody", None)
            if registry_params == None:
                continue

            if "redirect_uris" not in registry_params and "post_logout_redirect_uris" not in registry_params:
                continue

            registry_params = copy.deepcopy(oauth_client["deployOauthRawRequestBody"])
            registry_params["client_name"] = client_name
            
            # 处理回调
            if "redirect_uris" in registry_params:
                for idx, uri in enumerate(registry_params["redirect_uris"]):
                    registry_params["redirect_uris"][idx] = self._update_redirect_uri(uri)
            if "post_logout_redirect_uris" in registry_params:
                for idx, uri in enumerate(registry_params["post_logout_redirect_uris"]):
                    registry_params["post_logout_redirect_uris"][idx] = self._update_redirect_uri(uri)

  
            logger.info(f"refresh hydra client {client_name}")
            hydra_client.update_client(client_id, registry_params)

            oauth_registry_info_data[client_name] = oauth_client

        cms_oauth_registry_info.real_data = oauth_registry_info_data
        cms_oauth_registry_info.save(self.__cms_client)
        logger.info("[ end ] init or upgrade registry info")
        

    def patch_oauth_registry_info(self, hydra_client: Optional[HydraClient] = None, data:dict = dict()):
        logger.info("[start] init or upgrade registry info")    
        hydra_client = hydra_client or HydraClient(self.__cms_client)
        cms_oauth_registry_info = self.__cms_client.head_cms_data("oauth-registry-info") or CMSObject.create(
            "oauth-registry-info")  # fmt: skip
        oauth_registry_info_data = cms_oauth_registry_info.real_data
        # 开始注册
        for client_info in data:
            client_info: dict
            client_name = client_info.get("client_name", "")
            if client_name == "":
                continue

            oauth_enabled = client_info.get("oauthOn", False)
            if not oauth_enabled:
                # 未开启，不注册
                continue
            # 注册oauth
            raw_registry_params = copy.deepcopy(client_info["oauth2"])
            registry_params = client_info["oauth2"]
            registry_params["client_name"] = client_name
            # 处理回调
            if "redirect_uris" in registry_params:
                for idx, uri in enumerate(registry_params["redirect_uris"]):
                    registry_params["redirect_uris"][idx] = self._update_redirect_uri(uri)
            if "post_logout_redirect_uris" in registry_params:
                for idx, uri in enumerate(registry_params["post_logout_redirect_uris"]):
                    registry_params["post_logout_redirect_uris"][idx] = self._update_redirect_uri(uri)

            oauth_client:dict = oauth_registry_info_data.get(client_name, dict())
            """
            oaut_client = {
                "deployOauthRawRequestBody": "raw request body in data["oauth2], type is a dict",
                "oauthClientID":  "str",
                "oauthClientSecret": "str",
            }
            """
            oauth_client["deployOauthRawRequestBody"] = raw_registry_params
            if "oauthClientID" in oauth_client :
                # 已经注册过，做更新
                if "redirect_uris" in registry_params or "post_logout_redirect_uris" in registry_params:
                    client_id = oauth_registry_info_data[client_name]["oauthClientID"]
                    logger.info(f"update hydra client {client_name}")
                    hydra_client.update_client(client_id, registry_params)

            else:
                logger.info(f"registry hydra client {client_name}")
                oauth_client["oauthClientID"], oauth_client["oauthClientSecret"] = hydra_client.registry_client(registry_params)

            oauth_registry_info_data[client_name] = oauth_client

        cms_oauth_registry_info.real_data = oauth_registry_info_data
        cms_oauth_registry_info.save(self.__cms_client)
        logger.info("[ end ] init or upgrade registry info")

    def upgrade_ingress_class443(self):
        logger.info("[start] upgrade nginx-ingress-controller class-443")
        ingress_class_name: str = f"class-443"       
        old_config = self.__helm_client.get_current_values(ingress_class_name)

        logger.info(f"upgrade ingress class {ingress_class_name}")
        release = self.__helm_client.get_release_info(ingress_class_name)
        if release != None:
            chart_name, chart_version = self.__helm_client.split_version(release["chart"])
            self.__helm_client.install_or_upgrade_any(
                name=ingress_class_name,
                config=old_config,
                chart_name=chart_name,
                chart_version=chart_version,
            )
        logger.info("[ end ] upgrade nginx-ingress-controller class-443")

    def __change_access_addr_data(self, access_addr: str, access_type: str):
        as_info = self.get_as_info()
        as_info["access_addr"] = access_addr
        as_info["access_type"] = access_type
        cms_as = self.__cms_client.get_cms_data("anyshare")
        cms_as.real_data = as_info
        cms_as.save(self.__cms_client)

    async def change_access_addr_not_refresh_ingress_class(self, access_addr: str, access_type: str, force: bool=False):
        """
        暂时不修改ingress-class，防止连接中断
        """

        def addr_is_open(host, port) -> bool:
            try:
                ss = socket.create_connection((host, int(port)), timeout=5.0)
                ss.shutdown(2)
                ss.close()
            except socket.error:
                return False
            else:
                return True

        def check_access_addr(
            old_access_addr: str,  # noqa
            old_access_type: str,  # noqa
            access_addr: str,  # noqa
            access_type: str,  # noqa
        ):
            parse_rel = urlparse(access_addr)
            parse_old = urlparse(old_access_addr)
            host, port = parse_rel.hostname, parse_rel.port
            port = port or 443 if parse_rel.scheme == "https" else 80
            old_port = parse_old.port or 443 if parse_old.scheme == "https" else 80
            if access_type == "internal" and port != old_port and addr_is_open(host, port):  # 内部地址如果端口改变，新端口需要不被使用
                raise MyHttpError(code=MyHttpCode.NCT_INVALID_ACCESS_ADDR, message="new internal addr should closed")
            # 只检查内部地址，外部地址不进行检查

        old_access_addr: str = self.get_as_info()["access_addr"]
        old_access_type: str = self.get_as_info()["access_type"]
        force or check_access_addr(old_access_addr, old_access_type, access_addr, access_type)
        self.__change_access_addr_data(access_addr, access_type)
        try:
            hydra_client = HydraClient(self.__cms_client)
            hydra_client.check_health()
            self.refresh_oauth_client_uri(hydra_client)
            self.__helm_client.upgrade_hydra_redirect_url(access_addr)
            self.__k8s_client.sdk_client.delete_pods(
                filter_str="webservice",
                namespace=self.__k8s_client.rest_client.self_namespace,
            )
        except Exception:
            logger.error("set access addr failed, rollback.")
            self.__change_access_addr_data(old_access_addr, old_access_type)
            self.__helm_client.upgrade_hydra_redirect_url(access_addr)
            raise
    
    async def async_get_access_addr(self):
        as_info = await self.async_get_as_info()
        access_addr = as_info["access_addr"]
        access_type = as_info["access_type"]
        u_parse = urlparse(access_addr)
        port = u_parse.port or 443 if u_parse.scheme == "https" else 80
        return {
            "host": u_parse.hostname,
            "port": str(port),
            "type": access_type,
            "scheme": u_parse.scheme,
            "path": u_parse.path or "/"
        }
    
    async def async_get_as_info(self):
        return (await self.__cms_client.async_get_cms_data("anyshare")).real_data
    
    
    def get_as_info(self):
        return self.__cms_client.get_cms_data("anyshare").real_data

