package export

import (
	"fmt"
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
		return fmt.Errorf("Helm Chart path Contain Spaces")
	}
	return nil
}
