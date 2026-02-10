package validation

import (
	api_machinery_validation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

var ValidateNodeName = api_machinery_validation.NameIsDNS1035Label

func ValidateNodes(nodes []configuration.Node, fldPath *field.Path) (allErrs field.ErrorList) {
	allNames, allIPv4s, allIPv6s := sets.New[string](), sets.New[string](), sets.New[string]()
	for i, n := range nodes {
		path := fldPath.Index(i)

		for _, msg := range ValidateNodeName(n.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("name"), n.Name, msg))
		}

		if allNames.Has(n.Name) {
			allErrs = append(allErrs, field.Duplicate(path.Child("name"), n.Name))
		} else {
			allNames.Insert(n.Name)
		}

		if n.IP4 == "" && n.IP6 == "" {
			allErrs = append(allErrs, field.Invalid(path, "", "either .IP4 or .IP6 is required"))
		}

		if n.IP4 != "" {
			if !IsIPv4String(n.IP4) {
				allErrs = append(allErrs, field.Invalid(path.Child("ip4"), n.IP4, "must be an IPv4 address"))
			}
			if allIPv4s.Has(n.IP4) {
				allErrs = append(allErrs, field.Duplicate(path.Child("ip4"), n.IP4))
			} else {
				allIPv4s.Insert(n.IP4)
			}
		}

		if n.IP6 != "" {
			if !IsIPv6String(n.IP6) {
				allErrs = append(allErrs, field.Invalid(path.Child("ip6"), n.IP6, "must be an IPv6 address"))
			}
			if allIPv6s.Has(n.IP6) {
				allErrs = append(allErrs, field.Duplicate(path.Child("ip6"), n.IP6))
			} else {
				allIPv6s.Insert(n.IP6)
			}
		}
	}
	return
}

func NewNodeNameSet(nodes []configuration.Node) sets.Set[string] {
	var set = sets.New[string]()
	for _, n := range nodes {
		set.Insert(n.Name)
	}
	return set
}
