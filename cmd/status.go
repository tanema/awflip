package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tanema/ufbt/lib/flipper"
)

var statusCmd = &cobra.Command{
	Use:    "status",
	Short:  "display the status of the device",
	PreRun: establishDevice,
	Run: func(cmd *cobra.Command, args []string) {
		info, err := device.Info(flipper.InfoDevice)
		cobra.CheckErr(err)
		fmt.Println("Firmware")
		fmt.Printf("  API:     %v.%v\n", info["firmware.api.major"], info["firmware.api.minor"])
		fmt.Printf("  Version: %v\n", info["firmware.version"])
		fmt.Println("Hardware")
		fmt.Printf("  UID:     %v\n", info["hardware.uid"])
		fmt.Printf("  Name:    %v\n", info["hardware.name"])
		fmt.Printf("  Model:   %v\n", info["hardware.model"])
		fmt.Printf("  Version: %v\n", info["hardware.ver"])
	},
}
