import { request } from "../../../tools/request";
import {
    ServiceItem,
    IGetServiceParams,
    ServiceJSONSchemaItem,
    ApplicationItem,
    IGetApplicationParams,
    IConfigJobParams,
    IGetJobParams,
    JobItem,
    ComponentItem,
    ComponentType,
    ICreateAndExecuteJobParams,
    IGetLogParams,
    JobLogItem,
    IGetConfigTemplateParams,
    ConfigTemplateItem,
    ISortServiceParams,
    IGetDependenciesListParams,
    DependenciesListItem,
} from "./declare";
import { paramsSerializer } from "../../../components/common/utils/request";
import { getLangConfig } from "../../../core/language";

class ServiceApplication {
    url = "/api/deploy-installer/v1/application";

    /**
     * 获取服务列表
     */
    get(params: IGetServiceParams): Promise<ServiceItem[]> {
        return request.get(
            `${this.url}/instance/work?${paramsSerializer({
                ...params,
                lang: getLangConfig(),
            })}`
        );
    }

    /**
     * 获取服务依赖列表
     */
    getDependenciesList(
        sid: number,
        params: IGetDependenciesListParams[]
    ): Promise<{ [key: string]: DependenciesListItem }> {
        return request.post(
            `${this.url}/autodependence?${paramsSerializer({
                sid,
                lang: getLangConfig(),
            })}`,
            params
        );
    }

    /**
     * 获取应用包列表
     */
    getApplication(params: IGetApplicationParams): Promise<ApplicationItem[]> {
        return request.get(
            `${this.url}?${paramsSerializer({
                ...params,
                lang: getLangConfig(),
            })}`
        );
    }

    /**
     * 获取应用包上传状态
     */
    getApplicationUploadStatus(name: string, version: string): Promise<null> {
        return request.get(
            `${this.url}/name/${name}/${version}?lang=${getLangConfig()}`
        );
    }

    /**
     * 获取配置模板列表
     */
    getConfigTemplate(
        params: IGetConfigTemplateParams
    ): Promise<{ data: ConfigTemplateItem[]; totalNum: number }> {
        return request.get(`${this.url}/config?${paramsSerializer(params)}`);
    }

    /**
     * 上传一个配置模板
     */
    postConfigTemplate(params: ConfigTemplateItem): Promise<number> {
        return request.post(`${this.url}/config`, params);
    }

    /**
     * 删除一个配置模板
     */
    deleteConfigTemplate(tid: number): Promise<null> {
        return request.delete(`${this.url}/config/${tid}`);
    }

    // 初步排序批量服务
    sortService(
        sid: number,
        params: ISortServiceParams[]
    ): Promise<{ sorted: ISortServiceParams[]; outer: ISortServiceParams[] }> {
        return request.put(
            `${this.url}/dependencesort?sid=${sid}&lang=${getLangConfig()}`,
            params
        );
    }
}

class ServiceJob {
    url = "/api/deploy-installer/v1/job";

    /**
     * 获取任务列表
     */
    get(params: IGetJobParams): Promise<JobItem[]> {
        return request.get(
            `${this.url}?${paramsSerializer({
                ...params,
                lang: getLangConfig(),
            })}`
        );
    }

    /**
     * 获取服务详细信息和配置
     */
    getJSONSchema(jid: number): Promise<ServiceJSONSchemaItem> {
        return request.get(
            `${this.url}/jsonschema/${jid}?lang=${getLangConfig()}`
        );
    }

    /**
     * 通过aid获取应用详细信息和配置
     */
    getJSONSchemaSnapshot(
        aid: number,
        params: { tid?: number[]; sid: number }
    ): Promise<ServiceJSONSchemaItem> {
        if (params?.tid?.length) {
            return request.get(
                `${this.url}/jsonschema/snapshot/${aid}?${paramsSerializer({
                    ...params,
                    lang: getLangConfig(),
                })}`
            );
        } else {
            return request.get(
                `${this.url}/jsonschema/snapshot/${aid}?${paramsSerializer({
                    ...params,
                    lang: getLangConfig(),
                    tid: undefined,
                })}`
            );
        }
    }

    /**
     * 通过name和version获取应用详细信息和配置
     */
    getJSONSchemaSnapshotByName(
        info: { name: string; version?: string },
        params?: { tid?: number[]; sid?: number }
    ): Promise<ServiceJSONSchemaItem> {
        if (params?.tid?.length) {
            return request.get(
                `${this.url}/jsonschema/snapshot/name/${info.name}/${
                    info.version ? info.version : ""
                }?${paramsSerializer({
                    ...params,
                    lang: getLangConfig(),
                })}`
            );
        } else if (params) {
            return request.get(
                `${this.url}/jsonschema/snapshot/name/${info.name}/${
                    info.version ? info.version : ""
                }?${paramsSerializer({
                    ...params,
                    lang: getLangConfig(),
                    tid: undefined,
                })}`
            );
        } else {
            return request.get(
                `${this.url}/jsonschema/snapshot/name/${info.name}/${
                    info.version ? info.version : ""
                }?lang=${getLangConfig()}`
            );
        }
    }

    /**
     * 以指定配置项创建任务并执行
     */
    createAndExecuteJob(params: ICreateAndExecuteJobParams): Promise<null> {
        return request.post(`${this.url}/jsonschema/snapshot`, params);
    }

    /**
     * 创建一个安装或更新任务
     */
    createJob(params: { aid: number }): Promise<number> {
        return request.post(`${this.url}`, params);
    }

    /**
     * 配置一个安装或更新任务
     */
    configJob(jid: number, data: IConfigJobParams): Promise<null> {
        return request.put(`${this.url}/${jid}`, data);
    }

    /**
     * 执行任务
     */
    executeJob(jid: number): Promise<null> {
        return request.post(`${this.url}/executor/${jid}`, {});
    }

    /**
     * 暂停任务
     */
    pauseJob(jid: number): Promise<null> {
        return request.patch(`${this.url}/executor/${jid}`, {});
    }

    /**
     * 获取任务错误日志
     */
    getLog(params: IGetLogParams): Promise<JobLogItem[]> {
        return request.get(
            `${this.url}/log?${paramsSerializer({
                ...params,
                lang: getLangConfig(),
            })}`
        );
    }

    /**
     * 卸载服务
     */
    uninstallService(params: {
        name: string;
        sid?: number;
        force: boolean;
    }): Promise<number> {
        return request.delete(`${this.url}?${paramsSerializer(params)}`);
    }
}

class ServiceComponent {
    url = "/api/deploy-installer/v1/component";

    /**
     * 获取微服务信息和配置项
     */
    getComponentInfo(cid: number): Promise<ComponentItem> {
        return request.get(`${this.url}/instance/${cid}`);
    }

    /**
     * 获取微服务的依赖服务列表
     */
    getComponentDependence(cid: number): Promise<ComponentType[]> {
        return request.get(`${this.url}/instance/${cid}/dependence`);
    }
}

export const serviceApplication = new ServiceApplication();
export const serviceJob = new ServiceJob();
export const serviceComponent = new ServiceComponent();
