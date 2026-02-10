#!/usr/bin/env python
# -*- coding:utf-8 -*-

'''
提供简易封装的命令执行的公共函数
'''
import subprocess
import re
from threading import Timer

import tracer, logger


CMDPROCESS_OWNER = "cmdprocess"


@tracer.trace_func
def shell_cmd(cmdstr, timeout_seconds=600):
    """
    /**
     * 执行命令，并等待返回结果
     * 命令执行出错，则抛出异常
     * 注意：执行时间较长或输出信息较多的命令，不建议使用此函数.
     *
     * @param cmdstr(str)               命令的完整字符串
     * @param timeout_seconds(int)      命令执行超时秒数设置
     *                                  若命令执行超过该秒数，则终止运行，并抛出超时异常
     *                                  不超时，请设置为None
     * @return(tuple(str, str))         返回命令执行的结果(outmsg, errmsg)
     * @raise(Exception)                命令执行失败将抛出异常
     */
    """
    (returncode, outmsg, errmsg) = shell_cmd_not_raise(cmdstr, timeout_seconds)
    if returncode != 0:
        error = "Run cmd failed.(CMD:%s)(RET:%s)(STDERR:%s)(STDOUT:%s)." % (
            cmdstr, returncode, errmsg, outmsg)
        logger.syslog(CMDPROCESS_OWNER, error)
        raise Exception(error)

    # 返回命令输出
    return (outmsg, errmsg)


@tracer.trace_func
def shell_cmd_dict(cmdstr, timeout_seconds=600):
    """
    /**
     * 执行命令，并等待返回结果
     * 命令执行出错，则抛出异常
     * 注意：执行时间较长或输出信息较多的命令，不建议使用此函数.
     *
     * @param cmdstr(str)               命令的完整字符串
     * @param timeout_seconds(int)      命令执行超时秒数设置
     *                                  若命令执行超过该秒数，则终止运行，并抛出超时异常
     *                                  不超时，请设置为None
     * @return dict                     返回命令执行的结果 {"outmsg": str, "errmsg": str}
     * @raise(Exception)                命令执行失败将抛出异常
     */
    """
    cmd_ret = {}
    cmd_ret["outmsg"], cmd_ret["errmsg"] = shell_cmd(cmdstr, timeout_seconds)
    return cmd_ret


@tracer.trace_func
def shell_cmd_not_raise(cmdstr, timeout_seconds=600):
    """
    /**
     * 执行命令，并等待返回结果
     * 命令执行出错，不会抛出异常，而是将 exit code 作为结果返回，供调用者自行判断。
     * 注意：执行时间较长或输出信息较多的命令，不建议使用此函数.
     *
     * @param cmdstr(str)               命令的完整字符串
     * @param timeout_seconds(int)      命令执行超时秒数设置
     *                                  若命令执行超过该秒数，则终止运行，并抛出超时异常
     *                                  未设置，则默认超时时间为10分钟
     * @return(tuple(int, str, str))    返回命令执行的结果(returncode, outmsg, errmsg)
     */
    """
    try:
        # 检查命令字符串是否包含不安全字符
        unsafe_chars = set('`#$|<>;&\'"\
')
        if isinstance(cmdstr, str) and any(c in cmdstr for c in unsafe_chars):
            # 如果命令包含shell特殊字符，使用shell=True执行，但先进行安全检查
            if check_param_security_for_shell.check(cmdstr):
                error = "Unsafe command with special characters: %s" % cmdstr
                logger.syslog(CMDPROCESS_OWNER, error)
                raise Exception(error)
            proc = subprocess.Popen(cmdstr,
                                   shell=True,
                                   stdout=subprocess.PIPE,
                                   stderr=subprocess.PIPE,
                                   close_fds=True)
        else:
            # 如果是简单命令，拆分为参数列表并使用shell=False执行
            if isinstance(cmdstr, str) or isinstance(cmdstr, unicode):
                cmd_args = cmdstr.split()
            else:
                cmd_args = cmdstr  # 如果已经是列表，直接使用

            proc = subprocess.Popen(cmd_args,
                                   shell=False,
                                   stdout=subprocess.PIPE,
                                   stderr=subprocess.PIPE,
                                   close_fds=True)
    except Exception as ex:
        error = "Run cmd failed.(CMD:%s)(ERROR:%s)" % (cmdstr, ex)
        logger.syslog(CMDPROCESS_OWNER, error)
        raise Exception(error)

    my_timer = Timer(timeout_seconds, proc.kill)
    my_timer.start()
    try:
        # 读取命令的输出，并等待命令执行结束
        outmsg, errmsg = proc.communicate(input=None)
        if my_timer.is_alive():
            return (proc.returncode, outmsg, errmsg)

        # 命令执行超时
        error = "Run cmd failed, {0} seconds timeout.(CMD:{1}).".format(timeout_seconds, cmdstr)
        logger.syslog(CMDPROCESS_OWNER, error)
        raise Exception(error)
    finally:
        my_timer.cancel()
        if proc:
            if proc.stdin:
                proc.stdin.close()
            if proc.stdout:
                proc.stdout.close()
            if proc.stderr:
                proc.stderr.close()


@tracer.trace_func
def shell_cmd_async(cmdstr):
    """
    /**
     * 执行命令，不等待返回结果
     *
     * @param cmdstr(str)               命令的完整字符串
     */
    """
    try:
        shell_cmd("{0} >/dev/null 2>&1 &".format(cmdstr))
    except Exception as ex:
        error = "Run cmd asynchronously failed.(CMD:%s)(ERROR:%s)" % (cmdstr, ex)
        logger.syslog(CMDPROCESS_OWNER, error)
        raise Exception(error)


def output_to_lines(outmsg):
    """
    将命令输出分解成多行，且去除空白行
    @return lines(list<str>)
    """
    return [line for line in re.split(r"[\r\n]+", outmsg) if line[:-1].strip()]


def raise_cmd_fail(cmdstr, returncode, outmsg, errmsg):
    """
    将命令执行失败封装成异常抛出
    """
    raise Exception("Run cmd failed.(CMD:%s)(RET:%s)(STDERR:%s)(STDOUT:%s)." % (
        cmdstr, returncode, errmsg, outmsg))


def check_param_security_for_shell(func):
    """
    检查参数是否能安全地用于 Shell
    如果参数中包括 `#$|<>;& 则抛异常
    使用方法:
        @cmdprocess.check_param_security
        def foo(a, b):
            ...
    """
    unsafe_chars = set('`#$|<>;&\'"\n')

    def check(parameter):
        """ 检查参数是否安全 """
        if isinstance(parameter, (str, unicode)):
            return unsafe_chars & set(parameter)
        if isinstance(parameter, (list, tuple, set)):
            return any(check(each) for each in parameter)
        if isinstance(parameter, dict):
            return any(check(key) or check(val) for key, val in parameter.items())
        try:
            return check(parameter.__dict__)
        except AttributeError:
            return False

    def wrapper(*args, **kwargs):
        """装饰器"""
        unsafe_params = [each for each in list(args) + kwargs.values() if check(each)]
        if unsafe_params:
            message = 'Unsafe parameter(s): {params} include {chars!r}'.format(
                params=unsafe_params,
                chars=''.join(unsafe_chars),
            )
            raise Exception(message)
        return func(*args, **kwargs)
    return wrapper
