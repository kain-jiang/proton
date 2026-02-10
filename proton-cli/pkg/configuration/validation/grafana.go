package validation

import (
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/grafana"
)

// ValidateGrafana 检查 Grafana 的部署配置是否符合要求
func ValidateGrafana(spec *configuration.Grafana, nodeNameSet sets.Set[string], fldPath *field.Path) (allErrs field.ErrorList) {
	if len(spec.Hosts) > grafana.Replicas {
		allErrs = append(allErrs, field.TooMany(fldPath.Child("hosts"), len(spec.Hosts), grafana.Replicas))
	}
	for i, h := range spec.Hosts {
		if !nodeNameSet.Has(h) {
			allErrs = append(allErrs, field.NotFound(fldPath.Child("hosts").Index(i), h))
		}
	}

	if len(spec.Hosts) != 0 && spec.DataPath == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("data_path"), ""))
	}

	if spec.DataPath != "" && !filepath.IsAbs(spec.DataPath) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("data_path"), spec.DataPath, "should be absolute path"))
	}
	for _, part := range strings.Split(spec.DataPath, "/") {
		if part == ".." {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("data_path"), spec.DataPath, "must not contain '..'"))
			// even for `../../..`, one error is sufficient to make the point
			break
		}
	}

	if spec.StorageClassName != "" {
		if len(spec.Hosts) != 0 || spec.DataPath != "" {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("storageClassName"), spec.StorageClassName, "storageClassName is conflicted with host and data_path"))
		}
	} else if len(spec.Hosts) < grafana.Replicas {
		allErrs = append(allErrs, field.Required(fldPath.Child("hosts"), ""))
	}
	return
}

// ValidateGrafanaUpdate 检查 Grafana 能否从旧的配置更新为新的配置
func ValidateGrafanaUpdate(o, n *configuration.Grafana, fldPath *field.Path) (allErrs field.ErrorList) {
	for i, h := range n.Hosts {
		if i >= len(o.Hosts) || h != o.Hosts[i] {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("hosts").Index(i), "immutable"))
		}
	}
	if n.DataPath != o.DataPath {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("data_path"), "immutable"))
	}
	if n.StorageClassName != o.StorageClassName {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("storageClassName"), "immutable"))
	}
	return
}
