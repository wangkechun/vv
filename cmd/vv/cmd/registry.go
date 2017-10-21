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
		fmt.Println("register start on ", registerStartCmdCfg.Addr)
		fmt.Println("register start on ", registerStartCmdCfg.Addr2)
		return registry.New(registerStartCmdCfg.Config).Run()
	},
}
var registerStartCmdCfg struct {
	registry.Config
}

func init() {
	registerCmd.AddCommand(registerStartCmd)
	RootCmd.AddCommand(registerCmd)
	registerStartCmd.Flags().StringVarP(&registerStartCmdCfg.Addr, "addr", "A", "0.0.0.0:6655", "")
	registerStartCmd.Flags().StringVarP(&registerStartCmdCfg.Addr2, "addr2", "a", "0.0.0.0:6656", "")
}
