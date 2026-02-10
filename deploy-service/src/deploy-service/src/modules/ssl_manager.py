#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# @Time    : 2021/4/10 13:47
# @Author  : Jimmy.li
# @Email   : jimmy.li@aishu.cn

import base64
import configparser
import datetime
from functools import lru_cache
import os
import re
import time

import M2Crypto
import M2Crypto.X509
import M2Crypto.RSA

import OpenSSL

from src.clients.cms import CMSClient
from src.common import global_param
from src.common.lib import exec_command
from src.common.log_util import logger
from src.lib.db.cert import Cert
from src.lib.kubernetes.k8s_api import K8SAPI

ROOT = "/tmp/cert_work"  # 证书根目录
OPENSSL_CONF_PATH = os.path.join(ROOT, "openssl.cnf")  # 证书配置文件路径
OPENSSL_CONF_SRC_CANDIDATES = [
    "/app/conf/default/anyshare/openssl.cnf",
    "/app/conf/default/cert/anyshare/openssl.cnf",
]
CA_KEY_PATH = os.path.join(ROOT, "ca.key")  # ca私钥路径
CA_CRT_PATH = os.path.join(ROOT, "ca.crt")  # 根ca证书路径
KEY_PATH = os.path.join(ROOT, "cert.key")  # 私钥路径
CSR_PATH = os.path.join(ROOT, "cert.csr")  # 请求生成证书
CRT_PATH = os.path.join(ROOT, "cert.crt")  # 证书路径
headers = {"Content-Type": "application/json"}


class IngoreCaseConf(configparser.ConfigParser):
    def __init__(self, defaults=None):
        configparser.ConfigParser.__init__(self, defaults=None)

    def optionxform(self, optionstr):
        return optionstr


class SSLManager(object):
    def _resolve_openssl_conf_src(self):
        for p in OPENSSL_CONF_SRC_CANDIDATES:
            if os.path.isfile(p):
                return p
        return ""

    def generate_openssl_cert(self, host):
        """
        调用openssl生成访问地址自签名证书
        """

        # 判断openssl.cnf 文件是否存在
        if not os.path.isfile(OPENSSL_CONF_PATH):
            if not os.path.isdir(ROOT):
                exec_mkdir, result = exec_command("mkdir -p %s" % ROOT)
                if not exec_mkdir:
                    return result, "", "", ""

            src_conf = self._resolve_openssl_conf_src()
            if not src_conf:
                return "openssl.cnf not found", "", "", ""

            exec_cp, result = exec_command("\cp %s %s" % (src_conf, ROOT))
            logger.info("cmd ['\\cp %s %s']" % (src_conf, ROOT))
            if not exec_cp:
                return result, "", "", ""

        # 检查ca证书是否到期
        ca_status = self.check_ca_crt_expried("app")

        # 1. 生成ca证书
        caFile = ""
        caKey = ""
        if ca_status:
            genrsa_ca = "openssl genrsa -out %s 2048" % CA_KEY_PATH
            logger.info("cmd ['openssl genrsa -out %s 2048']" % CA_KEY_PATH)
            exec_genrsa_ca, result = exec_command(genrsa_ca)
            if not exec_genrsa_ca:
                return result, "", "", ""

            try:
                os.chmod(CA_KEY_PATH, 0o600)
            except Exception:
                pass

            gencrt = (
                "\
                    openssl req -new -x509 -days 3650 -key {0} -out {1} -subj "
                '"/C=CN/L=Shanghai/O=Eisoo/OU=AnyShare/CN="ca.{2}.cn -extensions v3_ca -config {3}\
                    '.format(
                    CA_KEY_PATH, CA_CRT_PATH, "aishu", OPENSSL_CONF_PATH
                )
            )
            exec_gencrt, result = exec_command(gencrt)
            logger.info("cmd [" + gencrt + "]")
            if not exec_gencrt:
                return result, "", "", ""
        else:
            _, _, caFile, caKey, _ = self.read_cert_from_db("app")
            with open(CA_CRT_PATH, "w") as f:
                f.write(caFile)
            with open(CA_KEY_PATH, "w") as f:
                f.write(caKey)

        # 创建服务器证书私钥
        genrsa = "openssl genrsa -out %s 2048" % KEY_PATH

        exec_genrsa, result = exec_command(genrsa)
        logger.info("cmd [" + genrsa + "]")
        if not exec_genrsa:
            return result, "", "", ""

        try:
            os.chmod(KEY_PATH, 0o600)
        except Exception:
            pass

        # 生成服务器证书请求
        gencsr = '\
        openssl req -new -key {0} -out {1} -subj "/C=CN/L=Shanghai/O=Eisoo/OU=AnyShare/CN="{2}\
        '.format(
            KEY_PATH, CSR_PATH, host
        )

        # 将san写入openssl 配置文件
        config = IngoreCaseConf()
        config.read(OPENSSL_CONF_PATH)
        if self.is_valid_ip(host):
            option = "IP.1"
            config.remove_option("SAN", "DNS.1")
        else:
            option = "DNS.1"
            config.remove_option("SAN", "IP.1")
        config.set("SAN", option, host)
        config.write(open(OPENSSL_CONF_PATH, "w"))

        # 3. 使用ca根证书颁发服务器证书
        gencrt = (
            "\
        openssl x509 -req -days 3650 -in {0} -CA {1} -CAkey {2} -CAcreateserial -out {3} "
            "-extfile {4} -extensions v3_req -sha256\
        ".format(
                CSR_PATH, CA_CRT_PATH, CA_KEY_PATH, CRT_PATH, OPENSSL_CONF_PATH
            )
        )

        exec_gencsr, result = exec_command(gencsr)
        logger.info("cmd [" + gencsr + "]")
        if not exec_gencsr:
            return result, "", "", ""
        exec_gencrt, result = exec_command(gencrt)
        logger.info("cmd [" + gencrt + "]")
        if not exec_gencrt:
            return result, "", "", ""

        # 读取私钥和证书
        with open(CRT_PATH, "r") as f:
            certFile = f.read()

        with open(KEY_PATH, "r") as f:
            privateKey = f.read()
        # 读取ca根证书秘钥
        with open(CA_KEY_PATH, "r") as f:
            caKey = f.read()

        # 读取ca根证书
        with open(CA_CRT_PATH, "r") as f:
            caFile = f.read()

        return certFile, privateKey, caFile, caKey
    
    def init_global_https(self, ip="*"):
        """
        幂等的初始化证书
        """
        secret_list = K8SAPI().list_namespaced_secret_names()
        if global_param.INGRESS_SECRET_NAME in secret_list:
            return "",""
        
        return self.set_global_https(ip)

    def set_global_https(self, ip):
        """
        设置全局的nginx证书，使用AnyShare自签
        """
        # 检查cert_type

        logger.info("create app cert begin.")
        certFile, privateKey, caFile, caKey = self.generate_openssl_cert(ip)
        if not privateKey:
            logger.info("create app cert failed %s" % certFile)
            # todo: 错误码待定
            return "500017003", certFile
        self.storage_cert_ca(
            certFile=certFile, privateKey=privateKey, caFile=caFile, caKey=caKey, certSource="self-signed"
        )

        crt_content = certFile
        key_content = privateKey
        secret_list = K8SAPI().list_namespaced_secret_names()
        data = dict()
        data["tls.crt"] = base64.b64encode(crt_content.encode()).decode("utf-8")
        data["tls.key"] = base64.b64encode(key_content.encode()).decode("utf-8")
        if global_param.INGRESS_SECRET_NAME not in secret_list:
            K8SAPI().create_namespaced_secret(
                secret_name=global_param.INGRESS_SECRET_NAME, secret_type="tls", data=data
            )
            return "", ""
        K8SAPI().update_namespaced_secret(secret_name=global_param.INGRESS_SECRET_NAME, secret_type="tls", data=data)
        return "", ""

    def get_cert_info(self, cert_type=None):
        """
        获取证书的相关信息
        """

        result = self.get_info_of_cert(cert_type=cert_type)
        return "", result

    def upload_cert(self, certFile, privateKey):
        """
        设置全局的nginx存储证书，使用用户自己上传的证书
        """

        # 检查证书和私钥是否匹配
        status, mesg = self.is_cert_key_match(certFile, privateKey)
        if not status:
            logger.info("cert.crt not match cert.key")
            return mesg, mesg
        # 把证书保存到数据库中
        self.storage_cert_ca(certFile=certFile, privateKey=privateKey, certSource="customer-upload")
        crt_content = certFile
        key_content = privateKey
        secret_list = K8SAPI().list_namespaced_secret_names()
        data = dict()
        data["tls.crt"] = base64.b64encode(crt_content.encode()).decode("utf-8")
        data["tls.key"] = base64.b64encode(key_content.encode()).decode("utf-8")
        K8SAPI().update_namespaced_secret(secret_name=global_param.INGRESS_SECRET_NAME, secret_type="tls", data=data)
        return "", ""

    def download_cert(self):
        ca_content = Cert.get_value_content_by_key("ca_crt")
        return ca_content["f_value"]

    def check_ca_crt_expried(self, cert_type):
        certFile, privateKey, caFile, caKey, _ = self.read_cert_from_db(cert_type)
        status = True
        if caFile:
            certificate = OpenSSL.crypto.load_certificate(OpenSSL.crypto.FILETYPE_PEM, caFile)
            # 有效期过期日期
            tmpDate = certificate.get_notAfter().decode("utf-8")[0:8]
            # 获取生效日期时间戳
            end_date = datetime.datetime(int(tmpDate[0:4]), int(tmpDate[4:6]), int(tmpDate[6:8]))
            end_time = time.mktime(end_date.timetuple())
            # 是否有效
            status = False if end_time > time.time() else True

        return status

    def is_valid_ip(self, ip):
        ipregex = (
            "^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|"
            "2[0-4][0-9]|25[0-5])$"
        )
        pat = re.compile(ipregex)
        ret = pat.match(ip)
        if ret:
            return True
        else:
            return False

    def get_info_of_cert(self, cert_type=None):
        cert_info_list = []
        if not cert_type:
            for cert_type in ["app"]:
                certInfo = self.check_cert_info(cert_type)
                cert_info_list.append(certInfo)
        else:
            certInfo = self.check_cert_info(cert_type)
            cert_info_list.append(certInfo)

        return cert_info_list

    def check_cert_info(self, cert_type):
        certInfo = {
            "issuer": "",  # 证书颁发者
            "accepter": "",  # 证书接受者
            "startDate": "",  # 有效期开始日期
            "expireDate": "",  # 有效期过期日期
            "hasExpired": "",  # 是否过期
            "certType": cert_type,  # 证书类型
            "certSource": "",  # 证书来源
        }
        certFile, privateKey, caFile, caKey, certSource = self.read_cert_from_db(cert_type)
        certInfo["certSource"] = certSource or "unknown"
        if certFile:
            certificate = OpenSSL.crypto.load_certificate(OpenSSL.crypto.FILETYPE_PEM, certFile)
            subj = dict(certificate.get_subject().get_components())
            issuer = dict(certificate.get_issuer().get_components())
            # 颁发者
            certInfo["issuer"] = issuer.get(b"CN", b"").decode("utf-8")
            # 颁发给
            certInfo["accepter"] = subj.get(b"CN", b"").decode("utf-8")

            # 有效期开始日期 2015/01/01
            tmpDate = certificate.get_notBefore().decode("utf-8")

            certInfo["startDate"] = tmpDate[0:4] + "/" + tmpDate[4:6] + "/" + tmpDate[6:8]

            # 获取生效日期时间戳
            start_date = datetime.datetime(int(tmpDate[0:4]), int(tmpDate[4:6]), int(tmpDate[6:8]))
            start_time = time.mktime(start_date.timetuple())

            # 有效期过期日期 2016/01/01
            tmpDate = certificate.get_notAfter().decode("utf-8")[0:8]
            certInfo["expireDate"] = tmpDate[0:4] + "/" + tmpDate[4:6] + "/" + tmpDate[6:8]
            # 是否过期
            certInfo["hasExpired"] = certificate.has_expired()

        return certInfo

    def read_cert_from_db(self, cert_type):

        certs = Cert.get_https_all_content()
        certFile, privateKey, caFile, caKey, certSource = None, None, None, None, None

        for cert in certs:

            if cert["f_key"] == "ca_cert":
                certFile = cert["f_value"]

            if cert["f_key"] == "ca_private_key":
                privateKey = cert["f_value"]

            if cert["f_key"] == "ca_crt":
                caFile = cert["f_value"]

            if cert["f_key"] == "ca_key":
                caKey = cert["f_value"]

            if cert["f_key"] == "cert_source":
                certSource = cert["f_value"]

        return certFile, privateKey, caFile, caKey, certSource

    def storage_cert_ca(self, certFile=None, privateKey=None, caFile=None, caKey=None, certSource=None):
        logger.info("start storage cert begin.")
        keys = {}

        # 添加ca证书
        if caFile:
            keys["ca_crt"] = caFile

        # 添加ca秘钥
        if caKey:
            keys["ca_key"] = caKey

        # 添加应用服务器证书
        if certFile:
            keys["ca_cert"] = certFile

        # 添加应用服务器秘钥
        if privateKey:
            keys["ca_private_key"] = privateKey

        # 设置证书类型
        if certSource:
            keys["cert_source"] = certSource

        for key, value in keys.items():
            Cert.update_https_ca_content(key, value)

        logger.info("start storage cert done.")

    def is_cert_key_match(self, certFile, privateKey):
        """
        检查证书文件与私钥是否匹配
        """
        try:
            certificate = M2Crypto.X509.load_cert_string(certFile, M2Crypto.X509.FORMAT_PEM)
            privateKey = M2Crypto.RSA.load_key_string(privateKey.encode())

            # 从证书中获取公钥
            puk = certificate.get_pubkey().get_rsa()
            # 用公钥加密
            encrypted = puk.public_encrypt(b"data", M2Crypto.RSA.pkcs1_padding)
            # 用私钥解密，若成功则说明与证书匹配
            privateKey.private_decrypt(encrypted, M2Crypto.RSA.pkcs1_padding)
            return True, ""
        except Exception:
            return False, "400017029"



class CertDownloadFeature(object):
    cms_config_filed: str = "cert_download_feature"
    cms_config_name: str = "anyshare"


    @classmethod
    def get_status(cls):
        _min = int(time.time()) // 5
        return cls.get_cache_status(_min)

    @classmethod
    @lru_cache(maxsize=1)
    def get_cache_status(cls, _):
        cli = CMSClient()
        return cli.get_cms_data(cls.cms_config_name).real_data.get(cls.cms_config_filed, True)

    @classmethod
    def set_status(cls, status: bool):
        cli = CMSClient()
        cms_obj = cli.get_cms_data(cls.cms_config_name)
        cms_data = cms_obj.real_data
        cms_data[cls.cms_config_filed] = status
        cms_obj.real_data = cms_data
        cms_obj.save(cli)

