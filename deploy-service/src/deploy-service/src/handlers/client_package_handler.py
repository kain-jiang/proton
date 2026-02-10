#!/usr/bin/env python
# -*- coding: utf-8 -*-
# @Time : 2020/5/19 11:24
# @Author : Louis.she
# @Site :
# @File : client_package.py
# @Software: PyCharm
import json
import traceback

import tornado.web

from src.common.config import ERROR_DEFINE, OS_TYPE
from src.modules.client_package_manager import ClientPackageInfoManager


class BaseHandler(tornado.web.RequestHandler):
    def set_default_headers(self):
        self.set_header("Access-Control-Allow-Credentials", "false")
        self.set_header("Access-Control-Max-Age", 1728000)
        self.set_header("Content-Type", "application/json")

    def get_error(self, obj):
        try:
            return "", ERROR_DEFINE[[str(obj)]]
        except Exception as e:
            return "", ""

    def data_received(self, chunk):
        pass

    def options(self):
        self.set_header("Content-Type", "text/plain; charset=UTF-8")
        self.set_status(204)
        self.finish()


class PackageHandler(BaseHandler):
    # 获取所有上传包的信息

    def get(self):
        try:
            ostype = self.get_argument("os_type", "")
            if ostype == "windows":
                ostype = "win64_advanced"
            result, status = ClientPackageInfoManager.get_package_info(os_type=ostype)
            if status:
                result = {"code": status, "message": "get package info failed", "cause": ERROR_DEFINE[status]}
                self.write(result)
                self.set_status(400)
                return
            self.write(json.dumps(result))
            self.set_status(200)
            self.finish()
        except Exception as e:
            self.write({"code": 500017000, "message": "get package info failed", "cause": str(e)})
            self.set_status(500)
            self.finish()
            raise e


    # 删除指定上传包的信息

    def delete(self, ostype):
        update_type = self.get_argument("update_type", "")
        if ostype == "windows":
            ostype = "win64_advanced"
        if not ostype or not update_type or ostype not in OS_TYPE:
            result = {"code": "400017005", "message": "delete package failed", "cause": ERROR_DEFINE["400017005"]}
            self.write(result)
            self.set_status(400)
            return
        try:

            os_type = OS_TYPE.index(ostype)

            result = ClientPackageInfoManager.delete_package(os_type, update_type)
            if result:
                result = {"code": result, "message": "delete package failed", "cause": ERROR_DEFINE[result]}
                self.write(result)
                self.set_status(400)
                return
            self.set_status(200)
            self.finish()
        except Exception as e:
            self.write({"code": 500017000, "message": "delete package failed", "cause": str(e)})
            self.set_status(500)
            self.finish()
            raise Exception(traceback.format_exc())

class ClientPackageUploadHandler(BaseHandler):
    def create_upload_information(self):
        # 创建上传信息 [GET]
        try:
            package_name = self.get_argument("package_name", "")
            if not package_name:
                raise Exception("package_name is required.")
            code, info = ClientPackageInfoManager.create_upload_infomation(package_name)
            if code:
                self.write( {"code": code, "message": "upload client package failed", "cause": ERROR_DEFINE[code]})
                self.set_status(400)
                self.finish()
                return
            else:
                self.write(info)
                self.set_status(200)
                self.finish()
        except Exception as e:
            self.write({"code": 500017000, "message": "upload client package failed", "cause": str(e)})
            self.set_status(500)
            self.finish()
            raise Exception(traceback.format_exc())

    def complete_upload(self):
        # 创建上传信息 [POST]
        try:
            params = json.loads(self.request.body or "{}")
            package_name = params["package_name"]
            filesize = params.get("filesize", 0)
            url = params["url"]
            version_description = params.get("version_description", "")
            code = ClientPackageInfoManager.complete_upload(package_name, filesize, version_description, url)
            if code:
                self.write( {"code": code, "message": "upload client package failed", "cause": ERROR_DEFINE[code]})
                self.set_status(400)
                self.finish()
                return
            else:
                self.set_status(200)
                self.finish()
        except Exception as e:
            self.write({"code": 500017000, "message": "upload client package failed", "cause": str(e)})
            self.set_status(500)
            self.finish()
            raise Exception(traceback.format_exc())

    get = create_upload_information
    post = complete_upload

class OsConfigHandler(BaseHandler):
    def get(self):
        osType = self.get_argument("os_type", "")
        if osType and osType not in OS_TYPE:
            result = {"code": "400017005", "message": "get config failed", "cause": ERROR_DEFINE["400017005"]}
            self.write(result)
            self.set_status(400)
            return
        try:
            result = ClientPackageInfoManager.get_all_os_config(os_type=str(osType))
            result = json.dumps(result)
            self.write(result)
            self.set_status(200)
            self.finish()
        except Exception as e:
            self.write({"code": 500017000, "message": "get config failed", "cause": str(e)})
            self.set_status(500)
            self.finish()
            raise Exception(traceback.format_exc())


class CheckStorageHandler(BaseHandler):
    def get(self):
        try:
            result = ClientPackageInfoManager.checkout_storage()
            self.write({"result": result})
            self.set_status(200)
            self.finish()
        except Exception as e:
            self.write({"code": 500017000, "message": "get config failed", "cause": str(e)})
            self.set_status(500)
            self.finish()
            raise Exception(traceback.format_exc())


class SetDownloadInfoHandler(BaseHandler):
    def post(self):
        data = self.request.body
        data = json.loads(data)
        name = data.get("name", "")
        url = data.get("url", "")
        ostype = data.get("os_type", "")
        version_description = data.get("version_description", "")

        try:
            result = ClientPackageInfoManager.set_package_download_info(name, url, ostype, version_description)
            if result:
                result = {"code": result, "message": "set download info failed", "cause": ERROR_DEFINE[result]}
                self.write(result)
                self.set_status(400)
                return
            self.set_status(200)
            self.finish()
        except Exception as e:
            self.write({"code": 500017000, "message": "set download info failed", "cause": str(e)})
            self.set_status(500)
            self.finish()
            raise Exception(traceback.format_exc())


class GetDownloadUrlHandler(BaseHandler):
    def post(self):
        data = self.request.body
        data = json.loads(data)
        osType = data.get("os_type", "")
        if osType == "windows":
            osType = "win64_advanced"
        reqHost = data.get("req_host", "")
        useHttps = True
        user_oss_id = data.get("user_oss_id", "")
        try:
            if not osType or not reqHost:
                result = {
                    "code": "400017005",
                    "message": "get package download url failed",
                    "cause": ERROR_DEFINE["400017005"],
                }
                self.write(result)
                self.set_status(400)
                return
            result, status = ClientPackageInfoManager.get_download_url(osType, reqHost, useHttps, user_oss_id)
            if status:
                result = {"code": status, "message": "get package download url failed", "cause": ERROR_DEFINE[status]}
                self.write(result)
                self.set_status(400)
                return
            self.write(result)
            self.set_status(200)
            self.finish()
        except Exception as e:
            self.write({"code": 500017000, "message": "get package download url failed", "cause": str(e)})
            self.set_status(500)
            self.finish()
            raise Exception(traceback.format_exc())

    def get(self):
        osType = self.get_argument("os_type", "")
        if osType == "windows":
            osType = "win64_advanced"
        reqHost = self.get_argument("req_host", "")
        useHttps = True
        user_oss_id = self.get_argument("user_oss_id", "")
        try:
            if not osType or not reqHost:
                result = {
                    "code": "400017005",
                    "message": "get package download url failed",
                    "cause": ERROR_DEFINE["400017005"],
                }
                self.write(result)
                self.set_status(400)
                return
            result, status = ClientPackageInfoManager.get_download_url(osType, reqHost, useHttps, user_oss_id)
            if status:
                result = {"code": status, "message": "get package download url failed", "cause": ERROR_DEFINE[status]}
                self.write(result)
                self.set_status(400)
                return
            self.write(result)
            self.set_status(200)
            self.finish()
        except Exception as e:
            self.write({"code": 500017000, "message": "get package download url failed", "cause": str(e)})
            self.set_status(500)
            self.finish()
            raise Exception(traceback.format_exc())


class PackageVersionHandler(BaseHandler):
    def get(self, ostype, version):
        if ostype == "windows":
            ostype = "win64_advanced"
        # user_str = self.get_argument("user", "")
        # 决策结果
        result_str = self.get_argument("result", "")
        mode_str = self.get_argument("mode", "")
        remark_str = self.get_argument("remark", "")
        silence_str = self.get_argument("silence", "")
        if not ostype or not version:
            result = {
                "code": "400017005",
                "message": "check package version failed",
                "cause": ERROR_DEFINE["400017005"],
            }
            self.write(result)
            self.set_status(400)
            return

        try:
            result, status = ClientPackageInfoManager.check_package_version(ostype, version, result_str, mode_str, remark_str, silence_str)
            if status:
                result = {"code": status, "message": "check package version failed", "cause": ERROR_DEFINE[status]}
                self.write(result)
                self.set_status(400)
                return
            self.write(result)
            self.set_status(200)
            self.finish()
        except Exception as e:
            self.write({"code": 500017000, "message": "check package version failed", "cause": str(e)})
            self.set_status(500)
            self.finish()
            raise Exception(traceback.format_exc())


class SetVersionDescriptionHandler(BaseHandler):
    def post(self):
        data = self.request.body
        data = json.loads(data)
        osType = data.get("os_type", "")
        version_description = data.get("version_description")
        update_type = data.get("update_type", "")
        open_download = data.get("open_download", "")
        if (
            not osType
            or not update_type
            or open_download
            and not isinstance(open_download, bool)
        ):
            result = {"code": "400017005", "message": "set version description failed", "cause": ERROR_DEFINE["400017005"]}
            self.write(result)
            self.set_status(400)
            return

        try:
            result = ClientPackageInfoManager.set_package_config(
                osType, update_type, version_description=version_description, open_download=open_download
            )
            if result:
                result = {"code": result, "message": "set version description failed", "cause": ERROR_DEFINE[result]}
                self.write(result)
                self.set_status(400)
                return
            self.set_status(200)
            self.finish()
        except Exception as e:
            self.write({"code": 500017000, "message": "set version description failed", "cause": str(e)})
            self.set_status(500)
            self.finish()
            raise Exception(traceback.format_exc())

class UpadteTypeHandler(BaseHandler):
    def put(self):
        data = self.request.body
        data = json.loads(data)
        osType = data.get("os_type", "")
        if osType == "windows":
            osType = "win64_advanced"
        update_type = data.get("update_type", "")
        if not osType or not update_type or osType != "android":
            result = {"code": "400017005", "message": "set update type failed", "cause": ERROR_DEFINE["400017005"]}
            self.write(result)
            self.set_status(400)
            return

        try:
            result = ClientPackageInfoManager.set_update_type(osType, update_type)
            if result:
                result = {"code": result, "message": "set update type failed", "cause": ERROR_DEFINE[result]}
                self.write(result)
                self.set_status(400)
                return
            self.set_status(200)
            self.finish()
        except Exception as e:
            self.write({"code": 500017000, "message": "set update type failed", "cause": str(e)})
            self.set_status(500)
            self.finish()
            raise Exception(traceback.format_exc())

    def get(self):
        osType = self.get_argument("os_type", "")
        if osType == "windows":
            osType = "win64_advanced"
        try:
            result = ClientPackageInfoManager.get_update_type(os_type=osType)
            try:
                if result:
                    int(result)
                    result = {
                        "code": result,
                        "message": "get package update type failed",
                        "cause": ERROR_DEFINE[result],
                    }
                    self.write(result)
                    self.set_status(400)
                    return
            except:
                pass
            self.write(result)
            self.set_status(200)
            self.finish()
        except Exception as e:
            self.write({"code": 500017000, "message": "get update type failed", "cause": str(e)})
            self.set_status(500)
            self.finish()
            raise Exception(traceback.format_exc())
