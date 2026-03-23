package registry

import (
	"context"
	"fmt"
	"sync"
)

type Lifecycle interface {
	Init(context.Context) error
	Start(context.Context) error
	Stop(context.Context) error
}

type Provider interface {
	Name() string
}

type Container struct {
	mu     sync.RWMutex
	values map[string]any
}

func New() *Container {
	return &Container{values: make(map[string]any)}
}

func (c *Container) Register(name string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[name] = value
}

func (c *Container) MustGet(name string) any {
	value, ok := c.Get(name)
	if !ok {
		panic(fmt.Sprintf("registry: component %q not found", name))
	}
	return value
}

func (c *Container) Get(name string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.values[name]
	return value, ok
}

func (c *Container) Names() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	items := make([]string, 0, len(c.values))
	for name := range c.values {
		items = append(items, name)
	}
	return items
}
