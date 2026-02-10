#!/usr/bin/env python
# -*- coding:utf-8 -*-
import copy
import json
import re
from typing import Dict, Tuple, Union

import requests

from src.common import global_param
from src.common.ignore_case_configparser import IgnoreCaseConfigParser
from src.common.log_util import log_response, logger
from src.common.wrapper import retry_by_exception
from src.utils.net import get_host_for_url

from src.clients.k8s import K8SClient
from src.clients.cms import CMSClient, CMSObject

requests.packages.urllib3.disable_warnings()
is_match = lambda p, s: re.match(p, s) is not None

CMS_REQUEST_DATA = {"name": "", "use": "", "data": {}, "encrypt_field": []}  # 服务的名字，表示该服务的配置  # 标识使用的配置项  # 需要加密的字段


def get_self_namespace() -> str:
    return K8SClient.instance().rest_client.self_namespace


@retry_by_exception(attempt=12, sleep_time=5)
def read_conf_in_config(config_name: "str", file_name: "str") -> Union["IgnoreCaseConfigParser", "str"]:
    """
    获取configMap中<file_name>.yaml文件的两层dict对象，转为ConfigPaser
    @return ConfigPaser()/''
    """
    logger.info(msg=f"read config from cms (name with {config_name})")
    cms_object = CMSClient.instance().get_cms_data(config_name)
    conf = _dict_to_config_parser(cms_object.real_data)
    return conf


@retry_by_exception(attempt=12, sleep_time=5)
def get_ossgateway_info():
    """获取数据库连接信息"""
    cms_object = CMSClient.instance().head_cms_data("ossgateway")
    return cms_object.real_data if cms_object else {}

def save_ossgateway_info(data: dict):
    cms_object = CMSClient.instance().head_cms_data("ossgateway")
    if not cms_object:
        cms_object = CMSObject.create("ossgateway", data)
    else:
        cms_object.real_data = data
    cms_object.save(CMSClient.instance())

@retry_by_exception(attempt=12, sleep_time=5)
def get_authserver_info():
    """获取autherserver信息"""
    config_name = "authserver"
    cmsobject = CMSClient.instance().get_cms_data(config_name)
    return cmsobject.real_data


def get_third_app_depserviceinfo(chart):
    chart_name = chart.get("name")
    chart_values = chart.get("values")
    if "thirdAppDepServices" not in chart_values:
        logger.debug("%s have not thirdAppDepServices ,skip." % chart_name)
        return None

    third_depservice_info = chart_values.get("thirdAppDepServices")
    return third_depservice_info


def check_json_format(raw_msg):
    """
    用于判断一个字符串是否符合Json格式
    :param self:
    :return:
    """
    if isinstance(raw_msg, (str, str)):  # 首先判断变量是否为字符串
        try:
            json.loads(raw_msg)
        except ValueError:
            return False
        return True
    else:
        return False



def _dict_to_config_parser(dict_obj: "Dict[str, dict]") -> "IgnoreCaseConfigParser":
    """
    dict_obj 需要为两层的字典，
    即：
    {
        'sectionA': {
            "option1": "value1",
            "option2": "value2",
            "option3": "value3"
        },
        'sectionB': {
            "option4": "value4",
            "option5": "value5"

        }
    }
    """
    conf = IgnoreCaseConfigParser()
    conf.read_dict(dict_obj)
    return conf




hydra_admin = ""

def get_hydra_admin_by_cache():
    global hydra_admin
    if hydra_admin:
        return hydra_admin
    try:
        access_conf = read_conf_in_config(global_param.SERVICE_ACCESS_CONFIG, global_param.SERVICE_ACCESS_FILE_NAME)
        hydra_admin_host = access_conf.get("hydra", "administrativeHost")
        hydra_admin_port = access_conf.get("hydra", "administrativePort")
    except Exception as e:
        logger.error(f"get hydra admin info failed: {str(e)}")
        return "hydra-admin:4445"
    else:
        hydra_admin = f"{hydra_admin_host}:{hydra_admin_port}"
        return hydra_admin




