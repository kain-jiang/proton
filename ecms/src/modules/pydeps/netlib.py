#!/usr/bin/env python
# -*- coding:utf-8 -*-

"""
网络相关公共库
"""

import re
import subprocess
from socket import inet_aton, inet_ntoa
from struct import unpack, pack
from netaddr.ip import IPAddress
from src.modules.pydeps.cmdprocess import shell_cmd_not_raise
from src.modules.pydeps.logger import syslog_cmd

MODULE_NAME = 'netlib'


def is_valid_ip(ip_addr):
    """
    检查IP合法
    @param string ip_addr: IPv4 or IPv6 address
    @return bool: True 符合IP规则, False 不符合IP规则
    """
    try:
        IPAddress(ip_addr)
    except Exception:
        return False
    else:
        return True


def get_ip_version(ip_addr):
    """
    the IP protocol version represented by this ip_addr object.
    @param string ip_addr: IPv4 or IPv6 address
    @return int: 4 表示 IPv4, 6 表示 IPv6, -1 表示非法 ip
    """
    try:
        ip = IPAddress(ip_addr)
    except Exception:
        return -1
    else:
        return ip.version


def is_valid_mask(mask):
    """
    检查子网掩码是否合法
    @param string mask: 子网掩码
    @return bool: True 符合掩码规则, False 不符合掩码规则
    """
    if is_valid_ip(mask):
        mask_num, = unpack("!I", inet_aton(mask))
        if mask_num == 0:
            return False

        # get inverted
        mask_num = ~mask_num + 1
        binstr = bin(mask_num)[3:]

        # convert to positive integer
        binstr = '0b%s' % ''.join('1' if b == '0' else '0' for b in binstr)
        mask_num = int(binstr, 2) + 1

        # check 2^n
        if mask_num & (mask_num - 1) == 0:
            return True
    return False


def is_useable_ip(ip_addr, mask=None):
    """
    根据 ip 地址与子网掩码,判断此IP是否可用
    @param string ip_addr: IP 地址
    @param string mask: 子网掩码
    @return bool: True 表示IP可用, False 表示IP不可用

    >>>netlib.is_useable_ip("192.168.0.1", "255.255.255.0")
    True
    >>>netlib.is_useable_ip("127.0.0.1") //Loopback address
    False
    >>>netlib.is_useable_ip("224.0.0.1") //Multicast address(224.0.0.0 - 239.255.255.255)
    False
    >>>netlib.is_useable_ip("169.254.0.1") //Failed dhcp allocation IP(169.254.x.x)
    False
    >>>netlib.is_useable_ip("192.168.77.128", "255.255.255.128") //Network number is 1
    False
    """
    if not is_valid_ip(ip_addr):
        return False

    ip_split = ip_addr.split('.')
    # 如果IP地址以0开头，则不可用
    if ip_split[0] == '0':
        return False
    # 如果IP地址以255开头，则不可用
    if ip_split[0] == '255':
        return False
    # 如果IP地址以127开头，则不可用
    if ip_split[0] == '127':
        return False
    # 如果IP地址以169.254开头，则不可用
    if ip_split[0] == '169' and ip_split[1] == '254':
        return False

    ip_num = ip_to_int(ip_addr)
    # 过滤全零地址
    if ip_num == 0:
        return False

    # 未指定掩码,则不进行其它验证
    if mask is None:
        return True

    if not is_valid_mask(mask):
        return False

    # 根据掩码计算子网地址，如果IP为子网地址，则不可用
    subnet = calc_subnet(ip_addr, mask)
    if ip_addr == subnet:
        return False
    # 根据子网以及掩码计算广播地址，如果IP为广播地址，则不可用
    if ip_addr == calc_broadcast_by_subnet(subnet, mask):
        return False
    return True


def ip_to_int(ip_addr):
    """
    将 ip 地址转换为整数
    @param string ip_addr: IP 地址
    @return int: 转换后的整数值
    >>>netlib.ip_to_int("192.168.0.1")
    3232235521
    """
    try:
        if is_valid_ip(ip_addr):
            result = unpack("!I", inet_aton(ip_addr))
            return result[0]
        else:
            errmsg = "IP address %s is not valid" % (ip_addr)
            raise Exception(errmsg)
    except Exception as ex:
        raise ex


def int_to_ip(int_num):
    """
    将整数值转换为IP地址
    @param int int_num: 一个整数值
    @return string: 转换后的IP地址
    >>>netlib.int2ip(3232235521)
    192.168.0.1
    """
    try:
        return inet_ntoa(pack("!I", int_num))
    except Exception:
        errmsg = "The integer %d conversion for IP failed" % (int_num)
        raise Exception(errmsg)


def calc_subnet(ip_addr, mask):
    """
    使用IP地址与子网掩码计算子网
    @param string ip_addr: IP 地址
    @param string mask: 子网掩码
    @param string: 网络地址

    >>>netlib.calcSubnet("192.168.0.1", "255.255.255.0")
    192.168.0.0
    """
    try:
        if is_valid_ip(ip_addr) and is_valid_ip(mask):
            ip_num, = unpack("!I", inet_aton(ip_addr))
            mask_num, = unpack("!I", inet_aton(mask))
            subnet_num = ip_num & mask_num
            sub_net = inet_ntoa(pack("!I", subnet_num))
            return sub_net
        else:
            errmsg = "calc subnet failed, ip_addr = %s, mask = %s" % (ip_addr, mask)
            raise Exception(errmsg)
    except Exception as ex:
        raise ex


def calc_host_num(mask):
    """
    计算主机数量
    @param string mask: 子网掩码
    @return int: 当前网络中主机数

    >>>netlib.calcHostNum("255.255.255.0")
    254
    """
    try:
        if is_valid_mask(mask):
            bit_num = bin(ip_to_int(mask)).count('1')
            return (2 ** (32 - bit_num)) - 2
        else:
            errmsg = "calc host num failed, mask = %s" % (mask)
            raise Exception(errmsg)
    except Exception as ex:
        raise ex


def exchange_mask_to_int(mask):
    """
    转换子网掩码格式,统计位数
    @param string mask : 子网掩码

    >>>netlib.exchange_mask_to_int("255.255.255.0")
    24
    """
    # 计算二进制字符串中 '1' 的个数
    count_bit = lambda bin_str: len([i for i in bin_str if i == '1'])
    mask_splited = mask.split('.')
    # 转换各段子网掩码为二进制, 计算十进制
    mask_count = [count_bit(bin(int(i))) for i in mask_splited]
    return sum(mask_count)


def exchange_int_to_mask(num):
    """
    转换子网掩码格式,统计位数
    @param string mask : 子网掩码

    >>>netlib.exchange_int_to_mask(24)
    255.255.255.0
    """
    bin_arr = ['0' for i in range(32)]
    for i in range(num):
        bin_arr[i] = '1'
    tmpmask = [''.join(bin_arr[i * 8:i * 8 + 8]) for i in range(4)]
    tmpmask = [str(int(tmpstr, 2)) for tmpstr in tmpmask]
    return '.'.join(tmpmask)


def is_same_network(ip_addr1, ip_addr2, mask):
    """
    判断两个IP是否在同一网段
    @param string ip_addr1: IP 地址
    @param string ip_addr2: IP 地址
    @param string mask: 子网掩码
    @return bool: True ip 在同一网段, False ip 不在同一网段

    >>>netlib.isInSameNetwork("192.168.77.1", "192.168.77.2", "255.255.255.0")
    True
    >>>netlib.isInSameNetwork("192.168.77.1", "192.168.8.2", "255.255.0.0")
    True
    >>>netlib.isInSameNetwork("192.168.77.1", "192.168.8.2", "255.255.255.0")
    False
    """
    try:
        if is_valid_ip(ip_addr1) and is_valid_ip(ip_addr2) and is_valid_mask(mask):
            ip1_num, = unpack("!I", inet_aton(ip_addr1))
            ip2_num, = unpack("!I", inet_aton(ip_addr2))
            mask_num, = unpack("!I", inet_aton(mask))
            if ip1_num & mask_num != ip2_num & mask_num:
                return False
            else:
                return True
        else:
            errmsg = "isInSameNetwork failed, ip_addr1 = %s, ip_addr2 = %s, mask = %s" \
                     % (ip_addr1, ip_addr2, mask)
            raise Exception(errmsg)
    except Exception as ex:
        raise ex


def calc_broadcast(ip_addr, mask):
    """
    根据IP地址和子网掩码计算广播地址
    @param string ip_addr: IP 地址
    @param string mask: 子网掩码
    @return string: 广播地址

    >>>netlib.calc_broadcast("192.168.77.12", "255.255.255.128")
    192.168.77.127
    """
    sub_net = calc_subnet(ip_addr, mask)
    broad_cast = calc_broadcast_by_subnet(sub_net, mask)
    return broad_cast


def calc_broadcast_by_subnet(subnet, mask):
    """
    根据子网地址计算广播地址
    @param string subnet: 子网地址
    @param string mask: 子网掩码
    @param string: 广播地址

    >>>netlib.calc_broadcast_by_subnet("192.168.77.0", "255.255.255.128")
    192.168.77.127
    """
    try:
        if not is_valid_mask(mask):
            errmsg = "Subnet mask %s is invalid." % (mask)
            raise Exception(errmsg)

        subnet_num = ip_to_int(subnet)

        # calc host bit num
        host_bit = bin(ip_to_int(mask)).count('1')

        # replace 32 - host_bit numbers 0 to 1
        binstr = ''
        if host_bit < 32:
            binstr = bin(subnet_num)[host_bit - 32:]

        binstr = ''.join('1' for b in binstr)
        binstr = ''.join([bin(subnet_num)[:host_bit + 2], binstr])

        broadcast_num = int(binstr, 2)
        return int_to_ip(broadcast_num)
    except Exception as ex:
        raise ex


def check_ipaddr_by_icmp(ipaddr, retry=3):
    """
    检查IP地址是否连通
    @param string ipaddr: IP地址
    @param string retry: 重试次数
    @return bool: True,可连通; False,不可连通

    >>>netlib.check_ipaddr_by_icmp("100.200.300.44", 1)
    >>>False
    >>>netlib.check_ipaddr_by_icmp("127.0.0.1")
    >>>True
    """
    cmdstr = 'ping -c %d %s' % (retry, ipaddr)
    try:
        process = subprocess.Popen(cmdstr,
                                   shell=True,
                                   stdout=subprocess.PIPE,
                                   stderr=subprocess.PIPE,
                                   close_fds=True)
        process.communicate()
        if not process.returncode:
            return True
        else:
            return False
    except Exception, err:
        raise err
    finally:
        if process.stdin:
            process.stdin.close()
        if process.stdout:
            process.stdout.close()
        if process.stderr:
            process.stderr.close()


def arping(nic_name, count, ip, dst_ip):
    """
    向相邻主机发送 ARP 请求
    @param string nic_name: 网卡名
    @param int count: 发送请求的次数
    @param string ip: IP
    @param string dst_ip: 目的地址
    """
    cmd = 'arping -I %s -c %d -s %s %s' % (nic_name, count, ip, dst_ip)
    (returncode, outmsg, errmsg) = shell_cmd_not_raise(cmd)
    syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg, returncode)
