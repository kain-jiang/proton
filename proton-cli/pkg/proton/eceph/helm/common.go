package helm

import "encoding/json"

type Values struct {
	Namespace string `json:"namespace,omitempty"`

	Image ValuesImage `json:"image,omitempty"`

	Service *ValuesService `json:"service,omitempty"`

	DepServices *ValuesDepServices `json:"depServices,omitempty"`
}

type Values4ECeph struct {
	Namespace string `json:"namespace,omitempty"`

	Image ValuesImage `json:"image,omitempty"`

	Service map[string]string `json:"service,omitempty"`

	DepServices *ValuesDepServices `json:"depServices,omitempty"`
}

func (v *Values) ToMap() map[string]any {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		panic(err)
	}
	return m
}

func (v *Values4ECeph) ToMap() map[string]any {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		panic(err)
	}
	return m
}

type ValuesImage struct {
	Registry string `json:"registry,omitempty"`
}

type ValuesService struct {
	HTTPPort     int    `json:"httpPort,omitempty"`
	IngressClass string `json:"ingressClass,omitempty"`
}

type ValuesDepServices struct {
	RDS *ValuesRDS `json:"rds,omitempty"`

	RGW *ValuesRGW `json:"rgw,omitempty"`

	ProtonECephConfigManager map[string]string `json:"proton-eceph-config-manager,omitempty"`

	ProtonECephTenantManager map[string]string `json:"proton-eceph-tenant-manager,omitempty"`
}

type ValuesRDS struct {
	Type string `json:"type,omitempty"`

	Host string `json:"host,omitempty"`
	Port int    `json:"port,omitempty"`

	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}

type ValuesRGW struct {
	Host string `json:"host,omitempty"`

	Port int `json:"port,omitempty"`

	Protocol string `json:"protocol,omitempty"`
}
