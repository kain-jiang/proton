package component

import "taskrunner/trait"

// HoleComponent hole resource component defined
type HoleComponent struct {
	trait.ComponentMeta `json:",inline"`
}

// Validate validate the obj
// TODO
func (c *HoleComponent) Validate(config map[string]interface{}, attribute map[string]interface{}) error {
	return nil
}

// ProtonResourceComponent do nothing, read config from proton conf
type ProtonResourceComponent = HoleComponent
