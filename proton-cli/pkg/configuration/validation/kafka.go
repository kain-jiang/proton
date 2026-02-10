package validation

import (
	"github.com/go-test/deep"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

// ValidateKafka tests whether required fields in th Kafka are set.
func ValidateKafka(k *configuration.Kafka, nodeNameSet sets.Set[string], fldPath *field.Path) (allErrs field.ErrorList) {
	if k.StorageClassName == "" {
		allErrs = append(allErrs, ValidateKafkaHosts(k.Hosts, nodeNameSet, fldPath.Child("hosts"))...)
		allErrs = append(allErrs, ValidateRequiredString(k.Data_path, fldPath.Child("data_path"))...)
	}
	if k.StorageClassName != "" && len(k.Hosts) > 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("hosts"), k.Hosts, ".storageClassName and .hosts cannot be set at the same time"))
	}
	if k.StorageClassName != "" && k.Data_path != "" {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("data_path"), k.Hosts, ".storageClassName and .data_path cannot be set at the same time"))
	}
	return
}

// ValidateKafkaHosts 检查Kafka hosts 是否已经被合法 支持hosts不按顺序
func ValidateKafkaHosts(hosts []string, nodeNameSet sets.Set[string], fldPath *field.Path) (allErrs field.ErrorList) {
	set := sets.New[string](hosts...)

	if s := set.Difference(nodeNameSet); s.Len() != 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, sets.List(s), "undefined"))
	}
	if set.Len() < len(hosts) {
		allErrs = append(allErrs, field.Duplicate(fldPath, hosts))
	}

	return
}
func ValidateKafkaUpdate(o, n *configuration.Kafka, fldPath *field.Path) (allErrs field.ErrorList) {
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
