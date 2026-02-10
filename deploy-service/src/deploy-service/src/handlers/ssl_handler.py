#!/usr/bin/env python
# -*- coding: utf-8 -*-
# @Time : 2020/6/19 10:48
# @Author : Louis.she
# @Site :
# @File : https_handler.py
# @Software: PyCharm
import json
import traceback

import tornado.web

from src.common.config import ERROR_DEFINE
from src.common.error import deal_error_and_finish
from src.modules.ssl_manager import CertDownloadFeature, SSLManager


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


class SetGlobalCertHandler(BaseHandler):
    def get(self, ip):
        try:
            result, msg = SSLManager().set_global_https(ip)
            if result:
                result = {"code": result, "message": msg, "cause": ERROR_DEFINE[result]}
                self.write(result)
                self.set_status(400)
                return
            self.set_status(200)
            self.finish()
        except Exception as e:
            self.write({"code": 500017000, "message": "set global cert failed", "cause": str(e)})
            self.set_status(500)
            self.finish()
            raise Exception(traceback.format_exc())


class GetCertInfoHandler(BaseHandler):
    def get(self):

        cert_type = self.get_argument("cert_type", "")
        try:
            status, result = SSLManager().get_cert_info(cert_type=cert_type)
            if status:
                result = {"code": status, "message": result, "cause": ERROR_DEFINE[status]}
                self.write(json.dumps(result))
                self.set_status(400)
                return
            self.write(json.dumps(result))
            self.set_status(200)
            self.finish()
        except Exception as e:
            self.write({"code": 500017000, "message": "get cert info failed", "cause": str(e)})
            self.set_status(500)
            self.finish()
            raise Exception(traceback.format_exc())


class DownloadCertHandler(BaseHandler):
    def get(self):
        try:

            self.set_header("Content-Type", "application/octet-stream")
            self.set_header("Content-Disposition", "attachment; filename=ca.crt")
            data = SSLManager().download_cert()
            self.write(data)
        except Exception as e:
            self.write({"code": 500017000, "message": "download ca cert failed", "cause": str(e)})
            self.set_status(500)
            self.finish()
            raise Exception(traceback.format_exc())


class UploadCertHandler(BaseHandler):
    def put(self):
        try:

            data = json.loads(self.request.body)
            cert_crt = data.get("cert_crt", "")
            cert_key = data.get("cert_key", "")
            if not cert_crt or not cert_key:
                result = {"code": "400017010", "message": "file is does not exist", "cause": ERROR_DEFINE["400017010"]}
                self.write(result)
                self.set_status(400)
                return

            status, _result = SSLManager().upload_cert(cert_crt, cert_key)
            if status:
                result = {"code": status, "message": _result, "cause": ERROR_DEFINE[status]}
                self.write(result)
                self.set_status(400)
                return
            self.set_status(200)
            return
        except Exception as e:
            self.write({"code": 500017000, "message": "upload cert failed", "cause": str(e)})
            self.set_status(500)
            self.finish()
            raise Exception(traceback.format_exc())


class CertDownloadFeatureHandler(BaseHandler):

    @deal_error_and_finish(message="get cert download feature status failed")
    def download_feature_status(self):
        self.set_status(200)
        self.write({"status": CertDownloadFeature.get_status()})

    @deal_error_and_finish(message="disable cert download feature failed")
    def disable_download_feature(self):
        CertDownloadFeature.set_status(False)
        CertDownloadFeature.get_cache_status(-1)
        self.set_status(200)

    @deal_error_and_finish(message="enable cert download feature failed")
    def enable_download_feature(self):
        CertDownloadFeature.set_status(True)
        CertDownloadFeature.get_cache_status(1)
        self.set_status(200)


    get = download_feature_status
    delete = disable_download_feature
    post = put = enable_download_feature

