package validation

import (
	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func ValidateCR(c *configuration.Cr, fldPath *field.Path) (allErrs field.ErrorList) {
	if c.Local == nil && c.External == nil {
		allErrs = append(allErrs, field.Invalid(fldPath, "", "either .Cr.Local or .Cr.External is required"))
		return allErrs
	}
	if c.Local != nil && c.External != nil {
		allErrs = append(allErrs, field.Invalid(fldPath, "", ".Cr.Local and .Cr.External are mutually exclusive"))
		return allErrs
	}
	if c.External != nil {
		if err := c.External.ValidateExternalCR(); err != nil {
			allErrs = append(allErrs, field.InternalError(fldPath, err))
			return allErrs
		}
	}
	return
}

func ValidateCRUpdate(o, n *configuration.Cr, fldPath *field.Path) (allErrs field.ErrorList) {
	if (n.Local != nil) != (o.Local != nil) || (n.External != nil) != (o.External != nil) {
		allErrs = append(allErrs, field.Invalid(fldPath, n.Local, "cr provisioner may not be changed"))
	}
	return
}
