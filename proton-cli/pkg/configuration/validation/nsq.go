package validation

import (
	"github.com/go-test/deep"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func ValidateMQNSQ(c *configuration.ProtonDataConf, nodeNameSet sets.Set[string], fldPath *field.Path) (allErrs field.ErrorList) {
	if c.StorageClassName == "" {
		allErrs = append(allErrs, ValidateMQNSQHosts(c.Hosts, nodeNameSet, fldPath.Child("hosts"))...)
		allErrs = append(allErrs, ValidateRequiredString(c.Data_path, fldPath.Child("data_path"))...)
	}
	if c.StorageClassName != "" && len(c.Hosts) > 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("hosts"), c.Hosts, ".storageClassName and .hosts cannot be set at the same time"))
	}
	if c.StorageClassName != "" && c.Data_path != "" {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("data_path"), c.Hosts, ".storageClassName and .data_path cannot be set at the same time"))
	}
	return
}

// ValidateMQNSQHosts 检查MQNSQ hosts 是否已经被合法 支持hosts不按顺序
func ValidateMQNSQHosts(hosts []string, nodeNameSet sets.Set[string], fldPath *field.Path) (allErrs field.ErrorList) {
	set := sets.New[string](hosts...)

	if set.Len() != 1 && set.Len() != 3 {
		allErrs = append(allErrs, field.Required(fldPath, "only support 1 or 3 host"))
	}
	if s := set.Difference(nodeNameSet); s.Len() != 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, sets.List(s), "undefined"))
	}
	if set.Len() < len(hosts) {
		allErrs = append(allErrs, field.Duplicate(fldPath, hosts))
	}

	return
}

// MQNSQ 暂不支持数据目录修改
func ValidateMQNSQUpdate(o, n *configuration.ProtonDataConf, fldPath *field.Path) (allErrs field.ErrorList) {
	if n.Data_path != o.Data_path {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("data_path"), n.Data_path, "data path is immutable"))
	}
	if o.StorageClassName != n.StorageClassName {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("storageClassName"), n.StorageClassName, "storage class name is immutable"))
	}

	// 仅支持扩容,不支持缩容
	if len(n.Hosts) < len(o.Hosts) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("hosts"), n.Hosts, "only support expand hosts"))
	} else {
		// 扩容时，新配置节点列表必须满足旧节点在最前
		for _, diff := range deep.Equal(n.Hosts[:len(o.Hosts)], o.Hosts) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("hosts"), n.Hosts, diff+" old-hosts must be in front"))
		}
	}

	return
}
