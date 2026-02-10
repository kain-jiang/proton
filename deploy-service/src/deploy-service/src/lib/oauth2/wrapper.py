## region oauth2 装饰器
import requests
from tornado import web

from src.common import utils, config
from src.common.log_util import logger

def __analysis_token(req_handler: web.RequestHandler) -> str:
    auth_header: str = req_handler.request.headers.get('Authorization', "")
    token: str = ""
    if auth_header:
        prefix = auth_header.split(" ")[0].lower()
        if prefix == "bearer":
            token = auth_header.split(' ', 1)[1]
    if not token:
        token = req_handler.get_cookie("deploy.oauth2_token", "") # deploy-web目前使用oauth2的方式
    return token

def __introspect_token(token) -> dict:
    try:
        resp: requests.Response = requests.post(
            url=f"http://{utils.get_hydra_admin_by_cache()}/admin/oauth2/introspect",
            data={"token": token}
        )
        response_json: dict = resp.json()
        logger.info(f"introspect token result: {resp.text}, token: {token}")
    except Exception as e:
        logger.error(f"introspect token failed: {str(e)}")
        return {"active": False}
    else:
        return response_json


def check_is_login(func):
    def wrapper(self, *args, **kwargs):
        self: web.RequestHandler
        token: str = __analysis_token(self)
        active: bool = __introspect_token(token).get("active", False)
        if active:
            return func(self, *args, **kwargs)
        else:
            self.set_status(401)
            self.write({
                "code": config.MyHttpCode.NCT_UNAUTHORIZED.value,
                "cause": config.MyHttpCode.NCT_UNAUTHORIZED.name,
                "message": "token introspect failed."
            })
            self.finish()
    return wrapper

def aysnc_check_is_login(func):
    async def wrapper(self, *args, **kwargs):
        self: web.RequestHandler
        token: str = __analysis_token(self)
        active: bool = __introspect_token(token).get("active", False)
        if active:
            return await func(self, *args, **kwargs)
        else:
            self.set_status(401)
            self.write({
                "code": config.MyHttpCode.NCT_UNAUTHORIZED.value,
                "cause": config.MyHttpCode.NCT_UNAUTHORIZED.name,
                "message": "token introspect failed."
            })
            await self.finish()
    return wrapper

## endregion