from typing import Tuple, Union

import requests
from requests.exceptions import HTTPError
from src.common.log_util import logger

from src.clients.cms import CMSClient, CMSObject
from src.utils.net import get_host_for_url



class HydraClient(object):
    def __init__(self, cms_client: Union[None, CMSClient] = None):
        if not cms_client:
            cms_client = CMSClient()
        cms_service_access: CMSObject = cms_client.head_cms_data("service-access") or CMSObject.create("service-access")  # fmt: skip
        self.__hydra_admin_host = cms_service_access.real_data.get("hydra", {}).get("administrativeHost", "hydra-admin")
        self.__hydra_admin_port = cms_service_access.real_data.get("hydra", {}).get("administrativePort", "4445")

    def url_str(self, path):
        return f"http://{get_host_for_url(self.__hydra_admin_host)}:{self.__hydra_admin_port}{path}"

    def registry_client(self, registry_params: dict) -> Tuple[str, str]:
        resp: requests.Response = requests.post(url=self.url_str("/admin/clients"), json=registry_params, timeout=30.0)
        resp.raise_for_status()
        resp_data = resp.json()
        return resp_data["client_id"], resp_data["client_secret"]

    def update_client(self, client_id: str, registry_params: dict):
        resp: requests.Response = requests.put(
            url=self.url_str(f"/admin/clients/{client_id}"), json=registry_params, timeout=30.0
        )
        try:
            resp.raise_for_status()
        except HTTPError as e:
            logger.error(f"resp: {resp.text}, request: {registry_params}")
            raise

    def check_health(self):
        resp = requests.get(self.url_str("/health/alive"))
        resp.raise_for_status()
