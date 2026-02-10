#!/usr/bin/env python
# -*- coding:utf-8 -*-

"""This is sysagent"""
import os
import re
import time
import subprocess

from src.modules.pydeps import tracer, filelib
from src.modules.pydeps.logger import syslog, syslog_debug
from src.modules.pydeps.cmdprocess import check_param_security_for_shell
from src.modules.pydeps import cmdprocess

PUBLIC_XML_PATH = "/etc/firewalld/zones/public.xml"
TRUSTED_XML_PATH = "/etc/firewalld/zones/trusted.xml"

DEFAULT_PUBLIC_XML = """<?xml version="1.0" encoding="utf-8"?>
<zone target="default">
  <short>Public</short>
  <description>For use in public areas. You do not trust the other computers on networks to not harm your computer. Only selected incoming connections are accepted.</description>
  <rule family="ipv4">
    <port port="22" protocol="tcp" />
    <accept />
  </rule>
</zone>
"""

DEFAULT_TRUSTED_XML = """<?xml version="1.0" encoding="utf-8"?>
<zone target="ACCEPT">
  <short>Trusted</short>
  <description>All network connections are accepted.</description>
</zone>
"""


MODULE_NAME = 'FirewallAgent'


class FirewallAgent(object):
    """
    This is system agent class
    """
    @classmethod
    @check_param_security_for_shell
    @tracer.trace_func
    def add_rich_rule(cls, rich_rule, zone, is_permanent=True):
        """
        添加多条复杂规则, 不重载
        """
        # 使用subprocess参数列表避免命令注入
        cmd_args = ['firewall-cmd', '--add-rich-rule=' + rich_rule, '--zone=' + zone]
        process = subprocess.Popen(cmd_args, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        outmsg, errmsg = process.communicate()
        if process.returncode != 0:
            raise Exception(errmsg)
        syslog(MODULE_NAME, 'Added rich rule %s to %s' % (rich_rule, zone))

        if is_permanent is False:
            return

        # 添加永久规则
        permanent_cmd_args = ['firewall-cmd', '--add-rich-rule=' + rich_rule, '--zone=' + zone, '--permanent']
        process = subprocess.Popen(permanent_cmd_args, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        outmsg, errmsg = process.communicate()
        if process.returncode != 0:
            raise Exception(errmsg)
        syslog(MODULE_NAME, 'Added rich rule %s to %s permanent' % (rich_rule, zone))

    @classmethod
    @check_param_security_for_shell
    @tracer.trace_func
    def remove_rich_rule(cls, rich_rule, zone, is_permanent=True):
        """
        移除指定区域的永久规则,不重载
        """
        # 使用subprocess参数列表避免命令注入
        cmd_args = ['firewall-cmd', '--remove-rich-rule=' + rich_rule, '--zone=' + zone]
        process = subprocess.Popen(cmd_args, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        outmsg, errmsg = process.communicate()
        if process.returncode != 0:
            raise Exception(errmsg)
        syslog(MODULE_NAME, 'Removed rich rule %s to %s' % (rich_rule, zone))

        if is_permanent is False:
            return

        # 移除永久规则
        permanent_cmd_args = ['firewall-cmd', '--remove-rich-rule=' + rich_rule, '--zone=' + zone, '--permanent']
        process = subprocess.Popen(permanent_cmd_args, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        outmsg, errmsg = process.communicate()
        if process.returncode != 0:
            raise Exception(errmsg)
        syslog(MODULE_NAME, 'Removed rich rule %s to %s permanent' % (rich_rule, zone))

    @classmethod
    @check_param_security_for_shell
    @tracer.trace_func
    def add_source(cls, source, zone, is_permanent=True):
        """添加单条源地址"""
        # 使用subprocess参数列表避免命令注入
        cmd_args = ['firewall-cmd', '--add-source=' + source, '--zone=' + zone]
        process = subprocess.Popen(cmd_args, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        outmsg, errmsg = process.communicate()
        if process.returncode != 0:
            raise Exception(errmsg)
        syslog(MODULE_NAME, 'Add source(%s) to zone(%s) is_permanent=%r' % (
            source, zone, is_permanent))
        if is_permanent is False:
            return

        # 永久规则
        cmd_args.append("--permanent")
        syslog(MODULE_NAME, 'Add source(%s)' % source)
        cmdprocess.shell_cmd(cmd_args)
        syslog(MODULE_NAME, 'Add source(%s) success' % source)

    @classmethod
    @check_param_security_for_shell
    @tracer.trace_func
    def remove_source(cls, source, zone, is_permanent=True):
        """移除单条源地址"""
        # 使用subprocess参数列表避免命令注入
        cmd_args = ['firewall-cmd', '--remove-source=' + source, '--zone=' + zone]
        process = subprocess.Popen(cmd_args, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        outmsg, errmsg = process.communicate()
        returncode = process.returncode
        if returncode != 0 and "UNKNOWN_SOURCE" not in errmsg.decode('utf-8', errors='ignore'):
            raise Exception(errmsg)

        # cmdprocess.shell_cmd(cmd_str)
        syslog(MODULE_NAME, 'Remove source(%s) from zone(%s) is_permanent=%r' % (
            source, zone, is_permanent))
        if is_permanent is False:
            return

        # 执行永久规则
        cmd_args.append("--permanent")

        syslog(MODULE_NAME, 'Remove source(%s)' % source)
        cmdprocess.shell_cmd(cmd_args)
        syslog(MODULE_NAME, 'Remove source(%s) success' % source)

    @classmethod
    @check_param_security_for_shell
    @tracer.trace_func
    def remove_service(cls, service_name, zone, is_permanent=True):
        """移除指定区域的服务规则"""
        syslog(MODULE_NAME, 'Remove service %s from %s begin' % (service_name, zone))
        cmd_str = "firewall-cmd --remove-service='%s' --zone=%s" % (service_name, zone)

        # 构造永久规则
        if is_permanent is True:
            cmd_str = cmd_str + ' --permanent'
        cmdprocess.shell_cmd(cmd_str)

        syslog(MODULE_NAME, 'Remove service %s from %s end' % (service_name, zone))

    @classmethod
    @tracer.trace_func
    def get_firewall_info(cls, option, zone, is_permanent=True):
        """
        获取指定区域防火墙信息:
        @param rich-rule: 获取所有复杂规则
               service: 获取服务
               source: 获取源地址
        """
        if option not in ['rich-rule', 'service', 'source']:
            raise Exception('Param error: %s' % option)

        result_list = list()

        if option == 'rich-rule':
            if is_permanent is False:
                cmd_str = "firewall-cmd --list-rich-rules --zone=%s" % zone
                out, err = cmdprocess.shell_cmd(cmd_str)
                return cmdprocess.output_to_lines(out)
            # 读取永久规则
            cmd_str = ["firewall-cmd", "--permanent", "--zone", zone, "--list-rich-rules"]
            out, _ = cmdprocess.shell_cmd(cmd_str)
            return cmdprocess.output_to_lines(out)

        elif option == 'service':
            cmd_str = "firewall-cmd --list-services --zone=%s" % zone
            if is_permanent is True:
                cmd_str = cmd_str + ' --permanent'
            out, err = cmdprocess.shell_cmd(cmd_str)
            for each_line in cmdprocess.output_to_lines(out):
                result_list.extend(each_line.split(' '))
            return result_list

        elif option == 'source':

            if is_permanent is False:
                cmd_str = "firewall-cmd --list-sources --zone=%s" % zone
                out, err = cmdprocess.shell_cmd(cmd_str)
                for each_line in cmdprocess.output_to_lines(out):
                    result_list.extend(each_line.split(' '))
                return result_list

            # 读取永久规则
            cmd_str = ["firewall-cmd", "--permanent", "--zone", zone,  "--list-sources"]
            out, err = cmdprocess.shell_cmd(cmd_str)
            for each_line in cmdprocess.output_to_lines(out):
                result_list.extend(each_line.split(' '))
            return result_list

    @classmethod
    @check_param_security_for_shell
    @tracer.trace_func
    def get_target(cls, zone):
        """获取指定区域的链规则"""
        cmd_str = "firewall-cmd --get-target --zone=%s --permanent" % zone
        out, err = cmdprocess.shell_cmd(cmd_str)
        return out.strip()

    @classmethod
    @tracer.trace_func
    def set_target(cls, option, zone):
        """设置指定区域的链规则, 需要重载"""
        syslog(MODULE_NAME, 'Set target on zone[%s]' % zone)
        cmd_str = ["firewall-cmd", "--permanent", "--zone", zone, "--set-target", option]
        cmdprocess.shell_cmd(cmd_str)
        syslog(MODULE_NAME, 'Set target on zone[%s] success' % zone)

    @classmethod
    @tracer.trace_func
    def get_default_zone(cls):
        """获取防火墙默认区域"""
        cmd_str = 'firewall-cmd --get-default-zone'
        out, err = cmdprocess.shell_cmd(cmd_str)
        return out.strip()

    @classmethod
    @check_param_security_for_shell
    @tracer.trace_func
    def set_default_zone(cls, zone):
        """设置防火墙默认区域"""
        cmd_str = 'firewall-cmd --set-default-zone=%s' % zone
        out, err = cmdprocess.shell_cmd(cmd_str)
        syslog(MODULE_NAME, 'Set default zone=%s' % zone)

    @classmethod
    @tracer.trace_func
    def reload_firewall(cls, is_complete=False):
        """
        载入防火墙规则,默认不完全重载
        --reload             Reload firewall and keep state information
        --complete-reload    Reload firewall and lose state information
        """
        cmd_str = "firewall-cmd --reload"

        if is_complete is True:
            cmd_str = "firewall-cmd --complete-reload"

        retry = 3
        while retry:
            try:
                cmdprocess.shell_cmd(cmd_str)
                break
            except Exception as ex:
                syslog_debug(MODULE_NAME, str(ex))
                if retry == 1:
                    raise ex

                time.sleep(10 * (4 - retry))
                retry -= 1

        syslog_debug(MODULE_NAME, "Run cmd[%s]" % cmd_str)

    @classmethod
    @tracer.trace_func
    def init_firewall_xml(cls):
        """
        初始化成默认防火墙规则,仅允许22端口连接
        注意:重载后生效,此处不reload
        """
        syslog(MODULE_NAME, 'Clear firewall rule begin')

        filelib.write_file(PUBLIC_XML_PATH, DEFAULT_PUBLIC_XML, "w+")
        syslog(MODULE_NAME, 'Init firewall public.xml')
        filelib.write_file(TRUSTED_XML_PATH, DEFAULT_TRUSTED_XML, "w+")
        syslog(MODULE_NAME, 'Init firewall trusted.xml')

        syslog(MODULE_NAME, 'Clear firewall rule end')

# =================================================================
# 内部函数
# =================================================================
    @classmethod
    @tracer.trace_func
    def _decode_from_xml_obj(cls, xml_element):
        """获取xml文件结构中的rich rule规则"""
        rich_rule = 'rule family="ipv4"'
        source_obj = xml_element.find('source')
        if source_obj is not None:
            rich_rule = rich_rule + ' source address="%s"' % source_obj.attrib['address']

        destination_obj = xml_element.find('destination')
        if destination_obj is not None:
            rich_rule = rich_rule + ' destination address="%s"' % destination_obj.attrib['address']

        port_obj = xml_element.find('port')
        if port_obj is not None:
            rich_rule = rich_rule + ' port port="%s" protocol="%s"' % (port_obj.attrib['port'],
                                                                       port_obj.attrib['protocol'])
        rich_rule = rich_rule + ' accept'
        return rich_rule

    @classmethod
    @tracer.trace_func
    def _indent(cls, elem, level=0):
        i = "\n" + level*" "
        if len(elem):
            if not elem.text or not elem.text.strip():
                elem.text = i + "  "
            if not elem.tail or not elem.tail.strip():
                elem.tail = i
            for elem in elem:
                cls._indent(elem, level+2)
            if not elem.tail or not elem.tail.strip():
                elem.tail = i
        else:
            if level and (not elem.tail or not elem.tail.strip()):
                elem.tail = i
