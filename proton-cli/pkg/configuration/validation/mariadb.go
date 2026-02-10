package validation

import (
	"github.com/go-test/deep"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func ValidateMariaDB(m *configuration.ProtonMariaDB, nodeNameSet sets.Set[string], fldPath *field.Path) (allErrs field.ErrorList) {
	if m.StorageClassName == "" {
		allErrs = append(allErrs, ValidateHosts(m.Hosts, nodeNameSet, fldPath.Child("hosts"))...)
		allErrs = append(allErrs, ValidateRequiredString(m.Data_path, fldPath.Child("data_path"))...)
	}
	if m.StorageClassName != "" && len(m.Hosts) > 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("hosts"), m.Hosts, ".storageClassName and .hosts cannot be set at the same time"))
	}
	if m.StorageClassName != "" && m.Data_path != "" {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("data_path"), m.Hosts, ".storageClassName and .data_path cannot be set at the same time"))
	}
	if m.Config != nil {
		allErrs = append(allErrs, ValidateMariaDBConfig(m.Config, fldPath.Child("config"))...)
	}
	return
}

func ValidateMariaDBConfig(c *configuration.ProtonMariaDBConfigs, fldPath *field.Path) (allErrs field.ErrorList) {
	if c.LowerCaseTableNames != nil && (*c.LowerCaseTableNames < 0 || *c.LowerCaseTableNames > 2) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("lower_case_table_names"), *c.LowerCaseTableNames, "must be 0, 1 or 2"))
	}
	allErrs = append(allErrs, ValidateResourceQuantityValueString(c.Resource_requests_memory, fldPath.Child("resource_requests_memory"))...)
	allErrs = append(allErrs, ValidateResourceQuantityValueString(c.Resource_limits_memory, fldPath.Child("resource_limits_memory"))...)
	return
}

func ValidateMariaDBUpdate(o, n *configuration.ProtonMariaDB, fldPath *field.Path) (allErrs field.ErrorList) {
	if len(o.Hosts) > len(n.Hosts) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("hosts"), n.Hosts, "mariadb doesn't support scaling down"))
	} else if deep.Equal(o.Hosts, n.Hosts[:len(o.Hosts)]) != nil {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("hosts"), n.Hosts, "old hosts must be in front of the slice"))
	}

	if o.Data_path != n.Data_path {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("data_path"), n.Data_path, "data path is immutable"))
	}
	if o.StorageClassName != n.StorageClassName {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("storageClassName"), n.StorageClassName, "storage class name is immutable"))
	}
	return
}
