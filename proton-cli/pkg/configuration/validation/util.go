package validation

import (
	"net"
	"sort"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func IsIPv4String(addr string) bool {
	return net.ParseIP(addr).To4() != nil
}

func IsIPv6String(addr string) bool {
	ip := net.ParseIP(addr)
	return ip.To16() != nil && ip.To4() == nil
}

// ValidateVersion tests whether version is valid.
func ValidateVersion(version string, fldPath *field.Path) (allErrs field.ErrorList) {
	allErrs = append(allErrs, ValidateRequiredString(version, fldPath)...)
	return
}

// ValidateHosts 检查各模块的 hosts 是否已经被合法
func ValidateHosts(hosts []string, nodeNameSet sets.Set[string], fldPath *field.Path) (allErrs field.ErrorList) {
	set := sets.New[string](hosts...)

	if s := set.Difference(nodeNameSet); s.Len() != 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, sets.List(s), "undefined"))
	}
	if set.Len() < len(hosts) {
		allErrs = append(allErrs, field.Duplicate(fldPath, hosts))
	}
	return
}

// ValidateHosts 检查各模块的 hosts 是否已经被合法
func ValidateOnlyOneHost(hosts []string, nodeNameSet sets.Set[string], fldPath *field.Path) (allErrs field.ErrorList) {
	set := sets.New[string](hosts...)

	if set.Len() == 0 {
		allErrs = append(allErrs, field.Required(fldPath, "at least one host"))
	}
	if set.Len() > 1 {
		allErrs = append(allErrs, field.Required(fldPath, "only one host allowed"))
	}
	if s := set.Difference(nodeNameSet); s.Len() != 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, sets.List(s), "undefined"))
	}
	if set.Len() < len(hosts) {
		allErrs = append(allErrs, field.Duplicate(fldPath, hosts))
	}
	if !sort.IsSorted(sort.StringSlice(hosts)) {
		allErrs = append(allErrs, field.Invalid(fldPath, hosts, "unsorted"))
	}
	return
}

// ValidateDataPath tests whether data path is valid.
func ValidateDataPath(version string, fldPath *field.Path) (allErrs field.ErrorList) {
	allErrs = append(allErrs, ValidateRequiredString(version, fldPath)...)
	return
}

// ValidateRequiredString tests whether string is provided.
func ValidateRequiredString(s string, fldPath *field.Path) (allErrs field.ErrorList) {
	if s == "" {
		allErrs = append(allErrs, field.Required(fldPath, ""))
	}
	return
}

// ValidatePort tests whether port is provided and valid 0-65535.
func ValidatePort(p int, fldPath *field.Path) (allErrs field.ErrorList) {
	if p <= 0 || p > 65535 {
		allErrs = append(allErrs, field.Invalid(fldPath, p, "valid port: 0-65535"))
	}
	return
}

// ValidateResourceQuantityValueString tests whether string is a valid quantity value
func ValidateResourceQuantityValueString(s string, fldPath *field.Path) (allErrs field.ErrorList) {
	if _, err := resource.ParseQuantity(s); err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath, s, err.Error()))
	}
	return
}

// ValidateHostsForPersistentData tests whether the hosts meet the requirements
// for persistent data. Because the persistent data uses the Persistent Volume's
// nodeAffinity field, which is immutable, the hosts list can append or pop, not
// update.
func ValidateHostsForPersistentData(oldHosts, newHosts []string, fldPath *field.Path) (allErrs field.ErrorList) {
	for i := 0; i < len(oldHosts) && i < len(newHosts); i++ {
		if oldHosts[i] != newHosts[i] {
			allErrs = append(allErrs, field.Invalid(fldPath.Index(i), oldHosts[i], "field is immutable"))
		}
	}
	return
}
