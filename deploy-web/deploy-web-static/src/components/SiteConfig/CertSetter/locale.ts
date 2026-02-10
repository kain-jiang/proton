import { i18nDeploy as i18n } from "../../../core/mediator";

export default i18n([
    ["选择证书类型：", "選擇證書類型：", "Select Type: "],
    ["生成自签名证书", "產生自簽名證書", "Generate"],
    ["密钥：", "秘密金鑰：", "Key: "],
    ["服务器证书：", "伺服器證書：", "Server Certificate: "],
    ["浏览", "瀏覽", "Browse"],
    ["上传", "上傳", "Upload"],
    ["取消", "取消", "Cancel"],
    ["生成证书成功", "產生證書成功", "Successful"],
    ["上传成功", "上傳成功", "Successful"],
    [
        "检测到您配置的CA证书还未失效，生成自签名证书后将覆盖原有证书，是否生成自签名证书？",
        "檢測到您設定的CA證書還未失效，生成自簽名證書後將複寫原有證書，是否生成自簽名證書？",
        "An effective CA certificate is detected, which will be overwritten by this operation. Are you sure to continue?",
    ],
    [
        "检测到您配置的CA证书还未失效，上传后将覆盖原有证书，是否上传？",
        "檢測到您設定的CA證書還未失效，上傳後將複寫原有證書，是否上傳？",
        "An effective CA certificate is detected, which will be overwritten by this operation. Are you sure to continue?",
    ],
    ["请先上传密钥。", "請先上傳秘密金鑰。", "Key needs uploading first."],
    [
        "请先上传服务器证书。",
        "請先上傳伺服器證書。",
        "Server Certificate needs uploading first.",
    ],
    [
        "系统自动生成${certType}${signType} 成功",
        "系统自動產生${certType}${signType} 成功",
        "System automatically generates ${certType}${signType} Successfully",
    ],
    [
        "本地上传${certType} 成功",
        "本機上傳${certType} 成功",
        "Upload Local ${certType} Certificate Successfully",
    ],
    ["提示", "提示", "Tips"],
    ["错误", "錯誤", "Error"],
]);
