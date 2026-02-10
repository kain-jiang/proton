package keepalived

import "github.com/hashicorp/go-version"

// proton-cli 创建的 helm release 的名称
const ReleaseName = "proton-rds-keepalived"

// anysharectl 部署的 keepalived 的 release 名字与 Proton CLI 不同，此处做兼
// 容处理一并删除。
//
// 兼容范围是 proton-cli 1.1。从 proton-cli 1.2 及以上版本升级不需要此兼容处
// 理。当不再支持从 proton-cli 1.1 升级后，此部分兼容处理可以被移除。
const ReleaseNameAnyShareCTLCreated = "proton-mariadb-keepalived"

// rds keepalived 的 chart 的名字
const ChartName = "proton-rds-keepalived"

// helm release 所用的 chart 的版本低于 MinimumVersion 则需要升级
const MinimumVersionString = "1.4.0"

// helm release 所用的 chart 的版本低于 MinimumVersion 则需要升级
var MinimumVersion = version.Must(version.NewSemver(MinimumVersionString))
