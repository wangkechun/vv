package cmd

import "os"

var defaultRegistryTCP = "evm.hi-hi.cn:6655"
var defaultRegistryRPC = "evm.hi-hi.cn:6656"
var defaultName = "hello"

func init() {
	if os.Getenv("TEST_VV") != "" {
		defaultRegistryTCP = "127.0.0.1:6655"
		defaultRegistryRPC = "127.0.0.1:6656"
	}
}
