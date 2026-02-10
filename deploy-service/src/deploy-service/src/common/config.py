#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# @Time    : 2021/4/10 14:40
# @Author  : Jimmy.li
# @Email   : jimmy.li@aishu.cn

from enum import Enum, unique

ERROR_DEFINE = {
    # ------500-------
    "500017000": "NCT_UNKNOWN_ERROR",  # 未知错误
    "500017001": "NCT_NODE_OFF_LINE",  # 节点离线
    "500017002": "NCT_SITE_STATION_INFO",  # 没有站点信息
    "500017003": "NCT_EXCE_COMMAND_FAILED",  # 执行命令失败
    "500017004": "NCT_GET_VIP_FAILED",  # 获取vip失败
    "500017005": "NCT_SYNCHRONIZATION_CERT_FAILED",  # 同步证书失败
    "500017007": "NCT_SYNCHRONIZATION_CERT_OSS_FAILED",  # 同步存储证书失败
    "500017006": "NCT_SET_SSL_FAILED",  # 设置节点SSL失败
    "500017008": "NCT_NO_DEFAULT_STORAGE",  # 未配置默认存储
    "500017009": "NCT_GET_NODE_INFO_FAILED",  # 获取节点信息失败
    "500017010": "NCT_GET_TIME_ZONE_FAILED",  # 获取时区失败
    "500017011": "NCT_GET_OS_LANGUAGE_FAILED",  # 获取系统语言失败
    "500017012": "NCT_GET_AS_LANGUAGE_FAILED",  # 获取AS语言失败
    "500017013": "NCT_GET_CUSTOMER_CONFIGURATION",  # 获取自定义服务失败
    "500017020": "NCT_HYDRA_NOT_AVAILABLE",  # hydra服务不可用
    "500017021": "NCT_GET_DEVICE_INFO_FAILED",  # 获取设备信息失败
    "500017022": "NCT_CLIENT_UPDATE_DECISION_EXCEPTION", # 客户端更新策略异常,检查策略引擎
    # 400
    "400017001": "NCT_GET_DOWNLOAD_URL_FAILED",  # 获取下载地址失败
    "400017002": "NCT_INVALID_OPTION",  # 非法的option
    "400017003": "NCT_INVALID_VALUE",  # 非法的value
    "400017004": "NCT_OPTION_NOT_EXISTS",  # option不存在
    "400017005": "NCT_PARAMS_ERROR",  # 参数错误（缺少）
    "400017006": "NCT_INVALID_FILENAME",  # 非法的文件名
    "400017007": "NCT_DEL_PKG_INFO_FAILED",  # 删除升级包信息失败
    "400017008": "NCT_UPLOAD_EOSS_FAILED",  # 上传包失败
    "400017009": "NCT_INVALID_OS_PARAMETER",  # 无效的os系统参数
    "400017010": "NCT_PACKAGE_NOT_EXIST",  # 文件不存在
    "400017011": "NCT_UPDATE_PACKAGE_ALREADY_EXIST",  # 升级包已存在
    "400017012": "NCT_FILE_SUFFIX_ERROR",  # 未知的文件后缀名
    "400017013": "NCT_FILE_PREFIX_ERROR",  # 未知的文件前缀
    "400017014": "NCT_OS_CANT_MATCH_SURFIX",  # 系统和后缀名不匹配
    "400017015": "NCT_NO_CLIENT_PACKAGE_UPLOAD",  # 无升级包上传
    "400017016": "NCT_ALREADY_SET_DOWNLOAD_LINK",  # 已配置下载地址（安卓）
    "400017017": "NCT_INVALID_PACKAGE_VERSION",  # 无效的升级包版本号
    "400017018": "NCT_INVALID_OS_SET_DOWNLOAD_URL",  # 无效的OS系统设置下载地址（仅支持ios，安卓）
    "400017019": "NCT_NO_STORAGE",  # 未配置存储
    "400017028": "NCT_INVALID_CERT_TYPE",  # 无效的证书类型
    "400017029": "NCT_CERT_NOT_MATCH",  # 证书不匹配
    "400017030": "NCT_UPDATE_TYPE_ERROR",  # 无效的升级类型
    "400017033": "NCT_INVALID_VERSION",  # 非法的版本号
    "400017034": "NCT_INVALID_MICRO_SERVICE_CONFIG",  # 无效的微服务配置项
    "400017035": "NCT_INVALID_MICRO_SERVICE_VALUE",  # 无效的微服务配置值
    "400017036": "NCT_INVALID_PARAMS_OF_BODY",  # body中缺少参数
    "400017205": "NCT_FILE_DAMAGED",  # 上传文件损坏
    "400017204": "NCT_UPLOAD_TASK_NOT_EXISTS",  # 上传任务不存在
    "400017206": "NCT_FILE_TYPE_NOT_SUPPORT",  # 文件类型不支持
    "400017207": "NCT_FILE_PART_NOT_COMPLETE",  # 分片数不完全
    "400017101": "NCT_INVALID_ACCESS_ADDR_TYPE",  # 无效的访问地址类型
    "400017103": "NCT_INVALID_ACCESS_ADDR",  # 无效的访问地址
    "400017104": "NCT_INVALID_PORT",  # 无效的端口
    "400017234": "NCT_INVALID_SERVICE_NAME",  # 无效的服务名
    "400017235": "NCT_SERVICE_PACKAGE_NOT_EXIST",  # 服务包不存在
    "400017236": "NCT_SERVICE_PACKAGE_VERSION_LOW",  # 服务包版本低于当前版本
    "400017237": "NCT_SERVICE_PACKAGE_FORMAT_INVALID",  # 无效的服务包名
    "400017238": "NCT_SERVICE_PACKAGE_NOT_COMPLETE",  # 服务包不完整
    "400017239": "NCT_NODE_NOT_IN_CLUSTER",  # 节点不在集群中
    "400017240": "NCT_SERVICE_NOT_INSTALLED",  # 服务未安装
    "400017241": "NCT_DELETE_SERVICE_NODE_NOT_ONE",  # 删除节点不是一个
    "400017242": "NCT_ES_CLUSTER_NOT_USE",  # es集群状态不可用
    "400017243": "NCT_NOT_FOUND_VERSION_NUM",  # 找不到有效版本号
    "400017244": "NCT_MILVUS_INSTALLED",  # 该节点已经安装milvus
    "400017245": "NCT_MILVUS_MORE_THAN_ONE",  # 多个节点已经安装milvus
    "400017246": "NCT_CONNECT_OSS_ERROR",  # 连接对象存储失败
    "400017247": "NCT_OSS_ALREADY_EXIST",  # BUCKET已存在
    "400017248": "NCT_OSSGATEWAY_NOT_INSTALL",  # OSS网关未安装
    "400017249": "NCT_OSS_BUCKET_ALREADY_EXIST_IN_AS",  # 其他站点存在同名bucket
    "400017250": "NCT_OSS_NAME_ALREADY_EXIST_IN_AS",  # 其他站点存在同名存储
    "400017251": "NCT_THIRD_APP_SERVICE_EXIST_IN_AS",  # 第三方服务已安装，无法删除对应依赖
    "400018001": "NCT_GET_VALUES_YAML_FAILED",  # 获取微服务values.yaml失败
    "400018002": "NCT_UPGRADE_MICRO_SERVICE_RDS_FAILED",  # 更新微服务rds失败
    "400018003": "NCT_UPGRADE_MICRO_SERVICE_ENV_TIMEZONE_FAILED",  # 更新微服务系统时区失败
    "400018004": "NCT_UPGRADE_MICRO_SERVICE_ENV_LANGUAGE_FAILED",  # 更新微服务系统语言失败
    "400018005": "NCT_UPGRADE_MICRO_SERVICE_SERVICE_LANGUAGE_FAILED",  # 更新微服务全球化语言失败
    "400018006": "NCT_UPGRADE_MICRO_SERVICE_DEP_SERVICE_FAILED",  # 更新微服务依赖失败
    "400018007": "NCT_UPGRADE_SERVICE_RUNNING",  # 服务正在运行,
    "400018008": "NCT_GET_CUSTOMER_CONFIGURATION_FAILED",  # 获取自定义服务失败
    # 第三方依赖服务配置错误码
    "400017201": "NCT_REQUIRE_THIRD_APP_DEPSERVICE_NOT_CONFIG",  # 未配置第三方必填依赖
    "400017202": "NCT_THIRD_APP_DEPSERVICE_NOT_EXISTS",  # 第三方依赖信息不存在
    "400017203": "NCT_NOT_FIND_MICRO_SERVICE_THIRD_DEPSERVICE",  # 微服务中的依赖信息未找到
    # 404
    "404017001": "NCT_MICRO_SERVICE_NOT_FOUND",  # 微服务不存在
    "404017002": "NCT_MICRO_SERVICE_ATTR_NOT_FOUND",  # 微服务属性不存在
    "404017003": "NCT_ORIENTDB_NOT_FOUND",
    "404017100": "NCT_MICRO_SERVICE_NOT_FOUND_IN_MODULE_SERVICE",  # 模块服务中未包含该微服务
    # 423
    "423017001": "NCT_FORBIDDEN_UNINSTALL",  # 禁止卸载服务
    # 423
    "423017002": "NCT_FORBIDDEN_DUPLICATE_REQUEST",  # 拒绝重复请求
    "409017001": "NCT_MICRO_SERVICE_ALREADY_INSTALLED_BY_OTHER_MODULE_SERVICE",  # 微服务已被其他模块服务安装
    # URL 校验
    "404017300": "NCT_THIRD_PARTY_SERVICE_NOT_FOUND",  # 第三方服务不存在
    "500017300": "NCT_ADDRESS_INACCESSIBLE",  # 地址不可达

    # 401
    "401017000": "NCT_UNAUTHORIZED",  # oauth验证未通过
}

HYDRA_DSN = (
    "mysql://{user}:{password}@tcp({host}:{port})/hydra_v2?parseTime=true&timeout=5s&readTimeout=5s&writeTimeout=5s"
)

# -------------------------------------------------------------
AS_LANGUAGE_FILE_PATH = "/sysvol/conf/language.conf"
OS_LANGUAGE_FILE_PATH = "/etc/locale.conf"

# 微服务配置修改对应字典
MICRO_SERVICE_UPGRADE = {
    "os-timezone": {"env": {"timezone": ""}},
    "os-language": {"env": {"language": ""}},
    "micro-service-language": {"service": {"language": ""}},
    "rds": {"depServices": {"rds": {"host": "", "password": "", "port": 3320, "user": ""}}},
    "services": {
        "urls": {
            "consent": "https://{host}:{port}/oauth2/consent",
            "login": "https://{host}:{port}/oauth2/signin",
            "logout": "https://{host}:{port}/oauth2/signout",
            "self": {"issuer": "https://{host}:{port}"},
        }
    },
}

# micro service 配置
MICRO_SERVICE_VALUES = {
    "os-language": ["en_US.UTF-8", "zh_TW.UTF-8", "zh_CN.UTF-8"],
    "micro-service-language": ["en_US", "zh_TW", "zh_CN"],
    "os-timezone": [],
    "customerConfiguration": [],
}

# ----------------------------------------------------------------------

# 升级包存放位置
PKG_LOCATION_LOCAL_TO_OSS = 1  # 表示本地上传到对象存储
PKG_LOCATION_STATIC_URL = 2  # 表示独立配置升级包下载地址

# OS_TYPE_DICT 的key 和 OS_TYPE 里面的value位置相对应，从而得到每个类型的数字代表
OS_TYPE_DICT = {
    # 0
    # 1
    "android": "2",
    "mac": "3",
    'win32_advanced': '4',
    # 'win64_advanced': '5',
    # 'office_plugin': '6',
    "ios": "7",
    "win64_advanced": "8",
    "linux_x64_rpm": "9",
    "linux_arm64_rpm": "10",
    "linux_mips64_rpm": "11",
    "linux_x64_deb": "12",
    "linux_x64_AppImage": "13",
    "linux_arm64_deb": "14",
    "linux_arm64_AppImage": "15",
    "linux_mips64_deb": "16",

    "officeplugin_x86": "17",  # 17
    "officeplugin_x64": "18",  # 18
    "officeplugin_mac": "19",  # 19
}

OS_TYPE_ABBREVIATION_DICT = {
    "android": ["android"],
    "mac": ["mac"],
    "win": ["win32_advanced", "win64_advanced"],
    "ios": ["ios"],
    "linux": ["linux_x64_rpm", "linux_arm64_rpm", "linux_mips64_rpm", "linux_x64_deb", "linux_x64_AppImage", "linux_arm64_deb", "linux_arm64_AppImage", "linux_mips64_deb"],
    "office": ["officeplugin_x86", "officeplugin_x64", "officeplugin_mac"],
}

OS_TYPE = [
    "",  # 0
    "",  # 1
    "android",  # 2
    "mac",  # 3
    "win32_advanced",  # 4
    "",  # 5
    "",  # 6
    "ios",  # 7
    "win64_advanced",  # 8
    "linux_x64_rpm",  # 9
    "linux_arm64_rpm",  # 10
    "linux_mips64_rpm",  # 11
    "linux_x64_deb",  # 12
    "linux_x64_AppImage",  # 13
    "linux_arm64_deb",  # 14
    "linux_arm64_AppImage",  # 15
    "linux_mips64_deb",  # 16

    "officeplugin_x86",  # 17
    "officeplugin_x64",  # 18
    "officeplugin_mac",  # 19
]

OS_TYPE_DEFINE = {
    # 0
    # 1
    "2": "android",  # 2
    "3": "mac",  # 3
    "4": "windows32Advanced",  # 4
    "5": "windows64Advanced",  # 5
    "6": "",  # 6
    "7": "ios",  # 7
    "8": "win64_advanced",  # 8
    "9": "linuxX64Rpm",  # 9
    "10": "linuxArm64Rpm",  # 10
    "11": "linuxMips64Rpm",  # 11
    "12": "linuxX64Deb",  # 12
    "13": "linuxX64AppImage",  # 13
    "14": "linuxArm64Deb",  # 14
    "15": "linuxArm64AppImage",  # 15
    "16": "linuxMips64Deb",  # 16

    "17": "officePluginX86",  # 17
    "18": "officePluginX64",  # 18
    "19": "officePluginMac",  # 19
}

# 文件后缀名
FILE_SUFFIX = ["exe", "deb", "rpm", "AppImage", "dmg", "ipa", "apk", "tgz", "pkg"]

OBJECTID = [
    "f959419ebb6611e58ecb000c2999898e",  # 0
    "0476542cbb6711e58ecb000c2999898e",  # 1
    "0db2f914bb6711e58ecb000c2999898e",  # 2
    "17126ac6bb6711e58ecb000c2999898e",  # 3
    "d8d227d0977f4eac96d4b48b56b017d2",  # 4
    "4af61b3c70534b18b8bf8c5c98e46fc2",  # 5
    "dbae1be21d6f11e9accf005056821098",  # 6
    "a09c65261b1c95aeaae008e884b5ff8b",  # 7
    "d7404da29f3511eab5d600505682e601",  # 8
    "dd93e9669f3511eab5d600505682e601",  # 9
    "e5a027969f3511eab5d600505682e601",  # 10
    "e60962889f3511eab5d600505682e601",  # 11
    "64e7ad9e257811eb92f90050568227c1",  # 12
    "7989ffea257811eb92f90050568227c1",  # 13
    "130deac4257811eb92f90050568227c1",  # 14
    "14010010257811eb92f90050568227c1",  # 15
    "13975d2c257811eb92f90050568227c1",  # 16

    "3e2b3474aabb427d8cb8af84fe64f43c",  # 17
    "b205a9e9e05049b68d410ca23cfd4315",  # 18
    "f0a9668b5e334b6b8d1b4aa95e227209",  # 19
]

STANDARD = "standard"
COSTOM = "custom"

ERROR_RESULT_TEM = {"code": "", "message": "", "cause": ""}


class ConfigFile(Enum):
    """
    配置文件枚举类
    """
    # 服务所在命名空间配置
    NAMESPACE_INFO = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"


@unique
class MyHttpCode(Enum):
    """
    自定义状态码
    """

    # ------500-------
    NCT_UNKNOWN_ERROR = "500017000"  # 未知错误
    NCT_INGRESS_CONFIG_CONFLICT = "500017001"  # Ingress 配置冲突
    NCT_SITE_STATION_INFO = "500017002"  # 没有站点信息
    NCT_EXCE_COMMAND_FAILED = "500017003"  # 执行命令失败
    # NCT_GET_VIP_FAILED = '500017004'  # 获取vip失败
    # NCT_SYNCHRONIZATION_CERT_FAILED = '500017005'  # 同步证书失败
    # NCT_SYNCHRONIZATION_CERT_OSS_FAILED = '500017007'  # 同步存储证书失败
    # NCT_SET_SSL_FAILED = '500017006'  # 设置节点SSL失败
    NCT_NO_DEFAULT_STORAGE = "500017008"  # 未配置默认存储
    # NCT_GET_NODE_INFO_FAILED = '500017009'  # 获取节点信息失败
    NCT_GET_TIME_ZONE_FAILED = "500017010"  # 获取时区失败
    NCT_GET_OS_LANGUAGE_FAILED = "500017011"  # 获取系统语言失败
    NCT_GET_AS_LANGUAGE_FAILED = "500017012"  # 获取AS语言失败
    NCT_HYDRA_NOT_AVAILABLE = "500017020"  # hydra服务不可用
    # NCT_GET_DEVICE_INFO_FAILED = '500017021'  # 获取设备信息失败
    # 400
    NCT_GET_DOWNLOAD_URL_FAILED = "400017001"  # 获取下载地址失败
    NCT_INVALID_OPTION = "400017002"  # 非法的option
    # NCT_INVALID_VALUE = '400017003'  # 非法的value
    # NCT_OPTION_NOT_EXISTS = '400017004'  # option不存在
    NCT_PARAMS_ERROR = "400017005"  # 参数错误（缺少）
    NCT_INVALID_FILENAME = "400017006"  # 非法的文件名
    NCT_DEL_PKG_INFO_FAILED = "400017007"  # 删除升级包信息失败
    # NCT_UPLOAD_EOSS_FAILED = '400017008'  # 上传包失败
    NCT_INVALID_OS_PARAMETER = "400017009"  # 无效的os系统参数
    NCT_PACKAGE_NOT_EXIST = "400017010"  # 文件不存在
    # NCT_UPDATE_PACKAGE_ALREADY_EXIST = '400017011'  # 升级包已存在
    NCT_FILE_SUFFIX_ERROR = "400017012"  # 未知的文件后缀名
    # NCT_FILE_PREFIX_ERROR = '400017013'  # 未知的文件前缀
    NCT_OS_CANT_MATCH_SURFIX = "400017014"  # 系统和后缀名不匹配
    NCT_NO_CLIENT_PACKAGE_UPLOAD = "400017015"  # 无升级包上传
    # NCT_ALREADY_SET_DOWNLOAD_LINK = '400017016'  # 已配置下载地址（安卓）
    # NCT_INVALID_PACKAGE_VERSION = '400017017'  # 无效的升级包版本号
    NCT_INVALID_OS_SET_DOWNLOAD_URL = "400017018"  # 无效的OS系统设置下载地址（仅支持ios，安卓）
    # NCT_NO_STORAGE = '400017019'  # 未配置存储
    # NCT_INVALID_CERT_TYPE = '400017028'  # 无效的证书类型
    NCT_CERT_NOT_MATCH = "400017029"  # 证书不匹配
    NCT_UPDATE_TYPE_ERROR = "400017030"  # 无效的升级类型
    NCT_INVALID_VERSION = "400017033"  # 非法的版本号
    NCT_INVALID_MICRO_SERVICE_CONFIG = "400017034"  # 无效的微服务配置项
    NCT_INVALID_MICRO_SERVICE_VALUE = "400017035"  # 无效的微服务配置值
    NCT_INVALID_PARAMS_OF_BODY = "400017036"  # body中缺少参数
    # NCT_FILE_DAMAGED = '400017205'  # 上传文件损坏
    # NCT_UPLOAD_TASK_NOT_EXISTS = '400017204'  # 上传任务不存在
    # NCT_FILE_TYPE_NOT_SUPPORT = '400017206'  # 文件类型不支持
    NCT_FILE_PART_NOT_COMPLETE = "400017207"  # 分片数不完全
    NCT_INVALID_ACCESS_ADDR_TYPE = "400017101"  # 无效的访问地址类型
    NCT_INVALID_ACCESS_ADDR = "400017103"  # 无效的访问地址
    NCT_INVALID_PORT = "400017104"  # 无效的端口
    # NCT_INVALID_SERVICE_NAME = '400017234'  # 无效的服务名
    NCT_SERVICE_PACKAGE_NOT_EXIST = "400017235"  # 服务包不存在
    NCT_SERVICE_PACKAGE_VERSION_LOW = "400017236"  # 服务包版本低于当前版本
    # NCT_SERVICE_PACKAGE_FORMAT_INVALID = '400017237'  # 无效的服务包名
    # NCT_SERVICE_PACKAGE_NOT_COMPLETE = '400017238'  # 服务包不完整
    NCT_NODE_NOT_IN_CLUSTER = "400017239"  # 节点不在集群中
    NCT_SERVICE_NOT_INSTALLED = "400017240"  # 服务未安装
    # NCT_DELETE_SERVICE_NODE_NOT_ONE = '400017241'  # 删除节点不是一个
    # NCT_ES_CLUSTER_NOT_USE = '400017242'  # es集群状态不可用
    # NCT_NOT_FOUND_VERSION_NUM = '400017243'  # 找不到有效版本号
    NCT_MILVUS_INSTALLED = "400017244"  # 该节点已经安装milvus
    NCT_MILVUS_MORE_THAN_ONE = "400017245"  # 多个节点已经安装milvus
    NCT_CONNECT_OSS_ERROR = "400017246"  # 连接对象存储失败
    NCT_OSS_ALREADY_EXIST = "400017247"  # BUCKET已存在
    NCT_OSSGATEWAY_NOT_INSTALL = "400017248"  # OSS网关未安装
    NCT_THIRD_APP_SERVICE_EXIST_IN_AS = "400017251"  # 第三方服务已安装，无法删除对应依赖
    NCT_GET_VALUES_YAML_FAILED = "400018001"  # 获取微服务values.yaml失败
    NCT_UPGRADE_MICRO_SERVICE_RDS_FAILED = "400018002"  # 更新微服务rds失败
    NCT_UPGRADE_MICRO_SERVICE_ENV_TIMEZONE_FAILED = "400018003"  # 更新微服务系统时区失败
    NCT_UPGRADE_MICRO_SERVICE_ENV_LANGUAGE_FAILED = "400018004"  # 更新微服务系统语言失败
    NCT_UPGRADE_MICRO_SERVICE_SERVICE_LANGUAGE_FAILED = "400018005"  # 更新微服务全球化语言失败
    # NCT_UPGRADE_MICRO_SERVICE_DEP_SERVICE_FAILED = '400018006'  # 更新微服务依赖失败

    NCT_UPGRADE_SERVICE_RUNNING = "400018007"  # 服务正在运行
    # 第三方依赖服务配置错误码
    NCT_REQUIRE_THIRD_APP_DEPSERVICE_NOT_CONFIG = "400017201"  # 未配置第三方必填依赖
    # NCT_THIRD_APP_DEPSERVICE_NOT_EXISTS = '400017202'  # 第三方依赖信息不存在
    # NCT_NOT_FIND_MICRO_SERVICE_THIRD_DEPSERVICE = '400017203'  # 微服务中的依赖信息未找到
    # 404
    NCT_MICRO_SERVICE_NOT_FOUND = "404017001"  # 微服务不存在
    NCT_MICRO_SERVICE_ATTR_NOT_FOUND = "404017002"  # 微服务属性不存在
    NCT_ORIENTDB_NOT_FOUND = "404017003"
    NCT_MICRO_SERVICE_NOT_FOUND_IN_MODULE_SERVICE = "404017100"  # 模块服务中未包含该微服务
    # 423
    NCT_FORBIDDEN_UNINSTALL = "423017001"  # 禁止卸载服务
    NCT_FORBIDDEN_DUPLICATE_REQUEST = "423017002"  # 拒绝重复请求
    NCT_MICRO_SERVICE_ALREADY_INSTALLED_BY_OTHER_MODULE_SERVICE = "409017001"  # 微服务已被其他模块服务安装
    # URL 校验
    NCT_THIRD_PARTY_SERVICE_NOT_FOUND = "404017300"  # 第三方服务不存在
    NCT_ADDRESS_INACCESSIBLE = "500017300"  # 地址不可达
    # 401
    NCT_UNAUTHORIZED = "401017000"  # oauth验证未通过
