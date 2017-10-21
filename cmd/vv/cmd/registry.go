package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wangkechun/vv/pkg/registry"
)

var registerCmd = &cobra.Command{
	Use: "register",
}

var registerStartCmd = &cobra.Command{
	Use: "start",
	RunE: func(cmd *cobra.Command, args []string) error {
		registerStartCmdCfg.RegistryAddrRPC = defaultRegistryRPC
		registerStartCmdCfg.RegistryAddrTCP = defaultRegistryTCP
		return registry.New(registerStartCmdCfg.Config).Run()
	},
}
var registerStartCmdCfg struct {
	registry.Config
}

func init() {
	registerCmd.AddCommand(registerStartCmd)
	RootCmd.AddCommand(registerCmd)
}
