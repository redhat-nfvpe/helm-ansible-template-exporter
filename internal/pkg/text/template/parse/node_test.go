package parse_test

import (
	"github.com/redhat-nfvpe/helm-ansible-template-exporter/internal/pkg/helm"
	template2 "github.com/redhat-nfvpe/helm-ansible-template-exporter/internal/pkg/text/template"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
)

const helmTemplatesDirectory = "templates"

// Represents the meta-information for a contrived chart test case.
type testCase struct {
	name     string
	chartDir string
}

var testCases = []testCase{
	{
		"basicconditionalchart_bool_case",
		"testdata/basicconditionalchart_bool_case",
	},
	{
		"basicconditionalchart_definition_case",
		"testdata/basicconditionalchart_definition_case",
	},
	{
		"nested_conditional",
		"testdata/nested_conditional",
	},
	{
		"nested_conditional_with_if_definition",
		"testdata/nested_conditional_with_if_definition",
	},
	{
		"basic_sprig",
		"testdata/basic_sprig",
	},
	{
		"basic_with",
		"testdata/basic_with",
	},
}

// Reads the files in a directory, exiting fatally if any errors occur.
func readDir(directory string, t *testing.T) ([]os.FileInfo, error) {
	logrus.Debugf("Attempting to read: %s", directory)
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	logrus.Debugf("Successfully read: %s", directory)
	return files, err
}

func TestToString(t *testing.T) {
	for _, testCase := range testCases {
		logrus.Infof("Running: %s", testCase.name)
		templatesDirectory := path.Join(testCase.chartDir, helmTemplatesDirectory)
		templateFiles, err := readDir(templatesDirectory, t)
		if err != nil {
			t.Errorf("Couldn't read: %s", templatesDirectory)
		}
		for _, file := range templateFiles {
			testFileName := file.Name()
			cwd, _ := os.Getwd()
			helm.HelmChartRef = path.Join(cwd, testCase.chartDir)
			templateFilePath := path.Join(templatesDirectory, testFileName)
			template, err := template2.New(testFileName).
				Option("missingkey=zero").
				Funcs(template2.HelmFuncMap()).
				ParseFiles(templateFilePath)
			if err != nil {
				t.Errorf("Unexpected error while parsing %s: %s", testFileName, err)
			}
			expectedFileName := path.Join(testCase.chartDir, testFileName+".j2")
			expectedByte, err := ioutil.ReadFile(expectedFileName)
			if err != nil {
				t.Errorf("Could not load expected file: %s", expectedFileName)
			}
			expected := string(expectedByte)
			expected = strings.TrimSpace(expected)
			actual := strings.TrimSpace(template.Root.String())
			if expected != actual {
				t.Errorf("Parsing error.  Expected=%s Actual=%s", expected, actual)
			}
		}
	}
}
