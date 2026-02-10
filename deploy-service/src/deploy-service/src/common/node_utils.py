#!/usr/bin/env python
# -*- coding: utf-8 -*-
# @Time    : 2021/4/17 22:07
# @Author  : Jimmy.li
# @Email   : jimmy.li@aishu.cn

import requests

from src.common import global_param, utils
from src.common.log_util import log_response, logger
from src.utils.net import get_host_for_url

from . import lib


def __get_node_mgnt_api_url(path: str):
    node_mgnt_host, node_mgnt_port = node_mgnt_connent()
    return f"http://{get_host_for_url(node_mgnt_host)}:{node_mgnt_port}{path}"


def get_all_node_info():
    resp = requests.get(__get_node_mgnt_api_url("/api/nodemgnt/v1/nodes"))
    log_response(resp)
    if resp.status_code not in range(200, 300):
        raise Exception("get all nodes info falied ," % resp.content)

    nodes_info = resp.json()
    return nodes_info


def get_node_info_by_ip(ip):
    logger.info("[PerformanceLogger] get_node_info_by_ip start")
    resp = requests.get(__get_node_mgnt_api_url(f"/api/nodemgnt/v1/nodes?node_ip={ip}"))
    log_response(resp)
    if resp.status_code not in range(200, 300):
        raise Exception("get all nodes info falied ," % resp.content)

    node_info = resp.json()
    logger.info("[PerformanceLogger] get_node_info_by_ip succeed")
    return node_info


def get_ha_info():
    resp = requests.get(__get_node_mgnt_api_url("/api/nodemgnt/v1/ha"))
    log_response(resp)
    if resp.status_code not in range(200, 300):
        raise Exception("get all nodes info falied ," % resp.content)

    ha_info = resp.json()
    return ha_info


node_mgnt_host = None
node_mgnt_port = None


def node_mgnt_connent():
    global node_mgnt_host, node_mgnt_port
    if node_mgnt_connent is None or node_mgnt_port is None:
        access_conf = utils.read_conf_in_config(
            global_param.SERVICE_ACCESS_CONFIG, global_param.SERVICE_ACCESS_FILE_NAME
        )
        node_mgnt_host = access_conf.get("nodemgnt", "host")
        node_mgnt_port = access_conf.get("nodemgnt", "port")
    return node_mgnt_host, node_mgnt_port


def get_service_installed_nodes(service_name):
    installed_nodes = []
    node_infos = get_all_node_info()
    for node_info in node_infos:
        if lib.NODE_ROLE_PRE + service_name in node_info["node_lables"].keys():
            installed_nodes.append(node_info["node_ip"])

    return installed_nodes
