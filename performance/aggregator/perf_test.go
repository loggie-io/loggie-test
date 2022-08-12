package filecollection_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"loggie-test/pkg/env"
	"loggie-test/pkg/resources/elasticsearch"
	"loggie-test/pkg/resources/k8s_loggie_aggre"
	"loggie-test/pkg/tools/k8s"
	"testing"
)

func TestAgentToAggregator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Agent to Aggregator test suite")
}

var _ = BeforeSuite(func() {
	k8s.InitCluster()
	Expect(env.Init().Setup(elasticsearch.Name, k8s_loggie_aggre.Name).
		Require(elasticsearch.Name, k8s_loggie_aggre.Name).Do()).To(Succeed())
})
