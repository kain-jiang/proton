#!/usr/bin/env python
# -*- coding: utf-8 -*-
# @Time    : 2021/4/29 9:52
# @Author  : mo.kang<mo.kang@eisoo.com>
# @Site    : 
# @File    : system_agent.py
# @Software: PyCharm
import os
import copy
import multiprocessing
import commands
import subprocess
import logging

from src.modules.pydeps import filelib
from src.modules.pydeps import cmdprocess

# 初始化日志记录器
logger = logging.getLogger(__name__)

# 配置日志记录
if not logger.handlers:
    handler = logging.StreamHandler()
    formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
    handler.setFormatter(formatter)
    logger.addHandler(handler)
    logger.setLevel(logging.INFO)

RPS_SOCK = 32768
MODULE_NAME = "SysAgent"
STSCTL_CONFIG_PATH = "/etc/sysctl.d"
SYSCTL_CONFIG = "/etc/sysctl.conf"
PROTON_SYSCTL_CONFIG = "/etc/sysctl.d/proton.conf"


class SystemAgent(object):
    @classmethod
    def get_sysctl_lasting_parameters(cls):
        all_conf = list()
        all_conf.append(SYSCTL_CONFIG)
        confd = os.listdir(STSCTL_CONFIG_PATH)
        if confd:
            for conf_file in confd:
                if conf_file.endswith(".conf"):
                    all_conf.append(os.path.join(STSCTL_CONFIG_PATH, conf_file))
        all_conf_dict = dict()
        for conf in all_conf:
            lines_list = filelib.read_file(conf, return_type="list")
            line_dict = cls.list_to_dict(lines_list)
            all_conf_dict.update(line_dict)
        return all_conf_dict

    @classmethod
    def get_sysctl_parameters(cls):
        outmsg, _ = cmdprocess.shell_cmd("sysctl -ea")
        conf_list = outmsg.split("\n")
        return cls.list_to_dict(conf_list)

    @classmethod
    def create_sysctl_parameters(cls, conf_dict):
        proton_dict = dict()
        if not isinstance(conf_dict, dict):
            raise Exception("%s: is not type of dict" % conf_dict)
        for key, value in conf_dict.iteritems():
            # 安全地设置 sysctl 参数，使用 subprocess 模块直接传递参数列表
            import subprocess
            # 使用参数列表而不是字符串拼接，避免命令注入
            cmd_args = ["sysctl", "-w", "{0}={1}".format(str(key), str(value))]
            process = subprocess.Popen(
                cmd_args,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE
            )
            _, errmsg = process.communicate()
            returncode = process.returncode
            
            if returncode != 0 or "sysctl: setting key" in errmsg:
                # 对于固定命令，可以继续使用 shell_cmd
                cmdprocess.shell_cmd("sysctl -p")
                cmdprocess.shell_cmd("sysctl -p /etc/sysctl.d/*")
                raise Exception(errmsg)
        if os.path.isfile(PROTON_SYSCTL_CONFIG):
            old_conf_list = filelib.read_file(PROTON_SYSCTL_CONFIG, return_type='list')
            proton_dict = cls.list_to_dict(old_conf_list)
        proton_dict.update(conf_dict)
        conf_list = cls.dict_to_list(proton_dict)
        filelib.write_file(PROTON_SYSCTL_CONFIG, conf_list, "w+")

    @classmethod
    def update_sysctl_parameters(cls, conf_dict):
        if not os.path.isfile(PROTON_SYSCTL_CONFIG):
            open(PROTON_SYSCTL_CONFIG, "w").close()
        for key, value in conf_dict.iteritems():
            # 安全地设置 sysctl 参数，使用 subprocess 模块直接传递参数列表
            import subprocess
            # 使用参数列表而不是字符串拼接，避免命令注入
            cmd_args = ["sysctl", "-w", "{0}={1}".format(str(key), str(value))]
            process = subprocess.Popen(
                cmd_args,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE
            )
            _, errmsg = process.communicate()
            returncode = process.returncode
            
            if returncode != 0 or "sysctl: setting key" in errmsg:
                # 对于固定命令，可以继续使用 shell_cmd
                cmdprocess.shell_cmd("sysctl -p")
                cmdprocess.shell_cmd("sysctl -p /etc/sysctl.d/*")
                raise Exception(errmsg)
        fina_conf_dict = copy.deepcopy(conf_dict)
        all_conf = list()
        confd = os.listdir(STSCTL_CONFIG_PATH)
        if confd:
            for conf_file in confd:
                if conf_file.endswith(".conf"):
                    all_conf.append(os.path.join(STSCTL_CONFIG_PATH, conf_file))
        all_conf.append(SYSCTL_CONFIG)
        for conf in all_conf:
            if fina_conf_dict:
                file_conf_list = filelib.read_file(conf, return_type='list')
                file_conf_dict = cls.list_to_dict(file_conf_list)
                for key, value in conf_dict.iteritems():
                    if key in file_conf_dict:
                        file_conf_dict[key] = value
                        fina_conf_dict.pop(key, None)
                file_conf_list = cls.dict_to_list(file_conf_dict)
                filelib.write_file(conf, file_conf_list, "w+")
        if fina_conf_dict:
            file_conf_list = filelib.read_file(PROTON_SYSCTL_CONFIG, return_type='list')
            file_conf_dict = cls.list_to_dict(file_conf_list)
            file_conf_dict.update(fina_conf_dict)
            file_conf_list = cls.dict_to_list(file_conf_dict)
            filelib.write_file(PROTON_SYSCTL_CONFIG, file_conf_list, "w+")
        cmdprocess.shell_cmd("sysctl -p")
        cmdprocess.shell_cmd("sysctl -p /etc/sysctl.d/*")

    @classmethod
    def delete_sysctl_parameters(cls, conf_list):
        all_conf = list()
        all_conf.append(SYSCTL_CONFIG)
        confd = os.listdir(STSCTL_CONFIG_PATH)
        if confd:
            for conf_file in confd:
                if conf_file.endswith(".conf"):
                    all_conf.append(os.path.join(STSCTL_CONFIG_PATH, conf_file))
        for conf_file in all_conf:
            file_conf_list = filelib.read_file(conf_file, return_type='list')
            file_conf_dict = cls.list_to_dict(file_conf_list)
            for conf in conf_list:
                if conf in file_conf_dict:
                    file_conf_dict.pop(conf)
            file_conf_dict = cls.dict_to_list(file_conf_dict)
            filelib.write_file(conf_file, file_conf_dict, "w+")
        cmdprocess.shell_cmd("sysctl -p")
        cmdprocess.shell_cmd("sysctl -p /etc/sysctl.d/*")

    @classmethod
    def list_to_dict(cls, conf_list):
        if not isinstance(conf_list, list):
            raise Exception("%s: is not type of list" % conf_list)
        conf_dict = dict()
        for line in conf_list:
            if line and not line.startswith("#") and not line.startswith("sysctl: ") and line.find("=") != -1:
                conf_line = line.split("=")
                if len(conf_line) == 2:
                    key, value = conf_line[0].strip(), conf_line[1].strip()
                    conf_dict[key] = value
        return conf_dict

    @classmethod
    def dict_to_list(cls, conf_dict):
        if not isinstance(conf_dict, dict):
            raise Exception("%s: is not type of list" % conf_dict)
        conf_list = list()
        for key, value in conf_dict.iteritems():
            conf_line = str(key) + " = " + str(value)
            conf_list.append(conf_line)
        return conf_list

    @classmethod
    def get_conf(cls, conf_name):
        conf_list = filelib.read_file(conf_name, return_type='list')
        conf_dict = cls.list_to_dict(conf_list)
        return {conf_name: conf_dict}

    @classmethod
    def restart_services(cls, service_name):
        # 验证服务名称为有效的 systemd 服务名称
        if not service_name or not isinstance(service_name, (str, unicode)):
            logger.error("Invalid service name: %s" % service_name)
            return -1
            
        # 检查服务名称是否只包含允许的字符（字母、数字、连字符、下划线、点）
        import re
        if not re.match(r'^[a-zA-Z0-9_\-\.]+$', service_name):
            logger.error("Invalid characters in service name: %s" % service_name)
            return -1
            
        # 使用subprocess参数列表避免命令注入
        try:
            cmd_args = ['systemctl', 'restart', service_name]
            process = subprocess.Popen(
                cmd_args, 
                stdout=subprocess.PIPE, 
                stderr=subprocess.PIPE,
                shell=False  # 确保不通过shell执行
            )
            outmsg, errmsg = process.communicate()
            
            if process.returncode != 0:
                logger.error("Failed to restart service %s: %s" % (service_name, errmsg))
            else:
                logger.info("Successfully restarted service: %s" % service_name)
                
            return process.returncode
            
        except Exception as e:
            logger.error("Error restarting service %s: %s" % (service_name, str(e)))
            return -1

    def get_rpscpus(self, cpunum):
        cpustr = hex(2 ** cpunum - 1)[2:]
        if 'L' in cpustr:
            cpustr = cpustr[:-1]
        cpulist = []
        while len(cpustr) > 8:
            cpulist.append("," + cpustr[len(cpustr) - 8:])
            cpustr = cpustr[:-8]
        return cpustr + "".join(cpulist)

    def get_queues_num(self, dirname):
        queues = os.listdir(dirname)
        myqueues = [i for i in queues if i.startswith("tx")]
        return len(myqueues)

    def get_priority_cpu(self, eth):
        # 直接读取文件而不是执行命令，避免命令注入
        try:
            numa_node_path = os.path.join('/sys/class/net', eth, 'device/numa_node')
            with open(numa_node_path, 'r') as f:
                node = f.read().strip()
        except Exception as e:
            raise Exception("%s: Network card node acquisition failed: %s" % (eth, str(e)))
            
        try:
            cpulist_path = os.path.join('/sys/devices/system/node', 'node' + str(node), 'cpulist')
            with open(cpulist_path, 'r') as f:
                cpulist = f.read().strip()
        except Exception as e:
            raise Exception("%s: Failed to get node's cpulist: %s" % (node, str(e)))
            
        cpu_list = []
        for i in cpulist.split(","):
            cpu_list += self.get_list(int(i.split("-")[0]), int(i.split("-")[1]))
        return cpu_list

    def get_list(self, start, end=None):
        if end:
            return [i for i in range(start, end + 1)]
        else:
            return [i for i in range(start)]

    def get_cpu_bind_list(self, cpunum, eth):
        all_cpu_list = self.get_list(cpunum)
        priority_cpu_list = self.get_priority_cpu(eth)
        final_cpu_list = priority_cpu_list + list(set(all_cpu_list) - set(priority_cpu_list))
        return final_cpu_list

    def get_eths(self):
        all_eths = os.listdir('/sys/class/net/')
        nets_list = copy.deepcopy(all_eths)
        for eth in nets_list:
            if eth.startswith("tun") or eth == 'lo' or eth.startswith('veth') or eth.startswith(
                    'docker') or eth.startswith('cali'):
                all_eths.remove(eth)
        return all_eths

    def get_interrupts(self, eth):
        status, result = commands.getstatusoutput("ls -U /sys/class/net/%s/device/msi_irqs" % eth)
        if status:
            raise Exception("%s: Failed to obtain network card interrupt" % eth)
        return [i.strip() for i in result.split("\n")]

    @classmethod
    def bind_core(cls):
        import subprocess
        
        # 使用 subprocess 替代 commands.getstatusoutput
        process = subprocess.Popen(
            ["systemctl", "disable", "irqbalance", "--now"],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE
        )
        _, _ = process.communicate()
        if process.returncode != 0:
            raise Exception("systemctl disable irqbalance --now : Execution failed")

        cpunum = int(multiprocessing.cpu_count())
        
        # 安全地写入文件而不是使用 echo 命令
        try:
            with open("/proc/sys/net/core/rps_sock_flow_entries", "w") as f:
                f.write(str(RPS_SOCK))
        except IOError:
            raise Exception("Failed to write to /proc/sys/net/core/rps_sock_flow_entries")

        rps_cpus = cls.get_rpscpus(cpunum)
        all_eths = cls.get_eths()
        for eth in all_eths:
            bound_core = []
            interrupts = cls.get_interrupts(eth)
            queues = cls.get_queues_num("/sys/class/net/%s/queues/" % eth)
            cpu_bind_list = cls.get_cpu_bind_list(cpunum, eth)
            for interrupt in interrupts:
                irq_path = "/proc/irq/%s/smp_affinity_list" % interrupt
                if os.path.exists("/proc/irq/%s/" % interrupt):
                    try:
                        # 安全地写入文件而不是使用 echo 命令
                        with open(irq_path, "w") as f:
                            f.write(str(cpu_bind_list[0]))
                    except IOError:
                        raise Exception("Failed to write to %s" % irq_path)
                    
                    cpu = cpu_bind_list[0]
                    bound_core.append(cpu)
                    cpu_bind_list.pop(0)
                    cpu_bind_list.append(cpu)
            
            for i in range(queues):
                rps_flow_cnt_path = "/sys/class/net/%s/queues/rx-%s/rps_flow_cnt" % (eth, str(i))
                try:
                    # 安全地写入文件而不是使用 echo 命令
                    with open(rps_flow_cnt_path, "w") as f:
                        f.write(str(RPS_SOCK / len(interrupts)))
                except IOError:
                    raise Exception("Failed to write to %s" % rps_flow_cnt_path)
                
                rps_cpus_path = "/sys/class/net/%s/queues/rx-%s/rps_cpus" % (eth, str(i))
                try:
                    # 安全地写入文件而不是使用 echo 命令
                    with open(rps_cpus_path, "w") as f:
                        f.write(str(rps_cpus))
                except IOError:
                    raise Exception("Failed to write to %s" % rps_cpus_path)

    @classmethod
    def get_sys_info(cls):
        import subprocess
        sysInfo = {}
        
        # 使用 subprocess 参数化执行命令
        # hostname
        process = subprocess.Popen(
            ["hostname"],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE
        )
        HostName, errmsg = process.communicate()
        returncode = process.returncode
        logger.syslog_cmd(MODULE_NAME, "hostname", HostName, errmsg, returncode)
        sysInfo["name"] = HostName.strip()
        
        # 直接读取文件而不是使用 grep 命令
        try:
            with open("/proc/meminfo", "r") as f:
                for line in f:
                    if "MemTotal" in line:
                        # 从行中提取数字
                        parts = line.split()
                        if len(parts) >= 2:
                            MemTotal = int(parts[1]) / 1024 / 1024  # 转换为 GB
                            break
            logger.syslog_cmd(MODULE_NAME, "read /proc/meminfo", str(MemTotal), "", 0)
            sysInfo["memory"] = str(MemTotal)
        except Exception as e:
            logger.syslog_cmd(MODULE_NAME, "read /proc/meminfo", "", str(e), 1)
            sysInfo["memory"] = "0"
        
        # dmidecode 命令
        # system-serial-number
        process = subprocess.Popen(
            ["dmidecode", "-s", "system-serial-number"],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE
        )
        SeriaNumber, errmsg = process.communicate()
        returncode = process.returncode
        logger.syslog_cmd(MODULE_NAME, "dmidecode -s system-serial-number", SeriaNumber, errmsg, returncode)
        sysInfo["sn"] = SeriaNumber.strip()
        
        # system-manufacturer
        process = subprocess.Popen(
            ["dmidecode", "-s", "system-manufacturer"],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE
        )
        Vendor, errmsg = process.communicate()
        returncode = process.returncode
        logger.syslog_cmd(MODULE_NAME, "dmidecode -s system-manufacturer", Vendor, errmsg, returncode)
        sysInfo["vendor"] = Vendor.strip()
        
        # system-product-name
        process = subprocess.Popen(
            ["dmidecode", "-s", "system-product-name"],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE
        )
        ProductName, errmsg = process.communicate()
        returncode = process.returncode
        logger.syslog_cmd(MODULE_NAME, "dmidecode -s system-product-name", ProductName, errmsg, returncode)
        sysInfo["product-name"] = ProductName.strip()

        return sysInfo