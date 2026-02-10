#!/usr/bin/env python
# -*- coding:utf-8 -*-

"""
数据库连接公共方法
"""

from src.modules.pydeps import clusterconf, dbconn

import dbscripts.global_info
import dbscripts.ecms

TIME_FORMAT = "%Y-%m-%d %H:%M:%S"


def get_db_connector(db_name=dbscripts.ecms.DB_NAME):
    """获取 ecms DB 连接器"""
    db_info = dbconn.DBInfo()

    # 第三方数据库
    if clusterconf.ClusterConfig.if_use_external_db():
        info = clusterconf.ClusterConfig.get_external_db_info()
        db_info.host = info['db_host']
        db_info.port = info['db_port']
        db_info.user = info['db_user']
        db_info.passwd = info['db_password']
    else:
        # 内置数据库
        db_info.host = clusterconf.ClusterConfig.get_db_host()
        db_info.port = clusterconf.ClusterConfig.get_db_port()
        db_info.user = dbscripts.global_info.DB_REMOTE_USER
        db_info.passwd = dbscripts.global_info.DB_REMOTE_PASSWD

    db_info.dbname = db_name
    db_info.read_timeout = 60
    db_info.write_timeout = 60
    return dbconn.Connector(db_info)

def get_local_db_connector(port):
    """获取数据库的本地连接器"""
    db_info = dbconn.DBInfo()
    db_info.host = "127.0.0.1"
    db_info.user = dbscripts.global_info.DB_LOCAL_USER
    db_info.passwd = dbscripts.global_info.DB_LOCAL_PASSWD
    db_info.port = port
    db_info.dbname = dbscripts.ecms.DB_NAME
    return dbconn.Connector(db_info)
