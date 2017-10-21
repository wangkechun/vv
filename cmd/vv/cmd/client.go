package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/wangkechun/vv/client"
	"os"
)

var clientCmd = &cobra.Command{
	Use: "client",
}

var clientEditCmd = &cobra.Command{
	Use: "edit",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("not file specified")
		}
		clientEditCmdCfg.FilePath = args[0]
		if clientEditCmdCfg.Name == "" {
			clientEditCmdCfg.Name, _ = os.Hostname()
		}
		fmt.Printf("vv client [%s] started\n", clientEditCmdCfg.Name)
		return client.New(clientEditCmdCfg.Config).Run()
	},
}
var clientEditCmdCfg struct {
	client.Config
}

func init() {
	clientCmd.AddCommand(clientEditCmd)
	RootCmd.AddCommand(clientCmd)
	clientEditCmd.Flags().StringVarP(&clientEditCmdCfg.Name, "name", "n", "", "client name, default is hostname")
	clientEditCmd.Flags().StringVarP(&clientEditCmdCfg.RegistryAddr, "registry_addr", "r", "127.0.0.1:6655", "registry addr to connect")
}
