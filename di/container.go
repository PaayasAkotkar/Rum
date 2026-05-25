package injection

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"
)

// Container is created just so that injection struct act as a bridge than the whole
type Container struct {
	registry map[reflect.Type]*Entry
	NodeID   string
	mu       sync.RWMutex
	ctx      context.Context
}

// Entry or descirption of the service
type Entry struct {
	Type         reflect.Type
	Factory      Factory
	instance     any
	Lifecycle    Lifecycle
	PoolConfig   *PoolConfig
	pool         chan any
	Build        bool
	mu           sync.RWMutex
	Dependencies []reflect.Type
}

func NewContainer(ctx context.Context, nodeID string) *Container {
	return &Container{
		registry: make(map[reflect.Type]*Entry),
		NodeID:   nodeID,
		ctx:      ctx,
	}
}

func (c *Container) getPooledService(entry *Entry) (any, error) {
	select {
	case conn := <-entry.pool:
		return conn, nil
	case <-time.After(entry.PoolConfig.ConnectionTimeout):
		instance, err := entry.Factory.Fn(c.ctx, c)
		if err != nil {
			return nil, fmt.Errorf("pool exhausted and creation failed: %w", err)
		}
		return instance, nil
	}
}

func (c *Container) ReturnPooledService(t reflect.Type, instance any) error {
	c.mu.RLock()
	entry, ok := c.registry[t]
	c.mu.RUnlock()

	if !ok {
		return fmt.Errorf("service not found")
	}

	select {
	case entry.pool <- instance:
		return nil
	default:
		return nil
	}
}

// Build builds all services
func (c *Container) Build(ctx context.Context) error {
	c.mu.RLock()
	entries := make([]*Entry, 0, len(c.registry))
	for _, e := range c.registry {
		entries = append(entries, e)
	}
	c.mu.RUnlock()

	built := make(map[reflect.Type]bool)

	for _, entry := range entries {
		if err := c.buildEntry(entry, built); err != nil {
			return fmt.Errorf("failed to build %s: %w", entry.Type, err)
		}
	}

	return nil
}

// buildEntry builds the service dependency graph
func (c *Container) buildEntry(entry *Entry, built map[reflect.Type]bool) error {

	if built[entry.Type] {
		return nil
	}

	for _, depType := range entry.Dependencies {
		dep, ok := c.registry[depType]

		if ok {
			if err := c.buildEntry(dep, built); err != nil {
				return err
			}
		}
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	switch entry.Lifecycle {
	case Singleton:
		instance, err := entry.Factory.Fn(c.ctx, c)
		if err != nil {
			return err
		}
		entry.instance = instance

	case Pooled:
		if entry.PoolConfig == nil {
			return fmt.Errorf("pooled lifecycle requires PoolConfig")
		}

		for i := 0; i < entry.PoolConfig.MaxConnections; i++ {
			instance, err := entry.Factory.Fn(c.ctx, c)
			if err != nil {
				return err
			}
			entry.pool <- instance
		}

	case Transient, Scoped:
		// Nothing to build
	}

	entry.Build = true
	built[entry.Type] = true
	return nil
}

// GetService returns the service as per the profile either single, pooled or transient
func (c *Container) GetService(t reflect.Type) (any, error) {
	entry, ok := c.registry[t]

	if !ok {
		return nil, fmt.Errorf("service %s not registered", t)
	}

	entry.mu.RLock()
	defer entry.mu.RUnlock()

	if !entry.Build {
		return nil, fmt.Errorf("service %s not built", t)
	}

	switch entry.Lifecycle {
	case Singleton:
		return entry.instance, nil

	case Transient:
		instance, err := entry.Factory.Fn(c.ctx, c)
		if err != nil {
			return nil, err
		}
		return instance, nil

	case Pooled:
		return c.getPooledService(entry)

	default:
		return nil, fmt.Errorf("unknown lifecycle: %s", entry.Lifecycle)
	}
}
