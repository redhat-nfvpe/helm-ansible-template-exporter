package export

import (
	"github.com/redhat-nfvpe/helm-ansible-template-exporter/internal/pkg/ansiblegalaxy"
	"github.com/redhat-nfvpe/helm-ansible-template-exporter/internal/pkg/convert"
	"github.com/redhat-nfvpe/helm-ansible-template-exporter/internal/pkg/helm"
	j2parse "github.com/redhat-nfvpe/helm-ansible-template-exporter/internal/pkg/text/template/parse"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
	"path/filepath"
)

var (
	helmChartRef      string
	workspace         string
	roleName          string
	generateFilters   bool
	emitKeysSnakeCase bool
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
	exportCmd.Flags().BoolVar(&generateFilters, "generateFilters", false,"whether or not to install Ansible Filter scaffolding")
	exportCmd.Flags().BoolVar(&emitKeysSnakeCase, "emitKeysSnakeCase", true, "whether or not to convert Ansible keys to snake_case")
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
	j2parse.ReplaceWithSnakeCase = emitKeysSnakeCase
	helm.HelmChartRef = helmChartRef
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
	convert.CopyTemplates(helmChartRef, roleDirectory)
	convert.CopyValuesToDefaults(helmChartRef, roleDirectory)
	convert.RemoveValuesReferencesInDefaults(roleDirectory)
	if emitKeysSnakeCase {
		log.Info("Locating keys that should be converted to snake_case")
		keySet := convert.ConvertDefaultsToSnakeCase(chartClient, roleDirectory)
		j2parse.KnownTextNodeSubstitutions = *keySet
	}
	convert.SuppressWhitespaceTrimmingInTemplates(roleDirectory)
	convert.ConvertControlFlowSyntax(roleDirectory)
	convert.RemoveValuesReferencesInTemplates(roleDirectory)
	// generate the task, which just renders the templates
	convert.InstallAnsibleTasks(roleDirectory)

	// Since Sprig Ansible Filters are not fully implemented, generateFilters CLI argument controls whether or not to
	// install the stub filters.
	if generateFilters {
		log.Info("Installing Sprig Ansible Filters")
		convert.InstallAnsibleFilters(roleDirectory)
	}

	return nil
}
