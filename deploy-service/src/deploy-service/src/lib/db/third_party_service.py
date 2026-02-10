#!/usr/bin/env python3
# -*- coding:utf-8 -*-
from src.lib.db.db_connector import get_db_operate_obj

TABLE_NAME = "third_party_service"


class ThirdPartyService():
    def __init__(self):
        pass

    @classmethod
    def set_service(cls, service: str, data: list):
        db_operator = get_db_operate_obj()

        servers = []
        for config in data:
            server = config.get("server")
            protocol = config.get("protocol")
            host = config.get("host")
            port = config.get("port")
            servers.append(server)

            rel = db_operator.fetch_all_result(
                f"select protocol, host, port from {TABLE_NAME} where service=%s and server=%s", service, server
            )
            if rel:
                # update
                db_operator.update(
                    f"update {TABLE_NAME} set protocol=%s, host=%s, port=%s where service=%s and server=%s",
                    protocol, host, port, service, server
                )
            else:
                # insert
                db_operator.insert(TABLE_NAME,{
                    "service": service,
                    "server": server,
                    "protocol": protocol,
                    "host": host,
                    "port": port,
                })

            # 删除
        sql = f"delete from {TABLE_NAME} where service=%s and server not in ({','.join(len(servers) * ['%s'])})"
        db_operator.delete(sql, service, *servers)



    @classmethod
    def remove_service(cls, service: str):
        db_operator = get_db_operate_obj()

        sql = f"delete from {TABLE_NAME} where `service`=%s"

        db_operator.delete(sql, service)

    @classmethod
    def get_services(cls):
        db_operator = get_db_operate_obj()
        sql = f"select `service`, `server`, `protocol`, `host`, `port` from {TABLE_NAME} where 1=%s"

        return db_operator.fetch_all_result(sql, 1)

    @classmethod
    def get_service_by_service(cls, service: str):
        db_operator = get_db_operate_obj()

        sql = f"select `service`, `server`, `protocol`, `host`, `port` from {TABLE_NAME} where `service`=%s"

        return db_operator.fetch_all_result(sql, service)
