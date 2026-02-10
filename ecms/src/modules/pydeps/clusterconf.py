#!/usr/bin/env python
# -*- coding:utf-8 -*-

"""
集群配置文件访问公共类
"""
import os

import safeconfig, tracer


class ClusterConfig:
    """
    集群配置文件访问类
    """
    def __init__(self):
        pass

    config_file = "/sysvol/conf/cluster.conf"

    @classmethod
    @tracer.trace_func
    def get_db_host(cls):
        """获取数据库地址"""
        return safeconfig.SafeConfig.get(conf_file_path=cls.config_file,
                                         section="cluster",
                                         option="db_host")

    @classmethod
    @tracer.trace_func
    def get_db_port(cls):
        """获取数据库端口"""
        return int(safeconfig.SafeConfig.get(conf_file_path=cls.config_file,
                                             section="cluster",
                                             option="db_port"))

    @classmethod
    @tracer.trace_func
    def set_db_host(cls, db_host):
        """设置数据库地址"""
        safeconfig.SafeConfig.set(conf_file_path=cls.config_file,
                                  section="cluster",
                                  option="db_host",
                                  value=db_host)

    @classmethod
    @tracer.trace_func
    def set_db_port(cls, db_port):
        """设置数据库端口"""
        safeconfig.SafeConfig.set(conf_file_path=cls.config_file,
                                  section="cluster",
                                  option="db_port",
                                  value=db_port)

    @classmethod
    @tracer.trace_func
    def get_app_master_node_uuid(cls):
        """
        获取应用主节点 UUID
        缓存, 用于判断应用主节点是否发生改变
        通过 ECMSManager.get_app_master_node_info
        获取应用主节点
        """
        return safeconfig.SafeConfig.get(
            conf_file_path=cls.config_file,
            section='application',
            option='master_node_uuid'
        )

    @classmethod
    @tracer.trace_func
    def set_app_master_node_uuid(cls, node_uuid):
        """
        设置应用主节点 UUID
        缓存, 用于判断应用主节点是否发生改变
        """
        return safeconfig.SafeConfig.set(
            conf_file_path=cls.config_file,
            section='application',
            option='master_node_uuid',
            value=node_uuid
        )

    @classmethod
    @tracer.trace_func
    def file_exists(cls):
        """
        查询配置文件是否存在
        @return bool
        """
        return os.path.exists(cls.config_file)

    @classmethod
    @tracer.trace_func
    def section_exists(cls, section):
        """判断指定section是否存在"""
        return section in safeconfig.SafeConfig.get_section(cls.config_file)

    @classmethod
    @tracer.trace_func
    def set_nsqlookupd_host(cls, nsqlookup_host):
        """
        设置nsqlookupd访问ip
        """
        safeconfig.SafeConfig.set(conf_file_path=cls.config_file,
                                  section="nsqlookupd",
                                  option="nsqlookupd_host",
                                  value=nsqlookup_host)
    @classmethod
    @tracer.trace_func
    def set_nsqlookupd_port(cls, nsqlookup_host, connect_type):
        """
        设置nsqlookupd访问port
        @param connect_type 参数：'http' or 'tcp'
        """
        if connect_type == 'http':
            option = 'http_port'
        elif connect_type == 'tcp':
            option = 'tcp_port'
        else:
            raise Exception("Param error")
        safeconfig.SafeConfig.set(conf_file_path=cls.config_file,
                                  section="nsqlookupd",
                                  option=option,
                                  value=nsqlookup_host)
    @classmethod
    @tracer.trace_func
    def get_nsqlookupd_host(cls):
        """获取nsqlookupd连接ip"""
        return safeconfig.SafeConfig.get(conf_file_path=cls.config_file,
                                         section="nsqlookupd",
                                         option="nsqlookupd_host")

    @classmethod
    @tracer.trace_func
    def get_nsqlookupd_port(cls, connect_type):
        """获取nsqlookupd连接port"""
        if connect_type == 'http':
            option = 'http_port'
        elif connect_type == 'tcp':
            option = 'tcp_port'
        else:
            raise Exception("Param error")
        return int(safeconfig.SafeConfig.get(conf_file_path=cls.config_file,
                                             section="nsqlookupd",
                                             option=option))

# =================================================================================================
# 第三方数据库部分
# =================================================================================================
    @classmethod
    @tracer.trace_func
    def if_use_external_db(cls):
        """数据库是否为第三方数据库"""
        use = safeconfig.SafeConfig.get(conf_file_path=cls.config_file,
                                        section="cluster",
                                        option="use_external_db")

        if use == 'True':
            return True
        else:
            return False

    @classmethod
    @tracer.trace_func
    def use_external_db(cls, value):
        """设置第三方数据库"""
        safeconfig.SafeConfig.set(conf_file_path=cls.config_file,
                                  section="cluster",
                                  option="use_external_db",
                                  value=value)

    @classmethod
    @tracer.trace_func
    def set_external_db_info(cls, info):
        """设置第三方数据库连接信息"""
        safeconfig.SafeConfig.set(conf_file_path=cls.config_file,
                                  section='cluster',
                                  option="db_host",
                                  value=info["db_host"])
        safeconfig.SafeConfig.set(conf_file_path=cls.config_file,
                                  section='cluster',
                                  option="db_port",
                                  value=info["db_port"])
        safeconfig.SafeConfig.set(conf_file_path=cls.config_file,
                                  section='cluster',
                                  option="db_user",
                                  value=info["db_user"])
        safeconfig.SafeConfig.set(conf_file_path=cls.config_file,
                                  section='cluster',
                                  option="db_password",
                                  value=cls.encrypt(info["db_password"]))

    @classmethod
    @tracer.trace_func
    def get_external_db_info(cls):
        """获取第三方数据库连接信息"""
        info = dict()
        info['db_host'] = safeconfig.SafeConfig.get(conf_file_path=cls.config_file,
                                                    section='cluster',
                                                    option="db_host")
        info['db_port'] = safeconfig.SafeConfig.get(conf_file_path=cls.config_file,
                                                    section='cluster',
                                                    option="db_port")
        info['db_port'] = int(info['db_port'])
        info['db_user'] = safeconfig.SafeConfig.get(conf_file_path=cls.config_file,
                                                    section='cluster',
                                                    option="db_user")

        pwd = safeconfig.SafeConfig.get(conf_file_path=cls.config_file,
                                        section='cluster',
                                        option="db_password")
        info["db_password"] = cls.decrypt(pwd)

        return info

    @classmethod
    @tracer.trace_func
    def update_external_db_info(cls, info):
        """修改第三方数据库用户名/密码"""
        safeconfig.SafeConfig.set(conf_file_path=cls.config_file,
                                  section='cluster',
                                  option="db_user",
                                  value=info["db_user"])
        safeconfig.SafeConfig.set(conf_file_path=cls.config_file,
                                  section='cluster',
                                  option="db_password",
                                  value=cls.encrypt(info["db_password"]))

    @classmethod
    @tracer.trace_func
    def encrypt(cls, s):
        b = bytearray(str(s))
        n = len(b)
        c = bytearray(n * 2)
        j = 0
        for i in range(0, n):
            b1 = b[i]
            b2 = b1 ^ 32
            c1 = b2 % 16
            c2 = b2 // 16
            c1 = c1 + 65
            c2 = c2 + 65
            c[j] = c1
            c[j + 1] = c2
            j = j + 2
        return str(c)

    @classmethod
    @tracer.trace_func
    def decrypt(cls, s):
        c = bytearray(str(s))
        n = len(c)
        if n % 2 != 0:
            return ""
        n = n // 2
        b = bytearray(n)
        j = 0
        for i in range(0, n):
            c1 = c[j]
            c2 = c[j + 1]
            j = j + 2
            c1 = c1 - 65
            c2 = c2 - 65
            b2 = c2 * 16 + c1
            b1 = b2 ^ 32
            b[i] = b1
        try:
            return str(b)
        except:
            return "failed"
