#!/usr/lib/env python
# -*- coding:utf-8 -*-
"""Deploy数据库连接类"""

# ---------------------------------------------------------------------------
# src.lib.db.db_connector.py:
#    Copyright (c) Eisoo Software, Inc.(2004 - 2013), All rights reserved.
#
# Author:
#    ruan.yulin (ruan.yulin@eisoo.com)
#
# Creating Time:
#    2019.08.02
# ---------------------------------------------------------------------------

from src.clients.config import ConfigClient
from src.lib.db import dbconn
from src.common.log_util import logger

_rds_info: dict = None

def get_rds_info():
    global _rds_info
    if _rds_info is None:
        refresh_rds_info()
    return _rds_info


def refresh_rds_info():
    global _rds_info
    _rds_info = ConfigClient.load_config(renew=True).rds_info()
    return _rds_info


_connector = None

def get_db_operate_obj():
    """连接数据库"""
    global _connector
    if _connector is None:
        _connector = dbconn.Connector(get_rds_info())
    try:
        return _connector.get_db_operate_obj()
    except Exception as e:
        logger.error(e)
        _connector = dbconn.Connector(refresh_rds_info())
        return _connector.get_db_operate_obj()
