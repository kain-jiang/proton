package validation

import (
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

var supportedFirewallModes = sets.New(
	configuration.FirewallFirewalld,
	configuration.FirewallUserManaged,
)

func ValidateFirewall(c *configuration.Firewall, fldPath *field.Path) (allErrs field.ErrorList) {
	if !supportedFirewallModes.Has(c.Mode) {
		allErrs = append(allErrs, field.NotSupported(fldPath.Child("mode"), c.Mode, sets.List(supportedFirewallModes)))
	}
	return
}
