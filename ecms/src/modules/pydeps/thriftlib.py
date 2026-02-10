#!/usr/bin/env python
#-*- coding:utf-8 -*-

'''
This is a python thrift library for Eisoo platform.

Client example:

from eisoo.thriftlib import BaseThriftClient
from EInfoworksLogger import ncTEInfoworksLoggerService
from EInfoworksLogger.constants import NCT_INFOWORKS_LOGGER_PORT

client = BaseThriftClient("localhost", NCT_INFOWORKS_LOGGER_PORT, ncTEInfoworksLoggerService)
client.Log_EnumOwners()

PS：注意及时关闭连接。

'''

from thrift.transport.TSocket import TSocket
from thrift.transport.TTransport import TBufferedTransport
from thrift.protocol.TBinaryProtocol import TBinaryProtocol


class BaseThriftClient(object):
    '''Base thrift client lib'''
    def __init__(self, dest_ip, dest_port, interface, timeout_s=0):
        '''
        初始化Thrift连接
        初始化后自动打开
        :param dest_ip Thrift 连接目的IP
        :param dest_port Thrift 连接目的端口
        :param interface Thrift 客户端接口
        :param timeout_s  Socket 超时时间，单位：秒，默认为0，不超时
        '''
        self.socket = TSocket(dest_ip, dest_port)
        if timeout_s:
            self.socket.setTimeout(timeout_s * 1000)
        self.transport = TBufferedTransport(self.socket)
        self.protocol = TBinaryProtocol(self.transport)
        self.client = interface.Client(self.protocol)
        self.transport.open()

    def __getattr__(self, name):
        '''
        本魔法函数，仅在调用本实例不存在的属性、方法时，才会触发
        本魔法函数会将所有本实例不存在的属性、方法重定向到本实例的client属性
        主要用于实例化之后直接调用相应Thrift接口方法
        使用方法请查看文件顶部的Example
        '''
        return getattr(self.client, name)

    def open(self):
        '''打开Thrift连接'''
        self.transport.open()

    def close(self):
        '''关闭Thrift连接'''
        self.transport.close()

    def __del__(self):
        '''将本实例删除后，自动关闭连接'''
        try:
            self.close()
        except Exception:
            pass
