USE `deploy`;

CREATE TABLE IF NOT EXISTS `cert` (
  `id` bigint(11) NOT NULL AUTO_INCREMENT,
  `f_key` varchar(64) NOT NULL COMMENT '配置选项',
  `f_value` varchar(8192) NOT NULL COMMENT '配置内容',
  PRIMARY KEY (`id`),
  UNIQUE KEY `f_key` (`f_key`)
) ENGINE=InnoDB AUTO_INCREMENT=6 COMMENT='证书信息表';



CREATE TABLE IF NOT EXISTS `chart` (
  `id` bigint(11) NOT NULL AUTO_INCREMENT,
  `chart_name` varchar(50) NOT NULL COMMENT 'chart包名',
  `service_name` varchar(50) NOT NULL COMMENT '服务名',
  `chart_version` varchar(50) NOT NULL COMMENT 'chart包版本',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=48 COMMENT='helm chart包信息';



CREATE TABLE IF NOT EXISTS `client_package_info` (
  `id` bigint(11) NOT NULL AUTO_INCREMENT,
  `f_name` varchar(150) NOT NULL COMMENT '升级包名称',
  `f_os` int(11) NOT NULL COMMENT 'OS系统',
  `f_size` bigint(20) NOT NULL COMMENT '大小',
  `f_version` varchar(50) NOT NULL COMMENT '版本',
  `f_time` varchar(50) NOT NULL COMMENT '上传时间',
  `f_mode` tinyint(1) NOT NULL COMMENT '升级类型',
  `f_pkg_location` tinyint(4) NOT NULL DEFAULT 1 COMMENT '升级包位置',
  `f_url` text NOT NULL COMMENT '下载地址',
  `f_oss_id` varchar(50) DEFAULT NULL COMMENT '对象存储ID',
  `f_update_type` varchar(32) NOT NULL COMMENT '升级包升级类型',
  `f_open_download` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否开放下载',
  `f_version_description` VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '安裝包',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 COMMENT='客户端升级包信息';


INSERT INTO `client_package_info`(
    `f_name`, `f_os`, `f_size`, `f_version`, `f_time`, `f_mode`, `f_pkg_location`, `f_url`, `f_oss_id`, `f_update_type`, `f_open_download`, `f_version_description`
) SELECT
      'ios',7,0,'7.0.0.0','2023/06/25 10:22:16',0,2,'https://apps.apple.com/cn/app/anyshare/id1538388902',NULL,'custom',1, ''
FROM DUAL
WHERE NOT EXISTS(
    SELECT `f_name` FROM `client_package_info` WHERE `f_name`='ios' AND `f_os`=7
);




CREATE TABLE IF NOT EXISTS `containerized_service` (
  `id` bigint(11) NOT NULL AUTO_INCREMENT,
  `service_name` varchar(50) NOT NULL COMMENT '服务名',
  `replicas` int(11) DEFAULT 0 COMMENT '服务副本数',
  `installed_version` varchar(50) DEFAULT '' COMMENT '已安装版本',
  `installed_package` varchar(100) DEFAULT '' COMMENT '已安装包',
  `available_version` varchar(50) DEFAULT '' COMMENT '当前可用版本',
  `available_package` varchar(100) DEFAULT '' COMMENT '当前可用包',
  `require_third_app_depservice` tinyint(1) NOT NULL DEFAULT 0 COMMENT '必填依赖',
  `optional_install_micro_service` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否可选安装微服务，用于升级时决定是否安装新增的微服务',
  PRIMARY KEY (`id`),
  UNIQUE KEY `service_name` (`service_name`)
) ENGINE=InnoDB AUTO_INCREMENT=2 COMMENT='容器服务信息';



-- CREATE TABLE IF NOT EXISTS `custom_name` (
--   `name` char(32) NOT NULL,
--   `value` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL CHECK (json_valid(`value`)),
--   UNIQUE KEY `f_index_custom_name` (`name`) USING BTREE
-- ) ENGINE=InnoDB;



CREATE TABLE IF NOT EXISTS `deployment_option` (
  `id` bigint(11) NOT NULL AUTO_INCREMENT,
  `option_key` varchar(256) NOT NULL COMMENT '配置选项',
  `option_value` varchar(2048) NOT NULL COMMENT '配置选项值',
  PRIMARY KEY (`id`),
  UNIQUE KEY `option_key` (`option_key`)
) ENGINE=InnoDB AUTO_INCREMENT=5 COMMENT='部署配置信息';



CREATE TABLE IF NOT EXISTS `depservice_oss` (
  `id` bigint(11) NOT NULL AUTO_INCREMENT,
  `oss_name` varchar(255) NOT NULL,
  `service_name` varchar(50) NOT NULL COMMENT '服务名',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT='Depservice OSS信息表';



CREATE TABLE IF NOT EXISTS `micro_service` (
  `id` bigint(11) NOT NULL AUTO_INCREMENT,
  `micro_service_name` varchar(100) NOT NULL COMMENT '微服务名称',
  `service_name` varchar(50) NOT NULL COMMENT '服务名称',
  `micro_service_version` varchar(50) NOT NULL COMMENT '微服务版本',
  `external_port` int(11) DEFAULT 0 COMMENT '外部端口',
  `internal_port` int(11) DEFAULT 0 COMMENT '内部端口',
  `need_ingress` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `micro_service_name` (`micro_service_name`)
) ENGINE=InnoDB AUTO_INCREMENT=47 COMMENT='微服务';



CREATE TABLE IF NOT EXISTS `micro_third_app_depservice` (
  `id` bigint(11) NOT NULL AUTO_INCREMENT,
  `service_name` varchar(50) NOT NULL COMMENT '模块化服务名',
  `micro_service` varchar(50) NOT NULL COMMENT '微服务名称',
  `third_app_service` varchar(50) NOT NULL COMMENT '依赖所属',
  `components_name` varchar(50) NOT NULL COMMENT '微服务依赖',
  `enable` tinyint(1) DEFAULT 0 COMMENT '是否启用第三方',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 COMMENT='第三方依赖管理信息表';



CREATE TABLE IF NOT EXISTS `obsinstance_oss` (
  `id` bigint(11) NOT NULL AUTO_INCREMENT,
  `bucket` varchar(255) NOT NULL COMMENT 'BUCKET名字',
  `obs_id` varchar(50) NOT NULL COMMENT 'BUCKET ID',
  `instance_id` varchar(50) NOT NULL COMMENT '服务实例ID',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT='Depservice OSS信息表';



CREATE TABLE IF NOT EXISTS `os_config` (
  `id` bigint(11) NOT NULL AUTO_INCREMENT,
  `f_os` varchar(50) NOT NULL COMMENT 'OS系统',
  `f_mode` varchar(255) NOT NULL COMMENT '升级包当前使用升级类型',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=10 COMMENT='客户端系统开放下载配置';


INSERT INTO `os_config`(`f_os`, `f_mode`)
SELECT '2','standard' FROM DUAL WHERE NOT EXISTS(SELECT `f_os` FROM `os_config` WHERE `f_os`='2');

INSERT INTO `os_config`(`f_os`, `f_mode`)
SELECT '7','custom' FROM DUAL WHERE NOT EXISTS(SELECT `f_os` FROM `os_config` WHERE `f_os`='7');

INSERT INTO `os_config`(`f_os`, `f_mode`)
SELECT '3','standard' FROM DUAL WHERE NOT EXISTS(SELECT `f_os` FROM `os_config` WHERE `f_os`='3');

INSERT INTO `os_config`(`f_os`, `f_mode`)
SELECT '4','standard' FROM DUAL WHERE NOT EXISTS(SELECT `f_os` FROM `os_config` WHERE `f_os`='4');

INSERT INTO `os_config`(`f_os`, `f_mode`)
SELECT '5','standard' FROM DUAL WHERE NOT EXISTS(SELECT `f_os` FROM `os_config` WHERE `f_os`='5');

INSERT INTO `os_config`(`f_os`, `f_mode`)
SELECT '8','custom' FROM DUAL WHERE NOT EXISTS(SELECT `f_os` FROM `os_config` WHERE `f_os`='8');

INSERT INTO `os_config`(`f_os`, `f_mode`)
SELECT '9','standard' FROM DUAL WHERE NOT EXISTS(SELECT `f_os` FROM `os_config` WHERE `f_os`='9');

INSERT INTO `os_config`(`f_os`, `f_mode`)
SELECT '10','standard' FROM DUAL WHERE NOT EXISTS(SELECT `f_os` FROM `os_config` WHERE `f_os`='10');

INSERT INTO `os_config`(`f_os`, `f_mode`)
SELECT '11','standard' FROM DUAL WHERE NOT EXISTS(SELECT `f_os` FROM `os_config` WHERE `f_os`='11');



CREATE TABLE IF NOT EXISTS `pv` (
  `id` bigint(11) NOT NULL AUTO_INCREMENT,
  `pv_name` varchar(50) NOT NULL DEFAULT '' COMMENT '存储名',
  `release_name` varchar(50) NOT NULL DEFAULT '' COMMENT '使用pv的服务名称',
  PRIMARY KEY (`id`),
  UNIQUE KEY `pv_name` (`pv_name`)
) ENGINE=InnoDB COMMENT='k8s持久化存储';



CREATE TABLE IF NOT EXISTS `third_party_service` (
  `id` bigint(11) NOT NULL AUTO_INCREMENT,
  `service` varchar(100) NOT NULL COMMENT '第三方服务名',
  `server` varchar(100) NOT NULL COMMENT '第三方服务d额服务器名',
  `protocol` varchar(100) NOT NULL COMMENT '协议',
  `host` varchar(100) NOT NULL COMMENT '第三方服务地址 host',
  `port` int(11) DEFAULT NULL COMMENT '第三方服务地址 port',
  PRIMARY KEY (`id`),
  UNIQUE KEY `service` (`service`,`server`)
) ENGINE=InnoDB COMMENT='第三方服务';



CREATE TABLE IF NOT EXISTS `upgradation_status` (
  `id` bigint(11) NOT NULL AUTO_INCREMENT,
  `upgrade_id` varchar(100) NOT NULL COMMENT '升级ID',
  `name` varchar(100) NOT NULL COMMENT '服务名',
  `type` varchar(100) NOT NULL COMMENT '服务类型：module-service|micro-service',
  `status` varchar(100) NOT NULL COMMENT '升级状态：start|running|failed|success',
  `start` datetime NOT NULL COMMENT '升级开始时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `upgrade_id` (`upgrade_id`)
) ENGINE=InnoDB COMMENT='服务升级状态';



CREATE TABLE IF NOT EXISTS `upgradation_status_records` (
  `id` bigint(11) NOT NULL AUTO_INCREMENT,
  `upgrade_id` varchar(100) NOT NULL COMMENT '升级ID',
  `time` datetime NOT NULL COMMENT '记录时间',
  `message` varchar(2048) NOT NULL COMMENT '记录内容',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT='服务升级记录';

INSERT INTO  `chart` (`chart_version`, `chart_name`, `service_name`) SELECT '1.0.0-master', 'ossgateway', 'ManagementConsole' FROM DUAL WHERE NOT EXISTS (SELECT `chart_name` FROM `chart` WHERE `chart_name` = 'ossgateway');