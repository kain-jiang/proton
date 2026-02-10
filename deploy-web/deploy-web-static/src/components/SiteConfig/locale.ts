import { i18nDeploy as i18n } from "../../core/mediator";

export default i18n([
    ["自签名证书", "自簽名證書", "Self-signed Certificate"],
    ["CA证书", "CA證書", "CA Certificate"],
    ["配置证书", "設定證書", "Configure"],
    ["证书配置", "證書設定", "Certificate Configuration"],
    ["访问配置", "存取設定", "Access Configuration"],
    ["修改访问配置", "修改存取設定", "Modify address"],
    [
        "当前配置对所有产品的访问地址生效，包含AnyShare的文档域地址，AnyBackup，AnyRobot，AnyDATA，AnyFabric的访问地址。",
        "當前設定對所有產品的存取位址生效，包含AnyShare的文件網域位址，AnyBackup，AnyRobot，AnyDATA，AnyFabric的存取位址。",
        "This configuration is available to all products' access addresses, including the document domain address of AnyShare, and the access address of AnyBackup, AnyRobot, AnyDATA and AnyFabric.",
    ],
    [
        "一个文档域对应一套产品服务系统",
        "一個文件網域對應一套產品服務系統",
        "One Document Domain is one complete product system",
    ],
    [
        "用户需要在各个终端输入文档域地址，来访问相应的终端。如果修改文档域地址，原地址将失效，无法继续登录。",
        "使用者需要在各個終端輸入文件網域位址，來存取相應的終端。如果修改文件網域位址，遠位址將失效，無法繼續登入。",
        "End users are required to input Doc Domain Address in each client to access the product service. If you change the address here, the original one will be invalid. ",
    ],
    [
        "勾选后，用户可在登录页下载证书及相关文档进行安装。",
        "勾選後，使用者可在登入頁下載證書及相關文件進行安裝。",
        "If checked, users can download certificate and relevant documents on the login page to install.",
    ],
    [
        "确定要关闭网页端证书下载入口吗？",
        "確定要關閉網頁端證書下載入口嗎？",
        "Are you sure to hide the certificate download on login page?",
    ],
    [
        "关闭后，终端用户将无法在网页端登录页下载证书及相关文档进行安装。",
        "關閉后，終端使用者將無法在網頁端登入頁下載證書及相關文件進行安裝。",
        "After this, the end users will not be able to download the certificate or the installation documents.",
    ],
    [
        "关闭证书下载入口成功",
        "關閉證書下載入口成功",
        "Certificate Download is hided successfully",
    ],
    [
        "开启证书下载入口成功",
        "啟用證書下載入口成功",
        "Show Certificate Download successfully",
    ],
    ["错误", "錯誤", "Error"],
]);
