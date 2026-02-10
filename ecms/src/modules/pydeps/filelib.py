#!/usr/bin/env python
#-*- coding:utf-8 -*-

"""
文件读写操作公共库
"""

import os
from src.modules.pydeps import tracer


@tracer.trace_func
def read_file(path, mode='r', return_type='str'):
    """
    读文件公共方法
    - path  文件路径
    - mode  文件打开方式
    - return_type  读取内容返回类型，默认字符串
    """
    result = return_type
    with open(path, mode) as file_obj:
        if return_type == 'str':
            result = file_obj.read()
        elif return_type == 'list':
            result = file_obj.readlines()
        else:
            errmsg = "unkonw variable type %s" % (return_type)
            raise Exception(errmsg)
    return result


@tracer.trace_func
def write_file(path, content, mode='w'):
    """
    写文件公共方法
    - path  文件路径
    - content  写入内容，类型可为list,str
    - mode  文件打开方式
    """
    directory = path[:path.rindex("/")]
    # 自动创建目录
    if not os.path.exists(directory):
        os.makedirs(directory)
    with open(path, mode) as file_obj:
        if isinstance(content, basestring):
            file_obj.write(content)
        elif isinstance(content, list):
            for i in range(len(content)):
                # 忽略空行
                if len(content[i]) == 0:
                    continue
                # 添加换行符
                if content[i][-1] != "\n":
                    content[i] += "\n"
            file_obj.writelines(content)
                
        else:
            errmsg = "unkonw variable type"
            raise Exception(errmsg)
