export interface Props extends React.ClassAttributes<any> {
    // 文档域访问配置信息
    documentConfigInfo: {
        // 访问地址
        host: string;
        // 端口
        port: string;
        // 前缀
        path: string;
        // 类型
        type: string;
    };
}
