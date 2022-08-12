package k8s_loggie

import (
	"bufio"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/prometheus/prom2json"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"loggie-test/pkg/resources"
	"loggie-test/pkg/tools/k8s"
	"loggie-test/pkg/tools/prometheus"
	"net/http"
	"strings"
	"sync"
)

const (
	Name = "k8sLoggie"

	ListenPort = 9196
)

func init() {
	resources.Register(Name, makeLoggie)
}

type Config struct {
	PodName   string `yaml:"podName"`
	Namespace string `yaml:"namespace" default:"loggie"`
}

var _ resources.Resource = (*Loggie)(nil)

type Loggie struct {
	config *Config
	once   sync.Once
}

func makeLoggie() interface{} {
	return &Loggie{
		config: &Config{},
	}
}

type ILoggie interface {
	GetPrometheusMetrics() (map[string]*prom2json.Family, error)
}

func (r *Loggie) Config() interface{} {
	return r.config
}

func (r *Loggie) Name() string {
	return Name
}

func (r *Loggie) Setup(ctx context.Context) error {
	r.PortForward()

	return nil
}

func (r *Loggie) CleanUp(ctx context.Context) error {
	return nil
}

func (r *Loggie) Ready() (bool, error) {
	pod := &corev1.Pod{}
	err := k8s.Mgr.GetClient().Get(context.Background(), types.NamespacedName{
		Namespace: r.config.Namespace,
		Name:      r.config.PodName,
	}, pod)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *Loggie) PortForward() {
	r.once.Do(func() {
		go func() {
			err := k8s.PortForward(r.config.Namespace, r.config.PodName, 9196)
			if err != nil {
				panic(err)
			}
		}()
	})
}

func (r *Loggie) GetPrometheusMetrics() (map[string]*prom2json.Family, error) {
	endpoint := fmt.Sprintf("http://localhost:%d/metrics", ListenPort)
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("server returned HTTP status %s", resp.Status)
	}

	out, err := prometheus.ToFamily(resp.Body)
	if err != nil {
		return nil, errors.WithMessage(err, "read response body failed")
	}

	return out, nil
}

func (r *Loggie) CheckERRLogs() (bool, error) {
	tailCount := int64(100)
	podLogOpts := corev1.PodLogOptions{
		TailLines: &tailCount,
	}
	req := k8s.Mgr.Kubeclient.CoreV1().Pods(r.config.Namespace).GetLogs(r.config.PodName, &podLogOpts)
	podLogs, err := req.Stream(context.Background())
	if err != nil {
		return false, err
	}
	defer podLogs.Close()

	scan := bufio.NewScanner(podLogs)

	for {
		if !scan.Scan() {
			if scan.Err() != nil { // close connection when scan error
				return false, err
			}

			break
		}

		body := string(scan.Bytes())
		if strings.Contains(body, "ERR") {
			return true, errors.Errorf("ERR Log Found: %s", body)
		}
	}

	return false, nil
}
