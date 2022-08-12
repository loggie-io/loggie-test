package resources

import (
	"context"
	"errors"
	"fmt"
)

type State int

const (
	ResourceStatusNotReady State = iota
	ResourceStatusFastFailed
	ResourceStatusReady
)

func (s State) String() string {
	switch s {
	case ResourceStatusNotReady:
		return "ResourceStatusNotReady"

	case ResourceStatusFastFailed:
		return "ResourceStatusFastFailed"

	case ResourceStatusReady:
		return "ResourceStatusReady"

	default:
		return "Unknown"
	}
}

var defaultResourceRegistry *ResourceFactory

func init() {
	defaultResourceRegistry = newRegistry()
}

type Resource interface {
	Name() string
	Ready() (bool, error)
	Setup(ctx context.Context) error
	CleanUp(ctx context.Context) error
}

type Config interface {
	Config() interface{}
}

type NewResource func() interface{}

type ResourceFactory struct {
	factoryFn map[string]NewResource
}

func newRegistry() *ResourceFactory {
	return &ResourceFactory{
		factoryFn: make(map[string]NewResource, 0),
	}
}

func (rf *ResourceFactory) addResource(name string, factory NewResource) {
	rf.factoryFn[name] = factory
}

func (rf *ResourceFactory) getResource(name string) (interface{}, error) {
	factory, ok := rf.factoryFn[name]
	if !ok {
		return nil, errors.New(fmt.Sprintf("resource %s not found", name))
	}

	res := factory()

	return res, nil
}

func Register(name string, factory NewResource) {
	defaultResourceRegistry.addResource(name, factory)
}

func NewResourceIns(name string) (interface{}, error) {
	return defaultResourceRegistry.getResource(name)
}
