#!/usr/bin/env python
# -*- coding: utf-8 -*-
# @Time    : 2021/4/11 23:56
# @Author  : Jimmy.li
# @Email   : jimmy.li@aishu.cn

"""
客户端
"""

from src.lib.db.db_connector import get_db_operate_obj


class ClientPackage(object):
    def __init__(self):
        pass

    @classmethod
    def get_all_client_package_info(cls):
        """
        获取所有升级包信息
        :return:所有升级包信息 dict
        """
        conn = get_db_operate_obj()
        sqlstr = "select * from client_package_info"
        result = conn.fetch_all_result(sqlstr)
        return result

    @classmethod
    def check_ostype_exist(cls, os_type, update_type):
        """
        检查给定的osType是否已存在升级包
        :return 升级包信息
        """
        conn = get_db_operate_obj()
        query_sql = """
        SELECT f_os
        FROM client_package_info
        WHERE f_os = %s and f_update_type = %s
        """
        result = conn.fetch_all_result(query_sql, os_type, update_type)
        return result

    @classmethod
    def get_client_package_by_ostype(cls, os_type):
        """
        根据升级包类型获取升级包信息
        :param os_type:  安装包类型 'android', 'mac', 'win32_advanced'。。。
        :return: 升级包信息
        """
        conn = get_db_operate_obj()
        query_sql = """
                        SELECT f_os, f_url, f_pkg_location, f_name, f_size, f_version, f_mode, 
                        f_time, f_open_download, f_update_type, f_version_description
                        FROM client_package_info
                        WHERE f_os = %s
                        """
        result = conn.fetch_all_result(query_sql, os_type)
        return result

    @classmethod
    def delete_client_package_by_ostype(cls, os_type, update_type):
        """
        删除升级包
        :param os_type: 安装包类型 'android', 'mac', 'win32_advanced'。。。
        :param update_type: 上传类型 标椎(standard)/自定义(custom)
        :return: None
        """
        conn = get_db_operate_obj()
        delete_sql = """
                        DELETE
                        FROM client_package_info
                        WHERE f_os = %s and f_update_type = %s
                        """
        conn.delete(delete_sql, os_type, update_type)


    @classmethod
    def insert_client_update_package(cls, obj):
        """
        升级包信息插入数据库
        :param obj: 插入的数据 dict
        :return: None
        """
        conn = get_db_operate_obj()
        conn.insert("client_package_info", obj)

    @classmethod
    def set_client_package_config(cls, os_type, update_type, version_description=None, open_download=None):
        """
        设置升级包是否开放下载，是否强制升级
        :param os_type: 安装包类型 'android', 'mac', 'win32_advanced'。。。
        :param update_type: 上传类型 标椎(standard)/自定义(custom)
        :param version_description: 客户端升级包相关描述
        :param open_download: 是否开放下载 bool
        :return: None
        """
        conn = get_db_operate_obj()
        
        if version_description != None:
            query_sql = """
                update client_package_info set f_version_description=%s where f_os=%s and f_update_type=%s
                """
            conn.update(query_sql, version_description, os_type, update_type)

        if isinstance(open_download, bool):
            query_sql = """
                            update client_package_info set f_open_download=%s where f_os=%s and f_update_type=%s
                            """
            conn.update(query_sql, open_download, os_type, update_type)

    @classmethod
    def get_client_package_by_ostype_updatetype(cls, os_type, update_type):
        """
        根据上传类型获取升级包信息
        :param os_type: 安装包类型 'android', 'mac', 'win32_advanced'。。。
        :param update_type: 上传类型 标椎(standard)/自定义(custom)
        :return:
        """
        conn = get_db_operate_obj()
        query_sql = """
                        SELECT f_os, f_url, f_pkg_location, f_name, f_size, f_version, f_mode, 
                        f_time, f_update_type, f_open_download, f_oss_id
                        FROM client_package_info
                        WHERE f_os = %s and f_update_type = %s
                        """
        result = conn.fetch_one_result(query_sql, os_type, update_type)

        return result

    @classmethod
    def set_client_package_update_type(cls, os_type, update_type):
        """
        设置升级包使用类型
        :param ostype:安装包类型 'android', 'mac', 'win32_advanced'。。。
        :param update_type:上传类型 标椎(standard)/自定义(custom)
        :return:None
        """
        conn = get_db_operate_obj()
        query_sql = """
                    update os_config set f_mode=%s where f_os=%s
                    """
        conn.update(query_sql, update_type, os_type)

    @classmethod
    def get_current_used_by_ostype(cls, os_type, set_custom_in_all_platform=False):
        """
        获取当前使用的升级包信息
        :param os_type: 安装包类型 'android', 'mac', 'win32_advanced'。。。
        :param set_custom_in_all_platform: 强制所有平台使用自定义下载选项，此项默认为False'
        :return: 升级包信息
        """
        conn = get_db_operate_obj()
        config = dict()
        config_query_sql = """select f_mode from os_config where f_os=%s"""
        if set_custom_in_all_platform:
            config["f_mode"] = "custom"
        else:
            if int(os_type) in [2, 7]:
                config = conn.fetch_one_result(config_query_sql, os_type)
            else:
                config["f_mode"] = "standard"

        query_sql = """
                        SELECT f_os, f_url, f_pkg_location, f_name, f_size, f_version, f_mode,
                         f_time, f_update_type, f_open_download, f_oss_id
                        FROM client_package_info
                        WHERE f_os = %s and f_update_type = %s
                        """
        result = conn.fetch_one_result(query_sql, os_type, config["f_mode"])

        return result

    @classmethod
    def get_all_current_used_package_info(cls):
        """
        获取所有客户端升级包当前使用的信息
        :return: 所有升级包类型，是否开放下载，上传类型 dict
        """
        conn = get_db_operate_obj()
        sqlstr = "select f_os, f_open_download, f_update_type from client_package_info"
        result = conn.fetch_all_result(sqlstr)
        return result

    @classmethod
    def get_client_package_update_type(cls, os_type=None):
        """
        获取升级包使用的升级类型
        :param os_type:安装包类型 'android', 'mac', 'win32_advanced'。。。
        :param update_type:上传类型 标椎(standard)/自定义(custom)
        :return:
        """
        conn = get_db_operate_obj()
        if os_type:
            query_sql = """
                        select f_os, f_mode from  os_config where f_os = %s
                        """
            result = conn.fetch_all_result(query_sql, os_type)
        else:
            query_sql = """
                            select f_os, f_mode from  os_config
                            """
            result = conn.fetch_all_result(query_sql)
        return result
