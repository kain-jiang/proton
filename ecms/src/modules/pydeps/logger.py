#!/usr/bin/env python
# -*- coding:utf-8 -*-

"""
记日志的公共库

注意：此模块需要依赖EThriftException.thrift接口，
因此import此模块的代码，需要将EThriftException.thrift转换为py代码，
并包含存在EThriftException文件夹的目录。
"""

import os
import traceback
import syslog as usyslog

from src.modules.pydeps.raiser import ncTException
from src.modules.pydeps import tracer


####################################################################
# 记本地日志的函数
@tracer.trace_func
def syslog(owner, msg):
    """
    记本地日志
    @param owner(str)：日志所属模块/日志来源
    @param msg(str)：日志内容
    @no-raise
    """
    os_log(owner, usyslog.LOG_INFO, msg)


@tracer.trace_func
def syslog_emerg(owner, msg):
    """
    紧急：系统不可用
    @param owner(str)：日志所属模块/日志来源
    @param msg(str)：日志内容
    @no-raise
    """
    os_log(owner, usyslog.LOG_EMERG, msg)


@tracer.trace_func
def syslog_alert(owner, msg):
    """
    警报：必须马上采取行动的事件
    @param owner(str)：日志所属模块/日志来源
    @param msg(str)：日志内容
    @no-raise
    """
    os_log(owner, usyslog.LOG_ALERT, msg)


@tracer.trace_func
def syslog_crit(owner, msg):
    """
    重要：关键的事件,临界状态
    @param owner(str)：日志所属模块/日志来源
    @param msg(str)：日志内容
    @no-raise
    """
    os_log(owner, usyslog.LOG_CRIT, msg)


@tracer.trace_func
def syslog_error(owner, msg):
    """
    错误：错误的事件
    @param owner(str)：日志所属模块/日志来源
    @param msg(str)：日志内容
    @no-raise
    """
    os_log(owner, usyslog.LOG_ERR, msg)


@tracer.trace_func
def syslog_warn(owner, msg):
    """
    警告：预警的事件
    @param owner(str)：日志所属模块/日志来源
    @param msg(str)：日志内容
    @no-raise
    """
    os_log(owner, usyslog.LOG_WARNING, msg)


@tracer.trace_func
def syslog_notice(owner, msg):
    """
    提醒：普通但重要的事件
    @param owner(str)：日志所属模块/日志来源
    @param msg(str)：日志内容
    @no-raise
    """
    os_log(owner, usyslog.LOG_NOTICE, msg)


@tracer.trace_func
def syslog_info(owner, msg):
    """
    信息：有用的信息
    @param owner(str)：日志所属模块/日志来源
    @param msg(str)：日志内容
    @no-raise
    """
    os_log(owner, usyslog.LOG_INFO, msg)


@tracer.trace_func
def syslog_debug(owner, msg):
    """
    调试：调试的信息
    @param owner(str)：日志所属模块/日志来源
    @param msg(str)：日志内容
    @no-raise
    """
    os_log(owner, usyslog.LOG_DEBUG, msg)


@tracer.trace_func
def syslog_exception(owner, desc, ex, need_traceback=False):
    """
    将异常记本地日志
    @param owner(str)：日志所属模块/日志来源
    @param desc(str)：异常发生的场景描述
    @param ex(BaseException及其派生类)：异常对象
    @param need_traceback(bool): 是否需要堆栈
    @no-raise
    """

    try:
        if isinstance(ex, ncTException):
            msg = "%s ncTException(msg=%s, provider=%s, errID=%s, type=%s, file=%s, line=%s, time=%s)."\
                % (desc, ex.expMsg, ex.errProvider, str(ex.errID), str(ex.expType), ex.fileName,
                   str(ex.codeLine), str(ex.time))
        else:
            msg = "%s %s:%s" % (desc, ex.__class__.__name__, str(ex))

        # 打印日志
        os_log(owner, usyslog.LOG_ERR, msg)

        # 打印堆栈
        if need_traceback:
            # 确认堆栈有信息则打印
            traceback_str = str(traceback.format_exc())
            if traceback_str != '':
                msg_list = traceback_str.split("\n")
                # 移除空字符串
                if '' in msg_list:
                    msg_list.remove('')
                for msg in msg_list:
                    os_log_only(owner, usyslog.LOG_ERR, msg)

    except BaseException:
        traceback.print_exc()


@tracer.trace_func
def syslog_cmd(owner, cmd, outmsg, errmsg, returncode=0):
    """
    将指定命令及其输出记本地日志
    @param owner(str)：日志所属模块/日志来源
    @param cmd(str)：命令行
    @param outmsg(str)：命令输出到stdout的信息
    @param errmsg(str): 命令输出到stderr的信息
    @param returncode(int): 命令执行的返回码
    @no-raise
    """
    try:
        msg = "[CMD] %s (RET: %s)" % (cmd, returncode)
        os_log(owner, usyslog.LOG_DEBUG, msg)

        if outmsg:
            msg_list = str(outmsg).split("\n")
            # 移除空字符串
            if '' in msg_list:
                msg_list.remove('')
            msg_list[0] = "[STDOUT] " + msg_list[0]
            for msg in msg_list:
                os_log_only(owner, usyslog.LOG_DEBUG, msg)
        if errmsg:
            msg_list = str(errmsg).split("\n")
            # 移除空字符串
            if '' in msg_list:
                msg_list.remove('')
            msg_list[0] = "[STDERR] " + msg_list[0]
            for msg in msg_list:
                os_log_only(owner, usyslog.LOG_DEBUG, msg)

    except BaseException:
        traceback.print_exc()


####################################################################
# 保留原记日志的函数，以便代码兼容


def log_info(owner, msg):
    """
    记信息级别日志
    @param owner(str)：日志所属模块/日志来源
    @param msg(str)：日志内容
    @no-raise
    """
    syslog_info(owner, msg)


def log_warn(owner, msg):
    """
    记警告级别日志
    @param owner(str)：日志所属模块/日志来源
    @param msg(str)：日志内容
    @no-raise
    """
    syslog_warn(owner, msg)


def log_error(owner, msg):
    """
    记错误级别日志
    @param owner(str)：日志所属模块/日志来源
    @param msg(str)：日志内容
    @no-raise
    """
    syslog_error(owner, msg)


def log_operation(owner, msg):
    """
    记操作级别日志
    @param owner(str)：日志所属模块/日志来源
    @param msg(str)：日志内容
    @no-raise
    """
    syslog_info(owner, msg)


def log_debug(owner, msg):
    """
    记调试级别日志
    @param owner(str)：日志所属模块/日志来源
    @param msg(str)：日志内容
    @no-raise
    """
    syslog_debug(owner, msg)


def log_exception(owner, desc, ex, need_traceback=False):
    """
    将异常记日志（LT_DEBUG类型）
    @param owner(str)：日志所属模块/日志来源
    @param desc(str)：异常发生的场景描述
    @param ex(BaseException及其派生类)：异常对象
    @param need_traceback(bool): 是否需要堆栈
    @no-raise
    """
    syslog_exception(owner, desc, ex, need_traceback)


def log_cmd(owner, cmd, outmsg, errmsg, returncode=0):
    """
    将指定命令及其输出记录日志（LT_DEBUG类型）
    @param owner(str)：日志所属模块/日志来源
    @param cmd(str)：命令行
    @param outmsg(str)：命令输出到stdout的信息
    @param errmsg(str): 命令输出到stderr的信息
    @param returncode(int): 命令执行的返回码
    @no-raise
    """
    syslog_cmd(owner, cmd, outmsg, errmsg, returncode)


####################################################################
# 记本地操作系统日志的函数
# 日志文件： /var/log/app/app.log

@tracer.trace_func
def os_log(owner, level, msg):
    """
    记本地操作系统日志
    @param owner(str)：日志所属模块/日志来源
    @param level(str): 日志级别
    @param msg(str)：日志内容
    @no-raise
    """
    # 计算字符串
    level_dict = {
        usyslog.LOG_EMERG: 'EMERG', usyslog.LOG_ALERT: 'ALERT', usyslog.LOG_CRIT: 'CRIT',
        usyslog.LOG_ERR: 'ERROR', usyslog.LOG_WARNING: 'WARNING', usyslog.LOG_NOTICE: 'NOTICE',
        usyslog.LOG_INFO: 'INFO', usyslog.LOG_DEBUG: 'DEBUG'}

    for msg_line in str(msg).split("\n"):
        try:
            msg_str = "[%s][%s] %s" % (os.getpid(), level_dict[level], msg_line)
            # 单次日志大小超过 1024 会导致记录不全，因此将日志切分记录
            # 该限制来源于syslog协议,最大允许1K
            step = 1024
            usyslog.openlog(str(owner), 0, usyslog.LOG_USER)
            for i in xrange(0, len(msg_str), step):
                usyslog.syslog(level, msg_str[i:i + step])
        except BaseException:
            traceback.print_exc()
        finally:
            usyslog.closelog()


@tracer.trace_func
def os_log_only(owner, level, msg):
    """
    单输出日志，日志消息体不记录模块
    @param owner(str)：日志所属模块/日志来源
    @param level(str): 日志级别
    @param msg(str)：日志内容
    @no-raise
    """
    try:
        # 单次日志大小超过 1024 会导致记录不全，因此将日志切分记录
        # 该限制来源于syslog协议,最大允许1K
        step = 1024
        usyslog.openlog(str(owner), 0, usyslog.LOG_USER)
        for i in xrange(0, len(msg), step):
            usyslog.syslog(level, msg[i:i + step])
    except BaseException:
        traceback.print_exc()
    finally:
        usyslog.closelog()

####################################################################
# 转换日志字符串描述


def exception_to_msg(desc, ex, need_traceback=False):
    """
    将异常对象转换为异常字符串描述
    @param desc(str)：异常发生的场景描述
    @param ex(BaseException及其派生类)：异常对象
    @param need_traceback(bool): 是否需要堆栈
    @return msg(str)：完整异常字符串描述
    """
    if isinstance(ex, ncTException):
        msg = "%s ncTException(msg=%s, provider=%s, errID=%s, type=%s, file=%s, line=%s, time=%s)."\
            % (desc, ex.expMsg, ex.errProvider, str(ex.errID), str(ex.expType), ex.fileName,
               str(ex.codeLine), str(ex.time))
    else:
        msg = "%s %s:%s" % (desc, ex.__class__.__name__, str(ex))

    if need_traceback:
        msg += '\n%s' % traceback.format_exc()

    return msg


def cmd_to_msg(cmd, outmsg, errmsg, returncode):
    """将命令及其输出信息组合成日志内容"""
    msg = "[CMD] %s (RET: %s)" % (cmd, returncode)
    if outmsg:
        msg += "\n[STDOUT] %s" % (outmsg)
    if errmsg:
        msg += "\n[STDERR] %s" % (errmsg)
    return msg


####################################################################
# 记本地操作系统日志的函数
# 日志文件： /var/log/app/app.log

@tracer.trace_func
def upgrade_log(owner, level, msg):
    """
    记本地服务端升级日志 local4 (/var/log/upgrade.log)
    @param owner(str)：日志所属模块/日志来源
    @param level(str): 日志级别
    @param msg(str)：日志内容
    @no-raise
    """
    # 计算字符串
    level_dict = {
        usyslog.LOG_EMERG: 'EMERG', usyslog.LOG_ALERT: 'ALERT', usyslog.LOG_CRIT: 'CRIT',
        usyslog.LOG_ERR: 'ERROR', usyslog.LOG_WARNING: 'WARNING', usyslog.LOG_NOTICE: 'NOTICE',
        usyslog.LOG_INFO: 'INFO', usyslog.LOG_DEBUG: 'DEBUG'}

    try:
        msg_str = "[%s][%s] %s" % (os.getpid(), level_dict[level], msg)
        # 单次日志大小超过 1024 会导致记录不全，因此将日志切分记录
        # 该限制来源于syslog协议,最大允许1K
        step = 1024
        for i in xrange(0, len(msg_str), step):
            usyslog.openlog(str(owner), 0, usyslog.LOG_LOCAL4)
            usyslog.syslog(level, msg_str[i:i + step])
    except BaseException:
        traceback.print_exc()
