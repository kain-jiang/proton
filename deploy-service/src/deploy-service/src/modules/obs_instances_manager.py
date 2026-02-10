#!/usr/bin/env python3
# -*- coding:utf-8 -*-

import base64
import hmac
import requests
import json
import hashlib
import urllib
import decimal
import datetime

from src.common.log_util import logger
from src.clients.oss import OssGatewayManagerClient
from src.common import global_param, utils
from src.common.utils import retry_by_exception
from src.lib.db.obsinstance_oss import OBSInstanceOss
from src.modules.ossgateway_service_manager import OSSGatewayManager
from src.utils.net import get_host_for_url

OATUH_URI = "/v1/token"
OATUH_URL = "https://{oauth_host}/v1/token?access_id={access_id}&signature={signature}"

OSS_SERVICE_NAME = "OSSGatewayService"

OSS_DATA = {
    "ossId": "",
    "ossName": "",
    "enabled": True,
    "siteId": "",
    "ossgwMergeSize": 0,
    "bucketInfo": {
        "accessId": "",
        "accessKey": "",
        "name": "",
        "cdnName": "",
        "internalServerName": "",
        "provider": "",
        "serverName": "",
        "providerDetail": "",
        "bucketStyle": 0,
        "isCacheBucket": False,
        "httpPort": None,
        "httpsPort": 443,
        "region": ""
    }
}


class EisJsonEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, datetime.datetime):
            return obj.strftime('%Y-%m-%d %H:%M:%S')
        elif isinstance(obj, datetime.date):
            return obj.strftime('%Y-%m-%d')
        elif isinstance(obj, datetime.time):
            return obj.strftime('%H:%M:%S')
        elif isinstance(obj, datetime.timedelta):
            return str(obj)
        elif isinstance(obj, decimal.Decimal):
            return float(obj)
        elif isinstance(obj, set):
            return list(obj)
        elif isinstance(obj, bytes):
            return obj.decode('utf-8')
        else:
            return json.JSONEncoder.default(self, obj)


class ObsInstanceManager:
    access_conf = None

    def __init__(self):
        AuthServer_Info = utils.get_authserver_info()
        self.oauth_host = AuthServer_Info["AuthServer"]["host"]
        self.access_key = AuthServer_Info["AuthServer"]["ak"]
        self.secret_key = AuthServer_Info["AuthServer"]["sk"]
        self.access_conf = utils.read_conf_in_config(
            global_param.SERVICE_ACCESS_CONFIG, global_param.SERVICE_ACCESS_FILE_NAME
        )

    @classmethod
    def get_ossgateway_client(cls):
        return OssGatewayManagerClient.instance()

    @classmethod
    def get_access_info(cls):
        if cls.access_conf is None:
            cls.access_conf = utils.read_conf_in_config(
                global_param.SERVICE_ACCESS_CONFIG, global_param.SERVICE_ACCESS_FILE_NAME
            )
        return cls.access_conf

    def add_obs_config(self, data, token_id):
        # 验证token
        self.verify_token(token_id)

        # 获取通过此接口安装的所有bucket
        all_bucket = OBSInstanceOss.get_oss_bucket()
        ossgateway_instance = OSSGatewayManager()

        # 如果没有安装过, 设置前缀
        if len(all_bucket) == 0:
            data_first_dir = OBSInstanceOss.aes_decrypt(data["data_first_dir"], self.secret_key)
            ossgateway_instance.set_storage_prefix(data_first_dir)

        # 查看是否安装过相同的bucket
        for bucket_info in all_bucket:
            if data["name"] == bucket_info["name"]:
                logger.info("%s is already installed, skip" % data["name"])
                return 1, {"obs_id": bucket_info["obs_id"]}

        # 通过安装OSS网关接口配置bucket
        oss_data = self.get_oss_data(data)
        code = ossgateway_instance.add_oss_config(OSS_SERVICE_NAME, oss_data)
        if code:
            return 0, code

        # 配置第一个存储为默认存储
        if len(all_bucket) == 0:
            ossinfos = ossgateway_instance._get_oss_info()
            for ossinfo in ossinfos:
                bucket_data = ossinfo
                if data["name"] == bucket_data["bucketInfo"]["name"]:
                    self.get_ossgateway_client().SetSiteDefaultOSS(bucket_data["ossId"])

        # 写入数据库,记录bucket和obs_id对应关系
        oss_info = dict()
        oss_info["bucket"] = data.get("bucket", "")
        oss_info["obs_id"] = data.get("obs_id", "")
        oss_info["instance_id"] = data.get("instance_id", "")
        OBSInstanceOss.insert_oss_info(oss_info)

        return 1, {"obs_id": oss_info["obs_id"]}

    def update_obs_config(self, data, token_id):
        # 验证token
        self.verify_token(token_id)

        # 由于AS未实现此接口参数的功能，故此接口不做任何事
        return 

    def get_oss_data(self, data):
        # 检查参数
        parameters = ["obs_id", "obs_type", "spare_key", "bucket", "region_endpoint"]
        parameter_err = []
        for parameter in parameters:
            if parameter not in data:
                parameter_err.append("%s not in body" % parameter)
            elif not data[parameter]:
                parameter_err.append("%s is null" % parameter)
        if parameter_err:
            raise Exception(parameter_err)

        # 构造配置OSS网关需要的参数
        OSS_DATA["ossName"] = data["bucket"]
        OSS_DATA["ossgwMergeSize"] = 0
        OSS_DATA["bucketInfo"]["accessId"] = OBSInstanceOss.aes_decrypt(data["spare_key"]["ak"], self.secret_key)
        OSS_DATA["bucketInfo"]["accessKey"] = OBSInstanceOss.eisoo_rsa_encrypt(OBSInstanceOss.aes_decrypt(data["spare_key"]["sk"], self.secret_key))
        OSS_DATA["bucketInfo"]["name"] = data["bucket"]
        OSS_DATA["bucketInfo"]["provider"] = data["obs_type"]
        OSS_DATA["bucketInfo"]["serverName"] = data["region_endpoint"]
        OSS_DATA["bucketInfo"]["region"] = data.get("region", "")
        return OSS_DATA

    @retry_by_exception(attempt=3, sleep_time=5)
    def verify_token(self, token_id):
        method = "POST"
        body = {"Token": token_id}
        # signature = self.get_signature(OATUH_URI, method, body)
        signature = self.gen_signature(self.secret_key, OATUH_URI, method, body)
        url = OATUH_URL.format(
            oauth_host=self.oauth_host,
            access_id=self.access_key,
            signature=signature,
        )
        body = json.dumps(body)
        headers = {
            'Content-Type': 'application/json'
        }
        response = requests.request(method, url, headers=headers, data=body)
        if response.status_code == 200 and response.json()["Data"]["Validate"] == 1:
            return
        else:
            error = (
                f"verify token failed; signature: {signature}, "
                f"code: {response.status_code}, msg: {response.content}, req_data: {body}"
            )
            raise Exception(error)

    def gen_signature(self, access_key, request_uri, request_type='GET', request_body=None, request_parameter=None):
        """
        生成签名-账户中心
        signature = base64.b64encode( HMAC-SHA256(access_key, UTF-8-Encoding-Of( StringToSign )))
        StringToSign = HTTPMethod + "\n" + URI + "\n" + Body(Parameter)
        """
        check_str = ''
        string_to_sign = request_type + '\n' + request_uri

        if request_body:
            check_str = json.dumps(request_body, cls=EisJsonEncoder)

        if request_parameter:
            keys = sorted(list(request_parameter.keys()))
            for k in keys:
                if k == 'access_id':
                    continue
                if check_str:
                    check_str += '&'
                check_str += f"{k}={request_parameter[k]}"               
        if check_str:
            string_to_sign += '\n' + check_str.replace(' ', '')
        signature_db_safe = base64.b64encode(
            hmac.new(str(access_key).encode('utf-8'), string_to_sign.encode('utf-8'), hashlib.sha256).digest())
        return urllib.parse.quote(signature_db_safe.decode('utf-8'))
