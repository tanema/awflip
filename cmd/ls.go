package cmd

import (
	"os"
	"text/template"

	"github.com/spf13/cobra"
)

var fileTmpl = template.Must(template.New("file").Parse(`{{.Path}}{{if .IsDir}}/{{else}}  {{.Size}}b{{end}}
`))

var lsCmd = &cobra.Command{
	Use:    "ls <[/int|/ext]path>",
	Short:  "list files in a specific directory",
	Args:   cobra.ExactArgs(1),
	PreRun: establishDevice,
	Run: func(cmd *cobra.Command, args []string) {
		paths, err := device.Ls(args[0])
		cobra.CheckErr(err)
		for _, info := range paths {
			fileTmpl.Execute(os.Stdout, info)
		}
	},
}
