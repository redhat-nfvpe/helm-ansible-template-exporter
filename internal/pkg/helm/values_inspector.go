package helm

import (
	"errors"
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/chartutil"
	"reflect"
	"regexp"
	"strings"
)

const goTemplateMemberAccessOperator = "."

//LogHelmReport - Collect Log data and print
type LogHelmReport struct {
	Name         string
	NewLine      string
	OriginalLine string
	FieldName    string
	Action       string
	LineNumber   int
}

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
	return isBooleanYamlValue(input, path, 0, len(*path)-1)
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
				return isBooleanYamlValue(&castedSubMap, path, pathIndex+1, indexCount)
			} else {
				return false, errors.New("invalid path")
			}
		} else {
			return false, errors.New("invalid path")
		}
	}
}


// GetValues takes a path that traverses a values that are stores in Values map and returns the value at the end of that path.
// Given the following data the value at path "chapter.one.title" is "PR Review".
//
//	chapter:
//	  one:
//	    title: "PR Review"
func GetValues(arg string) (*map[string][]*LogHelmReport, error) {
	argString := strings.ReplaceAll(arg, ".Values.", "")
	chartClient := NewChartClient()
	err := chartClient.LoadChartFrom(HelmChartRef)
	if err != nil {
		logrus.Warnf("error loading chart: %s", err)
		return nil, err
	}
	raw, _ := chartutil.ReadValues([]byte(chartClient.Chart.Values.Raw))
	result, err := raw.PathValue(argString)
	if err != nil {
		logrus.Warnf("Path value not found for path : %s ", argString)
		return nil, err
	}
	return dump(result), nil
}

// Converts and dumps the data of type interface{} slice
// returned by chartUtil.GetPathValues into *map[string]interface{}
func dump(result interface{}) *map[string][]*LogHelmReport {
	items := reflect.ValueOf(result)
	returnMap := make(map[string][]*LogHelmReport)
	if items.Kind() == reflect.Slice {
		for i := 0; i < items.Len(); i++ {
			item := items.Index(i)
			if item.Elem().Kind() == reflect.String {
				returnMap[item.Interface().(string)] = []*LogHelmReport{}
			} else if item.Elem().Kind() == reflect.Map {
				returnMaps := item.Interface().(map[string]interface{})
				for k := range returnMaps {
					returnMap[k] = []*LogHelmReport{}
				}
			}
		}
	}
	return &returnMap
}

// PreFixValuesWithItems...For single item list {% for item in hosts %}{{ item.name }}{% endfor %},
// This function will prefix the list variables with a loop variable, within the body of the for loop.
// There is two cases , one with field name that needs to be prefixed with loop variables
// other condition is dot field name is replaced by loop variable
func PreFixValuesWithItems(sb *strings.Builder, itemField string, pathValues *map[string][]*LogHelmReport) {
	input := sb.String()
	lines := strings.Split(string(input), "\n")
	lineNumbers := []int{}
	dotSuffixedField := " " + itemField + goTemplateMemberAccessOperator
	sb.Reset()
	for i, line := range lines {
		for vars := range *pathValues {
			if vars == goTemplateMemberAccessOperator { // if the replacing variable is dot then do this
				re, err := regexp.Compile(`\s\.\s`)
				if err == nil {
					matched := re.MatchString(line)
					if matched { //ignore line with template and includes
						if strings.Contains(line, "template") || strings.Contains(line, "include") {
							continue
						}
						lines[i] = re.ReplaceAllString(line, " "+itemField+" ")
						logrus.Debugf("Replaced item %s  to : %s on line %d", vars, itemField, i)
						r := new(LogHelmReport)
						r.Action = vars + "Replaced with " + itemField + "\n"
						r.NewLine = lines[i]
						r.OriginalLine = line
						r.FieldName = itemField
						report := (*pathValues)[goTemplateMemberAccessOperator]
						report = append(report, r)
						(*pathValues)[goTemplateMemberAccessOperator] = report
					}
				}
			} else {
				re, err := regexp.Compile("\\s." + vars + "\\s")
				if err == nil {
					matched := re.MatchString(line)
					if matched {
						lines[i] = re.ReplaceAllString(line, dotSuffixedField+vars+" ")
						logrus.Debugf("Appending item %s key to : %s on line %d", vars, dotSuffixedField+vars, i)
						r := new(LogHelmReport)
						r.Action = vars + " Replaced with " + itemField + "\n"
						r.NewLine = lines[i]
						r.OriginalLine = line
						r.FieldName = itemField
						report := (*pathValues)[goTemplateMemberAccessOperator]
						report = append(report, r)
						(*pathValues)[goTemplateMemberAccessOperator] = report
					}
				}
			}
			lineNumbers = append(lineNumbers, i)
		}
	}
	if len(lineNumbers) > 0 {
		logrus.Infof("Successfully updated references for range item field: %s on lines %d",
			itemField, lineNumbers)
	}
	output := strings.Join(lines, "\n")
	sb.WriteString(output)

}

// Remove $ from loop variables and body of the loop
func RemoveDollarPrefix(sb *strings.Builder) {
	input := sb.String()
	lines := strings.Split(string(input), "\n")
	lineNumbers := []int{}
	sb.Reset()
	re, err := regexp.Compile(`\$\b`)
	if err != nil {
		sb.WriteString(input)
		logrus.Errorf("Error parsing dollar sign %#v", err)
		return
	}
	for i, line := range lines {
		matched := re.MatchString(line)
		if matched {
			lines[i] = re.ReplaceAllString(line, "")
			logrus.Debugf("Removed `$`  prefix form variable %s on line %d", line, i)
			lineNumbers = append(lineNumbers, i)
		}
	}
	if len(lineNumbers) > 0 {
		logrus.Infof("Successfully cleaned `$` from range fields on lines %d", lineNumbers)
	}
	output := strings.Join(lines, "\n")
	sb.WriteString(output)
}

// LogReports - Printing report struct
func PrintReportItems(reports []*LogHelmReport) {
	for i, r := range reports {
		logrus.Info(i, ".Action: ", r.Action)
		logrus.Info(".Field Name: ", r.FieldName)
		logrus.Info(".Original Line: ", r.OriginalLine)
		logrus.Info(".New Line: ", r.NewLine)
	}
}
