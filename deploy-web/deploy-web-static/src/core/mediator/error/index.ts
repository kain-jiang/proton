import deploy from "./deploy";
import nodemgnt from "./nodemgnt";

/**
 * 弹窗类型
 */
export enum BoxType {
    // 成功
    success,
    // 警告
    alert,
    // 错误
    error,
    // 信息
    info,
}

// 错误模块
const errorModuleMap = {
    deploy: deploy,
    nodemgnt: nodemgnt,
};

// 例外错误
const extraErrorModuleMap = {
    deploy: ["500017021"],
    nodemgnt: [],
};

// 获取错误信息
const getErrorMessage = (
    reqModule: string,
    { errID, code, expMsg, message }: any
) => {
    const id = code || errID;
    const text = message || expMsg;

    if (
        id >= 500000000 &&
        extraErrorModuleMap[reqModule].every(
            (extraErrID: number) => id === extraErrID
        )
    ) {
        return text;
    } else {
        return errorModuleMap[reqModule][id];
    }
};

export default getErrorMessage;
