package reload

import (
	"context"
	"fmt"
	logconfigv1beta1 "github.com/loggie-io/loggie/pkg/discovery/kubernetes/apis/loggie/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"loggie-test/pkg/env"
	"loggie-test/pkg/resources/elasticsearch"
	"loggie-test/pkg/resources/genfilesdeploy"
	"loggie-test/pkg/resources/k8s_loggie"
	"loggie-test/pkg/tools/logconfig"
	"math/rand"
	"time"
)

const (
	CaseAlwaysReload = "alwaysReload"

	FieldsInterval  = "interval"
	FieldsTotalTime = "totalTime"
)

var _ = Describe(fmt.Sprintf("#%s# always reload and check consistently", CaseAlwaysReload), func() {

	var esIns *elasticsearch.ES
	var deploy *genfilesdeploy.GenFilesDeployment
	var lgc *logconfigv1beta1.LogConfig
	var loggieIns *k8s_loggie.Loggie
	var interval time.Duration
	var totalTime time.Duration

	stopScaleChan := make(chan struct{}, 1)

	BeforeEach(func() {
		_, ei := env.GetResource(elasticsearch.Name)
		esIns = ei.(*elasticsearch.ES)

		_, li := env.GetResource(k8s_loggie.Name)
		loggieIns = li.(*k8s_loggie.Loggie)

		_, dy := env.GetResource(genfilesdeploy.Name)
		deploy = dy.(*genfilesdeploy.GenFilesDeployment)

		caseCfg := env.GetCaseConfig(CaseAlwaysReload)
		inv := caseCfg[FieldsInterval].(string)
		t, err := time.ParseDuration(inv)
		Expect(err).ShouldNot(HaveOccurred())
		interval = t

		totalTimeStr := caseCfg[FieldsTotalTime].(string)
		tt, err := time.ParseDuration(totalTimeStr)
		Expect(err).ShouldNot(HaveOccurred())
		totalTime = tt
	})

	Context("No errors if delete pod every interval time duration", func() {
		It("No ERR logs", func() {

			By("create logconfig")
			logconf := logconfig.GenLogConfigWithESSink(esIns, deploy, nil)
			Expect(logconfig.CreateLogConfig(logconf)).ShouldNot(HaveOccurred())
			lgc = logconf

			By("create genfiles deployment")
			Expect(deploy.Setup(context.Background())).Should(Succeed())
			Eventually(func() (bool, error) {
				return deploy.Ready()
			}).WithTimeout(30 * time.Second).WithPolling(1 * time.Second).Should(BeTrue())

			go scaleDeploymentArbitrary(stopScaleChan, *deploy.Conf.Replicas, interval, totalTime, deploy)

			By("check loggie error log")
			Consistently(func() bool {
				result, err := loggieIns.CheckERRLogs()
				Expect(err).ShouldNot(HaveOccurred())
				return result
			}).WithPolling(1 * time.Second).WithTimeout(totalTime + 5*time.Second).ShouldNot(BeTrue())
		})

	})

	AfterEach(func() {
		stopScaleChan <- struct{}{}

		// delete logconfig
		Expect(logconfig.DeleteLogConfig(lgc)).ShouldNot(HaveOccurred())

		// delete deployment
		Expect(deploy.CleanUp(context.Background())).Should(Succeed())

		// delete elasticsearch index
		Expect(esIns.CleanUp(context.Background())).Should(Succeed())

	})

})

func scaleDeploymentArbitrary(stopChan chan struct{}, maxScale int32, interval time.Duration, totalTime time.Duration, deploy *genfilesdeploy.GenFilesDeployment) {
	By("scale deployment every interval time")

	totalT := time.NewTimer(totalTime)
	t := time.NewTicker(interval)
	defer func() {
		t.Stop()
		totalT.Stop()
	}()

	for {
		select {
		case <-stopChan:
			GinkgoWriter.Println("stop scaling deployment arbitrary")
			return

		case <-totalT.C:
			GinkgoWriter.Println("stop scaling deployment arbitrary when total time is over")
			return

		case <-t.C:
			rand.Seed(time.Now().UnixNano())
			next := rand.Intn(int(interval.Seconds()))
			du, err := time.ParseDuration(fmt.Sprintf("%ds", next))
			Expect(err).ShouldNot(HaveOccurred())
			t.Reset(du)
			GinkgoWriter.Printf("next hit: %+v ", du)

			d, err := deploy.GetDeployment()
			Expect(err).ShouldNot(HaveOccurred()) // FIXME cannot `expect` in this goroutine
			d.ResourceVersion = ""
			oldReplicas := d.Spec.Replicas
			var newReplicas int32 = 1

			for newReplicas == *oldReplicas {
				newReplicas = rand.Int31n(maxScale + 1)
			}

			d.Spec.Replicas = &newReplicas
			Expect(deploy.UpdateDeployment(d)).Should(Succeed())
		}
	}

}
