#!/usr/bin/env python
# -*- coding: utf-8 -*-
# @Time : 2021/7/30 16:24
# @Author : Kain.jiang@aishu.cn
# @File : wrapper.py
# @Software: PyCharm
import multiprocessing
import time

from tornado import web

from src.common.log_util import logger


def retry_by_exception(attempt=3, sleep_time=5):
    """异常重试的方法装饰器"""

    def decorator(func):
        def wrapper(*args, **kw):
            att = 0
            while att < attempt:
                try:
                    return func(*args, **kw)
                except Exception as e:
                    att = att + 1
                    logger.debug(
                        f"DeployService: function {func.__name__} produce exception [{str(e)}], retry {att} ..."
                    )
                    time.sleep(sleep_time)
            return func(*args, **kw)

        return wrapper

    return decorator


def retry_by_retvalue(attempt=3, sleep_time=5, ret=None):
    """根据返回值进行重试的类实例方法装饰器"""

    def decorator(func):
        def wrapper(self, *args, **kw):
            att = 0
            while att < attempt:
                rel = func(self, *args, **kw)
                if rel == ret:
                    att = att + 1
                    logger.debug(f"DeployService: function {func.__name__} return value equal {ret}, retry {att} ...")
                    time.sleep(sleep_time)
                else:
                    return rel
            return func(self, *args, **kw)

        return wrapper

    return decorator


def lock_handler(lock: multiprocessing.Lock):
    """锁， 只能被 `web.RequestHandler` 实例使用"""

    def decorator(func):
        def wrapper(*args, **kw):
            assert len(args) > 0 and isinstance(args[0], web.RequestHandler)
            if not lock.acquire(block=False):
                args[0].write(
                    {
                        "code": "423017002",
                        "message": "Refuse to process duplicate requests",
                        "cause": "NCT_FORBIDDEN_DUPLICATE_REQUEST",
                    }
                )
                args[0].set_status(423)
                args[0].finish()
                logger.debug(f"DeployService: {str(func)} is running ...")
                return
            logger.debug(f"DeployService: {str(func)} start lock ...")
            try:
                return func(*args, **kw)
            finally:
                lock.release()
                logger.debug(f"DeployService: {str(func)} lock release ...")

        return wrapper

    return decorator


def async_lock_handler(lock: multiprocessing.Lock):
    """锁， 只能被 `web.RequestHandler` 实例的 asnyc 处理器使用"""

    def decorator(func):
        async def wrapper(*args, **kw):
            assert len(args) > 0 and isinstance(args[0], web.RequestHandler)
            if not lock.acquire(block=False):
                args[0].write(
                    {
                        "code": "423017002",
                        "message": "Refuse to process duplicate requests",
                        "cause": "NCT_FORBIDDEN_DUPLICATE_REQUEST",
                    }
                )
                args[0].set_status(423)
                args[0].finish()
                logger.debug(f"DeployService: {str(func)} is running ...")
                return
            logger.debug(f"DeployService: {str(func)} start lock ...")
            try:
                return await func(*args, **kw)
            finally:
                lock.release()
                logger.debug(f"DeployService: {str(func)} lock release ...")

        return wrapper

    return decorator


def catch_exception(message="an exception occurred in the service"):
    def decorator(func):
        def wrapper(*args, **kwargs):
            assert len(args) > 0 and isinstance(args[0], web.RequestHandler)
            self: web.RequestHandler = args[0]
            try:
                func(*args, **kwargs)
            except Exception as e:
                self.write({
                    "code": "500017000",
                    "message": message,
                    "cause": str(e)
                })
                self.set_status(500)
                self.finish()

        return wrapper

    return decorator
