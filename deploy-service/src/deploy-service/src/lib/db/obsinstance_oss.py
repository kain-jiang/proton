#!/usr/lib/env python
# -*- coding:utf-8 -*-

import rsa
import base64

from Cryptodome.Cipher import AES
from src.common.log_util import logger
from src.lib.db.db_connector import get_db_operate_obj

TABLE_NAME = "obsinstance_oss"


class OBSInstanceOss(object):
    def __init__(self):
        pass

    @classmethod
    def insert_oss_info(cls, oss_info):
        db_oprator = get_db_operate_obj()

        colu = dict()
        colu["bucket"] = oss_info.get("name", "")
        colu["obs_id"] = oss_info.get("obs_id", "")
        colu["instance_id"] = oss_info.get("instance_id", "")

        db_oprator.insert(TABLE_NAME, colu)

    @classmethod
    def get_oss_bucket(cls):
        db_oprator = get_db_operate_obj()

        sql = "select bucket, obs_id  from obsinstance_oss"

        result = db_oprator.fetch_all_result(sql)

        return result

    @classmethod
    def aes_decrypt(self, content, secret_key):
        """AES解密 """
        try:
            cipher = AES.new(secret_key[:32].encode('utf-8'), AES.MODE_CBC, secret_key[:16].encode('utf-8'))
            content = base64.b64decode(content)
            text = cipher.decrypt(content)
            text = text.decode('utf-8')
            return text[:len(text)-ord(text[-1])]
        except Exception as ex:
            logger.error("error: %s", str(ex))
            raise Exception("ASE decrypt failed")


    @classmethod        
    def eisoo_rsa_encrypt(self, message):
        pub_key_1024 = """
    -----BEGIN PUBLIC KEY-----
    MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC7JL0DcaMUHumSdhxXTxqiABBC
    DERhRJIsAPB++zx1INgSEKPGbexDt1ojcNAc0fI+G/yTuQcgH1EW8posgUni0mcT
    E6CnjkVbv8ILgCuhy+4eu+2lApDwQPD9Tr6J8k21Ruu2sWV5Z1VRuQFqGm/c5vaT
    OQE5VFOIXPVTaa25mQIDAQAB
    -----END PUBLIC KEY-----
        """

        pub_key_2048 = """
    -----BEGIN PUBLIC KEY-----
    MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA5UaYuwHphUL1xAFr7mtd
    /dFA+X6MjApSjc2KL0KI5UlIUcGhTujIHsAdqtDSk8kHpeyb5zl6Y8NQsdnJ+Fg8
    /Yx1A29D6rFTSjFHbg//w12XX631QiDn+NStRsoMW9SLkvJScYtVyKngby3IbXpp
    J3o5ZST4ZqdenpJUWKU3rO1WDTGDCI1a9F+97YzUo3q5PFeoV2L/iLQ13+numCzH
    XXTfCj+PGyOiztNuc9/lDgObM73jmAXXHWC5cYBkFNsnsUGsWtZTpf4VHjuoTUfe
    laGkvQG4Ha3yVGLRsAnb0UjCUeFaQrbgADJo/BS1k+J6r9roSVUudmGj1lZ/qIFp
    TQIDAQAB
    -----END PUBLIC KEY-----
        """

        if not message:
            return ''
        pub_key = pub_key_2048
        pub_key = rsa.PublicKey.load_pkcs1_openssl_pem(pub_key.encode())
        cryptedMessage = rsa.encrypt(message.encode('utf-8'), pub_key)
        key_str_text = bytes.decode(base64.b64encode(cryptedMessage))
        return key_str_text
