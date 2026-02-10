package configuration

// ECeph deployment configuration
type ECeph struct {
	// If we should skip ECeph update this time
	SkipECephUpdate bool `json:"skip_eceph_update,omitempty"`
	// ECeph will be deployed on these hosts.
	Hosts []string `json:"hosts,omitempty"`
	// ECeph's high availability configuration.
	Keepalived *ECephKeepalived `json:"keepalived,omitempty"`
	// ECeph's server TLS
	TLS ECephTLS `json:"tls,omitempty"`
}

// ECeph's high-availability configuration.
type ECephKeepalived struct {
	// Internal virtual ip address as CIDR, example: 10.0.14.71/24
	Internal string `json:"internal,omitempty"`
	// External virtual ip address as CIDR, example: 192.168.14.71/24
	External string `json:"external,omitempty"`
}

// ECeph's server TLS
type ECephTLS struct {
	// Name of the secret which the eceph's tls in
	Secret string `json:"secret,omitempty"`
	// CertificateData contains PEM-encoded data from a cert file for TLS.
	CertificateData []byte `json:"certificate-data,omitempty"`
	// KeyData contains PEM-encoded data from a key file for TLS.
	KeyData []byte `json:"key-data,omitempty"`
}
