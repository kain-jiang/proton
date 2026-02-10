#!/usr/bin/env python
# -*- coding: utf-8 -*-
# @Time    : 2021/4/19 10:54
# @Author  : mo.kang<mo.kang@eisoo.com>
# @Site    : 
# @File    : file_agent.py
# @Software: PyCharm
import os
import shutil
import subprocess

from src.modules.pydeps import cmdprocess, tracer

CRT_FILE = "/usr/local/slb-nginx/ssl/eceph-server.crt"
BAK_CRT_FILE = "/usr/local/slb-nginx/ssl/eceph-server.crt.bak"
KEY_FILE = "/usr/local/slb-nginx/ssl/eceph-server.key"
BAK_KEY_FILE = "/usr/local/slb-nginx/ssl/eceph-server.key.bak"
MODULE_NAME = 'FileAgent'
SLB_NGINX_SBIN = "/usr/local/slb-nginx/sbin/slb-nginx"


class FileAgent(object):
    """
    文件和文件夹管理
    """
    @classmethod
    def directory_exists(cls, directory):
        return os.path.exists(directory)

    @classmethod
    def create_directory(cls, directory, mode=0o755):
        if not os.path.exists(directory):
            os.makedirs(directory)
        os.chmod(directory, mode)

    @classmethod
    def delete_directory(cls, directory):
        if os.path.exists(directory):
            shutil.rmtree(directory)

    @classmethod
    def creat_file(cls, file_name, mode=644):
        if not os.path.isfile(file_name):
            open(file_name, "w").close()
        os.chmod(file_name, mode)

    @classmethod
    def file_exists(cls, file_name):
        return os.path.isfile(file_name)

    @classmethod
    def delete_file(cls, file_name):
        if os.path.isfile(file_name):
             os.remove(file_name)

    @classmethod
    def update_tls(cls, content):
        if not cls.file_exists(KEY_FILE) or not cls.file_exists(CRT_FILE):
            return -1

        content_lines = [i.strip() for i in content.split("\n")]
        length = len(content_lines)
        start = -1
        end = -1
        for i in range(length-1):
            if content_lines[i] == "-----BEGIN CERTIFICATE-----":
                start = i
            elif content_lines[i] == "-----END CERTIFICATE-----":
                end = i
        if start == end or (start == 0 and length - end < 3):
            return 1
        else:
            if start != 0:
                key_content = "\n".join(content_lines[:start])
                crt_content = "\n".join(content_lines[start:])
            else:
                crt_content = "\n".join(content_lines[:end+1])
                key_content = "\n".join(content_lines[end+1:])
        if key_content.find("PRIVATE") != -1:
            os.rename(KEY_FILE, BAK_KEY_FILE)
            with open(KEY_FILE, "w") as key:
                key.write(key_content)
        if crt_content.find("CERTIFICATE") != -1:
            os.rename(CRT_FILE, BAK_CRT_FILE)
            with open(CRT_FILE, "w") as key:
                key.write(crt_content)
            # 使用subprocess参数列表避免命令注入
            cmd_args = [SLB_NGINX_SBIN, '-t']
            process = subprocess.Popen(cmd_args, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
            outmsg, errmsg = process.communicate()
            returncode = process.returncode
            if not returncode:
                os.remove(BAK_CRT_FILE)
                os.remove(BAK_KEY_FILE)
                # 使用subprocess调用systemctl
                reload_args = ['systemctl', 'reload', 'slb-nginx']
                subprocess.call(reload_args)
                return 0
            else:
                if cls.file_exists(BAK_CRT_FILE):
                    os.remove(CRT_FILE)
                    os.rename(BAK_CRT_FILE, CRT_FILE)
                if cls.file_exists(BAK_KEY_FILE):
                    os.remove(KEY_FILE)
                    os.rename(BAK_KEY_FILE, KEY_FILE)
                # 使用subprocess调用systemctl避免命令注入
                reload_args = ['systemctl', 'reload', 'slb-nginx']
                subprocess.call(reload_args)
        return 1
