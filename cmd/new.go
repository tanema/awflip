package cmd

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	//go:embed data/app_template/*
	appTemplate         embed.FS
	customizedFilenames = []string{
		"Cargo.toml",
		"README.md",
	}
)

var newCmd = &cobra.Command{
	Use:   "new [appID]",
	Short: "generate a new project for the flipper zero",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appID := args[0]
		fmt.Println("Generating project:", appID)
		cobra.CheckErr(os.Mkdir(appID, 0755))
		cobra.CheckErr(fs.WalkDir(appTemplate, "data/app_template", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			destFile := strings.TrimPrefix(strings.TrimPrefix(path, "data/app_template"), "/")
			if strings.HasPrefix(destFile, "_") {
				destFile = "." + strings.TrimPrefix(destFile, "_")
			} else if destFile == "" {
				return nil
			}
			fmt.Println("    ==>", destFile)
			if d.IsDir() {
				return os.MkdirAll(filepath.Join(appID, destFile), 0755)
			} else if data, err := appTemplate.ReadFile(path); err != nil {
				return err
			} else {
				return os.WriteFile(filepath.Join(appID, destFile), data, 0644)
			}
		}))
		for _, f := range customizedFilenames {
			cobra.CheckErr(findAndReplace(filepath.Join(appID, f), "new_flipper_app", appID))
		}
		fmt.Println("Done.")
	},
}

func findAndReplace(filePath, oldVal, newVal string) error {
	read, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	newContents := strings.Replace(string(read), oldVal, newVal, -1)
	return os.WriteFile(filePath, []byte(newContents), 0)
}
