#!/usr/bin/env python
# coding=utf-8

#!/usr/bin/env python
# -*- coding: utf-8 -*-
from tornado import web


class AliveHandler(web.RequestHandler):
    def get(self):
        self.set_status(200)
        self.finish()


class ReadyHandler(web.RequestHandler):
    def get(self):
        self.set_status(200)
        self.finish()
