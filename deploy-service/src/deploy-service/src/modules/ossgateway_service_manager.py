#!/usr/bin/env python3
# -*- coding:utf-8 -*-

import copy
import random
import socket
import time
import yaml
import base64
import traceback
from collections import Counter

from src.clients.oss import OssGatewayManagerClient, OSSTools
from src.lib.db.obsinstance_oss import OBSInstanceOss
from src.common import lib, node_utils, utils
from src.common.log_util import logger
from src.lib.db.depservice_oss import DepServiceOss
from src.lib.db.micro_service import MicroService
from src.lib.db.t_deployment_option import DeploymentOption
from src.lib.kubernetes import K8SAPI
from src.utils.net import get_host_for_url

from src.clients.k8s import K8SClient
from src.clients.helm import HelmClient
from src.clients.config import ConfigClient
from src.clients.protoncli import ProtonCliClient

INGRESS_RULE_TMP = {
    "apiVersion": "extensions/v1beta1",
    "kind": "Ingress",
    "metadata": {"name": "", "namespace": "", "annotations": {"kubernetes.io/ingress.class": ""}},
    "spec": {
        "rules": [
            {
                "http": {
                    "paths": [
                        # {
                        #     backend: {
                        #         "serviceName": "test-web1",
                        #         "servicePort": 1997
                        #     },
                        #     "path": "/ping",
                        #     "pathType": "ImplementationSpecific"
                        # }
                    ]
                }
            }
        ]
    },
}


class OSSGatewayManager:

    def __init__(self):
        self.core_v1_api = K8SClient.instance().sdk_client.core_v1_api
        self.apps_v1_api = K8SClient.instance().sdk_client.apps_v1_api
        self.namespace = K8SClient.instance().rest_client.self_namespace
        self.helm3 = HelmClient()
        self.success_nodes = []
        self.service_name = "ManagementConsole"
        self.chart = "ossgateway"
        self.chart_version = ConfigClient.load_config().ossgateway_version()
        self.internal_list = []


    def get_ingress_class(self) -> str:
        class443_values = self.helm3.get_current_values("class-443")
        return class443_values.get("depServices", {}).get("class-443", {}).get("ingressClass", "class-443")

    @classmethod
    def get_ossgateway_client(cls):
        return OssGatewayManagerClient.instance()


    def add_oss_config(self, service_name, data, flag=None):
        logger.info(f"add_oss_config start")
        # 检查是否存在同名的bucket
        if flag is None:
            if self._test_bucket_exist(data):
                logger.info(f"add_oss_config: bucket already exists")
                return "400017247"

        # 公有云安装仍按以前的方式部署
        # if utils.get_deploy_mode().get("mode", "standard") == "cloud":
        #     self.set_oss(data)
        # else:
        #     self._install_ossgateway_service(service_name, data)

        # 云端一体化不安装ossgageway
        if "ossgwMergeSize" in data:
            logger.info(f"add_oss_config: ossgwMergeSize in data, set oss")
            self.set_oss(data)
            return

        # 如果请求已经明确指出不安装OSS网关，则不安装
        if not data.get("hasOSSGW", False):
            logger.info(f"add_oss_config: hasOSSGW is False, set oss")
            self.set_oss(data)
            return

        # 只有数据存储需要安装oss gateway
        is_cache_bucket = data.get("isCacheBucket", False)
        if not is_cache_bucket:
            logger.info(f"add_oss_config: install ossgateway service")
            code = self._install_ossgateway_service(service_name, data, flag)
            if code:
                return code
        # 配置对象存储配置
        try:
            logger.info(f"add_oss_config: set oss")
            self.set_oss(data)

            # 设置内部ip
            if not is_cache_bucket:
                logger.info(f"add_oss_config: create internal addresses")
                self.get_ossgateway_client().CreateInternalAddresses(
                    bucket_name=data["bucketInfo"]["name"],
                    vendor_type=data["bucketInfo"]["provider"],
                    internal_list=self.internal_list,
                )
            logger.info(f"add_oss_config success")
        except Exception as e:
            if not is_cache_bucket:
                try:
                    self.uninstall_ossgateway_service(data)
                except Exception as exp:
                    raise exp from e
            raise

    def set_storage_prefix(self, storage_prefix):
        self.get_ossgateway_client().SetStoragePrefix(storage_prefix)

    def _install_ossgateway_service(self, service_name, data, flag=None):
        logger.info(f"_install_ossgateway_service start")
        # 安装前检查
        micro_service_name = self.get_micro_service_name(data["bucketInfo"]["name"])

        # 添加oss网关内部地址
        data["ossgwInfo"] = {}
        data["ossgwInfo"]["ossgwHost"] = self.get_ossgateway_internal_host(micro_service_name)

        bucket_name = data["bucketInfo"]["name"]
        logger.info(f"_install_ossgateway_service: query_ossgateway_releases begin, bucket_name: {bucket_name}")
        ossgateway_releases = self.query_ossgateway_releases(self.helm3.get_all_releases())
        logger.info(f"_install_ossgateway_service: query_ossgateway_releases end. ossgateway_releases: {ossgateway_releases}")
        micro_services_name = []
        for release in ossgateway_releases:
            micro_services_name.append(release["name"])
            if micro_service_name == release["name"]:
                msg = f"service[{micro_service_name}] already installed, skip."
                logger.info(msg)
                raise Exception(msg)

        service_name = self.service_name

        charts = [self.chart]

        # 无意义代码，去除
        # service_info = ContainerizedService.get_service_info(service_name)
        # if not service_info or not service_info["available_version"]:
        #     error = "no available package."
        #     raise Exception(error)

        # 检查micro_service_name是否被使用
        if self.helm3.get_release_info(micro_service_name) is not None:
            logger.info(f"_install_ossgateway_service: release name '{micro_service_name}' is used")
            raise Exception(f"release name '{micro_service_name}' is used")

        old_ossgateway_cms_data = utils.get_ossgateway_info()
        logger.info(f"old ossgateway info: [{old_ossgateway_cms_data}]")

        oss_nodes_ip, oss_nodes_name = self.select_oss_node()
        oss_nodes_port = self._get_free_port(micro_services_name, old_ossgateway_cms_data, oss_nodes_ip)
        data["ossgwInfo"]["ossgwPort"] = oss_nodes_port[0]
        # 给OSSGateway添加额外的端口
        oss_nodes_port = oss_nodes_port + self._get_free_port(
            micro_services_name, old_ossgateway_cms_data, oss_nodes_ip, 9500
        )
        logger.info(
            f"ossgateway will install on node [{oss_nodes_name}], node ip:[{oss_nodes_ip}], ports:[{oss_nodes_port}]"
        )

        # 成功安装的微服务
        installed_services = list()
        # 成功设置标签的节点
        self.success_nodes = list()
        # 绑定bucket和oss网关标识
        success_bind_ossgateway_bucket = False
        # 插入数据库成功标识
        success_insert_database = False

        inter_oss_nodes_ip = self._get_oss_inter_nodes_ip(oss_nodes_ip)
        logger.info(
            f"Get inter ip [{inter_oss_nodes_ip}]"
        )

        # 设置内部ip
        if not inter_oss_nodes_ip:
            oss_ips = oss_nodes_ip
        else:
            oss_ips = inter_oss_nodes_ip
        oss_port = oss_nodes_port[0]
        self.internal_list = [get_host_for_url(ip) + ":" + str(oss_port) for ip in oss_ips]

        try:
            # 给选择的节点打标签
            for node_name in oss_nodes_name:
                logger.info(f"_install_ossgateway_service: tag label to node {node_name}")
                self._tag_label_to_node(node_name, micro_service_name)
                self.success_nodes.append(node_name)
                logger.info(f"_install_ossgateway_service: tag label to node {node_name} success")

            # 安装服务
            logger.info(f"_install_ossgateway_service: install service begin")
            service_conf = lib.serviceConf()
            for chart in charts:
                service_conf["service_name"] = service_name
                service_conf["node_ips"] = inter_oss_nodes_ip
                if not inter_oss_nodes_ip:
                    service_conf["node_ips"] = oss_nodes_ip
                service_conf["ports"] = oss_nodes_port
                if flag == "AB_Restore":
                    service_conf["read_only"] = "true"
                logger.info(service_conf)
                self.helm3.install_or_upgrade_any(
                    name=micro_service_name,
                    config=self.oss_gateway_config(micro_service_name, service_conf),
                    chart_name=self.chart,
                    chart_version=self.chart_version,
                )
                installed_services.append(chart)
            logger.info(f"_install_ossgateway_service: install service end")
            # 更新cms配置
            new_cms_ossgateway_data = self.update_ossgateway_cms(
                micro_services_name,
                old_ossgateway_cms_data,
                micro_service_name,
                bucket_name,
                oss_nodes_ip,
                inter_oss_nodes_ip,
                oss_nodes_port,
            )

            # 更新ingress路由
            self.update_ossgateway_ingress(new_cms_ossgateway_data)

            # 绑定实例
            # if flag == "AB_Restore":
            #     self.unbind_ossgateway_bucket(data)
            service_conf["node_ips"] = oss_nodes_ip
            self.bind_ossgateway_bucket(data, service_conf)
            success_bind_ossgateway_bucket = True

            # 更新数据库
            # micro_service_infos = list()
            # for chart in installed_services:
            #     external_port = DeploymentOption.get_option(lib.SERVICE_EXTERNAL_PORT_KEY)
            #     micro_service_version = self.chart_version
            #     micro_service_info = dict()
            #     micro_service_info["micro_service_name"] = micro_service_name
            #     micro_service_info["service_name"] = "OSSGatewayService"
            #     micro_service_info["micro_service_version"] = micro_service_version

            #     micro_service_info["internal_port"] = 0
            #     micro_service_info["external_port"] = external_port
            #     micro_service_info["need_ingress"] = 1
            #     micro_service_infos.append(micro_service_info)
            # logger.info("insert_many_service micro_service_info[%s]." % micro_service_infos)
            # MicroService.insert_many_service(micro_service_infos)
            # replicas = len(service_conf["node_ips"])
            # ContainerizedService.update_replicas(service_name, replicas)
            # ContainerizedService.update_installed_version(service_name, service_info["available_version"])
            # success_insert_database = True

        except Exception as ex:
            logger.error("error: %s, try rollback", str(ex))
            try:
                # 回滚
                logger.info(
                    "install service[%s] failed, rollback. err:%s" % (micro_service_name, str(ex).encode("utf8"))
                )
                if success_bind_ossgateway_bucket:
                    self.unbind_ossgateway_bucket(data)

                # if success_insert_database:
                #     MicroService.delete_micro_service(micro_service_name)

                self.helm3.uninstall(micro_service_name)

                for node_name in self.success_nodes:
                    self._cancle_label_to_node(node_name, micro_service_name)
                # 回滚ingress(不回滚)
                # 回滚cms(不回滚)
                if hasattr(ex, "errID") and ex.errID == 10115:
                    logger.info(str(ex.expMsg).encode("utf8"))
                    return "400017250"
                if hasattr(ex, "errID") and ex.errID == 10116:
                    logger.info(str(ex.expMsg).encode("utf8"))
                    return "400017249"
                error = "install service[%s] failed, err:%s" % (micro_service_name, str(traceback.format_exc()))
                raise Exception(error)
            except Exception as ex:
                logger.error("error: %s", str(ex))
            raise

    def get_micro_service_name(self, bucket: str):
        return "ossgateway-" + bucket.replace("_", "-").lower()

    def query_ossgateway_releases(self, releases: list[dict]) -> list[dict]:
        return [release for release in releases if release["name"].startswith("ossgateway-") ]

    def get_ossgateway_internal_host(self, server_name):
        return server_name + ".{}.svc.cluster.local".format(self.namespace)

    def uninstall_ossgateway_service(self, data: dict):
        micro_service_name = self.get_micro_service_name(data["bucketInfo"]["name"])
        logger.info("uninstall service[%s]", micro_service_name)
        self.unbind_ossgateway_bucket(data)
        # MicroService.delete_micro_service(micro_service_name)
        self.helm3.uninstall(micro_service_name)

        for node_name in self.success_nodes:
            self._cancle_label_to_node(node_name, micro_service_name)

    def update_oss_config(self, service_name, data):
        logger.info(f"update_oss_config start, service_name: {service_name}")
        if not self._test_bucket_exist(data):
            msg = f"not allow to modify bucket."
            logger.info(msg)
            raise Exception(msg)

        # 公有云安装仍按以前的方式部署
        # if utils.get_deploy_mode().get("mode", "standard") == "cloud":
        #     self.set_oss(data)
        # else:
        #     code = self._update_oss_config(service_name, data)
        #     if code: return code
        is_cache_bucket = data.get("isCacheBucket", False)
        has_oss_gateway = data.get("hasOSSGW", False)
        if is_cache_bucket or not has_oss_gateway:
            logger.info(f"update_oss_config: is_cache_bucket or not has_oss_gateway, set oss")
            self._update_cache_oss_config(data)
            logger.info(f"update_oss_config end")
        else:
            logger.info(f"update_oss_config: _update_oss_config, service_name: {service_name}")
            code = self._update_oss_config(service_name, data)
            if code:
                return code

    def _update_cache_oss_config(self, data):
        self.set_oss(data)

    def _update_oss_config(self, service_name, data):
        logger.info(f"_update_oss_config start.")
        micro_service_name = "ossgateway-" + data["bucketInfo"]["name"].replace("_", "-").lower()

        # 修改oss网关内部地址
        data["ossgwInfo"] = {}
        data["ossgwInfo"]["ossgwHost"] = self.get_ossgateway_internal_host(micro_service_name)

        service_installed = False
        logger.info(f"_update_oss_config: query_ossgateway_releases begin.")
        ossgateway_releases = self.query_ossgateway_releases(self.helm3.get_all_releases())
        logger.info(f"_update_oss_config: query_ossgateway_releases end. ossgateway_releases: {ossgateway_releases}")
        for release in ossgateway_releases:
            if micro_service_name == release["name"]:
                service_installed = True
                break
        if not service_installed:
            msg = f"service[{micro_service_name}] not installed."
            logger.info(msg)
            return "400017248"

        ossgateway_cms_data = utils.get_ossgateway_info()
        logger.info(f"_update_oss_config: get_ossgateway_info end. ossgateway_cms_data: {ossgateway_cms_data}")
        data["ossgwInfo"]["ossgwPort"] = ossgateway_cms_data[micro_service_name]["ports"][0]
        service_conf = lib.serviceConf()
        service_conf["node_ips"] = ossgateway_cms_data[micro_service_name]["ips"]
        service_conf["ports"] = ossgateway_cms_data[micro_service_name]["ports"]
        logger.info(service_conf)
        bind_ossgateway_bucket_success = False

        # 如果是clouhub配置的需要设置merge_size的值为0
        cloud_hub_buckets = OBSInstanceOss.get_oss_bucket()
        for bucket_info in cloud_hub_buckets:
            if data["bucketInfo"]["name"] == bucket_info["bucket"]:
                data["ossgwMergeSize"] = 0
                logger.info(f"_update_oss_config: is cloud hub bucket, set ossgwMergeSize to 0")

        try:
            logger.info(f"_update_oss_config: bind_ossgateway_bucket begin, data: {data}, service_conf: {service_conf}")
            self.bind_ossgateway_bucket(data, service_conf)
            logger.info(f"_update_oss_config: bind_ossgateway_bucket end")
            bind_ossgateway_bucket_success = True
            self._pod_restart(self.namespace, micro_service_name)
            logger.info(f"_update_oss_config: pod_restart end, namespace: {self.namespace}, micro_service_name: {micro_service_name}")
            self.set_oss(data)
            logger.info(f"_update_oss_config: set_oss end")
        except Exception as ex:
            # 回滚
            logger.info("update service[%s] failed, rollback. err:%s" % (micro_service_name, str(ex).encode("utf8")))
            if bind_ossgateway_bucket_success:
                old_data = self.get_old_oss_data(data["bucketInfo"]["name"])
                self.bind_ossgateway_bucket(old_data, service_conf)
                if hasattr(ex, "errID"):
                    if ex.errID == 10115:
                        logger.info(str(ex.expMsg).encode("utf8"))
                        return "400017250"
                    if ex.errID == 10116:
                        logger.info(str(ex.expMsg).encode("utf8"))
                        return "400017249"
            error = "update service[%s] failed, err:%s" % (micro_service_name, str(traceback.format_exc()))
            raise Exception(error)

    def install_abrestore_ossgateway(self, ab_datas):
        """AB恢复AS的过程中批量安装OSS网关"""
        logger.info(f"install_abrestore_ossgateway start")
        ossinfos = self.get_ossgateway_client().GetOSSInfo()
        logger.info(f"install_abrestore_ossgateway: GetOSSInfo end. ossinfos: {ossinfos}")
        # 检查是否有重复bucket的情况
        self._check_reduplicate_bucket(ossinfos)

        ossgateway_releases = self.query_ossgateway_releases(self.helm3.get_all_releases())
        logger.info(f"install_abrestore_ossgateway: query_ossgateway_releases end. ossgateway_releases: {ossgateway_releases}")
        micro_services_name = [release["name"] for release in ossgateway_releases]
        logger.info(f"install_abrestore_ossgateway: micro_services_name end. micro_services_name: {micro_services_name}")
        buckets = [ossinfo["bucketInfo"]["name"] for ossinfo in ossinfos]
        for ab_data in ab_datas:
            # 考虑AB重复安装恢复的情况下，t_oss_conf表中已经有同名存储的情况
            if ab_data["bucket"] in buckets:
                for ossinfo in ossinfos:
                    if ossinfo["bucketInfo"]["name"] == ab_data["bucket"]:
                        micro_service_name = "ossgateway-" + ossinfo["bucketInfo"]["name"].replace("_", "-")
                        # 检查是否已经安装了OSS网关，如果安装则跳过
                        if micro_service_name in micro_services_name:
                            msg = f"service[{micro_service_name}] already installed, skip."
                            logger.info(msg)
                            continue
                        self._generate_sefossinfo_data(ossinfo, ab_data)
                        ossinfo["hasOSSGW"] = True
                        self.add_oss_config("OSSGatewayService", ossinfo, "AB_Restore")
            else:
                data = dict()
                data["storageName"] = ab_data["bucket"]
                data["storageId"] = ""
                data["enabled"] = True
                data["bucketInfo"] = dict()
                self._generate_sefossinfo_data(data, ab_data)
                data["bucketInfo"]["provider"] = ab_data["provider"]
                data["bucketInfo"]["name"] = ab_data["bucket"]
                data["hasOSSGW"] = True
                self.add_oss_config("OSSGatewayService", data, "AB_Restore")

    def upgrade_ossgateway_versoin(self):
        """AnyShare2.2升级到2.3版本,处理2.2中未安装oss网关的场景,更加已添加的对象存储安装OSS网关"""
        logger.info(f"upgrade_ossgateway_versoin start")
        ossinfos = self.get_ossgateway_client().GetOSSInfo()
        logger.info(f"upgrade_ossgateway_versoin: GetOSSInfo end. ossinfos: {ossinfos}")
        # 检查是否有重复bucket的情况
        self._check_reduplicate_bucket(ossinfos)
        ossgateway_releases = self.query_ossgateway_releases(self.helm3.get_all_releases())
        micro_services_name = [release["name"] for release in ossgateway_releases]
        logger.info(f"upgrade_ossgateway_versoin: micro_services_name end. micro_services_name: {micro_services_name}")
        for ossinfo in ossinfos:
            micro_service_name = "ossgateway-" + ossinfo["bucketInfo"]["name"].replace("_", "-")
            # 检查是否已经安装了OSS网关，如果安装则跳过
            if micro_service_name in micro_services_name:
                msg = f"service[{micro_service_name}] already installed, skip."
                logger.info(msg)
                continue
            self.add_oss_config("OSSGatewayService", ossinfo, "upgrade")
        logger.info(f"upgrade_ossgateway_versoin end")
    def upgrade_ossgateway_service(self, service_name):
        """ossgateway升级"""
        logger.info(f"upgrade_ossgateway_service start")
        ossinfos = self._get_oss_info()
        logger.info(f"upgrade_ossgateway_service: _get_oss_info end. ossinfos: {ossinfos}")
        ossgateway_releases = self.query_ossgateway_releases(self.helm3.get_all_releases())
        micro_services_name = [release["name"] for release in ossgateway_releases]
        logger.info(f"upgrade_ossgateway_service: query_ossgateway_releases end. micro_services_name: {micro_services_name}")
        if not micro_services_name:
            return
        service_name = self.service_name
        # charts = Chart.get_charts_by_service(service_name)
        # chart = charts[0]
        chart = self.chart
        ossgateway_cms_data = utils.get_ossgateway_info()
        # 升级7.0.3.4 修改ingress配置
        # self.update_ossgateway_ingress(ossgateway_cms_data)
        logger.info(f"upgrade_ossgateway_service: chart: {chart} ossgateway_cms_data: {ossgateway_cms_data}")
        for micro_service_name in micro_services_name:
            service_conf = lib.serviceConf()
            service_conf["service_name"] = service_name

            if "inter_ips" in ossgateway_cms_data[micro_service_name] and ossgateway_cms_data[micro_service_name]["inter_ips"]:
                service_conf["node_ips"] = ossgateway_cms_data[micro_service_name]["inter_ips"]
            else:
                inter_oss_nodes_ip = self._get_oss_inter_nodes_ip(ossgateway_cms_data[micro_service_name]["ips"])
                if inter_oss_nodes_ip:
                    ossgateway_cms_data[micro_service_name]["inter_ips"] = inter_oss_nodes_ip
                    service_conf["node_ips"] = inter_oss_nodes_ip
                else:
                    service_conf["node_ips"] = ossgateway_cms_data[micro_service_name]["ips"]

            service_conf["ports"] = ossgateway_cms_data[micro_service_name]["ports"]
            # ossgateway添加了端口，如果检测到端口是3个，升级需要添加端口
            if len(service_conf["ports"]) == 3:
                service_conf["ports"] = service_conf["ports"] + list(
                    self._get_free_port(micro_services_name, ossgateway_cms_data, service_conf["node_ips"], 9500)
                )
                ossgateway_cms_data[micro_service_name]["ports"] = service_conf["ports"]
                logger.info(f"upgrade_ossgateway_service: save_ossgateway_info end. ossgateway_cms_data: {ossgateway_cms_data}")
                utils.save_ossgateway_info(ossgateway_cms_data)
                logger.info(f"upgrade_ossgateway_service: save_ossgateway_info success")
            logger.info(f"upgrade_ossgateway_service: update_ossgateway_ingress begin")
            self.update_ossgateway_ingress(ossgateway_cms_data)
            logger.info(f"upgrade_ossgateway_service: update_ossgateway_ingress end")
            utils.save_ossgateway_info(ossgateway_cms_data)
            logger.info(f"upgrade_ossgateway_service: save_ossgateway_info success, ossgateway_cms_data: {ossgateway_cms_data}")
            # 升级7.0.4.3添加oss网关内部ip
            for ossinfo in ossinfos:
                data = ossinfo
                if micro_service_name == "ossgateway-" + data["bucketInfo"]["name"].replace("_", "-").lower():
                    # 设置内部ip
                    oss_ips = service_conf["node_ips"] 
                    oss_port = ossgateway_cms_data[micro_service_name]["ports"][0]
                    self.internal_list = [get_host_for_url(ip) + ":" + str(oss_port) for ip in oss_ips]
                    self.get_ossgateway_client().CreateInternalAddresses(
                        bucket_name=data["bucketInfo"]["name"],
                        vendor_type=data["bucketInfo"]["provider"],
                        internal_list=self.internal_list,
                    )
                    logger.info(f"upgrade_ossgateway_service: CreateInternalAddresses end")
            logger.info(f"upgrade_ossgateway_service: install or upgrade any begin, micro_service_name: {micro_service_name}, chart_version:{self.chart_version}")
            self.helm3.install_or_upgrade_any(
                name=micro_service_name,
                config=self.oss_gateway_config(micro_service_name, service_conf),
                chart_name=self.chart,
                chart_version=self.chart_version,
            )
            logger.info(f"upgrade_ossgateway_service: install or upgrade any end")
            # micro_service_version = self.chart_version
            # MicroService.update_service_version(micro_service_name, micro_service_version)

    def select_oss_node(self):
        all_nodes_info = node_utils.get_all_node_info()
        online_nodes_info = [node_info for node_info in all_nodes_info if node_info["node_online"]]
        oss_nodes_ip = list()
        oss_nodes_name = list()
        select_nodes_info = list()

        if len(all_nodes_info) == 0:
            error = "not found node to install "
            logger.info(error)
            raise Exception(error)
        # ossgateway目前安装节点选择：当节点小于3时,随机选择一个节点，当节点大于等于3时候，随机选择3个节点
        if len(all_nodes_info) < 3:
            if len(online_nodes_info) >= 1:
                select_nodes_info = random.sample(online_nodes_info, 1)
            else:
                error = "not have online node to install"
                logger.info(error)
                raise Exception(error)

        if len(all_nodes_info) >= 3:
            if len(online_nodes_info) >= 3:
                select_nodes_info = random.sample(online_nodes_info, 3)
            else:
                error = f"not have enough online node to install ,only find {len(online_nodes_info)} online node"
                logger.info(error)
                raise Exception(error)

        for node_info in select_nodes_info:
            oss_nodes_ip.append(node_info["node_ip"])
            oss_nodes_name.append(node_info["node_name"])

        return oss_nodes_ip, oss_nodes_name

    def addr_is_open(self, ip, port):
        # type: (str, int) -> bool
        """
        创建 socket 连接，检测地址是否可连
        :param ip: 需要检测的 IP 地址
        :param port: 需要检测的端口
        :return: 可连返回 True，否则返回 False
        """
        try:
            ss = socket.create_connection((ip, int(port)), timeout=5)  # type: socket
            ss.shutdown(2)
            ss.close()
            return True
        except socket.error:
            return False

    def _get_free_port(self, micro_services_name, old_ossgateway_cms_data, oss_nodes_ip, start_port=9000):
        """获取9000到10000中的3个未使用的连续端口, 目前AS网关用不到这么多端口，暂时不检测端口溢出混乱"""
        used_ports = [
            old_ossgateway_cms_data[micro_service_name]["ports"] for micro_service_name in micro_services_name
        ]
        for port in range(int(start_port), 10000, 3):
            for used_port in used_ports:
                if port in used_port:
                    break
            else:
                for oss_node_ip in oss_nodes_ip:
                    if self.addr_is_open(oss_node_ip, int(port)) or self.addr_is_open(
                        oss_node_ip, int(port) + 1) or self.addr_is_open(oss_node_ip, int(port) + 2):
                        break
                else:
                    return int(port), int(port) + 1, int(port) + 2
        else:
            error = "can't find usable ip"
            logger.info(error)
            raise Exception(error)

    def _tag_label_to_node(self, node_name, release_name):
        label_key = "ossgateway/%s" % release_name
        label = {label_key: "OSSGateway"}
        K8SAPI().set_node_label(node_name, label)

    def _cancle_label_to_node(self, node_name, release_name):
        label_key = "ossgateway/%s" % release_name
        K8SAPI().remove_node_label(node_name, label_key)


    def _generate_sefossinfo_data(self, data, ab_data):
        data["bucketInfo"]["providerDetail"] = ab_data.get("providerDetail", "")
        data["bucketInfo"]["accessId"] = ab_data["accessId"]
        data["bucketInfo"]["accessKey"] = ab_data["accessKey"]
        data["bucketInfo"]["serverName"] = ab_data["serverName"]
        data["bucketInfo"]["internalServerName"] = ab_data.get("internalServerName", "")
        data["bucketInfo"]["httpsPort"] = ab_data["httpsPort"]
        data["bucketInfo"]["httpPort"] = ab_data["httpPort"]
        data["bucketInfo"]["cdnName"] = ab_data.get("cdnName", "")
        data["bucketInfo"]["region"] = ab_data.get("region", "")
        # OSSGatewayManager REST接口已经移除bucketStyle，此处假设其已不再需要
        # data["accessInfo"]["bucketStyle"] = ab_data.get("bucketStyle", "")


    def _test_bucket_exist(self, data):
        ossinfos: list = self.get_ossgateway_client().GetOSSInfo()
        cacheossinfos: list = self.get_ossgateway_client().GetCacheOSSInfo()
        ossinfos.extend(cacheossinfos if cacheossinfos else [])
        for ossinfo in ossinfos:
            if data["bucketInfo"]["name"] == ossinfo["bucketInfo"]["name"]:
                return True
        return False

    def set_oss(self, data):
        if data["storageId"] is None or data["storageId"] == "":
            logger.info(f"set_oss: AddOSSInfo begin, data: {data}")
            self.get_ossgateway_client().AddOSSInfo(data)
            logger.info(f"set_oss: AddOSSInfo end")
        else:
            logger.info(f"set_oss: ModifyExistingOSSInfo begin, data: {data}")
            self.get_ossgateway_client().ModifyExistingOSSInfo(data)
            logger.info(f"set_oss: ModifyExistingOSSInfo end")

    def bind_ossgateway_bucket(self, data, service_conf):
        url_list = ["%s:%s" % (get_host_for_url(ip), service_conf["ports"][0]) for ip in service_conf["node_ips"]]
        self.get_ossgateway_client().BindBucket(
            bucket_name=data["bucketInfo"]["name"],
            access_key=DepServiceOss.eisoo_rsa_decrypt(data["bucketInfo"]["accessKey"]),
            access_key_id=data["bucketInfo"]["accessId"],
            url=OSSTools.to_url(
                host=data["bucketInfo"]["serverName"],
                http_port=data["bucketInfo"].get("httpPort"),
                https_port=data["bucketInfo"].get("httpsPort"),
            ),
            internal_url=OSSTools.to_url(
                host=data["bucketInfo"]["internalServerName"],
                http_port=data["bucketInfo"].get("httpPort"),
                https_port=data["bucketInfo"].get("httpsPort"),
            ),
            url_list=url_list,
            vendor_type=data["bucketInfo"]["provider"],
            region=data["bucketInfo"].get("region", ""),
            size=data.get("ossgwMergeSize", None)
        )

    def unbind_ossgateway_bucket(self, data):
        self.get_ossgateway_client().UnbindBucket(
            bucket_name=data["bucketInfo"]["name"],
            vendor_type=data["bucketInfo"]["provider"],
        )

    def update_ossgateway_cms(
        self,
        micro_services_name,
        old_ossgateway_cms_data,
        micro_service_name,
        bucket_name,
        oss_nodes_ip,
        inter_oss_nodes_ip,
        oss_nodes_port,
    ):
        new_cms_ossgateway_data = dict()
        for tmp_micro_service_name in micro_services_name:
            if tmp_micro_service_name in old_ossgateway_cms_data.keys():
                new_cms_ossgateway_data[tmp_micro_service_name] = old_ossgateway_cms_data[tmp_micro_service_name]
        new_cms_ossgateway_data[micro_service_name] = {
            "bucket": bucket_name,
            "ips": oss_nodes_ip,
            "inter_ips": inter_oss_nodes_ip,
            "ports": oss_nodes_port,
        }
        utils.save_ossgateway_info(new_cms_ossgateway_data)
        return new_cms_ossgateway_data

    def update_ossgateway_ingress(self, new_cms_ossgateway_data):
        req_data = self.get_ingress_rule_body(new_cms_ossgateway_data)
        K8SClient().rest_client.save_ingress(
            name=req_data["metadata"]["name"],
            namespace=req_data["metadata"]["namespace"],
            paths=req_data["spec"]["rules"][0]["http"]["paths"],
            annotations=req_data["metadata"]["annotations"],
        )

    def get_ingress_rule_body(self, new_cms_ossgateway_data):
        rule = list()
        for release_name in new_cms_ossgateway_data.keys():
            backend = {
                "extensions/v1beta1": {
                    "serviceName": release_name,
                    "servicePort": new_cms_ossgateway_data[release_name]["ports"][0],
                },
                "networking.k8s.io/v1": {
                    "service": {
                        "name": release_name,
                        "port": {"number": int(new_cms_ossgateway_data[release_name]["ports"][0])},  # must be int
                    }
                },
            }.get(K8SClient().rest_client.ingress_api_version)
            tmp_rule = {
                "backend": backend,
                "path": "/%s" % new_cms_ossgateway_data[release_name]["bucket"],
                "pathType": "ImplementationSpecific",
            }
            logger.info(tmp_rule)
            rule.append(tmp_rule)
        req_data = copy.deepcopy(INGRESS_RULE_TMP)
        req_data["metadata"]["name"] = "rule-1443"
        req_data["metadata"]["namespace"] = self.namespace
        req_data["metadata"]["annotations"]["kubernetes.io/ingress.class"] = self.get_ingress_class()
        self._add_cors_header(req_data)
        req_data["spec"]["rules"][0]["http"]["paths"] = rule
        logger.info(req_data)
        return req_data

    def _add_cors_header(self, ingress_rule_data):
        annotations = {
            "nginx.ingress.kubernetes.io/cors-allow-origin": "*",
            "nginx.ingress.kubernetes.io/cors-allow-methods": "GET,PUT,POST,DELETE,HEAD,OPTIONS,always",
            "nginx.ingress.kubernetes.io/cors-allow-headers": "DNT,Keep-Alive,User-Agent,X-Requested-With,"
            "If-Modified-Since,Cache-Control,Content-Type,"
            "Authorization,Location,pragma,Range,x-amz-date,always",
            "nginx.ingress.kubernetes.io/cors-allow-credentials": "true",
            "nginx.ingress.kubernetes.io/enable-cors": "true",
            "nginx.ingress.kubernetes.io/configuration-snippet": "more_set_headers 'Access-Control-Expose-Headers: Location,Content-Range,Content-Length,Accept-Ranges,Etag,always';",
            "nginx.ingress.kubernetes.io/proxy-max-temp-file-size": "0",
            "nginx.ingress.kubernetes.io/proxy-request-buffering": "off",
            "nginx.ingress.kubernetes.io/proxy-next-upstream": "error timeout http_502",
            "nginx.ingress.kubernetes.io/proxy-next-upstream-tries": "3",
        }
        for k, v in annotations.items():
            ingress_rule_data["metadata"]["annotations"][k] = v

    def get_old_oss_data(self, bucket):
        ossinfos = self.get_ossgateway_client().GetOSSInfo()
        for ossinfo in ossinfos:
            if ossinfo["bucketInfo"]["name"] == bucket:
                return ossinfo

    def _check_reduplicate_bucket(self, ossinfos):
        """检查是否有重复的bucket"""
        bucket_list = [ossinfo["bucketInfo"]["name"] for ossinfo in ossinfos]
        bucket_counter = dict(Counter(bucket_list))
        bucket_reduplicate = [name for name, counter in bucket_counter.items() if counter > 1]
        if bucket_reduplicate:
            msg = f"reduplicate_bucke {bucket_reduplicate}."
            logger.info(msg)
            raise Exception(msg)

    def _pod_restart(self, namespace, filter_value):
        old_pod_names = self._filter_pods(namespace, filter_value)
        for pod_name, v in old_pod_names.items():
            self._delete_pod(namespace, pod_name)
            logger.info(f"_pod_restart: delete_pod end, namespace: {namespace}, pod_name: {pod_name}")
        logger.info("pod reload sucessfully")
        return old_pod_names

    def _filter_pods(self, namespace, filter_value):
        api = self.core_v1_api
        pod_list = api.list_namespaced_pod(namespace=namespace)
        pod_names = dict()
        for pod in pod_list.items:
            if pod.metadata.name.find(filter_value) >= 0:
                pod_names[pod.metadata.name] = pod.status.phase
        return pod_names

    def _delete_pod(self, namespace, pod_name):
        api = self.core_v1_api
        return api.delete_namespaced_pod(name=pod_name, namespace=namespace)

    def _get_oss_info(self):
        """获取对象存储信息，考虑升级服务未启动，等待1分钟"""
        i = 0
        while True:
            logger.info(i)
            try:
                ossinfos = self.get_ossgateway_client().GetOSSInfo()
                return ossinfos
            except Exception as ex:
                if i == 5:
                    logger.info(ex)
                    raise Exception(ex)
                time.sleep(10)
                i += 1
                continue
            break

    def _get_oss_inter_nodes_ip(self, oss_nodes_ip):
        """根据k8s节点获取对应节点的内部ip"""
        inter_oss_nodes_ip = self._get_inter_ips(oss_nodes_ip)
        if inter_oss_nodes_ip:
            return inter_oss_nodes_ip

    def _get_inter_ips(self, oss_nodes_ip):
        inter_ip = []
        nodes_info = ProtonCliClient.instance().nodes()
        if not nodes_info:
            return oss_nodes_ip

        for node_info in nodes_info:
            if "internal_ip" in node_info:
                if "ip4" in node_info and node_info["ip4"] in oss_nodes_ip:
                    inter_ip.append(node_info["internal_ip"])
                if "ip6" in node_info and node_info["ip6"] in oss_nodes_ip:
                    inter_ip.append(node_info["internal_ip"]) 
        return  inter_ip

    def oss_gateway_config(self, micro_service_name: str, service_conf: dict) -> dict:
        return {
            "namespace": self.namespace,
            "image": {
                "registry": ProtonCliClient.instance().cr_info().registry_address()
            },
            "depServices": {
                "mq": ConfigClient.load_config().get_dep_service_info("mq"),
                "mongodb": ConfigClient.load_config().get_dep_service_info("mongodb"),
                "redis": ConfigClient.load_config().get_dep_service_info("redis"),
            },
            "ip": service_conf["node_ips"],
            "replicaCount": len(service_conf["node_ips"]),
            "nodeSelector": {
                f"ossgateway/{micro_service_name}": "OSSGateway"
            },
            "service": {
                "publicHttpPort": service_conf["ports"][0],
                "privateHttpPort": service_conf["ports"][1],
                "privateTcpPort": service_conf["ports"][2],
                "trafficPort": service_conf["ports"][3],
                "read_only": service_conf.get("read_only", "false")
            }
        }