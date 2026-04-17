import * as React from "react";
import Success from "../Success/component.view";
import CRConfig from "./CRConfig/component.view";
import NodeConfig from "./NodeConfig/component.view";
import { ConnectInfo } from "./ConnectInfo/index.view";
import NetworkConfig from "./NetWorkConfig/component.view";
import { DataBaseConfig } from "./DataBaseConfig/index.view";
import {
  Layout,
  Steps,
  Button,
  Space,
  Spin,
  ConfigProvider as AiConfigProvider,
} from "@aishutech/ui";
import zh_CN from "antd/lib/locale-provider/zh_CN";
import { ChooseTemplate } from "../DeployTools/ChooseTemplate/component.view";
import {
  OptionSteps,
  DataBaseStorageType,
  ConfigEditStatus,
  CHRONY_MODE,
  DefaultConnectInfoValidateState,
} from "./helper";
import DeployToolsBase from "./component.base";
import "./styles.view.scss";

const { Header, Content, Footer } = Layout;
const { Step } = Steps;

export default class DeployTools extends DeployToolsBase {
  render(): React.ReactNode {
    const { configEditStatus, dataBaseStorageType, configData } = this.state;

    return configEditStatus === ConfigEditStatus.Success ? (
      <Success />
    ) : (
      <AiConfigProvider locale={zh_CN}>
        <Layout className="container">
          <Header className="header">
            <h2 className="title-font">部署工具</h2>
          </Header>
          {dataBaseStorageType ? (
            this.renderContent()
          ) : (
            <ChooseTemplate
              configData={configData}
              changeDataBaseStorageType={this.changeDataBaseStorageType.bind(
                this,
              )}
              updateDeploy={this.updateDeploy.bind(this)}
            />
          )}
        </Layout>
        {configEditStatus === ConfigEditStatus.initing ? (
          <div className="waiting-init">
            <Spin size="large" tip="初始化中, 请耐心等待..."></Spin>
          </div>
        ) : null}
      </AiConfigProvider>
    );
  }

  /**
   * 渲染内容部分
   * @returns
   */
  renderContent() {
    const { stepStatus, dataBaseStorageType } = this.state;
    return (
      <>
        <Content className="site-layout-background main-content">
          {dataBaseStorageType === DataBaseStorageType.Standard ? (
            <div className="steps-position">
              <Steps
                type="navigation"
                labelPlacement="vertical"
                current={stepStatus}
                initial={OptionSteps.NodeConfig}
              >
                <Step title="节点配置" description="配置集群节点" />
                <Step
                  title="kubernetes配置"
                  description="设置kubernetes及docker"
                />
                <Step title="仓库配置" description="配置容器仓库" />
                <Step
                  title="基础服务配置"
                  description="设置存储服务、消息中间件等"
                />
                <Step title="连接配置" description="设置本地、第三方连接配置" />
              </Steps>
            </div>
          ) : (
            <div className="steps-position">
              <Steps
                type="navigation"
                labelPlacement="vertical"
                current={stepStatus - 1}
              >
                <Step
                  title="kubernetes配置"
                  description="设置kubernetes及docker"
                />
                <Step title="仓库配置" description="配置容器仓库" />
                <Step
                  title="基础服务配置"
                  description="设置存储服务、消息中间件等"
                />
                <Step title="连接配置" description="设置本地、第三方连接配置" />
              </Steps>
            </div>
          )}
          <div className="component-content">{this.getContentForm()}</div>
        </Content>
        <Footer className="foot-position">{this.getSubmitButton()}</Footer>
      </>
    );
  }

  /**
   * 获取按钮状态
   */
  getSubmitButton() {
    const {
      sshAccount,
      sshPassword,
      stepStatus,
      configData,
      nextStepButtonDisable,
      dataBaseStorageType,
    } = this.state;

    switch (stepStatus) {
      case OptionSteps.NodeConfig:
        return (
          <Button
            size="large"
            type="primary"
            disabled={nextStepButtonDisable}
            onClick={() => {
              this.checkNodeConfig();
            }}
          >
            {" "}
            下一步
          </Button>
        );
      case OptionSteps.NetworkConfig:
        return (
          <Space>
            {dataBaseStorageType !== DataBaseStorageType.DepositKubernetes ? (
              <Button
                size="large"
                onClick={() => {
                  this.updateNetworkNodesValidateState();
                  this.onChangeStepStatus(OptionSteps.NodeConfig);
                }}
              >
                {" "}
                上一步
              </Button>
            ) : null}
            <Button
              size="large"
              type="primary"
              onClick={() => {
                this.checkNetworkConfig();
              }}
            >
              {" "}
              下一步
            </Button>
          </Space>
        );
      case OptionSteps.RepositoryConfig:
        return (
          <Space>
            <Button
              size="large"
              onClick={() => {
                this.updateCRNodesValidateState();
                this.onChangeStepStatus(OptionSteps.NetworkConfig);
              }}
            >
              {" "}
              上一步
            </Button>
            <Button
              size="large"
              type="primary"
              onClick={() => {
                this.checkRepositoryConfig();
              }}
            >
              {" "}
              下一步
            </Button>
          </Space>
        );
      case OptionSteps.DataBaseConfig:
        return (
          <Space>
            <Button
              size="large"
              onClick={() => {
                this.updateGrafanaNodesValidateState();
                this.updatePrometheusNodesValidateState();
                this.updateMonitorNodesValidateState();
                this.onChangeStepStatus(OptionSteps.RepositoryConfig);
              }}
            >
              {" "}
              上一步
            </Button>
            <Button
              size="large"
              type="primary"
              onClick={() => {
                this.checkDataBaseConfig();
              }}
            >
              {" "}
              下一步
            </Button>
          </Space>
        );
      case OptionSteps.ConnectInfo:
        return (
          <Space>
            <Button
              size="large"
              onClick={() => {
                this.updateConnectInfoValidateState({
                  ...DefaultConnectInfoValidateState,
                });
                this.onChangeStepStatus(OptionSteps.DataBaseConfig);
              }}
            >
              {" "}
              上一步
            </Button>
            <Button
              size="large"
              type="primary"
              onClick={() => {
                this.checkConnectInfoConfig();
              }}
            >
              {" "}
              完成
            </Button>
          </Space>
        );
    }
  }

  /**
   * 获取内容表单
   * @returns
   */
  getContentForm() {
    const {
      stepStatus,
      configData,
      sshAccount,
      sshPassword,
      selectableServices,
      addableServices,
      dataBaseStorageType,
      nodesValidateState,
      networkNodesValidateState,
      crNodesValidateState,
      grafanaNodesValidateState,
      prometheusNodesValidateState,
      monitorNodesValidateState,
      connectInfoValidateState,
    } = this.state;

    switch (stepStatus) {
      case OptionSteps.NodeConfig:
        return (
          <NodeConfig
            dataBaseStorageType={dataBaseStorageType}
            updateChrony={this.updateChrony.bind(this)}
            updateFirewall={this.updateFirewall.bind(this)}
            updateNodesInfo={this.updateNodesInfo.bind(this)}
            updateSSHInfo={this.onChangeSSHInfo.bind(this)}
            updateNicCidr={this.onUpdateConfigInfo.bind(this)}
            updateNetworkConfig={this.updateNetworkConfig.bind(this)}
            updateNodeForm={this.updateNodeForm.bind(this)}
            updateNodesValidateState={this.updateNodesValidateState.bind(this)}
            setNextStepButtonDisable={this.setNextStepButtonDisable.bind(this)}
            configData={configData}
            accountInfo={{
              sshAccount: sshAccount,
              sshPassword: sshPassword,
            }}
            nodesValidateState={nodesValidateState}
            // ipConfig={{
            //   internal_cidr: configData.internal_cidr,
            //   internal_nic: configData.internal_nic,
            // }}
          />
        );
      case OptionSteps.NetworkConfig:
        return (
          <NetworkConfig
            configData={configData}
            dataBaseStorageType={dataBaseStorageType}
            networkNodesValidateState={networkNodesValidateState}
            onUpdateNetworkConfig={this.onUpdateConfigInfo.bind(this)}
            updateNetworkForm={this.updateNetworkForm.bind(this)}
            updateNetworkNodesValidateState={this.updateNetworkNodesValidateState.bind(
              this,
            )}
            updateDeploy={this.updateDeploy.bind(this)}
          />
        );
      case OptionSteps.RepositoryConfig:
        return (
          <CRConfig
            cRType={this.crType}
            configData={configData}
            dataBaseStorageType={dataBaseStorageType}
            crNodesValidateState={crNodesValidateState}
            onUpdateCRConfig={this.onUpdateCRConfigInfo.bind(this)}
            onUpDateCRTypeConfig={this.onUpDateCRTypeConfig.bind(this)}
            updateCRForm={this.updateCRForm.bind(this)}
            updateCRNodesValidateState={this.updateCRNodesValidateState.bind(
              this,
            )}
          />
        );
      case OptionSteps.DataBaseConfig:
        return (
          <DataBaseConfig
            configData={configData}
            dataBaseStorageType={dataBaseStorageType}
            addableServices={addableServices}
            selectableServices={selectableServices}
            grafanaNodesValidateState={grafanaNodesValidateState}
            prometheusNodesValidateState={prometheusNodesValidateState}
            monitorNodesValidateState={monitorNodesValidateState}
            onAddService={this.onAddService.bind(this)}
            onDeleteService={this.onDeleteService.bind(this)}
            onUpdateDataBaseConfig={this.onUpdateConfigInfo.bind(this)}
            onUpdateConnectInfo={this.onUpdateConnectInfo.bind(this)}
            updateDataBaseForm={this.updateDataBaseForm.bind(this)}
            updateGrafanaNodesValidateState={this.updateGrafanaNodesValidateState.bind(
              this,
            )}
            updatePrometheusNodesValidateState={this.updatePrometheusNodesValidateState.bind(
              this,
            )}
            updateMonitorNodesValidateState={this.updateMonitorNodesValidateState.bind(
              this,
            )}
          />
        );
      case OptionSteps.ConnectInfo:
        return (
          <ConnectInfo
            configData={configData}
            dataBaseStorageType={dataBaseStorageType}
            connectInfoValidateState={connectInfoValidateState}
            onAddResource={this.onAddResource.bind(this)}
            onDeleteResource={this.onDeleteResource.bind(this)}
            onUpdateConnectInfo={this.onUpdateConnectInfo.bind(this)}
            updateConnectInfoForm={this.updateConnectInfoForm.bind(this)}
            updateConnectInfoValidateState={this.updateConnectInfoValidateState.bind(
              this,
            )}
          />
        );
    }
  }
}
