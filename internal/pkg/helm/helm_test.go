package helm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-nfvpe/helm-ansible-template-exporter/internal/pkg/helm"
	"k8s.io/helm/pkg/chartutil"
)

var _ = Describe("Helm API", func() {
	Context("Load and read helm charts", func() {
		It("Return templates and values populated", func() {
			chartClient := helm.NewChartClient()
			err := chartClient.LoadChartFrom("../../../examples/helmcharts/nginx")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(chartClient.Chart.Templates)).NotTo(Equal(0))
			raw, _ := chartutil.ReadValues([]byte(chartClient.Chart.Values.Raw))
			image, _ := raw["image"].(map[string]interface{})
			Expect(image["pullPolicy"]).To(Equal("IfNotPresent"))
		})
	})

})
