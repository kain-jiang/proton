#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import os
import sys

import rdsdriver

chart_name="ossgateway"
service_name="ManagementConsole"
chart_version="1.0.0-master"

def get_conn(user, password, host, port):
    """
    获取数据库的连接
    """
    try:
        conn = rdsdriver.connect(host=host,
                               port=int(port),
                               user=user,
                               passwd=password,
                               autocommit=True,
                               )
        # cursor = conn.cursor()
    except Exception as e:
        print("connect eofs error: %s", str(e))
        sys.exit(1)
    return conn


def update_micro_service(conn_cursor):
    """Migration of ossgateway chart info into deploy-serivce"""
    try:
        update_sql = "UPDATE deploy.chart SET chart_version='{0}' WHERE chart_name='{1}';".format(chart_version, chart_name)
        insert_sql = "INSERT INTO  deploy.chart (chart_version, chart_name, service_name) VALUES ('{0}', '{1}', '{2}');".format(chart_version, chart_name, service_name)
        get_old = "SELECT chart_name FROM deploy.chart WHERE chart_name='{0}';".format(chart_name)
        conn_cursor.execute(get_old)
        result = conn_cursor.fetchone()
        if result:
            conn_cursor.execute(update_sql)
        else:
            conn_cursor.execute(insert_sql)
    except Exception as exception:
        raise Exception(f"Migration of ossgateway chart info into deploy-serivce failed:msg:{exception}")


if __name__ == "__main__":
    conn = get_conn(os.environ["DB_USER"], os.environ["DB_PASSWD"], os.environ["DB_HOST"], os.environ["DB_PORT"])
    conn_cursor = conn.cursor()
    try:
        update_micro_service(conn_cursor)
    except Exception as ex:
        raise Exception(f"Migration of ossgateway chart info into deploy-serivce failed:msg:{ex}")
    finally:
        conn_cursor.close()
        conn.close()