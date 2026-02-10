package cs

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"slices"
	"strings"

	exec "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/exec/v1alpha1"
)

// cleanNodeIPtables 清理 Kubernetes 相关的 IPTables 规则
func cleanNodeIPtables(executor exec.Executor) error {
	// Dual families
	for _, cmd := range []struct {
		iptablesSave string
		iptables     string
	}{
		{iptablesSave: "iptables-save", iptables: "iptables"},   // IPv4
		{iptablesSave: "ip6tables-save", iptables: "ip6tables"}, // IPv6
	} {
		// dump iptables
		out, err := executor.Command(cmd.iptablesSave).Output()
		if err != nil {
			return err
		}

		// generate args to clean
		list, err := generateIPTablesArgsListForCleaning(bytes.NewReader(out))
		if err != nil {
			return err
		}

		// execute command to clean
		for _, args := range list {
			if err := executor.Command(cmd.iptables, args...).Run(); err != nil {
				return err
			}
		}
	}

	return nil
}

// generate iptables arguments list for cleaning from iptables dumped data
func generateIPTablesArgsListForCleaning(in io.Reader) (list [][]string, err error) {
	var (
		table  string
		chains []string
	)
	s := bufio.NewScanner(in)
	for s.Scan() {
		line := s.Text()

		line = strings.TrimSpace(line)

		// skip empty lines
		if len(line) == 0 {
			continue
		}

		switch {
		// comment
		case strings.HasPrefix(line, "#"):
			continue

		// table
		case strings.HasPrefix(line, "*"):
			table = strings.TrimPrefix(line, "*")

		// chain
		case strings.HasPrefix(line, ":"):
			parts := strings.SplitN(line[1:], " ", 2)
			if len(parts) != 2 {
				err = fmt.Errorf("invalid rule line: %s", line)
				return
			}
			chain := parts[0]

			// filter chains for kubernetes, cni, calico
			if !matchChain(chain) {
				continue
			}
			chains = append(chains, chain)

			// Delete all rules from the chain
			list = append(list, []string{"-t", table, "-F", chain})

		// commit
		case line == "COMMIT":
			// delete chains from the table
			for _, chain := range chains {
				list = append(list, []string{"-t", table, "-X", chain})
			}

			table = ""
			chains = nil

		// rule
		default:
			fields := strings.Fields(line)

			// skip rules belonging to current chains
			if slices.Contains(chains, fields[1]) {
				continue
			}

			// skip target not belong to current chains
			{
				i := slices.IndexFunc(fields, func(f string) bool { return f == "-j" || f == "--jump" })
				// not found
				if i == -1 {
					continue
				}
				// -j, --jump without target
				if i == len(fields)-1 {
					continue
				}

				// target not belong to current chains
				if target := fields[i+1]; !slices.Contains(chains, target) {
					continue
				}
			}

			var args []string
			args = append(args, "-t", table)
			args = append(args, "-D", fields[1])
			args = append(args, fields[2:]...)

			// remove the rule
			list = append(list, args)
		}
	}

	return
}

// match chain for kubernetes, cni, calico
func matchChain(name string) bool {
	// calico
	if strings.HasPrefix(name, "cali-") {
		return true
	}

	// kubernetes
	if strings.HasPrefix(name, "KUBE-") {
		return true
	}

	// cni
	if strings.HasPrefix(name, "CNI-") {
		return true
	}

	return false
}
