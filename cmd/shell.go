package cmd

import (
	"fmt"
	"strings"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:    "shell",
	Short:  "run an interactive shell on the device.",
	PreRun: establishDevice,
	Run: func(cmd *cobra.Command, args []string) {
		rl, err := readline.New("> ")
		if err != nil {
			panic(err)
		}
		defer rl.Close()

		fmt.Println(device.Banner)
		for {
			line, err := rl.Readline()
			cobra.CheckErr(err)
			if strings.TrimSpace(line) == "exit" {
				break
			} else if resp, err := device.Request(line); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(resp)
			}
		}
	},
}
