#!/usr/bin/env python
# -*- coding:utf-8 -*-

"""
chrony agent 集群时间同步代理模块
"""
import os
import re
from src.modules.pydeps import cmdprocess, tracer, logger, filelib
import subprocess

CHRONY_CONFIG_PATH = '/etc/chrony.conf'
MODULE_NAME = 'ChronyAgent'
SERVICE_NAME = 'chronyd'

ncTChronyRole = {
    'UNKNOWN': 0,
    'MASTER': 1,
    'SLAVE': 2
}


CHRONY_SERVER_CONFIG = """
#master node
# These servers were defined in the installation:
# Use public servers from the pool.ntp.org project.
# Please consider joining the pool (http://www.pool.ntp.org/join.html).

# Ignore stratum in source selection.
stratumweight 0

# Record the rate at which the system clock gains/losses time.
driftfile /var/lib/chrony/drift

# Enable kernel RTC synchronization.
rtcsync

# In first three updates step the system clock instead of slew
# if the adjustment is larger than 10 seconds.
makestep 10 3

# Allow NTP client access from local network.
allow 0/0

# Listen for commands only on localhost.
bindcmdaddress 127.0.0.1
bindcmdaddress ::1

# Serve time even if not synchronized to any NTP server.
local stratum 10 orphan

keyfile /etc/chrony.keys

# Specify the key used as password for chronyc.
commandkey 1

# Generate command key if missing.
generatecommandkey

# Disable logging of client accesses.
noclientlog

# Send a message to syslog if a clock adjustment is larger than 0.5 seconds.
logchange 0.5

logdir /var/log/chrony
#log measurements statistics tracking
"""

CHRONY_CLIENT_CONFIG = """
#slave node
# These servers were defined in the installation:
# Use public servers from the pool.ntp.org project.
# Please consider joining the pool (http://www.pool.ntp.org/join.html).
server %s iburst minpoll 4 maxpoll 10

# Ignore stratum in source selection.
stratumweight 0

# Record the rate at which the system clock gains/losses time.
driftfile /var/lib/chrony/drift

# Enable kernel RTC synchronization.
rtcsync

# In first three updates step the system clock instead of slew
# if the adjustment is larger than 10 seconds.
makestep 10 3

# Allow NTP client access from local network.
#allow 192.168/16

# Listen for commands only on localhost.
bindcmdaddress 127.0.0.1
bindcmdaddress ::1

# Serve time even if not synchronized to any NTP server.
local stratum 10

keyfile /etc/chrony.keys

# Specify the key used as password for chronyc.
commandkey 1

# Generate command key if missing.
generatecommandkey

# Disable logging of client accesses.
noclientlog

# Send a message to syslog if a clock adjustment is larger than 0.5 seconds.
logchange 0.5

logdir /var/log/chrony
#log measurements statistics tracking
"""

CHRONY_DEFAULT_CONFIG = """
# These servers were defined in the installation:
# Use public servers from the pool.ntp.org project.
# Please consider joining the pool (http://www.pool.ntp.org/join.html).

# Ignore stratum in source selection.
stratumweight 0

# Record the rate at which the system clock gains/losses time.
driftfile /var/lib/chrony/drift

# Enable kernel RTC synchronization.
rtcsync

# In first three updates step the system clock instead of slew
# if the adjustment is larger than 10 seconds.
makestep 10 3

# Allow NTP client access from local network.
allow 0/0

# Listen for commands only on localhost.
bindcmdaddress 127.0.0.1
bindcmdaddress ::1

# Serve time even if not synchronized to any NTP server.
local stratum 10

keyfile /etc/chrony.keys

# Specify the key used as password for chronyc.
commandkey 1

# Generate command key if missing.
generatecommandkey

# Disable logging of client accesses.
noclientlog

# Send a message to syslog if a clock adjustment is larger than 0.5 seconds.
logchange 0.5

logdir /var/log/chrony
#log measurements statistics tracking
"""


class ChronyAgent(object):
    """
    This is chrony agent class
    """

    REGEX_SYSTEM_TIME = re.compile(
        r'^System time     : (\d+\.\d+) seconds (fast|slow) of NTP time$',
        re.MULTILINE,
    )

    def __init__(self):
        """
        pass
        """

    ########################################################################################
    # 以下函数为接口功能实现
    @classmethod
    @tracer.trace_func
    def set_chrony_server(cls):
        """设置 chrony server"""
        logger.syslog(MODULE_NAME, 'Set chrony server begin.')

        # 集群内部 chrony server 配置
        master_config = list()
        master_config.append(CHRONY_SERVER_CONFIG)
        master_config = "\n".join(master_config)

        # 重写配置
        filelib.write_file(CHRONY_CONFIG_PATH, master_config)

        logger.syslog(MODULE_NAME, 'Set chrony server end.')

    @classmethod
    @tracer.trace_func
    def set_chrony_client(cls, server_ip):
        """设置 chrony client"""
        logger.syslog(MODULE_NAME, 'Set chrony client begin, server=({0}).'.format(server_ip))

        # 集群 chrony client 配置
        slave_config = list()
        slave_config.append(CHRONY_CLIENT_CONFIG % (server_ip))
        slave_config = "\n".join(slave_config)

        # 重写配置
        filelib.write_file(CHRONY_CONFIG_PATH, slave_config)
        logger.syslog(MODULE_NAME, 'Set chrony client end, server=({0}).'.format(server_ip))

    @classmethod
    @tracer.trace_func
    def get_chrony_role(cls):
        """
        获取节点 chrony 角色，(server|master) or (client|slave)
        @return ncTChronyRole: chrony 角色
        """
        cfg_str = filelib.read_file(CHRONY_CONFIG_PATH)
        if re.search(r"(?m)^#master node$", cfg_str) is not None:
            role = ncTChronyRole["MASTER"]
        elif re.search(r"(?m)^#slave node$", cfg_str) is not None:
            role = ncTChronyRole["SLAVE"]
        else:
            role = ncTChronyRole["UNKNOWN"]
        return role

    @classmethod
    @cmdprocess.check_param_security_for_shell
    @tracer.trace_func
    def add_time_server(cls, server):
        """添加时间服务器"""
        logger.syslog(MODULE_NAME, 'Add time server %r begin' % server)

        # 添加时间服务器 - 使用subprocess参数列表避免命令注入
        cmd_args = ['chronyc', '-a', 'add', 'server', server]
        process = subprocess.Popen(cmd_args, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        outmsg, errmsg = process.communicate()
        cmd = ' '.join(cmd_args)  # 仅用于日志记录
        logger.syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg)

        # 写入配置文件
        data = 'server %s iburst minpoll 4 maxpoll 10' % server
        filelib.write_file(CHRONY_CONFIG_PATH, data + os.linesep, 'a')

        logger.syslog(MODULE_NAME, 'Add time server %r end' % server)

    @classmethod
    @cmdprocess.check_param_security_for_shell
    @tracer.trace_func
    def del_time_server(cls, server):
        """移除时间服务器"""
        logger.syslog(MODULE_NAME, 'Del time server %r begin' % server)

        # 移除时间服务器 - 使用subprocess参数列表避免命令注入
        cmd_args = ['chronyc', '-a', 'delete', server]
        process = subprocess.Popen(cmd_args, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        outmsg, errmsg = process.communicate()
        cmd = ' '.join(cmd_args)  # 仅用于日志记录
        logger.syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg)

        # 写入配置文件
        data = 'server %s iburst minpoll 4 maxpoll 10' % server
        conf = filelib.read_file(CHRONY_CONFIG_PATH)
        filelib.write_file(
            path=CHRONY_CONFIG_PATH,
            content=conf.replace(data + os.linesep, str())
        )

        logger.syslog(MODULE_NAME, 'Del time server %r end' % server)

    @classmethod
    @tracer.trace_func
    def clear_chrony_config(cls):
        """设置 chrony server"""
        logger.syslog(MODULE_NAME, 'Clear chrony config begin.')

        # 集群内部 chrony server 配置
        default_config = list()
        default_config.append(CHRONY_DEFAULT_CONFIG)
        default_config = "\n".join(default_config)

        # 重写配置
        filelib.write_file(CHRONY_CONFIG_PATH, default_config)

        logger.syslog(MODULE_NAME, 'Clear chrony config end.')

    @classmethod
    @tracer.trace_func
    def get_diff_from_ref(cls):
        """ 获取与当前使用的时间源的时间差异 """
        cmd = 'chronyc tracking'
        outmsg, errmsg = cmdprocess.shell_cmd(cmdstr=cmd)
        mobj = cls.REGEX_SYSTEM_TIME.search(outmsg)
        if mobj:
            diff, slow_or_fast = mobj.groups()
            return float(diff) * -1 * int(slow_or_fast == 'slow')
        else:
            raise Exception('cannot parse output of command %r' % cmd)

    @classmethod
    @tracer.trace_func
    def makestep(cls):
        """ 立刻与当前使用的时间源同步 """
        cmd = 'chronyc -a makestep'
        outmsg, errmsg = cmdprocess.shell_cmd(cmdstr=cmd)
        logger.syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg)
