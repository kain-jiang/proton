package trait

const (
	AppMetaFile = "appMeta.json"
)

const (
	AppBetav1Type = "app/v1betav1"
)

// const (
// 	ComponentHelmTaskType    = "helm/task"
// 	ComponentHelmServiceType = "helm/service"
// 	ComponentHoleype         = "hole"
// )

const (
	HelmDefinedPath   = "_componentMeta.json"
	HelmChartPath     = "Chart.yaml"
	HelmChartDir      = "helm_charts/"
	ConfigTemplateDir = "config_templates/"
)

const (
	// AppinitStatus app only init not start
	AppinitStatus = iota
	// AppConfirmedStatus the config has been confirmed by user
	AppConfirmedStatus
	// AppWaitingStatus waiting execute
	AppWaitingStatus
	// AppDoingStatus execute app
	AppDoingStatus
	// AppSucessStatus app sucess
	AppSucessStatus
	// AppFailStatus app fail for component task
	AppFailStatus
	// AppStopedStatus  app has been stop
	AppStopedStatus
	// AppStopingStatus app waiting for stop
	AppStopingStatus
	// AppFailMissStatus fail because of missing some component
	AppFailMissStatus

	// AppFailUninstallStatus fail when uninstall old component
	AppFailUninstallStatus

	// AppIgnoreStatus ignore the execute
	AppIgnoreStatus

	// AppDeleteingOldComponentStatus job is deleting the component in old version
	AppDeleteingOldComponentStatus

	// AppUpdatedComponentStatus job has updated the component to new version
	AppUpdatedComponentStatus

	// AppUpgradeParentComponentStatus job is upgrading the parent compoent in system
	AppUpgradeParentComponentStatus

	// AppUpgradeParentComponentFailStatus job failed when upgrad the parent compoent in system
	AppUpgradeParentComponentFailStatus

	// AppToCreateStatus job not create, it will be create if the task nomaly work
	AppToCreateStatus
)

// JobDoingStauts job is need executing. readonnly
var JobDoingStauts = []int{
	AppWaitingStatus,
	AppDoingStatus,
	AppStopingStatus,
	AppDeleteingOldComponentStatus,
	AppUpdatedComponentStatus,
	AppUpgradeParentComponentStatus,
}

const (
	// JobUpgradeOType 更新操作符
	JobUpgradeOType = 0
	// JobInstallOType 安装操作符
	JobInstallOType = 1
	// JobRollbackOType 回滚操作符
	JobRollbackOType = 2
	// JobDeleteOType 删除操作符
	JobDeleteOType = 3
)
