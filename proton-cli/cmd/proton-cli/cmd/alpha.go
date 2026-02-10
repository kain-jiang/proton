package cmd

import (
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
)

func newAlphaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alpha",
		Short: "Proton-cli experimental sub-commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newNetInterfaceMTUByAddress())

	return cmd
}

func newNetInterfaceMTUByAddress() *cobra.Command {
	return &cobra.Command{
		Use:   "net-interface-mtu-by-address",
		Short: "Get network interface mtu by address",
		RunE: func(cmd *cobra.Command, args []string) error {
			want := net.ParseIP(args[0])
			if want == nil {
				return fmt.Errorf("invalid address: %v", args[0])
			}

			interfaces, err := net.Interfaces()
			if err != nil {
				return err
			}

			seen := sets.New[string]()
			var mtu int
			for _, each := range interfaces {
				addrs, err := each.Addrs()
				if err != nil {
					return err
				}
				for _, addr := range addrs {
					n, ok := addr.(*net.IPNet)
					if !ok {
						continue
					}
					if !want.Equal(n.IP) {
						continue
					}

					seen.Insert(each.Name)
					mtu = each.MTU
				}
			}
			if mtu == 0 {
				return fmt.Errorf("network interface that contain address %q was not found", args[0])
			}
			if seen.Len() > 1 {
				return fmt.Errorf("address %q was found on multi network interfaces: %s", args[0], strings.Join(sets.List(seen), ", "))
			}
			fmt.Println(mtu)
			return nil
		},
		Args: cobra.ExactArgs(1),
	}
}
