// Package rum...
// client.go implements the roboust and more cleaner way to write and maintain codes
package rum

import (
	"context"
	"fmt"
	"time"
)

// DispatcherConfig holds configuration for dispatcher registration
type DispatcherConfig struct {
	Service  string
	Settings Settings
}

// ProfileBuilder constructs a profile using the builder pattern
type ProfileBuilder[In, Out any] struct {
	store       *RumStore[In, Out]
	profile     *IProfile[In, Out]
	sequence    ISequence[In]
	timeout     time.Duration
	kit         *Kit[In, Out]
	initialized bool
}

// Client manages Rum profile lifecycle
type Client[In, Out any] struct {
	store    *RumStore[In, Out]
	builders []*ProfileBuilder[In, Out]
	server   *Rum[In, Out]
}

// NewClient creates a new Rum client for building profiles
func NewClient[In, Out any](rum *Rum[In, Out]) *Client[In, Out] {
	if rum == nil {
		panic("server cannot be nil")
	}
	return &Client[In, Out]{
		store:    rum.store,
		server:   rum,
		builders: make([]*ProfileBuilder[In, Out], 0),
	}
}

func (c *Client[In, Out]) Server() *Rum[In, Out] {
	return c.server
}

// CreateProfile initializes a new profile builder
func (c *Client[In, Out]) CreateProfile(seq ISequence[In], timeout time.Duration) *ProfileBuilder[In, Out] {
	if timeout <= 0 {
		panic("timeout must be positive")
	}

	pb := &ProfileBuilder[In, Out]{
		store:       c.store,
		profile:     NewProfile[In, Out](),
		sequence:    seq,
		timeout:     timeout,
		kit:         NewKit[In, Out](),
		initialized: true,
	}
	c.builders = append(c.builders, pb)
	return pb
}

// Run run the rum server
func (pb *Client[In, Out]) Run(ctx context.Context, config RumServer) {
	server := New(ctx, pb.store)
	server.Serve(ctx, config)
}

// RegisterDispatch registers an event handler in a service's dispatcher
func (pb *ProfileBuilder[In, Out]) RegisterDispatch(
	ctx context.Context,
	serviceName string,
	eventName string,
	settings Settings,
	fn func(context.Context, In) (Out, error),
) *ProfileBuilder[In, Out] {
	pb.ensureInitialized()

	if serviceName == "" || eventName == "" {
		panic("serviceName and eventName cannot be empty")
	}
	if fn == nil {
		panic("handler function cannot be nil")
	}

	service := pb.getOrCreateService(ctx, serviceName, settings)
	dispatcher := pb.getOrCreateDispatcher(service, settings)
	dispatcher.Register(eventName, IRegister[In, Out]{Fn: fn})

	return pb
}

// PushService adds a pre-configured service to the profile
func (pb *ProfileBuilder[In, Out]) PushService(name string, service *Service[In, Out]) *ProfileBuilder[In, Out] {
	pb.ensureInitialized()

	if name == "" {
		panic("service name cannot be empty")
	}
	if service == nil {
		panic("service cannot be nil")
	}

	pb.kit.PushService(name, service)
	return pb
}

// PushKit replaces the entire kit with a custom one
func (pb *ProfileBuilder[In, Out]) PushKit(kit *Kit[In, Out]) *ProfileBuilder[In, Out] {
	if kit == nil {
		panic("kit cannot be nil")
	}
	pb.kit = kit
	return pb
}

// Build finalizes the profile and stores it
func (pb *ProfileBuilder[In, Out]) Build() *IProfile[In, Out] {
	pb.ensureInitialized()

	if pb.profile == nil {
		panic("profile was not properly initialized")
	}

	pb.profile.RegisterProfile(pb.sequence, pb.timeout, pb.kit)
	pb.store.SetProfile(pb.profile)

	return pb.profile
}

// BuildAll finalizes all profiles in the client
func (c *Client[In, Out]) BuildAll() error {
	if len(c.builders) == 0 {
		return fmt.Errorf("no profiles to build")
	}

	for i, pb := range c.builders {
		if pb == nil {
			return fmt.Errorf("builder at index %d is nil", i)
		}
		pb.Build()
	}

	return nil
}

// ensureInitialized validates that the builder was properly initialized
func (pb *ProfileBuilder[In, Out]) ensureInitialized() {
	if !pb.initialized {
		panic("ProfileBuilder was not properly initialized")
	}
	if pb.kit == nil {
		pb.kit = NewKit[In, Out]()
	}
}

// getOrCreateService retrieves an existing service or creates a new one
func (pb *ProfileBuilder[In, Out]) getOrCreateService(
	ctx context.Context,
	serviceName string,
	settings Settings,
) *Service[In, Out] {
	service, err := pb.kit.GetService(serviceName)
	if err == nil {
		return service
	}

	service = NewService[In, Out](ctx, settings, serviceName)
	pb.kit.PushService(serviceName, service)
	return service
}

// getOrCreateDispatcher retrieves an existing dispatcher or creates a new one
func (pb *ProfileBuilder[In, Out]) getOrCreateDispatcher(
	service *Service[In, Out],
	settings Settings,
) *Dispatcher[In, Out] {
	dispatcher := service.GetDispatch()
	if dispatcher != nil {
		return dispatcher
	}

	dispatcher = NewDispatcher[In, Out](settings)
	service.SetDispatch(dispatcher)
	return dispatcher
}

// DispatcherBuilder provides a fluent interface for building dispatchers
type DispatcherBuilder[In, Out any] struct {
	dispatcher *Dispatcher[In, Out]
	settings   Settings
}

// NewDispatcherBuilder creates a new dispatcher builder
func NewDispatcherBuilder[In, Out any](settings Settings) *DispatcherBuilder[In, Out] {
	return &DispatcherBuilder[In, Out]{
		dispatcher: NewDispatcher[In, Out](settings),
		settings:   settings,
	}
}

// Register registers an event handler in the dispatcher
func (db *DispatcherBuilder[In, Out]) Register(
	event string,
	fn func(context.Context, In) (Out, error),
) *DispatcherBuilder[In, Out] {
	if event == "" {
		panic("event name cannot be empty")
	}
	if fn == nil {
		panic("handler function cannot be nil")
	}

	db.dispatcher.Register(event, IRegister[In, Out]{Fn: fn})
	return db
}

// Build returns the constructed dispatcher
func (db *DispatcherBuilder[In, Out]) Build() *Dispatcher[In, Out] {
	return db.dispatcher
}
