package cmd

import (
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:     "read [devicePath] [localPath]",
	Example: "read /ext/apps/Games/snake_game.fap .",
	Short:   "read a file on the device, to a local path",
	Args:    cobra.ExactArgs(2),
	PreRun:  establishDevice,
	Run: func(cmd *cobra.Command, args []string) {
		cobra.CheckErr(device.Read(args[0], args[1]))
	},
}
