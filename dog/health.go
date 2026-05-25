package dog

import (
	"fmt"
	"log"
	"time"
)

// // updateProgress updates real-time progress
// func (rd *Dog[T]) updateProgress(policyName string, percent uint64) {
// 	if percent > 100 {
// 		percent = 100
// 	}

// 	rd.mu.Lock()
// 	defer rd.mu.Unlock()

// 	if progress, exists := rd.progress[policyName]; exists {
// 		progress.SetCompletion(time.Duration(percent))
// 		health := rd.genHealth(percent)
// 		progress.SetHealth(health)
// 		rd.health[policyName] = &health
// 	}
// }

// // genHealth returns health based on current progress
// // 100-50 : healhty
// // 0-30: danger
// // 0: not healthy
// func (rd *Dog[T]) genHealth(progress uint64) Health {
// 	h := Health{}

// 	switch {
// 	case progress >= 75:
// 		h.IsHealthy = true
// 		h.Silent = true
// 	case progress >= 50 && progress < 75:
// 		h.IsHealthy = true
// 		h.Mid = false
// 	case progress >= 30 && progress < 50:
// 		h.IsHealthy = true
// 		h.Mid = true
// 	case progress > 0 && progress < 30:
// 		h.Danger = true
// 		h.IsHealthy = false
// 	case progress == 0:
// 		h.IsHealthy = false
// 	}

// 	return h
// }

// // calculateStatus returns the current report status baseed on executions
// func (rd *Dog[T]) calculateStatus(report *WatchdogReport) string {
// 	if report.ExecutionCount.Load() == 0 {
// 		return "pending"
// 	}
// 	if report.SuccessRate >= 0.95 {
// 		return "healthy"
// 	} else if report.SuccessRate >= 0.80 {
// 		return "warning"
// 	} else {
// 		return "critical"
// 	}
// }

// handleSummon processes function execution
func (rd *Dog[T]) handleSummon(policyName string) {
	rd.mu.RLock()
	policy, pExists := rd.policy[policyName]
	lifecycle, lExists := rd.lifecycle[policyName]
	rd.mu.RUnlock()

	if !pExists {
		log.Printf("❌ Summon failed: policy %s not found", policyName)
		rd.Bark(IBark{
			Policy: policyName,
			Reason: fmt.Sprintf("policy %s not found", policyName),
		})
		return
	}

	if !lExists || !lifecycle.IsInState(StateMonitoring) {
		log.Printf("❌ Summon failed: policy %s not in monitoring state (current: %s)",
			policyName, lifecycle.GetState())
		rd.Bark(IBark{
			Policy: policyName,
			Reason: fmt.Sprintf("policy not in monitoring state: %s", lifecycle.GetState()),
		})
		return
	}

	log.Printf("🔥 Summoning policy: %s", policyName)

	funcs := policy.GetFunc()
	if len(funcs) == 0 {
		log.Printf("⚠️  No functions to summon for policy %s", policyName)
		return
	}

	// Sample metrics at start
	rd.mu.Lock()
	if metrics, ok := rd.metrics[policyName]; ok {
		metrics.Sample()
	}
	rd.mu.Unlock()

	// Execute each function
	for _, fn := range funcs {
		rdCopy := rd // Capture for closure
		rdCopy.executeFunction(policyName, fn)
	}
}

// executeFunction executes a single function and tracks it
func (rd *Dog[T]) executeFunction(policyName string, fn Funcs[T]) {
	startTime := time.Now()

	// Sample metrics before execution
	rd.mu.Lock()
	if metrics, ok := rd.metrics[policyName]; ok {
		metrics.Sample()
	}
	rd.mu.Unlock()

	var err error
	var duration time.Duration
	var output []byte

	log.Printf("  ▶️  Executing function: %s (rank: %d)", fn.Name, fn.Rank)

	// Execute void function
	if fn.Void != nil {
		err = (*fn.Void)()
		duration = time.Since(startTime)
		output = []byte("void execution succeeded")

		if err != nil {
			rd.Bark(IBark{
				Policy:   policyName,
				Reason:   err.Error(),
				Duration: duration,
			})
			return
		}

		rd.Done(IDone{
			PolicyName:   policyName,
			FuncName:     fn.Name,
			Rank:         fn.Rank,
			FuncDuration: duration,
			Output:       output,
		})
		return
	}

	// Execute function with return
	if fn.Fn != nil {
		resp, err := (*fn.Fn)()
		duration = time.Since(startTime)

		if err != nil {
			rd.Bark(IBark{
				Policy:   policyName,
				Reason:   err.Error(),
				Duration: duration,
			})
			return
		}

		// Try to serialize output
		if resp != nil {
			output = serializeOutput(resp)
		} else {
			output = []byte("nil response")
		}

		rd.Done(IDone{
			PolicyName:   policyName,
			FuncName:     fn.Name,
			Rank:         fn.Rank,
			FuncDuration: duration,
			Output:       output,
		})
		return
	}
}


// updateProgress updates real-time progress
func (rd *Dog[T]) updateProgress(policyName string, percent uint64) {
	if percent > 100 {
		percent = 100
	}

	rd.mu.Lock()
	defer rd.mu.Unlock()

	if progress, exists := rd.progress[policyName]; exists {
		progress.SetCompletion(time.Duration(percent))
		health := rd.genHealth(percent)
		progress.SetHealth(health)
		rd.health[policyName] = &health
	}
}

// genHealth returns health based on progress
func (rd *Dog[T]) genHealth(progress uint64) Health {
	h := Health{}

	switch {
	case progress >= 75:
		h.IsHealthy = true
		h.Silent = true
	case progress >= 50 && progress < 75:
		h.IsHealthy = true
		h.Mid = false
	case progress >= 30 && progress < 50:
		h.IsHealthy = true
		h.Mid = true
	case progress > 0 && progress < 30:
		h.Danger = true
		h.IsHealthy = false
	case progress == 0:
		h.IsHealthy = false
	}

	return h
}

// calculateStatus returns the current report status
func (rd *Dog[T]) calculateStatus(report *WatchdogReport) string {
	if report.ExecutionCount.Load() == 0 {
		return "pending"
	}
	if report.SuccessRate >= 0.95 {
		return "healthy"
	} else if report.SuccessRate >= 0.80 {
		return "warning"
	} else {
		return "critical"
	}
}
