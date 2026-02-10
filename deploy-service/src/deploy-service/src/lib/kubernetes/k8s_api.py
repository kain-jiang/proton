#!/usr/bin/env python
# coding=utf-8
import os
import traceback
from typing import Dict, List

from kubernetes import client as kube_client
from kubernetes import config as kube_config
from kubernetes.client import (
    V1ConfigMap,
    V1Namespace,
    V1Node,
    V1NodeSelector,
    V1NodeSelectorRequirement,
    V1NodeSelectorTerm,
    V1ObjectMeta,
    V1PersistentVolume,
    V1PersistentVolumeSpec,
    V1Pod,
    V1Secret,
    V1VolumeNodeAffinity,
)
from kubernetes.client.rest import ApiException

from src.clients.k8s import K8SClient
from src.common.log_util import logger


def kube_api_try_except(func):
    """用于API请求异常处理的装饰器，将异常信息body抛出"""

    def wrapper(*args, **kwargs):
        try:
            result = func(*args, **kwargs)
        except ApiException as kube_api_ex:
            kube_api_e = traceback.format_exc()
            logger.error("Request kubernetes API server failed, {0}".format(kube_api_ex.body))
            raise Exception(kube_api_e)
        return result

    return wrapper


class K8SAPI:
    """管理自己 namespace 下的资源"""

    MODULE_NAME = "K8SAPI"

    @kube_api_try_except
    def __init__(self):
        self.core_v1_api = K8SClient.instance().sdk_client.core_v1_api
        self.namespace = K8SClient.instance().rest_client.self_namespace

    # pod 部分
    @kube_api_try_except
    def list_namespaced_pod(self) -> "List[V1Pod]":
        return self.core_v1_api.list_namespaced_pod(self.namespace).items

    @kube_api_try_except
    def list_namespaced_pod_name(self) -> "List[str]":
        pod_list = self.list_namespaced_pod()
        pod_name_list = [pod.metadata.name for pod in pod_list]
        return pod_name_list

    @kube_api_try_except
    def delete_namespaced_pod(self, pod_name: "str", async_req=True) -> "None":
        self.core_v1_api.delete_namespaced_pod(pod_name, self.namespace, async_req=async_req)

    @kube_api_try_except
    def read_namespace(self) -> "V1Namespace":
        return self.core_v1_api.read_namespace(self.namespace)

    @kube_api_try_except
    def create_namespaced_secret(self, secret_name: "str", secret_type: "str", data: "Dict") -> "None":
        metadata = V1ObjectMeta()
        metadata.name = secret_name
        body = V1Secret(
            api_version="v1",
            data=data,
            kind="Secret",
            metadata=metadata,
            type=f"kuberbets.io/{secret_type}",
        )
        logger.debug(f"create secret[{secret_name}], type: {secret_type}")
        self.core_v1_api.create_namespaced_secret(namespace=self.namespace, body=body)

    def update_namespaced_secret(self, secret_name: "str", secret_type: "str", data: "Dict") -> "None":
        metadata = V1ObjectMeta()
        metadata.name = secret_name
        body = V1Secret(
            api_version="v1",
            data=data,
            kind="Secret",
            metadata=metadata,
            type=f"kuberbets.io/{secret_type}",
        )
        logger.debug(f"update secret[{secret_name}]")
        self.core_v1_api.patch_namespaced_secret(name=secret_name, namespace=self.namespace, body=body)

    def list_namespaced_secret(self) -> "List[V1Secret]":
        return self.core_v1_api.list_namespaced_secret(self.namespace).items

    def list_namespaced_secret_names(self) -> "List[str]":
        secret_list = self.list_namespaced_secret()
        secret_name_list = [secret.metadata.name for secret in secret_list]
        return secret_name_list

    def list_node(self) -> "List[V1Node]":
        return self.core_v1_api.list_node().items

    def list_node_ips(self) -> "List[str]":
        node_list = self.list_node()
        node_ip_list = list()
        for node in node_list:
            addr_list = node.status.addresses
            for addr in addr_list:
                if addr.type == "InternalIP":
                    node_ip = addr.address
                    node_ip_list.append(node_ip)
                    break
        return node_ip_list

    def set_node_label(self, node_name: "str", labels: "Dict") -> "None":
        """
        @param node_name str cs节点名
        @param labels dict 标签
        """
        node: "V1Node" = self.core_v1_api.read_node(node_name)
        metadata = node.metadata
        metadata.labels = labels
        logger.debug(f"patch labels[{labels}] of node[{node_name}]")
        try:
            self.core_v1_api.patch_node(name=node_name, body=node)
        except ApiException as ex:
            if int(ex.status) == 409:
                logger.info(f"label[{labels}] node[{node_name}] is already exists.")
            else:
                raise ex

    def remove_node_label(self, node_name, label):
        """
        @param node_name str cs节点名
        @param labels str 标签key
        """
        node: "V1Node" = self.core_v1_api.read_node(node_name)
        metadata: "V1ObjectMeta" = node.metadata
        metadata.labels.pop(label, None)
        logger.debug(f"remove labels[{label}] from node[{node_name}]")
        self.core_v1_api.replace_node(name=node_name, body=node)

    def delete_persistent_volume(self, pvname):
        try:
            self.core_v1_api.delete_persistent_volume(pvname)
        except ApiException as e:
            print("Exception when calling CoreV1Api->delete_persistent_volume: %s\n" % e)

    def get_persistent_volume_list(self):

        return self.core_v1_api.list_persistent_volume()

    def delete_pvc_persistent_volume_claim(self, pvc_name):
        try:
            self.core_v1_api.delete_namespaced_persistent_volume_claim(pvc_name, self.namespace)
        except ApiException as e:
            print("Exception when calling CoreV1Api->delete_namespaced_persistent_volume_claim: %s\n" % e)

    def list_namespaced_pods(self):
        try:
            api_response = self.core_v1_api.list_namespaced_pod(self.namespace)
            return api_response
        except ApiException as e:
            print("Exception when calling CoreV1Api->list_namespaced_pod: %s\n" % e)

    def list_namespaced_cm_names(self) -> "List[str]":
        cm_list = self.core_v1_api.list_namespaced_config_map(self.namespace).items
        cm_name_list = [cm.metadata.name for cm in cm_list]
        return cm_name_list

    def create_namespaced_config_map(self, name: str, data: Dict[str, str]):
        metadata = V1ObjectMeta()
        metadata.name = name
        body = V1ConfigMap(
            api_version="v1",
            data=data,
            kind="ConfigMap",
            metadata=metadata,
        )
        logger.info(f"create configmap[{name}]")
        self.core_v1_api.create_namespaced_config_map(namespace=self.namespace, body=body)

    def patch_namespaced_config_map(self, name: str, data: Dict[str, str]):
        metadata = V1ObjectMeta()
        metadata.name = name
        body = V1ConfigMap(api_version="v1", data=data, kind="ConfigMap", metadata=metadata)
        logger.info(f"patch configmap[{name}]")
        self.core_v1_api.patch_namespaced_config_map(name=name, namespace=self.namespace, body=body)

    def read_namespaced_cm_data(self, name: str) -> Dict[str, str]:
        if name not in self.list_namespaced_cm_names():
            return {}
        cm: V1ConfigMap = self.core_v1_api.read_namespaced_config_map(name=name, namespace=self.namespace)
        return cm.data
