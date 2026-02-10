#!/usr/bin/env python
# -*- coding:utf-8 -*-

"""ecms.swift 数据表管理模块"""

import time
from src.modules.pydeps import logger, calclib, tracer
from src.modules.ecmsdb import err, db_conn


MODULE_NAME = err.ECMSDB_NAME


class TSwiftDBManager(object):
    """
    swift 数据表管理模块
    """
    @classmethod
    def get_object_builder(cls):
        """
        获取 object.builder 文件内容
        若不存在该记录，则返回 None
        """
        tracer.trace("begin")
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_object.builder` FROM `swift` ORDER BY `f_id` DESC LIMIT 1;"""
        result = conn.fetch_one_result(sql)
        if result is not None:
            tracer.trace("return: not None")
            return result["f_object.builder"]
        else:
            tracer.trace("return: None")
            return None

    @classmethod
    def get_object_ring(cls):
        """
        获取 object.ring.gz 文件内容
        若不存在该记录，则返回 None
        """
        tracer.trace("begin")
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_object.ring.gz` FROM `swift` ORDER BY `f_id` DESC LIMIT 1;"""
        result = conn.fetch_one_result(sql)
        if result is not None:
            tracer.trace("return: not None")
            return result["f_object.ring.gz"]
        else:
            tracer.trace("return: None")
            return None

    @classmethod
    @tracer.trace_func
    def get_object_ring_md5(cls):
        """
        获取 object.ring.gz 的 md5 值
        若不存在该记录，则返回 None
        """
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """SELECT `f_object.ring.gz_md5` FROM `swift` ORDER BY `f_id` DESC LIMIT 1;"""
        result = conn.fetch_one_result(sql)
        if result is not None:
            return result["f_object.ring.gz_md5"]
        else:
            return None

    @classmethod
    def update(cls, object_builder, object_ring):
        """
        更新 swift 配置
        """
        object_ring_md5 = calclib.calc_md5(object_ring)

        tracer.trace("(object_ring_md5=%s) begin" % object_ring_md5)
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()

        # 先查询出第一条记录
        sql = """SELECT `f_id` FROM `swift` ORDER BY `f_id` DESC LIMIT 1;"""
        old_record = conn.fetch_one_result(sql)

        # 插入新记录
        values = {"f_update_time": time.strftime(db_conn.TIME_FORMAT, time.localtime()),
                  "f_object.builder": object_builder,
                  "f_object.ring.gz": object_ring,
                  "f_object.ring.gz_md5": object_ring_md5}
        conn.insert("swift", values)

        # 再删除旧记录
        if old_record is not None:
            sql = """DELETE FROM `swift` WHERE `f_id` = %s;""" % old_record["f_id"]
            conn.delete(sql)
        tracer.trace("(object_ring_md5=%s) end" % object_ring_md5)
        logger.syslog(
            MODULE_NAME, "Updated swift object ring(md5={0}).".format(object_ring_md5))

    @classmethod
    @tracer.trace_func
    def delete(cls):
        """
        删除 swift 配置
        """
        connector = db_conn.get_db_connector()
        conn = connector.get_db_operate_obj()
        sql = """DELETE FROM `swift`;"""
        conn.delete(sql)
        logger.syslog(MODULE_NAME, "Deleted swift ring config.")
