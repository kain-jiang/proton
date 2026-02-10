import { request } from "../../../tools/request";
import { paramsSerializer } from "../../../components/common/utils/request";
import {
    ApplicationItem,
    ICreateComposeJobParams,
    IGetApplicationParams,
    IGetJobParams,
    IGetSuiteParams,
    SuiteItem,
    SuiteManifestsItem,
} from "./declare";
import { getLangConfig } from "../../../core/language";

class ComposeJob {
    url = "/api/deploy-installer/v1/composejob";

    /**
     * 获取套件任务列表信息
     */
    getJobList(
        params: IGetJobParams,
        appFormat = "schema"
    ): Promise<SuiteItem[]> {
        return request.get(
            `${this.url}?${paramsSerializer({
                ...params,
                lang: getLangConfig(),
            })}`,
            {},
            {
                headers: {
                    "app-format": appFormat,
                },
            }
        );
    }

    /**
     * 获取套件任务信息
     */
    getJobInfo(jid: number, appFormat = "schema"): Promise<SuiteItem> {
        return request.get(
            `${this.url}/${jid}?lang=${getLangConfig()}`,
            {},
            {
                headers: {
                    "app-format": appFormat,
                },
            }
        );
    }

    /**
     * 创建套件部署任务
     */
    createJob(
        params: ICreateComposeJobParams,
        appFormat = "schema"
    ): Promise<number> {
        return request.post(this.url, params, {
            headers: {
                "app-format": appFormat,
            },
        });
    }

    /**
     * 启动套件部署任务
     */
    startJob(jid: number): Promise<null> {
        return request.post(`${this.url}/${jid}`, {});
    }

    /**
     * 暂停套件部署任务
     */
    pauseJob(jid: number): Promise<null> {
        return request.put(`${this.url}/${jid}`);
    }

    /**
     * 配置套件部署任务
     */
    patchJob(
        jid: number,
        params: SuiteItem,
        appFormat = "schema"
    ): Promise<null> {
        return request.patch(`${this.url}/${jid}`, params, {
            headers: {
                "app-format": appFormat,
            },
        });
    }
}

class SuiteManifests {
    url = "/api/deploy-installer/v1/manifests";

    /**
     * 获取已安装套件列表
     */
    get(params: IGetSuiteParams): Promise<SuiteItem[]> {
        return request.get(
            `${this.url}/work?${paramsSerializer({
                ...params,
                lang: getLangConfig(),
            })}`
        );
    }

    /**
     * 获取应用包列表
     */
    getApplication(
        params: IGetApplicationParams
    ): Promise<{ data: ApplicationItem[] }> {
        return request.get(
            `${this.url}?${paramsSerializer({
                ...params,
                lang: getLangConfig(),
            })}`
        );
    }

    // 获取套件配置清单
    getSuiteManifests(
        name: string,
        version: string
    ): Promise<SuiteManifestsItem> {
        return request.get(
            `${this.url}/${name}/${version}?lang=${getLangConfig()}`
        );
    }
}

export const composeJob = new ComposeJob();
export const suiteManifests = new SuiteManifests();
