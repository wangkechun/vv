package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wangkechun/vv/pkg/registry"
)

var registerCmd = &cobra.Command{
	Use: "register",
}

var registerStartCmd = &cobra.Command{
	Use: "start",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("register start on ", registerStartCmdCfg.addr)
		return registry.New(registry.Config{
			Addr: registerStartCmdCfg.addr,
		}).Run()
	},
}
var registerStartCmdCfg struct {
	addr string
}

func init() {
	registerCmd.AddCommand(registerStartCmd)
	RootCmd.AddCommand(registerCmd)
	registerStartCmd.Flags().StringVarP(&registerStartCmdCfg.addr, "addr", "A", "0.0.0.0:6655", "")
}
