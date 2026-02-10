#!/usr/bin/env python
#-*- coding:utf-8 -*-

"""
本模块实现各类线程锁或进程锁
"""

import os
import fcntl
from contextlib import contextmanager

from src.modules.pydeps import tracer, timelib

LOCK_DIR = "/var/lock/app/"
if not os.path.exists(LOCK_DIR):
    os.makedirs(LOCK_DIR)
LOCK_DIR = os.path.realpath(LOCK_DIR)


@contextmanager
@tracer.trace_func
def mutex_lock(mutex):
    """
    多线程互斥锁
    @param mutex:      锁对象，threading.Lock()
    """
    try:
        mutex.acquire()
        yield
    finally:
        mutex.release()


class FileLock(object):
    """
    文件锁，可用于进程间的互斥访问
    支持 with 语句
    #use as:
    with FileLock("myfile.txt"):
        # work with the file as it is now locked
        print("Lock acquired.")
    """
    def __init__(self, lock_name, timeout, delay, fail_error=""):
        """
        初始化
        @param string name          文件锁名称
        @param int/float timeout    尝试获取锁的超时时间, 单位为秒, 如 10
        @param int/float delay      每次尝试获取锁之间的延时时间, 单位为秒, 如 0.1
        @param string fail_error    请求文件锁失败时的异常描述
        """
        self.is_locked = False
        self.lockfile_path = os.path.join(LOCK_DIR, lock_name + ".lock")
        self.lockfile = None
        self.lock_name = lock_name
        self.timeout = timeout
        self.delay = delay
        if fail_error:
            self.fail_error = fail_error
        else:
            self.fail_error = "Acquire file lock({0}) timeout.".format(self.lock_name)

    @tracer.trace_func
    def acquire(self):
        """
        请求文件锁
        Acquire the lock, if possible. If the lock is in use, it check again
        every `wait` seconds. It does this until it either gets the lock or
        exceeds `timeout` number of seconds, in which case it throws
        an exception.
        """
        if self.lockfile:
            raise Exception("Not allowed call acquired() twice before a matched release() called.")

        try:
            self.lockfile = open(self.lockfile_path, "w")
            timeout = timelib.Timeout(self.timeout, self.delay)
            while True:
                try:
                    # 文件锁
                    fcntl.flock(self.lockfile, fcntl.LOCK_EX | fcntl.LOCK_NB)
                    break
                except Exception, ex:
                    if str(ex).find("Resource temporarily unavailable") == -1:
                        raise

                    if timeout.is_timeout():
                        raise Exception(self.fail_error)
            self.is_locked = True
        except:
            if self.lockfile:
                self.lockfile.close()
                self.lockfile = None
            raise

    @tracer.trace_func
    def release(self):
        """
        释放文件锁
        """
        if self.is_locked:
            fcntl.flock(self.lockfile, fcntl.LOCK_UN)
            self.lockfile.close()
            self.lockfile = None
            self.is_locked = False

    def __enter__(self):
        """
        Activated when used in the with statement.
        Should automatically acquire a lock to be used in the with block.
        """
        if not self.is_locked:
            self.acquire()
        return self

    def __exit__(self, type, value, traceback):
        """
        Activated at the end of the with statement.
        It automatically releases the lock if it isn't locked.
        """
        if self.is_locked:
            self.release()

    def __del__(self):
        """
        Make sure that the FileLock instance doesn't leave a lockfile
        lying around.
        """
        self.release()
