package filecollection_test

import (
	"context"
	"fmt"
	logconfigv1beta1 "github.com/loggie-io/loggie/pkg/discovery/kubernetes/apis/loggie/v1beta1"
	filesource "github.com/loggie-io/loggie/pkg/source/file"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"loggie-test/pkg/env"
	"loggie-test/pkg/report"
	"loggie-test/pkg/resources/elasticsearch"
	"loggie-test/pkg/resources/genfilesdeploy"
	"loggie-test/pkg/resources/k8s_loggie"
	"loggie-test/pkg/tools/logconfig"
	"time"
)

const (
	CaseCollectFileToES = "fileToES"

	FieldSearchESTimeout = "searchESTimeout"
)

var _ = Describe(fmt.Sprintf("#%s# file source to elasticsearch sink", CaseCollectFileToES), Label("file collection"), func() {

	var esIns *elasticsearch.ES
	var searchESTimeout time.Duration
	var deploy *genfilesdeploy.GenFilesDeployment
	var lgc *logconfigv1beta1.LogConfig
	var loggieIns *k8s_loggie.Loggie

	stopMetrics := make(chan struct{})

	BeforeEach(func() {
		_, ei := env.GetResource(elasticsearch.Name)
		esIns = ei.(*elasticsearch.ES)

		_, li := env.GetResource(k8s_loggie.Name)
		loggieIns = li.(*k8s_loggie.Loggie)

		_, dy := env.GetResource(genfilesdeploy.Name)
		deploy = dy.(*genfilesdeploy.GenFilesDeployment)

		caseCfg := env.GetCaseConfig(CaseCollectFileToES)
		esTimeoutConfig := caseCfg[FieldSearchESTimeout].(string)
		timeout, err := time.ParseDuration(esTimeoutConfig)
		Expect(err).ShouldNot(HaveOccurred())
		searchESTimeout = timeout

	})

	AfterEach(func() {
		stopMetrics <- struct{}{}

		// delete logconfig
		Expect(logconfig.DeleteLogConfig(lgc)).ShouldNot(HaveOccurred())

		// delete deployment
		Expect(deploy.CleanUp(context.Background())).Should(Succeed())
		//Eventually(func() error {
		//	_, err := deploy.GetDeployment()
		//	return err
		//}).WithTimeout(10*time.Second).WithPolling(1*time.Second).Should(HaveOccurred())

		// delete elasticsearch index
		Expect(esIns.CleanUp(context.Background())).Should(Succeed())
	})

	Context("FileCount=10, FileSize=100MB (LineCount=100000, LineBytes=1KiB)", func() {
		It("FileCount=10", func() {
			By("create logconfig")
			logconf := logconfig.GenLogConfigWithESSink(esIns, deploy, nil)
			Expect(logconfig.CreateLogConfig(logconf)).ShouldNot(HaveOccurred())
			lgc = logconf

			By("record metrics")
			go report.RecordMetrics(stopMetrics, loggieIns, "FileCount10")

			deploy.Conf.FileCount = 10
			deploy.Conf.LineCount = 100000
			deploy.Conf.LineBytes = 1024
			collectLogs(esIns, deploy, searchESTimeout)
		})
	})

	Context("FileCount=20, FileSize=100MB (LineCount=100000, lineBytes=1KiB)", func() {
		It("FileCount=20", func() {
			By("create logconfig")
			logconf := logconfig.GenLogConfigWithESSink(esIns, deploy, nil)
			Expect(logconfig.CreateLogConfig(logconf)).ShouldNot(HaveOccurred())
			lgc = logconf

			By("record metrics")
			go report.RecordMetrics(stopMetrics, loggieIns, "FileCount20")

			deploy.Conf.FileCount = 20
			deploy.Conf.LineCount = 100000
			deploy.Conf.LineBytes = 1024
			collectLogs(esIns, deploy, searchESTimeout)
		})
	})

	Context("FileCount=20, addonMeta=true", func() {
		It("addonMeta=true", func() {
			By("create logconfig")
			logconf := logconfig.GenLogConfigWithESSink(esIns, deploy, &filesource.Config{
				CollectConfig: filesource.CollectConfig{
					AddonMeta: true,
				},
			})
			Expect(logconfig.CreateLogConfig(logconf)).ShouldNot(HaveOccurred())
			lgc = logconf

			By("record metrics")
			go report.RecordMetrics(stopMetrics, loggieIns, "addonMeta")

			deploy.Conf.FileCount = 20
			deploy.Conf.LineCount = 100000
			deploy.Conf.LineBytes = 1024
			collectLogs(esIns, deploy, searchESTimeout)
		})
	})

})

func collectLogs(esIns *elasticsearch.ES, deploy *genfilesdeploy.GenFilesDeployment, searchESTimeout time.Duration) {
	By("create genfiles deployment")
	Expect(deploy.Setup(context.Background())).Should(Succeed())
	Eventually(func() (bool, error) {
		return deploy.Ready()
	}).WithTimeout(30 * time.Second).WithPolling(1 * time.Second).Should(BeTrue())

	By("get counts in elasticsearch index")
	// expect same count of events in elasticsearch
	Eventually(func() (int64, error) {
		return esIns.Count()
	}).WithTimeout(searchESTimeout).WithPolling(1 * time.Second).
		Should(Equal(deploy.AllCount()))
}
