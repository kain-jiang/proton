#!/usr/lib/env python
# -*- coding:utf-8 -*-

import rdsdriver
from pymysql.converters import escape_string

from src.lib.db.db_connector import get_db_operate_obj

TABLE_NAME = "micro_service"


class MicroService(object):
    def __init__(self):
        pass

    @classmethod
    def get_all_service(cls):
        db_operator = get_db_operate_obj()

        sql = "select * from micro_service"

        results = db_operator.fetch_all_result(sql)

        return results

    @classmethod
    def get_ingress_service(cls):
        db_operator = get_db_operate_obj()

        sql = "select * from micro_service where need_ingress=1"

        results = db_operator.fetch_all_result(sql)

        return results

    @classmethod
    def update_service_version(cls, service_name, version):
        db_operator = get_db_operate_obj()

        sql = "update micro_service set micro_service_version=%s " "where micro_service_name=%s"

        db_operator.update(sql, version, service_name)

    @classmethod
    def insert_service(cls, service_info):
        db_operator = get_db_operate_obj()

        colu = dict()

        colu["micro_service_name"] = escape_string(service_info["micro_service_name"])
        colu["service_name"] = escape_string(service_info["service_name"])
        colu["micro_service_version"] = escape_string(service_info["micro_service_version"])
        colu["external_port"] = service_info["external_port"]
        colu["internal_port"] = service_info["internal_port"]
        colu["need_ingress"] = service_info["need_ingress"]

        db_operator.insert(TABLE_NAME, colu)

    @classmethod
    def get_service_by_micro_name(cls, micro_service_name):
        db_operator = get_db_operate_obj()

        sql = "select * from micro_service where micro_service_name=%s"

        result = db_operator.fetch_one_result(sql, micro_service_name)

        return result

    @classmethod
    def get_services_by_service_name(cls, service_name):
        db_operator = get_db_operate_obj()

        sql = "select micro_service_name from micro_service where service_name=%s"

        results = db_operator.fetch_all_result(sql, service_name)

        micro_services = list()
        for result in results:
            micro_services.append(result["micro_service_name"])

        return micro_services

    @classmethod
    def delete_micro_service(cls, micro_service_name):
        db_operator = get_db_operate_obj()

        sql = "delete from micro_service where micro_service_name=%s"

        db_operator.delete(sql, micro_service_name)

    @classmethod
    def delete_micro_services(cls, micro_services):
        db_operator = get_db_operate_obj()

        sql = f"delete from micro_service where micro_service_name in ({','.join(len(micro_services) * ['%s'])})"

        db_operator.delete(sql, *micro_services)

    @classmethod
    def delete_service(cls, service_name):
        db_operator = get_db_operate_obj()

        sql = "delete from micro_service where service_name=%s"

        db_operator.delete(sql, service_name)

    @classmethod
    def insert_many_service(cls, service_infos):
        db_operator = get_db_operate_obj()

        col = [
            "micro_service_name",
            "service_name",
            "micro_service_version",
            "external_port",
            "internal_port",
            "need_ingress",
        ]

        values = list()
        for info in service_infos:
            value = (
                info["micro_service_name"],
                info["service_name"],
                info["micro_service_version"],
                info["external_port"],
                info["internal_port"],
                info["need_ingress"],
            )
            values.append(value)

        db_operator.insert_many(TABLE_NAME, col, values)

    @classmethod
    def get_services_info_by_service_name(cls, service_name):
        db_operator = get_db_operate_obj()

        sql = "select * from micro_service where service_name=%s"

        results = db_operator.fetch_all_result(sql, service_name)

        micro_services = list()
        for result in results:
            micro_services.append(result)

        return micro_services
