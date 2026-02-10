#!/usr/bin/env python
# -*- coding:utf-8 -*-

"""ecms.node 数据表管理模块"""
import time
from copy import deepcopy

from src.modules.pydeps import raiser, tracer
from src.modules.pydeps.logger import syslog_info
from src.modules.ecmsdb import err, db_conn


MODULE_NAME = err.ECMSDB_NAME

ncTHaSys = {
    'NORMAL': 0,                         # 非ha节点
    'BASIC': 1,                          # 集群使用全局高可用，vip和ivip在同一节点情况
    'APP': 2,                            # 应用子系统vip
    'STORAGE': 3,                        # 存储子系统vip
    'DB': 4,                             # 数据库使用的ivip标识
}


ncTNodeInfo = {
    'node_uuid': None,             # 节点uuid
    'role_db': 0,                     #数据库节点标识(0:非数据库节点  1:数据库master节点 2:数据库slave节点)
    'role_ecms': 0,                   # 集群管理节点标识(0:非集群管理节点；1:集群管理主节点)
    'role_app': 0,                    # 应用节点标识   0:非应用节点；1:应用节点
    'role_storage': 0,                # 存储节点标识   0:非存储节点；1:存储节点
    'node_alias': '',            # 节点别名
    'node_ip': '',               # 节点加入集群使用的IP
    'is_online':  True,             # 节点在线状态(true:在线 | false离线)
    'is_ha': 0,                     # 节点是否是ha节点
    'is_etcd': 0,                  # 节点是否有etcd实例(0: 没有实例 1: 只有有一个实例 2: 有两个实例(ecms节点的备实例存在)
    'consistency_status': True,   # 节点一致性状态(true: 与集群状态一致 | false: 不一致)
}


ncTHaNodeInfo = {
    'node_uuid': '',        # 节点uuid
    'node_alias': '',       # 节点别名
    'node_ip': '',          # 节点ip
    'is_online': True,      # 节点在线状态
    'is_master': False,     # 节点是否是ha主节点
    'sys': ncTHaSys['BASIC'],   # vip所属子系统

}


class TNodeDBManager(object):
    """
    node 数据表管理模块
    """
    def __init__(self):
        """
        pass
        """
    @classmethod
    @tracer.trace_func
    def add_node_info(cls, node_info):
        """
        向t_node表中添加一条节点信息
        @param <ncTNodeInfo> node_info 节点minion_id
        @param <TNodeRole> node_info 节点角色信息
        @f_role_db str True 数据库节点/False 非数据库节点
        """
        if cls.exists_node_info(node_info.node_uuid):
            raiser.raise_e(err.ECMSDB_NAME, err.ECMSDB_ERR_INTERNAL,
                           "node %s already existed." % (node_info.node_uuid))

        # 参数检测
        if node_info.role_db not in [0, 1, 2]:
            raiser.raise_e(err.ECMSDB_NAME, err.ECMSDB_ERR_INTERNAL,
                           "Parameter error, f_role_db: %s" % node_info.role_db)
        if node_info.role_ecms not in [1, 0]:
            raiser.raise_e(err.ECMSDB_NAME, err.ECMSDB_ERR_INTERNAL,
                           "Parameter error, f_role_ecms: %s" % node_info.role_ecms)

        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()

        # 插入新记录
        values = {'f_node_uuid': node_info.node_uuid,
                  'f_role_db': node_info.role_db,
                  'f_role_ecms': node_info.role_ecms,
                  'f_role_app': node_info.role_app,
                  'f_role_storage': node_info.role_storage,
                  'f_node_alias': node_info.node_alias,
                  'f_node_ip': node_info.node_ip,
                  'f_heartbeat_time': cls._get_time_now()}
        conn.insert('node', values)
        syslog_info(MODULE_NAME, "Added node, %r." % node_info)

    @classmethod
    @tracer.trace_func
    def delete_node_info(cls, node_uuid):
        """
        删除 node 表中信息
        """
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """
        Delete FROM `node`
        WHERE `f_node_uuid` = %s
        """
        conn.delete(sql, node_uuid)
        syslog_info(MODULE_NAME, "Deleted node {0}.".format(node_uuid))

    @classmethod
    @tracer.trace_func
    def update_node_info(cls, node_info):
        """更改节点信息"""
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()

        escaped_node_alias = connector.escape(node_info.node_alias)
        sql = """UPDATE `node`
        SET `f_role_db` = %d,
        `f_role_app` = %d,
        `f_role_storage` = %d,
        `f_role_ecms` = %d,
        `f_node_alias` = '%s',
        `f_node_ip` = '%s'
        WHERE `f_node_uuid` = '%s';""" % (node_info.role_db,
                                          node_info.role_app,
                                          node_info.role_storage,
                                          node_info.role_ecms,
                                          escaped_node_alias,
                                          node_info.node_ip,
                                          node_info.node_uuid)

        conn.update(sql)
        syslog_info(MODULE_NAME, "Updated node: %r" % node_info)

    @classmethod
    @tracer.trace_func
    def update_node_ha_status(cls, node_uuid, ha_status):
        """更改节点ha状态"""
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()

        sql = """UPDATE `node` SET `f_ha` = %s
                 WHERE `f_node_uuid` = '%s';""" % (ha_status, node_uuid)
        conn.update(sql)
        syslog_info(MODULE_NAME, "Updated node[%s] ha status to %s" % (node_uuid, ha_status))

    @classmethod
    @tracer.trace_func
    def update_node_etcd_status(cls, node_uuid, etcd_status):
        """更改节点etcd状态"""
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()

        sql = """UPDATE `node` SET `f_etcd` = %s
                 WHERE `f_node_uuid` = '%s';""" % (etcd_status, node_uuid)
        conn.update(sql)
        syslog_info(MODULE_NAME, "Updated node[%s] etcd status to %s" % (node_uuid, etcd_status))

    @classmethod
    @tracer.trace_func
    def get_node_info(cls, node_uuid):
        """
        获取指定节点信息
        @return <ncTNodeInfo> 节点信息结构
        """
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_node_uuid`,
        `f_role_db`,
        `f_role_ecms`,
        `f_role_app`,
        `f_role_storage`,
        `f_node_alias`,
        `f_node_ip`,
        `f_ha`,
        `f_etcd`,
        `f_consistency_status` FROM `node` WHERE `f_node_uuid` = '%s'""" % node_uuid
        result = conn.fetch_one_result(sql)
        if result is None:
            raiser.raise_e(err.ECMSDB_NAME, err.ECMSDB_ERR_INTERNAL,
                           "node %s not existed." % node_uuid)
        return cls._record_to_node_info(result)

    @classmethod
    @tracer.trace_func
    def get_node_info_on_local_db(cls, port, node_uuid):
        """
        在本地数据库获取指定节点信息
        @param port 本地数据库端口
        @param node_uuid 节点uuid
        """
        try:
            connector = db_conn.get_local_db_connector(port)
            conn = connector.get_db_operate_obj()
        # 没有本地数据库, 返回节点结构
        except:
            syslog_info(MODULE_NAME, "Not found local database")
            return ncTNodeInfo()

        sql = """SELECT `f_node_uuid`,
        `f_role_db`,
        `f_role_ecms`,
        `f_role_app`,
        `f_role_storage`,
        `f_node_alias`,
        `f_node_ip`,
        `f_ha`,
        `f_etcd`,
        `f_consistency_status` FROM `node` WHERE `f_node_uuid` = '%s'""" % node_uuid
        result = conn.fetch_one_result(sql)
        if result is None:
            raiser.raise_e(err.ECMSDB_NAME, err.ECMSDB_ERR_INTERNAL,
                           "node %s not existed." % node_uuid)
        node_info = cls._local_record_to_node_info(result)
        return node_info

    @classmethod
    @tracer.trace_func
    def get_all_node_info_on_local_db(cls, port):
        """在本地数据库中获取所有节点信息"""
        connector = db_conn.get_local_db_connector(port)
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_node_uuid`,
        `f_role_db`,
        `f_role_ecms`,
        `f_role_app`,
        `f_role_storage`,
        `f_node_alias`,
        `f_node_ip`,
        `f_ha`,
        `f_etcd`,
        `f_consistency_status` FROM `node`"""
        result = conn.fetch_all_result(sql)

        node_info_list = list()
        # 循环构造节点信息结构体
        for each_result in result:
            node_info = cls._local_record_to_node_info(each_result)
            node_info_list.append(node_info)
        return node_info_list

    @classmethod
    @tracer.trace_func
    def get_node_info_by_ip(cls, ipaddr):
        """
        获取指定节点信息
        @return <ncTNodeInfo> 节点信息结构
        """
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_node_uuid`,
        `f_role_db`,
        `f_role_ecms`,
        `f_role_app`,
        `f_role_storage`,
        `f_node_alias`,
        `f_node_ip`,
        `f_ha`,
        `f_etcd`,
        `f_consistency_status` FROM `node` WHERE `f_node_ip` = '%s'""" % ipaddr
        result = conn.fetch_one_result(sql)
        if result is None:
            raiser.raise_e(err.ECMSDB_NAME, err.ECMSDB_ERR_INTERNAL, "ip %s not existed." % ipaddr)
        return cls._record_to_node_info(result)

    @classmethod
    @tracer.trace_func
    def get_all_node_info(cls):
        """获取所有节点信息"""
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_node_uuid`,
        `f_role_db`,
        `f_role_ecms`,
        `f_role_app`,
        `f_role_storage`,
        `f_node_alias`,
        `f_node_ip`,
        `f_ha`,
        `f_etcd`,
        `f_consistency_status` FROM `node`"""
        result = conn.fetch_all_result(sql)
        node_info_list = list()

        # 循环构造节点信息结构体
        for each_result in result:
            node_info_list.append(cls._record_to_node_info(each_result))

        return node_info_list

    @classmethod
    @tracer.trace_func
    def get_ha_node_info(cls):
        """获取所有ha节点"""
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_node_uuid`,
        `f_node_alias`,
        `f_node_ip`,
        `f_ha` FROM `node` WHERE `f_ha` != 0;"""
        result = conn.fetch_all_result(sql)
        node_info_list = list()

        # 循环构造节点信息结构体
        for each_result in result:
            node_info_list.append(cls._record_to_ha_node_info(each_result))

        return node_info_list

    @classmethod
    @tracer.trace_func
    def get_ha_node_info_by_sys(cls, sys):
        """根据ha系统标签获取指定ha集群节点"""
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_node_uuid`,
        `f_node_alias`,
        `f_node_ip`,
        `f_ha` FROM `node` WHERE `f_ha` = %d;""" % sys
        result = conn.fetch_all_result(sql)
        node_info_list = list()

        # 循环构造节点信息结构体
        for each_result in result:
            node_info_list.append(cls._record_to_ha_node_info(each_result))

        return node_info_list

    @classmethod
    @tracer.trace_func
    def get_role_db_uuid(cls):
        """
        获取数据库子系统所有的节点uuid
        @return list uuid列表
        """
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_node_uuid` FROM `node` WHERE `f_role_db` != 0"""
        result = conn.fetch_all_result(sql)
        return list(each['f_node_uuid'] for each in result)

    @classmethod
    @tracer.trace_func
    def get_role_db_master_uuid(cls):
        """获取数据库主库节点信息"""
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_node_uuid` FROM `node` WHERE `f_role_db` = 1"""
        result = conn.fetch_one_result(sql)
        if result is None:
            return ""
        return result['f_node_uuid']

    @classmethod
    @tracer.trace_func
    def get_role_db_slave_uuid(cls):
        """获取数据库主库节点信息"""
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_node_uuid` FROM `node` WHERE `f_role_db` = 2"""
        result = conn.fetch_one_result(sql)
        if result is None:
            return ""
        return result['f_node_uuid']

    @classmethod
    @tracer.trace_func
    def get_role_app_uuid(cls):
        """
        获取数据库子系统所有的节点uuid
        @return list uuid列表
        """
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_node_uuid` FROM `node` WHERE `f_role_app` != 0"""
        result = conn.fetch_all_result(sql)
        return list(each['f_node_uuid'] for each in result)

    @classmethod
    @tracer.trace_func
    def get_role_storage_uuid(cls):
        """
        获取数据库子系统所有的节点uuid
        @return list uuid列表
        """
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_node_uuid` FROM `node` WHERE `f_role_storage` != 0"""
        result = conn.fetch_all_result(sql)
        return list(each['f_node_uuid'] for each in result)

    @classmethod
    @tracer.trace_func
    def get_role_ecms_uuid(cls):
        """
        获取集群管理子系统所有的节点uuid
        @return list uuid列表
        """
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_node_uuid` FROM `node` WHERE `f_role_ecms` != 0"""
        result = conn.fetch_all_result(sql)
        return list(each['f_node_uuid'] for each in result)

    @classmethod
    @tracer.trace_func
    def get_role_ecms_master_uuid(cls):
        """获取ecms主节点uuid"""
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_node_uuid` FROM `node` WHERE `f_role_ecms` = 1"""
        result = conn.fetch_one_result(sql)
        return result['f_node_uuid']

    @classmethod
    @tracer.trace_func
    def get_role_ecms_slave_uuid(cls):
        """获取ecms从节点uuid"""
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_node_uuid` FROM `node` WHERE `f_role_ecms` = 2"""
        result = conn.fetch_one_result(sql)
        if not result:
            return None
        return result['f_node_uuid']

    @classmethod
    @tracer.trace_func
    def get_role_ecms_ip(cls):
        """
        获取集群管理子系统所有的节点ip
        @return list ip列表
        """
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_node_ip` FROM `node` WHERE `f_role_ecms` = 1"""
        result = conn.fetch_all_result(sql)
        return list(each['f_node_ip'] for each in result)

    @classmethod
    @tracer.trace_func
    def get_node_info_by_etcd_status(cls):
        """获取集群中所有装有etcd实例的节点信息"""
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_node_uuid`,
        `f_role_db`,
        `f_role_ecms`,
        `f_role_app`,
        `f_role_storage`,
        `f_node_alias`,
        `f_node_ip`,
        `f_ha`,
        `f_etcd`,
        `f_consistency_status` FROM `node` WHERE `f_etcd` != 0"""
        result = conn.fetch_all_result(sql)
        node_info_list = list()

        # 循环构造节点信息结构体
        for each_result in result:
            node_info_list.append(cls._record_to_node_info(each_result))

        return node_info_list

    @classmethod
    @tracer.trace_func
    def exists_node_info(cls, node_uuid):
        """
        判断 node 表中是否已经存在信息
        @param string node_uuid: 节点uuid
        @return bool: True 存在, False 不存在
        """
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()

        sql = """
        SELECT `f_node_uuid` FROM `node`
        WHERE `f_node_uuid` = '%s'
        """ % node_uuid

        result = conn.fetch_one_result(sql)
        if result is None:
            return False
        else:
            return True

    @classmethod
    @tracer.trace_func
    def exists_node_alias(cls, node_alias):
        """判断t_node表中是否已经存在信息"""
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()

        escaped_node_alias = connector.escape(node_alias)
        sql = """
        SELECT `f_node_alias` FROM `node`
        WHERE `f_node_alias` = '%s'
        """ % escaped_node_alias
        result = conn.fetch_one_result(sql)
        if result is None:
            return False
        else:
            return True

    @classmethod
    @tracer.trace_func
    def get_consistency_status_by_node_uuid(cls, node_uuid):
        """获取指定节点的一致性状态"""
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_consistency_status`
                 FROM `node`
                 WHERE `f_node_uuid`='%s';""" % node_uuid
        result = conn.fetch_one_result(sql)

        if result is not None:
            return result['f_consistency_status']
        else:
            return None

    @classmethod
    @tracer.trace_func
    def update_consistency_status_by_node_uuid(cls, node_uuid, consistency_status):
        """更新指定节点的ssh连接信息"""
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()

        sql = """UPDATE `node`
                 SET `f_consistency_status` = %d
                 WHERE `f_node_uuid` = '%s';""" % (consistency_status, node_uuid)
        conn.update(sql)

# ============================================================================
# 处理节点心跳
# ============================================================================
    @classmethod
    @tracer.trace_func
    def get_heartbeat_time(cls, node_uuid):
        """
        根据节点 uuid 获取节点心跳时间
        若不存在该记录，则返回 None
        @return string heartbeat_time
        """
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()

        sql = """
        SELECT `f_heartbeat_time` FROM `node`
        WHERE `f_node_uuid` = %s
        """
        result = conn.fetch_one_result(sql, node_uuid)
        if result is not None:
            return result["f_heartbeat_time"]
        else:
            return None

    @classmethod
    @tracer.trace_func
    def get_heartbeat_seconds_passed_by(cls, node_uuid):
        """
        根据节点 uuid 获取节点心跳更新时间距离当前时间的秒数
        若不存在该记录，则返回 None
        @return float seconds
        """
        heartbeat_time = cls.get_heartbeat_time(node_uuid)
        heartbeat_timestamp = time.mktime(time.strptime(heartbeat_time, db_conn.TIME_FORMAT))

        now_time = cls._get_time_now()
        now_timestamp = time.mktime(time.strptime(now_time, db_conn.TIME_FORMAT))

        seconds_passed_by = now_timestamp - heartbeat_timestamp
        return seconds_passed_by

    @classmethod
    @tracer.trace_func
    def is_node_online(cls, node_uuid):
        """
        判断指定节点是否在线
        """
        # 节点的心跳时间超过60s则认为已离线
        if abs(cls.get_heartbeat_seconds_passed_by(node_uuid)) >= 60:
            return False
        return True

    @classmethod
    @tracer.trace_func
    def update_heartbeat_time(cls, node_uuid):
        """
        更新指定节点的心跳时间
        """
        cls.update_heartbeat_time_ex(node_uuid, cls._get_time_now())

    @classmethod
    @tracer.trace_func
    def update_heartbeat_time_ex(cls, node_uuid, heartbeat_time):
        """
        更新指定节点的心跳时间
        @param string node_uuid         指定节点唯一标识
        @param string heartbeat_time    格式为 '2015-09-02 08:54:36'
        """
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """UPDATE `node` SET `f_heartbeat_time` = %s WHERE `f_node_uuid` = %s;"""
        conn.update(sql, heartbeat_time, node_uuid)

# ============================================================================
# 内部函数
# ============================================================================
    @classmethod
    @tracer.trace_func
    def _record_to_node_info(cls, record):
        """
        单条查询结果转换成TNodeInfo对象
        @param record 单条查询结果
        @return TNodeInfo
        """
        node_info = deepcopy(ncTNodeInfo)
        node_info['node_uuid'] = record['f_node_uuid']
        node_info['role_db'] = record['f_role_db']
        node_info['role_ecms'] = record['f_role_ecms']
        node_info['role_app'] = record['f_role_app']
        node_info['role_storage'] = record['f_role_storage']
        node_info['node_alias'] = record['f_node_alias']
        node_info['node_ip'] = record['f_node_ip']
        node_info['is_online'] = cls.is_node_online(record['f_node_uuid'])
        node_info['is_ha'] = record['f_ha']
        node_info['is_etcd']= record['f_etcd']
        node_info['consistency_status'] = record['f_consistency_status']
        return node_info

    @classmethod
    @tracer.trace_func
    def _local_record_to_node_info(cls, record):
        """
        本地单条查询结果转换成TNodeInfo对象
        @param record 单条查询结果
        @return TNodeInfo
        """
        node_info = deepcopy(ncTNodeInfo)
        node_info['node_uuid'] = record['f_node_uuid']
        node_info['role_db'] = record['f_role_db']
        node_info['role_ecms'] = record['f_role_ecms']
        node_info['role_app'] = record['f_role_app']
        node_info['role_storage'] = record['f_role_storage']
        node_info['node_alias'] = record['f_node_alias']
        node_info['node_ip'] = record['f_node_ip']
        node_info['is_online'] = cls.is_node_online(record['f_node_uuid'])
        node_info['is_ha'] = record['f_ha']
        node_info['is_etcd'] = record['f_etcd']
        node_info['consistency_status'] = record['f_consistency_status']
        node_info['is_online'] = True
        return node_info

    @classmethod
    @tracer.trace_func
    def _record_to_ha_node_info(cls, record):
        """
        单条查询结果转换成ncTHaNodeInfo对象
        @param record 单条查询结果
        @return ncTHaNodeInfo
        """
        ha_node_info = deepcopy(ncTHaNodeInfo)
        ha_node_info['node_uuid'] = record['f_node_uuid']
        ha_node_info['node_alias'] = record['f_node_alias']
        ha_node_info['node_ip'] = record['f_node_ip']
        ha_node_info['is_online'] = cls.is_node_online(record['f_node_uuid'])
        ha_node_info['sys'] = record['f_ha']
        return ha_node_info

    @classmethod
    @tracer.trace_func
    def _get_time_now(cls):
        """
        使用 sql 命令获取当前数据库节点时间
        @return string time     格式为 '2015-09-02 08:54:36'
        """
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()

        sql = """
        SELECT NOW()
        """
        result = conn.fetch_one_result(sql)
        current_time = result["NOW()"]
        now = current_time.strftime(db_conn.TIME_FORMAT)
        return now

    @classmethod
    @tracer.trace_func
    def exchange_ecms_role(cls, old_ecms_uuid, new_ecms_uuid):
        """交换ecms角色

        :param old_ecms_uuid: 当前ecms节点uuid
        :type old_ecms_uuid: str
        :param new_ecms_uuid: 新的ecms节点uuid
        :type new_ecms_uuid: str
        """
        syslog_info(
            MODULE_NAME,
            "Change ecms_role from node[%s] to node[%s} begin" % (old_ecms_uuid, new_ecms_uuid)
        )

        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        cursor = conn.conn.cursor()
        # 开始事务
        cursor.execute("set autocommit=0")
        cursor.execute("begin")
        sql = "UPDATE `node` SET `f_role_ecms` = 0 WHERE `f_node_uuid` = '%s'" % old_ecms_uuid
        cursor.execute(sql)
        syslog_info(MODULE_NAME, "execute sql[%s]" % sql)

        sql = "UPDATE `node` SET `f_role_ecms` = 1 WHERE `f_node_uuid` = '%s'" % new_ecms_uuid
        cursor.execute(sql)
        syslog_info(MODULE_NAME, "execute sql[%s]" % sql)

        cursor.execute("commit")
        cursor.execute("set autocommit=1")

        cursor.close()
        conn.conn.close()

        syslog_info(
            MODULE_NAME,
            "Change ecms_role from node[%s] to node[%s} end" % (old_ecms_uuid, new_ecms_uuid)
        )

    @classmethod
    @tracer.trace_func
    def exchange_db_role(cls, old_db1_uuid, new_db1_uuid):
        """交换db角色

        :param old_db1_uuid(str): 当前数据库主uuid
        :param new_db1_uuid(str): 新的数据库主uuid
        """
        syslog_info(
            MODULE_NAME,
            "Change db_master_role from node[%s] to node[%s} begin" % (old_db1_uuid, new_db1_uuid)
        )

        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        cursor = conn.conn.cursor()
        # 开始事务
        cursor.execute("set autocommit=0")
        cursor.execute("begin")
        sql = "UPDATE `node` SET `f_role_db` = 2 WHERE `f_node_uuid` = '%s'" % old_db1_uuid
        cursor.execute(sql)
        syslog_info(MODULE_NAME, "execute sql[%s]" % sql)

        sql = "UPDATE `node` SET `f_role_db` = 1 WHERE `f_node_uuid` = '%s'" % new_db1_uuid
        cursor.execute(sql)
        syslog_info(MODULE_NAME, "execute sql[%s]" % sql)

        cursor.execute("commit")
        cursor.execute("set autocommit=1")

        cursor.close()
        conn.conn.close()

        syslog_info(
            MODULE_NAME,
            "Change db_master_role from node[%s] to node[%s} end" % (old_db1_uuid, new_db1_uuid)
        )
