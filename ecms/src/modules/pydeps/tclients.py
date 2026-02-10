#!/usr/bin/env python
# -*- coding:utf-8 -*-

'''
此库文件为各个thrift接口，提供简易的client对象创建接口.

For example:

from eisoo.tclients import TClient

with TClient("ShareMgnt") as client:
    client.Usrm_GetAllUserCount()
'''
from contextlib import contextmanager

from src.modules.pydeps import thriftlib


##################################################################
# 新集群模块接口
def ecmsagent_tclient(ip_addr="127.0.0.1", timeout_s=600):
    from ECMSAgent import ncTECMSAgent
    from ECMSAgent.constants import NCT_ECMSAGENT_PORT
    return thriftlib.BaseThriftClient(ip_addr, NCT_ECMSAGENT_PORT, ncTECMSAgent, timeout_s)


def ecmsmanager_tclient(ip_addr="127.0.0.1", timeout_s=600):
    from ECMSManager import ncTECMSManager
    from ECMSManager.constants import NCT_ECMSMANAGER_PORT
    return thriftlib.BaseThriftClient(ip_addr, NCT_ECMSMANAGER_PORT, ncTECMSManager, timeout_s)


def ecmsupgrademanager_tclient(ip_addr="127.0.0.1", timeout_s=600):
    from ECMSUpgrade import ncTECMSUpgradeManager
    from ECMSUpgrade.constants import NCT_UPGRADE_MANAGER_PORT
    return thriftlib.BaseThriftClient(
        ip_addr, NCT_UPGRADE_MANAGER_PORT, ncTECMSUpgradeManager, timeout_s)


def ecmsupgradeagent_tclient(ip_addr="127.0.0.1", timeout_s=600):
    from ECMSUpgrade import ncTECMSUpgradeAgent
    from ECMSUpgrade.constants import NCT_UPGRADE_AGENT_PORT
    return thriftlib.BaseThriftClient(
        ip_addr, NCT_UPGRADE_AGENT_PORT, ncTECMSUpgradeAgent, timeout_s)


##################################################################
# 数据模块接口
def eofs_tclient(ip_addr="127.0.0.1", timeout_s=0):
    """
    Get ncTEOFS thrift client
    """
    from EOFS import ncTEOFS
    from EOFS.constants import NCT_EOFS_PORT
    return thriftlib.BaseThriftClient(ip_addr, NCT_EOFS_PORT, ncTEOFS, timeout_s)


def evfs_tclient(ip_addr="127.0.0.1", timeout_s=0):
    """
    Get ncTEVFS thrift client
    """
    from EVFS import ncTEVFS
    from EVFS.constants import NCT_EVFS_PORT
    return thriftlib.BaseThriftClient(ip_addr, NCT_EVFS_PORT, ncTEVFS, timeout_s)


def efast_tclient(ip_addr="127.0.0.1", timeout_s=0):
    """
    Get ncTEFAST thrift client
    """
    from EFAST import ncTEFAST
    from EFAST.constants import NCT_EFAST_PORT
    return thriftlib.BaseThriftClient(ip_addr, NCT_EFAST_PORT, ncTEFAST, timeout_s)


def ecnjy_tclient(ip_addr="127.0.0.1", timeout_s=0):
    """
    Get ncTECNJY thrift client
    """
    from ECNJY import ncTECNJY
    from ECNJY.constants import NCT_ECNJY_PORT
    return thriftlib.BaseThriftClient(ip_addr, NCT_ECNJY_PORT, ncTECNJY, timeout_s)


def esearchmgnt_tclient(ip_addr="127.0.0.1", timeout_s=0):
    """
    Get ncTESearchMgnt thrift client
    """
    from ESearchMgnt import ncTESearchMgnt
    from ESearchMgnt.constants import NCT_ESM_PORT
    return thriftlib.BaseThriftClient(ip_addr, NCT_ESM_PORT, ncTESearchMgnt, timeout_s)


def ekeyscanmonitor_tclient(ip_addr="127.0.0.1", timeout_s=0):
    """
    Get ncTEKeyScanMonitor thrift client
    """
    from EKeyScanMonitor import ncTEKeyScanMonitor
    from EKeyScanMonitor.constants import NCT_KSM_PORT
    return thriftlib.BaseThriftClient(ip_addr, NCT_KSM_PORT, ncTEKeyScanMonitor, timeout_s)


##################################################################
# 应用模块接口
def eacp_tclient(ip_addr="127.0.0.1", timeout_s=0):
    """
    Get ncTEACP thrift client
    """
    from EACP import ncTEACP
    from EACP.constants import NC_T_EACP_PORT
    return thriftlib.BaseThriftClient(ip_addr, NC_T_EACP_PORT, ncTEACP, timeout_s)


def eacplog_tclient(ip_addr="127.0.0.1", timeout_s=0):
    """
    Get ncTEACPLog thrift client
    """
    from EACPLog import ncTEACPLog
    from EACPLog.constants import NC_T_EACP_LOG_PORT
    return thriftlib.BaseThriftClient(ip_addr, NC_T_EACP_LOG_PORT, ncTEACPLog, timeout_s)


def sharemgnt_tclient(ip_addr="127.0.0.1", timeout_s=0):
    """
    Get ncTShareMgnt thrift client
    """
    from ShareMgnt import ncTShareMgnt
    from ShareMgnt.constants import NCT_SHAREMGNT_PORT
    return thriftlib.BaseThriftClient(ip_addr, NCT_SHAREMGNT_PORT, ncTShareMgnt, timeout_s)


def sharesite_tclient(ip_addr="127.0.0.1", timeout_s=0):
    """
    Get ncTShareMgnt thrift client
    """
    from ShareSite import ncTShareSite
    from ShareSite.constants import NCT_SHARESITE_PORT
    return thriftlib.BaseThriftClient(ip_addr, NCT_SHARESITE_PORT, ncTShareSite, timeout_s)

def license_tclient(ip_addr="127.0.0.1", timeout_s=0):
    """
    Get ncTLicense thrift client
    """
    from License import ncTLicense
    from License.constants import NCT_LICENSE_PORT
    return thriftlib.BaseThriftClient(ip_addr, NCT_LICENSE_PORT, ncTLicense, timeout_s)

def deploymanager_tclient(ip_addr="127.0.0.1", timeout_s=0):
    """
    Get ncTDeployManager thrift client
    """
    from Deploy import ncTDeployManager
    from Deploy.constants import NCT_DEPLOYMANAGER_PORT
    return thriftlib.BaseThriftClient(ip_addr, NCT_DEPLOYMANAGER_PORT, ncTDeployManager, timeout_s)

def deployagent_tclient(ip_addr="127.0.0.1", timeout_s=0):
    """
    Get ncTDeployAgent thrift client
    """
    from Deploy import ncTDeployAgent
    from Deploy.constants import NCT_DEPLOYAGENT_PORT
    return thriftlib.BaseThriftClient(ip_addr, NCT_DEPLOYAGENT_PORT, ncTDeployAgent, timeout_s)

TCLIENT_CONFIG = {
    "ECMSAgent": ecmsagent_tclient,
    "ECMSManager": ecmsmanager_tclient,
    "ECMSUpgradeManager": ecmsupgrademanager_tclient,
    "ECMSUpgradeAgent": ecmsupgradeagent_tclient,
    "EOFS": eofs_tclient,
    "EVFS": evfs_tclient,
    "EFAST": efast_tclient,
    "ECNJY": ecnjy_tclient,
    "ESearchMgnt": esearchmgnt_tclient,
    "EACP": eacp_tclient,
    "EACPLog": eacplog_tclient,
    "ShareMgnt": sharemgnt_tclient,
    "ShareSite": sharesite_tclient,
    "License": license_tclient,
    "EKeyScanMonitor": ekeyscanmonitor_tclient,
    "DeployManager": deploymanager_tclient,
    "DeployAgent": deployagent_tclient,
}


@contextmanager
def TClient(TName="", ip_addr="127.0.0.1", timeout_s=0):
    """
    Get thrift client by name
    :param timeout_s  Socket 超时时间，单位：秒，默认为0，不超时
    """
    if timeout_s:
        client = TCLIENT_CONFIG[TName](ip_addr, timeout_s)
    else:
        client = TCLIENT_CONFIG[TName](ip_addr)
    try:
        yield client
    finally:
        client.close()
