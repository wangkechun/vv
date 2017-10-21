package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wangkechun/vv/server"
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
	serverStartCmd.Flags().StringVarP(&serverStartCmdCfg.RegistryAddr, "registry_addr", "r", "127.0.0.1:6655", "registry addr to connect")
}
