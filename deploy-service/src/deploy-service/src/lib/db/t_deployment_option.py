#!/usr/bin/env python3
# -*- coding:utf-8 -*-
"""
数据库表deployment_option操作类
"""
import rdsdriver
from pymysql.converters import escape_string

from src.lib.db.db_connector import get_db_operate_obj

# global keys

TABLE_NAME = "deployment_option"


class DeploymentOption(object):
    def __init__(self):
        pass

    @classmethod
    def insert_option(cls, key, value):
        db_oprator = get_db_operate_obj()

        columns = dict()
        columns["option_key"] = escape_string(key)
        columns["option_value"] = escape_string(value)

        db_oprator.insert(TABLE_NAME, columns)

    @classmethod
    def insert_option_not_escape_string(cls, key, value):
        db_oprator = get_db_operate_obj()

        columns = dict()
        columns["option_key"] = key
        columns["option_value"] = value

        db_oprator.insert(TABLE_NAME, columns)

    @classmethod
    def get_option(cls, key):
        db_oprator = get_db_operate_obj()

        sqlstr = "select option_value from deployment_option where option_key = %s"

        key = escape_string(key)
        result = db_oprator.fetch_one_result(sqlstr, key)

        return result["option_value"] if result else 0

    @classmethod
    def delete_option(cls, key):
        db_oprator = get_db_operate_obj()

        sqlstr = "delete from deployment_option where option_key = %s"

        key = escape_string(key)
        db_oprator.delete(sqlstr, key)

    @classmethod
    def update_option(cls, key, value):
        db_oprator = get_db_operate_obj()

        sqlstr = "update deployment_option set option_value=%s where option_key=%s"

        db_oprator.update(sqlstr, value, key)
