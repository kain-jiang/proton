package validation

import (
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func ValidatePackageStore(spec *configuration.PackageStore, nodeNameSet sets.Set[string], fldPath *field.Path) (allErrs field.ErrorList) {
	// Record the name of the host that has appeared
	var seen = sets.New[string]()
	for i, h := range spec.Hosts {
		if !nodeNameSet.Has(h) {
			allErrs = append(allErrs, field.NotFound(fldPath.Child("hosts").Index(i), h))
			continue
		}
		if seen.Has(h) {
			allErrs = append(allErrs, field.Duplicate(fldPath.Child("hosts").Index(i), h))
		} else {
			seen.Insert(h)
		}
	}

	if len(spec.Hosts) != 0 && *spec.Replicas != len(spec.Hosts) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("replicas"), spec.Replicas, ".replicas should be equal to .hosts length if set"))
	}

	if len(spec.Hosts) != 0 {
		allErrs = append(allErrs, ValidatePackageBaredStorage(&spec.Storage, fldPath.Child("storage"))...)
	} else {
		allErrs = append(allErrs, ValidatePackageHostedStorage(&spec.Storage, fldPath.Child("storage"))...)
	}

	return
}

func ValidatePackageBaredStorage(storage *configuration.PackageStoreStorage, fldPath *field.Path) (allErrs field.ErrorList) {
	if storage.StorageClassName != "" {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("storageClassName"), storage.StorageClassName, "only for hosted environments"))
	}
	if !filepath.IsAbs(storage.Path) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("path"), storage.Path, "should be absolute path"))
	}
	for _, part := range strings.Split(storage.Path, "/") {
		if part == ".." {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("path"), storage.Path, "must not contain '..'"))
			// even for `../../..`, one error is sufficient to make the point
			break
		}
	}
	return
}

func ValidatePackageHostedStorage(storage *configuration.PackageStoreStorage, fldPath *field.Path) (allErrs field.ErrorList) {
	if storage.StorageClassName == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("storageClassName"), ""))
	}
	if storage.Path != "" {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("path"), storage.Path, "only for bared environments"))
	}
	return
}
