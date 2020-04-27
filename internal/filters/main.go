/*
A CLI->Go template shim used to render arbitrary sprig functions.

To Run:
> go run main.go sprigFunctionName sprigFunctionArguments...

Example:
> go run main.go trim "      some string      "

This program takes textual input and attempts to resolve the sprigFunctionArguments types through inspecting
sprigFunctionName's signature.  Bad input results in fatal exit with error code 1.  Debug output is intentionally
suppressed as stdout is utilized by the calling program.
*/

package main

import (
	"errors"
	"fmt"
	"github.com/Masterminds/sprig"
	"os"
	"reflect"
	"strconv"
)

const fatalExitReturnCode = 1
const funcArgsCLIArgsFirstIndex = 2
const funcNameCLIArgsIndex = 1
const funcReturnOutputIndex = 0

// Wrapper to call a Golang function through reflection.
func CallFunc(funcMap map[string]interface{}, name string, arguments ... interface{}) (result []reflect.Value,
	err error) {

	function := reflect.ValueOf(funcMap[name])
	if len(arguments) != function.Type().NumIn() {
		err = errors.New("number of arguments is not adapted")
		return
	}
	in := make([]reflect.Value, len(arguments))
	for k, param := range arguments {
		in[k] = reflect.ValueOf(param)
	}
	result = function.Call(in)
	return
}

// Exits 1 with an appropriate message.
func fatalError(message string) {
	fmt.Fprintln(os.Stderr, "ERROR: " + message)
	os.Exit(fatalExitReturnCode)
}

// Ensures that a sprig/template function was provided as a CLI argument.
func checkForFuncName(args *[]string) {
	if len(*args) < 2 {
		fatalError("No Sprig Function name was provided")
	}
}

// Extract a function given a funcMap.
func extractFunction(funcMap map[string]interface{}, funcName string) (reflect.Value, error) {
	functionInterface := funcMap[funcName]
	if functionInterface == nil || reflect.TypeOf(functionInterface).Kind() != reflect.Func {
		return reflect.Value{}, errors.New(funcName + " is not a valid function")
	}
	function := reflect.ValueOf(functionInterface)
	return function, nil
}



// Invokes the CLI shim.  The first argument is the sprig/template function name.  The remaining arguments are the
// sprig/template function arguments.
func main() {
	args := os.Args
	// This would indicate a function with no funcName argument.
	checkForFuncName(&args)

	funcName := args[funcNameCLIArgsIndex]
	funcArgs := args[funcArgsCLIArgsFirstIndex:]

	// Loads only the text/template and Sprig Function Map.
	funcMap := sprig.TxtFuncMap()

	// Used to convert CLI args into functional arguments.
	var iFuncArgs []interface{}
	function, err := extractFunction(funcMap, funcName)

	if err != nil {
		fatalError(err.Error())
	}

	var argType = reflect.String
	for i, b := range funcArgs {
		// Attempt to derive the argument type from the method signature, and cast/convert the input appropriately.
		if i < function.Type().NumIn() {
			argType = function.Type().In(i).Kind()
		}
		// TODO: We must address other types of input when possible.
		if argType == reflect.Int {
			c, _ := strconv.Atoi(b)
			iFuncArgs = append(iFuncArgs, c)
		} else if argType == reflect.Uint32 {
			c, _ := strconv.Atoi(b)
			iFuncArgs = append(iFuncArgs, uint32(c))
		} else {
			iFuncArgs = append(iFuncArgs, b)
		}
	}

	out, err := CallFunc(funcMap, funcName, iFuncArgs...)
	if err != nil {
		fatalError(err.Error())
	}
	// Note:  this is the only emission to stdout of the entire program
	fmt.Print(out[funcReturnOutputIndex])
}
