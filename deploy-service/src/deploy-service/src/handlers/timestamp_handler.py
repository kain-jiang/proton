#!/usr/bin/env python
# -*- coding: utf-8 -*-
from tornado import web
import tornado.concurrent, tornado.gen
from concurrent.futures import ThreadPoolExecutor
import time



class TimestampHandler(web.RequestHandler):
    executor = ThreadPoolExecutor(100, "proton-timestamp")
    
    @tornado.gen.coroutine
    def get(self):
        """
        迁移proton-openapi接口,返回当前时间戳。
        所有使用该接口的服务都应该考虑重构
        TODO: 等待废弃移除
        """
        try:
            t  = yield self.time()
            self.write(t)
            self.set_status(200)
        finally:
            self.finish()
    
    @tornado.concurrent.run_on_executor
    def time(self) ->dict:
        return {"apiVersion":"v1alpha1","kind":"Timestamp","name":"timestamp","data":int(time.time()),"description":"服务器当前时间","status":""}