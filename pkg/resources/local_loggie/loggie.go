package local_loggie

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/prometheus/prom2json"
	"io/ioutil"
	"loggie-test/pkg/resources"
	"loggie-test/pkg/tools/prometheus"
	"net/http"
	"os"
	"path"
	"time"
)

const Name = "localLoggie"

func init() {
	resources.Register(Name, makeLoggie)
}

type Config struct {
	PipelinePath string `yaml:"pipelinePath,omitempty"`
	Endpoint     string `yaml:"endpoint,omitempty"`
}

var _ resources.Resource = (*Loggie)(nil)

type Loggie struct {
	config *Config
}

func makeLoggie() interface{} {
	return &Loggie{
		config: &Config{},
	}
}

func (r *Loggie) Config() interface{} {
	return r.config
}

func (r *Loggie) Name() string {
	return Name
}

func (r *Loggie) Setup(ctx context.Context) error {
	return nil
}

func (r *Loggie) CleanUp(ctx context.Context) error {
	return nil
}

func (r *Loggie) Ready() (bool, error) {
	resp, err := http.Get(r.config.Endpoint + "/api/v1/reload/config")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, errors.Errorf("request to loggie %s failed", r.config.Endpoint)
	}
	return true, nil
}

func (r *Loggie) UpdatePipelineAndWaitForReloadPeriod(pipeConfigContent string) error {
	_, err := os.Stat(r.config.PipelinePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err := os.MkdirAll(r.config.PipelinePath, 0777); err != nil {
			return err
		}
	}

	err = ioutil.WriteFile(path.Join(r.config.PipelinePath, "pipeline.yml"), []byte(pipeConfigContent), 0777)
	time.Sleep(10 * time.Second)
	return err
}

func (r *Loggie) Restart() {
}

func (r *Loggie) GetPrometheusMetrics() (map[string]*prom2json.Family, error) {

	endpoint := fmt.Sprintf("%s/%s", r.config.Endpoint, "metrics")
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
