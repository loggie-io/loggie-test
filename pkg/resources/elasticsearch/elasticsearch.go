package elasticsearch

import (
	"context"
	"github.com/olivere/elastic/v7"
	"loggie-test/pkg/resources"
	"loggie-test/pkg/tools/k8s"
	"sync"
)

const Name = "elasticsearch"

func init() {
	resources.Register(Name, makeES)
}

type Config struct {
	Namespace string `yaml:"namespace,omitempty"`
	Service string `yaml:"service,omitempty"`

	Index   string   `yaml:"index,omitempty" validate:"required"`
}

var _ resources.Resource = (*ES)(nil)

type ES struct {
	Client *elastic.Client
	Conf   *Config
	LocalAddress string

	once sync.Once
}

func makeES() interface{} {
	return &ES{
		Conf: &Config{},
	}
}

func (r *ES) Config() interface{} {
	return r.Conf
}

func (r *ES) Name() string {
	return Name
}

func (r *ES) Setup(ctx context.Context) error {
	r.portForward()
	client, err := elastic.NewClient(elastic.SetSniff(false))
	if err != nil {
		return err
	}

	r.Client = client

	return nil
}

func (r *ES) CleanUp(ctx context.Context) error {
	_, err := r.Client.DeleteIndex(r.Conf.Index).Do(context.Background())
	return err
}

func (r *ES) Ready() (bool, error) {
	_, _, err := r.Client.Ping("http://localhost:9200").Do(context.Background())
	if err != nil {
		return false, err
	}

	return true, nil
}

type CountResp struct {
	Count int `json:"count"`
}

func (r *ES) Count() (int64, error) {
	return r.Client.Count(r.Conf.Index).Do(context.Background())
}

func (r *ES) portForward() {
	if r.Conf.Namespace == "" || r.Conf.Service == "" {
		return
	}

	r.once.Do(func() {
		go func() {
			err := k8s.PortForwardByService(r.Conf.Namespace, r.Conf.Service, 9200)
			if err != nil {
				panic(err)
			}
		}()
	})
}
