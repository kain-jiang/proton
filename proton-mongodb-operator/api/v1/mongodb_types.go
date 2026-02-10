/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MongoDBSpec defines the desired state of MongoDB
type MongoDBSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Selector *metav1.LabelSelector `json:"selector,omitempty"`
	Image           string                      `json:"image,omitempty"`
	ImagePullPolicy corev1.PullPolicy           `json:"imagePullPolicy,omitempty"`
	Replicas        int32                       `json:"replicas,omitempty"`
	Service         *Service                    `json:"service,omitempty"`
	Resources       corev1.ResourceRequirements `json:"resources,omitempty"`
	Storage         *Storage                    `json:"storage,omitempty"`
	Replset         *ReplsetSpec                `json:"replset,omitempty"`
	Mongodconf      *Mongodconf                 `json:"conf,omitempty"`
	Debug           string                      `json:"debug,omitempty"`
	// UpgradeOptions          UpgradeOptions                       `json:"upgradeOptions,omitempty"`
	// ClusterServiceDNSSuffix string `json:"clusterServiceDNSSuffix,omitempty"`
	// ClusterServiceDNSMode   DnsMode `json:"clusterServiceDNSMode,omitempty"`
	// Sharding                Sharding `json:"sharding,omitempty"`
	// MultiCluster            MultiCluster                         `json:"multiCluster,omitempty"`
	// PodAffinity        *PodAffinity        `json:"podaffinity,omitempty"`
	// NodeSelector       map[string]string   `json:"nodeSelector,omitempty"`
}

type Service struct {
	Type            corev1.ServiceType `json:"type,omitempty"`
	EnableDualStack bool               `json:"enableDualStack"`
	Port            int32              `json:"port,omitempty"`
}

type Mongodconf struct {
	WiredTigerCacheSizeGB int32   `json:"wiredTigerCacheSizeGB,omitempty"`
	TLS                   *TLS    `json:"tls,omitempty"`
	AuditLog              LogConf `json:"auditLog,omitempty"`
}

type ReplsetSpec struct {
	Name string `json:"name,omitempty"`
	// Size int32  `json:"size"`
}

type ReplsetStatus struct {
	Size int32 `json:"size"`
}

type Volume struct {
	Host string `json:"host"`
	Path string `json:"path"`
	// BackupPath string `json:"backupPath"`
}

type Storage struct {
	Capacity         string   `json:"capacity,omitempty"`
	StorageClassName string   `json:"storageClassName"`
	VolumeSpec       []Volume `json:"volume,omitempty"`
}

// TLS is the configuration used to set up TLS encryption
type TLS struct {
	Enabled       bool   `json:"enabled"`
	TLSSecretName string `json:"tlssecretname,omitempty"`

	// // CertificateKeySecret is a reference to a Secret containing a private key and certificate to use for TLS.
	// // The key and cert are expected to be PEM encoded and available at "tls.key" and "tls.crt".
	// // This is the same format used for the standard "kubernetes.io/tls" Secret type, but no specific type is required.
	// // Alternatively, an entry tls.pem, containing the concatenation of cert and key, can be provided.
	// // If all of tls.pem, tls.crt and tls.key are present, the tls.pem one needs to be equal to the concatenation of tls.crt and tls.key
	// // +optional
	// CertificateKeySecret LocalObjectReference `json:"certificateKeySecretRef"`

	// // CaCertificateSecret is a reference to a Secret containing the certificate for the CA which signed the server certificates
	// // The certificate is expected to be available under the key "ca.crt"
	// // +optional
	// CaCertificateSecret *LocalObjectReference `json:"caCertificateSecretRef,omitempty"`
}

type LogConf map[string]intstr.IntOrString

type PodAffinity struct {
	TopologyKey *string          `json:"antiAffinityTopologyKey,omitempty"`
	Advanced    *corev1.Affinity `json:"advanced,omitempty"`
}

// var affinityValidTopologyKeys = map[string]struct{}{
// 	AffinityOff:                                {},
// 	"kubernetes.io/hostname":                   {},
// 	"failure-domain.beta.kubernetes.io/zone":   {},
// 	"failure-domain.beta.kubernetes.io/region": {},
// }

// var defaultAffinityTopologyKey = "kubernetes.io/hostname"

const AffinityOff = "none"
