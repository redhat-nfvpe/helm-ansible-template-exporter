package parse

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"strings"
)

const defaultPermissions = 0660
const keyValueSeparator = ":"

var DefaultsFile string
var ReplaceWithSnakeCase bool

// Substitute field with snakeField within defaults/main.yaml.
func SubstituteSnakeCaseDefaultValue(field, snakeField string) {
	// Must live in text/template as including in convert package would cause a circular dependency.
	input, err := ioutil.ReadFile(DefaultsFile)
	if err != nil {
		logrus.Warnf("Skipping snake case substitution. Couldn't read file: %s", DefaultsFile)
		return
	}
	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		if strings.Contains(line, field) {
			logrus.Infof("Converting default value to snake case line %d: %s -> %s", i, field, snakeField)
			lines[i] = strings.ReplaceAll(lines[i], field + keyValueSeparator, snakeField + keyValueSeparator)
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(DefaultsFile, []byte(output), defaultPermissions)
	if err != nil {
		logrus.Warnf("Skipping snake case substitution. Couldn't write file: %s", DefaultsFile)
	} else {
		logrus.Infof("Successfully converted default values for %s", DefaultsFile)
	}
}