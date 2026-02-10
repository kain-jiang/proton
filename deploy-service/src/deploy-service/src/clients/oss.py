import requests
from typing import Optional

from src.clients.cms import CMSClient, CMSObject
from src.utils.net import get_host_for_url
from src.common.utils import retry_by_exception
from src.common.log_util import log_response_with_request
from src.common.error import OtherHttpError


def custom_raise_for_status(resp: requests.Response):
    if not resp.ok:
        raise OtherHttpError(resp)
    resp.raise_for_status()

class OssGatewayManagerClient(object):
    def __init__(self, cms_client: Optional[CMSClient] = None):
        if not cms_client:
            cms_client = CMSClient()
        cms_service_access: CMSObject = cms_client.head_cms_data("service-access") or CMSObject.create(
            "service-access")  # fmt: skip
        ossgatewaymanager_private_host = cms_service_access.real_data.get("ossgatewaymanager", {}).get("privateHost",
                                                                                                       "ossgatewaymanager-private")
        ossgatewaymanager_private_port = cms_service_access.real_data.get("ossgatewaymanager", {}).get("privatePort",
                                                                                                       "9002")
        self.private_host = f"{get_host_for_url(ossgatewaymanager_private_host)}:{ossgatewaymanager_private_port}"

    _instance = None

    @classmethod
    def instance(cls) -> "OssGatewayManagerClient":
        if cls._instance is None:
            cls._instance = cls()
        return cls._instance

    def AddOSSInfo(self, storageInfoBody: dict):
        resp: requests.Response = requests.post(
            url=f"http://{self.private_host}/api/ossgateway/v1/objectstorageinfo",
            json=storageInfoBody
        )
        log_response_with_request(resp)
        resp.raise_for_status()
        return resp.json()

    def SetStoragePrefix(self, storage_prefix):
        resp: requests.Response = requests.post(
            url=f"http://{self.private_host}/api/ossgateway/v1/storageprefix",
            json={
                "storagePrefix": storage_prefix
            }
        )
        log_response_with_request(resp)
        resp.raise_for_status()
        return resp.json()

    def GetCacheOSSInfo(self):
        resp: requests.Response = requests.get(
            url=f"http://{self.private_host}/api/ossgateway/v1/objectstorageinfo?isCache=true",
        )
        resp.raise_for_status()
        return resp.json()

    def GetDownloadInfo(self, ossid: str, key: str, file_name: str, user_oss_id: str) -> dict:
        params = {"type": "query_string", "save_name": file_name}
        if user_oss_id:
            params["user"] = user_oss_id
        resp: requests.Response = requests.get(
            url=f"http://{self.private_host}/api/ossgateway/v1/download/{ossid}/{key}",
            params = params
        )
        resp.raise_for_status()
        return resp.json()

    def GetOSSInfo(self):
        resp: requests.Response = requests.get(
            url=f"http://{self.private_host}/api/ossgateway/v1/objectstorageinfo?isCache=false",
        )
        resp.raise_for_status()
        return resp.json()

    def GetSiteDefaultOSS(self):
        resp: requests.Response = requests.get(
            url=f"http://{self.private_host}/api/ossgateway/v1/default-storage",
        )
        resp.raise_for_status()
        return resp.json()

    def GetUploadInfo(self, ossid: str, key: str, request_method: str = "PUT", query_string: bool = True) -> dict:
        resp: requests.Response = requests.get(
            url=f"http://{self.private_host}/api/ossgateway/v1/upload/{ossid}/{key}",
            params={"type": "query_string", "request_method": request_method}
            if query_string
            else {"request_method": request_method},
        )
        resp.raise_for_status()
        return resp.json()

    def GetDeleteInfo(self, ossid: str, key: str) -> dict:
        resp: requests.Response = requests.get(url=f"http://{self.private_host}/api/ossgateway/v1/delete/{ossid}/{key}")
        resp.raise_for_status()
        return resp.json()

    def ModifyExistingOSSInfo(self, storageInfoBody: dict):
        resp: requests.Response = requests.put(
            url=f"http://{self.private_host}/api/ossgateway/v1/objectstorageinfo",
            json=storageInfoBody
        )
        log_response_with_request(resp)
        custom_raise_for_status(resp)

    def SetSiteDefaultOSS(self, storageId: str):
        resp: requests.Response = requests.put(
            url=f"http://{self.private_host}/api/ossgateway/v1/default-storage/" + storageId,
        )
        resp.raise_for_status()

    def UnbindBucket(self, *, bucket_name: str, vendor_type: str):
        resp: requests.Response = requests.delete(
            url=f"http://{self.private_host}/api/ossgateway/v1/bucket",
            json={
                "bucket_name": bucket_name,
                "vendor_type": vendor_type,
            }
        )
        if resp.status_code != 404:
            resp.raise_for_status()

    @retry_by_exception(attempt=15)
    def BindBucket( # noqa
        self, *,
        bucket_name: str,
        access_key: str,  # 需要明文密码
        access_key_id: str,
        url: str,
        internal_url: str,
        vendor_type: str,
        url_list: list,
        region: str = "",
        size: int = None
    ):
        request_json = {
            "bucket_name": bucket_name,
            "access_key": access_key,
            "access_key_id": access_key_id,
            "url": url,
            "internal_url": internal_url,
            "url_list": url_list,
            "vendor_type": vendor_type,
            "region": region,
        }
        if size is not None:
            request_json["size"] = size

        resp: requests.Response = requests.post(
            url=f"http://{self.private_host}/api/ossgateway/v1/bucket",
            json=request_json
        )
        resp.raise_for_status()

    def CreateInternalAddresses(
        self, *,
        bucket_name: str,
        vendor_type: str,
        internal_list: list,
    ):
        resp: requests.Response = requests.post(
            url=f"http://{self.private_host}/api/ossgateway/v1/internal-addresses",
            json={
                "bucket_name": bucket_name,
                "vendor_type": vendor_type,
                "internal_list": internal_list
            }
        )
        resp.raise_for_status()


class OSSTools:

    @staticmethod
    def to_url(*, host: str, http_port: int = None, https_port: int = None):
        if not host:
            # 没传入host
            return ""
        if https_port:
            return f"https://{get_host_for_url(host)}:{https_port}"
        if http_port:
            return f"http://{get_host_for_url(host)}:{http_port}"  # noqa
        return f"https://{get_host_for_url(host)}"
