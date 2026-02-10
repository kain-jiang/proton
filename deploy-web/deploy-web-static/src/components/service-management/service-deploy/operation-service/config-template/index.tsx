import React, { FC, useEffect, useMemo, useRef, useState } from "react";
import {
  Row,
  Col,
  Table,
  Button,
  Input,
  Tag,
  Empty,
  Upload,
  Spin,
  message,
  Modal,
} from "@kweaver-ai/ui";
import { CloseOutlined, InformationOutlined } from "@kweaver-ai/ui/lib/icons";
import type { InputRef } from "@kweaver-ai/ui";
import type { ColumnsType, ColumnType } from "@kweaver-ai/ui/es/table";
import { ContentLayout, Toolbar } from "../../../../common/components";
import { ConfigTemplateItem } from "../../../../../api/service-management/service-deploy/declare";
import { serviceApplication } from "../../../../../api/service-management/service-deploy";
import { handleError } from "../../../utils/handleError";
import { ReactComponent as DeleteTemplateIcon } from "../../../assets/DeleteTemplateIcon.svg";
import { noop } from "lodash";
import styles from "./styles.module.less";
import { SERVICE_PREFIX } from "../../../config";
import { safetyStr } from "../../../../common/utils/string";
import { defaultEmptyImg } from "./asset";
import { CommonEnum } from "../../../utils/common.type";
import { OperationType } from "../type";
import __ from "./locale";
import { ServiceMode } from "../../../../../core/service-management/service-deploy";

interface IProps {
  // 应用名称
  aname: string;
  // 应用版本
  aversion: string;
  // 选中的配置模板id
  selectedTid: number[];
  // 修改选中的配置模板id
  changeSelectedTid: (item: number[]) => void;
  // 操作服务的类型（安装或更新）
  operationType: OperationType;
  // 自定义版本
  customVersion: string;
  // 修改自定义版本
  changeCustomVersion: (item: string) => void;
}

export const ConfigTemplate: FC<IProps> = ({
  aname,
  aversion,
  selectedTid,
  changeSelectedTid,
  operationType,
  customVersion,
  changeCustomVersion,
}) => {
  // 标签筛选输入框
  const searchInput = useRef<InputRef>(null);
  // 配置模板列表数据
  const [data, setData] = useState<ConfigTemplateItem[]>([]);
  // 标签筛选的标签数组
  const [tagsTemp, setTagsTemp] = useState<string[]>([]);
  // 标签筛选确认的标签数组
  const [tagsConfirm, setTagsConfirm] = useState<string[]>([]);
  // 标签筛选输入框内容
  const [currentVal, setCurrentVal] = useState<string>("");
  // 选中的配置模板
  const [templateChosenData, setTemplateChosenData] = useState<
    ConfigTemplateItem[]
  >([]);
  // 过滤后的模板数据
  const [filteredTemplateData, setFilteredTemplateData] = useState<
    ConfigTemplateItem[]
  >([]);
  // 导入模板解析文件时的加载情况
  const [isLoading, setIsLoading] = useState<boolean>(false);

  const inputValueRef = useRef<string>("");

  // 获取配置模板列表和已选模板
  useEffect(() => {
    getConfigTemplates({ updateList: true, filterVersion });
  }, []);

  // 根据已选模板的数据控制所有模板列表的选中情况
  useEffect(() => {
    changeSelectedTid(templateChosenData.map((item) => item.tid!));
  }, [templateChosenData]);

  const filterVersion = useMemo(
    () => encodeURIComponent(customVersion || aversion),
    [customVersion, aversion]
  );

  // 获取所有配置模板数据，和标签过滤后的模板数据
  const getConfigTemplates = async ({
    updateList,
    tags,
    filterVersion,
    isUpdateTemplateChosenData = true,
  }: {
    updateList: boolean;
    tags?: string[];
    filterVersion: string;
    isUpdateTemplateChosenData?: boolean;
  }) => {
    try {
      // 更新配置模板列表数据和已选模板
      if (updateList) {
        const res = await serviceApplication.getConfigTemplate({
          offset: 0,
          limit: 10000,
          l: [],
          v: filterVersion,
          vt: 2,
          aname,
          count: "false",
        });
        const configData = res?.data || [];
        setData(configData);
        isUpdateTemplateChosenData &&
          setTemplateChosenData(
            selectedTid.map(
              (tid) => configData.find((data) => data.tid === tid)!
            )
          );
      }
      if (tags) {
        const res = await serviceApplication.getConfigTemplate({
          offset: 0,
          limit: 10000,
          l: tags.map((tag) => encodeURIComponent(tag)),
          v: filterVersion,
          vt: 2,
          aname,
          count: "false",
        });
        setFilteredTemplateData(res?.data || []);
      }
    } catch (error: any) {
      // 错误状态为400并且updateList为true（初次渲染组件）
      if (error.status === 400 && updateList) {
        Modal.confirm({
          wrapClassName: `${SERVICE_PREFIX}-custom-modal`,
          title: __("提示"),
          okText: __("确定"),
          content: (
            <>
              <div className={styles["modal-content"]}>
                {__("无法识别当前版本，您可以按照规范自行填写版本。")}
              </div>
              <div className={styles["modal-content-tip"]}>
                {__(
                  "版本格式需遵循语义化版本2.0.0规则，常用格式为：\n主版本号.次版本号.修订号，例如7.5.6"
                )}
              </div>
              <Input
                placeholder={__("请输入版本号")}
                onChange={(e) => {
                  inputValueRef.current = e.target.value;
                }}
              />
            </>
          ),
          onOk: handleCustomVersionOk,
          onCancel: () => {
            inputValueRef.current = "";
          },
          icon: (
            <InformationOutlined
              style={{ color: "#126EE3" }}
              onPointerEnterCapture={noop}
              onPointerLeaveCapture={noop}
            />
          ),
        });
      } else {
        handleError(error);
      }
    }
  };

  // 用户手动选择/取消某行
  const onSelectChange = (record: ConfigTemplateItem, selected: boolean) => {
    if (selected) {
      setTemplateChosenData([record, ...templateChosenData]);
    } else {
      setTemplateChosenData(
        templateChosenData.filter((item) => item.tid !== record.tid)
      );
    }
  };

  // 全选/取消全选的回调
  const onSelectAllChange = (
    selected: boolean,
    selectedRows: ConfigTemplateItem[],
    changeRows: ConfigTemplateItem[]
  ) => {
    if (selected) {
      setTemplateChosenData([...changeRows, ...templateChosenData]);
    } else {
      setTemplateChosenData(
        templateChosenData.filter((item) => {
          return changeRows.every((row) => {
            row.tid !== item.tid;
          });
        })
      );
    }
  };

  const rowSelection = {
    selectedRowKeys: selectedTid,
    onSelect: onSelectChange,
    onSelectAll: onSelectAllChange,
  };

  const handleConfirm = (
    confirm: (param?: any) => void,
    setSelectedKeys: any
  ) => {
    setSelectedKeys([...tagsTemp]);
    setTagsConfirm([...tagsTemp]);
    getConfigTemplates({
      updateList: false,
      tags: [...tagsTemp],
      filterVersion,
    });
    confirm();
  };

  const handlePressEnter = (e: any) => {
    if (e.target.value) {
      setTagsTemp([...tagsTemp, e.target.value as string]);
      setCurrentVal("");
    }
  };

  const handleCustomVersionOk = () => {
    const version = inputValueRef.current;
    changeCustomVersion(version);
    inputValueRef.current = "";
    getConfigTemplates({
      updateList: true,
      filterVersion: version,
    });
  };

  const removeTemplate = (record: ConfigTemplateItem) => {
    setTemplateChosenData(
      templateChosenData.filter((item) => item.tid !== record.tid)
    );
  };

  const handleUpload = (file: any) => {
    setIsLoading(true);
    const reader = new FileReader();
    reader.readAsText(file);
    reader.onload = async function () {
      try {
        const configTemplateData = JSON.parse(reader.result as string);
        try {
          const tid = await serviceApplication.postConfigTemplate(
            configTemplateData
          );
          getConfigTemplates({
            updateList: false,
            tags: tagsConfirm,
            filterVersion,
          });
          const res = await serviceApplication.getConfigTemplate({
            offset: 0,
            limit: 10000,
            l: [],
            v: filterVersion,
            vt: 2,
            aname,
            count: "false",
          });
          const newTemplateData = res?.data || [];
          setData(newTemplateData);
          // 如果导入的模板已经被选，不做处理
          if (!templateChosenData.some((item) => item.tid === tid)) {
            setTemplateChosenData([
              ...newTemplateData.filter((item) => item.tid === tid),
              ...templateChosenData,
            ]);
          }
          message.success(__("导入成功"));
        } catch (error) {
          handleError(error);
        }
      } catch (error) {
        // json文件转换失败处理
        Modal.info({
          title: __("导入失败"),
          okText: __("确定"),
          content: __("文件内容格式错误，无法解析。"),
        });
      } finally {
        setIsLoading(false);
      }
    };
    return false;
  };

  const handleDeleteClick = (record: ConfigTemplateItem) => {
    Modal.confirm({
      title: __("提示"),
      okText: __("确定"),
      content: __("您确定要删除此模板吗？删除后不可恢复。"),
      onOk: () => handleDeleteConfirm(record),
      icon: (
        <InformationOutlined
          style={{ color: "#126EE3" }}
          onPointerEnterCapture={noop}
          onPointerLeaveCapture={noop}
        />
      ),
    });
  };

  // 确认删除模板
  const handleDeleteConfirm = async (record: ConfigTemplateItem) => {
    try {
      await serviceApplication.deleteConfigTemplate(record.tid!);
      getConfigTemplates({
        updateList: true,
        tags: tagsConfirm,
        filterVersion,
        isUpdateTemplateChosenData: false,
      });
      setTemplateChosenData(
        templateChosenData.filter((item) => item.tid !== record.tid)
      );
      message.success(__("删除成功"));
    } catch (error) {
      handleError(error);
    }
  };

  const getColumnSearchProps = (): ColumnType<ConfigTemplateItem> => ({
    filterDropdown: ({ setSelectedKeys, confirm }) => {
      return (
        <div
          style={{ padding: "8px", overflow: "hidden" }}
          onKeyDown={(e) => e.stopPropagation()}
        >
          <div className={styles["filter-input-container"]}>
            <div className={styles["tag"]}>
              {tagsTemp.map((tag) => {
                return (
                  <Tag
                    key={tag}
                    closable
                    className={styles["tag-item"]}
                    onClose={() => {
                      setTagsTemp(tagsTemp.filter((item) => item != tag));
                    }}
                  >
                    {tag}
                  </Tag>
                );
              })}
            </div>
            <Input
              ref={searchInput}
              placeholder={tagsTemp.length ? "" : __("请输入标签名称")}
              bordered={false}
              className={styles["inputTag"]}
              value={currentVal}
              onChange={(e) => setCurrentVal(e.target.value)}
              onPressEnter={handlePressEnter}
            />
          </div>
          <Button
            className={styles["confirm-btn"]}
            size="small"
            onClick={() => handleConfirm(confirm, setSelectedKeys)}
          >
            {__("确定")}
          </Button>
        </div>
      );
    },
    onFilter: (value, record) => {
      return filteredTemplateData.some((templateData) => {
        return templateData.tid === record.tid;
      });
    },
    onFilterDropdownOpenChange: (visible) => {
      if (visible) {
        setTagsTemp([...tagsConfirm]);
        setCurrentVal("");
        setTimeout(() => searchInput.current?.select(), 100);
      }
    },
    filteredValue: tagsConfirm,
    render: (text, record) => {
      return (
        <>
          {record.labels?.length
            ? record.labels.map((label) => {
                return (
                  <Tag key={label} closable={false}>
                    {label}
                  </Tag>
                );
              })
            : CommonEnum.PLACEHOLDER}
        </>
      );
    },
    tooltip: (text, record) => {
      return (
        <>
          {record.labels?.length ? (
            <div className={styles["tag-tooltip-container"]}>
              {record.labels.map((label) => {
                return (
                  <Tag
                    key={label}
                    closable={false}
                    className={styles["tag-tooltip"]}
                  >
                    {label}
                  </Tag>
                );
              })}
            </div>
          ) : (
            CommonEnum.PLACEHOLDER
          )}
        </>
      );
    },
  });

  const templateListColumns: ColumnsType<ConfigTemplateItem> = [
    {
      title: __("模板名称"),
      dataIndex: "tname",
      render: (value: string) => safetyStr(value),
      tooltip: (value: string) => safetyStr(value),
    },
    {
      title: __("版本"),
      dataIndex: "tversion",
      render: (value: string) => safetyStr(value),
      tooltip: (value: string) => safetyStr(value),
    },
    {
      title: __("标签"),
      ...getColumnSearchProps(),
    },
    {
      title: __("描述"),
      dataIndex: "tdescription",
      render: (value: string) => safetyStr(value),
      tooltip: (value: string) => safetyStr(value),
    },
    {
      title: __("操作"),
      width: 100,
      render: (value, record: ConfigTemplateItem) => {
        return (
          <DeleteTemplateIcon
            style={{ cursor: "pointer" }}
            onClick={() => handleDeleteClick(record)}
          />
        );
      },
      tooltip: () => __("删除"),
    },
  ];

  const templateChosenColumns: ColumnsType<ConfigTemplateItem> = [
    {
      title: __("模板名称"),
      dataIndex: "tname",
      render: (value: string) => safetyStr(value),
      tooltip: (value: string) => safetyStr(value),
    },
    {
      title: __("版本"),
      dataIndex: "tversion",
      render: (value: string) => safetyStr(value),
      tooltip: (value: string) => safetyStr(value),
    },
    {
      title: __("操作"),
      render: (value, record) => {
        return (
          <CloseOutlined
            style={{ cursor: "pointer" }}
            onClick={() => removeTemplate(record)}
            onPointerEnterCapture={noop}
            onPointerLeaveCapture={noop}
          />
        );
      },
      tooltip: false,
    },
  ];

  const templateListHeader = (
    <Toolbar
      left={
        <Upload
          accept=".json"
          showUploadList={false}
          beforeUpload={handleUpload}
          disabled={isLoading}
          maxCount={1}
        >
          <Button type="default">{__("导入模板")}</Button>
        </Upload>
      }
      leftSize={24}
      moduleName={SERVICE_PREFIX}
    />
  );

  const templateChosenHeader = (
    <Toolbar
      left={<span className={styles["header-text"]}>{__("已选模板：")}</span>}
      right={
        <Button type="link" onClick={() => setTemplateChosenData([])}>
          {__("清空")}
        </Button>
      }
      cols={[{ span: 16 }, { span: 8 }]}
      moduleName={SERVICE_PREFIX}
    />
  );

  const windowHeight = window.innerHeight > 720 ? "100vh" : "720px";

  return (
    <>
      <div className={styles["text-tip"]}>
        {__(
          "选择多个模板后将自动合并配置。若配置存在冲突，后选中的模板配置将覆盖前者。下方将按照选中时间降序展示已选模板。"
        )}
      </div>
      {isLoading ? (
        <div className={styles["spin-container"]}>
          <Spin />
        </div>
      ) : (
        <Row gutter={40}>
          <Col span={16}>
            <ContentLayout
              header={templateListHeader}
              moduleName={SERVICE_PREFIX}
            >
              <Table
                rowKey={"tid"}
                rowSelection={rowSelection}
                columns={templateListColumns}
                dataSource={data}
                pagination={{
                  pageSize: data.length,
                  hideOnSinglePage: true,
                }}
                locale={{
                  emptyText: (
                    <Empty
                      image={<img src={defaultEmptyImg} />}
                      imageStyle={{
                        height: 144,
                      }}
                      description={
                        <span className={styles["table-empty-text"]}>
                          {__("暂无可用的配置模板")}
                        </span>
                      }
                    />
                  ),
                }}
                scroll={{
                  y:
                    operationType === ServiceMode.Install
                      ? `calc(${windowHeight} - 626px)`
                      : `calc(${windowHeight} - 490px)`,
                }}
              />
            </ContentLayout>
          </Col>
          <Col span={8}>
            <ContentLayout
              header={templateChosenHeader}
              moduleName={SERVICE_PREFIX}
            >
              <Table
                rowKey={"tid"}
                columns={templateChosenColumns}
                dataSource={templateChosenData}
                pagination={{
                  pageSize: data.length,
                  hideOnSinglePage: true,
                }}
                scroll={{
                  y:
                    operationType === ServiceMode.Install
                      ? `calc(${windowHeight} - 626px)`
                      : `calc(${windowHeight} - 490px)`,
                }}
                locale={{
                  emptyText: (
                    <Empty
                      image={<img src={defaultEmptyImg} />}
                      imageStyle={{
                        height: 144,
                      }}
                      description={
                        <span className={styles["table-empty-text"]}>
                          {__("暂未选择任何配置模板")}
                        </span>
                      }
                    />
                  ),
                }}
              />
            </ContentLayout>
          </Col>
        </Row>
      )}
    </>
  );
};
