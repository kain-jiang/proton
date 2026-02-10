#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# @Time    : 2021/3/9 16:52
# @Author  : Jimmy.li
# @Email   : jimmy.li@aishu.cn

from src.lib.db.db_connector import get_db_operate_obj

TABLE_NAME = "micro_third_app_depservice"


class MicroThirdAppDerService(object):
    def __init__(self):
        pass

    @classmethod
    def get_thirdsvc_by_svcname_and_micname_and_components(
        cls, service_name, micro_service_name, third_app_service, components_name
    ):
        db_operator = get_db_operate_obj()

        sql = "select id,service_name,micro_service,third_app_service,components_name,`enable` from micro_third_app_depservice where service_name=%s and micro_service = %s and components_name = %s and third_app_service = %s"

        result = db_operator.fetch_one_result(sql, service_name, micro_service_name, components_name, third_app_service)

        return result

    @classmethod
    def insert_micro_third_dpservice(cls, micro_third_info):
        db_operator = get_db_operate_obj()

        col = dict()
        col["service_name"] = micro_third_info["service_name"]
        col["micro_service"] = micro_third_info["micro_service"]
        col["third_app_service"] = micro_third_info["third_app_service"]
        col["components_name"] = micro_third_info["components_name"]
        col["enable"] = micro_third_info["enable"]

        db_operator.insert(TABLE_NAME, col)

    @classmethod
    def get_third_app_by_service(cls, service_name):
        db_operator = get_db_operate_obj()

        sql = "select id,service_name,micro_service,third_app_service,components_name,`enable` from micro_third_app_depservice where service_name=%s"

        results = db_operator.fetch_all_result(sql, service_name)
        return results

    @classmethod
    def update_third_app_service_enable(cls, containerized_service, third_app_service, enable):
        db_operator = get_db_operate_obj()

        sql = "update micro_third_app_depservice set `enable`=%s where service_name=%s and third_app_service=%s"

        db_operator.update(sql, enable, containerized_service, third_app_service)

    @classmethod
    def delete_third_info_by_service_name(cls, service_name):
        """删除模块化服务第三方依赖信息，删除数据库中数据，而非将enable置为0"""
        db_operator = get_db_operate_obj()
        sql = "delete from micro_third_app_depservice where service_name=%s"

        db_operator.delete(sql, service_name)

    ####################################

    @classmethod
    def get_enable_third_app_depservice_by_micro_service(cls, containerized_service, micro_service):
        db_operator = get_db_operate_obj()

        sql = "select third_app_service ,components_name from micro_third_app_depservice where service_name=%s and micro_service=%s and  enable=%s"

        results = db_operator.fetch_all_result(sql, containerized_service, micro_service, True)

        return results

    @classmethod
    def get_third_app_depservice_by_micro_service(cls, containerized_service, micro_service):
        db_operator = get_db_operate_obj()

        sql = "select third_app_service ,components_name ,`enable` from micro_third_app_depservice where service_name=%s and micro_service=%s"

        results = db_operator.fetch_all_result(sql, containerized_service, micro_service)

        return results

    @classmethod
    def get_enable_third_app_depservice_by_service(cls, containerized_service):
        db_operator = get_db_operate_obj()

        sql = "select third_app_service, components_name from micro_third_app_depservice where service_name=%s and  `enable`=%s"

        results = db_operator.fetch_all_result(sql, containerized_service, True)

        return results

    @classmethod
    def get_enabled_micro_services_by_service(cls, containerized_service):
        """
        通过模块服务获取已启用第三方依赖的微服务
        """
        db_operator = get_db_operate_obj()

        sql = "select micro_service from micro_third_app_depservice where service_name=%s and  `enable`=%s"

        results = db_operator.fetch_all_result(sql, containerized_service, True)

        micro_services = [s["micro_service"] for s in results]

        return micro_services

    @classmethod
    def get_micro_services_by_service(cls, containerized_service):
        """
        通过模块服务获取依赖第三方服务的微服务
        """
        db_operator = get_db_operate_obj()

        sql = "select micro_service from micro_third_app_depservice where service_name=%s"

        results = db_operator.fetch_all_result(sql, containerized_service)

        micro_services = []
        if results:
            micro_services = [r["micro_service"] for r in results]

        return micro_services

    @classmethod
    def disable_third_app_service(cls, service_name):
        db_operator = get_db_operate_obj()

        sql = "update micro_third_app_depservice set `enable`=0 where service_name=%s"

        db_operator.update(sql, service_name)
