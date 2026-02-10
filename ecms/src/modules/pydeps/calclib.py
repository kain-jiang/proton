#!/usr/bin/env python
#-*- coding:utf-8 -*-

'''
提供简易封装的计算相关的公共函数
'''

import hashlib


def calc_md5(buf):
    """
    计算指定buf的md5值
    @param string buf
    @return string md5
    """
    md5 = ""
    if buf:
        hash_obj = hashlib.md5()
        hash_obj.update(buf)
        md5 = hash_obj.hexdigest().lower()
    return md5


def calc_file_md5(path):
    """
    计算指定文件的md5值
    @param string path 文件路径
    @return string md5
    """
    with open(path, "r") as fobj:
        return calc_md5(fobj.read())
