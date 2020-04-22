package helm_test

import (
	"errors"
	"github.com/redhat-nfvpe/helm-ansible-template-exporter/internal/pkg/helm"
	"testing"
)

type execTest struct {
	name   string
	input  map[string]interface{}
	path   []string
	output bool
	err    error
}

var complicatedMap = map[string]interface{} {
	"image": map[string]interface{} {
		"registry": "docker.io",
		"repository": "bitnami/nginx",
		"tag": "1.17.9-debian-10-r0",
		"imagePullPolicy": "IfNotPresent",
		"level3Nesting": map[string]interface{}{
			"level3Key": "level3Val",
			"level3BooleanKeyTrue": true,
			"level3BooleanKeyFalse": false,
		},
	},
	"replicaCount": 1,
	"podAnnotations": map[string]interface{} {},
	"metrics": map[string]interface{} {
		"enabled": true,
		"disabled": false,
		"trueString": "true",
		"falseString": "false",
	},
}

var execTests = []execTest{
	// Negative Tests.  These tests espouse bad input criteria, and are all expected to return error(s) of some form.

	// 1. Test nil input map.  Since the map is nil, an error should be raised stating that "input cannot be nil".
	{
		"nil-input",
		nil,
		[]string{"path"},
		false,
		errors.New("input cannot be nil"),
	},

	// 2. Test nil path.  A path is expected to be non-zero-length slice, and thus nil should result in an error.
	{
		"nil-path",
		map[string]interface{} {
			"key": "val",
		},
		nil,
		false,
		errors.New("path slice must have at least one element"),
	},

	// 3. Test a zero-length path.  A path is expected to be non-zero-length slice, and thus a zero-length path should
	// result in an error.
	{
		"zero-length-path",
		map[string]interface{} {
			"key": "val",
		},
		[]string{},
		false,
		errors.New("path slice must have at least one element"),
	},

	// Positive Tests for a non-nested input map.

	// 1. A simple single element key/value where value (someImageValue) is a string, so IsBooleanYamlValue should
	// return false.
	{
		"simple-key-value--value-is-string",
		map[string]interface{} {
			"image": "someImageValue",
		},
		[]string{"image"},
		false,
		nil,
	},
	// 2. A simple single element key/value where value (nil) is not a boolean, so IsBooleanYamlValue should return
	// false.
	{
		"simple-key-value--value-is-empty",
		map[string]interface{} {
			"image": nil,
		},
		[]string{"image"},
		false,
		nil,
	},
	// 3. A simple single element key/value where value (false) is a boolean, so IsBooleanYamlValue should return true.
	{
		"simple-key-value--value-is-false",
		map[string]interface{} {
			"image": false,
		},
		[]string{"image"},
		true,
		nil,
	},
    // 4. A simple single element key/value where value (true) is a boolean, so IsBooleanYamlValue should return true.
	{
		"simple-key-value--value-is-true",
		map[string]interface{} {
			"image": true,
		},
		[]string{"image"},
		true,
		nil,
	},

	// Tests for a more complicated input map/path.
	// 1. Try to determine a level 2 nested evaluation for a non-boolean.
	{
		"level2nesting",
		complicatedMap,
		[]string{"image", "registry"},
		false,
		nil,
	},
	// 2. Try to determine a level 3 nested evaluation for a non-boolean.
	{
		"level3nesting",
		complicatedMap,
		[]string{"image", "level3Nesting", "level3Key"},
		false,
		nil,
	},
	// 3. An intermediary nesting level does not exist;  should return an error.
	{
		"intermediary-level-does-not-exist",
		complicatedMap,
		[]string{"image", "dne", "level3Key"},
		false,
		errors.New("invalid path"),
	},
	// 4. A tertiary level is a true boolean
	{
		"tertiary-level-is-true-bool",
		complicatedMap,
		[]string{"image", "level3Nesting", "level3BooleanKeyTrue"},
		true,
		nil,
	},
	// 5. A tertiary level is a false boolean
	{
		"tertiary-level-is-false-bool",
		complicatedMap,
		[]string{"image", "level3Nesting", "level3BooleanKeyFalse"},
		true,
		nil,
	},
	// 6. A tertiary level doesn't even exist
	{
		"tertiary-level-doesnt-exist",
		complicatedMap,
		[]string{"image", "level3Nesting", "dne"},
		false,
		errors.New("invalid path"),
	},
	// 7. "true" string is not treated as a bool
	{
		"true-string-shouldnt-be-treated-as-bool",
		complicatedMap,
		[]string{"metrics", "trueString"},
		false,
		nil,
	},
	// 8. "false" string is not treated as a bool
	{
		"false-string-shouldnt-be-treated-as-bool",
		complicatedMap,
		[]string{"metrics", "falseString"},
		false,
		nil,
	},
	// 9. but "metrics.enabled" on the same level should be treated as bool
	{
		"metrics.enabled-treated-as-bool",
		complicatedMap,
		[]string{"metrics", "enabled"},
		true,
		nil,
	},
	// 10. as well as "metrics.disabled" on the same level (should be treated as bool)
	{
		"metrics.disabled-treated-as-bool",
		complicatedMap,
		[]string{"metrics", "disabled"},
		true,
		nil,
	},

}

func TestIsBooleanYamlValue(t *testing.T) {
	for _, test := range execTests {
		isBool, err := helm.IsBooleanYamlValue(&test.input, &test.path)
		if test.err != nil {
			if test.err.Error() != err.Error() {
				t.Errorf("Error(%s):  Expected: %s Actual: %s", test.name, test.err, err)
			}
		} else {
			if test.output != isBool {
				t.Errorf("Error(%s):  Expected: %t Actual: %t", test.name, test.output, isBool)
			}
		}
	}
}