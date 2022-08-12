package filecollection_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"loggie-test/pkg/env"
	"loggie-test/pkg/resources/elasticsearch"
	"testing"
)

func TestK8sFileCollection(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "K8s log collection E2E Suite")
}

var _ = BeforeSuite(func() {
	Expect(env.Init().Setup(elasticsearch.Name).Require(elasticsearch.Name).Do()).To(Succeed())
})
