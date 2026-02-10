#!/usr/bin/env python
# -*- coding: utf-8 -*-
import copy
import json
import traceback

from tornado import web

from src.common.config import ERROR_DEFINE
from src.modules.ossgateway_service_manager import OSSGatewayManager
from src.modules.obs_instances_manager import ObsInstanceManager
from src.common.log_util import logger

error_result_tem = {"code": "", "message": "", "cause": ""}
message = dict()
message["400017246"] = "can't connect to bucket store."
message["400017247"] = "A bucket with the same name already exists."
message["400017248"] = "OSSGateway service not installed."
message["400017249"] = "A bucket with the same name already exists in other sites"
message["400017250"] = "A oss with the same name already exists in other sites."





class MultiContainerizedManagerHandler(web.RequestHandler):
    def post(self, service):
        try:
            data = json.loads(self.request.body)
            logger.info(f"MultiContainerizedManagerHandler: post request begin, service: {service}, body: {data}")
            code = OSSGatewayManager().add_oss_config(service, data)
            if code:
                ossgateway_error_result = {
                    "code": code,
                    "message": message[code],
                    "cause": ERROR_DEFINE[code],
                }
                self.write(ossgateway_error_result)
                self.set_status(400)
            else:
                self.set_status(200)
                logger.info(f"MultiContainerizedManagerHandler: post request success, service: {service}")
        except Exception as ex:
            code = "500017000"
            error_result = copy.deepcopy(error_result_tem)
            error_result["code"] = code
            error_result["message"] = str(ex)
            error_result["cause"] = "install ossgateway faild."
            self.write(error_result)
            self.set_status(500)
            raise Exception(traceback.format_exc())
        finally:
            self.finish()

    def put(self, service):
        try:
            data = json.loads(self.request.body)
            logger.info(f"MultiContainerizedManagerHandler: put request begin, service: {service}, body: {data}")
            code = OSSGatewayManager().update_oss_config(service, data)
            if code:
                ossgateway_error_result = {
                    "code": code,
                    "message": message[code],
                    "cause": ERROR_DEFINE[code],
                }
                self.write(ossgateway_error_result)
                self.set_status(400)
            else:
                self.set_status(200)
                logger.info(f"MultiContainerizedManagerHandler: put request success, service: {service}")
        except Exception as ex:
            code = "500017000"
            error_result = copy.deepcopy(error_result_tem)
            error_result["code"] = code
            error_result["message"] = str(ex)
            error_result["cause"] = "update ossgateway config faild."
            self.write(error_result)
            self.set_status(500)
            raise Exception(traceback.format_exc())
        finally:
            self.finish()


class UPMultiContainerizedManagerHandler(web.RequestHandler):
    def post(self, service, nodes=None):
        try:
            body_data = json.loads(self.request.body) if self.request.body else []
            logger.info(f"UPMultiContainerizedManagerHandler: post request begin, service: {service}, body: {body_data}")
            if nodes is not None:
                node_ips = nodes.split(",")
            if service == "OSSGatewayService":
                if body_data:
                    OSSGatewayManager().install_abrestore_ossgateway(body_data)
                else:
                    OSSGatewayManager().upgrade_ossgateway_versoin()
                logger.info(f"UPMultiContainerizedManagerHandler: post request success, service: {service}")
        except Exception as ex:
            code = "500017000"
            error_result = copy.deepcopy(error_result_tem)
            error_result["code"] = code
            error_result["message"] = str(ex)
            error_result["cause"] = "install ossgateway faild."
            self.write(error_result)
            self.set_status(500)
            raise Exception(traceback.format_exc())
        finally:
            self.finish()

    def put(self, service, nodes=None):
        try:
            if nodes is not None:
                node_ips = nodes.split(",")
            logger.info(f"UPMultiContainerizedManagerHandler: put request begin, service: {service}")
            OSSGatewayManager().upgrade_ossgateway_service(service)
            logger.info(f"UPMultiContainerizedManagerHandler: put request success, service: {service}")
        except Exception as ex:
            code = "500017000"
            error_result = copy.deepcopy(error_result_tem)
            error_result["code"] = code
            error_result["message"] = str(ex)
            error_result["cause"] = "upgrade ossgateway faild."
            self.write(error_result)
            self.set_status(500)
            raise Exception(traceback.format_exc())
        finally:
            self.finish()



class ObsInstancesHandler(web.RequestHandler):
    def post(self):
        try:
            data = json.loads(self.request.body)
            headers = self.request.headers
            # token_id = headers['Authorization'].lstrip("Bearer ")
            token_id = headers['Authorization'][7:]
            flag, result_data = ObsInstanceManager().add_obs_config(data, token_id)
            if flag:
                self.write(result_data)
                self.set_status(201)
            else:
                code = result_data
                ossgateway_error_result = {
                    "code": code,
                    "description": message[code],
                    "solution": "",
                    "detail": ERROR_DEFINE[code],
                    "link": "",
                }
                self.write(ossgateway_error_result)
                self.set_status(400)                
        except Exception as ex:
            code = "500017000"
            error_result = copy.deepcopy(error_result_tem)
            error_result = {"code": "", "description": "", "solution": "", "detail": "",  "link": ""}
            error_result["code"] = code
            error_result["description"] = "install ossgateway faild."
            error_result["detail"] = str(ex)
            self.write(error_result)
            self.set_status(500)
            raise Exception(traceback.format_exc())
        finally:
            self.finish()

    def put(self):
        try:
            data = json.loads(self.request.body)
            headers = self.request.headers
            token_id = headers['Authorization'][7:]
            ObsInstanceManager().update_obs_config(data, token_id)
            self.set_status(204)              
        except Exception as ex:
            code = "500017000"
            error_result = copy.deepcopy(error_result_tem)
            error_result = {"code": "", "description": "", "solution": "", "detail": "",  "link": ""}
            error_result["code"] = code
            error_result["description"] = "update obs faild."
            error_result["detail"] = str(ex)
            self.write(error_result)
            self.set_status(500)
            raise Exception(traceback.format_exc())
        finally:
            self.finish()
