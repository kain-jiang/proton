package validation

import (
	"net"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func Validate(spec *configuration.ECeph, nodes []configuration.Node, info *configuration.ResourceConnectInfo, fldPath *field.Path) (allErrs field.ErrorList) {
	if spec == nil {
		return
	}
	allErrs = append(allErrs, validateHosts(spec.Hosts, nodes, fldPath.Child("hosts"))...)
	allErrs = append(allErrs, validateResourceConnectInfo(info, fldPath)...)
	return
}

func ValidatePost(spec *configuration.ECeph, nodes []configuration.Node, info *configuration.ResourceConnectInfo, fldPath *field.Path) (allErrs field.ErrorList) {
	if spec == nil {
		return
	}
	allErrs = append(allErrs, validateKeepalived(spec.Keepalived, nodes, fldPath.Child("keepalived"))...)
	allErrs = append(allErrs, validateTLS(&spec.TLS, fldPath.Child("tls"))...)
	return
}

func validateHosts(hosts []string, nodes []configuration.Node, fldPath *field.Path) (allErrs field.ErrorList) {
	names := sets.NewString()
	for _, n := range nodes {
		names.Insert(n.Name)
	}
	seen := sets.NewString()
	for i, h := range hosts {
		if !names.Has(h) {
			allErrs = append(allErrs, field.NotFound(fldPath.Index(i), h))
			continue
		}
		if seen.Has(h) {
			allErrs = append(allErrs, field.Duplicate(fldPath.Index(i), h))
			continue
		}
		for _, n := range nodes {
			if n.Name == h && len(n.Internal_ip) == 0 {
				allErrs = append(allErrs, field.Required(field.NewPath("nodes").Child(n.Name).Child("internal_ip"), "Nodes to deploy ECeph must have an internal IP"))
			}
			if n.Name == h && (len(n.IP4) > 0 && len(n.IP6) > 0) {
				allErrs = append(allErrs, field.Forbidden(field.NewPath("nodes").Child(n.Name), "Nodes to deploy ECeph cannot have IPV4 and IPV6 address at the same time."))
			}
		}
		seen.Insert(h)
	}
	return
}

func validateKeepalived(k *configuration.ECephKeepalived, nodes []configuration.Node, fldPath *field.Path) (allErrs field.ErrorList) {
	if k == nil || (k.Internal == "" && k.External == "") {
		return
	}
	if len(k.Internal) > 0 {
		allErrs = append(allErrs, validateKeepalivedInternalVirtualAddress(k.Internal, nodes, fldPath.Child("internal"))...)
	}
	if len(k.External) > 0 {
		allErrs = append(allErrs, validateKeepalivedExternalVirtualAddress(k.External, nodes, fldPath.Child("external"))...)
	}
	return
}

func validateKeepalivedInternalVirtualAddress(cidr string, nodes []configuration.Node, fldPath *field.Path) (allErrs field.ErrorList) {
	if len(cidr) == 0 {
		return
	}
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath, cidr, "not a valid CIDR"))
		return
	}

	for _, n := range nodes {
		internal := net.ParseIP(n.Internal_ip)
		if ip.Equal(internal) {
			allErrs = append(allErrs, field.Duplicate(fldPath, ""))
			continue
		}
		if !ipNet.Contains(internal) {
			allErrs = append(allErrs, field.Invalid(fldPath, ip, "not in same network with "+internal.String()))
			continue
		}
	}
	return
}

func validateKeepalivedExternalVirtualAddress(cidr string, nodes []configuration.Node, fldPath *field.Path) (allErrs field.ErrorList) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath, cidr, "not a valid CIDR"))
		return
	}

	var ips []net.IP
	for _, n := range nodes {
		if ip.To4() != nil {
			ips = append(ips, net.ParseIP(n.IP4))
		} else {
			ips = append(ips, net.ParseIP(n.IP6))
		}
	}
	for _, nodeIP := range ips {
		if ip.Equal(nodeIP) {
			allErrs = append(allErrs, field.Duplicate(fldPath, ""))
			continue
		}
		if !ipNet.Contains(nodeIP) {
			allErrs = append(allErrs, field.Invalid(fldPath, ip, "not in same network with "+nodeIP.String()))
			continue
		}
	}
	return
}

// TODO: Add more validation.
//  1. If the certificate contains intermediates, it is a valid chain.
//  2. Format etc.
func validateTLS(tls *configuration.ECephTLS, fldPath *field.Path) (allErrs field.ErrorList) {

	if tls.Secret == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("tls"), ""))
	}
	if len(tls.CertificateData) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("certificate-data"), ""))
	}
	if len(tls.KeyData) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("key-data"), ""))
	}
	return
}

func validateResourceConnectInfo(info *configuration.ResourceConnectInfo, fldPath *field.Path) (allErrs field.ErrorList) {
	if info == nil || info.Rds == nil {
		allErrs = append(allErrs, field.Required(fldPath, "requires rds connection info"))
		return
	}

	return
}

func ValidateUpdate(o, n *configuration.ECeph, fldPath *field.Path) (allErrs field.ErrorList) {
	if o != nil && n == nil {
		allErrs = append(allErrs, field.Invalid(fldPath, n, "not support uninstall"))
		return
	}
	// no other limits
	return
}
