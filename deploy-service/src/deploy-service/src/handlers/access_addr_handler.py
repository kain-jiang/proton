#!/usr/bin/env python
# -*- coding: utf-8 -*-
import json

from tornado import web
from nslookup import Nslookup
from IPy import IP

from src.common.error import deal_error_and_finish
from src.common.error import MyHttpCode
from src.common.error import async_deal_error_and_finish
from src.common.error import MyHttpError
from src.modules.communication_manager import CommunicationManager
from src.utils.net import get_host_for_url


class RedirectHTTPSHandler(web.RequestHandler):
    @async_deal_error_and_finish(message="get access info failed")
    async def get(self,uri):
        as_info = await CommunicationManager().async_get_as_info()
        url = "{}/{}".format(as_info["access_addr"].removesuffix("/"), uri)
        self.set_status(301)
        self.set_header("Location", url)


class AccessAddrHandler(web.RequestHandler):

    @async_deal_error_and_finish(message="change access addr failed")
    async def change_access_addr_exclude_ingress_classes(self):
        if not self.request.body:
            raise MyHttpError(
                code=MyHttpCode.NCT_INVALID_PARAMS_OF_BODY,
                message="Request body is not allowed to be empty",
            )
        data: dict = json.loads(self.request.body)
        if "host" in data:
            try:
                # 1、域名有效 (能解析)
                # 2、IP合法
                if not Nslookup().dns_lookup_all(data["host"]).answer:
                    IP(data["host"])
            except Exception as e:
                raise MyHttpError(
                    code=MyHttpCode.NCT_INVALID_ACCESS_ADDR,
                    message="Host must be a available domain name or a valid IP",
                    cause=e
                )
        if "port" in data:
            try:
                port = int(data["port"])
            except Exception as e:
                raise MyHttpError(
                    code=MyHttpCode.NCT_INVALID_PORT,
                    message="The port must be a number",
                    cause=e
                )
            else:
                if not 1 <= port <= 65536:
                    raise MyHttpError(
                        code=MyHttpCode.NCT_INVALID_PORT,
                        message="The port must be between 0 and 65535"
                    )
        if "scheme" in data:
            if data["scheme"] not in ("http", "https"):
                raise MyHttpError(
                    code=MyHttpCode.NCT_INVALID_ACCESS_ADDR,
                    message="The scheme must be http or https"
                )

        if "path" in data:
            if not data["path"].startswith("/"):
                raise MyHttpError(
                    code=MyHttpCode.NCT_INVALID_ACCESS_ADDR,
                    message="The path must be start with /"
                )


        cli = CommunicationManager()
        access_info = await cli.async_get_access_addr()
        host = data.get("host", access_info["host"])
        port = data.get("port", access_info["port"])
        access_scheme = data.get("scheme", access_info["scheme"])
        access_path = data.get("path", access_info["path"])
        access_type = data.get("type", access_info["type"])

        if access_path != "/" and access_type != "external":
            raise MyHttpError(
                code=MyHttpCode.NCT_INVALID_ACCESS_ADDR,
                message="If a prefix is owned, the access address type must be external"
            )

        access_addr = f"{access_scheme}://{get_host_for_url(host)}:{port}{access_path}"
        _force = "force" in self.request.arguments
        await cli.change_access_addr_not_refresh_ingress_class(access_addr, access_type, force=_force)

    @async_deal_error_and_finish(message="change access addr failed")
    async def put(self):
        await self.change_access_addr_exclude_ingress_classes()
        cli = CommunicationManager()
        cli.upgrade_ingress_class443()


    @async_deal_error_and_finish(message="get access info failed")
    async def get(self):
        self.write(await CommunicationManager().async_get_access_addr())
        self.set_status(200)


            