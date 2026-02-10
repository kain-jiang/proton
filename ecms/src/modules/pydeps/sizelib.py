#!/usr/bin/env python
#-*- coding:utf-8 -*-

'''
提供简易封装的size转换的公共函数
'''

import math
import re

# 支持的size单位
SIZE_UNIT_LIST = [['Bytes', 'B'],
                  ['KB', 'K'],
                  ['MB', 'M'],
                  ['GB', 'G'],
                  ['TB', 'T'],
                  ['PB', 'P']]


def convert_size(size, ori_unit, dest_unit):
    """
    将 size 转换为期望的单位
    @param size(int/long/float):以 ori_unit 为单位的大小
    @param ori_unit(str):       size的原始单位，如 'Bytes', 'KB', 'MB', 'GB', 'TB', 'PB'
    @param dest_unit(str):      size的目标单位，如 'Bytes', 'KB', 'MB', 'GB', 'TB', 'PB'
    @return dest_size(float):   以 dest_unit 为单位的大小
    """
    ori_unit_index = -1
    for i, units in enumerate(SIZE_UNIT_LIST):
        if ori_unit in units:
            ori_unit_index = i
            break

    dest_unit_index = -1
    for i, units in enumerate(SIZE_UNIT_LIST):
        if dest_unit in units:
            dest_unit_index = i
            break

    if ori_unit_index == -1:
        raise Exception("Size unit %s does not support." % ori_unit)
    if dest_unit_index == -1:
        raise Exception("Size unit %s does not support." % dest_unit)

    return size * math.pow(1024, ori_unit_index - dest_unit_index)


def bytes_to_string(bytes):
    """
    将 Bytes 转换可读字符串，取最大的单位
    @param bytes(long):     字节数
    @return size(str)
    """
    # 计算最大的单位
    # 求对数(对数：若 a**b = N 则 b 叫做以 a 为底 N 的对数)，舍弃小数点，取小
    unit_i = int(math.floor(math.log(bytes, 1024)))
    if unit_i >= len(SIZE_UNIT_LIST):
        unit_i = len(SIZE_UNIT_LIST) - 1
    size = bytes / math.pow(1024, unit_i)
    return "%.3f" % size + " " + SIZE_UNIT_LIST[unit_i][0]


def parse_size_string(size_str):
    """
    解析 size 的可读字符串
    @param size_str(str):               size 的可读字符串，如 999.123 GB 或 100MB 等
    @return (size(float), unit(str)):   (size数值，单位)
    """
    reobj = re.search(r"^\s*(.*\d+)\s*((?:[a-z]|[A-Z])+)\s*$", size_str)
    if not reobj:
        if size_str == "0":
            return (0, "Bytes")
        raise Exception("Can not parse %s." % size_str)
    size = float(reobj.group(1))
    unit = reobj.group(2)

    unit_index = -1
    for i, units in enumerate(SIZE_UNIT_LIST):
        if unit in units:
            unit_index = i
            break
    if unit_index == -1:
        raise Exception("Size unit %s does not support." % unit)

    return (size, unit)
