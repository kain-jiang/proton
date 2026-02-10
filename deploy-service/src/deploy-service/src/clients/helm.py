import os,copy,re
import pathlib
from dataclasses import dataclass
from typing import Optional, Union, Callable, Tuple
import requests

import yaml
import tempfile
import tenacity

from urllib.parse import urlsplit, urlunparse

from src.clients.cmd import CmdClient
from src.clients.k8s import K8SClient
from src.clients.protoncli import ProtonCliClient
from src.clients.cms_data.cr import CRInfo
from src.common.log_util import logger

from src.common.utils import get_host_for_url



@dataclass
class SearchedChartInfo(object):
    name: str
    version: str
    app_version: str
    description: str

    @classmethod
    def from_dict(cls, data: dict) -> "SearchedChartInfo":
        return cls(**data)

    @property
    def chart_repo(self) -> str:
        return self.name.split("/")[0]

    @property
    def chart_name(self) -> str:
        return self.name.split("/")[1]

    @property
    def chart_version(self) -> str:
        return self.version

    @property
    def chart_ref(self):
        return self.name


@dataclass
class InspectChartInfo(object):
    chart: dict
    values: dict

    @property
    def chart_name(self) -> str:
        return self.chart["name"]

    @property
    def chart_version(self) -> str:
        return self.chart["version"]


class HelmRepos(object):
    def __init__(self, cr_info: CRInfo):
        self.charts: list[SearchedChartInfo] = []
        self.refresh_charts()
        self.cr_info = cr_info
        self.__push_urls: list = None

    @property
    def push_urls(self) -> bool:
        if self.__push_urls is None:
            if not self.cr_info.chartmuseum.push or urlsplit(self.cr_info.chartmuseum.get_repo_url()).hostname in [
                "acr.aishu.cn",
                "acr-arm.aishu.cn",
            ]:
                self.__push_urls = []
            else:
                # 过于复杂的URL组装
                repo_url_split = urlsplit(self.cr_info.chartmuseum.get_repo_url())
                repo_url_paths = repo_url_split.path.split("/")
                api_path = "/".join(["api", *[path for path in repo_url_paths if path], "charts"])
                chartmuseum_scheme = repo_url_split.scheme
                chartmuseum_hosts = self.cr_info.chartmuseum.hosts or [repo_url_split.hostname]
                chartmuseum_port = (
                    self.cr_info.chartmuseum.port if self.cr_info.chartmuseum.hosts else repo_url_split.port
                )
                self.__push_urls = [
                    urlunparse(
                        (
                            chartmuseum_scheme,
                            f"{get_host_for_url(chartmuseum_host)}:{chartmuseum_port}"
                            if chartmuseum_port
                            else get_host_for_url(chartmuseum_host),
                            api_path,
                            None,
                            None,
                            None,
                        )
                    )
                    for chartmuseum_host in chartmuseum_hosts
                ]
        return self.__push_urls

    @tenacity.retry(stop=tenacity.stop_after_attempt(3), wait=tenacity.wait_fixed(3), reraise=True)
    def push_chart_after_del(
        self,
        chart_path: Union[str, pathlib.Path],
        chart_name: str,
        chart_version: str,
    ):
        chartmuseum_user = self.cr_info.chartmuseum.auth_user or None
        chartmuseum_pass = self.cr_info.chartmuseum.auth_passwd or None
        for push_url in self.push_urls:
            resp = requests.get(
                url=f"{push_url}/{chart_name}/{chart_version}",
                verify=False,
                auth=(chartmuseum_user, chartmuseum_pass),
            )
            if resp.status_code != 404:
                resp.raise_for_status()
                resp = requests.delete(
                    url=f"{push_url}/{chart_name}/{chart_version}",
                    verify=False,
                    auth=(chartmuseum_user, chartmuseum_pass),
                )
                if resp.status_code != 404:
                    resp.raise_for_status()
            chart_file_name = os.path.basename(chart_path)
            resp = requests.post(
                url=push_url,
                files={"chart": (chart_file_name, open(chart_path, "rb"), "multipart/form-data")},
                verify=False,
                auth=(chartmuseum_user, chartmuseum_pass),
            )
            if resp.status_code != 409:
                resp.raise_for_status()
        logger.info(f"push chart success [{chart_path}]")

    def refresh_charts(self):
        CmdClient.run_or_raise(f"helm repo update --fail-on-repo-update-fail")
        msg, _ = CmdClient.run_or_raise(f"helm search repo --devel --versions --output yaml")
        charts: list[dict] = yaml.safe_load(msg) or []
        self.charts = [SearchedChartInfo.from_dict(chart) for chart in charts]

    def search_chart_ref(self, chart_name: str, chart_version: str) -> str:
        for c in self.charts:
            if c.chart_name == chart_name and c.chart_version == chart_version:
                return c.chart_ref
        raise Exception(f"Cannot find chart {chart_name} version {chart_version}")

    def search_chart_ref_and_version(self, chart_and_version: str) -> Tuple[str, str]:
        for c in self.charts:
            if f"{c.chart_name}-{c.chart_version}" == chart_and_version:
                return c.chart_ref, c.chart_version
        raise Exception(f"Cannot find chart {chart_and_version}")


class HelmClient(object):
    def __init__(self, namespace: str = K8SClient.instance().rest_client.self_namespace) -> None:
        self.namespace = namespace
        # property
        self.__cr_info: Optional[CRInfo] = None
        self.__helm_repos: Optional[HelmRepos] = None

    @property
    def cr_info(self) -> CRInfo:
        if not self.__cr_info:
            self.__cr_info = ProtonCliClient.instance().cr_info()
        return self.__cr_info

    @property
    def helm_repos(self) -> HelmRepos:
        if not self.__helm_repos and self.cr_info.chart_repository == "chartmuseum":
            self.__helm_repos = HelmRepos(self.cr_info)
        return self.__helm_repos

    def inspect(self, chart_path: Union[str, pathlib.Path]) -> dict:
        chart_yaml, _ = CmdClient.run_or_raise(f"helm inspect chart {str(chart_path)}")
        value_yaml, _ = CmdClient.run_or_raise(f"helm inspect values {str(chart_path)}")
        result = yaml.safe_load(chart_yaml)
        result["values"] = yaml.safe_load(value_yaml)
        return result

    def get_current_values(self, release_name: str, all: bool=False) -> dict:
        is_all = "--all" if all else ""
        values_yaml,_ = CmdClient.run_or_raise(f"helm --namespace {self.namespace} get values {release_name} {is_all} --output yaml")
        return yaml.safe_load(values_yaml)
    
    def get_release_info(self, release_name: str)-> Union[dict, None]:
        """
        get helm3 release info
        args:
            release_name: type(str), release name for index release
        return:
            None: if release not exists
            dict: if release indexed, return a release info dict like 
                {
                    "app_version": "str",
                    "chart": "str, chart full name with version",
                    "name": "str, release name"
                    "namespace": "str, release namespace",
                }
        """
        list_yaml,_ = CmdClient.run_or_raise(f"helm ls --output yaml --namespace {self.namespace} --filter '^{release_name}$'")
        release_list = yaml.safe_load(list_yaml)
        if release_list:
            return release_list[0]
        return None

    def get_all_releases(self) -> list[dict]:
        '''
        get all helm3 releases info
        return:
            list[dict]: all releases info list
            each dict item like:
                {
                    "app_version": "str",
                    "chart": "str, chart full name with version",
                    "name": "str, release name"
                    "namespace": "str, release namespace",
                }
        '''
        list_yaml,_ = CmdClient.run_or_raise(f"helm ls --output yaml --namespace {self.namespace}")
        return yaml.safe_load(list_yaml) or []

    @classmethod
    def merge_dict(cls, basic, upper:dict) -> dict:
        bcopy = copy.deepcopy(basic)
        for k,v in upper.items():
            if isinstance(v, dict):
                v0 = bcopy.get(k)
                if isinstance(v0,dict):
                    bcopy[k] = cls.merge_dict(v0,v)
                    continue
            bcopy[k] = v
        return bcopy

    def install_or_upgrade(self, name: str, config: dict, chart_ref: str, chart_version: str, atomic: bool = True):
        with tempfile.NamedTemporaryFile(suffix=f"{name}.yaml") as tmp_values_file:
            with open(tmp_values_file.name, "w") as fw:
                yaml.safe_dump(config, fw)
            automic: str = "--atomic" if atomic else ""
            CmdClient.run_or_raise(
                f"helm upgrade {name} --install {automic} {chart_ref} --version {chart_version}"
                f" --disable-openapi-validation --namespace {self.namespace} -f {tmp_values_file.name}"
            )

    def install_or_upgrade_oci(self, name: str, config: dict, chart_name: str, chart_version: str, atomic: bool = True):
        oci = self.cr_info.oci
        chart_ref = f"oci://{oci.get_registry().removesuffix('/')}/{chart_name}"
        return self.install_or_upgrade(name, config, chart_ref, chart_version, atomic)
        
    def install_or_upgrade_any(self, name: str, config: dict, chart_name: str, chart_version: str, atomic: bool = True):
        if self.cr_info.chart_repository == "chartmuseum":
            chart_ref = self.helm_repos.search_chart_ref(chart_name, chart_version)
            return self.install_or_upgrade(name, config, chart_ref,chart_version, atomic)
        if self.cr_info.chart_repository == "oci":
            return self.install_or_upgrade_oci(name, config, chart_name, chart_version, atomic)
        else:
            raise Exception(f"not support chart repository {self.cr_info.chart_repository}")

    def uninstall(self, name: str):
        CmdClient.run_or_raise(
            f"helm uninstall {name} --namespace {self.namespace} --ignore-not-found"
        )

    def split_version(self, name_version: str) -> Tuple[str, str]:
        match = re.match(r'(.*)-(\d+\.\d+.*)', name_version)
        chart_name = match.group(1)
        chart_version = match.group(2)
        return chart_name, chart_version

    def upgrade_all_releases(
        self,
        condition: Callable[[dict], bool],
        do: Callable[[dict], dict],
        exclude_release_condition: Optional[Callable[[str], bool]] = None,
        atomic: bool = True,
    ):
        """
        如果微服务实例的配置信息condition返回了True，则将微服务配置更新成do的结果
        """
        # 1. 获取所有微服务实例
        releases_yaml, _ = CmdClient.run_or_raise(f"helm list --output yaml --namespace {self.namespace}")
        releases_infos: list = yaml.safe_load(releases_yaml)
        for release_info in releases_infos:
            release_name = release_info["name"]
            if exclude_release_condition and exclude_release_condition(release_name):
                # 排除列表
                continue
            release_name = release_info["name"]
            # 2. 获取旧配置
            config_yaml, _ = CmdClient.run_or_raise(f"helm --namespace {self.namespace} get values {release_name} --output yaml")
            old_config = yaml.safe_load(config_yaml)
            if condition(old_config):
                # 3. 需要更新
                chart_name, chart_version = self.split_version(release_info["chart"])
                self.install_or_upgrade_any(
                    name=release_name,
                    config=do(old_config),
                    chart_name=chart_name,
                    chart_version=chart_version,
                    atomic=atomic,
                )

    def upgrade_hydra_redirect_url(self, access_addr: str):
        def hydra_url(path: str):
            return f"{access_addr.removesuffix('/')}{path}"

        def get_hydra_version() -> str:
            all_rls, _ = CmdClient.run_or_raise(f"helm --namespace {self.namespace} list --all --filter hydra --output yaml")
            chart_str: str = yaml.safe_load(all_rls)[0]["chart"]
            return chart_str.removeprefix("hydra-")

        hydra_version = get_hydra_version()
        rel, _ = CmdClient.run_or_raise(f"helm --namespace {self.namespace} get values hydra --output yaml")
        hydra_config: dict = yaml.safe_load(rel)
        # hydra_config["hydra"]["config"]["urls"] = {
        #     "login": hydra_url("/oauth2/signin"),
        #     "consent": hydra_url("/oauth2/consent"),
        #     "logout": hydra_url("/oauth2/signout"),
        #     "self": {
        #         "issuer": hydra_url(""),
        #     },
        # }
        self.install_or_upgrade_any(
            name="hydra",
            config=hydra_config,
            chart_name="hydra",
            chart_version=hydra_version,
        )
