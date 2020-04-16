package export_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cmd "github.com/redhat-nfvpe/helm-ansible-template-exporter/cmd/hamsible/export"
)

var _ = Describe("Export", func() {
	Context("When role is not passed", func() {
		It("Should thrown an error, expecting role", func() {
			args := []string{}
			exportCmd := cmd.GetExportCmd()
			exportCmd.SetArgs(args)
			err := exportCmd.Execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("please specify role name"))
		})
	})
	Context("When a flag is passed without `role` argument ", func() {
		It("Should thrown an error, expecting role", func() {
			args := []string{"--workspace=workspace"}
			exportCmd := cmd.GetExportCmd()
			exportCmd.SetArgs(args)
			err := exportCmd.Execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("please specify role name"))
		})
	})

	Context("When role argument is passed without any flag", func() {
		It("Should ask for helm chart path", func() {
			args := []string{"test"}
			exportCmd := cmd.GetExportCmd()
			exportCmd.SetArgs(args)
			err := exportCmd.Execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("please specify helm chart path"))
		})
	})

	Context("When bad helm chart path is passed", func() {
		It("Should return a bad path error", func() {
			args := []string{"test", "--helm-chart=./temp"}
			exportCmd := cmd.GetExportCmd()
			exportCmd.SetArgs(args)
			err := exportCmd.Execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("helm chart path doesn't exists"))
		})
	})

})
