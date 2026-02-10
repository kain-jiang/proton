#!/usr/bin/env python
#-*- coding:utf-8 -*-

"""
抛异常的公共库：
1.按EThriftException.thrift定义，抛ncTException异常
2.抛Exception异常

注意：此模块需要依赖EInfoworksLogger.thrift和EThriftException.thrift接口，
因此import此模块的代码，需要将EInfoworksLogger.thrift和EThriftException.thrift
转换为py代码，并包含存在EInfoworksLogger和EThriftException文件夹的目录。
"""

import inspect
import time
import traceback
import sys
import os

# from EThriftException.ttypes import ncTExpType, ncTException

from src.modules.pydeps import tracer
import logger

####################################################################
# 按EThriftException.thrift定义，抛ncTException异常

ncTExpType = {
    "NCT_FATAL": 0,
    "NCT_CRITICAL": 1,
    "NCT_WARN": 2,
    "NCT_INFO": 3,
}


class ncTException(Exception):
    def __init__(self, expType=None, codeLine=None, errID=None, fileName=None, expMsg=None, errProvider=None, time=None,
                errDetail=None):
        self.expType = expType
        self.codeLine = codeLine
        self.errID = errID
        self.fileName = fileName
        self.expMsg = expMsg
        self.errProvider = errProvider
        self.time = time
        self.errDetail = errDetail

    def __str__(self):
        return repr(self)

    def __repr__(self):
        L = ['%s=%r' % (key, value)
             for key, value in self.__dict__.iteritems()]
        return '%s(%s)' % (self.__class__.__name__, ', '.join(L))

    def __eq__(self, other):
        return isinstance(other, self.__class__) and self.__dict__ == other.__dict__

    def __ne__(self, other):
        return not (self == other)


@tracer.trace_func
def raise_e(err_provider,
            err_id,
            err_msg,
            err_type=ncTExpType["NCT_FATAL"],
            exp_detail=''):
    """
    抛ncTException类型异常
    @param err_provider(str)：错误提供者
    @param err_id(int): 错误号
    @param err_msg(str)：错误内容
    @param err_type(ncTExpType):错误类型，默认为 NCT_FATAL
    """
    # 检查参数类型
    check_type("err_provider", err_provider, basestring)
    check_type("err_msg", err_msg, basestring)
    check_type("err_id", err_id, int)
    check_type("err_type", err_type, int)

    # 堆栈回退2级：跳过本函数及tracer.trace_func，取异常抛出的代码位置
    expt_frame = inspect.currentframe().f_back.f_back
    ex = ncTException()
    ex.errProvider = err_provider
    ex.errID = err_id
    ex.expMsg = err_msg
    ex.expType = err_type
    ex.codeLine = expt_frame.f_lineno
    file_name = os.path.basename(expt_frame.f_code.co_filename)
    dir_name = os.path.basename(os.path.dirname(expt_frame.f_code.co_filename))
    ex.fileName = "{0}/{1}".format(dir_name, file_name)
    ex.time = time.ctime()
    ex.errDetail = exp_detail
    raise ex


def equal_ncTException(ex, err_provider, err_id):
    """
    判断是否为指定异常
    @param ex(ncTException): 异常对象
    @param err_provider(str)：错误提供者
    @param err_id(int): 错误号
    @return: 若 ex 与指定err_provider和err_id匹配，则return True，否则return False
    """
    # 检查参数类型
    check_type("ex", ex, ncTException)
    check_type("err_provider", err_provider, basestring)
    check_type("err_id", err_id, int)

    if ex.errProvider == err_provider and ex.errID == err_id:
        return True
    else:
        return False

####################################################################
# 参数类型检查函数


def check_type(name, value, type_obj):
    """
    检查变量类型是否为指定类型.
    @name: 变量名称
    @value: 变量值
    @type_obj: 类型对象
    @raise TypeError with traceback.
    """
    if not isinstance(value, type_obj):
        frame = inspect.currentframe().f_back
        raise TypeError("%s:%d - Type of %s(%s) need be %r."
                        % (frame.f_code.co_filename, frame.f_lineno,
                           name, str(value), type_obj))

####################################################################
# 装饰器


def wrapin_ncTException(err_provider, err_id, err_desc,
                        need_log_args=True, need_log_nctexception=True):
    """
    装饰器：用于捕获非ncTException异常后，记录日志，并封装成 ncTException 异常抛出
    场景: thrift接口使用该装饰器，避免抛出非TException类型异常，
          导致thrift客户端收到Tsocket read 0 Bytes的问题
    @param err_provider(str)：错误提供者
    @param err_id(int): 错误号
    @param err_desc(str)：对异常场景的描述
    @param need_log_args(bool): 记日志时是否包含函数参数
    @param need_log_nctexception(bool): 若捕获的是 ncTException，是否仍需要记日志

    用法：
    @raiser.wrapin_ncTException("provider", 123, "test_func出错。")
    def test_func(param):
        raise Exception("error.")
    """
    def _wrapin_ncTException(func):
        def __wrapin_ncTException(*args, **kwargs):
            try:
                return func(*args, **kwargs)
            except BaseException, ex:
                def _log_expt():
                    """将异常记调试日志"""
                    if need_log_args:
                        desc = "%s %s(%s)" % (err_desc, func.__name__, tracer.args2str(*args, **kwargs))
                    else:
                        desc = "%s %s(...)" % (err_desc, func.__name__)
                    logger.syslog_exception(err_provider, desc, ex, need_traceback=True)

                # 转换成ncTException抛出
                exc_type, exc_value, exc_traceback = sys.exc_info()
                if exc_type == ncTException:
                    if need_log_nctexception:
                        _log_expt()
                    raise
                else:
                    _log_expt()
                    last_tb = traceback.extract_tb(exc_traceback)[-1]
                    file_name = os.path.basename(last_tb[0])
                    dir_name = os.path.basename(os.path.dirname(last_tb[0]))
                    code_file = "{0}/{1}".format(dir_name, file_name)
                    code_line = last_tb[1]
                    code_func = last_tb[2]
                    raise_e(err_provider, err_id, '%s %s:%s (File "%s", line %d, in %s)' % (
                        err_desc, ex.__class__.__name__, str(ex), code_file, code_line, code_func))
        return __wrapin_ncTException

    # 检查参数类型
    check_type("err_provider", err_provider, basestring)
    check_type("err_desc", err_desc, basestring)
    check_type("err_id", err_id, int)
    return _wrapin_ncTException
