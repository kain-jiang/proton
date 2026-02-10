#!/usr/bin/env python
# -*- coding: utf-8 -*-
# @Time    : 2021/4/8 9:27
# @Author  : mo.kang<mo.kang@eisoo.com>
# @Site    : 
# @File    : server.py
# @Software: PyCharm
import os
import sys
import logging

import tornado.log
import tornado.httpserver
import tornado.web
import tornado.ioloop
from tornado.options import options

CURR_SCRIPT_PATH = os.path.dirname(os.path.abspath(sys.argv[0]))
if CURR_SCRIPT_PATH.find("/src/") != -1:
    sys.path.append(os.path.realpath(os.path.join(CURR_SCRIPT_PATH, "../")))
sys.path.append(os.path.dirname(CURR_SCRIPT_PATH))

from src.handler.router import ECMS_ROUTER
from src.modules.pydeps import safeconfig


CONF_FILE = "/etc/ecms/ecms.conf"
MODULE_NAME = "ECMS"


class LogFormatter(tornado.log.LogFormatter):
    def __init__(self):
        super(LogFormatter, self).__init__(
            fmt='%(color)s[%(asctime)s %(levelname)s]%(end_color)s %(message)s',
            datefmt='%Y-%m-%d %H:%M:%S'
        )


class Application(tornado.web.Application):

    def __init__(self):
        settings = dict(
            log_stdout=True
        )
        # 设置log输出格式
        try:
            logging = safeconfig.SafeConfig.get(conf_file_path=CONF_FILE, section="ecms", option="log_level")
            options.logging = logging
            log_path = safeconfig.SafeConfig.get(conf_file_path=CONF_FILE,
                                                 section="ecms", option="log_path").replace('"', "")
            options.log_file_prefix = log_path
        except:
            pass
        try:
            options.parse_command_line()
        except IOError:
            # 文件夹不存在则创建
            os.makedirs(os.path.dirname(log_path))
            open(log_path, 'a').close()
            options.parse_command_line()
        super(Application, self).__init__(ECMS_ROUTER, settings)


def main():
    app = Application()
    [i.setFormatter(LogFormatter()) for i in logging.getLogger().handlers]
    ssl_options = None
    tls = False
    try:
        tls = False if safeconfig.SafeConfig.get(conf_file_path=CONF_FILE, section="ecms", option="tls") \
                       not in ["true", "True"] else True
    except:
        pass
    if tls:
        try:
            certfile = safeconfig.SafeConfig.get(conf_file_path=CONF_FILE, section="ecms", option="certfile")
            keyfile = safeconfig.SafeConfig.get(conf_file_path=CONF_FILE, section="ecms", option="keyfile")
        except:
            raise Exception("No certificate file specified")
        ssl_options = {"certfile": certfile, "keyfile": keyfile}
    try:
        port = int(safeconfig.SafeConfig.get(conf_file_path=CONF_FILE, section="ecms", option="listen"))
    except:
        port = 9202
    http_server = tornado.httpserver.HTTPServer(app, ssl_options=ssl_options)
    http_server.listen(port)
    tornado.ioloop.IOLoop.current().start()


if __name__ == '__main__':
    main()