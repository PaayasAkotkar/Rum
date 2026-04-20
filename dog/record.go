package dog

import (
	"fmt"
	"log"
	rumpaint "rum/app/paint"
	"time"
)

// // RecordDuration records the execution duration for a policy
func (rd *Dog[T]) addDuration(policyName string, duration time.Duration) {
	rd.mu.Lock()
	defer rd.mu.Unlock()

	title := fmt.Sprintf("Record-Duration Policy ⏲️ %s\n", policyName)
	desc := fmt.Sprintf("duration: %v\n", duration)
	t := rumpaint.Card(title, desc)
	log.Println(t)
	rd.durations[policyName] = duration
}

// recordStart records the start time for monitoring
func (rd *Dog[T]) recordStart(policyName string, startTime time.Time) {
	if progress, exists := rd.progress[policyName]; exists {
		progress.StartedAtNano = startTime.UnixNano()
	}
}
