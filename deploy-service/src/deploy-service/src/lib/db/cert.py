#!/usr/bin/env python
# -*- coding: utf-8 -*-
# @Time    : 2021/4/13 17:04
# @Author  : Jimmy.li
# @Email   : jimmy.li@aishu.cn

from src.lib.db.db_connector import get_db_operate_obj

TABLE_NAME = "cert"


class Cert(object):
    def __init__(self):
        pass

    @classmethod
    def get_https_all_content(cls):
        conn = get_db_operate_obj()
        query_sql = """ select * from cert"""
        result = conn.fetch_all_result(query_sql)
        return result

    @classmethod
    def update_https_ca_content(self, cert_key, cert_value):
        conn = get_db_operate_obj()
        query_sql = """ select * from cert where f_key=%s"""
        result = conn.fetch_one_result(query_sql, cert_key)
        if result:
            update_query_sql = """update cert set f_value=%s where f_key=%s"""
            conn.update(update_query_sql, cert_value, cert_key)
        else:
            insert_obj = {"f_key": cert_key, "f_value": cert_value}
            conn.insert("cert", insert_obj)

    @classmethod
    def get_value_content_by_key(cls, key):
        conn = get_db_operate_obj()
        query_sql = """ select f_value from cert where f_key = %s"""
        result = conn.fetch_one_result(query_sql, key)
        return result
