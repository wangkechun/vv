package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wangkechun/vv/pkg/fuse/lab"
	"github.com/wangkechun/vv/pkg/server"
)

var mountCmd = &cobra.Command{
	Use: "mount",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("pre mount .")
		lab.FunFuseG1()
		return nil
	},
}

func init() {
	RootCmd.AddCommand(mountCmd)
}
