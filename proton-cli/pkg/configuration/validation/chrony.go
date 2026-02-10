package validation

import (
	"golang.org/x/exp/slices"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func ValidateChrony(c *configuration.Chrony, cs *configuration.Cs, fldPath *field.Path) (allErrs field.ErrorList) {
	// 如果c.Mode存在的话，只支持设计文档中描述的三种模式
	supportedChronyMode := []string{
		configuration.ChronyModeUserManaged,
		configuration.ChronyModeLocalMaster,
		configuration.ChronyModeExternalNTP,
	}
	switch c.Mode {
	// 用户管理模式下，server项目不能有值
	case configuration.ChronyModeUserManaged:
		if len(c.Server) > 0 {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("server"), "Proton-CLI will not manage NTP servers in usermanaged mode, please configure NTP manually"))
		}
	// 本地master节点模式下，server项目只能有一个服务器，且该服务器必须为一个master节点的节点名，该模式下server项目理论上用户不得更改
	case configuration.ChronyModeLocalMaster:
		if len(c.Server) != 1 {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("server"), "Only one server is allowed in localmaster mode"))
		}
		if !slices.Contains(cs.Master, c.Server[0]) {
			allErrs = append(allErrs, field.NotSupported(fldPath.Child("server"), c.Server[0], cs.Master))
		}
	// 外部时间服务器模式下，server项目不能为空，理论可以填写多个但是不对用户公开这一信息
	case configuration.ChronyModeExternalNTP:
		if len(c.Server) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("server"), "A NTP server address is required in chrony externalntp mode"))
		}
	default:
		allErrs = append(allErrs, field.NotSupported(fldPath.Child("mode"), c.Mode, supportedChronyMode))
	}
	return
}
