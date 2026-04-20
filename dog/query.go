package dog

import (
	"time"
)

// GetPolicy returns a policy by name
func (rd *Dog[T]) GetPolicy(name string) *Policy[T] {
	rd.mu.RLock()
	defer rd.mu.RUnlock()
	return rd.policy[name]
}

// GetAllPolicies returns all registered policy names
func (rd *Dog[T]) GetAllPolicies() []string {
	rd.mu.RLock()
	defer rd.mu.RUnlock()

	names := make([]string, 0, len(rd.policy))
	for name := range rd.policy {
		names = append(names, name)
	}
	return names
}

// GetProgress returns progress for a policy
func (rd *Dog[T]) GetProgress(name string) *ExeProgress {
	rd.mu.RLock()
	defer rd.mu.RUnlock()

	if progress, exists := rd.progress[name]; exists {
		return progress
	}
	return nil
}

// GetHealth returns health status for a policy
func (rd *Dog[T]) GetHealth(name string) *Health {
	rd.mu.RLock()
	defer rd.mu.RUnlock()

	if health, exists := rd.health[name]; exists {
		return health
	}
	return nil
}

// GetTimeout returns the timeout limit for a policy
func (rd *Dog[T]) GetTimeout(name string) time.Duration {
	rd.mu.RLock()
	defer rd.mu.RUnlock()

	if policy, exists := rd.policy[name]; exists {
		return policy.GetBase()
	}

	return rd.base
}

// GetDuration returns the latest recorded duration for a policy
func (rd *Dog[T]) GetDuration(name string) time.Duration {
	rd.mu.RLock()
	defer rd.mu.RUnlock()

	return rd.durations[name]
}

// Pakkun returns the full report for a policy
// inspiration of the name from Naruto -> Kakashi's summing dog
func (rd *Dog[T]) Pakkun(name string) *WatchdogReport {

	ch := rd.chakra.Subscribe(name)
	defer rd.chakra.Kai(name, ch)

	for {
		select {
		case <-rd.ctx.Done():
			return nil
		case result := <-ch:
			if rd.Settings.ShowReport {
				rd.generateReport(name)
			}
			return result
		}
	}
}
