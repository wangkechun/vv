package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const keyUsername = "username"
const keyPassword = "password"

var passwordCmd = &cobra.Command{
	Use: "password",
}

var passwordSetCmd = &cobra.Command{
	Use: "set <username> <password>",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 2 {
			return cmd.Help()
		}
		viper.Set(keyUsername, args[0])
		viper.Set(keyPassword, args[1])
		return nil
	},
}

var passwordGetCmd = &cobra.Command{
	Use: "get",
	RunE: func(cmd *cobra.Command, args []string) error {
		username := viper.Get(keyUsername).(string)
		password := viper.Get(keyPassword).(string)
		fmt.Printf("user: %s\n", username)
		fmt.Printf("password: %s\n", password)
		fmt.Printf("token: %s:%s\n", username, password)
		return nil
	},
}

func init() {
	passwordCmd.AddCommand(passwordSetCmd)
	passwordCmd.AddCommand(passwordGetCmd)
	RootCmd.AddCommand(passwordCmd)
}
