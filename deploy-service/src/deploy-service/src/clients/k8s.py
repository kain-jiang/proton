import os
from typing import Optional
import json
import base64

import requests
from kubernetes import client as kube_client
from kubernetes import config as kube_config
from packaging import version
from src.utils.net import get_host_for_url


class K8SClient:

    class RestClient(object):
        def __init__(self):
            self.__self_namespace: str = ""
            self.__token: str = ""
            self.__version: str = ""
            self.__ingress_api_version: str = ""
            self.__kubernetes_host: str = os.environ.get("KUBERNETES_SERVICE_HOST", "kubernetes.default")
            self.__kubernetes_port: str = os.environ.get("KUBERNETES_SERVICE_PORT", "443")

        @property
        def self_namespace(self) -> str:
            if not self.__self_namespace:
                with open("/var/run/secrets/kubernetes.io/serviceaccount/namespace", "r") as fr:
                    self.__self_namespace = fr.read()
            return self.__self_namespace

        @property
        def token(self) -> str:
            if not self.__token:
                with open("/var/run/secrets/kubernetes.io/serviceaccount/token", "r") as fr:
                    self.__token = fr.read()
            return self.__token

        @property
        def ingress_api_version(self) -> str:
            if not self.__ingress_api_version:
                if version.parse("1.18") <= version.parse(self.version) < version.parse("1.20"):
                    self.__ingress_api_version = "extensions/v1beta1"
                elif version.parse("1.20") <= version.parse(self.version):
                    self.__ingress_api_version = "networking.k8s.io/v1"
                else:
                    self.__ingress_api_version = "extensions/v1beta1"
            return self.__ingress_api_version

        @property
        def lower_k8s(self) -> bool:
            return version.parse(self.version) < version.parse("1.18")

        @property
        def version(self) -> str:
            if not self.__version:
                resp: requests.Response = requests.get(
                    url=f"https://{get_host_for_url(self.__kubernetes_host)}:{self.__kubernetes_port}/version",
                    verify=False,
                    headers={"Accept": "application/json", "Authorization": f"bearer {self.token}"},
                )
                resp.raise_for_status()
                git_ver: str = resp.json()["gitVersion"]
                self.__version = git_ver.split("-")[0]
            return self.__version

        def save_ingress(self, name: str, namespace: str, paths: list, annotations: dict):
            data: dict = {
                "apiVersion": self.ingress_api_version,
                "kind": "Ingress",
                "metadata": {
                    "name": name,
                    "namespace": namespace,
                    "annotations": annotations,
                    "labels": {"ingress/owner": "deploy-service", "ingress/version": "v1.0.0"},
                },
                "spec": {"rules": [{"http": {"paths": paths}}]},
            }
            resp: requests.Response = requests.put(
                url=f"https://{get_host_for_url(self.__kubernetes_host)}:{self.__kubernetes_port}/apis/{self.ingress_api_version}/namespaces/{namespace}/ingresses/{name}",
                # noqa
                verify=False,
                headers={
                    "Accept": "application/json",
                    "Authorization": f"bearer {self.token}",
                },
                json=data,
                timeout=30.0,
            )
            if resp.status_code == 404:
                resp = requests.post(
                    url=f"https://{get_host_for_url(self.__kubernetes_host)}:{self.__kubernetes_port}/apis/{self.ingress_api_version}/namespaces/{namespace}/ingresses",
                    # noqa
                    verify=False,
                    headers={
                        "Accept": "application/json",
                        "Authorization": f"bearer {self.token}",
                    },
                    json=data,
                    timeout=30.0,
                )
            resp.raise_for_status()

    class SDKClient(object):

        ROLE_PRE = "node-role.aishu.cn/"

        def __init__(self):
            self.__core_v1_api: Optional[kube_client.CoreV1Api] = None
            self.__apps_v1_api: Optional[kube_client.AppsV1Api] = None
            self.__batch_v1_api: Optional[kube_client.BatchV1Api] = None
            self.__config_loaded = False

        def config_load(self):
            if not self.__config_loaded:
                if os.environ.get("IN_CLUSTER", "true") != "true":
                    kube_config.load_kube_config()
                else:
                    kube_config.load_incluster_config()
                self.__config_loaded = True

        @property
        def core_v1_api(self):
            if not self.__core_v1_api:
                self.config_load()
                self.__core_v1_api = kube_client.CoreV1Api()
            return self.__core_v1_api

        @property
        def apps_v1_api(self):
            if not self.__apps_v1_api:
                self.config_load()
                self.__apps_v1_api = kube_client.AppsV1Api()
            return self.__apps_v1_api

        @property
        def batch_v1_api(self):
            if not self.__batch_v1_api:
                self.config_load()
                self.__batch_v1_api = kube_client.BatchV1Api()
            return self.__batch_v1_api

        def get_configmap(self, name: str, namespace: str,) -> dict:
            return self.core_v1_api.read_namespaced_config_map(name=name, namespace=namespace).data

        def get_secret(self, name: str, namespace: str,) -> dict[str, bytes]:
            data = self.core_v1_api.read_namespaced_secret(name=name, namespace=namespace).data
            return {k: base64.b64decode(v) for k, v in data.items()}

        def save_secret(self, name: str, namespace: str, data: dict[str, str]):
            try:
                secret: kube_client.V1Secret = self.core_v1_api.read_namespaced_secret(name, namespace)
            except kube_client.ApiException as api_exp:
                if api_exp.status != 404:
                    raise
                # 不存在，创建
                self.core_v1_api.create_namespaced_secret(namespace, kube_client.V1Secret(
                    metadata=kube_client.V1ObjectMeta(name=name, namespace=namespace),
                    data=data
                ))
            else:
                # 更新
                secret.data = data
                self.core_v1_api.replace_namespaced_secret(name, namespace, secret)


        def delete_secret(self, name: str, namespace: str):
            try:
                self.core_v1_api.delete_namespaced_secret(name, namespace)
            except kube_client.ApiException as api_exp:
                if api_exp.status != 404:
                    raise

        def delete_pods(self, filter_str: str, namespace: str):
            pod_list: kube_client.V1PodList = self.core_v1_api.list_namespaced_pod(namespace=namespace)
            for pod in pod_list.items:
                pod: kube_client.V1Pod
                metadata: kube_client.V1ObjectMeta = pod.metadata
                if filter_str in metadata.name:
                    self.core_v1_api.delete_namespaced_pod(metadata.name, namespace=namespace)

        def set_label(self, node: str, labels: dict):
            v1node: kube_client.V1Node = self.core_v1_api.read_node(node)
            v1meta: kube_client.V1ObjectMeta = v1node.metadata
            v1meta.labels = labels
            try:
                self.core_v1_api.patch_node(name=node, body=v1node)
            except kube_client.ApiException as api_exp:
                if api_exp.status != 409:
                    raise

        def unset_label(self, node: str, labels: list):
            labels = [labels] if isinstance(labels, str) else labels
            v1node: kube_client.V1Node = self.core_v1_api.read_node(node)
            v1meta: kube_client.V1ObjectMeta = v1node.metadata
            exist_labels: dict = v1meta.labels
            for label in labels:
                if label in exist_labels:
                    del exist_labels[label]
            self.core_v1_api.replace_node(name=node, body=v1node)

        def get_label(self, node: str) -> dict:
            v1node: kube_client.V1Node = self.core_v1_api.read_node(node)
            v1meta: kube_client.V1ObjectMeta = v1node.metadata
            return v1meta.labels

        def get_all_labels(self) -> dict:
            v1nodelist: kube_client.V1NodeList = self.core_v1_api.list_node()
            v1nodes: list[kube_client.V1Node] = v1nodelist.items
            result: dict = {}
            for v1node in v1nodes:
                v1meta: kube_client.V1ObjectMeta = v1node.metadata
                result[v1meta.name] = v1meta.labels
            return result
        
        def get_all_node_internal_ip(self):
            v1nodelist: kube_client.V1NodeList = self.core_v1_api.list_node()
            v1nodes: list[kube_client.V1Node] = v1nodelist.items
            result = []
            for v1node in v1nodes:
                addrs = v1node.status.addresses
                for addr in addrs:
                    if addr.type == "InternalIP":
                        result.append(addr.address)
            return result

        def set_role(self, node: str, role: str):
            self.set_label(node, {f"{self.ROLE_PRE}{role}": role})

        def unset_role(self, node: str, role: str):
            self.unset_label(node, [f"{self.ROLE_PRE}{role}"])

        def get_role(self, node: str):
            """
            获取指定节点node-role.aishu.cn角色
            :param node: 节点名
            :return: 指定角色信息
            """
            labels: dict[str, str] = self.get_label(node)
            return [key.removeprefix(self.ROLE_PRE) for key, _ in labels.items() if key.startswith(self.ROLE_PRE)]

        def get_all_role(self):
            """
            获取所有节点的node-role.aishu.cn角色
            :return: 所有角色信息
            """
            all_labels: dict[str, dict[str, str]] = self.get_all_labels()
            return {
                name: [key.removeprefix(self.ROLE_PRE) for key, val in labels.items() if key.startswith(self.ROLE_PRE)]
                for name, labels in all_labels.items()
            }
        
        def create_secret_docker_registry(self, ns: str, name: str, host: str, username: str, password: str, email: str = ""):
            server = host.split("/")[0]
            dockerConfig = {
                "docker-server": server,
                "docker-username": username,
                "docker-password": password,
            }
            if email:
                dockerConfig["docker-email"] = email
            dockerConfigBytes = json.dumps(dockerConfig).encode("utf-8")
            try:
                secret: kube_client.V1Secret = self.core_v1_api.read_namespaced_secret(name, ns)
            except kube_client.ApiException as api_exp:
                if api_exp.status != 404:
                    raise
                # 不存在，创建
                self.core_v1_api.create_namespaced_secret(ns, kube_client.V1Secret(
                    type="kubernetes.io/dockerconfigjson",
                    metadata=kube_client.V1ObjectMeta(name=name, namespace=ns),
                    data={
                        ".dockerconfigjson": base64.b64encode(dockerConfigBytes).decode("utf-8")
                    }
                ))
            else:
                # 更新
                secret.data[".dockerconfigjson"] = base64.b64encode(dockerConfigBytes).decode("utf-8")
                self.core_v1_api.replace_namespaced_secret(name, ns, secret)

    def __init__(self) -> None:
        self.__rest_client: Optional[K8SClient.RestClient] = None
        self.__sdk_client: Optional[K8SClient.SDKClient] = None

    @property
    def rest_client(self):
        if not self.__rest_client:
            self.__rest_client = K8SClient.RestClient()
        return self.__rest_client

    @property
    def sdk_client(self):
        if not self.__sdk_client:
            self.__sdk_client = K8SClient.SDKClient()
        return self.__sdk_client

    _instance: "K8SClient" = None

    @classmethod
    def instance(cls) -> "K8SClient":
        if cls._instance is None:
            cls._instance = cls()
        return cls._instance


if __name__ == "__main__":
    cli = K8SClient()
    print(cli.sdk_client.get_all_role())
