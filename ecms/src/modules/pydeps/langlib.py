#!/usr/bin/env python
#-*- coding:utf-8 -*-

"""
国际化语言资源配置的公共库
1.初始化加载本地语言资源
2.管理国际化配置文件/sysvol/conf/language.conf

注意：此模块需要依赖EInfoworksLogger.thrift和EThriftException.thrift接口，
因此import此模块的代码，需要将EInfoworksLogger.thrift和EThriftException.thrift
转换为py代码，并包含存在EInfoworksLogger和EThriftException文件夹的目录。
"""

import os
import gettext

from src.modules.pydeps import logger

LANG_CONFIG_FILE = "/sysvol/conf/language.conf"
LANG_CONFIG_OPTION = "LANG"
LANG_DEFAULT = "zh_CN"
LANGLIB_OWNER = "langlib"


def init_language(res_name, local_path=".."):
    """
    初始化加载本地语言资源
    场景：在服务启动时调用即可。
    - 若全球化语言未配置，默认加载"zh_CN"
    - 资源文件目录结构：local/zh_CN/LC_MESSAGES/xxx.mo
    - 资源文件请使用utf8编码
    @param res_name(str): 资源文件名，如ENMC.mo，请传"ENMC"
    @param local_path(str):指定local目录所在路径，默认为上级目录
    """
    try:
        # 获取全球化配置
        langstr = get_lang()
        if not langstr:
            langstr = LANG_DEFAULT

        # 获取语言资源路径：从当前程序运行目录下查找local目录
        local_path = os.path.realpath(os.path.join(local_path, "local"))

        logger.syslog(LANGLIB_OWNER,
                      "init_language() %s %s %s.mo" % (local_path,
                                                       langstr,
                                                       res_name))

        # 装载语言资源
        gettext.install(res_name, local_path, unicode="True")
        gettext.translation(res_name, local_path, languages=[langstr]).install()
    except Exception, ex:
        logger.log_exception(LANGLIB_OWNER, "Init language failed.", ex)
        raise


def get_lang():
    """
    获取全球化语言配置
    - 从 /sysvol/conf/language.conf 中读取 LANG 字段的值
    - 配置文件不存在或LANG字段未设置，则默认返回"zh_CN"
    - 若配置多个LANG字段，则以第一个为准
    @return str：语言，如"zh_CN"等
    """
    langstr = LANG_DEFAULT
    if os.path.exists(LANG_CONFIG_FILE):
        with open(LANG_CONFIG_FILE) as fo:
            for line in fo.readlines():
                # 去除换行、中间空格等，再按=分割
                vec = line.strip(os.linesep).replace(' ', '').split('=')
                if len(vec) == 2:
                    if vec[0] == LANG_CONFIG_OPTION:
                        langstr = vec[1]
                        break
    return langstr


def set_lang(langstr):
    """
    修改全球化语言配置
    @param langstr(str): 语言，如zh_CN、en_US、zh_TW
    """
    conf_path = os.path.dirname(LANG_CONFIG_FILE)
    if not os.path.exists(conf_path):
        os.mkdir(conf_path)
    with open(LANG_CONFIG_FILE, mode='w') as fo:
        fo.write("%s=%s" % (LANG_CONFIG_OPTION, langstr))
