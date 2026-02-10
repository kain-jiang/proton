package testing

import (
	"context"
	"errors"
	"sort"

	slb "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/slb/v2"
)

type KeepalivedHA struct {
	Items map[string]slb.KeepalivedHA

	Err error
}

// Create implements v2.KeepalivedHAInterface.
func (c *KeepalivedHA) Create(ctx context.Context, name string, kha *slb.KeepalivedHA) error {
	if c.Err != nil {
		return c.Err
	}
	if _, ok := c.Items[name]; ok {
		return errors.New("already exists")
	}
	if c.Items == nil {
		c.Items = make(map[string]slb.KeepalivedHA)
	}
	c.Items[name] = *kha
	return nil
}

// Delete implements v2.KeepalivedHAInterface.
func (c *KeepalivedHA) Delete(ctx context.Context, name string) error {
	if c.Err != nil {
		return c.Err
	}
	if _, ok := c.Items[name]; !ok {
		return errors.New("not found")
	}
	delete(c.Items, name)
	return nil
}

// Get implements v2.KeepalivedHAInterface.
func (c *KeepalivedHA) Get(ctx context.Context, name string) (*slb.KeepalivedHA, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	kha, ok := c.Items[name]
	if !ok {
		return nil, errors.New("not found")
	}
	return &kha, nil
}

func (c *KeepalivedHA) GetRaw(ctx context.Context, name string) (map[string]interface{}, error) {
	panic("unimplemented")
}

// List implements v2.KeepalivedHAInterface.
func (c *KeepalivedHA) List(ctx context.Context) ([]string, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	var names []string
	for k := range c.Items {
		names = append(names, k)
	}
	sort.Strings(names)
	return names, nil
}

// Update implements v2.KeepalivedHAInterface.
func (c *KeepalivedHA) Update(ctx context.Context, name string, kha *slb.KeepalivedHA) error {
	if c.Err != nil {
		return c.Err
	}
	if _, ok := c.Items[name]; !ok {
		return errors.New("not found")
	}
	c.Items[name] = *kha
	return nil
}

var _ slb.KeepalivedHAInterface = &KeepalivedHA{}
