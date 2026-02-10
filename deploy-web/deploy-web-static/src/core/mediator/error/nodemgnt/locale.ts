import { i18nDeploy as i18n } from "../..";

export default i18n([
    ["访问IP已经被使用", "存取IP已被使用", "Access IP is in use"],
    [
        "当前节点不在集群中",
        "當前節點不在此叢集中",
        "This node doesn’t exist in the current cluster",
    ],
    [
        "当前节点已经是高可用节点，不能重复设置",
        "當前節點已經是高可用節點，不能重複設定",
        "This is HA Master Node. You cannot set it again",
    ],
    [
        "已经存在高可用主，不能重复设置",
        "已經存在高可用主節點，不能重複設定",
        "HA master node already exists in this cluster. You cannot set it again",
    ],
    [
        "网卡不存在或者网卡停用",
        "網路卡不存在或已停用",
        "Netcard doesn't exist or is out of service",
    ],
    [
        "已经存在三个高可用节点，无法设置更多节点",
        "已存在三個高可用節點，無法設定更多節點",
        "The number HA nodes cannot exceed 3",
    ],
    [
        "没有主节点，无法设置从节点",
        "沒有主節點，無法設定從節點",
        "You should set HA master node first",
    ],
]);
