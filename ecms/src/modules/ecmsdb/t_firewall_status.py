#!/usr/bin/env python
# -*- coding:utf-8 -*-

"""集群 ecms.firewall_status 数据表管理模块"""

from src.modules.pydeps import logger, tracer
from src.modules.ecmsdb import err, db_conn

MODULE_NAME = err.ECMSDB_NAME


class TFirewallStatusDBManager(object):
    """
    firewall_status 数据表管理模块
    """
    def __init__(self):
        """
        pass
        """

    @classmethod
    @tracer.trace_func
    def get_status(cls, key):
        """
        根据子系统名字,获取开启状态
        """
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_status` FROM `firewall_status` WHERE `f_sys_name` = %s"""
        result = conn.fetch_one_result(sql, key)

        if result["f_status"]:
            return True
        else:
            return False

    @classmethod
    @tracer.trace_func
    def update_status(cls, key, value):
        """
        更新 firewall_status 表中指定键信息
        """
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()

        sql = """
        UPDATE `firewall_status`
        SET `f_status` = %s
        WHERE `f_sys_name` = '%s'
        """ % (value, key)
        conn.update(sql)
        logger.syslog(
            MODULE_NAME,
            "Updated firewall_status.f_sys_name {0} -> {1}."
            .format(key, value))
