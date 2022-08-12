package unixsock

import (
	"context"
	"loggie-test/pkg/resources"
	"loggie-test/pkg/tools/generator"
	"net"
)

const Name = "unixsock"

func init() {
	resources.Register(Name, makeunixsock)
}

type Config struct {
	Path       string `yaml:"path,omitempty"`
	LinesCount int    `yaml:"maxLines,omitempty" default:"1024"`
	LineBytes  int    `yaml:"lineBytes,omitempty" default:"1024"`
}

var _ resources.Resource = (*unixsock)(nil)

type unixsock struct {
	envType string
	config  *Config

	setupErr  error
	setupDone bool
}

func makeunixsock() interface{} {
	return &unixsock{
		config: &Config{},
	}
}

func (r *unixsock) Config() interface{} {
	return r.config
}

func (r *unixsock) Name() string {
	return Name
}

func (r *unixsock) Setup(ctx context.Context) error {
	r.setupInLocal(ctx)
	return nil
}

func (r *unixsock) setupInLocal(ctx context.Context) {
	conn, err := net.Dial("unix", r.config.Path)
	if err != nil {
		//Fail(fmt.Sprintf("unix dial failed: %v", err))
	}

	err = generator.WriteLines(ctx, conn, r.config.LinesCount, r.config.LineBytes)
	if err != nil {
		r.setupErr = err
		return
	}
	r.setupDone = true
	return
}

func (r *unixsock) setupInKubernetes(ctx context.Context) {

}

func (r *unixsock) CleanUp(ctx context.Context) error {
	return nil
}

func (r *unixsock) Ready() (bool, error) {
	if r.setupErr != nil {
		return false, r.setupErr
	}

	return r.setupDone, nil
}
