package filecollection_test

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"loggie-test/pkg/env"
	"loggie-test/pkg/resources/elasticsearch"
	"loggie-test/pkg/resources/genfiles"
	"loggie-test/pkg/resources/local_loggie"
	"time"
)

const (
	CaseCollectFileToES = "fileToES"

	FieldPipelines       = "pipelines"
	FieldSearchESTimeout = "searchESTimeout"
)

var _ = Describe(fmt.Sprintf("#%s# collect log files and send data to elasticsearch", CaseCollectFileToES), Label("file collection"), func() {

	var esIns *elasticsearch.ES
	var loggieIns *local_loggie.Loggie
	var files *genfiles.GenFiles
	var pipelineConfig string
	var searchESTimeout time.Duration

	BeforeEach(func() {
		_, ei := env.GetResource(elasticsearch.Name)
		esIns = ei.(*elasticsearch.ES)

		_, li := env.GetResource(local_loggie.Name)
		loggieIns = li.(*local_loggie.Loggie)

		caseCfg := env.GetCaseConfig(CaseCollectFileToES)
		pipelineConfig = caseCfg[FieldPipelines].(string)
		esTimeoutConfig := caseCfg[FieldSearchESTimeout].(string)
		timeout, err := time.ParseDuration(esTimeoutConfig)
		Expect(err).ShouldNot(HaveOccurred())
		searchESTimeout = timeout
	})

	Context("collect n files", func() {

		It("should collect n files", func() {

			By("make pipelines.yml")
			Expect(loggieIns.UpdatePipelineAndWaitForReloadPeriod(pipelineConfig)).Should(Succeed())

			By("generate log files")
			fileRes, fileIns := env.GetResource(genfiles.Name)
			Expect(fileRes.Setup(context.Background())).Should(Succeed())
			files = fileIns.(*genfiles.GenFiles)

			By("get counts in elasticsearch index")
			// expect same count of events in elasticsearch
			Eventually(func() (int64, error) {
				return esIns.Count()
			}).WithTimeout(searchESTimeout).WithPolling(1 * time.Second).
				Should(Equal(files.AllCount()))
		})

	})

	AfterEach(func() {
		// clean log files
		Expect(files.CleanUp(context.Background())).Should(Succeed())

		// delete elasticsearch index
		Expect(esIns.CleanUp(context.Background())).Should(Succeed())
	})

})
