#!/usr/bin/env python
# -*- coding:utf-8 -*-

"""
集群节点配置文件访问公共类
"""

import os

from src.modules.pydeps import safeconfig, tracer


class NodeConfig:
    """
    集群配置文件访问类
    """
    def __init__(self):
        pass

    config_file = "/sysvol/conf/nodeinfo.conf"

    @classmethod
    @tracer.trace_func
    def get_node_uuid(cls):
        """
        获取集群节点uuid
        """
        return safeconfig.SafeConfig.get(conf_file_path=cls.config_file,
                                         section="node",
                                         option="node_uuid")

    @classmethod
    @tracer.trace_func
    def set_node_uuid(cls, node_uuid):
        """
        设置集群节点uuid
        """
        safeconfig.SafeConfig.set(conf_file_path=cls.config_file,
                                  section="node",
                                  option="node_uuid",
                                  value=node_uuid)

    @classmethod
    @tracer.trace_func
    def file_exists(cls):
        """
        查询配置文件是否存在
        @return bool
        """
        return os.path.exists(cls.config_file)
