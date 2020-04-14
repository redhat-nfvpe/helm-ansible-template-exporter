/*
Package helm provides utilities to aid in the conversion from a Helm chart into an Ansible Playbook Role.
 */
package helm

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const ansibleRoleDefaultsDirectory = "defaults"
const ansibleRoleTemplatesDirectory = "templates"
const ansibleRoleMainYamlFileName = "main.yml"
const defaultPermissions = 0600
const helmDefaultsContainsSelfReference =
	"# TODO: Replace \".Values.\" reference with a literal, as Ansible Playbook doesn't allow self-reference\n"
const helmTemplatesDirectory = "templates"
const helmValuesFilePrefix = "values"
const j2Extension = "j2"
const valuesString = ".Values."
const yamlSuffix = "yaml"
const ymlSuffix = "yml"

// Given an Ansible Role Directory, return the path to the templates directory.  This does not check for the existence
// or readability of the underlying directory.
func getAnsibleRoleTemplatesDirectory(roleDirectory string) string {
	return filepath.Join(roleDirectory, ansibleRoleTemplatesDirectory)
}

// Given an Ansible Role directory, return the path to the defaults directory.  This does not check for the existence
// or readability of the underlying directory.
func getAnsibleRoleDefaultsDirectory(roleDirectory string) string {
	return filepath.Join(roleDirectory, ansibleRoleDefaultsDirectory)
}

// Given an Ansible Role directory, return the path to the defaults main.yml file name.  This does not check the
// existence or permissions of the underlying file.
func getAnsibleRoleDefaultsFileName(roleDirectory string) string {
	return filepath.Join(getAnsibleRoleDefaultsDirectory(roleDirectory), ansibleRoleMainYamlFileName)
}

// Given a Helm chart root directory, return the path to the templates directory.
func getHelmChartTemplatesDirectory(helmChartRootDirectory string) string {
	return filepath.Join(helmChartRootDirectory, helmTemplatesDirectory)
}

// Determine if the given filename is representative of a Helm values file (i.e., values.yml or values.yaml).
func isHelmValuesFile(fileName string) bool {
	return isYamlFile(fileName) && strings.HasPrefix(fileName, helmValuesFilePrefix)
}

// Given a Helm chart root directory, return the file path of the values file.  If a values file cannot be found, an
// appropriate Error is returned.
func getHelmChartValuesFile(helmChartRootDirectory string) (string, error) {
	files, _ := readDir(helmChartRootDirectory)
	for _, file := range files {
		fileName := file.Name()
		if isHelmValuesFile(fileName) {
			return filepath.Join(helmChartRootDirectory, fileName), nil
		}
	}
	return "", errors.New("Cannot resolve values.yml or values.yaml")
}

// Checks a directory for existence, exiting fatally if said directory does not exist.
func checkDirectoryExistence(directory string, errorMessage string) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		logrus.Errorf("%s: %s", errorMessage, directory)
		logrus.Fatal(err)
	}
}

// Reads the files in a directory, exiting fatally if any errors occur.
func readDir(directory string) ([]os.FileInfo, error) {
	logrus.Debugf("Attempting to read: %s", directory)
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Debugf("Successfully read: %s", directory)
	return files, err
}

// Extract whether a fileName represents a YAML file.  This function does not check for file existence.
func isYamlFile(fileName string) bool {
	return strings.HasSuffix(fileName, yamlSuffix) || strings.HasSuffix(fileName, ymlSuffix)
}

// Translates a YAML fileName into a Jinja2 fileName
func yamlToJ2FileName(fileName string) string {
	return fileName + "." + j2Extension
}

// Copies Helm Yaml templates to the appropriate Ansible Playbook roles template, post-fixing each YAML file with a
// ".j2" extension.  The path to the ansible playbook templates directory is returned.
func CopyTemplates(helmChartRootDirectory string, rolesDirectory string) string {
	chartTemplatesDir := getHelmChartTemplatesDirectory(helmChartRootDirectory)
	checkDirectoryExistence(chartTemplatesDir, "Cannot read the template directory")
	ansiblePlaybookTemplatesDirectory := getAnsibleRoleTemplatesDirectory(rolesDirectory)

	files, _ := readDir(chartTemplatesDir)
	for _, file := range files {
		fileName := file.Name()
		if isYamlFile(fileName) {
			chartTemplateFileName := filepath.Join(chartTemplatesDir, fileName)
			contents, err := ioutil.ReadFile(chartTemplateFileName)
			j2FileName := yamlToJ2FileName(fileName)
			ansiblePlaybookTemplateFilename := filepath.Join(ansiblePlaybookTemplatesDirectory, j2FileName)
			if err != nil {
				// TODO Implement a strict option which fails the conversion.
				logrus.Warnf("Read failure, skipping copy of: %s to %s", chartTemplateFileName,
					ansiblePlaybookTemplateFilename)
			}

			logrus.Debugf("Attempting to copy: %s to %s", chartTemplateFileName, ansiblePlaybookTemplateFilename)
			err = ioutil.WriteFile(ansiblePlaybookTemplateFilename, contents, defaultPermissions)
			if err != nil {
				// TODO Implement a strict option which fails the conversion.
				logrus.Warnf("Write failure, skipping copy of: %s to %s", chartTemplateFileName,
					ansiblePlaybookTemplateFilename)
			} else {
				logrus.Infof("Successfully copied: %s to %s", chartTemplateFileName,
					ansiblePlaybookTemplateFilename)
			}
		}
	}
	return ansiblePlaybookTemplatesDirectory
}

// Appends contents to the end of a file.
func appendFile(contents string, destinationFileName string) {
	file, err := os.OpenFile(destinationFileName, os.O_APPEND|os.O_WRONLY, defaultPermissions)
	if err != nil {
		logrus.Fatal(err)
	}
	defer file.Close()
	if _, err := file.WriteString(
		"\n\n# Everything Below this line was inserted by helmExport\n\n" + contents); err != nil {
		logrus.Fatal(err)
	}
}

// Copies the contents of a Helm templates values.yml or values.yaml to the corresponding Ansible Role's
// defaults/main.yaml.
func CopyValuesToDefaults(chartRoot string, roleDirectory string) {
	valuesFileName, err := getHelmChartValuesFile(chartRoot)
	if err != nil {
		logrus.Warnf("Skipping copying values file as it could not be found at: %s", chartRoot)
		return
	}
	rolesDefaultsFileName := getAnsibleRoleDefaultsFileName(roleDirectory)
	logrus.Debugf("Processing values file: %s to %s", valuesFileName, rolesDefaultsFileName)

	contents, err := ioutil.ReadFile(valuesFileName)
	if err != nil {
		logrus.Errorf("Couldn't read: %s", valuesFileName)
		logrus.Fatal(err)
	}

	appendFile(string(contents), rolesDefaultsFileName)
}

// Forms a hint comment that a manual fix is needed in defaults/main.yml due to a ".Values." self reference.
func formManualFixIsRequiredHint(line string) string {
	return helmDefaultsContainsSelfReference + "# " + line
}

// Given an Ansible role directory, correct the defaults/main.yml file for ".Values." references.  Although Helm allows
// self-reference, Ansible Playbook does not support this behavior.  Thus, report to the user that the file will need
// manual inspection/editing after export, and put the appropriate hint in the exported Ansible Playbook.
func RemoveValuesReferencesInDefaults(roleDirectory string) {
	valuesFileName := getAnsibleRoleDefaultsFileName(roleDirectory)

	input, err := ioutil.ReadFile(valuesFileName)
	if err != nil {
		logrus.Fatalln(err)
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if strings.Contains(line, valuesString) {
			logrus.Warnf("Self-reference in %s line %d requires a manual fix after helmConvert finishes",
				valuesFileName, i)
			lines[i] = formManualFixIsRequiredHint(lines[i])
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(valuesFileName, []byte(output), defaultPermissions)
	if err != nil {
		logrus.Fatalln(err)
	} else {
		logrus.Infof("Successfully commented out references to .Values. in: %s", valuesFileName)
	}
}

// Given a template file name, replace all references to ".Values." with the empty string.  Ansible Playbook allows
// direct access to defaults defined in defaults/main.yml.
func removeValuesReferencesInTemplate(templateFileName string) {
	input, err := ioutil.ReadFile(templateFileName)
	if err != nil {
		logrus.Warnf("Skipping .Values. substitution, couldn't read file: %s", templateFileName)
		return
	}

	lines := strings.Split(string(input), "\n")

	lineNumbers := []int{}

	for i, line := range lines {
		if strings.Contains(line, valuesString) {
			logrus.Debugf("Removing .Values. reference in %s on line %d", templateFileName, i)
			lines[i] = strings.ReplaceAll(lines[i], valuesString, "")
			lineNumbers = append(lineNumbers, i)
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(templateFileName, []byte(output), defaultPermissions)
	if err != nil {
		logrus.Warnf("Skipping .Values. removal, couldn't write file: %s", templateFileName)
	} else {
		logrus.Infof("Successfully removed references to .Values. in: %s on lines %d",
			templateFileName, lineNumbers)
	}
}

// Remove references to to ".Values." in all templates by replacing with an empty string.  Ansible Playbook can directly
// access defaults defined in defaults/main.yml.
func RemoveValuesReferencesInTemplates(roleDirectory string) {
	templatesDirectory := getAnsibleRoleTemplatesDirectory(roleDirectory)
	files, _ := readDir(templatesDirectory)
	for _, file := range files {
		templateFileName := filepath.Join(templatesDirectory, file.Name())
		removeValuesReferencesInTemplate(templateFileName)
	}
}