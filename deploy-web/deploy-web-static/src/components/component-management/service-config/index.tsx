import React, { FC, useState, useRef, useEffect, useMemo } from "react";
import {
  Button,
  Col,
  Divider,
  Form,
  FormInstance,
  Input,
  InputNumber,
  Radio,
  Row,
  Select,
} from "@kweaver-ai/ui";
import {
  ConfigData,
  DefaultConfigData,
  NEBULA_COMPONENTS,
  OperationType,
  RESOURCES,
  RESOURCES_TYPE,
  SERVICES,
  SOURCE_TYPE,
  booleanEmptyValidatorRules,
  buildInComponentsText,
  emptyValidatorRules,
  getLogNumberValidatorRules,
  getReplicaValidatorRules,
  getUsernameValidatorRules,
  portValidatorRules,
} from "../helper";
import {
  DeleteOutlined,
  PlusOutlined,
  QuestionCircleOutlined,
} from "@kweaver-ai/ui/lib/icons";
import styles from "./styles.module.less";
import __ from "./locale";
import { noop } from "lodash";
import { ExternalServiceConfig } from "../index.d";

interface IProps {
  component: SERVICES;
  operationType: OperationType;
  componentConfigData: ConfigData;
  setComponentConfigData: any;
  sourceType: string;
  nodeOptions: string[];
  setComponentForm: (item: any) => void;
  showTitle: boolean;
  originReplicaCount: number;
  hasDatabaseConnectInfo?: boolean;
}

export const ServiceConfig: FC<IProps> = ({
  component,
  operationType,
  componentConfigData,
  setComponentConfigData,
  sourceType,
  nodeOptions,
  setComponentForm,
  showTitle,
  originReplicaCount,
  hasDatabaseConnectInfo,
}) => {
  const form = useRef<FormInstance>(null);
  useEffect(() => {
    setComponentForm((componentForm: any) => {
      return {
        ...componentForm,
        [component]: form,
      };
    });
  }, []);
  // 更新表单数据
  useEffect(() => {
    form.current?.setFieldsValue({
      ...componentConfigData[component]?.params,
    });
  }, [componentConfigData]);

  const onChangeConfig = (config: object) => {
    setComponentConfigData({
      ...componentConfigData,
      [component]: {
        ...componentConfigData[component],
        params: {
          ...componentConfigData[component]?.params,
          ...config,
        },
      },
    });
  };

  const onAddExternalServiceList = () => {
    const external_service_list = [
      ...(componentConfigData[component]?.params?.external_service_list || []),
      { name: "", port: null, nodePortBase: null, enableSSL: false },
    ];
    setComponentConfigData({
      ...componentConfigData,
      [component]: {
        ...componentConfigData[component],
        params: {
          ...componentConfigData[component]?.params,
          external_service_list: external_service_list,
        },
      },
    });
  };

  const onChangeExternalServiceList = (
    index: number,
    config: Partial<ExternalServiceConfig>
  ) => {
    const external_service_list = [
      ...(componentConfigData[component]?.params?.external_service_list || []),
    ];

    setComponentConfigData({
      ...componentConfigData,
      [component]: {
        ...componentConfigData[component],
        params: {
          ...componentConfigData[component]?.params,
          external_service_list: external_service_list.map(
            (item, itemIndex) => {
              if (index === itemIndex) {
                return {
                  ...item,
                  ...config,
                };
              } else {
                return { ...item };
              }
            }
          ),
        },
      },
    });
  };

  const onDeleteExternalServiceList = (index: number) => {
    const external_service_list = [
      ...(componentConfigData[component]?.params?.external_service_list || []),
    ];
    external_service_list.splice(index, 1);
    setComponentConfigData({
      ...componentConfigData,
      [component]: {
        ...componentConfigData[component],
        params: {
          ...componentConfigData[component]?.params,
          external_service_list: external_service_list,
        },
      },
    });
  };

  const onChangeEnv = (config: object) => {
    setComponentConfigData({
      ...componentConfigData,
      [component]: {
        ...componentConfigData[component],
        params: {
          ...componentConfigData[component]?.params,
          env: {
            ...componentConfigData[component]?.params?.env,
            ...config,
          },
        },
      },
    });
  };

  const onChangeParamsConfig = (config: object) => {
    setComponentConfigData({
      ...componentConfigData,
      [component]: {
        ...componentConfigData[component],
        params: {
          ...componentConfigData[component]?.params,
          config: {
            ...componentConfigData[component]?.params?.config,
            ...config,
          },
        },
      },
    });
  };

  const onChangeSettings = (config: object) => {
    setComponentConfigData({
      ...componentConfigData,
      [component]: {
        ...componentConfigData[component],
        params: {
          ...componentConfigData[component]?.params,
          settings: {
            ...componentConfigData[component]?.params?.settings,
            ...config,
          },
        },
      },
    });
  };

  const onChangeNFSConfig = (type: string, config: object) => {
    const newNFSConfig = {
      ...componentConfigData[component]?.params?.extraValues?.storage?.repo,
      [type]:
        type === "nfs" && Object.keys(config).includes("enabled")
          ? Object.values(config).includes(true)
            ? { enabled: true, server: "", path: "" }
            : { enabled: false }
          : {
              ...componentConfigData[component]?.params?.extraValues?.storage
                ?.repo?.[type],
              ...config,
            },
    };

    setComponentConfigData({
      ...componentConfigData,
      [component]: {
        ...componentConfigData[component],
        params: {
          ...componentConfigData[component]?.params,
          extraValues: {
            ...componentConfigData[component]?.params?.extraValues,
            storage: {
              ...componentConfigData[component]?.params?.extraValues?.storage,
              repo: {
                ...newNFSConfig,
              },
            },
          },
        },
      },
    });
  };

  const onChangeConfigResources = (val: boolean) => {
    if (val) {
      setComponentConfigData({
        ...componentConfigData,
        [component]: {
          ...componentConfigData[component],
          params: {
            ...componentConfigData[component]?.params,
            resources: (DefaultConfigData[component] as any).resources,
          },
        },
      });
    } else {
      setComponentConfigData((componentConfigData: any) => {
        const newComponentConfigData = { ...componentConfigData };
        delete newComponentConfigData[component].params.resources;
        return newComponentConfigData;
      });
    }
  };

  const onChangeResource = (resources: string, key: string, val: string) => {
    setComponentConfigData({
      ...componentConfigData,
      [component]: {
        ...componentConfigData[component],
        params: {
          ...componentConfigData[component]?.params,
          resources: {
            ...componentConfigData[component]?.params.resources,
            [resources]: {
              ...componentConfigData[component]?.params.resources[resources],
              [key]: val,
            },
          },
        },
      },
    });
  };
  const onChangeConfigComponentResources = (val: boolean) => {
    if (val) {
      setComponentConfigData({
        ...componentConfigData,
        [component]: {
          ...componentConfigData[component],
          params: {
            ...componentConfigData[component]?.params,
            exporter_resources: (DefaultConfigData[component] as any)
              .exporter_resources,
          },
        },
      });
    } else {
      setComponentConfigData((componentConfigData: any) => {
        const newComponentConfigData = { ...componentConfigData };
        delete newComponentConfigData[component].params.exporter_resources;
        return newComponentConfigData;
      });
    }
  };

  const onChangeComponentResource = (key: string, val: string) => {
    setComponentConfigData({
      ...componentConfigData,
      [component]: {
        ...componentConfigData[component],
        params: {
          ...componentConfigData[component]?.params,
          exporter_resources: {
            ...componentConfigData[component]?.params.exporter_resources,
            requests: {
              ...componentConfigData[component]?.params.exporter_resources
                .requests,
              [key]: val,
            },
          },
        },
      },
    });
  };

  const onChangeNebulaComponentResources = (
    nebulaComponent: string,
    resourcesType: string,
    config: object
  ) => {
    const originConfig = componentConfigData[component]?.params;
    let nebulaConfig: object;
    if (resourcesType === RESOURCES.ALL) {
      let resources = {
        [RESOURCES.LIMITS]: {
          [RESOURCES_TYPE.CPU]: "",
          [RESOURCES_TYPE.MEMORY]: "",
        },
        [RESOURCES.REQUESTS]: {
          [RESOURCES_TYPE.CPU]: "100m",
          [RESOURCES_TYPE.MEMORY]: "128Mi",
        },
      };

      nebulaConfig = {
        ...originConfig,
        [nebulaComponent]: {
          ...originConfig?.[nebulaComponent],
          resources: config ? resources : null,
        },
      };
    } else if (resourcesType === RESOURCES.LIMITS) {
      nebulaConfig = {
        ...originConfig,
        [nebulaComponent]: {
          ...originConfig?.[nebulaComponent],
          resources: {
            ...originConfig[nebulaComponent].resources,
            [RESOURCES.LIMITS]: {
              ...originConfig[nebulaComponent].resources[RESOURCES.LIMITS],
              ...config,
            },
          },
        },
      };
    } else {
      nebulaConfig = {
        ...originConfig,
        [nebulaComponent]: {
          ...originConfig?.[nebulaComponent],
          resources: {
            ...originConfig[nebulaComponent].resources,
            [RESOURCES.REQUESTS]: {
              ...originConfig[nebulaComponent].resources[RESOURCES.REQUESTS],
              ...config,
            },
          },
        },
      };
    }
    setComponentConfigData({
      ...componentConfigData,
      [component]: {
        ...componentConfigData[component],
        params: nebulaConfig,
      },
    });
  };

  const onChangeNebulaComponentConfig = (
    nebulaComponent: string,
    config: object
  ) => {
    const originConfig = componentConfigData[component]?.params;
    const nebulaConfig = {
      ...originConfig,
      [nebulaComponent]: {
        ...originConfig?.[nebulaComponent],
        config: {
          ...originConfig?.[nebulaComponent]?.config,
          ...config,
        },
      },
    };

    setComponentConfigData({
      ...componentConfigData,
      [component]: {
        ...componentConfigData[component],
        params: nebulaConfig,
      },
    });
  };

  const options = useMemo(() => {
    return nodeOptions.map((host) => {
      return { label: host, value: host, disabled: true };
    });
  }, [nodeOptions]);

  const handleNodeChange = (val: string[]) => {
    setComponentConfigData({
      ...componentConfigData,
      [component]: {
        ...componentConfigData[component],
        params: {
          ...componentConfigData[component]?.params,
          hosts: val,
        },
      },
    });
  };

  const config = componentConfigData[component]?.params;

  return (
    <>
      {[SERVICES.Kafka].includes(component) || showTitle ? (
        <div className={styles["component-title"]}>
          {__("${service}服务", {
            service: buildInComponentsText[component],
          })}
        </div>
      ) : null}
      <Form
        labelAlign="left"
        initialValues={config}
        validateTrigger="onBlur"
        ref={form}
      >
        {sourceType === SOURCE_TYPE.INTERNAL ? (
          <Form.Item
            labelCol={{ span: 4 }}
            label={__("部署节点")}
            name="hosts"
            required
            rules={emptyValidatorRules}
          >
            <div>
              <Select
                mode="tags"
                style={{ width: "95%" }}
                placeholder={__("请选择部署节点")}
                onChange={handleNodeChange}
                options={options}
                value={config?.hosts}
                notFoundContent={null}
                getPopupContainer={(node) =>
                  node.parentElement || document.body
                }
                disabled={
                  [SERVICES.ETCD, SERVICES.Nebula].includes(component) &&
                  operationType === OperationType.Edit
                }
              />
              <QuestionCircleOutlined
                onPointerEnterCapture={noop}
                onPointerLeaveCapture={noop}
                style={{
                  marginLeft: "6px",
                }}
                title={__("必须填写Kubernetes节点名称。")}
              />
            </div>
          </Form.Item>
        ) : (
          <Form.Item
            labelCol={{ span: 4 }}
            label={__("副本数")}
            name="replica_count"
            required
            rules={getReplicaValidatorRules(originReplicaCount)}
          >
            <InputNumber
              style={{
                width: "100%",
              }}
              value={config?.replica_count}
              onChange={(val) => {
                onChangeConfig({
                  replica_count: val,
                });
              }}
              disabled={
                [SERVICES.ETCD, SERVICES.Nebula].includes(component) &&
                operationType === OperationType.Edit
              }
            />
          </Form.Item>
        )}
        {component === SERVICES.MariaDB ? (
          <Row gutter={24}>
            <Col span={8}>
              <Form.Item
                label="Innodb_buffer_size:"
                name={["config", "innodb_buffer_pool_size"]}
                required
                rules={emptyValidatorRules}
              >
                <Input
                  style={{ width: "100px" }}
                  value={config?.config?.innodb_buffer_pool_size}
                  onChange={(e) => {
                    onChangeParamsConfig({
                      innodb_buffer_pool_size: e.target.value,
                    });
                  }}
                  disabled={operationType === OperationType.Edit}
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label="Requests.Memory:"
                name={["config", "resource_requests_memory"]}
                required
                rules={emptyValidatorRules}
              >
                <div>
                  <Input
                    style={{ width: "100px" }}
                    value={config?.config?.resource_requests_memory}
                    onChange={(e) => {
                      onChangeParamsConfig({
                        resource_requests_memory: e.target.value,
                      });
                    }}
                  />
                  <QuestionCircleOutlined
                    onPointerEnterCapture={noop}
                    onPointerLeaveCapture={noop}
                    style={{
                      marginLeft: "6px",
                    }}
                    title={__(
                      "填写规则为整数或浮点数+单位，如(Mi,Gi,Ti,M,G,T)。\n为保证服务正常运行，请满足：Requests.Memory ≤ Limits.Memory"
                    )}
                  />
                </div>
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label="Limits.Memory:"
                name={["config", "resource_limits_memory"]}
                required
                rules={emptyValidatorRules}
              >
                <div>
                  <Input
                    style={{ width: "100px" }}
                    value={config?.config?.resource_limits_memory}
                    onChange={(e) => {
                      onChangeParamsConfig({
                        resource_limits_memory: e.target.value,
                      });
                    }}
                  />
                  <QuestionCircleOutlined
                    onPointerEnterCapture={noop}
                    onPointerLeaveCapture={noop}
                    style={{
                      marginLeft: "6px",
                    }}
                    title={__(
                      "填写规则为整数或浮点数+单位，如(Mi,Gi,Ti,M,G,T)。\n为保证服务正常运行，请满足：Requests.Memory ≤ Limits.Memory"
                    )}
                  />
                </div>
              </Form.Item>
            </Col>
          </Row>
        ) : null}
        {[SERVICES.MariaDB, SERVICES.MongoDB, SERVICES.Redis].includes(
          component
        ) ? (
          <Row>
            <Col span={12}>
              <Form.Item
                labelCol={{ span: 8 }}
                label={__("管理账户")}
                name="admin_user"
                required
                rules={emptyValidatorRules}
              >
                <Input
                  style={{ width: "200px" }}
                  value={config?.admin_user}
                  onChange={(e) => {
                    onChangeConfig({
                      admin_user: e.target.value,
                    });
                  }}
                  disabled={operationType === OperationType.Edit}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                labelCol={{ span: 8 }}
                label={__("管理密码")}
                name="admin_passwd"
                required
                rules={
                  operationType === OperationType.Edit
                    ? undefined
                    : emptyValidatorRules
                }
              >
                <div>
                  <Input.Password
                    style={{ width: "200px" }}
                    value={
                      operationType === OperationType.Edit
                        ? "******"
                        : config?.admin_passwd
                    }
                    onChange={(e) => {
                      onChangeConfig({
                        admin_passwd: e.target.value,
                      });
                    }}
                    disabled={operationType === OperationType.Edit}
                  />
                  <QuestionCircleOutlined
                    onPointerEnterCapture={noop}
                    onPointerLeaveCapture={noop}
                    style={{
                      marginLeft: "6px",
                    }}
                    title={__(
                      "密码要求3种字符，支持大写、小写、数字、特殊字符（!@#$%^&*()_+-.=）。"
                    )}
                  />
                </div>
              </Form.Item>
            </Col>
          </Row>
        ) : null}
        {component === SERVICES.OpenSearch ? (
          <Row>
            <Col span={12}>
              <Form.Item
                labelCol={{ span: 8 }}
                label={__("模式")}
                name="mode"
                required
                rules={emptyValidatorRules}
              >
                <Input
                  style={{ width: "200px" }}
                  value={config?.mode}
                  onChange={(e) => {
                    onChangeConfig({
                      mode: e.target.value,
                    });
                  }}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                labelCol={{ span: 8 }}
                label={__("JVM配置")}
                name={["config", "jvmOptions"]}
                required
                rules={emptyValidatorRules}
              >
                <Input
                  style={{ width: "200px" }}
                  value={config?.config?.jvmOptions}
                  onChange={(e) => {
                    onChangeParamsConfig({
                      jvmOptions: e.target.value,
                    });
                  }}
                />
              </Form.Item>
            </Col>
          </Row>
        ) : null}
        {[SERVICES.Kafka].includes(component) ? (
          <Row>
            <Col span={12}>
              <Form.Item
                label={__("JVM配置")}
                labelCol={{ span: 8 }}
                name={["env", "KAFKA_HEAP_OPTS"]}
                required
                rules={emptyValidatorRules}
              >
                <Input
                  style={{ width: "200px" }}
                  value={config?.env?.["KAFKA_HEAP_OPTS"]}
                  onChange={(e) => {
                    onChangeEnv({
                      ["KAFKA_HEAP_OPTS"]: e.target.value,
                    });
                  }}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label={
                  <span
                    title={__("日志保留字节数")}
                    className={styles["form-label"]}
                  >
                    {__("日志保留字节数")}
                  </span>
                }
                labelCol={{ span: 8 }}
                name={["env", "KAFKA_LOG_RETENTION_BYTES"]}
                rules={getLogNumberValidatorRules(true)}
              >
                <Input
                  style={{ width: "200px" }}
                  placeholder={__("默认值：${default}", {
                    default: "-1",
                  })}
                  value={config?.env?.KAFKA_LOG_RETENTION_BYTES}
                  onChange={(e) => {
                    onChangeEnv({
                      KAFKA_LOG_RETENTION_BYTES: e.target.value || undefined,
                    });
                  }}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label={__("日志保留小时数")}
                labelCol={{ span: 8 }}
                name={["env", "KAFKA_LOG_RETENTION_HOURS"]}
                rules={getLogNumberValidatorRules(false)}
              >
                <Input
                  style={{ width: "200px" }}
                  placeholder={__("默认值：${default}", {
                    default: "168",
                  })}
                  value={config?.env?.KAFKA_LOG_RETENTION_HOURS}
                  onChange={(e) => {
                    onChangeEnv({
                      KAFKA_LOG_RETENTION_HOURS: e.target.value || undefined,
                    });
                  }}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label={
                  <span
                    title={__("日志段最大小时数")}
                    className={styles["form-label"]}
                  >
                    {__("日志段最大小时数")}
                  </span>
                }
                labelCol={{ span: 8 }}
                name={["env", "KAFKA_LOG_ROLL_HOURS"]}
                rules={getLogNumberValidatorRules(false)}
              >
                <Input
                  style={{ width: "200px" }}
                  placeholder={__("默认值：${default}", {
                    default: "24",
                  })}
                  value={config?.env?.KAFKA_LOG_ROLL_HOURS}
                  onChange={(e) => {
                    onChangeEnv({
                      KAFKA_LOG_ROLL_HOURS: e.target.value || undefined,
                    });
                  }}
                />
              </Form.Item>
            </Col>
          </Row>
        ) : null}
        {[SERVICES.Zookeeper].includes(component) ? (
          <Form.Item
            label={__("JVM配置")}
            labelCol={{ span: 4 }}
            name={["env", "JVMFLAGS"]}
            required
            rules={emptyValidatorRules}
          >
            <Input
              value={config?.env?.["JVMFLAGS"]}
              onChange={(e) => {
                onChangeConfig({
                  env: {
                    ["JVMFLAGS"]: e.target.value,
                  },
                });
              }}
            />
          </Form.Item>
        ) : null}
        {component !== SERVICES.Nebula ? (
          <Row>
            <Col span={12}>
              <Form.Item labelCol={{ span: 8 }} label={__("存储卷容量")}>
                <Input
                  style={{ width: "200px" }}
                  placeholder={
                    [SERVICES.MariaDB, SERVICES.MongoDB].includes(component)
                      ? __("默认值：${default}", {
                          default: "10Gi",
                        })
                      : undefined
                  }
                  value={config?.storage_capacity}
                  onChange={(e) => {
                    onChangeConfig({
                      storage_capacity: e.target.value,
                    });
                  }}
                />
                <QuestionCircleOutlined
                  onPointerEnterCapture={noop}
                  onPointerLeaveCapture={noop}
                  style={{
                    marginLeft: "6px",
                  }}
                  title={__("填写规则为整数或浮点数+单位，如(Mi,Gi,Ti)。")}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              {sourceType === SOURCE_TYPE.INTERNAL ? (
                <Form.Item
                  labelCol={{ span: 8 }}
                  label={__("数据路径")}
                  name="data_path"
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    value={config?.data_path}
                    onChange={(e) => {
                      onChangeConfig({
                        data_path: e.target.value,
                      });
                    }}
                    placeholder={DefaultConfigData[component]?.data_path}
                    disabled={operationType === OperationType.Edit}
                  />
                </Form.Item>
              ) : (
                <Form.Item
                  labelCol={{ span: 8 }}
                  label="storageClassName:"
                  name="storageClassName"
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    value={config?.storageClassName}
                    onChange={(e) => {
                      onChangeConfig({
                        storageClassName: e.target.value,
                      });
                    }}
                    disabled={operationType === OperationType.Edit}
                  />
                </Form.Item>
              )}
            </Col>
          </Row>
        ) : null}
        {component === SERVICES.OpenSearch ? (
          <>
            <Form.Item
              labelCol={{ span: 4 }}
              label={__("远程词库")}
              name={["config", "hanlpRemoteextDict"]}
            >
              <Input
                style={{
                  width: "95%",
                }}
                value={config?.config?.hanlpRemoteextDict}
                onChange={(e) => {
                  onChangeParamsConfig({
                    hanlpRemoteextDict: e.target.value,
                  });
                }}
              />
            </Form.Item>
            <Form.Item
              labelCol={{ span: 4 }}
              label={__("去停词")}
              name={["config", "hanlpRemoteextStopwords"]}
            >
              <Input
                style={{
                  width: "95%",
                }}
                value={config?.config?.hanlpRemoteextStopwords}
                onChange={(e) => {
                  onChangeParamsConfig({
                    hanlpRemoteextStopwords: e.target.value,
                  });
                }}
              />
            </Form.Item>
            <Form.Item
              labelCol={{ span: 4 }}
              label={__("低警戒水位线")}
              name={[
                "settings",
                "cluster.routing.allocation.disk.watermark.low",
              ]}
              required
              rules={emptyValidatorRules}
            >
              <div>
                <Input
                  style={{
                    width: "95%",
                  }}
                  value={
                    config?.settings?.[
                      "cluster.routing.allocation.disk.watermark.low"
                    ]
                  }
                  onChange={(e) => {
                    onChangeSettings({
                      ["cluster.routing.allocation.disk.watermark.low"]:
                        e.target.value,
                    });
                  }}
                />
                <QuestionCircleOutlined
                  onPointerEnterCapture={noop}
                  onPointerLeaveCapture={noop}
                  style={{
                    marginLeft: "6px",
                  }}
                  title={__(
                    "控制磁盘使用的低警戒水位线。当设置为百分比时，OpenSearch不会将分片分配给使用率超过该百分比磁盘的节点。这也可以设置为比值，如0.85。最后，也可以设置为字节值，如400mb。此设置不会影响新创建索引的主分片，但会阻止分配其副本。默认值为85%。\n为保证服务正常运行，请设置合理的警戒水位线：低警戒水位线<高警戒水位线<洪泛警戒水位线"
                  )}
                />
              </div>
            </Form.Item>
            <Form.Item
              labelCol={{ span: 4 }}
              label={__("高警戒水位线")}
              name={[
                "settings",
                "cluster.routing.allocation.disk.watermark.high",
              ]}
              required
              rules={emptyValidatorRules}
            >
              <div>
                <Input
                  style={{
                    width: "95%",
                  }}
                  value={
                    config?.settings?.[
                      "cluster.routing.allocation.disk.watermark.high"
                    ]
                  }
                  onChange={(e) => {
                    onChangeSettings({
                      ["cluster.routing.allocation.disk.watermark.high"]:
                        e.target.value,
                    });
                  }}
                />
                <QuestionCircleOutlined
                  onPointerEnterCapture={noop}
                  onPointerLeaveCapture={noop}
                  style={{
                    marginLeft: "6px",
                  }}
                  title={__(
                    "控制磁盘使用的高警戒水位线。当设置为百分比时，OpenSearch将尝试从磁盘使用率高于该百分比的节点重新迁移碎片。这也可以设置为比值，如0.85。最后，也可以设置为字节值，如400mb。此设置会影响所有碎片的分配。默认值为90%。\n为保证服务正常运行，请设置合理的警戒水位线：低警戒水位线<高警戒水位线<洪泛警戒水位线"
                  )}
                />
              </div>
            </Form.Item>
            <Form.Item
              labelCol={{ span: 4 }}
              label={__("洪泛警戒水位线")}
              name={[
                "settings",
                "cluster.routing.allocation.disk.watermark.flood_stage",
              ]}
              required
              rules={emptyValidatorRules}
            >
              <div>
                <Input
                  style={{
                    width: "95%",
                  }}
                  value={
                    config?.settings?.[
                      "cluster.routing.allocation.disk.watermark.flood_stage"
                    ]
                  }
                  onChange={(e) => {
                    onChangeSettings({
                      ["cluster.routing.allocation.disk.watermark.flood_stage"]:
                        e.target.value,
                    });
                  }}
                />
                <QuestionCircleOutlined
                  onPointerEnterCapture={noop}
                  onPointerLeaveCapture={noop}
                  style={{
                    marginLeft: "6px",
                  }}
                  title={__(
                    "控制磁盘使用的洪泛警戒水位线。这是防止节点耗尽磁盘空间的最后手段。当有一块磁盘超过洪泛警戒水位线时，OpenSearch会强制将位于该节点上所有分片的所有索引置为只读模式。一旦磁盘利用率低于高水位线，索引块就被释放。这也可以设置为比值，如0.85。最后，也可以设置为字节值，如400mb。默认值为95%。\n为保证服务正常运行，请设置合理的警戒水位线：低警戒水位线<高警戒水位线<洪泛警戒水位线"
                  )}
                />
              </div>
            </Form.Item>
            <Form.Item
              labelCol={{ span: 6 }}
              label="http.max_initial_line_length:"
              name={["settings", "http.max_initial_line_length"]}
              required
              rules={emptyValidatorRules}
            >
              <Input
                style={{
                  width: "95%",
                }}
                value={config?.settings?.["http.max_initial_line_length"]}
                onChange={(e) => {
                  onChangeSettings({
                    ["http.max_initial_line_length"]: e.target.value,
                  });
                }}
              />
            </Form.Item>
            <Form.Item
              labelCol={{ span: 6 }}
              label="cluster.max_shards_per_node:"
              name={["settings", "cluster.max_shards_per_node"]}
              required
              rules={emptyValidatorRules}
            >
              <Input
                style={{
                  width: "95%",
                }}
                value={config?.settings?.["cluster.max_shards_per_node"]}
                onChange={(e) => {
                  onChangeSettings({
                    ["cluster.max_shards_per_node"]: e.target.value,
                  });
                }}
              />
            </Form.Item>
            <Form.Item
              label={__("内存锁定")}
              name={["settings", "bootstrap.memory_lock"]}
              required
              rules={booleanEmptyValidatorRules}
            >
              <Radio.Group
                value={config?.settings?.["bootstrap.memory_lock"]}
                onChange={(e) => {
                  onChangeSettings({
                    ["bootstrap.memory_lock"]: e.target.value,
                  });
                }}
              >
                <Radio value={true}>{__("是")}</Radio>
                <Radio value={false}>{__("否")}</Radio>
              </Radio.Group>
            </Form.Item>
            <Form.Item
              label={__("开启NFS快照仓库")}
              name={["extraValues", "storage", "repo", "nfs", "enabled"]}
              required
              rules={booleanEmptyValidatorRules}
            >
              <Radio.Group
                value={config?.extraValues?.storage?.repo?.nfs?.enabled}
                onChange={(e) => {
                  onChangeNFSConfig("nfs", {
                    enabled: e.target.value,
                  });
                }}
              >
                <Radio value={true}>{__("是")}</Radio>
                <Radio value={false}>{__("否")}</Radio>
              </Radio.Group>
            </Form.Item>
            {config?.extraValues?.storage?.repo?.nfs?.enabled ? (
              <Row>
                <Col span={12}>
                  <Form.Item
                    label={__("NFS快照仓库IP")}
                    name={["extraValues", "storage", "repo", "nfs", "server"]}
                    required
                    rules={emptyValidatorRules}
                  >
                    <Input
                      style={{ width: "200px" }}
                      value={config?.extraValues?.storage?.repo?.nfs?.server}
                      onChange={(e) => {
                        onChangeNFSConfig("nfs", {
                          server: e.target.value,
                        });
                      }}
                    />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    label={__("NFS快照仓库路径")}
                    name={["extraValues", "storage", "repo", "nfs", "path"]}
                    required
                    rules={emptyValidatorRules}
                  >
                    <Input
                      style={{ width: "200px" }}
                      value={config?.extraValues?.storage?.repo?.nfs?.path}
                      onChange={(e) => {
                        onChangeNFSConfig("nfs", {
                          path: e.target.value,
                        });
                      }}
                    />
                  </Form.Item>
                </Col>
              </Row>
            ) : null}
            <Form.Item
              label={__("开启HDFS快照仓库")}
              name={["extraValues", "storage", "repo", "hdfs", "enabled"]}
              required
              rules={booleanEmptyValidatorRules}
            >
              <Radio.Group
                value={config?.extraValues?.storage?.repo?.hdfs?.enabled}
                onChange={(e) => {
                  onChangeNFSConfig("hdfs", {
                    enabled: e.target.value,
                  });
                }}
              >
                <Radio value={true}>{__("是")}</Radio>
                <Radio value={false}>{__("否")}</Radio>
              </Radio.Group>
            </Form.Item>
          </>
        ) : null}
        {component === SERVICES.Kafka ? (
          <>
            <Form.Item
              label={__("禁用开放外部端口")}
              name="disable_external_service"
              required
              rules={booleanEmptyValidatorRules}
            >
              <div>
                <Radio.Group
                  value={config?.disable_external_service}
                  onChange={(e) => {
                    onChangeConfig({
                      disable_external_service: e.target.value,
                    });
                  }}
                >
                  <Radio value={true}>{__("是")}</Radio>
                  <Radio value={false}>{__("否")}</Radio>
                </Radio.Group>
                <QuestionCircleOutlined
                  style={{
                    marginLeft: "6px",
                  }}
                  title={__("未禁用时将根据下列端口信息进行开放。")}
                  onPointerEnterCapture={noop}
                  onPointerLeaveCapture={noop}
                />
              </div>
            </Form.Item>
            <Form.Item
              label={__("外部端口信息")}
              name="external_service_list"
              required
              style={{ marginBottom: "0" }}
            >
              <Button
                type="default"
                icon={
                  <PlusOutlined
                    onPointerEnterCapture={noop}
                    onPointerLeaveCapture={noop}
                  />
                }
                style={{ marginBottom: "24px" }}
                onClick={() => onAddExternalServiceList()}
              >
                {__("新增")}
              </Button>
              {config?.external_service_list.map(
                (serviceInfo: ExternalServiceConfig, index: number) => {
                  return (
                    <>
                      <Row>
                        <Col span={12}>
                          <Form.Item
                            label={__("名称")}
                            name={["external_service_list", index, "name"]}
                            required
                            rules={emptyValidatorRules}
                          >
                            <Input
                              style={{
                                width: "200px",
                              }}
                              value={serviceInfo?.name}
                              onChange={(e) => {
                                onChangeExternalServiceList(index, {
                                  name: e.target.value,
                                });
                              }}
                            />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item
                            label={__("地址")}
                            name={["external_service_list", index, "ip"]}
                          >
                            <Input
                              style={{
                                width: "200px",
                              }}
                              value={serviceInfo?.ip}
                              onChange={(e) => {
                                onChangeExternalServiceList(index, {
                                  ip: e.target.value,
                                });
                              }}
                            />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item
                            label={__("端口")}
                            name={["external_service_list", index, "port"]}
                            required
                            rules={portValidatorRules}
                          >
                            <div>
                              <InputNumber
                                style={{
                                  width: "200px",
                                }}
                                value={serviceInfo?.port}
                                onChange={(value) => {
                                  onChangeExternalServiceList(index, {
                                    port: value as number,
                                  });
                                }}
                              />
                              <QuestionCircleOutlined
                                style={{
                                  marginLeft: "6px",
                                }}
                                title={__("将开放kafka的此端口。")}
                                onPointerEnterCapture={noop}
                                onPointerLeaveCapture={noop}
                              />
                            </div>
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item
                            label={__("节点Base端口")}
                            name={[
                              "external_service_list",
                              index,
                              "nodePortBase",
                            ]}
                            required
                            rules={portValidatorRules}
                          >
                            <div>
                              <InputNumber
                                style={{
                                  width: "200px",
                                }}
                                value={serviceInfo?.nodePortBase}
                                onChange={(value) => {
                                  onChangeExternalServiceList(index, {
                                    nodePortBase: value as number,
                                  });
                                }}
                              />
                              <QuestionCircleOutlined
                                style={{
                                  marginLeft: "6px",
                                }}
                                title={__(
                                  "每个节点都将开放一个端口，从此端口开始依次递增。"
                                )}
                                onPointerEnterCapture={noop}
                                onPointerLeaveCapture={noop}
                              />
                            </div>
                          </Form.Item>
                        </Col>
                        <Col span={24}>
                          <Form.Item
                            label="TLS"
                            name={["external_service_list", index, "enableSSL"]}
                            required
                            rules={emptyValidatorRules}
                          >
                            <Radio.Group
                              value={serviceInfo?.enableSSL}
                              onChange={(e) => {
                                onChangeExternalServiceList(index, {
                                  enableSSL: e.target.value,
                                });
                              }}
                            >
                              <Radio value={true}>{__("是")}</Radio>
                              <Radio value={false}>{__("否")}</Radio>
                            </Radio.Group>
                            <Button
                              icon={
                                <DeleteOutlined
                                  onPointerEnterCapture={noop}
                                  onPointerLeaveCapture={noop}
                                />
                              }
                              type="link"
                              disabled={
                                config?.external_service_list?.length <= 1
                              }
                              style={{
                                marginLeft: "16px",
                                color: "black",
                              }}
                              onClick={() => onDeleteExternalServiceList(index)}
                            />
                          </Form.Item>
                        </Col>
                      </Row>
                      {index !== config.external_service_list.length - 1 ? (
                        <div className={styles["list-split"]}></div>
                      ) : null}
                    </>
                  );
                }
              )}
            </Form.Item>
          </>
        ) : null}
        {[
          SERVICES.MongoDB,
          SERVICES.Redis,
          SERVICES.PolicyEngine,
          SERVICES.OpenSearch,
          SERVICES.Kafka,
          SERVICES.Zookeeper,
        ].includes(component) ? (
          <>
            <Row>
              <Col span={12}>
                <Form.Item label={__("自定义配置资源限制")}>
                  <Radio.Group
                    value={!!config?.resources}
                    onChange={(e) => {
                      onChangeConfigResources(e.target.value);
                    }}
                  >
                    <Radio value={true}>{__("是")}</Radio>
                    <Radio value={false}>{__("否")}</Radio>
                  </Radio.Group>
                </Form.Item>
              </Col>
            </Row>
            {config?.resources ? (
              <>
                {config?.resources?.limits &&
                JSON.stringify(config?.resources?.limits) !== "{}" ? (
                  <Row gutter={24}>
                    <Col span={12}>
                      <Form.Item
                        label="Limits.CPU:"
                        name={["resources", "limits", "cpu"]}
                        required
                        rules={emptyValidatorRules}
                      >
                        <div>
                          <Input
                            style={{
                              width: "200px",
                            }}
                            value={config?.resources?.limits?.cpu}
                            onChange={(e) => {
                              onChangeResource("limits", "cpu", e.target.value);
                            }}
                          />
                          <QuestionCircleOutlined
                            onPointerEnterCapture={noop}
                            onPointerLeaveCapture={noop}
                            style={{
                              marginLeft: "6px",
                            }}
                            title={__(
                              "填写规则为整数或浮点数+单位，如(C,m)。\n为保证服务正常运行，请满足：Requests.CPU ≤ Limits.CPU"
                            )}
                          />
                        </div>
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        label="Limits.Memory:"
                        name={["resources", "limits", "memory"]}
                        required
                        rules={emptyValidatorRules}
                      >
                        <div>
                          <Input
                            style={{
                              width: "200px",
                            }}
                            value={config?.resources?.limits?.memory}
                            onChange={(e) => {
                              onChangeResource(
                                "limits",
                                "memory",
                                e.target.value
                              );
                            }}
                          />
                          <QuestionCircleOutlined
                            onPointerEnterCapture={noop}
                            onPointerLeaveCapture={noop}
                            style={{
                              marginLeft: "6px",
                            }}
                            title={__(
                              "填写规则为整数或浮点数+单位，如(Mi,Gi,Ti,M,G,T)。\n为保证服务正常运行，请满足：Requests.Memory ≤ Limits.Memory"
                            )}
                          />
                        </div>
                      </Form.Item>
                    </Col>
                  </Row>
                ) : null}
                <Row gutter={24}>
                  <Col span={12}>
                    <Form.Item
                      label="Requests.CPU:"
                      name={["resources", "requests", "cpu"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={config?.resources?.requests?.cpu}
                          onChange={(e) => {
                            onChangeResource("requests", "cpu", e.target.value);
                          }}
                        />
                        {config?.resources?.limits ? (
                          <QuestionCircleOutlined
                            onPointerEnterCapture={noop}
                            onPointerLeaveCapture={noop}
                            style={{
                              marginLeft: "6px",
                            }}
                            title={__(
                              "填写规则为整数或浮点数+单位，如(C,m)。\n为保证服务正常运行，请满足：Requests.CPU ≤ Limits.CPU"
                            )}
                          />
                        ) : null}
                      </div>
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item
                      label="Requests.Memory:"
                      name={["resources", "requests", "memory"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={config?.resources?.requests?.memory}
                          onChange={(e) => {
                            onChangeResource(
                              "requests",
                              "memory",
                              e.target.value
                            );
                          }}
                        />
                        {config?.resources?.limits ? (
                          <QuestionCircleOutlined
                            onPointerEnterCapture={noop}
                            onPointerLeaveCapture={noop}
                            style={{
                              marginLeft: "6px",
                            }}
                            title={__(
                              "填写规则为整数或浮点数+单位，如(Mi,Gi,Ti,M,G,T)。\n为保证服务正常运行，请满足：Requests.Memory ≤ Limits.Memory"
                            )}
                          />
                        ) : null}
                      </div>
                    </Form.Item>
                  </Col>
                </Row>
              </>
            ) : null}
          </>
        ) : null}
        {[SERVICES.Kafka, SERVICES.OpenSearch].includes(component) ? (
          <>
            <Form.Item label={__("Exporter自定义配置资源限制")}>
              <Radio.Group
                value={!!config?.exporter_resources}
                onChange={(e) => {
                  onChangeConfigComponentResources(e.target.value);
                }}
              >
                <Radio value={true}>{__("是")}</Radio>
                <Radio value={false}>{__("否")}</Radio>
              </Radio.Group>
            </Form.Item>
            {config?.exporter_resources ? (
              <Row gutter={24}>
                <Col span={12}>
                  <Form.Item
                    label="Requests.CPU:"
                    name={["exporter_resources", "requests", "cpu"]}
                    required
                    rules={emptyValidatorRules}
                  >
                    <Input
                      style={{ width: "200px" }}
                      value={config?.exporter_resources?.requests?.cpu}
                      onChange={(e) => {
                        onChangeComponentResource("cpu", e.target.value);
                      }}
                    />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    label="Requests.Memory:"
                    name={["exporter_resources", "requests", "memory"]}
                    required
                    rules={emptyValidatorRules}
                  >
                    <Input
                      style={{ width: "200px" }}
                      value={config?.exporter_resources?.requests?.memory}
                      onChange={(e) => {
                        onChangeComponentResource("memory", e.target.value);
                      }}
                    />
                  </Form.Item>
                </Col>
              </Row>
            ) : null}
          </>
        ) : null}
        {component === SERVICES.Nebula ? (
          <>
            <Form.Item labelCol={{ span: 4 }} label={__("密码")}>
              <Input.Password
                style={{ width: "200px" }}
                value={
                  operationType === OperationType.Edit
                    ? "******"
                    : config?.password
                }
                onChange={(e) => {
                  onChangeConfig({
                    password: e.target.value,
                  });
                }}
                disabled={operationType === OperationType.Edit}
              />
              <QuestionCircleOutlined
                onPointerEnterCapture={noop}
                onPointerLeaveCapture={noop}
                style={{
                  marginLeft: "6px",
                }}
                title={__(
                  "Nebula Graph 的 root 帐户的密码，长度不超过 24。如果为空则使用生成的随机密码。"
                )}
              />
            </Form.Item>
            <Form.Item
              labelCol={{ span: 4 }}
              label={__("数据路径")}
              name="data_path"
              required
              rules={emptyValidatorRules}
            >
              <Input
                value={config?.data_path}
                onChange={(e) => {
                  onChangeConfig({
                    data_path: e.target.value,
                  });
                }}
                placeholder={DefaultConfigData[component]?.data_path}
                disabled={operationType === OperationType.Edit}
              />
            </Form.Item>

            <Form.Item label={__("Metad自定义配置")}></Form.Item>
            <Row>
              <Col span={12}>
                <Form.Item
                  labelAlign="left"
                  label="memory_tracker_limitratio:"
                  name={["metad", "config", "memory_tracker_limitratio"]}
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    placeholder={__("默认值：${default}", {
                      default: "0.99",
                    })}
                    value={config?.metad?.config?.memory_tracker_limitratio}
                    onChange={(e) => {
                      onChangeNebulaComponentConfig(NEBULA_COMPONENTS.METAD, {
                        memory_tracker_limitratio: e.target.value,
                      });
                    }}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  labelAlign="left"
                  label="system_memory_high_watermark_ratio:"
                  name={[
                    "metad",
                    "config",
                    "system_memory_high_watermark_ratio",
                  ]}
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    placeholder={__("默认值：${default}", {
                      default: "0.99",
                    })}
                    value={
                      config?.metad?.config?.system_memory_high_watermark_ratio
                    }
                    onChange={(e) => {
                      onChangeNebulaComponentConfig(NEBULA_COMPONENTS.METAD, {
                        system_memory_high_watermark_ratio: e.target.value,
                      });
                    }}
                  />
                </Form.Item>
              </Col>
            </Row>
            <Form.Item
              labelAlign="left"
              labelCol={{ span: 4 }}
              label="enable_authorize:"
              name={["metad", "config", "enable_authorize"]}
              required
              rules={emptyValidatorRules}
            >
              <Radio.Group
                value={config?.metad?.config?.enable_authorize}
                onChange={(e) => {
                  onChangeNebulaComponentConfig(NEBULA_COMPONENTS.METAD, {
                    enable_authorize: e.target.value,
                  });
                }}
              >
                <Radio value={"true"}>{__("是")}</Radio>
                <Radio value={"false"}>{__("否")}</Radio>
              </Radio.Group>
            </Form.Item>

            <Form.Item label={__("graphd自定义配置")}></Form.Item>
            <Row>
              <Col span={12}>
                <Form.Item
                  labelAlign="left"
                  label="memory_tracker_limitratio:"
                  name={["graphd", "config", "memory_tracker_limitratio"]}
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    placeholder={__("默认值：${default}", {
                      default: "0.99",
                    })}
                    value={config?.graphd?.config?.memory_tracker_limitratio}
                    onChange={(e) => {
                      onChangeNebulaComponentConfig(NEBULA_COMPONENTS.GRAPHD, {
                        memory_tracker_limitratio: e.target.value,
                      });
                    }}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  labelAlign="left"
                  label="system_memory_high_watermark_ratio:"
                  name={[
                    "graphd",
                    "config",
                    "system_memory_high_watermark_ratio",
                  ]}
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    placeholder={__("默认值：${default}", {
                      default: "0.99",
                    })}
                    value={
                      config?.graphd?.config?.system_memory_high_watermark_ratio
                    }
                    onChange={(e) => {
                      onChangeNebulaComponentConfig(NEBULA_COMPONENTS.GRAPHD, {
                        system_memory_high_watermark_ratio: e.target.value,
                      });
                    }}
                  />
                </Form.Item>
              </Col>
            </Row>
            <Form.Item
              labelAlign="left"
              labelCol={{ span: 4 }}
              label="enable_authorize:"
              name={["graphd", "config", "enable_authorize"]}
              required
              rules={emptyValidatorRules}
            >
              <Radio.Group
                value={config?.graphd?.config?.enable_authorize}
                onChange={(e) => {
                  onChangeNebulaComponentConfig(NEBULA_COMPONENTS.GRAPHD, {
                    enable_authorize: e.target.value,
                  });
                }}
              >
                <Radio value={"true"}>{__("是")}</Radio>
                <Radio value={"false"}>{__("否")}</Radio>
              </Radio.Group>
            </Form.Item>

            <Form.Item label={__("Storaged自定义配置")}></Form.Item>
            <Row>
              <Col span={12}>
                <Form.Item
                  labelAlign="left"
                  label="memory_tracker_limitratio:"
                  name={["storaged", "config", "memory_tracker_limitratio"]}
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    placeholder={__("默认值：${default}", {
                      default: "0.99",
                    })}
                    value={config?.storaged?.config?.memory_tracker_limitratio}
                    onChange={(e) => {
                      onChangeNebulaComponentConfig(
                        NEBULA_COMPONENTS.STORAGED,
                        {
                          memory_tracker_limitratio: e.target.value,
                        }
                      );
                    }}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  labelAlign="left"
                  label="system_memory_high_watermark_ratio:"
                  name={[
                    "storaged",
                    "config",
                    "system_memory_high_watermark_ratio",
                  ]}
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    placeholder={__("默认值：${default}", {
                      default: "0.99",
                    })}
                    value={
                      config?.storaged?.config
                        ?.system_memory_high_watermark_ratio
                    }
                    onChange={(e) => {
                      onChangeNebulaComponentConfig(
                        NEBULA_COMPONENTS.STORAGED,
                        {
                          system_memory_high_watermark_ratio: e.target.value,
                        }
                      );
                    }}
                  />
                </Form.Item>
              </Col>
            </Row>
            <Form.Item
              labelAlign="left"
              labelCol={{ span: 4 }}
              label="enable_authorize:"
              name={["storaged", "config", "enable_authorize"]}
              required
              rules={emptyValidatorRules}
            >
              <Radio.Group
                value={config?.storaged?.config?.enable_authorize}
                onChange={(e) => {
                  onChangeNebulaComponentConfig(NEBULA_COMPONENTS.STORAGED, {
                    enable_authorize: e.target.value,
                  });
                }}
              >
                <Radio value={"true"}>{__("是")}</Radio>
                <Radio value={"false"}>{__("否")}</Radio>
              </Radio.Group>
            </Form.Item>

            <Form.Item label={__("Metad自定义配置资源限制")}>
              <Radio.Group
                value={!!config?.metad?.resources}
                onChange={(e) => {
                  onChangeNebulaComponentResources(
                    NEBULA_COMPONENTS.METAD,
                    RESOURCES.ALL,
                    e.target.value
                  );
                }}
              >
                <Radio value={true}>{__("是")}</Radio>
                <Radio value={false}>{__("否")}</Radio>
              </Radio.Group>
            </Form.Item>
            {config?.metad?.resources ? (
              <>
                <Row gutter={24}>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Limits.CPU:"
                      name={["metad", "resources", "limits", "cpu"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={config?.metad?.resources?.limits?.cpu}
                          onChange={(e) => {
                            onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.METAD,
                              RESOURCES.LIMITS,
                              {
                                [RESOURCES_TYPE.CPU]: e.target.value,
                              }
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          onPointerEnterCapture={noop}
                          onPointerLeaveCapture={noop}
                          style={{
                            marginLeft: "6px",
                          }}
                          title={__(
                            "填写规则为整数或浮点数+单位，如(C,m)。\n为保证服务正常运行，请满足：Requests.CPU ≤ Limits.CPU"
                          )}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Limits.Memory:"
                      name={["metad", "resources", "limits", "memory"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={config?.metad?.resources?.limits?.memory}
                          onChange={(e) => {
                            onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.METAD,
                              RESOURCES.LIMITS,
                              {
                                [RESOURCES_TYPE.MEMORY]: e.target.value,
                              }
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          onPointerEnterCapture={noop}
                          onPointerLeaveCapture={noop}
                          style={{
                            marginLeft: "6px",
                          }}
                          title={__(
                            "填写规则为整数或浮点数+单位，如(Mi,Gi,Ti,M,G,T)。\n为保证服务正常运行，请满足：Requests.Memory ≤ Limits.Memory"
                          )}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                </Row>
                <Row gutter={24}>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Requests.CPU:"
                      name={["metad", "resources", "requests", "cpu"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={config?.metad?.resources?.requests?.cpu}
                          onChange={(e) => {
                            onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.METAD,
                              RESOURCES.REQUESTS,
                              {
                                [RESOURCES_TYPE.CPU]: e.target.value,
                              }
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          onPointerEnterCapture={noop}
                          onPointerLeaveCapture={noop}
                          style={{
                            marginLeft: "6px",
                          }}
                          title={__(
                            "填写规则为整数或浮点数+单位，如(C,m)。\n为保证服务正常运行，请满足：Requests.CPU ≤ Limits.CPU"
                          )}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Requests.Memory:"
                      name={["metad", "resources", "requests", "memory"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={config?.metad?.resources?.requests?.memory}
                          onChange={(e) => {
                            onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.METAD,
                              RESOURCES.REQUESTS,
                              {
                                [RESOURCES_TYPE.MEMORY]: e.target.value,
                              }
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          onPointerEnterCapture={noop}
                          onPointerLeaveCapture={noop}
                          style={{
                            marginLeft: "6px",
                          }}
                          title={__(
                            "填写规则为整数或浮点数+单位，如(Mi,Gi,Ti,M,G,T)。\n为保证服务正常运行，请满足：Requests.Memory ≤ Limits.Memory"
                          )}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                </Row>
              </>
            ) : null}
            <Form.Item label={__("graphd自定义配置资源限制")}>
              <Radio.Group
                value={!!config?.graphd?.resources}
                onChange={(e) => {
                  onChangeNebulaComponentResources(
                    NEBULA_COMPONENTS.GRAPHD,
                    RESOURCES.ALL,
                    e.target.value
                  );
                }}
              >
                <Radio value={true}>{__("是")}</Radio>
                <Radio value={false}>{__("否")}</Radio>
              </Radio.Group>
            </Form.Item>
            {config?.graphd?.resources ? (
              <>
                <Row gutter={24}>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Limits.CPU:"
                      name={["graphd", "resources", "limits", "cpu"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={config?.graphd?.resources?.limits?.cpu}
                          onChange={(e) => {
                            onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.GRAPHD,
                              RESOURCES.LIMITS,
                              {
                                [RESOURCES_TYPE.CPU]: e.target.value,
                              }
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          onPointerEnterCapture={noop}
                          onPointerLeaveCapture={noop}
                          style={{
                            marginLeft: "6px",
                          }}
                          title={__(
                            "填写规则为整数或浮点数+单位，如(C,m)。\n为保证服务正常运行，请满足：Requests.CPU ≤ Limits.CPU"
                          )}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Limits.Memory:"
                      name={["graphd", "resources", "limits", "memory"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={config?.graphd?.resources?.limits?.memory}
                          onChange={(e) => {
                            onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.GRAPHD,
                              RESOURCES.LIMITS,
                              {
                                [RESOURCES_TYPE.MEMORY]: e.target.value,
                              }
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          onPointerEnterCapture={noop}
                          onPointerLeaveCapture={noop}
                          style={{
                            marginLeft: "6px",
                          }}
                          title={__(
                            "填写规则为整数或浮点数+单位，如(Mi,Gi,Ti,M,G,T)。\n为保证服务正常运行，请满足：Requests.Memory ≤ Limits.Memory"
                          )}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                </Row>
                <Row gutter={24}>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Requests.CPU:"
                      name={["graphd", "resources", "requests", "cpu"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={config?.graphd?.resources?.requests?.cpu}
                          onChange={(e) => {
                            onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.GRAPHD,
                              RESOURCES.REQUESTS,
                              {
                                [RESOURCES_TYPE.CPU]: e.target.value,
                              }
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          onPointerEnterCapture={noop}
                          onPointerLeaveCapture={noop}
                          style={{
                            marginLeft: "6px",
                          }}
                          title={__(
                            "填写规则为整数或浮点数+单位，如(C,m)。\n为保证服务正常运行，请满足：Requests.CPU ≤ Limits.CPU"
                          )}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Requests.Memory:"
                      name={["graphd", "resources", "requests", "memory"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={config?.graphd?.resources?.requests?.memory}
                          onChange={(e) => {
                            onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.GRAPHD,
                              RESOURCES.REQUESTS,
                              {
                                [RESOURCES_TYPE.MEMORY]: e.target.value,
                              }
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          onPointerEnterCapture={noop}
                          onPointerLeaveCapture={noop}
                          style={{
                            marginLeft: "6px",
                          }}
                          title={__(
                            "填写规则为整数或浮点数+单位，如(Mi,Gi,Ti,M,G,T)。\n为保证服务正常运行，请满足：Requests.Memory ≤ Limits.Memory"
                          )}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                </Row>
              </>
            ) : null}
            <Form.Item label={__("Storaged自定义配置资源限制")}>
              <Radio.Group
                value={!!config?.storaged?.resources}
                onChange={(e) => {
                  onChangeNebulaComponentResources(
                    NEBULA_COMPONENTS.STORAGED,
                    RESOURCES.ALL,
                    e.target.value
                  );
                }}
              >
                <Radio value={true}>{__("是")}</Radio>
                <Radio value={false}>{__("否")}</Radio>
              </Radio.Group>
            </Form.Item>
            {config?.storaged?.resources ? (
              <>
                <Row gutter={24}>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Limits.CPU:"
                      name={["storaged", "resources", "limits", "cpu"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={config?.storaged?.resources?.limits?.cpu}
                          onChange={(e) => {
                            onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.STORAGED,
                              RESOURCES.LIMITS,
                              {
                                [RESOURCES_TYPE.CPU]: e.target.value,
                              }
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          onPointerEnterCapture={noop}
                          onPointerLeaveCapture={noop}
                          style={{
                            marginLeft: "6px",
                          }}
                          title={__(
                            "填写规则为整数或浮点数+单位，如(C,m)。\n为保证服务正常运行，请满足：Requests.CPU ≤ Limits.CPU"
                          )}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Limits.Memory:"
                      name={["storaged", "resources", "limits", "memory"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={config?.storaged?.resources?.limits?.memory}
                          onChange={(e) => {
                            onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.STORAGED,
                              RESOURCES.LIMITS,
                              {
                                [RESOURCES_TYPE.MEMORY]: e.target.value,
                              }
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          onPointerEnterCapture={noop}
                          onPointerLeaveCapture={noop}
                          style={{
                            marginLeft: "6px",
                          }}
                          title={__(
                            "填写规则为整数或浮点数+单位，如(Mi,Gi,Ti,M,G,T)。\n为保证服务正常运行，请满足：Requests.Memory ≤ Limits.Memory"
                          )}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                </Row>
                <Row gutter={24}>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Requests.CPU:"
                      name={["storaged", "resources", "requests", "cpu"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={config?.storaged?.resources?.requests?.cpu}
                          onChange={(e) => {
                            onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.STORAGED,
                              RESOURCES.REQUESTS,
                              {
                                [RESOURCES_TYPE.CPU]: e.target.value,
                              }
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          onPointerEnterCapture={noop}
                          onPointerLeaveCapture={noop}
                          style={{
                            marginLeft: "6px",
                          }}
                          title={__(
                            "填写规则为整数或浮点数+单位，如(C,m)。\n为保证服务正常运行，请满足：Requests.CPU ≤ Limits.CPU"
                          )}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Requests.Memory:"
                      name={["storaged", "resources", "requests", "memory"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={config?.storaged?.resources?.requests?.memory}
                          onChange={(e) => {
                            onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.STORAGED,
                              RESOURCES.REQUESTS,
                              {
                                [RESOURCES_TYPE.MEMORY]: e.target.value,
                              }
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          onPointerEnterCapture={noop}
                          onPointerLeaveCapture={noop}
                          style={{
                            marginLeft: "6px",
                          }}
                          title={__(
                            "填写规则为整数或浮点数+单位，如(Mi,Gi,Ti,M,G,T)。\n为保证服务正常运行，请满足：Requests.Memory ≤ Limits.Memory"
                          )}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                </Row>
              </>
            ) : null}
          </>
        ) : null}
        {[SERVICES.MariaDB, SERVICES.MongoDB].includes(component) ? (
          <>
            <Divider orientation="left" orientationMargin="0">
              {__("连接配置")}
            </Divider>
            <div className={styles["component-title"]}>{__("账户信息")}</div>
            <Row>
              <Col span={12}>
                <Form.Item
                  labelCol={{ span: 4 }}
                  labelAlign="left"
                  label={__("用户名")}
                  name="username"
                  required
                  rules={getUsernameValidatorRules(SOURCE_TYPE.INTERNAL)}
                >
                  <Input
                    style={{ width: "200px" }}
                    value={config?.username}
                    onChange={(e) => {
                      onChangeConfig({
                        username: e.target.value,
                      });
                    }}
                    disabled={
                      operationType === OperationType.Edit &&
                      hasDatabaseConnectInfo
                    }
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  labelCol={{ span: 4 }}
                  labelAlign="left"
                  label={__("密码")}
                  name="password"
                  required
                  rules={
                    operationType === OperationType.Edit &&
                    hasDatabaseConnectInfo
                      ? undefined
                      : emptyValidatorRules
                  }
                >
                  <div>
                    <Input.Password
                      style={{ width: "200px" }}
                      value={
                        operationType === OperationType.Edit &&
                        hasDatabaseConnectInfo
                          ? "******"
                          : config?.password
                      }
                      onChange={(e) => {
                        onChangeConfig({
                          password: e.target.value,
                        });
                      }}
                      disabled={
                        operationType === OperationType.Edit &&
                        hasDatabaseConnectInfo
                      }
                    />
                    <QuestionCircleOutlined
                      onPointerEnterCapture={noop}
                      onPointerLeaveCapture={noop}
                      style={{
                        marginLeft: "6px",
                      }}
                      title={__(
                        "密码要求3种字符，支持大写、小写、数字、特殊字符（!@#$%^&*()_+-.=）。"
                      )}
                    />
                  </div>
                </Form.Item>
              </Col>
            </Row>
          </>
        ) : null}
      </Form>
    </>
  );
};
