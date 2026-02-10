#!/usr/bin/env python
# -*- coding: utf-8 -*-
# @Time    : 2021/4/2 9:58
# @Author  : mo.kang<mo.kang@eisoo.com>
# @Site    :
# @File    : ecms_handler.py
# @Software: PyCharm
import base64
import hashlib
import json
import re
import time
import uuid
import logging

from concurrent.futures import ThreadPoolExecutor

from tornado.web import RequestHandler
from tornado.concurrent import run_on_executor

from src.modules.ecms_agent.firewall_agent import FirewallAgent
from src.modules.ecms_agent.net_agent import NetAgent
from src.modules.ecms_agent.file_agent import FileAgent
from src.modules.ecms_agent.system_agent import SystemAgent
from src.modules.ecms_agent.chrony_agent import ChronyAgent
from src.modules.pydeps.netlib import get_ip_version, is_valid_ip

logger = logging.getLogger(__name__)

errmsg = {
    "code": "",
    "message": "",
    "cause": "",
    "detail": ""
}


SECRET_PATH = ["/bin", "/boot", "/dev", "/etc", "/home", "/lib", "/lib64", "/media", "/mnt", "/opt",
               "/proc", "/root", "/run", "/sbin", "/srv", "/sys", "/tmp", "/usr", "/var", "/"]


def handler_try_except(func):
    """用于handler异常处理的装饰器"""

    def wrapper(self, *args, **kwargs):
        try:
            func(self, *args, **kwargs)
        except Exception as e:
            logger.error("handle exception: %s", e)
            status = 500010000
            cause = ""
            message = "server exception"
            detail = ""
            if e.message:
                all_message = re.findall('\((.*?)\)', str(e.message), re.S)
                if len(all_message) > 3:
                    cause = all_message[2].strip().replace('\n', '').replace('\r', '').split("STDERR:")[-1]
                    if not cause:
                        cause = all_message[3].strip().replace('\n', '').replace('\r', '').split("STDOUT:")[-1]
                    detail = all_message[0].strip().replace('\n', '').replace('\r', '')
                else:
                    cause = str(e.message).strip().replace('\n', '').replace('\r', '')
            else:
                message = "Unknown exception"
            errmsg = {"code": status, "message": message, "cause": cause, "detail": detail}
            self.write(json.dumps(errmsg))
            self.set_status(500)
        return

    return wrapper


def check_type(parameter, ptype):
    errmsg["code"] = "400010000"
    errmsg["message"] = "Parameter type error"
    if not parameter and parameter is not False:
        errmsg["cause"] = "Parameter cannot be empty"
        return errmsg
    if ptype == "path":
        if not isinstance(parameter, unicode):
            errmsg["cause"] = "%s :Parameter is not of %s" % (parameter, unicode)
            return errmsg
        if parameter.find("..") != -1:
            errmsg["message"] = "Path not allowed"
            errmsg["cause"] = "..: Is not allowed to be used in path"
            return errmsg
        if not parameter.startswith("/"):
            errmsg["message"] = "Path not allowed"
            errmsg["cause"] = "path: Absolute path is not used"
            return errmsg
        if parameter in SECRET_PATH:
            errmsg["message"] = "Non-secure directory"
            errmsg["cause"] = "%s: Non-secure directory" % parameter
            return errmsg
        return
    if ptype == "ip":
        if not isinstance(parameter, unicode):
            errmsg["cause"] = "%s :Parameter is not of %s" % (parameter, unicode)
            return errmsg
        if not is_valid_ip(parameter):
            errmsg["cause"] = "%s : Not an ip address" % parameter
            return errmsg
        return
    if ptype == "prefix":
        if isinstance(parameter, int) and parameter > 128:
            errmsg["cause"] = "%s :Parameter should be less than or equal to %d" % (parameter, 128)
            return errmsg
        if not isinstance(parameter, (str, unicode)):
            errmsg["cause"] = "%s :Parameter is not of %s" % (parameter, (str, unicode))
            return errmsg
        if (
            not parameter.isdigit()
            or int(parameter) > 128
            or int(parameter) <= 0
        ):
            errmsg["cause"] = "%s :Parameter should be 0 < Parameter <= 128" % (parameter)
            return errmsg
        return
    if not isinstance(parameter, ptype):
        errmsg["cause"] = "%s :Parameter is not of %s" % (parameter, ptype)
        return errmsg

def get_basic_auth(handler):
    """get basic auth from the RequestHandler"""
    header = handler.request.headers.get("Authorization")
    if not header or not header.startswith("Basic "):
        return "", ""
    logger.debug("Get http header, Authorization: %s", header)

    auth = base64.b64decode(header[6:])
    u, p = auth.decode("utf-8").split(":", 2)

    return u, p

def generate_simple_token(username, epoch):
    """generate simple token"""
    seed = "{0}:{1}".format(username, epoch)
    logger.debug("Generate simple token for seed: %s", seed)
    return hashlib.sha256(seed.encode()).hexdigest()

def simple_auth(func):
    """simple auth"""
    def wrapper(self, *args, **kwargs):
        """wrapper"""
        username, password = get_basic_auth(self)
        logger.debug("Get basic auth, username: %s, password: %s", username, password)

        epoch = int(time.time())
        epoch_min = epoch - epoch % 60
        logger.debug("Current epoch mod minute: %d", epoch_min)
        for e in (epoch_min, epoch_min - 60, epoch_min + 60):
            want = generate_simple_token(username, e)
            logger.debug("Check password, got: %s, want: %s", password, want)
            if password == want:
                return func(self, *args, **kwargs)

        msg = {
                "code": "401000000",
                "message": "Invalid username or password",
        }
        self.set_status(401)
        json.dump(msg, self)

    return wrapper


class BaseHandler(RequestHandler):
    """解决JS跨域请求问题"""
    executor = ThreadPoolExecutor(60)

    def data_received(self, chunk):
        pass

    def set_default_headers(self):
        # self.set_header('Access-Control-Allow-Origin', '*')
        # self.set_header('Access-Control-Allow-Methods', 'POST, GET')
        # self.set_header('Access-Control-Max-Age', 1000)
        # self.set_header('Access-Control-Allow-Headers', '*')
        self.set_header('Content-type', 'application/json')


class FirewalldHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def post(self):
        complete = False
        if self.request.body:
            body = json.loads(self.request.body)
            if "is_complete" in body:
                complete = body["is_complete"]
                mess = check_type(complete, bool)
                if mess:
                    self.write(mess)
                    self.set_status(400)
                    return
        FirewallAgent.reload_firewall(complete)
        self.set_status(201)


class FirewalldDefaultZoneHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def get(self):
        default_zone = FirewallAgent.get_default_zone()
        self.write(json.dumps({"default-zone": default_zone}))

    @run_on_executor
    @handler_try_except
    def post(self):
        zone = "public"
        if self.request.body:
            body = json.loads(self.request.body)
            if "zone" in body:
                zone = body["zone"]
                mess = check_type(zone, unicode)
                if mess:
                    self.write(mess)
                    self.set_status(400)
                    return
        FirewallAgent.set_default_zone(zone)
        self.set_status(201)


class FirewalldFirewallHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def post(self):
        FirewallAgent.init_firewall_xml()
        self.set_status(201)


class FirewalldRichRulesHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def get(self):
        is_permanent = self.get_argument('is_permanent', 'true')
        if is_permanent == 'true':
            is_permanent = True
        elif is_permanent == "false":
            is_permanent = False
        else:
            mess = check_type(is_permanent, bool)
            if mess:
                self.write(mess)
                self.set_status(400)
                return
        zone = self.get_argument('zone', 'public')
        mess = check_type(zone, (unicode, str))
        if mess:
            self.write(mess)
            self.set_status(400)
            return
        rich_rules = json.dumps(FirewallAgent.get_firewall_info("rich-rule", zone, is_permanent))
        self.write(rich_rules)

    @run_on_executor
    @handler_try_except
    def post(self):
        zone = "public"
        is_permanent = True
        if self.request.body:
            body = json.loads(self.request.body)
            if body:
                if "zone" in body:
                    zone = body["zone"]
                    mess = check_type(zone, unicode)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                if "is_permanent" in body:
                    is_permanent = body["is_permanent"]
                    mess = check_type(is_permanent, bool)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                if "rich_rule" in body:
                    mess = check_type(body["rich_rule"], unicode)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                    FirewallAgent.add_rich_rule(body["rich_rule"], zone, is_permanent)
                    self.set_status(201)
                    return
        errmsg["code"] = "400010000"
        errmsg["message"] = "Parameter abnormal"
        errmsg["cause"] = "Missing required parameters: rich_rule"
        self.write(errmsg)
        self.set_status(400)

    @run_on_executor
    @handler_try_except
    def delete(self):
        zone = "public"
        is_permanent = True
        if self.request.body:
            body = json.loads(self.request.body)
            if body:
                if "zone" in body:
                    zone = body["zone"]
                    mess = check_type(zone, unicode)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                if "is_permanent" in body:
                    is_permanent = body["is_permanent"]
                    mess = check_type(is_permanent, bool)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                if "rich_rule" in body:
                    mess = check_type(body["rich_rule"], unicode)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                    FirewallAgent.remove_rich_rule(body["rich_rule"], zone, is_permanent)
                    self.set_status(204)
                    return
        errmsg["code"] = "400010000"
        errmsg["message"] = "Parameter abnormal"
        errmsg["cause"] = "Missing required parameters: rich_rule"
        self.write(errmsg)
        self.set_status(400)


class FirewalldServicesHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def get(self):
        is_permanent = self.get_argument('is_permanent', 'true')
        if is_permanent == 'true':
            is_permanent = True
        elif is_permanent == "false":
            is_permanent = False
        else:
            mess = check_type(is_permanent, bool)
            if mess:
                self.write(mess)
                self.set_status(400)
                return
        zone = str(self.get_argument('zone', 'public'))
        mess = check_type(zone, (unicode, str))
        if mess:
            self.write(mess)
            self.set_status(400)
            return
        services = json.dumps(FirewallAgent.get_firewall_info("service", zone, is_permanent))
        self.write(services)

    @run_on_executor
    @handler_try_except
    def delete(self):
        zone = "public"
        is_permanent = True
        if self.request.body:
            body = json.loads(self.request.body)
            if body:
                if "zone" in body:
                    zone = body["zone"]
                    mess = check_type(zone, unicode)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                if "is_permanent" in body:
                    is_permanent = body["is_permanent"]
                    mess = check_type(is_permanent, bool)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                if "service_name" in body:
                    service_name = body["service_name"]
                    mess = check_type(service_name, unicode)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                    FirewallAgent.remove_service(service_name, zone, is_permanent)
                    self.set_status(204)
                    return
        errmsg["code"] = "400010000"
        errmsg["message"] = "Parameter abnormal"
        errmsg["cause"] = "Missing required parameters: service_name"
        self.write(errmsg)
        self.set_status(400)


class FirewalldSourcesHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def get(self):
        is_permanent = self.get_argument('is_permanent', 'true')
        if is_permanent == 'true':
            is_permanent = True
        elif is_permanent == "false":
            is_permanent = False
        else:
            mess = check_type(is_permanent, bool)
            if mess:
                self.write(mess)
                self.set_status(400)
                return
        zone = self.get_argument('zone', 'public')
        mess = check_type(zone, (unicode, str))
        if mess:
            self.write(mess)
            self.set_status(400)
            return
        services = json.dumps(FirewallAgent.get_firewall_info("source", zone, is_permanent))
        self.write(services)

    @run_on_executor
    @handler_try_except
    def post(self, *args, **kwargs):
        zone = "public"
        is_permanent = True
        if self.request.body:
            body = json.loads(self.request.body)
            if body:
                if "zone" in body:
                    zone = body["zone"]
                    mess = check_type(zone, unicode)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                if "is_permanent" in body:
                    is_permanent = body["is_permanent"]
                    mess = check_type(is_permanent, bool)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                if "source" in body:
                    source = body["source"]
                    mess = check_type(source, unicode)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                    FirewallAgent.add_source(source, zone, is_permanent)
                    self.set_status(201)
                    return
        errmsg["code"] = "400010000"
        errmsg["message"] = "Parameter abnormal"
        errmsg["cause"] = "Missing required parameters: source"
        self.write(errmsg)
        self.set_status(400)

    @run_on_executor
    @handler_try_except
    def delete(self):
        zone = "public"
        is_permanent = True
        if self.request.body:
            body = json.loads(self.request.body)
            if body:
                if "zone" in body:
                    zone = body["zone"]
                    mess = check_type(zone, unicode)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                if "is_permanent" in body:
                    is_permanent = body["is_permanent"]
                    mess = check_type(is_permanent, bool)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                if "source" in body:
                    source = body["source"]
                    mess = check_type(source, unicode)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                    FirewallAgent.remove_source(source, zone, is_permanent)
                    self.set_status(204)
                    return
        errmsg["code"] = "400010000"
        errmsg["message"] = "Parameter abnormal"
        errmsg["cause"] = "Missing required parameters: source"
        self.write(errmsg)
        self.set_status(400)


class FirewalldTargetsHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def get(self):
        zone = self.get_argument("zone", "public")
        target = json.dumps(FirewallAgent.get_target(zone))
        self.write(json.dumps({"default-target": target}))

    @run_on_executor
    @handler_try_except
    def post(self, *args, **kwargs):
        zone = "public"
        if self.request.body:
            body = json.loads(self.request.body)
            if body:
                if "zone" in body:
                    zone = body["zone"]
                    mess = check_type(zone, unicode)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                if "option" in body:
                    option = body["option"]
                    mess = check_type(option, unicode)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                    if option not in ["default", "ACCEPT", "%%REJECT%%", "DROP"]:
                        errmsg["code"] = "400010000"
                        errmsg["message"] = "Parameter abnormal"
                        errmsg["cause"] = 'option : should be in ["default", "ACCEPT", "%%REJECT%%", "DROP"]'
                        self.write(errmsg)
                        self.set_status(400)
                        return
                    else:
                        FirewallAgent.set_target(option, zone)
                        self.set_status(201)
                        return
        errmsg["code"] = "400010000"
        errmsg["message"] = "Parameter abnormal"
        errmsg["cause"] = 'Missing required parameters: option'
        self.write(errmsg)
        self.set_status(400)


class NetArpHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def get(self, ipaddr):
        mess = check_type(ipaddr, "ip")
        if not mess:
            if NetAgent.exists_arp(ipaddr):
                self.write({"result": True})
            else:
                self.write({"result": False})
        else:
            self.write(mess)
            self.set_status(400)

    @run_on_executor
    @handler_try_except
    def delete(self, ipaddr):
        mess = check_type(ipaddr, "ip")
        if not mess:
            NetAgent.del_arp(ipaddr)
            self.set_status(204)
        else:
            self.write(mess)
            self.set_status(400)


class NetInfNameHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def get(self):
        inf_list = NetAgent.get_interface_name_for_vip()
        self.write(json.dumps(inf_list))


class NetIpsHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def get(self):
        ips = NetAgent.get_ip_addrs()
        self.write(json.dumps(ips))


class NetNicIpaddrHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def post(self, nic_name):
        gateway = None
        cause = ""
        if self.request.body:
            body = json.loads(self.request.body)
            if body:
                if "ipaddr" not in body:
                    cause = 'Missing required parameters: ipaddr'
                elif "label" not in body:
                    cause = 'Missing required parameters: label'
                elif ":" not in body["ipaddr"] and "netmask" not in body:   # ipv4 时 netmask 必须
                    cause = 'Missing required parameters: netmask'
                elif ":" in body["ipaddr"] and "prefix" not in body:        # ipv6 时 prefix 必须
                    cause = 'Missing required parameters: prefix'
                else:
                    if "gateway" in body:
                        gateway = body["gateway"]
                        mess = check_type(gateway, "ip")
                        if mess:
                            self.write(mess)
                            self.set_status(400)
                            return
                    ipaddr = body["ipaddr"]
                    mess = check_type(ipaddr, "ip")
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                    label = body["label"]
                    mess = check_type(label, unicode)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                    netmask = body.get("netmask", "")
                    if netmask:
                        mess = check_type(netmask, "ip")
                        if mess:
                            self.write(mess)
                            self.set_status(400)
                            return
                    prefix = body.get("prefix", "")
                    if prefix:
                        mess = check_type(prefix, "prefix")
                        if mess:
                            self.write(mess)
                            self.set_status(400)
                            return
                    if get_ip_version(ipaddr) != get_ip_version(gateway):
                        # ip地址和网关的IP协议版本需要相同
                        self.write({
                            "code": "400010000",
                            "message": "the IP protocol versions of ipaddr and gateway are not equal.",
                            "cause": "",
                            "detail": ""
                        })
                        self.set_status(400)
                        return
                    ifaddr = {
                        "nic_dev_name": nic_name,
                        "label": label,
                        "ipaddr": ipaddr,
                        "netmask": netmask,
                        "gateway": gateway,
                        "prefix": prefix
                    }
                    NetAgent.set_ifaddr(ifaddr)
                    self.set_status(201)
                    return
        errmsg["code"] = "400010000"
        errmsg["message"] = "Missing required parameters"
        errmsg["cause"] = cause
        self.write(errmsg)
        self.set_status(400)


class NetNicIfaddrHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def get(self):
        type = self.get_argument('type', '')
        value = self.get_argument('value', '')
        if type and value:
            if type == "label":
                mess = check_type(value, unicode)
                if mess:
                    self.write(mess)
                    self.set_status(400)
                    return
                ips = NetAgent.get_ifaddr(value)
                self.write(json.dumps(ips))
            elif type == "ipaddr":
                mess = check_type(value, "ip")
                if mess:
                    self.write(mess)
                    self.set_status(400)
                    return
                ips = NetAgent.get_ifaddr_by_ipaddr(value)
                self.write(json.dumps(ips))
            else:
                errmsg["code"] = "400010000"
                errmsg["message"] = "Parameter abnormal"
                errmsg["cause"] = 'type : should be in ["ipaddr", "label"]'
                self.write(errmsg)
                self.set_status(400)
        else:
            errmsg["code"] = "400010000"
            errmsg["message"] = "Parameter abnormal"
            errmsg["cause"] = "Missing required parameters: type and value"
            self.write(errmsg)
            self.set_status(400)

    @run_on_executor
    @handler_try_except
    def delete(self):
        label = self.get_argument('label', '')
        if label:
            NetAgent.del_ifaddr(label)
            self.set_status(204)
        else:
            errmsg["code"] = "400010000"
            errmsg["message"] = "Missing required parameters"
            errmsg["cause"] = "Missing required parameters: label"
            self.write(errmsg)
            self.set_status(400)


class NetNicsHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def get(self):
        interface_list = NetAgent.get_nics()
        self.write(json.dumps(interface_list))

    @run_on_executor
    @handler_try_except
    def post(self):
        if self.request.body:
            body = json.loads(self.request.body)
            if body:
                if "nic_name_list" in body:
                    nic_list = body["nic_name_list"]
                    if nic_list:
                        mess = check_type(nic_list, list)
                        if mess:
                            self.write(mess)
                            self.set_status(400)
                            return
                        nic_names = NetAgent.get_interface_name_for_bond()
                        for nic in nic_list:
                            if nic not in nic_names:
                                errmsg["code"] = "400010000"
                                errmsg["message"] = "Network card does not exist"
                                errmsg["cause"] = "Network card does not exist: %s" % nic
                                self.write(errmsg)
                                self.set_status(400)
                                return
                        NetAgent.bind_nics(nic_list)
                        self.set_status(201)
                        return
                    else:
                        errmsg["code"] = "400010000"
                        errmsg["message"] = "Parameter is empty"
                        errmsg["cause"] = "The value of the nic_name_list parameter cannot be empty"
                        self.write(errmsg)
                        self.set_status(400)
                        return
        errmsg["code"] = "400010000"
        errmsg["message"] = "Missing required parameters"
        errmsg["cause"] = "Missing required parameters: nic_name_list"
        self.write(errmsg)
        self.set_status(400)

    # @gen.coroutine
    @run_on_executor
    @handler_try_except
    def delete(self):
        if self.request.body:
            body = json.loads(self.request.body)
            if body:
                if "bond_dev_name" in body:
                    bond_dev_name = body["bond_dev_name"]
                    mess = check_type(bond_dev_name, unicode)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                    if bond_dev_name:
                        NetAgent.unbind_nic(bond_dev_name)
                        # self.finish()
                        self.set_status(204)
                        # NetAgent.reload_network()
                        return
        errmsg["code"] = "400010000"
        errmsg["message"] = "Missing required parameters"
        errmsg["cause"] = "Missing required parameters: bond_dev_name"
        self.write(errmsg)
        self.set_status(400)


class DirectoryHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def get(self):
        path_type = self.get_argument('type', 'directory')
        path = str(self.get_argument('path', ''))
        if not path:
            errmsg["code"] = "400010000"
            errmsg["message"] = "Missing required parameters"
            errmsg["cause"] = "Missing required parameters: path"
            self.write(errmsg)
            self.set_status(400)
        else:
            if path_type == "directory":
                flag = FileAgent.directory_exists(path)
            elif path_type == "file":
                flag = FileAgent.file_exists(path)
            else:
                errmsg["code"] = "400010000"
                errmsg["message"] = "Parameter abnormal"
                errmsg["cause"] = 'type : should be in ["directory", "file"]'
                self.write(errmsg)
                self.set_status(400)
                return
            self.write(json.dumps(flag))

    @run_on_executor
    @handler_try_except
    def post(self):
        path_type = "directory"
        mode = None
        if self.request.body:
            body = json.loads(self.request.body)
            if body:
                if "path" in body:
                    path = body["path"]
                    mess = check_type(path, unicode)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                    if not path.startswith("/"):
                        errmsg["code"] = "400010000"
                        errmsg["message"] = "Path error"
                        errmsg["cause"] = "path: Absolute path is not used"
                        self.write(errmsg)
                        self.set_status(400)
                        return
                    if "mode" in body:
                        mess = check_type(body["mode"], int)
                        if mess:
                            self.write(mess)
                            self.set_status(400)
                            return
                        mode = str(body["mode"]).zfill(4)
                        if not mode.isdigit() or len(mode) > 4 or not re.search('^[0-7]{4}$', mode):
                            errmsg["code"] = "400010000"
                            errmsg["message"] = "Parameter abnormal"
                            errmsg["cause"] = "mode: wrong format"
                            self.write(errmsg)
                            self.set_status(400)
                            return
                    if "type" in body:
                        path_type = body["type"]
                        mess = check_type(path, unicode)
                        if mess:
                            self.write(mess)
                            self.set_status(400)
                            return
                    if path_type == "directory":
                        if mode:
                            mode = int(mode, base=8)
                            FileAgent.create_directory(path, mode)
                        else:
                            FileAgent.create_directory(path)
                    elif path_type == "file":
                        if mode:
                            mode = int(mode, base=8)
                            FileAgent.creat_file(path, mode)
                        else:
                            FileAgent.creat_file(path)
                    else:
                        errmsg["code"] = "400010000"
                        errmsg["message"] = "Parameter abnormal"
                        errmsg["cause"] = 'type : should be in ["directory", "file"]'
                        self.write(errmsg)
                        self.set_status(400)
                        return
                    self.set_status(201)
                    return
        errmsg["code"] = "400010000"
        errmsg["message"] = "Missing required parameters"
        errmsg["cause"] = "Missing required parameters: path"
        self.write(errmsg)
        self.set_status(400)

    @run_on_executor
    @handler_try_except
    def delete(self):
        path_type = "directory"
        if self.request.body:
            body = json.loads(self.request.body)
            if body:
                if "path" in body:
                    path = body["path"]
                    mess = check_type(path, "path")
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                    if path.endswith("/") and len(path) > 1:
                        path = path[:-1]
                    if "type" in body:
                        path_type = body["type"]
                        mess = check_type(path_type, unicode)
                        if mess:
                            self.write(mess)
                            self.set_status(400)
                            return
                    if path_type == "directory":
                        FileAgent.delete_directory(path)
                    elif path_type == "file":
                        FileAgent.delete_file(path)
                    else:
                        errmsg["code"] = "400010000"
                        errmsg["message"] = "Parameter abnormal"
                        errmsg["cause"] = 'type : should be in ["directory", "file"]'
                        self.write(errmsg)
                        self.set_status(400)
                        return
                    self.set_status(204)
                    return
        errmsg["code"] = "400010000"
        errmsg["message"] = "Missing required parameters"
        errmsg["cause"] = "Missing required parameters: path"
        self.write(errmsg)
        self.set_status(400)


class CodeHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def get(self):
        machine_code = uuid.UUID(int=uuid.getnode()).hex[-12:].upper()
        self.write(json.dumps({"machine_code": machine_code}))


class SysctlHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def get(self):
        sysctl_dict = SystemAgent.get_sysctl_parameters()
        self.write(json.dumps(sysctl_dict))
        self.set_status(200)

    @run_on_executor
    @handler_try_except
    def post(self):
        if self.request.body:
            body = json.loads(self.request.body)
            if body:
                for key, value in body.iteritems():
                    if not isinstance(value, (unicode, int)) or isinstance(value, bool) or not value:
                        errmsg["code"] = "400010000"
                        errmsg["message"] = "Parameter type error"
                        errmsg["cause"] = "%s:Parameters should be of type str or int" % value
                        self.write(errmsg)
                        self.set_status(400)
                        return
                conf_dict = SystemAgent.get_sysctl_lasting_parameters()
                intersection = None
                if conf_dict:
                    intersection = set(body.keys()) & set(conf_dict.keys())
                if not intersection:
                    SystemAgent.create_sysctl_parameters(body)
                    self.set_status(201)
                    return
                else:
                    errmsg["code"] = "400010000"
                    errmsg["message"] = "Parameter already exists"
                    errmsg["cause"] = "Parameter already exists: %s" % intersection
                    self.write(errmsg)
                    self.set_status(400)
                    return
        errmsg["code"] = "400010000"
        errmsg["message"] = "Data cannot be empty"
        errmsg["cause"] = "Data cannot be empty"
        self.write(errmsg)
        self.set_status(400)

    @run_on_executor
    @handler_try_except
    def put(self):
        if self.request.body:
            body = json.loads(self.request.body)
            if body:
                for key, value in body.iteritems():
                    if not isinstance(value, (unicode, int)) or isinstance(value, bool) or not value:
                        errmsg["code"] = "400010000"
                        errmsg["message"] = "Parameter type error"
                        errmsg["cause"] = "%s:Parameters should be of type str or int" % value
                        self.write(errmsg)
                        self.set_status(400)
                        return
                conf_dict = SystemAgent.get_sysctl_parameters()
                intersection = None
                if conf_dict:
                    intersection = set(body.keys()) - (set(body.keys()) & set(conf_dict.keys()))
                if not intersection:
                    SystemAgent.update_sysctl_parameters(body)
                    self.set_status(204)
                    return
                else:
                    errmsg["code"] = "400010000"
                    errmsg["message"] = "Non-system parameters"
                    errmsg["cause"] = "Non-system parameters: %s" % intersection
                    self.write(errmsg)
                    self.set_status(400)
                    return
        errmsg["code"] = "400010000"
        errmsg["message"] = "Data cannot be empty"
        errmsg["cause"] = "Data cannot be empty"
        self.write(errmsg)
        self.set_status(400)

    @run_on_executor
    @handler_try_except
    def delete(self):
        if self.request.body:
            body = json.loads(self.request.body)
            if body:
                if "parameters" in body:
                    parameters = body["parameters"]
                    mess = check_type(parameters, list)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                    conf_dict = SystemAgent.get_sysctl_parameters()
                    subtraction = set(parameters) - set(conf_dict.keys())
                    if not subtraction:
                        SystemAgent.delete_sysctl_parameters(parameters)
                        self.set_status(204)
                        return
                    else:
                        errmsg["code"] = "400010000"
                        errmsg["message"] = "Parameter does not exist"
                        errmsg["cause"] = "Parameter does not exist: %s" % subtraction
                        self.write(errmsg)
                        self.set_status(400)
                        return
        errmsg["code"] = "400010000"
        errmsg["message"] = "Missing required parameters"
        errmsg["cause"] = "Missing required parameters : parameters"
        self.write(errmsg)
        self.set_status(400)


class ChronyRoleHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def get(self):
        role = ChronyAgent.get_chrony_role()
        self.write(json.dumps(role))
        self.set_status(200)


class ChronyDiffHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def get(self):
        diff = ChronyAgent.get_diff_from_ref()
        self.write(json.dumps(diff))
        self.set_status(200)


class ChronyServerHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def post(self):
        if self.request.body:
            body = json.loads(self.request.body)
            if body:
                if "server" in body:
                    server = body['server']
                    mess = check_type(server, unicode)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                    ChronyAgent.add_time_server(server)
                    self.set_status(201)
                    return
        errmsg["code"] = "400010000"
        errmsg["message"] = "Missing required parameters"
        errmsg["cause"] = "Missing required parameters : server"
        self.write(errmsg)
        self.set_status(400)

    @run_on_executor
    @handler_try_except
    def delete(self):
        if self.request.body:
            body = json.loads(self.request.body)
            if body:
                if "server" in body:
                    server = body['server']
                    mess = check_type(server, unicode)
                    if mess:
                        self.write(mess)
                        self.set_status(400)
                        return
                    ChronyAgent.del_time_server(server)
                    self.set_status(204)
                    return
        errmsg["code"] = "400010000"
        errmsg["message"] = "Missing required parameters"
        errmsg["cause"] = "Missing required parameters : server"
        self.write(errmsg)
        self.set_status(400)


class ChronyHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def post(self):
        if self.request.body:
            body = json.loads(self.request.body)
            if body:
                if "role" in body:
                    role = body['role']
                    if role == "server":
                        ChronyAgent.set_chrony_server()
                        self.set_status(201)
                        return
                    elif role == "client":
                        if "server" in body:
                            server = body["server"]
                            mess = check_type(server, unicode)
                            if mess:
                                self.write(mess)
                                self.set_status(400)
                                return
                            ChronyAgent.set_chrony_client(server)
                            self.set_status(201)
                            return
                        else:
                            errmsg["code"] = "400010000"
                            errmsg["message"] = "Missing required parameters"
                            errmsg["cause"] = "Missing required parameters : server"
                            self.write(errmsg)
                            self.set_status(400)
                            return
                    else:
                        errmsg["code"] = "400010000"
                        errmsg["message"] = "Parameter abnormal"
                        errmsg["cause"] = 'type : should be in ["server", "client"]'
                        self.write(errmsg)
                        self.set_status(400)
                        return
        errmsg["code"] = "400010000"
        errmsg["message"] = "Missing required parameters"
        errmsg["cause"] = "Missing required parameters : role"
        self.write(errmsg)
        self.set_status(400)

    @run_on_executor
    @handler_try_except
    def delete(self):
        ChronyAgent.clear_chrony_config()
        self.set_status(204)


class TLSHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def put(self):
        if self.request.body:
            status = FileAgent.update_tls(self.request.body)
            if status == 0:
                self.set_status(204)
                return
            elif status < 0:
                errmsg["code"] = "400010000"
                errmsg["message"] = "TLS error"
                errmsg["cause"] = "Certificate does not exist"
                self.write(errmsg)
                self.set_status(404)
                return
            else:
                errmsg["code"] = "400010000"
                errmsg["message"] = "TLS error"
                errmsg["cause"] = "Certificate or key content error"
                self.write(errmsg)
                self.set_status(400)
                return
        errmsg["code"] = "400010000"
        errmsg["message"] = "The certificate and key error"
        errmsg["cause"] = "The certificate and key content cannot be empty"
        self.write(errmsg)
        self.set_status(400)


class ServiceHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def post(self):
        if self.request.body:
            body = json.loads(self.request.body)
            if body and "service_name" in body and body["service_name"]:
                service_name = body["service_name"]
                if not isinstance(service_name, (str, unicode)):
                    errmsg["code"] = "400010000"
                    errmsg["message"] = "Type error"
                    errmsg["cause"] = "The service_name should be of type str"
                    self.write(errmsg)
                    self.set_status(400)
                    return
                status = SystemAgent.restart_services(body["service_name"])
                if status:
                    errmsg["code"] = "400010000"
                    errmsg["message"] = "Service restart error"
                    errmsg["cause"] = "The service name does not exist or the service is abnormal"
                    self.write(errmsg)
                    self.set_status(400)
                    return
                else:
                    self.set_status(204)
                    return
        errmsg["code"] = "400010000"
        errmsg["message"] = "Missing required parameters"
        errmsg["cause"] = "Missing required parameters : service_name"
        self.write(errmsg)
        self.set_status(400)

class SysInfoHandler(BaseHandler):
    @run_on_executor
    @handler_try_except
    def get(self):
        sys_info = SystemAgent.get_sys_info()
        self.write(json.dumps(sys_info))
        self.set_status(200)
