package cmd

import (
	"fmt"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pborman/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "vv",
	Short: "",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return clientEditCmd.RunE(cmd, args)
	},
}

// Execute ...
func Execute() {
	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "client", "help", "server", "register", "password":
		default:
			os.Args = []string{os.Args[0], "client", "edit", os.Args[1]}
		}
	}
	if viper.Get(keyUsername) == nil || viper.Get(keyPassword) == nil {
		hostName, _ := os.Hostname()
		if hostName == "" {
			hostName = "unknow"
		}
		password := uuid.NewUUID().String()
		viper.Set(keyUsername, hostName)
		viper.Set(keyPassword, password)

	}
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.vv.yaml)")
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".vv")
	}

	viper.AutomaticEnv() // read in environment variables that match
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
	// } else {
	// 	fmt.Println("ReadInConfig error", viper.ConfigFileUsed())
	// 	os.Exit(1)
	// }
}
