#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import subprocess

from src.common.log_util import logger


SERVICE_EXTERNAL_PORT_KEY = "micro_service_external_port"

SERVICE_INTERNAL_PORT_KEY = "micro_service_internal_port"

# k8s 节点角色标签前缀
NODE_ROLE_PRE = "node-role.aishu.cn/"


def serviceConf():

    ServiceConf = {
        "service_name": "",
        "node_ips": [],
    }
    return ServiceConf


def exec_command(command, shell=True):
    """
    执行shell命令
    """
    logger.info("start exec[%s] begin." % command)
    if isinstance(command, str):
        command = [command]
    proc = subprocess.Popen(command, shell=shell, close_fds=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    (outdata, _) = proc.communicate()
    if proc.returncode != 0:
        logger.info("exec [%s] error %s." % (command, outdata))
        return False, "exec_command %s error: %s" % (command, outdata)
    return True, ""
