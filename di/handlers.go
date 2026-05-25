package injection

import (
	"log"
	"time"
)

func (h *Injection) handleGetService(req *ServiceRequest) {
	instance, err := h.container.GetService(req.ServiceType)

	select {
	case req.ResponseCh <- &ServiceResponse{
		Instance: instance,
		Error:    err,
	}:
	case <-time.After(req.Timeout):
		log.Printf("[%s] Timeout sending service response\n", h.nodeID)
	}
}

func (h *Injection) handleAddService(reg *ServiceRegistration, container *Container) {
	entry := &Entry{
		Type:       reg.Type,
		Factory:    reg.Factory,
		Lifecycle:  reg.Lifecycle,
		PoolConfig: reg.PoolConfig,
	}

	if reg.Lifecycle == Pooled && reg.PoolConfig != nil {
		entry.pool = make(chan any, reg.PoolConfig.MaxConnections)
	}

	container.mu.Lock()
	defer container.mu.Unlock()

	container.registry[reg.Type] = entry
}

func (h *Injection) handleBuild(container *Container) error {
	return container.Build(h.ctx)
}
