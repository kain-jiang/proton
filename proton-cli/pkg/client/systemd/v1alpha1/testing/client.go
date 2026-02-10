package testing

import (
	systemd "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/systemd/v1alpha1"
)

type Unit struct {
	Name string

	Active, Enabled bool
}

type Client struct {
	Units []Unit

	Err error
}

// Start implements v1alpha1.Interface.
func (c *Client) Start(name string) error {
	if c.Err != nil {
		return c.Err
	}
	for _, u := range c.Units {
		if u.Name == name {
			u.Active = true
			return nil
		}
	}
	c.Units = append(c.Units, Unit{Name: name, Active: true})
	return nil
}

// Enabled implements v1alpha1.Interface.
func (c *Client) Enabled(name string, now bool) error {
	if c.Err != nil {
		return c.Err
	}
	for _, u := range c.Units {
		if u.Name == name {
			u.Enabled = true
			return nil
		}
	}
	c.Units = append(c.Units, Unit{Name: name, Enabled: true})
	return nil
}

// IsActive implements v1alpha1.Interface.
func (c *Client) IsActive(name string) (bool, error) {
	if c.Err != nil {
		return false, c.Err
	}
	for _, u := range c.Units {
		if u.Name == name {
			return u.Active, nil
		}
	}
	return false, nil
}

// IsEnabled implements v1alpha1.Interface.
func (c *Client) IsEnabled(name string) (bool, error) {
	if c.Err != nil {
		return false, c.Err
	}
	for _, u := range c.Units {
		if u.Name == name {
			return u.Enabled, nil
		}
	}
	return false, nil
}

var _ systemd.Interface = &Client{}
