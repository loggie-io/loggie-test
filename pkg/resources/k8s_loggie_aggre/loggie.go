package k8s_loggie_aggre

import (
	"context"
	"errors"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"loggie-test/pkg/resources"
	"loggie-test/pkg/tools/k8s"
)

const (
	Name = "k8sLoggieAggre"
)

func init() {
	resources.Register(Name, makeLoggie)
}

type Config struct {
	StatefulSetName string `yaml:"statefulSetName"`
	Namespace string `yaml:"namespace" default:"loggie-aggregator"`
}

var _ resources.Resource = (*Loggie)(nil)

type Loggie struct {
	Conf *Config
}

func makeLoggie() interface{} {
	return &Loggie{
		Conf: &Config{},
	}
}

func (r *Loggie) Config() interface{} {
	return r.Conf
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
	sts := &v1.StatefulSet{}
	err := k8s.Mgr.GetClient().Get(context.Background(), types.NamespacedName{
		Namespace: r.Conf.Namespace,
		Name:      r.Conf.StatefulSetName,
	}, sts)

	if err != nil {
		return false, err
	}

	if sts.Status.Replicas == *sts.Spec.Replicas {
		return true, nil
	}

	return false, errors.New("loggie aggregator statefulSet expect replicas failed")
}