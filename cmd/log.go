package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// logCmd represents the log command
var logCmd = &cobra.Command{
	Use:    "log",
	Short:  "stream logs from the device",
	PreRun: establishDevice,
	RunE: func(cmd *cobra.Command, args []string) error {
		return device.Log(os.Stdout, args[0])
	},
}
