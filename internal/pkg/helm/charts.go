package helm

import (
	"fmt"
	"github.com/redhat-nfvpe/helm-ansible-template-exporter/internal/pathconfig"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"os"
)

//HelmChartClient ....
type HelmChartClient struct {
	Chart      *chart.Chart
	ChartName  string
	PathConfig *pathconfig.PathConfig
}

//NewChartClient creates a new chart client
func NewChartClient() *HelmChartClient {
	client := HelmChartClient{}
	client.PathConfig, _ = GetBasePathConfig()
	return &client
}

//LoadChart uses the chart client's values to retrieve  the appropriate chart
func (hc *HelmChartClient) LoadChartFrom(chartPath string) (err error) {
	loadedChart, err := chartutil.Load(chartPath)
	if err != nil {
		return err
	}
	hc.Chart = loadedChart
	hc.ChartName = hc.Chart.Metadata.GetName()
	return

}

// GetBasePathConfig ....
func GetBasePathConfig() (*pathconfig.PathConfig, error) {
	// get the current directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting working directory")
	}
	basePath := cwd
	pathConfig := pathconfig.NewConfig(basePath)
	return pathConfig, nil
}
