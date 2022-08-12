package filecollection_test

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"loggie-test/pkg/env"
	"loggie-test/pkg/resources/elasticsearch"
	"loggie-test/pkg/resources/genfilesdeploy"
	"loggie-test/pkg/tools/logconfig"
	"time"
)

const (
	CaseCollectFileToES = "fileToES"

	FieldSearchESTimeout = "searchESTimeout"
)

var _ = Describe(fmt.Sprintf("#%s# collect log files and send data to elasticsearch", CaseCollectFileToES), Label("file collection"), func() {

	var esIns *elasticsearch.ES
	var searchESTimeout time.Duration
	var deploy *genfilesdeploy.GenFilesDeployment

	BeforeEach(func() {
		_, ei := env.GetResource(elasticsearch.Name)
		esIns = ei.(*elasticsearch.ES)

		_, dy := env.GetResource(genfilesdeploy.Name)
		deploy = dy.(*genfilesdeploy.GenFilesDeployment)

		caseCfg := env.GetCaseConfig(CaseCollectFileToES)
		esTimeoutConfig := caseCfg[FieldSearchESTimeout].(string)
		timeout, err := time.ParseDuration(esTimeoutConfig)
		Expect(err).ShouldNot(HaveOccurred())
		searchESTimeout = timeout
	})

	Context("collect n files", func() {

		It("should collect n files", func() {

			By("create genfiles deployment")
			Expect(deploy.Setup(context.Background())).Should(Succeed())
			Eventually(func() (bool, error) {
				return deploy.Ready()
			}).WithTimeout(30 * time.Second).WithPolling(1 * time.Second).Should(BeTrue())

			By("create logconfig")
			lgc := logconfig.GenLogConfigWithESSink(esIns, deploy, nil)
			Expect(logconfig.CreateLogConfig(lgc)).ShouldNot(HaveOccurred())

			By("get counts in elasticsearch index")
			// expect same count of events in elasticsearch
			Eventually(func() (int64, error) {
				return esIns.Count()
			}).WithTimeout(searchESTimeout).WithPolling(1 * time.Second).
				Should(Equal(deploy.AllCount()))

			By("delete logconfig")
			Expect(logconfig.DeleteLogConfig(lgc)).ShouldNot(HaveOccurred())

		})

	})

	AfterEach(func() {
		// delete deployment
		Expect(deploy.CleanUp(context.Background())).Should(Succeed())

		// delete elasticsearch index
		Expect(esIns.CleanUp(context.Background())).Should(Succeed())
	})

})
