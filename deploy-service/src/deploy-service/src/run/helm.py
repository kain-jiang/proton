import os

import tenacity
from src.clients.helm import HelmClient
from src.clients.cmd import CmdClient
from src.clients.config import ConfigClient
from multiprocessing import Process


def create_helm_init_daemon():

    config = ConfigClient.load_config()
    if not config.use_protoncli():
        return

    cr_info = HelmClient().cr_info

    @tenacity.retry(wait=tenacity.wait_fixed(3), reraise=True)
    def _add_helm_repo():
        repo_name: str = cr_info.chartmuseum.projects[-1]
        repo_url: str = cr_info.chartmuseum.get_repo_url()
        auth_user: str = cr_info.chartmuseum.auth_user
        auth_passwd: str = cr_info.chartmuseum.auth_passwd
        with_auth_str = f"--username {auth_user} --password {auth_passwd}" if auth_user else ""
        CmdClient.run_or_raise(f"helm repo add {repo_name} {repo_url} {with_auth_str}", retry_attempt=1)

    
    @tenacity.retry(wait=tenacity.wait_fixed(3), reraise=True)
    def _login_helm_registry():
        if cr_info.oci.username:
            registry = cr_info.oci.get_registry()
            username = cr_info.oci.username
            password = cr_info.oci.password
            insecure = "--insecure" if cr_info.oci.plain_http else ""
            CmdClient.run_or_raise(f"helm registry login {registry} --username {username} --password {password} {insecure}")

    target = _add_helm_repo
    if cr_info.chart_repository == "oci":
        target = _login_helm_registry

    add_repo_process = Process(target=target)
    add_repo_process.daemon = True
    add_repo_process.start()