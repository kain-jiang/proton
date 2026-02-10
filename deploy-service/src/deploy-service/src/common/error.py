from functools import wraps
from typing import Optional, Union

from tornado import web
import requests

from src.common.config import MyHttpCode


class MyHttpError(Exception):
    def __init__(
        self,
        code: MyHttpCode = MyHttpCode.NCT_UNKNOWN_ERROR,
        message: str = "",
        cause: Union[str, Exception] = "",
        extras: Optional[dict] = None,
    ):
        """创建自定义错误

        :param code: 错误码，参考 MyHttpCode
        :param message: 错误信息
        :param cause: 错误原因，可以是Exception
        :param extras: 额外需要传递的参数
        """
        self.code: MyHttpCode = code  # noqa
        self.message: str = message or code.name
        if cause:
            self.cause = str(cause)
        else:
            self.cause = code.name
        self.extras: dict = extras or {}

    def reply(self, h: web.RequestHandler, msg: str = ""):
        """
        :msg str 错误里面如果没有设置message，message和枚举名一致时，会覆盖message
        """
        if self.message == self.code.name and msg:
            self.message = msg

        try:
            http_code = int(self.code.value[:3])
        except Exception:  # noqa
            http_code = 500

        h.set_status(http_code)
        h.write({"code": self.code.value, "message": self.message, "cause": self.cause, **self.extras})


class OtherHttpError(Exception):
    def __init__(
        self,
        resp: requests.Response
    ):
        if resp.ok:
            raise Exception("service internal error, cannot use other http_error")
        self.http_code = resp.status_code
        self.err_info = resp.json()

    def reply(self, h: web.RequestHandler):
        h.set_status(self.http_code)
        h.write(self.err_info)


def deal_error_and_finish(message: str = ""):
    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kw):
            assert len(args) > 0 and isinstance(args[0], web.RequestHandler)
            try:
                return func(*args, **kw)
            except MyHttpError as err:
                err.reply(args[0], message)
                raise
            except Exception as e:
                MyHttpError(code=MyHttpCode.NCT_UNKNOWN_ERROR, message=message, cause=e).reply(args[0])
                raise
            finally:
                args[0].finish()

        return wrapper

    return decorator


def async_deal_error_and_finish(message: str = ""):
    def decorator(func):
        @wraps(func)
        async def wrapper(*args, **kw):
            assert len(args) > 0 and isinstance(args[0], web.RequestHandler)
            try:
                return await func(*args, **kw)
            except MyHttpError as err:
                err.reply(args[0], message)
                raise
            except Exception as e:
                MyHttpError(code=MyHttpCode.NCT_UNKNOWN_ERROR, message=message, cause=e).reply(args[0])
                raise
            finally:
                await args[0].finish()

        return wrapper

    return decorator


__all__ = ["MyHttpError", "MyHttpCode", "deal_error_and_finish", "async_deal_error_and_finish"]
