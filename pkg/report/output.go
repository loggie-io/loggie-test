package report

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"loggie-test/pkg/report/echarts"
	"loggie-test/pkg/resources/k8s_loggie"
	"loggie-test/pkg/tools/prometheus"
	"loggie-test/pkg/tools/unit"
	"sync"
	"time"
)

const tSLayout = "15:04:05"

var mutex sync.Mutex

func RecordMetrics(stop chan struct{}, loggieIns k8s_loggie.ILoggie, id string) {
	mutex.Lock()
	defer mutex.Unlock()

	lineChart := echarts.NewLineCharts()
	var xAxis []string
	sinkEventQpsSeries := echarts.NewSeriesSet()
	cpuSeries := echarts.NewSeriesSet()
	memSeries := echarts.NewSeriesSet()

	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-stop:
			lineChart.AddChart("Loggie CPU", "percentage(%)", "time", xAxis, cpuSeries)
			lineChart.AddChart("Loggie Memory", "memory(MiB)", "time", xAxis, memSeries)
			lineChart.AddChart("Loggie Sink", "qps(line/s)", "time", xAxis, sinkEventQpsSeries)
			Expect(lineChart.Render(id)).ShouldNot(HaveOccurred())
			return

		case <-t.C:
			metric, err := loggieIns.GetPrometheusMetrics()
			if err != nil {
				GinkgoWriter.Printf("get prometheus metrics failed: %v\n", err)
			}

			timeStamp := time.Now().Format(tSLayout)
			xAxis = append(xAxis, timeStamp) // add timestamp as X

			if cpuPer, err := prometheus.SingleGAUGE(metric, "loggie_sys_cpu_percent"); err == nil {
				cpuSeries.AddData("loggie_sys_cpu_percent", timeStamp, cpuPer)
			}

			if memRss, err := prometheus.SingleGAUGE(metric, "loggie_sys_mem_rss"); err == nil {
				mib := unit.BytesToMiB(memRss)
				memSeries.AddData("loggie_sys_mem_rss", timeStamp, mib)
			}
			if heapInUsed, err := prometheus.SingleGAUGE(metric, "go_memstats_heap_inuse_bytes"); err == nil {
				mib := unit.BytesToMiB(heapInUsed)
				memSeries.AddData("go_memstats_heap_inuse_bytes", timeStamp, mib)
			}

			multiseries, err := prometheus.GetLoggieMultiGAUGE(metric, "loggie_sink_event_qps")
			if err != nil {
				GinkgoWriter.Printf("get loggie gauge failed: %v\n", err)
			}
			for _, se := range multiseries {
				sinkEventQpsSeries.AddData(se.Name, timeStamp, se.Value)
			}
		}
	}

}
