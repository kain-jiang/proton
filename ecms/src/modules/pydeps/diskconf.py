#!/usr/bin/env python
#-*- coding:utf-8 -*-

"""
节点磁盘配置文件访问公共类
"""

import os
from src.modules.pydeps import safeconfig, tracer


class DiskConfig:
    """
    集群配置文件访问类
    """
    config_file = "/sysvol/conf/disk.conf"

    @classmethod
    @tracer.trace_func
    def get_mount_extend_args(cls):
        """
        获取磁盘挂载扩展参数
        """
        return safeconfig.SafeConfig.get(conf_file_path=cls.config_file,
                                         section="disk",
                                         option="mount_extend_args")

    @classmethod
    @tracer.trace_func
    def set_mount_extend_args(cls, args):
        """
        设置disk.conf mount扩展参数
        """
        safeconfig.SafeConfig.set(conf_file_path=cls.config_file,
                                  section="disk",
                                  option="mount_extend_args",
                                  value=args)

    @classmethod
    @tracer.trace_func
    def exists_disk_conf(cls):
        """
        查询磁盘配置文件(disk.conf)是否存在
        """
        return os.path.exists(cls.config_file)
