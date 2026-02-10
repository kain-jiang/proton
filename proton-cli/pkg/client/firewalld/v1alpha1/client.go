package v1alpha1

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	exec "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/exec/v1alpha1"
)

// Client implements Interface
type Client struct {
	executor exec.Executor
}

func New(e exec.Executor) *Client {
	return &Client{executor: e}
}

const commandFirewallCMD = "firewall-cmd"

const cmd = "firewall-cmd"

// IsRunning implements Interface. Alias of State().
func (c *Client) IsRunning() (bool, error) { return c.State() }

// AddPort implements Interface.
func (c *Client) AddPort(port string, permanent bool, zone string, timeout string) error {
	var args []string
	args = append(args, "--add-port", port)
	if permanent {
		args = append(args, "--permanent")
	}
	if zone != "" {
		args = append(args, fmt.Sprintf("--zone=%s", zone))
	}
	if timeout != "" {
		args = append(args, fmt.Sprintf("--timeout=%s", timeout))
	}

	return c.executor.Command(commandFirewallCMD, args...).Run()
}

// ListPorts implements Interface.
func (c *Client) ListPorts(permanent bool, zone string) ([]string, error) {
	var args []string
	args = append(args, "--list-ports")
	if permanent {
		args = append(args, "--permanent")
	}
	if zone != "" {
		args = append(args, fmt.Sprintf("--zone=%s", zone))
	}

	out, err := c.executor.Command(commandFirewallCMD, args...).Output()
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(string(out)), " "), nil
}

// 添加防火墙
func (c *Client) RemoveAndAddFirewallSource(ip string, oldIP string, permanent bool) error {
	// strip subnet mask
	ip = strings.Split(ip, "/")[0]
	oldIP = strings.Split(oldIP, "/")[0]
	if !permanent {
		if len(oldIP) > 0 {
			_ = c.executor.Command(commandFirewallCMD, "--remove-source", oldIP, "--zone=trusted").Run()
		}
		args := []string{"--add-source", ip, "--zone=trusted"}
		out, err := c.executor.Command(commandFirewallCMD, args...).Output()
		if err != nil {
			return fmt.Errorf("failed to add firewall source: exec %v failed, output %v, error %v", fmt.Sprintln(append([]string{commandFirewallCMD}, args...)), out, err)
		}
	} else {
		if len(oldIP) > 0 {
			_ = c.executor.Command(commandFirewallCMD, "--zone=trusted", "--permanent", "--remove-source", ip).Run()
		}

		args := []string{"--zone=trusted", "--permanent", "--add-source", ip}
		out, err := c.executor.Command(commandFirewallCMD, args...).Output()
		if err != nil {
			return fmt.Errorf("failed to add firewall source: exec %v failed, output %v, error %v", fmt.Sprintln(append([]string{commandFirewallCMD}, args...)), out, err)
		}
	}
	return nil
}

// Check whether the firewalld daemon is active.
func (c *Client) State() (bool, error) {
	var opts []string
	opts = appendOption(opts, optState)

	err := c.executor.Command(commandFirewallCMD, opts...).Run()
	if err == nil {
		return true, nil
	}

	ee := new(exec.ErrExitError)
	if !errors.As(err, &ee) {
		return false, err
	}

	if ee.ExitCode == ExitCodeNotRunning && strings.TrimSpace(string(ee.Stderr)) == "not running" {
		return false, nil
	}

	return false, err
}

// Reload firewall rules and keep state information. Current permanent
// configuration will become new runtime configuration, i.e. all runtime only
// changes done until reload are lost with reload if they have not been also in
// permanent configuration.
func (c *Client) Reload() error {
	var opts []string
	opts = appendOption(opts, optReload)

	return c.executor.Command(cmd, opts...).Run()
}

// Add a new permanent and empty zone.
func (c *Client) NewZone(z string) error {
	var opts []string
	opts = appendOption(opts, optPermanent)
	opts = appendOptionWithValue(opts, optNewZone, z)

	return c.executor.Command(commandFirewallCMD, opts...).Run()
}

// Delete an existing permanent zone.
func (c *Client) DeleteZone(z string) error {
	var opts []string
	opts = appendOption(opts, optPermanent)
	opts = appendOptionWithValue(opts, optDeleteZone, z)

	return c.executor.Command(cmd, opts...).Run()
}

// Get predefined zones.
func (c *Client) GetZones(p bool) ([]string, error) {
	var opts []string
	opts = appendOptionCondition(opts, optPermanent, p)
	opts = appendOption(opts, optGetZones)

	out, err := c.executor.Command(commandFirewallCMD, opts...).Output()
	return strings.Fields(string(out)), err
}

// Set the zone's target.
func (c *Client) SetZoneTarget(z string, t Target) error {
	var opts []string
	opts = appendOption(opts, optPermanent)
	opts = appendOptionWithValue(opts, optZone, z)
	opts = appendOptionWithValue(opts, optSetTarget, string(t))

	return c.executor.Command(commandFirewallCMD, opts...).Run()
}

// Get the zone's target.
func (c *Client) GetZoneTarget(z string) (Target, error) {
	var opts []string
	opts = appendOption(opts, optPermanent)
	opts = appendOptionWithValue(opts, optZone, z)
	opts = appendOption(opts, optGetTarget)

	out, err := c.executor.Command(commandFirewallCMD, opts...).Output()
	return Target(bytes.TrimSpace(out)), err
}

// Add a new permanent and empty ipset with specifying the type and optional the
// family.
func (c *Client) NewIPSet(s string, t IPSetType, f IPSetFamily) error {
	var opts []string
	opts = appendOption(opts, optPermanent)
	opts = appendOptionWithValue(opts, optNewIPSet, s)
	opts = appendOptionWithValue(opts, optType, t.String())
	opts = appendOptionWithValue(opts, optFamily, string(f))

	return c.executor.Command(commandFirewallCMD, opts...).Run()
}

// Delete an existing permanent ipset.
func (c *Client) DeleteIPSet(s string) error {
	var opts []string
	opts = appendOption(opts, optPermanent)
	opts = appendOptionWithValue(opts, optDeleteIPSet, s)

	return c.executor.Command(commandFirewallCMD, opts...).Run()
}

// Get ipset list.
func (c *Client) GetIPSets(p bool) ([]string, error) {
	var opts []string
	opts = appendOptionCondition(opts, optPermanent, p)
	opts = appendOption(opts, optGetIPSets)

	out, err := c.executor.Command(commandFirewallCMD, opts...).Output()
	return strings.Fields(string(out)), err
}

// Bind the source to zone. If zone is omitted, default zone will be used.
func (c *Client) AddZoneSources(p bool, z string, s []string) error {
	var opts []string
	opts = appendOptionCondition(opts, optPermanent, p)
	opts = appendOptionWithValue(opts, optZone, z)
	opts = appendOptionWithValue(opts, optAddSource, s...)

	// TODO: firewall-cmd 多次指定参数 --add-source 时只要有一个成功，命令就返回成功，可能部分失败。
	return c.executor.Command(commandFirewallCMD, opts...).Run()
}

// Remove binding of the source from zone it was previously added to.
func (c *Client) RemoveZoneSources(p bool, z string, s []string) error {
	var opts []string
	opts = appendOptionCondition(opts, optPermanent, p)
	opts = appendOptionWithValue(opts, optZone, z)
	opts = appendOptionWithValue(opts, optRemoveSource, s...)

	return c.executor.Command(commandFirewallCMD, opts...).Run()
}

// List sources that are bound to zone. If zone is omitted, default zone will be
// used.
func (c *Client) ListZoneSources(p bool, z string) ([]string, error) {
	var opts []string
	opts = appendOptionCondition(opts, optPermanent, p)
	opts = appendOptionWithValue(opts, optZone, z)
	opts = appendOption(opts, optListSources)

	out, err := c.executor.Command(commandFirewallCMD, opts...).Output()
	return strings.Fields(string(out)), err
}

// Bind interface interface to zone. If zone is omitted, default zone will be
// used.
func (c *Client) AddZoneInterfaces(p bool, z string, i []string) error {
	var opts []string
	opts = appendOptionCondition(opts, optPermanent, p)
	opts = appendOptionWithValue(opts, optZone, z)
	opts = appendOptionWithValue(opts, optAddInterface, i...)

	return c.executor.Command(commandFirewallCMD, opts...).Run()
}

// Change zone the interface is bound to.
func (c *Client) ChangeZoneInterfaces(permanent bool, zone string, interfaces []string) error {
	panic("unimplemented")
}

// Remove binding of interfaces from zone those were previously added to.
func (c *Client) RemoveZoneInterfaces(p bool, z string, i []string) error {
	var opts []string
	opts = appendOptionCondition(opts, optPermanent, p)
	opts = appendOptionWithValue(opts, optZone, z)
	opts = appendOptionWithValue(opts, optRemoveInterface, i...)

	return c.executor.Command(commandFirewallCMD, opts...).Run()
}

// List interfaces that are bound to zone as a space separated list. If zone is
// omitted, default zone will be used.
func (c *Client) ListZoneInterfaces(p bool, z string) ([]string, error) {
	var opts []string
	opts = appendOptionCondition(opts, optPermanent, p)
	opts = appendOptionWithValue(opts, optZone, z)
	opts = appendOption(opts, optListInterfaces)

	out, err := c.executor.Command(commandFirewallCMD, opts...).Output()
	return strings.Fields(string(out)), err
}

// Add new entries to the ipset.
func (c *Client) AddIPSetEntries(p bool, s string, e []string) error {
	var opts []string
	opts = appendOptionCondition(opts, optPermanent, p)
	opts = appendOptionWithValue(opts, optIPSet, s)
	opts = appendOptionWithValue(opts, optAddEntry, e...)

	return c.executor.Command(commandFirewallCMD, opts...).Run()
}

// Remove entries from the ipset.
func (c *Client) RemoveIPSetEntries(p bool, s string, e []string) error {
	var opts []string
	opts = appendOptionCondition(opts, optPermanent, p)
	opts = appendOptionWithValue(opts, optIPSet, s)
	opts = appendOptionWithValue(opts, optRemoveEntry, e...)

	return c.executor.Command(commandFirewallCMD, opts...).Run()
}

// List all entries of the ipset.
func (c *Client) GetIPSetEntries(p bool, s string) ([]string, error) {
	var opts []string
	opts = appendOptionCondition(opts, optPermanent, p)
	opts = appendOptionWithValue(opts, optIPSet, s)
	opts = appendOption(opts, optGetEntries)

	out, err := c.executor.Command(commandFirewallCMD, opts...).Output()
	return strings.Fields(string(out)), err
}

var _ Interface = &Client{}
