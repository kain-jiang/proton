package base

import (
	"fmt"
	"net"
	"strconv"

	"component-manage/internal/global"
	"component-manage/pkg/ecms"

	corev1 "k8s.io/api/core/v1"
)

const ecmsPort = 9202

func ClearStorage(hosts []string, path string) error {
	nodes, err := global.K8sCli.ListNodes()
	if err != nil {
		return fmt.Errorf("list nodes error: %w", err)
	}

	getECMSCli := func(name string) (ecms.Client, error) {
		for _, node := range nodes {
			if node.ObjectMeta.Name == name {
				for _, addr := range node.Status.Addresses {
					if addr.Type == corev1.NodeInternalIP {
						return ecms.New(net.JoinHostPort(addr.Address, strconv.Itoa(ecmsPort))), nil
					}
				}
				return nil, fmt.Errorf("node %s cannot get internal ip", name)
			}
		}
		return nil, fmt.Errorf("node %s not found", name)
	}

	for _, host := range hosts {
		cli, err := getECMSCli(host)
		if err != nil {
			return fmt.Errorf("get ecms client error: %w", err)
		}
		ok, err := cli.DirectoryExist(path)
		if err != nil {
			return fmt.Errorf("check data path error: %w", err)
		}

		if ok {
			err := cli.DirectoryDelete(path)
			if err != nil {
				return fmt.Errorf("delete data path error: %w", err)
			}
		} else {
			global.Logger.
				WithField("cli", "ecms").
				WithField("host", host).
				WithField("path", path).
				Info("data path is not exist, skip clean")
		}
	}
	return nil
}

func PrepareStorage(hosts []string, path string) error {
	nodes, err := global.K8sCli.ListNodes()
	if err != nil {
		return fmt.Errorf("list nodes error: %w", err)
	}

	getECMSCli := func(name string) (ecms.Client, error) {
		for _, node := range nodes {
			if node.ObjectMeta.Name == name {
				for _, addr := range node.Status.Addresses {
					if addr.Type == corev1.NodeInternalIP {
						return ecms.New(net.JoinHostPort(addr.Address, strconv.Itoa(ecmsPort))), nil
					}
				}
				return nil, fmt.Errorf("node %s cannot get internal ip", name)
			}
		}
		return nil, fmt.Errorf("node %s not found", name)
	}

	for _, host := range hosts {
		cli, err := getECMSCli(host)
		if err != nil {
			return fmt.Errorf("get ecms client error: %w", err)
		}
		ok, err := cli.DirectoryExist(path)
		if err != nil {
			return fmt.Errorf("check data path error: %w", err)
		}

		if !ok {
			err := cli.DirectoryCreate(path)
			if err != nil {
				return fmt.Errorf("create data path error: %w", err)
			}
		} else {
			global.Logger.
				WithField("cli", "ecms").
				WithField("host", host).
				WithField("path", path).
				Info("data path exist, skip create")
		}
	}
	return nil
}
