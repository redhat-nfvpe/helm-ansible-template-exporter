package main

import "path/filepath"

func main() {
	// TODO Each of these should be converted to a CLI option
	rolesDirectory := "./workspace"
	roleName := "test"
	helmChartRootDirectory := "/Users/ryangoulding/workspace/bitnami_nginx_charts/charts/bitnami/nginx"

	// Contains the directory of the role within the scratch space.
	roleDirectory := filepath.Join(rolesDirectory, roleName)

	// Does the conversion work.
	InstallAnsibleRole(roleName, rolesDirectory)
	CopyTemplates(helmChartRootDirectory, roleDirectory)
	CopyValuesToDefaults(helmChartRootDirectory, roleDirectory)
	RemoveValuesReferencesInDefaults(roleDirectory)
	RemoveValuesReferencesInTemplates(roleDirectory)
}