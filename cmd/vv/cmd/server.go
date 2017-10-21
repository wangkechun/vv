package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wangkechun/vv/pkg/fuse/lab2"
	"github.com/wangkechun/vv/pkg/server"
	"net/http"
	"qiniupkg.com/x/log.v7"
)

var serverCmd = &cobra.Command{
	Use: "server",
}

var serverStartCmd = &cobra.Command{
	Use: "start",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info("connecting to evm.hi-hi.cn:6656")
		log.Info("connecting success")
		ch := make(chan int, 5)
		http.HandleFunc("/", func(http.ResponseWriter, *http.Request) {
			ch <- 1
		})
		go http.ListenAndServe(":8855", nil)
		lab.FunFuse(ch)
		return nil
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
