package export

import (
	"github.com/redhat-nfvpe/helm-ansible-template-exporter/internal/pkg/ansiblegalaxy"
	"github.com/redhat-nfvpe/helm-ansible-template-exporter/internal/pkg/helm"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
	"path/filepath"
)

var (
	helmChartRef  string
	workspace     string
	roleName      string
)

func GetExportCmd() *cobra.Command {

	var exportCmd = &cobra.Command{
		Use:   "export <Role Name>",
		Short: "Export helm charts to ansible roles",
		Long:  "Export helm charts to ansible roles",
		RunE:  exportFunc,
	}
	exportCmd.Flags().StringVar(&helmChartRef, "helm-chart", "", "Path is downloaded helm chart folder.")
	exportCmd.Flags().StringVar(&workspace, "workspace", "workspace", "workspace to generate exported ansible role.")

	return exportCmd
}

func exportFunc(cmd *cobra.Command, args []string) error {

	chartClient := helm.NewChartClient()

	if err := parse(args); err != nil {
		log.Error("error parsing arguments: ", err)
		return err
	}
	if err := verifyFlags(); err != nil {
		log.Error("error verifying flags: ", err)
		return err
	}

	err := chartClient.LoadChartFrom(helmChartRef)

	if err != nil {
		log.Error("error loading chart: ", err)
		return err
	}

	// Contains the directory of the role within the scratch space.
	roleDirectory := filepath.Join(workspace, roleName)

	// Does the conversion work.
	/*TODO: use helmcnart client to read loaded helm templates and values*/

	/*Example
	use chartClient.Chart.Templates for reading templates
	*/
	ansiblegalaxy.InstallAnsibleRole(roleName, workspace)
	helm.CopyTemplates(helmChartRef, roleDirectory)
	helm.CopyValuesToDefaults(helmChartRef, roleDirectory)
	helm.RemoveValuesReferencesInDefaults(roleDirectory)
	helm.SuppressWhitespaceTrimmingInTemplates(roleDirectory)
	helm.ConvertControlFlowSyntax(roleDirectory)
	helm.RemoveValuesReferencesInTemplates(roleDirectory)
	// generate the task, which just renders the templates
	helm.InstallAnsibleTasks(roleDirectory)

	return nil
}
