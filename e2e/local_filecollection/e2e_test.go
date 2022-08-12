package filecollection_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"loggie-test/pkg/env"
	"loggie-test/pkg/resources/elasticsearch"
	"loggie-test/pkg/resources/local_loggie"
	"testing"
)

func TestLoggie(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Loggie local log collection E2E Suite")
}

var _ = BeforeSuite(func() {
	Expect(env.Init().Setup(elasticsearch.Name, local_loggie.Name).Require(elasticsearch.Name, local_loggie.Name).Do()).To(Succeed())
})
