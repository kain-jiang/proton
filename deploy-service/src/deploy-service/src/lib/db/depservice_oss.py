#!/usr/lib/env python
# -*- coding:utf-8 -*-
import os
from base64 import urlsafe_b64decode

from M2Crypto import RSA

from src.lib.db.db_connector import get_db_operate_obj

TABLE_NAME = "depservice_oss"


class DepServiceOss(object):
    def __init__(self):
        pass

    @classmethod
    def insert_oss_info(cls, oss_info):
        db_oprator = get_db_operate_obj()

        colu = dict()
        colu["service_name"] = oss_info.get("service_name", "")
        colu["oss_name"] = oss_info.get("oss_name", "")

        db_oprator.insert(TABLE_NAME, colu)

    @classmethod
    def get_oss_info_by_service_name(cls, service):
        db_oprator = get_db_operate_obj()

        sql = "select * from depservice_oss where service_name=%s"

        result = db_oprator.fetch_one_result(sql, service)

        return result

    @classmethod
    def delete_oss_info_by_service_name(cls, service_name):
        db_oprator = get_db_operate_obj()

        sql = "delete from depservice_oss where service_name=%s"

        db_oprator.delete(sql, service_name)

    @classmethod
    def eisoo_rsa_decrypt(cls, data):
        try:
            return cls.eisoo_rsa_decrypt_new(data)
        except Exception:
            return cls.eisoo_rsa_decrypt_old(data)

    @classmethod
    def eisoo_rsa_decrypt_new(cls, data):
        """
        rsa解密， 先用safe_base64解密，再用私钥解密
        """
        private_key_2048_new = os.environ.get("DEPLOY_SERVICE_RSA_PRIVATE_KEY_2048_NEW", "")
        if not private_key_2048_new:
            raise Exception("missing rsa private key")

        private_key_2048 = RSA.load_key_string(private_key_2048_new.encode("utf-8"))
        # 去除所有的\r和\n
        data = data.replace(r"\n", "")
        data = data.replace(r"\r", "")
        data = urlsafe_b64decode(data)

        password = private_key_2048.private_decrypt(data, RSA.pkcs1_padding)
        if not password:
            raise Exception("decrypt failed")

        return bytes.decode(password)

    @classmethod
    def eisoo_rsa_decrypt_old(cls, data):
        """
        rsa解密， 先用safe_base64解密，再用私钥解密
        """
        private_key_1024_elder = os.environ.get("DEPLOY_SERVICE_RSA_PRIVATE_KEY_1024_OLD", "")
        if not private_key_1024_elder:
            raise Exception("missing rsa private key")

        private_key_1024 = RSA.load_key_string(private_key_1024_elder.encode("utf-8"))
        # 去除所有的\r和\n
        data = data.replace(r"\n", "")
        data = data.replace(r"\r", "")
        data = urlsafe_b64decode(data)

        password = private_key_1024.private_decrypt(data, RSA.pkcs1_padding)
        if not password:
            raise Exception("decrypt failed")

        return bytes.decode(password)
         