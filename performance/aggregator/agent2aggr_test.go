package filecollection

import (
	"context"
	"fmt"
	logconfigv1beta1 "github.com/loggie-io/loggie/pkg/discovery/kubernetes/apis/loggie/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"loggie-test/pkg/env"
	"loggie-test/pkg/resources/elasticsearch"
	"loggie-test/pkg/resources/genfilesdeploy"
	"loggie-test/pkg/resources/k8s_loggie_aggre"
	"loggie-test/pkg/tools/logconfig"
	"time"
)

const (
	CaseAgentToAggre = "agent2aggr"

	FieldSearchESTimeout = "searchESTimeout"
)

var _ = Describe(fmt.Sprintf("#%s# agent -> aggregator -> es", CaseAgentToAggre), func() {

	var esIns *elasticsearch.ES
	var searchESTimeout time.Duration
	var deploy *genfilesdeploy.GenFilesDeployment
	var lgc *logconfigv1beta1.LogConfig
	var clgc *logconfigv1beta1.ClusterLogConfig
	var loggieAggreIns *k8s_loggie_aggre.Loggie

	BeforeEach(func() {
		_, ei := env.GetResource(elasticsearch.Name)
		esIns = ei.(*elasticsearch.ES)

		_, ag := env.GetResource(k8s_loggie_aggre.Name)
		loggieAggreIns = ag.(*k8s_loggie_aggre.Loggie)

		_, dy := env.GetResource(genfilesdeploy.Name)
		deploy = dy.(*genfilesdeploy.GenFilesDeployment)

		caseCfg := env.GetCaseConfig(CaseAgentToAggre)
		esTimeoutConfig := caseCfg[FieldSearchESTimeout].(string)
		timeout, err := time.ParseDuration(esTimeoutConfig)
		Expect(err).ShouldNot(HaveOccurred())
		searchESTimeout = timeout
	})

	Context("One agent send to one aggregator", func() {
		It("Should no log data lost", func() {
			By("create clusterLogconfig for loggie Aggregator")

			clusterlogconf := logconfig.GenClusterLogConfigWithGrpcSourceAndESSink(esIns)
			Expect(logconfig.CreateClusterLogConfig(clusterlogconf)).ShouldNot(HaveOccurred())
			clgc = clusterlogconf

			By("create logconfig for agent")
			logconf := logconfig.GenLogConfigWithGrpcSink(nil, loggieAggreIns, deploy)
			Expect(logconfig.CreateLogConfig(logconf)).ShouldNot(HaveOccurred())
			lgc = logconf

			collectLogs(esIns, deploy, searchESTimeout)
		})
	})

	AfterEach(func() {
		// delete deployment
		Expect(deploy.CleanUp(context.Background())).Should(Succeed())

		// delete logconfig
		Expect(logconfig.DeleteLogConfig(lgc)).ShouldNot(HaveOccurred())
		Expect(logconfig.DeleteClusterLogConfig(clgc)).ShouldNot(HaveOccurred())

		// delete elasticsearch index
		Expect(esIns.CleanUp(context.Background())).Should(Succeed())
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

