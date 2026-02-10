package validation

import (
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func ValidateOpenSearch(c *configuration.OpenSearch, nodeNameSet sets.Set[string], fldPath *field.Path) (allErrs field.ErrorList) {
	// 组件管理服务中做过的校验，这里就不再做了。这里只保留组件管理服务未做的校验
	if c.StorageClassName == "" {
		allErrs = append(allErrs, ValidateHosts(c.Hosts, nodeNameSet, fldPath.Child("hosts"))...)
	}
	return
}

// ValidateOpenSearchHosts 检查OpenSearch hosts 是否已经被合法 支持hosts不按顺序
func ValidateOpenSearchHosts(hosts []string, nodeNameSet sets.Set[string], fldPath *field.Path) (allErrs field.ErrorList) {
	set := sets.New[string](hosts...)

	if s := set.Difference(nodeNameSet); s.Len() != 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, sets.List(s), "undefined"))
	}
	if set.Len() < len(hosts) {
		allErrs = append(allErrs, field.Duplicate(fldPath, hosts))
	}

	return
}
