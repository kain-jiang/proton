#!/usr/bin/env python3
# -*- coding: utf-8 -*-

from src.clients.config import ConfigClient
import os
os.environ["DB_TYPE"] = ConfigClient.load_config().rds_type()

from src.run import start_deploy_service
from src.run import create_helm_init_daemon
from src.run import init_anyshare_cms
from src.common.log_util import logger

def main():
    logger.info("init_anyshare_cms")
    init_anyshare_cms()
    logger.info("create_helm_init_daemon")
    create_helm_init_daemon()
    logger.info("start_deploy_service")
    start_deploy_service()

if __name__ == "__main__":
    main()
