package slb

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	slbv2 "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/slb/v2"
)

const (
	keepalivedConfigPath  = "/etc/keepalived/keepalived.conf"
	keepalivedScriptRoot  = "/etc/slb/scripts"
	keepalivedMasterSrc   = "/etc/slb/scripts/entering_master.py"
	keepalivedBackupSrc   = "/etc/slb/scripts/entering_backup.py"
	keepalivedServiceName = "keepalived"
)

func (r *Remote) EnsureKeepalivedScriptSecurity(ctx context.Context) error {
	current, exists, err := r.readFileIfExists(ctx, keepalivedConfigPath)
	if err != nil || !exists {
		return err
	}

	conf, err := parseKeepalivedConfig(current)
	if err != nil {
		return err
	}
	if len(conf.blocks("vrrp_instance")) == 0 {
		return nil
	}

	globalDefs := conf.ensureGlobalDefs()
	changed := false
	if globalDefs.key("enable_script_security") == nil {
		globalDefs.setKey("enable_script_security", "")
		changed = true
	}
	if globalDefs.key("script_user") == nil {
		globalDefs.setKey("script_user", "root")
		changed = true
	}
	if !changed {
		return nil
	}

	rendered := conf.render()
	if err := r.writeFile(ctx, keepalivedConfigPath, rendered); err != nil {
		return err
	}
	if msg, err := r.testKeepalivedConfig(); err != nil {
		_ = r.writeFile(ctx, keepalivedConfigPath, current)
		if msg != "" {
			return fmt.Errorf("invalid keepalived config: %s", msg)
		}
		return fmt.Errorf("invalid keepalived config: %w", err)
	}

	return r.reloadOrRestartKeepalived()
}

func (r *Remote) EnsureKeepalivedHAInstance(ctx context.Context, name string, instance *slbv2.KeepalivedHA) error {
	current, exists, err := r.readFileIfExists(ctx, keepalivedConfigPath)
	if err != nil {
		return err
	}
	if !exists {
		return fs.ErrNotExist
	}

	conf, err := parseKeepalivedConfig(current)
	if err != nil {
		return err
	}
	if len(conf.Items) == 0 || conf.block("global_defs", "") == nil {
		globalDefs := conf.ensureGlobalDefs()
		if globalDefs.key("router_id") == nil {
			globalDefs.setKey("router_id", uuid.NewString())
		}
		if globalDefs.key("vrrp_garp_master_refresh") == nil {
			globalDefs.setKey("vrrp_garp_master_refresh", "30")
		}
	}

	existing := conf.block("vrrp_instance", name)
	if existing == nil {
		masterPath, backupPath, err := r.ensureKeepalivedScripts(ctx, name)
		if err != nil {
			return err
		}
		block := buildKeepalivedHAInstanceBlock(name, instance, masterPath, backupPath)
		conf.upsertBlock(block)
	} else {
		updateKeepalivedHAInstanceBlock(existing, instance)
	}

	rendered := conf.render()
	if !bytes.Equal(current, rendered) {
		if err := r.writeFile(ctx, keepalivedConfigPath, rendered); err != nil {
			return err
		}
		if msg, err := r.testKeepalivedConfig(); err != nil {
			_ = r.writeFile(ctx, keepalivedConfigPath, current)
			if msg != "" {
				return fmt.Errorf("invalid keepalived config: %s", msg)
			}
			return fmt.Errorf("invalid keepalived config: %w", err)
		}
	}

	return r.reloadOrRestartKeepalived()
}

func (r *Remote) ensureKeepalivedScripts(ctx context.Context, name string) (string, string, error) {
	dir := filepath.Join(keepalivedScriptRoot, name)
	if err := r.ensureDir(ctx, dir); err != nil {
		return "", "", err
	}

	masterDst := filepath.Join(dir, filepath.Base(keepalivedMasterSrc))
	backupDst := filepath.Join(dir, filepath.Base(keepalivedBackupSrc))
	if err := r.copyFile(ctx, keepalivedMasterSrc, masterDst); err != nil {
		return "", "", err
	}
	if err := r.copyFile(ctx, keepalivedBackupSrc, backupDst); err != nil {
		return "", "", err
	}
	return masterDst, backupDst, nil
}

func buildKeepalivedHAInstanceBlock(name string, instance *slbv2.KeepalivedHA, masterPath, backupPath string) *keepalivedBlock {
	block := &keepalivedBlock{Kind: "vrrp_instance", Name: name}
	block.setKey("interface", instance.Interface)
	block.setKey("state", "BACKUP")
	block.setKey("virtual_router_id", instance.VirtualRouterID)
	block.setKey("priority", instance.Priority)
	block.setKey("advert_int", "1")
	block.setKey("nopreempt", "")
	block.setKey("unicast_src_ip", instance.UnicastSRC_IP)
	block.setKey("notify_master", masterPath)
	block.setKey("notify_backup", backupPath)
	if len(instance.UnicastPeer) != 0 {
		block.setChild(buildKeepalivedKeyBlock("unicast_peer", toStringSlice(instance.UnicastPeer)))
	}
	if len(instance.VirtualIPAddress) != 0 {
		block.setChild(buildKeepalivedMapBlock("virtual_ipaddress", instance.VirtualIPAddress))
	}
	return block
}

func updateKeepalivedHAInstanceBlock(block *keepalivedBlock, instance *slbv2.KeepalivedHA) {
	block.setKey("interface", instance.Interface)
	block.setKey("virtual_router_id", instance.VirtualRouterID)
	block.setKey("priority", instance.Priority)
	block.setKey("unicast_src_ip", instance.UnicastSRC_IP)
	if block.key("state") == nil {
		block.setKey("state", "BACKUP")
	}
	if block.key("advert_int") == nil {
		block.setKey("advert_int", "1")
	}
	if block.key("nopreempt") == nil {
		block.setKey("nopreempt", "")
	}
	if len(instance.NotifyMaster) != 0 {
		block.setKey("notify_master", instance.NotifyMaster)
	}
	if len(instance.NotifyBackup) != 0 {
		block.setKey("notify_backup", instance.NotifyBackup)
	}
	if len(instance.UnicastPeer) != 0 {
		block.setChild(buildKeepalivedKeyBlock("unicast_peer", toStringSlice(instance.UnicastPeer)))
	}
	if len(instance.VirtualIPAddress) != 0 {
		block.setChild(buildKeepalivedMapBlock("virtual_ipaddress", instance.VirtualIPAddress))
	}
}

func buildKeepalivedKeyBlock(kind string, values []string) *keepalivedBlock {
	block := &keepalivedBlock{Kind: kind}
	for _, value := range values {
		block.Items = append(block.Items, &keepalivedKey{Name: value})
	}
	return block
}

func buildKeepalivedMapBlock(kind string, values map[string]string) *keepalivedBlock {
	block := &keepalivedBlock{Kind: kind}
	for _, key := range sortedKeys(values) {
		block.Items = append(block.Items, &keepalivedKey{Name: key, Value: values[key]})
	}
	return block
}

func (r *Remote) testKeepalivedConfig() (string, error) {
	msg, err := r.testCommand("keepalived -t 2>&1")
	msg = strings.TrimSpace(msg)
	if msg == "SECURITY VIOLATION - scripts are being executed but script_security not enabled." {
		return "", nil
	}
	return msg, err
}

func (r *Remote) reloadOrRestartKeepalived() error {
	r.resetFailedService(keepalivedServiceName)
	if r.serviceActive(keepalivedServiceName) {
		return r.exec.Command("systemctl", "reload", keepalivedServiceName).Run()
	}
	return r.exec.Command("systemctl", "restart", keepalivedServiceName).Run()
}

func toStringSlice[T ~string](items []T) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, string(item))
	}
	return result
}
