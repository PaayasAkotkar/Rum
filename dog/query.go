package dog

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

func (d *Dog[T]) summon(name string) {
	log.Printf("[summing] %s", name)
	policy := d.GetPolicy(name)
	if policy == nil {
		// Policy not found, report via Bark and exit
		d.Bark(IBark{Policy: name, Reason: fmt.Errorf("policy %s not found", name).Error()})
		return
	}

	for _, f := range policy.Fn {
		if f.Void != nil {
			functionStartTime := time.Now()
			err := (*f.Void)()
			measuredDuration := time.Since(functionStartTime)
			if err != nil {
				d.Bark(IBark{
					Policy: "properTracking",
					Reason: err.Error(),
				})
			} else {
				d.Done(IDone{
					PolicyName:   "properTracking",
					FuncName:     "operation",
					Rank:         f.Rank,
					FuncDuration: measuredDuration,
					Output:       []byte("void succeed"),
				})
			}
		} else if f.Fn != nil {
			functionStartTime := time.Now()
			resp, err := (*f.Fn)()
			measuredDuration := time.Since(functionStartTime)
			if err != nil {
				d.Bark(IBark{ // placeholder
					Policy: "properTracking",
					Reason: err.Error(),
				})
			} else {
				p, err := json.Marshal(resp)
				if err != nil {
					d.Bark(IBark{
						Policy: "properTracking",
						Reason: err.Error(),
					})
				}
				d.Done(IDone{
					PolicyName:   "properTracking",
					FuncName:     "operation",
					Rank:         f.Rank,
					FuncDuration: measuredDuration,
					Output:       p,
				})
			}
		}
	}
}

// // GetPolicy returns a policy by name
// func (rd *Dog[T]) GetPolicy(name string) *Policy[T] {
// 	rd.mu.RLock()
// 	defer rd.mu.RUnlock()
// 	return rd.policy[name]
// }

// // GetAllPolicies returns all registered policy names
// func (rd *Dog[T]) GetAllPolicies() []string {
// 	rd.mu.RLock()
// 	defer rd.mu.RUnlock()

// 	names := make([]string, 0, len(rd.policy))
// 	for name := range rd.policy {
// 		names = append(names, name)
// 	}
// 	return names
// }

// // GetProgress returns progress for a policy
// func (rd *Dog[T]) GetProgress(name string) *ExeProgress {
// 	rd.mu.RLock()
// 	defer rd.mu.RUnlock()

// 	if progress, exists := rd.progress[name]; exists {
// 		return progress
// 	}
// 	return nil
// }

// // GetHealth returns health status for a policy
// func (rd *Dog[T]) GetHealth(name string) *Health {
// 	rd.mu.RLock()
// 	defer rd.mu.RUnlock()

// 	if health, exists := rd.health[name]; exists {
// 		return health
// 	}
// 	return nil
// }

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

// GetPolicy retrieves a policy by name
func (rd *Dog[T]) GetPolicy(name string) *Policy[T] {
	rd.mu.RLock()
	defer rd.mu.RUnlock()
	return rd.policy[name]
}

// GetProgress retrieves progress for a policy
func (rd *Dog[T]) GetProgress(name string) *ExeProgress {
	rd.mu.RLock()
	defer rd.mu.RUnlock()

	if progress, exists := rd.progress[name]; exists {
		return progress
	}
	return nil
}

// GetHealth retrieves health status for a policy
func (rd *Dog[T]) GetHealth(name string) *Health {
	rd.mu.RLock()
	defer rd.mu.RUnlock()

	if health, exists := rd.health[name]; exists {
		return health
	}
	return nil
}

// GetMetrics retrieves system metrics for a policy
func (rd *Dog[T]) GetMetrics(name string) *SystemMetrics {
	rd.mu.RLock()
	defer rd.mu.RUnlock()

	if metrics, exists := rd.metrics[name]; exists {
		return metrics
	}
	return nil
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
