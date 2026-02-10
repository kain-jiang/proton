package utils

import (
	store "taskrunner/pkg/store/proton"
	"taskrunner/pkg/utils"
	"taskrunner/trait"
)

// ProtonCli proton client info
type ProtonCli struct {
	Namespace string `json:"namespace,omitempty"`
	ConfName  string `json:"confName,omitempty"`
	ConfKey   string `json:"confKey,omitempty"`
}

// GetProtonCli get proton client operator
func GetProtonCli(cfg *ProtonCli) (pcli *store.ProtonClient, err *trait.Error) {
	kcli, err := utils.NewKubeclient()
	if err != nil {
		return nil, err
	}

	pcli = &store.ProtonClient{
		Namespace: cfg.Namespace,
		ConfName:  cfg.ConfName,
		Confkey:   cfg.ConfKey,
		Kcli:      kcli,
	}
	return
}
