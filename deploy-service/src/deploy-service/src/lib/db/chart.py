#!/usr/bin/env python3
# -*- coding:utf-8 -*-

from pymysql.converters import escape_string

from src.lib.db.db_connector import get_db_operate_obj

TABLE_NAME = "chart"


class Chart(object):
    def __init__(self):
        pass

    @classmethod
    def push_chart(cls, chart_info):
        db_operator = get_db_operate_obj()

        col = dict()
        col["chart_name"] = escape_string(chart_info["chart_name"])
        col["service_name"] = escape_string(chart_info["service_name"])
        col["chart_version"] = escape_string(chart_info["chart_version"])

        db_operator.insert(TABLE_NAME, col)

    @classmethod
    def delete_chart(cls, chart_name, service_name):
        db_operator = get_db_operate_obj()

        sql = "delete from chart where chart_name=%s and service_name=%s"

        db_operator.delete(sql, chart_name, service_name)

    @classmethod
    def delete_service(cls, service_name):
        db_operator = get_db_operate_obj()

        sql = "delete from chart where service_name=%s"

        db_operator.delete(sql, service_name)

    @classmethod
    def update_chart_version(cls, chart_name, chart_version, service_name):
        db_operator = get_db_operate_obj()

        sql = "update chart set chart_version=%s where chart_name=%s and service_name=%s"

        db_operator.update(sql, chart_version, chart_name, service_name)

    @classmethod
    def update_chart_service_name(cls, chart_name, service_name):
        db_operator = get_db_operate_obj()

        sql = "update chart set service_name=%s where chart_name=%s"

        db_operator.update(sql, service_name, chart_name)

    @classmethod
    def get_chart_info(cls, chart_name, service_name):
        db_operator = get_db_operate_obj()

        sql = "select * from chart where chart_name=%s and service_name=%s"

        chart_info = db_operator.fetch_one_result(sql, chart_name, service_name)

        return chart_info

    @classmethod
    def get_charts_by_service(cls, service_name):
        db_operator = get_db_operate_obj()

        sql = "select chart_name from chart where service_name=%s"

        results = db_operator.fetch_all_result(sql, service_name)

        charts = list()
        for chart in results:
            charts.append(chart["chart_name"])

        return charts

    @classmethod
    def get_charts_info_by_service(cls, service_name):
        db_operator = get_db_operate_obj()

        sql = "select chart_name,chart_version from chart where service_name=%s"

        results = db_operator.fetch_all_result(sql, service_name)

        return results
