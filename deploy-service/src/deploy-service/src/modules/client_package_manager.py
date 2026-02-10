#!/usr/bin/env python
# -*- coding: utf-8 -*-
# @Time : 2020/5/19 11:23
# @Author : Louis.she
# @Site :
# @File : PackageManager.py
# @Software: PyCharm
import os
import time
import uuid
from src.clients.opa import OPAClient
from src.common.log_util import logger

import requests

from src.clients.cms import CMSClient
from src.common import global_param, utils
from src.common.config import (
    COSTOM,
    FILE_SUFFIX,
    OBJECTID,
    OS_TYPE,
    OS_TYPE_ABBREVIATION_DICT,
    OS_TYPE_DEFINE,
    OS_TYPE_DICT,
    PKG_LOCATION_LOCAL_TO_OSS,
    PKG_LOCATION_STATIC_URL,
    STANDARD,
)
from src.clients.oss import OssGatewayManagerClient
from src.lib.db.client_package import ClientPackage
from urllib.parse import urlparse

class ClientPackageInfoManager(object):
    access_conf = None
    oss_key_prefix = "babc09f8-673b-4137-a497-b02eb279fbd5"
    # 是否已经检查过AS云端一体化CMS信息
    _as_is_saas_instance = None

    @classmethod
    def get_ossgateway_client(cls):
        return OssGatewayManagerClient.instance()

    @classmethod
    def get_access_info(cls):
        if cls.access_conf is None:
            cls.access_conf = utils.read_conf_in_config(
                global_param.SERVICE_ACCESS_CONFIG, global_param.SERVICE_ACCESS_FILE_NAME
            )
        return cls.access_conf

    @classmethod
    def check_anyshare_saas_instance(cls):
        """
        获取当前示例是不是AnyShare订阅服务实例
        :return: 是否为订阅实例 bool
        """
        if cls._as_is_saas_instance is None:
            cms_client_instance = CMSClient()
            if cms_client_instance.head_cms_data("as-saas-instance-info") is None:
                cls._as_is_saas_instance = False
            else:
                cls._as_is_saas_instance = True
        return cls._as_is_saas_instance

    @classmethod
    def get_package_info(cls, os_type=None):
        """
        获取安装包信息
        :param os_type: 安装包类型（可选）'android', 'mac', 'win32_advanced'。。。
        :return: 安装包信息 dict
        """
        package_info_list = list()
        # 仅获取该类型的安装包
        if os_type:
            if not cls.check_ostype_valid(os_type):
                return "", "400017009"
            single_packages = ClientPackage.get_client_package_by_ostype(OS_TYPE_DICT[os_type])
            if not single_packages:
                return "", "400017015"
            else:
                for single_package in single_packages:
                    package_info_dict = dict()
                    package_info_dict["name"] = (
                        "" if single_package["f_name"] in ["ios", "android"] else single_package["f_name"]
                    )
                    package_info_dict["url"] = (
                        single_package["f_url"]
                        if int(single_package["f_pkg_location"]) == PKG_LOCATION_STATIC_URL
                        else ""
                    )
                    package_info_dict["version"] = single_package["f_version"]
                    package_info_dict["mode"] = True if single_package["f_mode"] else False
                    package_info_dict["time"] = single_package["f_time"]
                    package_info_dict["os_type"] = OS_TYPE[int(single_package["f_os"])]
                    package_info_dict["size"] = single_package["f_size"]
                    package_info_dict["update_type"] = single_package["f_update_type"]
                    package_info_dict["open_download"] = True if single_package["f_open_download"] else False
                    package_info_dict["version_description"] = single_package["f_version_description"]
                    package_info_list.append(package_info_dict)
            return package_info_list, ""

        # 获取所有的安装包
        packages = ClientPackage.get_all_client_package_info()

        os_type_list = list()
        os_type_exist_ios = list()
        os_type_exist_android = list()
        for osc in packages:
            package_info_dict = dict()
            package_info_dict["name"] = "" if osc["f_name"] in ["ios", "android"] else osc["f_name"]
            package_info_dict["url"] = osc["f_url"] if int(osc["f_pkg_location"]) == PKG_LOCATION_STATIC_URL else ""
            package_info_dict["version"] = osc["f_version"]
            package_info_dict["mode"] = True if osc["f_mode"] else False
            package_info_dict["time"] = osc["f_time"]
            package_info_dict["os_type"] = OS_TYPE[int(osc["f_os"])]
            package_info_dict["size"] = osc["f_size"]
            package_info_dict["update_type"] = osc["f_update_type"]
            package_info_dict["open_download"] = True if osc["f_open_download"] else False
            package_info_dict["version_description"] = osc["f_version_description"]

            os_type_list.append(osc["f_os"])
            package_info_list.append(package_info_dict)

            if int(osc["f_os"]) == 2:
                os_type_exist_android.append(int(osc["f_os"]))
                os_type_exist_android.append(osc["f_update_type"])

            if int(osc["f_os"]) == 7:
                os_type_exist_ios.append(int(osc["f_os"]))
                os_type_exist_ios.append(osc["f_update_type"])

        types_keys = OS_TYPE_DEFINE.keys()
        for key in types_keys:
            if int(key) not in os_type_list and int(key) not in [2, 7]:
                package_info_list.append(cls.package_init(key, STANDARD))

        # 判断IOS和安卓类型
        if 2 in os_type_exist_android:
            if STANDARD not in os_type_exist_android:
                package_info_list.append(cls.package_init(2, STANDARD))

            if COSTOM not in os_type_exist_android:
                package_info_list.append(cls.package_init(2, COSTOM))
        else:
            for u_t in [STANDARD, COSTOM]:
                package_info_list.append(cls.package_init(2, u_t))

        if 7 in os_type_exist_ios:
            if STANDARD not in os_type_exist_ios:
                package_info_list.append(cls.package_init(7, STANDARD))
            if COSTOM not in os_type_exist_ios:
                package_info_list.append(cls.package_init(7, COSTOM))
        else:
            for u_t in [STANDARD, COSTOM]:
                package_info_list.append(cls.package_init(7, u_t))

        return package_info_list, ""

    @classmethod
    def package_init(cls, os_type, update_type):
        """
        安装包信息初始化
        :param os_type: 安装包类型 'android', 'mac', 'win32_advanced'。。。
        :param update_type: 上传类型 标椎(standard)/自定义(custom)
        :return: 安装包初始化信息 dict
        """

        package_info = {
            "name": "",
            "url": "",
            "version": "",
            "mode": "",
            "time": "",
            "os_type": OS_TYPE[int(os_type)],
            "size": 0,
            "update_type": update_type,
            "open_download": "",
        }

        return package_info

    @classmethod
    def delete_package(cls, os_type, update_type):
        """
        删除安装包
        :param os_type:安装包类型 'android', 'mac', 'win32_advanced'。。。
        :param update_type:上传类型 标椎(standard)/自定义(custom)
        :return: None
        """

        try:
            if not cls.check_ostype_valid(OS_TYPE[os_type]):
                return "400017009"
        except:
            return "400017009"

        result = ClientPackage.get_client_package_by_ostype_updatetype(os_type, update_type)
        if result:
            if int(result["f_pkg_location"]) == PKG_LOCATION_LOCAL_TO_OSS:
                objectid = OBJECTID[int(os_type)]
                url: str = result["f_url"]
                key = cls.analyse_key(url, objectid)
                ossid = result["f_oss_id"]
                delete_info = cls.get_ossgateway_client().GetDeleteInfo(ossid, key)
                rsp = requests.request(
                    delete_info["method"], delete_info["url"], headers=delete_info["headers"], verify=False
                )
                if ("[2" not in str(rsp)) and ("[404]" not in str(rsp)):
                    return "400017007"
            ClientPackage.delete_client_package_by_ostype(os_type, update_type)

        else:
            return "400017015"

    @classmethod
    def get_update_info_from_filename(cls, filename):
        """
        检查包名是否符合规范
        :param filename: 文件名称 AnyShare_All_Linux_arm64-7.0.1.2-20200221-Terminator-520.rpm
        :return: 文件类型，版本号
        """
        # 检查文件后缀名
        file_suffix = str(filename.split(".")[-1])
        if file_suffix not in FILE_SUFFIX:
            return "400017012", ""

        # 检查文件名分割后的格式
        file_name_list = filename.split("-")
        if len(file_name_list) != 5:
            return "400017006", ""

        # 切割获取文件名第一部分
        file_name_first = file_name_list[0].split("_")
        if len(file_name_first) < 4:
            return "400017006", ""

        # 获取构建号
        file_latest = file_name_list[-1]
        build_num = file_latest.split(".")[0]
        if not build_num.isdigit():
            return "400017006", ""

        # 检查系统类型和后缀名的匹配是否合法
        filenameupper = filename.upper()
        if "WIN" in filenameupper and file_suffix != "exe":
            return "400017014", ""
        elif "LINUX" in filenameupper and file_suffix not in ["deb", "rpm", "AppImage"]:
            return "400017014", ""
        elif "IOS" in filenameupper and file_suffix != "ipa":
            return "400017014", ""
        elif "MAC" in filenameupper and file_suffix not in ("dmg", "tgz", "pkg"):
            return "400017014", ""
        elif "ANDROID" in filenameupper and file_suffix != "apk":
            return "400017014", ""
        if "Linux_mips64" in filename and "AppImage" in filename:
            return "400017014", ""
        # 获取os
        fileos = ""
        if file_name_first[1] != "All":
            return "400017006", ""
        os_type = file_name_first[2].upper()
        arch = file_name_first[3].upper()

        # AnyShare_All_Linux_arm64-7.0.1.2-20200221-Terminator-520.rpm
        # file_name_first -> [AnyShare, All, Linux, arm64]
        # os_type   2-> LINUX
        # arch      3-> ARM64

        if os_type == "WINDOWS" and arch == "X86":
            fileos = 4
        elif os_type == "WINDOWS" and arch == "X64":
            fileos = 8
        elif os_type == "ANDROID" and arch == "ALL":
            fileos = 2
        elif os_type == "MAC" and arch == "X64":
            fileos = 3
        elif os_type == "OFFICE":
            fileos = 6
        elif os_type == "IOS" and arch == "ALL":
            fileos = 7
        elif os_type == "LINUX" and arch == "X64" and file_suffix == "rpm":
            fileos = 9
        elif os_type == "LINUX" and arch == "ARM64" and file_suffix == "rpm":
            fileos = 10
        elif os_type == "LINUX" and arch == "MIPS64" and file_suffix == "rpm":
            fileos = 11
        elif os_type == "LINUX" and arch == "X64" and file_suffix == "deb":
            fileos = 12
        elif os_type == "LINUX" and arch == "X64" and file_suffix == "AppImage":
            fileos = 13
        elif os_type == "LINUX" and arch == "ARM64" and file_suffix == "deb":
            fileos = 14
        elif os_type == "LINUX" and arch == "ARM64" and file_suffix == "AppImage":
            fileos = 15
        elif os_type == "LINUX" and arch == "MIPS64" and file_suffix == "deb":
            fileos = 16
        elif os_type == "WINDOWS" and arch == "ALL":
            fileos = 8
        elif os_type == "OFFICEPLUGIN" and arch == "X86":
            fileos = 17
        elif os_type == "OFFICEPLUGIN" and arch == "X64":
            fileos = 18
        elif os_type == "OFFICEPLUGIN" and arch == "MAC":
            fileos = 19
        else:
            return "400017006", ""

        # 获取版本号,并判断是否合法
        version = file_name_list[1]
        if len(version.split(".")) not in [3, 4]:
            return "400017033", ""

        filever = "%s(%s)" % (version, str(build_num))
        return fileos, filever

    @classmethod
    def get_all_os_config(cls, os_type=None):
        """
        获取客户端开放下载配置
        :param os_type: 安装包类型（可选） 'android', 'mac', 'win32_advanced'。。。
        :return: 开放下载的类型 list
        """

        configs = ClientPackage.get_all_current_used_package_info()
        single_update_type = list()
        update_types = dict()
        if os_type:
            single_update_type = cls.get_update_type(os_type)
        else:
            update_types = cls.get_update_type()
        config_list = list()
        for config in configs:
            if os_type:
                if (
                    config["f_os"] == int(OS_TYPE_DICT[os_type])
                    and config["f_update_type"] == single_update_type[os_type]
                ):
                    if config["f_open_download"]:
                        return [os_type]
                    else:
                        return []
            else:
                if update_types.get(OS_TYPE[config["f_os"]], "") == config["f_update_type"]:
                    if not config["f_open_download"]:
                        f_os = OS_TYPE[config["f_os"]]
                        if f_os in config_list:
                            config_list.pop(config_list.index(f_os))
                    else:
                        config_list.append(OS_TYPE[config["f_os"]])
        if os_type:
            return []
        config_list = list(filter(None, config_list))
        return list(set(config_list))

    @classmethod
    def checkout_storage(cls):
        """
        检查是否有存储
        :return: bool
        """
        # 获取对象存储信息
        access_conf = utils.read_conf_in_config(
            global_param.SERVICE_ACCESS_CONFIG, global_param.SERVICE_ACCESS_FILE_NAME
        )
        oss_info = cls.get_ossgateway_client().GetOSSInfo()

        if not oss_info:
            return False
        return True

    @classmethod
    def set_package_download_info(cls, name, url, os_type, version_description):
        """
        自定义安装包信息
        :param name: 安装包名称 string
        :param url: 下载地址 string
        :param os_type: 安装包类型 'android', 'mac', 'win32_advanced'。。。
        :param version_description: 安装包信息描述
        :return: None
        """
        if os_type not in ["ios", "android"]:
            return "400017018"

        # 判断是否已存在的升级包，如果存在则删除
        if ClientPackage.check_ostype_exist(OS_TYPE.index(os_type), COSTOM):
            cls.delete_package(OS_TYPE.index(os_type), COSTOM)

        # time 获取当前时间作为用户设置的时间
        ctime = time.time()
        timearray = time.localtime(ctime)
        filetime = time.strftime("%Y/%m/%d %H:%M:%S", timearray)

        # 将记录保存到数据库中
        if name:
            fileos, filever = cls.get_update_info_from_filename(name)
            if not filever:
                return fileos

            file_name = name
            version = filever
        else:
            file_name = "ios" if os_type == 7 else "android"
            version = "7.0.0.0"

        package_dict = {
            "f_name": file_name,
            "f_os": OS_TYPE_DICT[os_type],
            "f_size": 0,
            "f_version": version,
            "f_time": filetime,
            "f_mode": 0,
            "f_pkg_location": PKG_LOCATION_STATIC_URL,
            "f_update_type": COSTOM,
            "f_url": url,
            "f_version_description": version_description
        }

        ClientPackage.insert_client_update_package(package_dict)
        return

    @classmethod
    def get_download_url(cls, os_type, req_host, use_https, user_oss_id):
        """
        获取客户端升级包下载链接
        :param os_type: 安装包类型（可选） 'android', 'mac', 'win32_advanced'。。。
        :param req_host: 请求主机ip
        :param use_https: 下载模式 https http
        :param user_oss_id: 用户所在站点的对应的oss_id
        :return: 下载地址 dict
        """
        if not cls.check_ostype_valid(os_type):
            return "", "400017009"

        result = ClientPackage.get_current_used_by_ostype(OS_TYPE_DICT[os_type], cls.check_anyshare_saas_instance())

        if result:
            if (os_type in ["ios", "android"] and int(result["f_pkg_location"]) == PKG_LOCATION_STATIC_URL) or cls.check_anyshare_saas_instance():
                return {"url": result["f_url"]}, ""

            objectid = OBJECTID[int(OS_TYPE_DICT[os_type])]
            url: str = result["f_url"]
            key = cls.analyse_key(url, objectid)
            ossid = result["f_oss_id"]
            filename = result["f_name"]
            delete_info = cls.get_ossgateway_client().GetDownloadInfo(ossid, key, filename, user_oss_id)
            return delete_info, ""
        return "", "400017015"

    @classmethod
    def check_ostype_valid(cls, os_type):
        """
        检查系统类型是否合法
        :param os_type: 安装包类型 'android', 'mac', 'win32_advanced'。。。
        :return: bool
        """
        try:
            if os_type not in OS_TYPE:
                return False
        except:
            return False
        return True

    @classmethod
    def check_package_version(cls, os_type, version, result_str, mode_str, remark_str, silence_str):
        """
        检查是否有升级版本
        :param os_type:安装包类型（可选） 'android', 'mac', 'win32_advanced'。。。
        :param version: 版本号 string
        # :param user_str: 一个封装了用户的userid和用户所属的所有部门/组织/用户组id组成的字符串，逗号分割
        :param result_str: 是否允许下载客户端
        :param mode_str: 是否采用强制更新,取值0/1
        :param remark_str:  保留字段
        :param silence_str:  静默更新
        :return: 升级包信息 dict
        """
        if not cls.check_ostype_valid(os_type):
            return "", "400017009"

        # 根据os_type获取对应的升级包信息，比较版本号

        package = ClientPackage.get_current_used_by_ostype(OS_TYPE_DICT[os_type], cls.check_anyshare_saas_instance())
        # 根据OPA获取策略引擎的数据
        if result_str and mode_str and silence_str:
            silence = True
            if silence_str.lower() == 'false':
                silence = False
            logger.info(f"{result_str}")
            if result_str == "accepted" and package:
                status, msg = cls.compare_version(package["f_version"], version)
                if status:
                    update_type = package["f_update_type"]
                    return {
                        "name": package["f_name"],
                        "os_type": os_type,
                        "version": package["f_version"],
                        "update_type": update_type,
                        "terminator": True if not int(mode_str) else False,
                        "silence": silence
                    }, ""
                elif msg:
                    return "", msg
            else:
                return {"name": "", "os_type": os_type, "version": "-1", "terminator": "", "silence": ""}, ""
        else:
            if package:
                status, msg = cls.compare_version(package["f_version"], version)
                if status:
                    update_type = package["f_update_type"]
                    return {
                        "name": package["f_name"],
                        "os_type": os_type,
                        "version": package["f_version"],
                        "update_type": update_type,
                        "terminator": True if package["f_mode"] else False,
                    }, ""
                elif msg:
                    return "", msg
        return {"name": "", "os_type": os_type, "version": "-1", "terminator": "", "silence": ""}, ""
    
    @classmethod
    def get_os_type_abbreviation(cls, os_type: str) -> str:
        """
        获取客户端类型的简称
        """
        for key, values in OS_TYPE_ABBREVIATION_DICT.items():
            if os_type in values:
                return key
        return None

    @classmethod
    def compare_version(cls, db_ver, com_ver):
        """
        版本比较
        :param db_ver: 数据库中获取的版本号   7.0.0.1(520)
        :param com_ver: 目前请求的版本号      7.0.0.1.555
        :return: Bool
        """
        com_ver_list = com_ver.split(".")
        if len(com_ver_list) not in [4, 5]:
            for x in range(4 - len(com_ver_list)):
                com_ver_list.append(0)
            com_ver_build = 0
        else:
            com_ver_list = com_ver.split(".")[:-1]
            com_ver_build = com_ver.split(".")[-1]

        db_ver = db_ver.split("(")

        db_ver_list = db_ver[0].split(".")
        db_ver_build = 0
        if len(db_ver) > 1:
            db_ver_build = db_ver[-1].split(")")[0]

        if len(com_ver_list) not in [3, 4, 5]:
            return False, "400017033"
        # 当版本位数不同时，以0补位
        if len(com_ver_list) < len(db_ver_list):
            for x in range(len(db_ver_list) - len(com_ver_list)):
                com_ver_list.append(0)
        elif len(com_ver_list) > len(db_ver_list):
            for x in range(len(com_ver_list) - len(db_ver_list)):
                db_ver_list.append(0)
        # 把列表中的字符串转成int

        com_ver_list.append(com_ver_build)
        db_ver_list.append(db_ver_build)

        index = 0
        for com in com_ver_list:
            try:
                com_ver_list[index] = int(com)
            except:
                return False, "400017033"
            index += 1

        index = 0
        for db_ver in db_ver_list:
            db_ver_list[index] = int(db_ver)
            index += 1

        if db_ver_list > com_ver_list:
            return True, ""

        return False, ""

    @classmethod
    def set_package_config(cls, os_type, update_type, version_description=None, open_download=None):
        """
        设置升级包是否开放下载，是否强制下载
        :param os_type:安装包类型 'android', 'mac', 'win32_advanced'。。。
        :param update_type: 上传类型 标椎(standard)/自定义(custom)
        :param version_description: 客户端升级包描述信息
        :param open_download: 是否开放下载 bool
        :return: None
        """
        if not cls.check_ostype_valid(os_type):
            return "400017009"
        if update_type not in [STANDARD, COSTOM]:
            return "400017030"
        if not ClientPackage.get_client_package_by_ostype_updatetype(OS_TYPE_DICT[os_type], update_type):
            return "400017015"
        ClientPackage.set_client_package_config(
            OS_TYPE_DICT[os_type], update_type, version_description=version_description, open_download=open_download
        )
        return

    @classmethod
    def set_update_type(cls, os_type, update_type):
        """
        设置升级使用类型
        :param os_type: 安装包类型 'android', 'mac', 'win32_advanced'。。。
        :param update_type: 上传类型 标椎(standard)/自定义(custom)
        :return: None
        """
        if update_type not in [STANDARD, COSTOM]:
            return "400017030"
        ClientPackage.set_client_package_update_type(OS_TYPE_DICT[os_type], update_type)
        return

    @classmethod
    def get_update_type(cls, os_type=None):
        """
        获取安装包的上传类型 自定义/标椎
        :param os_type: 安装包类型（可选） 'android', 'mac', 'win32_advanced'。。。
        :return: dict
        """
        os_type_param = None
        if os_type:
            if not cls.check_ostype_valid(os_type):
                return "400017009"
            os_type_param = OS_TYPE_DICT[os_type]
        configs = ClientPackage.get_client_package_update_type(os_type=os_type_param)
        config_dict = dict()
        for config in configs:
            if int(config["f_os"]) in [2, 7]:
                config_dict[OS_TYPE[int(config["f_os"])]] = config["f_mode"]
        if os_type and int(os_type_param) not in [2, 7]:
            config_dict[os_type] = STANDARD
        if config_dict and os_type:
            return config_dict
        for k, v in OS_TYPE_DICT.items():
            if int(v) in [2, 7]:
                continue
            config_dict[k] = STANDARD
        # 如果是云端一体化场景的话，直接全改成自定义更新
        if cls.check_anyshare_saas_instance():
            for k, v in OS_TYPE_DICT.items():
                config_dict[k] = COSTOM
        return config_dict

    @classmethod
    def get_default_ossid(cls):
        return cls.get_ossgateway_client().GetSiteDefaultOSS()["storage_id"]

    @classmethod
    def get_ossgateway_client(cls):
        return OssGatewayManagerClient.instance()

    @classmethod
    def get_access_info(cls):
        if cls.access_conf is None:
            cls.access_conf = utils.read_conf_in_config(
                global_param.SERVICE_ACCESS_CONFIG, global_param.SERVICE_ACCESS_FILE_NAME
            )
        return cls.access_conf

    @classmethod
    def create_upload_infomation(cls, package_name):
        CID = str(uuid.uuid1()).replace("-", "")

        fileos, filever = cls.get_update_info_from_filename(package_name)

        if not filever:
            err_code = fileos
            return err_code, ""
        # 获取对象存储信息
        ossid = cls.get_default_ossid()
        if not ossid:
            return "500017008", ""

        key = "/".join([cls.oss_key_prefix, CID, OBJECTID[fileos]])

        return "", cls.get_ossgateway_client().GetUploadInfo(ossid, key, query_string=False)

    @classmethod
    def complete_upload(cls, package_name, filesize, version_description, url):
        fileos, filever = cls.get_update_info_from_filename(package_name)
        if not filever:
            err_code = fileos
            return err_code
        # time 获取当前时间作为用户上传时间
        ctime = time.time()
        timearray = time.localtime(ctime)
        filetime = time.strftime("%Y/%m/%d %H:%M:%S", timearray)
        # 获取对象存储信息
        ossid = cls.get_default_ossid()
        if not ossid:
            return "500017008"
        package_dict = {
            "f_name": package_name,
            "f_os": fileos,
            "f_size": filesize,
            "f_version": filever,
            "f_time": filetime,
            "f_mode": 0,
            "f_pkg_location": PKG_LOCATION_LOCAL_TO_OSS,
            "f_url": url,
            "f_oss_id": ossid,
            "f_update_type": STANDARD,
            "f_version_description": version_description
        }
        # 删除已存在的升级包
        if ClientPackage.check_ostype_exist(fileos, STANDARD):
            cls.delete_package(fileos, STANDARD)
        ClientPackage.insert_client_update_package(package_dict)
        return ""

    @classmethod
    def analyse_key(cls, url, objectid):
        path = urlparse(url).path
        path_sp = path.split("/")
        objectid_index = path_sp.index(objectid)
        return "/".join(path_sp[objectid_index - 2 : objectid_index + 1])
