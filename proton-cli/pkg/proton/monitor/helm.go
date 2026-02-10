package monitor

import (
	"strconv"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

// HelmValuesFor return helm values for given configuration and registry
func HelmValuesFor(spec *configuration.ProtonMonitor, registry string) helm3.M {
	v := helm3.M{
		"namespace":    ReleaseNamespace,
		"replicaCount": HelmValuesReplicaCountFor(len(spec.Hosts), 1),
		"image":        helm3.M{"registry": registry},
		"config":       helm3.M{},
		"storage":      HelmValuesStorageFor(spec.Hosts, spec.DataPath, ""),
	}

	// 添加配置
	if spec.Config != nil {
		config := helm3.M{}

		// 添加 fluentbit 配置
		if spec.Config.Fluentbit != nil {
			fluentbit := helm3.M{}

			if spec.Config.Fluentbit.Port != 0 {
				fluentbit["port"] = spec.Config.Fluentbit.Port
			}

			if len(spec.Config.Fluentbit.Namespaces) > 0 {
				fluentbit["namespaces"] = spec.Config.Fluentbit.Namespaces
			}
			// 处理远程日志服务器配置
			if len(spec.Config.Fluentbit.RemoteLogServers) > 0 {
				remoteLogServers := []helm3.M{}

				for _, server := range spec.Config.Fluentbit.RemoteLogServers {
					serverConfig := helm3.M{}

					if server.Host != "" {
						serverConfig["host"] = server.Host
					}

					if server.Port != 0 {
						serverConfig["port"] = server.Port
					}

					if server.URI != "" {
						serverConfig["uri"] = server.URI
					}

					remoteLogServers = append(remoteLogServers, serverConfig)
				}

				fluentbit["remoteLogServers"] = remoteLogServers
			}
			config["fluentbit"] = fluentbit
		}

		// 添加 vmagent 配置
		if spec.Config.Vmagent != nil {
			vmagent := helm3.M{
				"k8sEtcdCerts": K8sEtcdCertsSecretName,
			}

			if spec.Config.Vmagent.ScrapeInterval != "" {
				vmagent["scrape_interval"] = spec.Config.Vmagent.ScrapeInterval
			}

			if spec.Config.Vmagent.ScrapeTimeout != "" {
				vmagent["scrape_timeout"] = spec.Config.Vmagent.ScrapeTimeout
			}

			if spec.Config.Vmagent.Port != 0 {
				vmagent["port"] = spec.Config.Vmagent.Port
			}

			// 添加 RemoteWrite 配置
			if spec.Config.Vmagent.RemoteWrite != nil {
				remoteWrite := helm3.M{}

				if spec.Config.Vmagent.RemoteWrite.Host != "" {
					remoteWrite["host"] = spec.Config.Vmagent.RemoteWrite.Host
				}

				if spec.Config.Vmagent.RemoteWrite.Port != 0 {
					remoteWrite["port"] = spec.Config.Vmagent.RemoteWrite.Port
				}

				if spec.Config.Vmagent.RemoteWrite.Path != "" {
					remoteWrite["path"] = spec.Config.Vmagent.RemoteWrite.Path
				}

				// 添加 extraServers 配置
				if len(spec.Config.Vmagent.RemoteWrite.ExtraServers) > 0 {
					var extraServers []helm3.M
					for _, server := range spec.Config.Vmagent.RemoteWrite.ExtraServers {
						extraServers = append(extraServers, helm3.M{
							"url": server.URL,
						})
					}
					remoteWrite["extraServers"] = extraServers
				}

				vmagent["remoteWrite"] = remoteWrite
			}

			config["vmagent"] = vmagent
		}

		// 添加 vmetrics 配置
		if spec.Config.Vmetrics != nil {
			vmetrics := helm3.M{}

			if spec.Config.Vmetrics.Retention != "" {
				vmetrics["retention"] = spec.Config.Vmetrics.Retention
			}

			if spec.Config.Vmetrics.Port != 0 {
				vmetrics["port"] = spec.Config.Vmetrics.Port
			}

			config["vmetrics"] = vmetrics
		}

		// 添加 vlogs 配置
		if spec.Config.Vlogs != nil {
			vlogs := helm3.M{}

			if spec.Config.Vlogs.Retention != "" {
				vlogs["retention"] = spec.Config.Vlogs.Retention
			}

			if spec.Config.Vlogs.Port != 0 {
				vlogs["port"] = spec.Config.Vlogs.Port
			}

			config["vlogs"] = vlogs
		}

		// 添加 grafana 配置
		if spec.Config.Grafana != nil {
			grafana := helm3.M{}

			if spec.Config.Grafana.Port != 0 {
				grafana["port"] = spec.Config.Grafana.Port
			}

			if spec.Config.Grafana.NodePort != 0 {
				grafana["nodePort"] = spec.Config.Grafana.NodePort
			}

			// 添加 SMTP 配置
			if spec.Config.Grafana.SMTP != nil && spec.Config.Grafana.SMTP.Enabled {
				smtp := helm3.M{
					"enabled": true,
				}

				if spec.Config.Grafana.SMTP.Host != "" {
					smtp["host"] = spec.Config.Grafana.SMTP.Host
				}

				if spec.Config.Grafana.SMTP.User != "" {
					smtp["user"] = spec.Config.Grafana.SMTP.User
				}

				if spec.Config.Grafana.SMTP.Password != "" {
					smtp["password"] = spec.Config.Grafana.SMTP.Password
				}

				smtp["skip_verify"] = spec.Config.Grafana.SMTP.SkipVerify

				if spec.Config.Grafana.SMTP.From != "" {
					smtp["from"] = spec.Config.Grafana.SMTP.From
				}

				if spec.Config.Grafana.SMTP.FromName != "" {
					smtp["from_name"] = spec.Config.Grafana.SMTP.FromName
				}

				if spec.Config.Grafana.SMTP.StartTLSPolicy != "" {
					smtp["startTLS_policy"] = spec.Config.Grafana.SMTP.StartTLSPolicy
				}

				smtp["enable_tracing"] = spec.Config.Grafana.SMTP.EnableTracing

				grafana["smtp"] = smtp
			}

			config["grafana"] = grafana
		}

		v["config"] = config
	}

	// 添加资源配置
	if spec.Resources != nil {
		resources := helm3.M{}

		// 添加 fluentbit 资源配置
		if spec.Resources.Fluentbit != nil {
			resources["fluentbit"] = spec.Resources.Fluentbit.DeepCopy()
		}

		// 添加 dcgmExporter 资源配置
		if spec.Resources.DcgmExporter != nil {
			resources["dcgmExporter"] = spec.Resources.DcgmExporter.DeepCopy()
		}

		// 添加 nodeExporter 资源配置
		if spec.Resources.NodeExporter != nil {
			resources["nodeExporter"] = spec.Resources.NodeExporter.DeepCopy()
		}

		// 添加 grafana 资源配置
		if spec.Resources.Grafana != nil {
			resources["grafana"] = spec.Resources.Grafana.DeepCopy()
		}

		// 添加 vmetrics 资源配置
		if spec.Resources.Vmetrics != nil {
			resources["vmetrics"] = spec.Resources.Vmetrics.DeepCopy()
		}

		// 添加 vlogs 资源配置
		if spec.Resources.Vlogs != nil {
			resources["vlogs"] = spec.Resources.Vlogs.DeepCopy()
		}

		// 添加 vmagent 资源配置
		if spec.Resources.Vmagent != nil {
			resources["vmagent"] = spec.Resources.Vmagent.DeepCopy()
		}

		v["resources"] = resources
	}

	return v
}

// HelmValuesStorageFor return storage configuration for .Values.storage
func HelmValuesStorageFor(hosts []string, dataPath string, storageClassName string) helm3.M {
	storage := helm3.M{}

	if storageClassName != "" {
		storage["storageClassName"] = storageClassName
	}

	if dataPath != "" {
		rel := make(helm3.M)
		for i, h := range hosts {
			rel[strconv.Itoa(i)] = helm3.M{
				"host": h,
				"path": dataPath,
			}
		}
		storage["local"] = rel
	}

	return storage
}

func HelmValuesReplicaCountFor(count, defaultReplicaCount int) int {
	if count == 0 {
		return defaultReplicaCount
	}
	return count
}
