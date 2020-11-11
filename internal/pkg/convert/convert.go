/*
Package helm provides utilities to aid in the conversion from a Helm chart into an Ansible Playbook Role.
 */
package convert

import (
	"bytes"
	"github.com/operator-framework/operator-sdk/pkg/ansible/paramconv"
	"github.com/pkg/errors"
	"github.com/redhat-nfvpe/helm-ansible-template-exporter/internal/pkg/helm"
	j2template "github.com/redhat-nfvpe/helm-ansible-template-exporter/internal/pkg/text/template"
	j2parse "github.com/redhat-nfvpe/helm-ansible-template-exporter/internal/pkg/text/template/parse"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/helm/pkg/chartutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	upstreamtemplate "text/template"
)

const ansibleRoleFilterPluginsDirectory = "filter_plugins"
const ansibleRoleDefaultsDirectory = "defaults"
const ansibleRoleTemplatesDirectory = "templates"
const ansibleRoleMainYamlFileName = "main.yml"
const ansibleRoleTasksDirectory = "tasks"
const ansibleTasksTemplateLeftDelimiter = "{{{"
const ansibleTasksTemplateLocation = "internal/pkg/helm/templates/tasks/main.yml"
const ansibleTasksTemplateRightDelimiter = "}}}"
const defaultDirectoryPermissions = 0777
const defaultPermissions = 0660
const filtersDirectory = "internal/filters"
const helmDefaultsContainsSelfReference =
	"# TODO: Replace \".Values.\" reference with a literal, as Ansible Playbook doesn't allow self-reference\n"
const HelmTemplatesDirectory = "templates"
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
	return filepath.Join(helmChartRootDirectory, HelmTemplatesDirectory)
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
	defaultsFileName := getAnsibleRoleDefaultsFileName(roleDirectory)
	j2parse.DefaultsFile = defaultsFileName
	ansibleRoleTemplatesDirectory := getAnsibleRoleTemplatesDirectory(roleDirectory)
	files, _ := readDir(ansibleRoleTemplatesDirectory)

	for _, file := range files {
		fileName := file.Name()
		templateFilePath := filepath.Join(ansibleRoleTemplatesDirectory, fileName)
		logrus.Infof("Attempting translation of branch nodes for: %s", templateFilePath)
		template, err := j2template.New(fileName).
			Option("missingkey=zero").
			Funcs(j2template.HelmFuncMap()).
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

func getFilterPluginsDirectory(roleDirectory string) string {
	return path.Join(roleDirectory, ansibleRoleFilterPluginsDirectory)
}

func createAnsibleFilterPluginsDirectory(roleDirectory string) error {
	filterPluinsDirectory := getFilterPluginsDirectory(roleDirectory)
	return os.Mkdir(filterPluinsDirectory, defaultDirectoryPermissions)
}

// Installs some default Ansible Filters into the generated workspace to aid in the Sprig transition.
func InstallAnsibleFilters(roleDirectory string) {
	err := createAnsibleFilterPluginsDirectory(roleDirectory)
	if err != nil {
		logrus.Warnf("Skipping Ansible Filter installation;  couldn't create %s in %s",
			ansibleRoleFilterPluginsDirectory, roleDirectory)
		return
	}
	filtersFiles, err := readDir(filtersDirectory)
	if err != nil {
		logrus.Warnf("Skipping Ansible Filter installation;  couldn't find: %s", filtersDirectory)
		return
	}
	for _, file := range filtersFiles {
		fileName := file.Name()
		filePath := path.Join(filtersDirectory, fileName)
		fileContents, err := ioutil.ReadFile(filePath)
		if err != nil {
			logrus.Warnf("Skipping Ansible Filter installation;  couldn't read: %s %s", filePath, err)
		} else {
			filterPluginsDirectory := getFilterPluginsDirectory(roleDirectory)
			destinationFileName := path.Join(filterPluginsDirectory, fileName)
			logrus.Infof("Attempting to write: %s", destinationFileName)
			err := ioutil.WriteFile(destinationFileName, fileContents, defaultPermissions)
			if err != nil {
				logrus.Warnf("Skipping writing %s due to %s", destinationFileName, err)
			} else {
				logrus.Infof("Successfully installed %s", destinationFileName)
			}
		}
	}
}

// Tail-recursively build up a map of Ansible keys in defaults/main.yaml that should be converted to snake_case.
func getTargetedReplacementKeysRecursive(input *map[string]interface{}, keys *map[string]string) {
	for key := range *input {
		snakeKey := paramconv.ToSnake(key)
		if snakeKey != key {
			(*keys)[key] = snakeKey
		}
		subMap := (*input)[key]
		if castedSubMap, ok := subMap.(map[string]interface{}); ok {
			getTargetedReplacementKeysRecursive(&castedSubMap, keys)
		}
	}
}

// Build up a map of Ansible keys in defaults/main.yaml that should be converted to snake_case.
func getTargetedReplacementKeys(chartClient *helm.HelmChartClient) *map[string]string {
	keys := map[string]string{}
	if chartClient.Chart.Values!=nil{
		raw, _ := chartutil.ReadValues([]byte(chartClient.Chart.Values.Raw))
		chartMap := raw.AsMap()
		getTargetedReplacementKeysRecursive(&chartMap, &keys)
	}
	return &keys
}

// Convert keys in defaults/main.yaml to snake case, and return a list of the substitutions for use in templates later.
func ConvertDefaultsToSnakeCase(chartClient *helm.HelmChartClient, roleDirectory string) *map[string]string {
	defaultsFile := getAnsibleRoleDefaultsFileName(roleDirectory)
	logrus.Infof("Attempting to convert keys to snake_case: %s", defaultsFile)
	input, err := ioutil.ReadFile(defaultsFile)
	if err != nil {
		logrus.Warnf("Skipping snake_case substitution, couldn't read file: %s", defaultsFile)
		return nil
	}

	lines := strings.Split(string(input), "\n")
	keysForConversion := getTargetedReplacementKeys(chartClient)

	for keyForConversion := range *keysForConversion {
		conversionLocations := []int{}
		snakeKey := paramconv.ToSnake(keyForConversion)
		for i, line := range lines {
			if strings.Contains(line, keyForConversion) {
				lines[i] = strings.ReplaceAll(lines[i], keyForConversion, snakeKey)
				conversionLocations = append(conversionLocations, i)
			}
		}
		logrus.Infof("converting defaults/main.yaml: %s -> %s lines: %d", keyForConversion, snakeKey,
			conversionLocations)
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(defaultsFile, []byte(output), defaultPermissions)
	if err != nil {
		logrus.Warnf("Skipping defaults/main.yaml snake_case conversion, couldn't write file: %s", defaultsFile)
	} else {
		logrus.Infof("Successfully performed snake_case conversion: %s", defaultsFile)
	}
	return keysForConversion
}
