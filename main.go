package main

import (
	"github.com/redhat-nfvpe/helm-ansible-template-exporter/pkg/ansiblegalaxy"
	"github.com/redhat-nfvpe/helm-ansible-template-exporter/pkg/helm"
	"path/filepath"
)

func main() {
	// TODO Each of these should be converted to a CLI option
	// https://github.com/redhat-nfvpe/helm-ansible-template-exporter/issues/2
	rolesDirectory := "./workspace"
	roleName := "test"
	helmChartRootDirectory := "/Users/ryangoulding/workspace/bitnami_nginx_charts/charts/bitnami/nginx"

	// Contains the directory of the role within the scratch space.
	roleDirectory := filepath.Join(rolesDirectory, roleName)

	// Does the conversion work.
	ansiblegalaxy.InstallAnsibleRole(roleName, rolesDirectory)
	helm.CopyTemplates(helmChartRootDirectory, roleDirectory)
	helm.CopyValuesToDefaults(helmChartRootDirectory, roleDirectory)
	helm.RemoveValuesReferencesInDefaults(roleDirectory)
	helm.RemoveValuesReferencesInTemplates(roleDirectory)
}