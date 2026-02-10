import json
from abc import ABC

from tornado import web,httputil

from src.common.config import ERROR_DEFINE
from src.modules import communication_manager

from src.common.error import *
from typing import Union


class CommunicationHandler(web.RequestHandler, ABC):

    @async_deal_error_and_finish(message="init communication failed")
    async def init_communication(self):
        body = self.parse_body()
        if body is None:
            return
        communication_manager.init_communication(body)

    @async_deal_error_and_finish(message="update communication failed")
    async def update_communication(self):
        body = self.parse_body()
        if body is None:
            return
        await communication_manager.update_communication(body)
    
    def parse_body(self) -> Union[dict, None]:
        try:
            # only support json body
            body_dict = json.loads(self.request.body)
        except Exception as e:
            self.write({
                "code": "400017005", 
                "message": "body parse error, only support json, error: {}".format(e),
                "cause": ERROR_DEFINE["400017005"],
                })
            self.set_status(400)
            return None
            
        if body_dict is None:
            body_dict = {
                "backend-service": {},
                "oauth2": {}
            }
        if body_dict.get("backend-service") is None:
            body_dict["backend-service"] = {}
        if body_dict["oauth2"] is None:
            body_dict["oauth2"] = {}
        return body_dict

    post = init_communication
    put = update_communication
