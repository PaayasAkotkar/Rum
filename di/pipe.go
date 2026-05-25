package injection

import "log"

func (h *Injection) pipe() {
	defer close(h.done)

	container := NewContainer(h.ctx, h.nodeID)
	h.container = container

	for {
		select {
		// Get service request
		case req := <-h.getService:
			h.handleGetService(req)

		// Register service
		case reg := <-h.addService:
			h.handleAddService(reg, container)

		// Build all services
		case resultCh := <-h.buildServices:
			resultCh <- h.handleBuild(container)

		// Scale event: rebuild entire DI
		case <-h.rebuildSignal:
			log.Printf("[%s] Rebuild signal received, reinitializing DI\n", h.nodeID)
			container = NewContainer(h.ctx, h.nodeID)
			h.container = container
			// Injection stays alive, ready to re-register and rebuild

		// Stop
		case <-h.stopChan:
			return
		case <-h.ctx.Done():
			return
		}
	}
}
