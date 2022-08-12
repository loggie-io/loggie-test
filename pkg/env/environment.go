package env

import (
	"context"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/creasty/defaults"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"loggie-test/pkg/cfg"
	"loggie-test/pkg/resources"
	"time"
)

var globalEnv *Env

type Env struct {
	config    *cfg.Config
	resources map[string]*resInstance

	err error
}

func newEnv() *Env {
	return &Env{
		config:    &cfg.Config{},
		resources: make(map[string]*resInstance, 0),
	}
}

type resInstance struct {
	status   resources.State
	resource resources.Resource
	instance interface{}
}

func newRes(status resources.State, resource resources.Resource, ins interface{}) *resInstance {
	return &resInstance{
		status:   status,
		resource: resource,
		instance: ins,
	}
}

func Init() *Env {
	globalEnv = newEnv()

	Expect(globalEnv.readConfig()).To(Succeed(), "read environment config")
	Expect(globalEnv.addResources()).To(Succeed(), "add resources")
	return globalEnv
}

func (e *Env) readConfig() error {
	defaultConfigPath := "./config.yml"
	content, err := ioutil.ReadFile(defaultConfigPath)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(content, e.config); err != nil {
		return err
	}

	return nil
}

func (e *Env) addResources() error {
	if e.err != nil {
		return e.err
	}

	for name, resWrapper := range e.config.Resources {
		resIns, err := resources.NewResourceIns(name)
		if err != nil {
			return err
		}

		conf, ok := resIns.(resources.Config)
		if ok {
			if err = e.injectConfig(resWrapper.ResourceIns, conf); err != nil {
				return err
			}
		}

		res := resIns.(resources.Resource)
		e.resources[name] = newRes(resources.ResourceStatusNotReady, res, resIns)
	}
	return nil
}

func (e *Env) injectConfig(rawConf interface{}, configRes resources.Config) error {
	raw, err := yaml.Marshal(rawConf)
	if err != nil {
		return err
	}

	conf := configRes.Config()
	err = yaml.Unmarshal(raw, conf)
	if err != nil {
		return err
	}

	if err := defaults.Set(conf); err != nil {
		return err
	}

	return nil
}

func GetResource(name string) (resources.Resource, interface{}) {
	r, ins, exist := globalEnv.getResource(name)
	Expect(exist).Should(Equal(true), fmt.Sprintf("get resource %s", name))
	return r, ins
}

func (e *Env) getResource(name string) (resources.Resource, interface{}, bool) {
	res, ok := e.resources[name]
	if !ok {
		return nil, nil, false
	}
	return res.resource, res.instance, true
}

func (e *Env) SetupAll() *Env {
	return globalEnv.Setup()
}

func (e *Env) Setup(resName ...string) *Env {
	if e.err != nil {
		return e
	}

	if len(e.resources) == 0 {
		return e
	}

	e.setup(resName...)
	return e
}

func (e *Env) setup(resName ...string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	for _, name := range resName {
		resIns, ok := e.resources[name]
		if !ok {
			continue
		}

		Expect(resIns.resource.Setup(ctx)).Should(Succeed(), fmt.Sprintf("setup resource %s", resIns.resource.Name()))
	}
}

func (e *Env) Require(resName ...string) *Env {
	if e.err != nil {
		return e
	}

	if len(resName) == 0 {
		return e.RequireAll()
	}

	resMap := make(map[string]*resInstance)
	for _, r := range resName {
		res, ok := e.resources[r]
		if !ok {
			e.err = errors.Errorf("resource %s not found", r)
			return e
		}
		resMap[r] = res
	}

	return e.require(resMap)
}

func (e *Env) RequireAll() *Env {
	if globalEnv.err != nil {
		return globalEnv
	}

	// check and retry
	return globalEnv.require(globalEnv.resources)
}

func (e *Env) require(resMap map[string]*resInstance) *Env {
	if len(resMap) == 0 {
		return nil
	}
	count := len(resMap)
	err := retry.Do(func() error {
		for name, res := range resMap {
			if res.status == resources.ResourceStatusReady {
				continue
			}

			if res.status == resources.ResourceStatusFastFailed {
				continue
			}

			ready, err := res.resource.Ready()
			if err != nil {
				Fail(fmt.Sprintf("resource %s is failed: %v", name, err))
				res.status = resources.ResourceStatusFastFailed
				count--
				continue
			}
			if ready {
				res.status = resources.ResourceStatusReady
				count--
				continue
			}
		}

		if count <= 0 {
			return nil
		}

		return errors.New("NotReady")

	}, retry.Delay(1*time.Second), retry.Attempts(10))
	if err != nil {
		e.err = err
		return e
	}
	return e
}

func (e *Env) Do() error {
	return e.err
}

func (e *Env) TearDown() {
	for _, r := range e.resources {
		r.resource.CleanUp(context.Background())
	}
}

func GetCaseConfig(caseName string) map[string]interface{} {
	config, ok := globalEnv.config.Cases[caseName]
	Expect(ok).To(BeTrue())

	return config.CaseConfig
}
