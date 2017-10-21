package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wangkechun/vv/pkg/server"
	"os"
)

var serverCmd = &cobra.Command{
	Use: "server",
}

var serverStartCmd = &cobra.Command{
	Use: "start",
	RunE: func(cmd *cobra.Command, args []string) error {
		if serverStartCmdCfg.Name == "" {
			serverStartCmdCfg.Name, _ = os.Hostname()
		}
		serverStartCmdCfg.Name = "123456"
		serverStartCmdCfg.RegistryAddrRPC = defaultRegistryRPC
		serverStartCmdCfg.RegistryAddrTCP = defaultRegistryTCP
		fmt.Printf("vv server [%s] started\n", serverStartCmdCfg.Name)
		return server.New(serverStartCmdCfg.Config).Run()
	},
}

var serverStartCmdCfg struct {
	server.Config
}

func init() {
	serverCmd.AddCommand(serverStartCmd)
	RootCmd.AddCommand(serverCmd)
	serverStartCmd.Flags().StringVarP(&serverStartCmdCfg.Name, "name", "n", "", "server name, default is hostname")
}
