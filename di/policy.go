package injection

import (
	"context"
	"reflect"
	"time"
)

type Lifecycle string

const (
	Singleton Lifecycle = "singleton"
	Transient Lifecycle = "transient"
	Scoped    Lifecycle = "scoped"
	Pooled    Lifecycle = "pooled"
)

type ServiceRequest struct {
	ServiceType reflect.Type
	ResponseCh  chan *ServiceResponse
	Timeout     time.Duration
}

type ServiceResponse struct {
	Instance any
	Error    error
}

type ServiceRegistration struct {
	Type       reflect.Type
	Factory    Factory
	Lifecycle  Lifecycle
	PoolConfig *PoolConfig
}

type PoolConfig struct {
	MaxConnections    int
	MinConnections    int
	MaxIdleTime       time.Duration
	ConnectionTimeout time.Duration
}

// Factory push how the container suppose to create service instance
type Factory struct {
	Fn func(ctx context.Context, c *Container) (any, error)
}
