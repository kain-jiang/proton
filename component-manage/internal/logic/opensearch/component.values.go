package opensearch

import (
	"errors"
	"fmt"
	"strconv"

	"component-manage/internal/global"
	"component-manage/internal/logic/base"
	"component-manage/pkg/models/types/components"

	"component-manage/pkg/helm3"
	"component-manage/pkg/models/types"

	"github.com/go-test/deep"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	OpensearchMasterServicePort = 9200

	// 内置opensearch默认用户名和密码
	OpensearchDefaultUsername = "admin"
	OpensearchDefaultPassword = "fake_pass"
	// Opensearch 默认连接协议
	OpensearchDefaultProtocol = "http"
	// Opensearch 内置版本
	OpensearchDefaultVersion = "7.10.0"
)

var SupportedOpenSearchModes = sets.NewString(
	"master",
	"hot",
	"warm",
)

func OpensearchValues(name string, p *types.OpensearchComponentParams) helm3.M {
	defaultSettings := map[string]interface{}{
		"bootstrap.memory_lock":    false,
		"action.auto_create_index": "-company*,-ar-*,-dip-*,-adp-*,+*",
		"node.search.cache.size":   "5gb",
	}

	v := make(map[string]interface{})

	// 镜像仓库
	v["image"] = map[string]interface{}{
		"registry": global.Config.Config.Registry,
	}

	v["namespace"] = p.Namespace

	replicas := len(p.Hosts)
	if replicas == 0 {
		replicas = p.ReplicaCount
	}
	v["replicaCount"] = replicas

	v["env"] = map[string]string{
		"language": "en_US.UTF-8",
		"timezone": "Asia/Shanghai",
	}

	v["service"] = map[string]interface{}{
		"enableDualStack": global.Config.Config.EnableDualStack,
	}

	v["config"] = map[string]interface{}{
		"hanlpRemote": map[string]interface{}{
			"extDict":      p.Config.HanlpRemoteextDict,
			"extStopwords": p.Config.HanlpRemoteextStopwords,
		},
		"nodeGroup":   p.Mode,
		"jvmOptions":  p.Config.JvmOptions,
		"clusterName": name,
	}

	// 填充默认值
	allSettings := make(map[string]interface{})
	if p.Settings != nil {
		allSettings = p.Settings
	}
	for dK, dV := range defaultSettings {
		if _, ok := allSettings[dK]; !ok {
			allSettings[dK] = dV
		}
	}
	v["settings"] = allSettings

	// 存储
	storage := make(map[string]interface{})
	storage["local"] = make(map[string](map[string]string))
	for i, h := range p.Hosts {
		storage["local"].(map[string](map[string]string))[strconv.Itoa(i)] = map[string]string{
			"host": h,
			"path": p.Data_path,
		}
	}
	storage["storageClassName"] = p.StorageClassName
	if len(p.StorageCapacity) > 0 {
		storage["capacity"] = p.StorageCapacity
	}
	v["storage"] = storage

	// 资源配额
	if p.Resources == nil && p.ExporterResources == nil {
	} else {
		v["resources"] = map[string]interface{}{}
		if p.Resources != nil {
			v["resources"].(map[string]interface{})["opensearch"] = p.Resources
		}
		if p.ExporterResources != nil {
			v["resources"].(map[string]interface{})["exporter"] = p.ExporterResources
		}
	}

	v = base.MergeHelmValues(v, p.ExtraValues)

	return v
}

func prepareOpensearch(param *types.OpensearchComponentParams) error {
	// 存储类无需准备
	if param.StorageClassName != "" {
		return nil
	}

	return base.PrepareStorage(param.Hosts, param.Data_path)
}

func clearOpensearch(param *types.OpensearchComponentParams) error {
	// 存储类无需准备
	if param.StorageClassName != "" {
		return nil
	}
	return base.ClearStorage(param.Hosts, param.Data_path)
}

func generateInfo(name string, p *types.OpensearchComponentParams) *components.OpensearchComponentInfo {
	return &types.OpensearchComponentInfo{
		SourceType:   "internal",
		Hosts:        fmt.Sprintf("%s.%s.%s", nameForMaster(name), p.Namespace, global.Config.ServiceSuffix()),
		Port:         OpensearchMasterServicePort,
		Username:     OpensearchDefaultUsername,
		Password:     OpensearchDefaultPassword,
		Protocol:     OpensearchDefaultProtocol,
		Version:      OpensearchDefaultVersion,
		Distribution: "opensearch",
	}
}

func checkOpensearch(p, oldp *types.OpensearchComponentParams) error {
	p.Namespace = base.DefaultString(p.Namespace, "resource")
	if p.Settings == nil {
		p.Settings = make(map[string]interface{})
	}
	if _, ok := p.Settings["action.auto_create_index"]; !ok {
		p.Settings["action.auto_create_index"] = "-company*,-ar-*,-dip-*,-adp-*,+*"
	}
	if _, ok := p.Settings["bootstrap.memory_lock"]; !ok {
		p.Settings["bootstrap.memory_lock"] = false
	}
	if !SupportedOpenSearchModes.Has(string(p.Mode)) {
		return fmt.Errorf("opensearch does not support the mode:%v, opensearch supports:%v", p.Mode, SupportedOpenSearchModes.List())
	}
	if p.StorageClassName != "" && len(p.Hosts) > 0 {
		return errors.New(".storageClassName and .hosts cannot be set at the same time")
	}
	if p.StorageClassName != "" && p.Data_path != "" {
		return errors.New(".storageClassName and .data_path cannot be set at the same time")
	}

	// 由于组件管理服务没有（本地K8S集群）所有节点主机的信息，所以不进行“节点是否属于集群节点”类校验，此处只检查节点列表是否重复
	// 数据目录创建逻辑会检测数据目录路径是否有效，不在此处检验
	if len(p.Hosts) > 0 {
		set := sets.New[string](p.Hosts...)
		if set.Len() < len(p.Hosts) {
			return errors.New("opensearch host list contains duplicated entries")
		}
	}

	////////////////////////////////////////////////////////////////////

	if oldp != nil {
		if base.DefaultString(oldp.Namespace, "resource") != p.Namespace {
			return errors.New("namespace is immutable")
		}
		if p.Data_path != oldp.Data_path {
			return errors.New("opensearch data_path is immutable")
		}
		if p.StorageClassName != oldp.StorageClassName {
			return errors.New("opensearch storage class name is immutable")
		}
		if p.StorageCapacity != oldp.StorageCapacity {
			return errors.New("storageCapacity can not be changed")
		}
		// 暂不支持mode修改
		if string(p.Mode) != string(oldp.Mode) {
			return errors.New("opensearch mode is immutable")
		}
		// 仅支持扩容,不支持缩容
		replicaCountActual := func(p0 *components.OpensearchComponentParams) int {
			replicas := len(p0.Hosts)
			if replicas == 0 {
				return p0.ReplicaCount
			}
			return replicas
		}
		if replicaCountActual(p) < replicaCountActual(oldp) {
			return errors.New("Real replicaCount can not be reduced")
		} else if p.Hosts != nil && oldp.Hosts != nil {
			// 扩容时，新配置节点列表必须满足旧节点在最前
			for _, diff := range deep.Equal(p.Hosts[:len(oldp.Hosts)], oldp.Hosts) {
				return fmt.Errorf("previous hosts must be in front of new hosts when expanding opensearch deployment: %v, %v", p.Hosts, diff)
			}
		}
	}

	return nil
}

func nameForMaster(name string) string {
	return fmt.Sprintf("%s-master", name)
}
