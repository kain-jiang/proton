package store

import (
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/rds/mgmt/v1alpha1"
)

const (
	// Default replicas for hosted environments.
	DefaultReplicas = 3
	// 10Gi
	DefaultStorageCapacity = 10 << 30

	DefaultStoragePath = "/sysvol/package-store"
)

const (
	DatabaseName                         = "deploy"
	DatabaseCharset   v1alpha1.Charset   = v1alpha1.CharsetUTF8MB4
	DatabaseCollation v1alpha1.Collation = v1alpha1.CollationUTF8MB4GeneralCI

	RDSUserPrivilege v1alpha1.PrivilegeType = v1alpha1.PrivilegeReadWrite
)
