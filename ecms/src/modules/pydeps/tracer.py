#!/usr/bin/env python
#-*- coding:utf-8 -*-

'''
This is a trace library for Eisoo platform.
'''

import sys
import os
import ConfigParser
import logging
import logging.handlers
import traceback
import functools
import inspect

TRACE_CONF_FILE = '/sysvol/apphome/conf/trace.conf'
TRACE_CONF_SECTION = 'trace'
TRACE_CONF_ENABLE = 'enable'
TRACE_CONF_OUTFILE = 'outfile'
TRACE_CONF_INCLUDE_FILENAMES = 'include_filenames'


###########################################################################
# trace 相关函数定义

def trace_func(func):
    '''
    Decorator for trace functions(include member functiions of class).
    注意：@trace_func必须为最底层装饰
    Usage:
        @other_decos
        @tracer.trace_func
        def funcxxx():
            pass
    '''
    @functools.wraps(func)
    def _trace_func_wrapper(*args, **kwargs):
        """func wrapper"""
        if not TRACE_HANDLER.isEnabled():
            return func(*args, **kwargs)

        try:
            codestr = '%s - %s()' \
                % (get_call_code_str(inspect.currentframe().f_back),
                   func.__name__)
            argsstr = args2str(*args, **kwargs)
            if argsstr == '':
                beginstr = '%s ^^no-args^^.' % codestr
            else:
                beginstr = '%s ^^args^^(%s)' % (codestr, argsstr)
            TRACE_HANDLER.log(get_call_code_filename(inspect.currentframe().f_back), beginstr)
        except BaseException:
            traceback.format_exc()

        ret = func(*args, **kwargs)

        try:
            codestr = '%s - %s()' \
                % (get_call_code_str(inspect.currentframe().f_back),
                   func.__name__)
            returnsstr = return2str(ret)
            if returnsstr == '':
                endstr = '%s ^^no-return^^.' % codestr
            else:
                endstr = '%s ^^return^^(%s)' % (codestr, returnsstr)
            TRACE_HANDLER.log(get_call_code_filename(inspect.currentframe().f_back), endstr)
        except BaseException:
            traceback.format_exc()

        return ret
    return _trace_func_wrapper


def trace(msg, *args, **kwargs):
    '''
    function for add trace message.
    Usage:
        def funcxxx():
            tracer.trace('This is a trace for %s', 'xxx')
    '''
    if not TRACE_HANDLER.isEnabled():
        return
    try:
        tracestr = '%s : %s' \
            % (get_call_code_str(inspect.currentframe().f_back), msg)
        TRACE_HANDLER.log(get_call_code_filename(inspect.currentframe().f_back),
                          tracestr, *args, **kwargs)
    except BaseException:
        traceback.format_exc()


class TraceHandler(object):

    '''Trace 管理器'''

    def __init__(self):
        self._enable = False
        self._logger = None
        self._fh = None
        self._include_filenames = []
        self.load_conf()

    def isEnabled(self):
        '''查询是否开启了trace'''
        return self._enable

    def log(self, filename, msg, *args, **kwargs):
        '''记录trace'''
        if self._logger is not None:
            if len(self._include_filenames) == 0 or filename in self._include_filenames:
                self._logger.debug(msg, *args, **kwargs)

    def load_conf(self):
        '''加载trace配置(/sysvol/apphome/conf/trace.conf)'''

        def _syslog(msg):
            try:
                try:
                    cmd = 'echo -ne "%s" | logger -i -t "tracer"' % msg
                    os.system(cmd)
                except:
                    traceback.print_exc()
            except:
                pass

        try:
            self._enable = False
            if self._logger is not None and self._fh is not None:
                # 先移除已有的logger handler
                self._logger.removeHandler(self._fh)
                self._logger = None

            if os.path.exists(TRACE_CONF_FILE):
                cfg = ConfigParser.ConfigParser()
                cfg.read(TRACE_CONF_FILE)
                self._enable = (cfg.get(TRACE_CONF_SECTION,
                                        TRACE_CONF_ENABLE) == str(True))
            else:
                _syslog('Trace config file(%s) does not exist.' % TRACE_CONF_FILE)

            if self._enable:
                filenames_str = cfg.get(TRACE_CONF_SECTION, TRACE_CONF_INCLUDE_FILENAMES)
                self._include_filenames = [name.strip() for name in filenames_str.split(',') \
                                           if name.strip() != ""]

                outfile_path = cfg.get(TRACE_CONF_SECTION, TRACE_CONF_OUTFILE)
                if outfile_path != 'stdout':
                    self._fh = logging.handlers.WatchedFileHandler(
                        outfile_path, 'a')
                else:
                    self._fh = logging.StreamHandler(sys.stdout)
                self._fh.setLevel(logging.DEBUG)
                formatter = logging.Formatter(
                    "%(asctime)s pid:%(process)-5d %(message)s")
                self._fh.setFormatter(formatter)
                self._logger = logging.getLogger('trace')
                self._logger.setLevel(logging.DEBUG)
                self._logger.addHandler(self._fh)
                self._logger.debug(
                    '\n\n\n========== Trace started ==========\n\n\n')
                _syslog('Trace enabled, log at %s' % outfile_path)
        except BaseException:
            self._enable = False
            self._logger = None
            _syslog('Load trace config failed. Err:%s' % (traceback.format_exc()))


##########################################################################
# 代码位置、参数、返回值等字符串转换函数定义


def get_call_code_str(frame):
    '''获取调用函数的字符串描述'''
    if frame is None:
        return 'NO-FRAME'
    filename = os.path.basename(frame.f_code.co_filename)
    return '%s:%d [%s()]' % (filename, frame.f_lineno, frame.f_code.co_name)


def get_call_code_filename(frame):
    '''获取调用函数的函数名称'''
    if frame is None:
        return ''
    filename = os.path.basename(frame.f_code.co_filename)
    return filename


def args2str(*args, **kwargs):
    """args 转换成字符串描述"""
    tmp = ''
    for arg in args:
        if isinstance(arg, basestring):
            argstr = ''.join(arg)
        else:
            argstr = str(arg)
        tmp += '%s, ' % argstr

    for key in kwargs:
        value = kwargs[key]
        if isinstance(value, basestring):
            vstr = ''.join(value)
        else:
            vstr = str(value)
        tmp += '%s=%s, ' % (str(key), vstr)

    if tmp != '':
        return tmp[:-2]
    else:
        return tmp


def return2str(returns):
    """return 值转换成字符串描述"""
    returnsstr = ''
    if returns is None:
        pass
    elif isinstance(returns, basestring):
        returnsstr = ''.join(returns)
    elif isinstance(returns, (list, tuple)):
        for ret in returns:
            if isinstance(ret, basestring):
                retstr = ''.join(ret)
            else:
                retstr = str(ret)
            returnsstr += '%s, ' % retstr
        if returnsstr != '':
            returnsstr = returnsstr[:-2]
    else:
        returnsstr = str(returns)

    return returnsstr

#########################################################################
# 全局变量

# 初始化全局trace管理器
TRACE_HANDLER = TraceHandler()
