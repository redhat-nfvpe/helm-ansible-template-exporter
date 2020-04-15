package cmd

import (
	"fmt"
	cmd "github.com/redhat-nfvpe/helm-ansible-template-exporter/cmd/hamsible/export"
	"github.com/spf13/cobra"
	"os"
)

var root = &cobra.Command{
	Use:   "helmExport",
	Short: "A tool to convert helm charts into ansible templates and roles",
}

func Execute() {
	root.AddCommand(cmd.GetExportCmd())
	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
