#!/usr/bin/env python
# -*- coding: utf-8 -*-

import copy
import json
import traceback

from tornado import web

from src.common.config import ERROR_DEFINE
from src.common.error import OtherHttpError
from src.common.log_util import logger
from src.modules.ossgateway_service_manager import OSSGatewayManager

error_result_tem = {"code": "", "message": "", "cause": ""}

OSS_SERVICE_NAME = "OSSGatewayService"


class OSSHandler(web.RequestHandler):
    def post(self):
        try:
            data = json.loads(self.request.body)
            logger.info(f"OSSHandler: post request begin, body: {data}")
            code = OSSGatewayManager().add_oss_config(OSS_SERVICE_NAME, data)
            if code:
                ossgateway_error_result = {
                    "code": code,
                    "message": "can't connect to bucket store.",
                    "cause": ERROR_DEFINE[code],
                }
                if code == "400017247":
                    ossgateway_error_result = {
                        "code": code,
                        "message": "A bucket with the same name already exists.",
                        "cause": ERROR_DEFINE[code],
                    }
                self.write(ossgateway_error_result)
                self.set_status(400)
            else:
                self.set_status(200)
                logger.info(f"OSSHandler: post request success")
        except OtherHttpError as oex:
            oex.reply(self)
        except Exception as ex:
            code = "500017000"
            error_result = copy.deepcopy(error_result_tem)
            error_result["code"] = code
            error_result["message"] = str(ex)
            error_result["cause"] = "add oss config failed."
            self.write(error_result)
            self.set_status(500)
            raise Exception(traceback.format_exc())
        finally:
            self.finish()

    def put(self):
        try:
            data = json.loads(self.request.body)
            logger.info(f"OSSHandler: put request begin, body: {data}")
            code = OSSGatewayManager().update_oss_config(OSS_SERVICE_NAME, data)
            if code:
                ossgateway_error_result = {
                    "code": code,
                    "message": "can't connect to bucket store.",
                    "cause": ERROR_DEFINE[code],
                }
                if code == "400017248":
                    ossgateway_error_result = {
                        "code": code,
                        "message": "OSSGateway service not installed",
                        "cause": ERROR_DEFINE[code],
                    }

                self.write(ossgateway_error_result)
                self.set_status(400)
            else:
                self.set_status(200)
                logger.info(f"OSSHandler: put request success")
        except OtherHttpError as oex:
            oex.reply(self)
        except Exception as ex:
            code = "500017000"
            error_result = copy.deepcopy(error_result_tem)
            error_result["code"] = code
            error_result["message"] = str(ex)
            error_result["cause"] = "update oss config failed."
            self.write(error_result)
            self.set_status(500)
            raise Exception(traceback.format_exc())
        finally:
            self.finish()
