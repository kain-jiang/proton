#!/usr/bin/env python
# -*- coding: utf-8 -*-
# @Time    : 2021/12/21 15:11
# @Author  : Lee.li
# @Site    : 
# @File    : test_client_package_manager.py
# @Software: PyCharm
import unittest
from unittest.mock import patch

from src.lib.db.client_package import ClientPackage
from src.modules.client_package_manager import ClientPackageInfoManager


class TestClientPackageInfoManager(unittest.TestCase):
    def setUp(self) -> None:
        self.get_client_package_by_ostype_patcher = patch.object(ClientPackage, "get_client_package_by_ostype")
        self.get_all_client_package_info_patcher = patch.object(ClientPackage, "get_all_client_package_info")
        self.read_conf_in_config_patcher = patch("src.modules.client_package_manager.utils.read_conf_in_config")
        self.get_client_package_by_ostype_updatetype_patcher = patch.object(ClientPackage,
                                                                            "get_client_package_by_ostype_updatetype")
        self.tclinet_patcher = patch("src.common.tclients.TClient")
        self.delete_client_package_by_ostype_patcher = patch.object(ClientPackage, "delete_client_package_by_ostype")
        # self.GetDeleteInfo_patcher = patch("src.modules.client_package_manager.Client.GetDeleteInfo")
        self.get_client_package_update_type_patcher = patch.object(ClientPackage, "get_client_package_update_type")
        self.get_current_used_by_ostype_patcher = patch.object(ClientPackage, "get_current_used_by_ostype")

    def tearDown(self) -> None:
        pass

    def test_get_package_info(self):
        # os_type is not none
        os_type = "android"
        single_packages = [
            {
                "f_name": "ios",
                "f_url": "url",
                "f_pkg_location": "2",
                "f_version": "1.0.0",
                "f_mode": "mode",
                "f_time": "time",
                "f_os": "2",
                "f_size": "100",
                "f_update_type": "update",
                "f_open_download": "True"
            }
        ]
        get_get_client_package_by_ostype_patcher = self.get_client_package_by_ostype_patcher.start()

        get_get_client_package_by_ostype_patcher.return_value = single_packages

        ClientPackageInfoManager().get_package_info(os_type)

        # os_type is none
        os_type = ""
        packages = [{
            "f_name": "ios",
            "f_url": "url",
            "f_pkg_location": "2",
            "f_version": "1.0.0",
            "f_mode": "mode",
            "f_time": "time",
            "f_os": "2",
            "f_size": "200",
            "f_update_type": "update",
            "f_open_download": "True"
        }]
        get_get_all_client_package_info_patcher = self.get_all_client_package_info_patcher.start()

        get_get_all_client_package_info_patcher.return_value = packages

        ClientPackageInfoManager().get_package_info(os_type)

    def test_delete_package(self):
        os_type = 3
        update_type = "standard"
        result = {
            "f_name": "ios",
            "f_url": "url",
            "f_pkg_location": "2",
            "f_version": "1.0.0",
            "f_mode": "mode",
            "f_time": "time",
            "f_os": "2",
            "f_size": "200",
            "f_update_type": "update",
            "f_open_download": "True"
        }
        result2 = ""
        access_conf = {"efast", "evfsThriftHost"}

        get_get_client_package_by_ostype_updatetype_patcher = self.get_client_package_by_ostype_updatetype_patcher.start()
        self.tclinet_patcher.start()
        # self.GetDeleteInfo_patcher.start()

        get_get_client_package_by_ostype_updatetype_patcher.return_value = result2

        ClientPackageInfoManager().delete_package(os_type, update_type)

    def test_compare_version(self):
        db_ver = "7.0.0.1(520)"
        com_ver = "7.0.0.1.555"

        ClientPackageInfoManager().compare_version(db_ver, com_ver)

    def test_get_update_type(self):
        os_type = "android"
        configs = [{
            "f_name": "ios",
            "f_url": "url",
            "f_pkg_location": "2",
            "f_version": "1.0.0",
            "f_mode": "mode",
            "f_time": "time",
            "f_os": "2",
            "f_size": "200",
            "f_update_type": "update",
            "f_open_download": "True"
        }]

        get_get_client_package_update_type = self.get_client_package_update_type_patcher.start()

        get_get_client_package_update_type.return_value = configs

        ClientPackageInfoManager().get_update_type(os_type)

    def test_get_update_info_from_filename(self):
        #
        filename = "AnyShare_All_Linux_arm64-7.0.1.2-20200221-Terminator-520.rpm"

        ClientPackageInfoManager().get_update_info_from_filename(filename)

        #
        filename = "AnyShare_All_WINDOWS_X64-7.0.1.2-20200221-Terminator-520.rpm"

        ClientPackageInfoManager().get_update_info_from_filename(filename)

        #
        filename = "AnyShare_All_WINDOWS_ALL-7.0.1.2-20200221-Terminator-520.rpm"

        ClientPackageInfoManager().get_update_info_from_filename(filename)

        #
        filename = "AnyShare_All_Linux_X64-7.0.1.2-20200221-Terminator-520.rpm"

        ClientPackageInfoManager().get_update_info_from_filename(filename)

        #
        filename = "AnyShare_All_Linux_X64-7.0.1.2-20200221-Terminator-520.deb"

        ClientPackageInfoManager().get_update_info_from_filename(filename)

        #
        filename = "AnyShare_All_Linux_arm64-7.0.1.2-20200221-Terminator-520.AppImage"

        ClientPackageInfoManager().get_update_info_from_filename(filename)

        #
        filename = "AnyShare_All_Linux_arm64-7.0.1.2-20200221-Terminator-520.deb"

        ClientPackageInfoManager().get_update_info_from_filename(filename)

        #
        filename = "AnyShare_All_Linux_X64-7.0.1.2-20200221-Terminator-520.AppImage"

        ClientPackageInfoManager().get_update_info_from_filename(filename)

        #
        filename = "AnyShare_All_Linux_MIPS64-7.0.1.2-20200221-Terminator-520.deb"

        ClientPackageInfoManager().get_update_info_from_filename(filename)

        #
        filename = "AnyShare_All_Linux_MIPS64-7.0.1.2-20200221-Terminator-520.rpm"

        ClientPackageInfoManager().get_update_info_from_filename(filename)

    def test_get_download_url(self):
        os_type = "android"
        req_host = "1.0.0.0"
        use_https = "1.0.0.0"

        result = {
            "f_name": "ios",
            "f_url": "url",
            "f_pkg_location": "2",
            "f_version": "1.0.0",
            "f_mode": "mode",
            "f_time": "time",
            "f_os": "2",
            "f_size": "200",
            "f_update_type": "update",
            "f_open_download": "True"
        }

        get_get_current_used_by_ostype_patcher = self.get_current_used_by_ostype_patcher.start()

        get_get_current_used_by_ostype_patcher.return_value = result

        ClientPackageInfoManager().get_download_url(os_type, req_host, use_https)


if __name__ == '__main__':
    unittest.main()
