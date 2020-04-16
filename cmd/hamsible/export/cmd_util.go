package export

import (
	"fmt"
	"os"
	"strings"
)

//use this function for parsing arguments
func parse(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("please specify role name")
	}
	roleName = args[0]
	if len(roleName) == 0 {
		return fmt.Errorf("role name should not be empty")
	}

	return nil
}

func verifyFlags() error {
	if len(helmChartRef) == 0 {
		return fmt.Errorf("Please Specify Helm Chart path")
	}
	if strings.ContainsAny(helmChartRef, " ") {
		return fmt.Errorf("helm chart path contain spaces")
	}
	if _, err := os.Stat(helmChartRef); os.IsNotExist(err) {
		return fmt.Errorf("helm chart path doesn't exists")
	}

	return nil
}
