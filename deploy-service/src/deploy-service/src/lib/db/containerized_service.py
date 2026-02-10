#!/usr/bin/env python3
# -*- coding:utf-8 -*-
"""
数据库表t_deployment_option操作类
"""

from src.lib.db.db_connector import get_db_operate_obj

TABLE_NAME = "containerized_service"


class ContainerizedService(object):
    def __init__(self):
        pass

    @classmethod
    def insert_service(cls, service_info):
        db_operator = get_db_operate_obj()

        col = dict()
        col["service_name"] = service_info["service_name"]
        col["replicas"] = service_info["replicas"]
        col["installed_version"] = service_info["installed_version"]
        col["installed_package"] = service_info["installed_package"]
        col["available_version"] = service_info["available_version"]
        col["available_package"] = service_info["available_package"]

        db_operator.insert(TABLE_NAME, col)

    @classmethod
    def update_installed_version(cls, service_name, version):
        db_operator = get_db_operate_obj()

        sql = "update containerized_service set installed_version=%s " "where service_name=%s"

        db_operator.update(sql, version, service_name)

    @classmethod
    def update_installed_package(cls, service_name, package):
        db_operator = get_db_operate_obj()

        sql = "update containerized_service set installed_package=%s " "where service_name=%s"

        db_operator.update(sql, package, service_name)

    @classmethod
    def update_installed_package_by_available(cls, service_name):
        db_operator = get_db_operate_obj()

        sql = "select available_package from containerized_service where service_name=%s"

        available_package = db_operator.fetch_one_result(sql, service_name)["available_package"]
        cls.update_installed_package(service_name, available_package)

    @classmethod
    def update_node_role(cls, service_name, node_role):
        db_operator = get_db_operate_obj()

        sql = "update containerized_service set node_role=%s " "where service_name=%s"

        db_operator.update(sql, node_role, service_name)

    @classmethod
    def update_available_version(cls, service_name, version):
        db_operator = get_db_operate_obj()

        sql = "update containerized_service set available_version=%s " "where service_name=%s"

        db_operator.update(sql, version, service_name)

    @classmethod
    def update_available_package(cls, service_name, package_name):
        db_operator = get_db_operate_obj()

        sql = "update containerized_service set available_package=%s " "where service_name=%s"

        db_operator.update(sql, package_name, service_name)

    @classmethod
    def update_replicas(cls, service_name, replicas):
        db_operator = get_db_operate_obj()

        sql = "update containerized_service set replicas=%s " "where service_name=%s"

        db_operator.update(sql, replicas, service_name)

    @classmethod
    def update_service_replicas(cls, service_name, replicas):
        db_operator = get_db_operate_obj()

        sql = "update containerized_service set replicas=%s " "where service_name=%s"

        db_operator.update(sql, replicas, service_name)

    @classmethod
    def delete_service(cls, service_name):
        db_operator = get_db_operate_obj()

        sql = "delete from containerized_service where service_name=%s"

        db_operator.delete(sql, service_name)

    @classmethod
    def get_service_info(cls, service_name):
        db_operator = get_db_operate_obj()

        sql = (
            "select service_name,replicas,installed_version,installed_package,available_version,available_package,"
            "require_third_app_depservice from containerized_service where service_name = %s"
        )

        service_info = db_operator.fetch_one_result(sql, service_name)

        return service_info


    @classmethod
    def get_requir_third_app_depservice(cls, service_name):
        db_operator = get_db_operate_obj()

        sql = "select require_third_app_depservice from containerized_service where service_name=%s"

        require = db_operator.fetch_one_result(sql, service_name)
        return require

    @classmethod
    def update_require_third_app_depservice(cls, service_name, require):
        db_operator = get_db_operate_obj()

        sql = "update containerized_service set require_third_app_depservice=%s " "where service_name=%s"

        db_operator.update(sql, require, service_name)

    @classmethod
    def get_all_module_services_info(self):
        db_operator = get_db_operate_obj()
        sql = (
            "select service_name,replicas,installed_version,installed_package,available_version,available_package,"
            "require_third_app_depservice from containerized_service"
        )

        result = db_operator.fetch_all_result(sql)
        return result

    @classmethod
    def update_optional_install_micro_service(cls, service_name: str, optional_install_micro_service: bool):
        """
        更新模块服务的 optional_install_micro_service(是否可选安装微服务) 字段，
        """
        db_operator = get_db_operate_obj()
        sql = "update containerized_service set optional_install_micro_service=%s where service_name=%s"
        db_operator.update(sql, optional_install_micro_service, service_name)

    @classmethod
    def get_optional_install_micro_service(cls, service_name: str):
        """
        获取模块服务的 optional_install_micro_service(是否可选安装微服务) 字段，
        """
        db_operator = get_db_operate_obj()
        sql = "select optional_install_micro_service from containerized_service where service_name=%s"
        result = db_operator.fetch_one_result(sql, service_name)
        if result:
            return result["optional_install_micro_service"]


