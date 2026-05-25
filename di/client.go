package injection

import (
	"context"
	"fmt"
	"reflect"
	cheetah "rum/app/cheetah"
	"time"
)

// Client implemented client so that it can be more roboust
type Client struct {
	Injection *Injection
	pending   map[reflect.Type]*Entry
	err       error
	cheetah   *cheetah.Cheetah[string]
}

// NewClient creates a new DI client
func NewClient(ctx context.Context, nodeID string) *Client {
	return &Client{
		Injection: New(ctx, nodeID),
		pending:   make(map[reflect.Type]*Entry),
		cheetah:   cheetah.New[string](),
	}
}

// AddSingleton create once and shutdown
func (c *Client) AddSingleton(t reflect.Type, factory Factory) *Client {
	if c.err != nil {
		return c
	}

	select {
	case c.Injection.addService <- &ServiceRegistration{
		Type:      t,
		Factory:   factory,
		Lifecycle: Singleton,
	}:
	case <-time.After(5 * time.Second):
		c.err = fmt.Errorf("timeout registering service")
	}

	return c
}

// AddTransient create for every call
func (c *Client) AddTransient(t reflect.Type, factory Factory) *Client {
	if c.err != nil {
		return c
	}

	select {
	case c.Injection.addService <- &ServiceRegistration{
		Type:      t,
		Factory:   factory,
		Lifecycle: Transient,
	}:
	case <-time.After(5 * time.Second):
		c.err = fmt.Errorf("timeout registering service")
	}

	return c
}

// BuildStatus returns if the status is ready or not
func (c *Client) BuildStatus() chan *string {
	return c.cheetah.Subscribe("build_status")
}

// CloseBuildStatus unsubscribes from the build status event via cheetah
func (c *Client) CloseBuildStatus(status chan *string) {
	c.cheetah.Unsubscribe("build_status", status)
}

// AddPooled creates a pool of instances and manages them through a pool.
func (c *Client) AddPooled(t reflect.Type, factory Factory, poolConfig *PoolConfig) *Client {
	if c.err != nil {
		return c
	}

	select {
	case c.Injection.addService <- &ServiceRegistration{
		Type:       t,
		Factory:    factory,
		Lifecycle:  Pooled,
		PoolConfig: poolConfig,
	}:
	case <-time.After(5 * time.Second):
		c.err = fmt.Errorf("timeout registering service")
	}

	return c
}

// Build builds the service dependency graph as per the profile
// note: make sure to subscribe to the buildstatus before calling it
// note: it can be triggered on scaleup using TriggerRebuild
func (c *Client) Build(ctx context.Context) error {
	if c.err != nil {
		return c.err
	}

	resultCh := make(chan error, 1)
	select {
	case c.Injection.buildServices <- resultCh:
		err := <-resultCh
		if err == nil {
			msg := "ready"
			c.cheetah.Publish("build_status", &msg)
		}
		return err
	case <-time.After(10 * time.Second):
		return fmt.Errorf("timeout building services")
	case <-ctx.Done():
		return fmt.Errorf("context cancelled")
	}
}

// GetService returns the service from the registry as per the profile
// note: make sure to use the buildstatus before calling it
func (c *Client) GetService(t reflect.Type) (any, error) {
	responseCh := make(chan *ServiceResponse, 1)

	req := &ServiceRequest{
		ServiceType: t,
		ResponseCh:  responseCh,
		Timeout:     5 * time.Second,
	}

	select {
	case c.Injection.getService <- req:
		response := <-responseCh
		return response.Instance, response.Error
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout getting service")
	}
}

// ReturnPooledService returns the service to the pool
// note: the service is returned to the pool and can be reused
func (c *Client) ReturnPooledService(t reflect.Type, instance any) error {
	return c.Injection.container.ReturnPooledService(t, instance)
}

// TriggerRebuild created especailly only for the cluster scaling
func (c *Client) TriggerRebuild() error {
	select {
	case c.Injection.rebuildSignal <- struct{}{}:
		return nil
	case <-time.After(1 * time.Second):
		return fmt.Errorf("timeout triggering rebuild")
	}
}

// Stop cleanups
func (c *Client) Stop() error {
	if c.Injection == nil {
		return nil
	}

	if c.Injection.cancel != nil {
		c.Injection.cancel()
	}

	select {
	case c.Injection.stopChan <- struct{}{}:
	case <-c.Injection.done:
	}

	<-c.Injection.done

	return nil
}
