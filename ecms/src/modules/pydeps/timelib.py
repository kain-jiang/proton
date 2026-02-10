#!/usr/bin/env python
#-*- coding:utf-8 -*-

"""
本模块提供 time 相关的公共函数/类
"""

import time


class Timeout(object):
    """
    超时检测机制
    """
    def __init__(self, timeout, delay):
        """
        @param int/float timeout      设定超时时间，单位为秒
        @param int/float delay        设定当检测未超时时延迟返回的时间，单位为秒
        """
        self.timeout = timeout
        self.delay = delay
        self.start_time = time.time()
        self.last_time = self.start_time

    def is_timeout(self):
        """
        判断当前时间距本对象创建之时是否已超时
        若超时，则返回 True
        若未超时，则 sleep 指定的delay时长后，返回 False
        """
        now = time.time()
        last_time_passed = now - self.last_time
        start_time_passed = now - self.start_time

        if last_time_passed < 0 or last_time_passed > start_time_passed:
            # 说明系统时间可能发生了跃变，忽略本次超时检测
            pass
        elif start_time_passed >= self.timeout:
            return True

        time.sleep(self.delay)
        return False

    def check(self, err_prefix):
        """
        检测当前时间距本对象创建之时是否已超时
        若超时，则抛出异常
        若未超时，则 sleep 指定的delay时长
        @param string err_prefix     超时异常信息的前缀
        """
        if self.is_timeout():
            raise Exception("{0} timeout.".format(err_prefix))
