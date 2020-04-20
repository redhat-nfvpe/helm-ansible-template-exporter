/*
Package helm provides utilities to aid in the conversion from a Helm chart into an Ansible Playbook Role.
 */
package helm

import (
	"bytes"
	"github.com/pkg/errors"
	j2template "github.com/redhat-nfvpe/helm-ansible-template-exporter/internal/pkg/text/template"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	upstreamtemplate "text/template"
)

const ansibleRoleDefaultsDirectory = "defaults"
const ansibleRoleTemplatesDirectory = "templates"
const ansibleRoleMainYamlFileName = "main.yml"
const ansibleRoleTasksDirectory = "tasks"
const ansibleTasksTemplateLeftDelimiter = "{{{"
const ansibleTasksTemplateLocation = "internal/pkg/helm/templates/tasks/main.yml"
const ansibleTasksTemplateRightDelimiter = "}}}"
const defaultPermissions = 0600
const goTemplateMemberAccessOperator = "."
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

// Given an Ansible Role directory, return the path to the defaults main.yml file.  This does not check the
// existence or permissions of the underlying file.
func getAnsibleRoleDefaultsFileName(roleDirectory string) string {
	return filepath.Join(getAnsibleRoleDefaultsDirectory(roleDirectory), ansibleRoleMainYamlFileName)
}

// Given an Ansible Role directory, return the path to the tasks directory.  This dos not check for the existence or
// readability of the underlying directory.
func getAnsibleRoleTasksDirectory(roleDirectory string) string {
	return filepath.Join(roleDirectory, ansibleRoleTasksDirectory)
}

// Given an Ansible Role directory, return the path to the tasks main.yml file.  This does not check the existence
// or permissions of the underlying file.
func getAnsibleRoleTasksMainFileName(roleDirectory string) string {
	return filepath.Join(getAnsibleRoleTasksDirectory(roleDirectory), ansibleRoleMainYamlFileName)
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

// Removes Whitespace Trimming calls "{{-" and "-}}" and replaces them with "{{" and "}}" respectively in a given
// template.  This is due to the fact that the Go text/template lexer is destructive, and ends up eating this
// white-space.
func suppressWhitespaceTrimmingInTemplate(templateFileName string) {
	input, err := ioutil.ReadFile(templateFileName)
	if err != nil {
		logrus.Warnf("Skipping whitespace surpression, couldn't read file: %s", templateFileName)
		return
	}

	lines := strings.Split(string(input), "\n")

	lineNumbers := []int{}

	for i, line := range lines {
		if strings.Contains(line, "{{-") || strings.Contains(line, "-}}") {
			logrus.Debugf("Suppressing whitespace in: %s on line %d", templateFileName, i)
			lines[i] = strings.ReplaceAll(lines[i], "{{-", "{{")
			lines[i] = strings.ReplaceAll(lines[i], "-}}", "}}")
			lineNumbers = append(lineNumbers, i)
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(templateFileName, []byte(output), defaultPermissions)
	if err != nil {
		logrus.Warnf("Skipping whitespace suppression couldn't write file: %s", templateFileName)
	} else {
		logrus.Infof("Successfully suppressed whitespace in: %s on lines %d",
			templateFileName, lineNumbers)
	}
}

// Removes Whitespace Trimming calls "{{-" and "-}}" and replaces them with "{{" and "}}" respectively for all templates
// in an Ansible Role.  This is due to the fact that the Go text/template lexer is destructive, and ends up eating this
// white-space.
func SuppressWhitespaceTrimmingInTemplates(roleDirectory string) {
	templatesDirectory := getAnsibleRoleTemplatesDirectory(roleDirectory)
	logrus.Infof("templates dir: %s", templatesDirectory)
	files, _ := readDir(templatesDirectory)
	for _, file := range files {
		templateFileName := filepath.Join(templatesDirectory, file.Name())
		suppressWhitespaceTrimmingInTemplate(templateFileName)
	}
}

// Invokes a custom text/template implementation in order to convert possibly-nested Branch Nodes into the Ansible
// counterparts.  For example, the following Golang template:
//   {{ if conditional }}
//   ...
//   {{ end }}
// Becomes:
//   {% if conditional %}
//   ...
//   {% endif %}
func ConvertControlFlowSyntax(roleDirectory string) {
	ansibleRoleTemplatesDirectory := getAnsibleRoleTemplatesDirectory(roleDirectory)
	files, _ := readDir(ansibleRoleTemplatesDirectory)

	for _, file := range files {
		fileName := file.Name()
		templateFilePath := filepath.Join(ansibleRoleTemplatesDirectory, fileName)
		logrus.Infof("Attempting translation of branch nodes for: %s", templateFilePath)
		template, err := j2template.New(fileName).
			Option("missingkey=zero").
			Funcs(HelmFuncMap()).
			ParseFiles(templateFilePath)
		if err != nil {
			logrus.Fatalf("Couldn't instantiate the Go Template engine %s", err)
		}
		err = ioutil.WriteFile(templateFilePath, []byte(template.Tree.Root.String()), defaultPermissions)
		if err != nil {
			logrus.Warnf("Skipping translation of branch nodes couldn't write file: %s", templateFilePath)
		} else {
			logrus.Infof("Successfully translated branch nodes in: %s",
				templateFilePath)
		}
	}
}

// Installs the Ansible Playbook Role task responsible for invoking the translated templates.
func InstallAnsibleTasks(roleDirectory string) {
	ansibleRoleTemplatesDirectory := getAnsibleRoleTemplatesDirectory(roleDirectory)
	files, _ := readDir(ansibleRoleTemplatesDirectory)

	// Generate a list of filenames to toss in the Ansible Playbook Role tasks/main.yml
	var fileNames []string
	for _, file := range files {
		fileName := file.Name()
		fileNames = append(fileNames, fileName)
	}

	// Custom delimiters are used in this template since ansible uses "{{" and "}}" as well.
	template, err := upstreamtemplate.New(ansibleRoleMainYamlFileName).
		Delims(ansibleTasksTemplateLeftDelimiter, ansibleTasksTemplateRightDelimiter).
		ParseFiles(ansibleTasksTemplateLocation)

	// This should not happen, as no user input has been included yet.  However, to be safe, we check anyway.
	if err != nil {
		logrus.Fatal(err)
	}

	buf := &bytes.Buffer{}
	err = template.Execute(buf, fileNames)
	if err != nil {
		logrus.Warnf("Couldn't generate the tasks main.yml file: %s", err)
	}

	destinationTasksYamlFile := getAnsibleRoleTasksMainFileName(roleDirectory)
	err = ioutil.WriteFile(destinationTasksYamlFile, buf.Bytes(), defaultPermissions)
	if err != nil {
		logrus.Warnf("Skipping creating/installing Ansible Tasks, couldn't write file: %s",
			destinationTasksYamlFile)
	} else {
		logrus.Infof("Successfully created/installed Ansible Tasks: %s",
			destinationTasksYamlFile)
	}
}


// Determines whether path within the input context refers to a boolean type.  Consulting a Helm Chart's values is
// helpful for determining whether a Helm Chart template conditional is checking for boolean equality v.s. definition.
func IsBooleanYamlValue(input *map[string]interface{}, path *[]string) (bool, error) {
	if input == nil || *input == nil {
		return false, errors.New("input cannot be nil")
	}
	if path == nil || *path == nil || len(*path) < 1 {
		return false, errors.New("path slice must have at least one element")
	}
	// The recursive stopping condition is triggered when the pathIndex is equal to the indexCount (number of subpaths)
	return isBooleanYamlValue(input, path, 0, len(*path) - 1)
}

// Internal recursive helper function that determines whether path within the input context refers to a boolean type.
func isBooleanYamlValue(input *map[string]interface{}, path *[]string, pathIndex int, indexCount int) (bool, error) {
	pathKey := (*path)[pathIndex]

	// Recursive base case.  When we have reached the last sub-path within a path (i.e., "pullPolicy" for
	// path=["metrics", "image", "pullPolicy"]
	if pathIndex == indexCount {
		// Handles the case in which the last pathKey does not exist.  For example, if "pullPolicy" wasn't valid.
		if value, ok := (*input)[pathKey]; ok {
			// Handles invalid YAML such as "pullPolicy:";  provides a specific hint for the invalid YAML.
			if value != nil {
				if reflect.TypeOf(value).Kind() == reflect.Bool {
					return true, nil
				} else {
					return false, nil
				}
			} else {
				return false, errors.New("invalid path;  \"" + pathKey + "\" has no value in input YAML")
			}
		} else {
			return false, errors.New("invalid path")
		}
	} else {
		// Recursive block;  checks the intermediary path and then recurse.
		if subMap, ok := (*input)[pathKey]; ok {
			// Handles the case in which an intermediary pathKey does not exist.  I.e., "metrics.doesntexist.lastkey".
			if subMap != nil {
				castedSubMap := subMap.(map[string]interface{})
				return isBooleanYamlValue(&castedSubMap, path, pathIndex + 1, indexCount)
			} else {
				return false, errors.New("invalid path")
			}
		} else {
			return false, errors.New("invalid path")
		}
	}
}

