package reload

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"loggie-test/pkg/env"
	"loggie-test/pkg/resources/elasticsearch"
	"loggie-test/pkg/resources/k8s_loggie"
	"loggie-test/pkg/tools/k8s"
	"testing"
)

func TestChaos(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Log collection reload chaos test suite")
}

var _ = BeforeSuite(func() {
	k8s.InitCluster()
	Expect(env.Init().Setup(elasticsearch.Name, k8s_loggie.Name).Require(elasticsearch.Name, k8s_loggie.Name).Do()).To(Succeed())
})
