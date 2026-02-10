package nebula

import (
	"encoding/base64"
	"errors"
	"fmt"

	"component-manage/internal/global"
	"component-manage/internal/logic/base"
	"component-manage/pkg/models/types/components"

	"component-manage/pkg/helm3"
	"component-manage/pkg/models/types"
)

// tools

//

func nebulaManifest(name string, param *types.NebulaComponentParams, plgInfo *types.NebulaPluginConfig) helm3.M {
	type m = map[string]any
	type l = []any

	allPathes := getDataPathes(param.DataPath)

	var (
		replicas = len(param.Hosts)
		logVC    = m{
			"resources": m{
				"requests": m{
					"storage": "1Gi",
				},
			},
		}
		serviceCfg = m{
			"enableDualStack": global.Config.Config.EnableDualStack,
		}
	)

	var (
		graphdVolume = func() l {
			rel := make(l, 0, replicas)
			for _, host := range param.Hosts {
				rel = append(rel, m{
					"host":    host,
					"logpath": allPathes.GraphDLog,
				})
			}
			return rel
		}()
		graphdConfig = m{
			"enable_authorize": "true",
		}

		metadDataVC = m{
			"resources": m{
				"requests": m{
					"storage": "5Gi",
				},
			},
		}
		metadVolume = func() l {
			rel := make(l, 0, replicas)
			for _, host := range param.Hosts {
				rel = append(rel, m{
					"host":     host,
					"logpath":  allPathes.MetaDLog,
					"datapath": allPathes.MetaDData,
				})
			}
			return rel
		}()

		storagedDataVCs = l{
			m{
				"resources": m{
					"requests": m{
						"storage": "10Gi",
					},
				},
			},
		}
		storagedVolume = func() l {
			rel := make(l, 0, replicas)
			for _, host := range param.Hosts {
				rel = append(rel, m{
					"host":     host,
					"logpath":  allPathes.MetaDLog,
					"datapath": l{allPathes.StorageDData0},
				})
			}
			return rel
		}()
	)

	values := m{
		"apiVersion": "apps.nebula-graph.io/v1alpha1",
		"kind":       "NebulaCluster",
		"metadata": m{
			"name":      name,
			"namespace": param.Namespace,
		},
		"spec": m{
			"enablePVReclaim": true,
			"secretName":      param.AdminSecretName,
			"imagePullPolicy": "IfNotPresent",
			"reference": m{
				"name":    "statefulsets.apps",
				"version": "v1",
			},

			"exporter": m{
				"image":       base.ImageName(plgInfo.Images.Exporter),
				"maxRequests": 20,
				"replicas":    1,
				"service":     serviceCfg,
				"version":     base.ImageTag(plgInfo.Images.Exporter),
			},

			"graphd": m{
				"config":         base.MergeHelmValues(graphdConfig, defaultMapTo(param.Graphd.Config, map[string]any{})),
				"image":          base.ImageName(plgInfo.Images.GraphD),
				"logVolumeClaim": logVC,
				"replicas":       replicas,
				"service":        serviceCfg,
				"version":        base.ImageTag(plgInfo.Images.GraphD),
				"volume":         graphdVolume,
			},

			"metad": m{
				"config":          defaultMapTo(param.Metad.Config, map[string]any{}),
				"dataVolumeClaim": metadDataVC,
				"image":           base.ImageName(plgInfo.Images.MetaD),
				"logVolumeClaim":  logVC,
				"replicas":        replicas,
				"service":         serviceCfg,
				"version":         base.ImageTag(plgInfo.Images.MetaD),
				"volume":          metadVolume,
			},
			"storaged": m{
				"config":           defaultMapTo(param.Storaged.Config, map[string]any{}),
				"dataVolumeClaims": storagedDataVCs,
				"image":            base.ImageName(plgInfo.Images.StorageD),
				"logVolumeClaim":   logVC,
				"replicas":         replicas,
				"service":          serviceCfg,
				"version":          base.ImageTag(plgInfo.Images.StorageD),
				"volume":           storagedVolume,
			},
		},
	}

	if param.Graphd.Resource != nil {
		values["spec"].(m)["graphd"].(m)["resource"] = param.Graphd.Resource.ResourceMap()
	}
	if param.Metad.Resource != nil {
		values["spec"].(m)["metad"].(m)["resource"] = param.Graphd.Resource.ResourceMap()
	}
	if param.Storaged.Resource != nil {
		values["spec"].(m)["storaged"].(m)["resource"] = param.Graphd.Resource.ResourceMap()
	}
	return values
}

func prepareNebula(param *types.NebulaComponentParams) error {
	// 创建 账户
	err := global.K8sCli.SecretSet(param.AdminSecretName, param.Namespace, map[string][]byte{
		"username": []byte("root"),
		"password": []byte(base64.StdEncoding.EncodeToString([]byte(param.Password))),
	})
	if err != nil {
		return fmt.Errorf("create admin secret failed: %w", err)
	}

	pathes := getDataPathes(param.DataPath).Pathes()
	for _, p := range pathes {
		if err := base.PrepareStorage(param.Hosts, p); err != nil {
			return err
		}
	}
	return nil
}

func clearNebula(param *types.NebulaComponentParams) error {
	// 删除密码
	err := global.K8sCli.SecretDel(param.AdminSecretName, param.Namespace)
	if err != nil {
		return fmt.Errorf("delete admin secret failed: %w", err)
	}

	pathes := getDataPathes(param.DataPath).Pathes()
	for _, p := range pathes {
		if err := base.ClearStorage(param.Hosts, p); err != nil {
			return err
		}
	}
	return nil
}

func generateInfo(name string, param *types.NebulaComponentParams) *components.NebulaComponentInfo {
	return &types.NebulaComponentInfo{
		Host:     fmt.Sprintf("%s-graphd-svc.%s.%s", name, param.Namespace, global.Config.ServiceSuffix()),
		Port:     9669,
		Username: "root",
		Password: param.Password,
	}
}

func checkNebula(name string, param, oldParam *types.NebulaComponentParams) error {
	param.AdminSecretName = defaultTo(param.AdminSecretName, fmt.Sprintf("nebula-admin-account-%s", name))
	param.Namespace = defaultTo(param.Namespace, "resource")
	param.Password = defaultTo(param.Password, createRootPassword())

	if oldParam != nil {
		// 更新
		param.Namespace = oldParam.Namespace
		param.AdminSecretName = oldParam.AdminSecretName
		param.Password = oldParam.Password

		if param.DataPath != oldParam.DataPath {
			return errors.New("data_path can not be changed")
		}

		if len(param.Hosts) < len(oldParam.Hosts) {
			return errors.New("hosts can not be reduced")
		}

		for idx, oH := range toList(oldParam.Hosts) {
			h := param.Hosts[idx]
			if h != oH {
				return fmt.Errorf("hosts can not be change to %s from %s", h, oH)
			}
		}

	}

	return nil
}
