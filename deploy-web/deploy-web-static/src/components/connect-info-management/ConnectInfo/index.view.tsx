import * as React from "react";
import { MQConnectInfo } from "./MQConnectInfo/component.view";
import { RDSConnectInfo } from "./RDSConnectInfo/component.view";
import { ETCDConnectInfo } from "./ETCDConnectInfo/component.view";
import { RedisConnectInfo } from "./RedisConnectInfo/component.view";
import { MongoDBConnectInfo } from "./MongoDBConnectInfo/component.view";
import { OpenSearchConnectInfo } from "./OpenSearchConnectInfo/component.view";
import { PolicyEngineConnectInfo } from "./PolicyEngineConnectInfo/component.view";
import { CONNECT_SERVICES } from "../../component-management/helper";
import { DataBaseConfigType } from "./index";

export const ConnectInfo: React.FC<DataBaseConfigType.Props> = React.memo(
    ({
        configData,
        connectInfoValidateState,
        onUpdateConnectInfo,
        updateConnectInfoForm,
        updateConnectInfoValidateState,
        originSourceType,
        originConnectInfoType,
    }) => {
        const updateConnectInfo = (key: string, value: any, type?: string) => {
            onUpdateConnectInfo(key, value, type);
        };

        const getConnectInfoTemplates = () => {
            return (
                <>
                    {configData?.resource_connect_info?.rds ? (
                        <RDSConnectInfo
                            configData={configData}
                            connectInfoValidateState={connectInfoValidateState}
                            updateConnectInfo={(o, val) =>
                                updateConnectInfo(CONNECT_SERVICES.RDS, o, val)
                            }
                            updateConnectInfoForm={(form) =>
                                updateConnectInfoForm({
                                    [CONNECT_SERVICES.RDS]: form,
                                })
                            }
                            updateConnectInfoValidateState={
                                updateConnectInfoValidateState
                            }
                            originSourceType={originSourceType}
                            originConnectInfoType={originConnectInfoType}
                        />
                    ) : null}
                    {configData?.resource_connect_info?.mongodb ? (
                        <MongoDBConnectInfo
                            configData={configData}
                            connectInfoValidateState={connectInfoValidateState}
                            updateConnectInfo={(o) =>
                                updateConnectInfo(CONNECT_SERVICES.MONGODB, o)
                            }
                            updateConnectInfoForm={(form) =>
                                updateConnectInfoForm({
                                    [CONNECT_SERVICES.MONGODB]: form,
                                })
                            }
                            updateConnectInfoValidateState={
                                updateConnectInfoValidateState
                            }
                            originSourceType={originSourceType}
                        />
                    ) : null}
                    {configData?.resource_connect_info?.redis ? (
                        <RedisConnectInfo
                            configData={configData}
                            connectInfoValidateState={connectInfoValidateState}
                            updateConnectInfo={(o, val) =>
                                updateConnectInfo(
                                    CONNECT_SERVICES.REDIS,
                                    o,
                                    val
                                )
                            }
                            updateConnectInfoForm={(form) =>
                                updateConnectInfoForm({
                                    [CONNECT_SERVICES.REDIS]: form,
                                })
                            }
                            updateConnectInfoValidateState={
                                updateConnectInfoValidateState
                            }
                            originSourceType={originSourceType}
                            originConnectInfoType={originConnectInfoType}
                        />
                    ) : null}
                    {configData?.resource_connect_info?.mq ? (
                        <MQConnectInfo
                            configData={configData}
                            connectInfoValidateState={connectInfoValidateState}
                            updateConnectInfo={(o, val) =>
                                updateConnectInfo(CONNECT_SERVICES.MQ, o, val)
                            }
                            updateConnectInfoForm={(form) =>
                                updateConnectInfoForm({
                                    [CONNECT_SERVICES.MQ]: form,
                                })
                            }
                            updateConnectInfoValidateState={
                                updateConnectInfoValidateState
                            }
                            originSourceType={originSourceType}
                            originConnectInfoType={originConnectInfoType}
                        />
                    ) : null}
                    {configData?.resource_connect_info?.opensearch ? (
                        <OpenSearchConnectInfo
                            configData={configData}
                            connectInfoValidateState={connectInfoValidateState}
                            updateConnectInfo={(o) =>
                                updateConnectInfo(
                                    CONNECT_SERVICES.OPENSEARCH,
                                    o
                                )
                            }
                            updateConnectInfoForm={(form) =>
                                updateConnectInfoForm({
                                    [CONNECT_SERVICES.OPENSEARCH]: form,
                                })
                            }
                            updateConnectInfoValidateState={
                                updateConnectInfoValidateState
                            }
                            originSourceType={originSourceType}
                        />
                    ) : null}

                    {configData?.resource_connect_info?.[
                        CONNECT_SERVICES.POLICY_ENGINE
                    ] ? (
                        <PolicyEngineConnectInfo
                            configData={configData}
                            updateConnectInfo={(o) =>
                                updateConnectInfo(
                                    CONNECT_SERVICES.POLICY_ENGINE,
                                    o
                                )
                            }
                            updateConnectInfoForm={(form) =>
                                updateConnectInfoForm({
                                    [CONNECT_SERVICES.POLICY_ENGINE]: form,
                                })
                            }
                        />
                    ) : null}
                    {configData?.resource_connect_info?.[
                        CONNECT_SERVICES.ETCD
                    ] ? (
                        <ETCDConnectInfo
                            configData={configData}
                            updateConnectInfo={(o) =>
                                updateConnectInfo(CONNECT_SERVICES.ETCD, o)
                            }
                            updateConnectInfoForm={(form) =>
                                updateConnectInfoForm({
                                    [CONNECT_SERVICES.ETCD]: form,
                                })
                            }
                        />
                    ) : null}
                </>
            );
        };

        return <div>{getConnectInfoTemplates()}</div>;
    }
);
