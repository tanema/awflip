package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tanema/ufbt/lib/flipper"
)

var (
	device    *flipper.Device
	deviceErr error
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ufbt",
	Short: "A brief description of your application",
	//Run: func(cmd *cobra.Command, args []string) { },
}

func establishDevice(cmd *cobra.Command, args []string) {
	device, deviceErr = flipper.Open("auto")
	cobra.CheckErr(deviceErr)
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(
		shellCmd,
		logCmd,
		readCmd,
		writeCmd,
		statusCmd,
		newCmd,
		lsCmd,
	)
	rootCmd.PersistentFlags().StringP("port", "p", "auto", "Device portname to specify a specific device if there are multiple connected. If set to auto, the device will be found automatically.")
}
