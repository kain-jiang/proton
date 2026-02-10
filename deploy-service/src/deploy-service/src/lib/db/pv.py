#!/usr/bin/env python3
# -*- coding:utf-8 -*-
"""
数据库表pv操作类
"""
import rdsdriver
from pymysql.converters import escape_string

from src.lib.db.db_connector import get_db_operate_obj

TABLE_NAME = "pv"


class PvManager(object):
    def __init__(self):
        pass

    @classmethod
    def create_pv(cls, pv_name, release_name):
        db_oprator = get_db_operate_obj()

        columns = dict()
        columns["pv_name"] = escape_string(pv_name)
        columns["release_name"] = escape_string(release_name)

        db_oprator.insert(TABLE_NAME, columns)

    @classmethod
    def get_pv(cls, pv_name):
        db_oprator = get_db_operate_obj()

        sqlstr = "select * from pv where pv_name=%s"

        pv_name = escape_string(pv_name)
        return db_oprator.fetch_one_result(sqlstr, pv_name)

    @classmethod
    def get_pv_by_service_name(cls, service_name):
        db_oprator = get_db_operate_obj()

        sqlstr = "select * from pv where release_name=%s"

        service_name = escape_string(service_name)
        return db_oprator.fetch_one_result(sqlstr, service_name)

    @classmethod
    def delete_pv(cls, pv_name):
        db_oprator = get_db_operate_obj()

        sqlstr = "delete from pv where pv_name=%s"

        pv_name = escape_string(pv_name)
        db_oprator.delete(sqlstr, pv_name)
