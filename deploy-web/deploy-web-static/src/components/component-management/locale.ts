import { i18nDeploy as i18n } from "../../core/mediator";

export default i18n([
    ["添加可选组件", "添加可選組件", "Add optional components"],
    ["添加-${component}", "新增-${component}", "Add-${component}"],
    ["编辑-${component}", "編輯-${component}", "Edit-${component}"],
    [
        "修改成功，配置将在服务更新后生效。请前往【服务管理-服务部署】页面更新服务。",
        "修改成功，設定將在服務更新後生效。請前往【服務管理-服務部署】頁面更新服務。",
        "Modified. This configuration will be effective after the service updates. You can update the following services on [Service Management-Service Deployment].",
    ],
    ["立即前往", "立即前往", "Go"],
    [
        "添加${component}成功",
        "新增{component}成功",
        "Add {component} successfully",
    ],
    [
        "设置${component}配置成功",
        "設定${component}成功",
        "${component} has been configured successfully",
    ],
    ["此项不允许为空。", "必填欄位。", "This field is required."],
    [
        "请输入1~65535范围内的整数。",
        "請輸入1~65535範圍內的整數。",
        "Enter integer from 1 to 65535.",
    ],
    [
        "请输入1~99范围内的整数。",
        "請輸入1~99範圍內的整數。",
        "Enter integer from 1 to 99.",
    ],
    [
        "请输入${originVal}~99范围内的整数。",
        "請輸入${originVal}~99範圍內的整數。",
        "Enter integer from ${originVal} to 99.",
    ],
    ["此项不允许为root。", "此項不允許為root。", "This field cannot be root."],
    ["提示", "提示", "Notes"],
    ["保存配置中...", "儲存設定中...", "Saving configuration..."],
    ["错误", "錯誤", "Error"],
    ["确定", "確定", "OK"],
    [
        "获取${errServices}失败",
        "獲取${errServices}失敗",
        "Failed to get the information of ${errServices}",
    ],
    ["单机模式", "單機模式", "Standalone Mode"],
    ["哨兵模式", "哨兵模式", "Sentinel Mode"],
    ["主从模式", "主從模式", "Master-slave Mode"],
    ["集群模式", "集群模式", "Cluster Mode"],
    [
        "达梦数据库的管理权限和普通账户信息必须相同。",
        "達夢數據庫的管理權限和普通帳戶信息必須相同。",
        "The Information of two accounts ( Admin and General ) for DM Database must be the same.",
    ],

    [
        "请输入非负整数。",
        "請輸入非負整數。",
        "Please enter a non-negative integer.",
    ],
    ["共${total}条", "共${total}條", "Total ${total} Item(s)"],
    ["组件名称", "組件名稱", "Component Name"],
    ["组件类型", "組件類型", "Component Type"],
    ["系统空间", "系統空間", "System Space"],
    ["系统空间ID", "系統空間ID", "System Space ID"],
    ["命名空间", "命名空間", "Namespace"],
    ["操作", "操作", "Action"],
    ["编辑", "編輯", "Edit"],
    [
        "名称只能包含英文或数字或-字符。",
        "名稱只能包含英文或數字或-字元。",
        "The name can only contain English letters, numbers, or hyphen (-) characters.",
    ],
]);
