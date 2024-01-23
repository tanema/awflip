package cmd

import (
	"github.com/spf13/cobra"
)

var writeCmd = &cobra.Command{
	Use:    "write [srcPath] [dstPath]",
	Short:  "write a file from local storage to the device",
	Args:   cobra.ExactArgs(2),
	PreRun: establishDevice,
	Run: func(cmd *cobra.Command, args []string) {
		cobra.CheckErr(device.Write(args[0], args[1]))
	},
}
