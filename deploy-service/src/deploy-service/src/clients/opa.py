import os

import requests
from src.common.log_util import logger
from src.utils.net import get_host_for_url


class OPAClient(object):
    def __init__(self) -> None:
        self.host = os.environ.get('POLICY_ENGINE_HOST')
        self.port = os.environ.get('POLICY_ENGINE_PORT')

    _instance = None

    @classmethod
    def instance(cls) -> "OPAClient":
        if cls._instance is None:
            cls._instance = cls()
        return cls._instance
    
    def download_strategy(self, user_arr: list, os_type: str) -> dict:
        data = {
            "input": {
                "user": user_arr,
                "client": os_type
            }
        }
        resp: requests.Response = requests.post(
            url=f"http://{get_host_for_url(host=self.host)}:{self.port}/api/proton-policy-engine/v2/data/client_update_policy/update_detection",
            json=data
        )
        resp.raise_for_status()
        return resp.json()



