#!/usr/bin/env python
#-*- coding:utf-8 -*-

"""
多进程并发安全地配置管理类, 配置文件格式见 ConfigParser
"""

import os
import ConfigParser

from src.modules.pydeps import locklib, tracer, logger, raiser

ERR_PROVIDER = "safeconfig"
ERR_ID = 1


class SafeConfig:

    """
    多进程并发安全地配置管理类, 配置文件格式见 ConfigParser
    """

    def __init__(self):
        pass

    @classmethod
    @tracer.trace_func
    def set(cls, conf_file_path, section, option, value):
        """
        在配置文件中设置或更新指定配置项
        线程/进程安全
        若配置文件或section不存在，则自动创建
        """
        with locklib.FileLock(os.path.basename(conf_file_path), timeout=30, delay=0.1):
            # 自动创建配置目录及文件
            conf_dir = os.path.dirname(conf_file_path)
            if not os.path.exists(conf_dir):
                os.makedirs(conf_dir)
            if not os.path.exists(conf_file_path):
                fobj = open(conf_file_path, 'w')
                fobj.close()

            config = ConfigParser.ConfigParser()
            config.read(conf_file_path)
            if not config.has_section(section):
                config.add_section(section)
            config.set(section, option, value)

            with open(conf_file_path, 'w') as fobj:
                config.write(fobj)

        if option not in ['db_password']:
            # 包含密码项,不打日志
            logger.syslog(owner="SafeConfig",
                          msg="Set config [%s]%s=%s in %s." % (section,
                                                               option, value, conf_file_path))

    @classmethod
    @tracer.trace_func
    def get(cls, conf_file_path, section, option):
        """
        在配置文件中获取指定配置项
        线程/进程安全
        若配置文件或option不存在，且设置了default_value，则返回default_value，否则抛错.
        """
        value = None

        with locklib.FileLock(os.path.basename(conf_file_path), timeout=30, delay=0.1):
            if os.path.exists(conf_file_path):
                config = ConfigParser.ConfigParser()
                config.read(conf_file_path)
                if config.has_option(section, option):
                    value = config.get(section, option)

        if value is None:
            errmsg = "Not found [%s]%s in %s." % (section, option, conf_file_path)
            raiser.raise_e(ERR_PROVIDER,
                           ERR_ID,
                           errmsg)

        return value

    @classmethod
    @tracer.trace_func
    def get_section(cls, conf_file_path):
        """获取所有section"""
        with locklib.FileLock(os.path.basename(conf_file_path), timeout=30, delay=0.1):
            if os.path.exists(conf_file_path):
                config = ConfigParser.ConfigParser()
                config.read(conf_file_path)
                return config.sections()
