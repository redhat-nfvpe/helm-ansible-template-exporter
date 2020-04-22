package helm

import (
	"errors"
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/chartutil"
	"reflect"
	"strings"
)

const goTemplateMemberAccessOperator = "."

// HelmChartRef is a global variable.  This is not ideal, but it is necessary in this case in order to avoid a circular
// dependency issue.  Essentially, the cmd package is originally responsible for determining the HelmChartRef via
// CLI interactions.  This package cannot depend on the "cmd" package, as it would cause a circular dependency.  The
// cmd package depends on our forked "text/template", and our "text/template" depends on the "helm" package.  Thus,
// "helm" cannot depend on "cmd".  This is a language limitation, and there may be more elegant ways to solve this in
// the future, but utilizing a global works for now, although it is ugly.  "cmd" is instructed to set this global
// variable for later use.
var HelmChartRef string

// A heuristic to determine whether a given input argument is likely a boolean value.  This is done through inspecting
// the charts Values.  For example, say we have an arbitrary argument in a conditional "metrics".  This function
// inspects the values file for the "metrics" definition.  If metrics looks like the following:
// metrics: true
// or
// metrics: false
// then metrics is likely a boolean (i.e., its value is a boolean value).  Otherwise, it is likely not a boolean.
func ArgIsLikelyBooleanYamlValue(arg string) (bool, error) {
	chartClient := NewChartClient()
	err := chartClient.LoadChartFrom(HelmChartRef)
	if err != nil {
		logrus.Warnf("error loading chart: %s", err)
	}

	raw, _ := chartutil.ReadValues([]byte(chartClient.Chart.Values.Raw))
	chartMap := raw.AsMap()
	pathArray := strings.Split(arg, goTemplateMemberAccessOperator)
	return IsBooleanYamlValue(&chartMap, &pathArray)
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
