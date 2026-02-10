#!/usr/bin/env python
# -*- coding:utf-8 -*-

"""公共定义"""

import time
import traceback
import uuid
import yaml

import err

from cmdprocess import shell_cmd_not_raise
import logger, tclients, tracer, clusterconf, netlib, safeconfig

from src.modules.ecmsdb import t_anyshare_conf
from src.modules.ecmsdb.t_node import TNodeDBManager
from src.modules.ecmsdb.t_service import TServiceDBManager

from ECMSManager.ttypes import ncTECMSManagerError
from ECMSAgent.ttypes import ncTServiceStatus
from EThriftException.ttypes import (ncTExpType, ncTException)


MODULE_NAME = err.ECMS_MANAGER_SERVICE_NAME


@tracer.trace_func
def troubleshoot_service(log_module, service):
    """
    在当前节点上检查并启动指定服务
    @param string log_module    记日志使用的模块名
    @param string service       服务名称
    """
    try:
        with tclients.TClient("ECMSAgent") as client:
            if client.get_service_status(service) != ncTServiceStatus.SS_STARTED:
                logger.syslog(
                    log_module, "Troubleshoot: service {0} is not running.".format(service))
                client.start_service(service)
                logger.syslog(
                    log_module, "Troubleshoot: start service {0} ok.".format(service))
    except Exception as ex:
        logger.syslog_exception(
            log_module, "Troubleshoot: check and start service {0} failed.".format(service), ex)

@tracer.trace_func
def verify_ecms_node():
    """
    验证当前节点是否为 ecms 节点
    """
    ecms_node_uuid = TNodeDBManager.get_role_ecms_master_uuid()
    if not ecms_node_uuid:
        raise Exception("Not found ecms node from cluster db.")

    # 验证够在 ecms 节点上执行的操作
    with tclients.TClient('ECMSAgent') as client:
        curr_node_uuid = client.get_node_uuid()
        if curr_node_uuid != ecms_node_uuid:
            raise Exception("This operation can only be performed on the master node.")
    return ecms_node_uuid

@tracer.trace_func
def is_external_db():
    """判断应用是否使用了第三方数据库"""
    try:
        return clusterconf.ClusterConfig.if_use_external_db()
    except Exception as ex:
        # 如果不存在app section, 为没有使用第三方
        if ex.errID == safeconfig.ERR_ID:
            return False
        else:
            raise ex

def raise_t_exception(exp_msg, file_name="", code_line=0,
                      exp_type=ncTExpType.NCT_WARN, exp_id=0, exp_detail=''):
    """
    丢出异常
    - exp_type   异常类型，默认警告
    - file_name  异常文件，默认为执行本函数的上一个堆栈所在文件
    - code_line  异常行号，默认为执行本函数的上一个堆栈所在行
    - exp_id     异常编号
    - exp_msg    异常内容
    """
    frame = traceback.extract_stack()
    if not file_name:
        file_name = frame[-2][0]
        code_line = frame[-2][1]

    exp = ncTException()
    exp.expType = exp_type
    exp.fileName = file_name
    exp.codeLine = code_line
    exp.errID = exp_id
    exp.expMsg = exp_msg
    exp.errProvider = MODULE_NAME
    exp.time = time.ctime()
    exp.errDetail = exp_detail
    raise exp

@tracer.trace_func
def angle_bracket_filter(old_string):
    new_string = []
    for c in old_string:
        if c == "<" or c == ">":
            continue
        new_string.append(c)
    return ''.join(new_string)

@tracer.trace_func
def validate_uuid(uuid_string, para_name):
    try:
        uuid.UUID(uuid_string)
    except ValueError:
        # If it's a value error, then the string is not a valid hex code for a UUID.
        raise_t_exception(exp_msg="%s is an invalid uuid" % (para_name),
                          exp_id=ncTECMSManagerError.NCT_INVALID_ARGUMENT)

@tracer.trace_func
def validate_ip(ip_string, para_name):
    if not netlib.is_valid_ip(ip_string):
        raise_t_exception(exp_msg="%s is an invalid IP" % (para_name),
                          exp_id=ncTECMSManagerError.NCT_INVALID_ARGUMENT)

@tracer.trace_func
def validate_usable_ip(ip_string):
    if netlib.check_ipaddr_by_icmp(ip_string):
        raise_t_exception(exp_msg="%s is already used " % (ip_string),
                          exp_id=ncTECMSManagerError.NCT_INVALID_ARGUMENT)

@tracer.trace_func
def validate_nonnegative_int(num, para_name):
    if (not isinstance(num, int)) or num < 0:
        raise_t_exception(exp_msg="%s is not a nonnegative int" % (para_name),
                          exp_id=ncTECMSManagerError.NCT_INVALID_ARGUMENT)

@tracer.trace_func
def validate_same_subnet(vip_info, ha_slave_ip, netmask=''):
    """
    判断高可用从节点是否和高可用主节点在同一网段
    @param netmask 子网掩码 255.255.255.0
    """
    # vip 与 从节点是否在同一网段
    if netmask == '':
        with tclients.TClient('ECMSAgent', ha_slave_ip) as client:
            node_ifaddr = client.get_ifaddr_by_ipaddr(ipaddr=ha_slave_ip)
            netmask = node_ifaddr.netmask
    if vip_info.mask != netmask or \
    not netlib.is_same_network(vip_info.vip, ha_slave_ip, vip_info.mask):
        raise_t_exception(
            exp_msg="ha master node ip {0} and ha slave node ip {1} are not in the same subnet."
            .format(vip_info.vip, ha_slave_ip),
            exp_id=ncTECMSManagerError.NCT_NOT_IN_SAME_SUNNET)

@tracer.trace_func
def no_blankspace_in_string(para_str, para_name):
    if " " in para_str:
        raise_t_exception(exp_msg="%s with blankspace is invalid" % (para_name),
                          exp_id=ncTECMSManagerError.NCT_INVALID_ARGUMENT)

@tracer.trace_func
def get_lvs_port(role):
    """获取lvs端口"""
    port_list = list()

    if role == "app":
        port_list.extend(TServiceDBManager.get_lvs_port_by_sys_role("app"))

        web_client_port = t_anyshare_conf.TAnyShareConfDBManager.get(
            t_anyshare_conf.WEB_CLIENT_PORT_KEY)
        if web_client_port is None:
            web_client_port = 443
        port_list.append(web_client_port)

        web_client_http_port = t_anyshare_conf.TAnyShareConfDBManager.get(
            t_anyshare_conf.WEB_CLIENT_HTTP_PORT_KEY)
        if web_client_http_port is None:
            web_client_http_port = 80
        port_list.append(web_client_http_port)

        eacp_https_port = t_anyshare_conf.TAnyShareConfDBManager.get(
            t_anyshare_conf.EACP_HTTPS_PORT_KEY)
        if eacp_https_port is None:
            eacp_https_port = 9999
        port_list.append(eacp_https_port)

        efast_https_port = t_anyshare_conf.TAnyShareConfDBManager.get(
            t_anyshare_conf.EFAST_HTTPS_PORT_KEY)
        if efast_https_port is None:
            efast_https_port = 9124
        port_list.append(efast_https_port)

    elif role == "storage":
        port_list.extend(TServiceDBManager.get_lvs_port_by_sys_role("storage"))

        eoss_http_port = t_anyshare_conf.TAnyShareConfDBManager.get(
            t_anyshare_conf.EOSS_HTTP_PORT_KEY)
        if eoss_http_port is None:
            eoss_http_port = 9028
        port_list.append(eoss_http_port)

        eoss_https_port = t_anyshare_conf.TAnyShareConfDBManager.get(
            t_anyshare_conf.EOSS_HTTPS_PORT_KEY)
        if eoss_https_port is None:
            eoss_https_port = 9029
        port_list.append(eoss_https_port)

    return list(set(port_list))

@tracer.trace_func
def get_container_platform_info(debug=True):
    """获取容器平台信息

    :raises Exception: anysharectl 命令执行失败
    :return(dict): 返回容器平台信息
    """
    cmd = "anysharectl get conf -v"
    code, out, err_msg = shell_cmd_not_raise(cmd)
    if debug:
        logger.syslog_cmd(MODULE_NAME, cmd, out, err_msg, code)
    if code != 0:
        err_msg = "Get container platform info failed, cmd:%s, code:%s, err:%s" \
            % (cmd, str(code), err_msg)
        raise Exception(err_msg)

    return yaml.load(out, Loader=yaml.Loader)

class BaseProcessor(object):
    """
    base processor class
    """

    @classmethod
    @tracer.trace_func
    def on_is_env_dirty(cls, config):
        """检查节点环境操作"""
        logger.syslog(MODULE_NAME, 'skip on_is_env_dirty in %s' % cls.__name__)

    @classmethod
    @tracer.trace_func
    def on_clear_node(cls, config):
        """清理节点"""
        logger.syslog(MODULE_NAME, 'skip on_clear_node in %s' % cls.__name__)

    @classmethod
    @tracer.trace_func
    def on_prepare_envrionment(cls, config):
        """检查节点环境操作"""
        logger.syslog(MODULE_NAME, 'skip on_prepare_envrionment in %s' % cls.__name__)

    @classmethod
    @tracer.trace_func
    def on_active_cluster(cls, config):
        """激活集群"""
        logger.syslog(MODULE_NAME, 'skip on_active_cluster in %s' % cls.__name__)

    @classmethod
    @tracer.trace_func
    def on_add_node_into_cluster(cls, config):
        """添加节点"""
        logger.syslog(MODULE_NAME, 'skip on_add_node_into_cluster in %s' % cls.__name__)

    @classmethod
    @tracer.trace_func
    def on_node_join_cluster(cls, config):
        """节点加入到集群"""
        logger.syslog(MODULE_NAME, 'skip on_node_join_cluster in %s' % cls.__name__)

    @classmethod
    @tracer.trace_func
    def on_node_enter_ecms(cls, config):
        """节点进入ecms"""
        logger.syslog(MODULE_NAME, 'skip on_node_enter_ecms in %s' % cls.__name__)

    @classmethod
    @tracer.trace_func
    def on_node_enter_db_master(cls, conf):
        """节点数据库成为主库"""
        logger.syslog(MODULE_NAME, 'on_node_enter_db_master in %s' % cls.__name__)

    @classmethod
    @tracer.trace_func
    def on_node_enter_db_slave(cls, conf):
        """节点数据库成为从库"""
        logger.syslog(MODULE_NAME, 'on_node_enter_db_slave in %s' % cls.__name__)

    @classmethod
    @tracer.trace_func
    def on_prepare_ecms(cls, config):
        """准备ecms环境"""
        logger.syslog(MODULE_NAME, 'skip on_prepare_ecms in %s' % cls.__name__)

    @classmethod
    @tracer.trace_func
    def on_remove_node_from_cluster(cls, config):
        """从集群中移除节点"""
        logger.syslog(MODULE_NAME, 'skip on_remove_node_from_cluster in %s' % cls.__name__)

    @classmethod
    @tracer.trace_func
    def on_consistency_check(cls):
        """一致性检查"""
        logger.syslog(MODULE_NAME, 'skip on_consistency_check in %s' % cls.__name__)

    @classmethod
    @tracer.trace_func
    def on_consistency_repair(cls):
        """一致性修复"""
        logger.syslog(MODULE_NAME, 'skip on_consistency_repair in %s' % cls.__name__)

    @classmethod
    @tracer.trace_func
    def on_change_node_ip(cls, config):
        """更改节点ip"""
        raise Exception("Called a pure virtual function.")

    ###########################################################################
    # 模块管理
    ###########################################################################

    @classmethod
    @tracer.trace_func
    def on_enable_chrony(cls, config):
        """启用 chrony 模块"""
        logger.syslog(MODULE_NAME, 'skip on_enable_chrony in %s' % cls.__name__)

    @classmethod
    @tracer.trace_func
    def on_disable_chrony(cls, config):
        """禁用 chrony 模块"""
        logger.syslog(MODULE_NAME, 'skip on_disable_chrony in %s' % cls.__name__)

    @classmethod
    @tracer.trace_func
    def on_enable_keepalived(cls, config):
        """启用 keepalived 模块"""
        logger.syslog(MODULE_NAME, 'skip on_enable_keepalived in %s' % cls.__name__)

    @classmethod
    @tracer.trace_func
    def on_disable_keepalived(cls, config):
        """禁用 keepalived 模块"""
        logger.syslog(MODULE_NAME, 'skip on_disable_keepalived in %s' % cls.__name__)

    @classmethod
    @tracer.trace_func
    def on_enable_zabbix(cls, config):
        """启用 zabbix 模块"""
        logger.syslog(MODULE_NAME, 'skip on_enable_zabbix in %s' % cls.__name__)

    @classmethod
    @tracer.trace_func
    def on_disable_zabbix(cls, config):
        """禁用 zabbix 模块"""
        logger.syslog(MODULE_NAME, 'skip on_disable_zabbix in %s' % cls.__name__)
