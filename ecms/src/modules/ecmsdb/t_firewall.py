#!/usr/bin/env python
# -*- coding:utf-8 -*-

"""ecms.t_firewall 数据表管理模块"""

import src.modules.pydeps.tracer, src.modules.pydeps.raiser
from src.modules.ecmsdb import err
from src.modules.ecmsdb import db_conn
from src.modules.pydeps.logger import syslog_info

from ECMSManager.ttypes import ncTFirewallInfo

MODULE_NAME = err.ECMSDB_NAME


class TFirewallDBManager(object):
    """
    t_firewall 数据表管理模块
    """
    @classmethod
    @src.modules.pydeps.tracer.trace_func
    def add_firewall_rule(cls, firewall_info):
        """
        添加一条防火墙规则到数据库
        @param  <ncTFirewallInfo> firewall_info 防火墙信息结构
        """
        if cls.exists_firewall_rule(firewall_info):
            src.modules.pydeps.raiser.raise_e(err.ECMSDB_NAME, err.ECMSDB_ERR_INTERNAL,
                           "firewall rule %r already existed." % firewall_info)

        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        # 插入新记录
        values = {'f_port': firewall_info.port,
                  'f_protocol': firewall_info.protocol,
                  'f_source_net': firewall_info.source_net,
                  'f_dest_net': firewall_info.dest_net,
                  'f_sys_role': firewall_info.role_sys,
                  'f_service_desc': firewall_info.service_desc}
        conn.insert('firewall', values)
        syslog_info(MODULE_NAME, "Added firewall rule: %r." % firewall_info)

    @classmethod
    @src.modules.pydeps.tracer.trace_func
    def del_firewall_rule(cls, firewall_info):
        """删除一条防火墙规则"""
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """Delete FROM `firewall`
        WHERE `f_port`=%d and `f_protocol`='%s' and `f_source_net`='%s' and `f_dest_net`='%s';
        """ % (firewall_info.port,
               firewall_info.protocol,
               firewall_info.source_net,
               firewall_info.dest_net)

        conn.delete(sql)
        syslog_info(MODULE_NAME, "Deleted firewall rule %r." % firewall_info)

    @classmethod
    @src.modules.pydeps.tracer.trace_func
    def update_firewall_rule(cls, old_info, new_info):
        """
        更新一条防火墙规则
        @param <ncTFirewallInfo> old_info 待更新规则
        @param <ncTFirewallInfo> new_info 更新后的规则
        """
        if not cls.exists_firewall_rule(old_info):
            src.modules.pydeps.raiser.raise_e(err.ECMSDB_NAME, err.ECMSDB_ERR_INTERNAL,
                           "firewall rule %r not existed." % old_info)

        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_id` FROM `firewall`
        WHERE `f_port` = %d and `f_protocol`='%s' and `f_source_net`='%s' and `f_dest_net`='%s'
        """ % (old_info.port,
               old_info.protocol,
               old_info.source_net,
               old_info.dest_net)

        result = conn.fetch_one_result(sql)

        sql = """UPDATE `firewall`
        SET `f_port`=%d, `f_protocol`='%s', `f_source_net`='%s', `f_dest_net`='%s',
            `f_sys_role`='%s', `f_service_desc`='%s'
        WHERE `f_id` = %d;""" % (new_info.port,
                                 new_info.protocol,
                                 new_info.source_net,
                                 new_info.dest_net,
                                 new_info.role_sys,
                                 new_info.service_desc,
                                 result['f_id'])
        conn.update(sql)
        syslog_info(MODULE_NAME, "Updated firewall rule %r to rule %r" % (old_info, new_info))

    @classmethod
    @src.modules.pydeps.tracer.trace_func
    def get_firewall_rule_by_role(cls, sys_role):
        """获取集群节点基础放行端口"""
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_port`,
        `f_protocol`,
        `f_source_net`,
        `f_dest_net`,
        `f_sys_role`,
        `f_service_desc` FROM `firewall` WHERE `f_sys_role` = '%s'""" % (sys_role)
        result = conn.fetch_all_result(sql)
        firewall_info_list = list()

        for each_result in result:
            firewall_info = cls._record_to_firewall_info(each_result)
            firewall_info_list.append(firewall_info)
        return firewall_info_list

    @classmethod
    @src.modules.pydeps.tracer.trace_func
    def get_firewall_rule_by_role_on_local_db(cls, port, sys_role):
        """在本地数据库获取集群节点基础放行端口"""
        firewall_info_list = list()

        try:
            connector = db_conn.get_local_db_connector(port)
            conn = connector.get_db_operate_obj()
        except:
            syslog_info(MODULE_NAME, "Not found local database")
            return firewall_info_list

        sql = """SELECT `f_port`,
        `f_protocol`,
        `f_source_net`,
        `f_dest_net`,
        `f_sys_role`,
        `f_service_desc` FROM `firewall` WHERE `f_sys_role` = '%s'""" % (sys_role)
        result = conn.fetch_all_result(sql)

        for each_result in result:
            firewall_info = cls._record_to_firewall_info(each_result)
            firewall_info_list.append(firewall_info)
        return firewall_info_list

    @classmethod
    @src.modules.pydeps.tracer.trace_func
    def exists_firewall_rule(cls, firewall_info):
        """
        判断一条防火墙规则是否存在
        """
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """
        SELECT `f_id` FROM `firewall`
        WHERE `f_port` = %s and `f_protocol`='%s' and `f_source_net`='%s' and `f_dest_net`='%s'
            and `f_sys_role`='%s'
        """ % (firewall_info.port,
               firewall_info.protocol,
               firewall_info.source_net,
               firewall_info.dest_net,
               firewall_info.role_sys)
        result = conn.fetch_one_result(sql)

        if result is None:
            return False
        else:
            return True


# ============================================================================
# 内部函数
# ============================================================================
    @classmethod
    @src.modules.pydeps.tracer.trace_func
    def _record_to_firewall_info(cls, record):
        """
        单条查询结果转换成ncTFirewallInfo对象
        @param record 单条查询结果
        @return ncTFirewallInfo
        """
        if record is None:
            return None

        firewall_info = ncTFirewallInfo()
        firewall_info.port = record['f_port']
        firewall_info.protocol = record['f_protocol']
        firewall_info.source_net = record['f_source_net']
        firewall_info.dest_net = record['f_dest_net']
        firewall_info.service_desc = record['f_service_desc']
        firewall_info.role_sys = record['f_sys_role']
        return firewall_info
