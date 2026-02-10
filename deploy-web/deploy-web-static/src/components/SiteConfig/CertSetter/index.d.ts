import { CertInfo } from "../index.d";

export interface Props extends React.ClassAttributes<void> {
    // 修改页面状态
    changePageState: (certType, pageState) => any;

    // 证书信息
    certInfo: CertInfo;
}

export interface State {
    // 签名类型
    signType: string;

    text: {
        // 密钥
        key: string;

        // 服务端证书
        serverCert: string;
    };

    validateState: {
        // 密钥
        key: string;

        // 服务端证书
        serverCert: string;
    };
}
