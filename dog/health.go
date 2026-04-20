package dog

import "time"

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

// genHealth returns health based on current progress
// 100-50 : healhty
// 0-30: danger
// 0: not healthy
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

// calculateStatus returns the current report status baseed on executions
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
