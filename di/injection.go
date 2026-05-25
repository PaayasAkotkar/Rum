// Package injection provides channel-based DI with Injection pattern for cluster scaling
// Design philosophy:
// - DI created once per node
// - All access through channels (async, non-blocking)
// - Each request queued to channel
// - Injection handles service resolution
// - On scale event: rebuild DI for new instance
// - No cross-node metrics needed
package injection

import (
	"context"
)

// Injection it only does one thing is to push the channels
type Injection struct {
	container *Container

	getService    chan *ServiceRequest
	addService    chan *ServiceRegistration
	buildServices chan chan error
	rebuildSignal chan struct{}
	stopChan      chan struct{}
	done          chan struct{}

	nodeID string
	ctx    context.Context
	cancel context.CancelFunc
}

func New(ctx context.Context, nodeID string) *Injection {
	ctx, cancel := context.WithCancel(ctx)

	Injection := &Injection{
		getService:    make(chan *ServiceRequest, 100),
		addService:    make(chan *ServiceRegistration, 50),
		buildServices: make(chan chan error, 1),
		rebuildSignal: make(chan struct{}, 1),
		stopChan:      make(chan struct{}),
		done:          make(chan struct{}),
		nodeID:        nodeID,
		ctx:           ctx,
		cancel:        cancel,
	}

	go Injection.pipe()

	return Injection
}
