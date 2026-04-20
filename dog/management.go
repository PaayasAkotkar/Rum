package dog

import (
	"fmt"
	"time"
)

// resetPolicy releases the call counts for specific policy
func (rd *Dog[T]) resetPolicy(name string) {
	rd.mu.Lock()
	defer rd.mu.Unlock()

	if policy, exists := rd.policy[name]; exists {
		policy.Release()
		policy.Succeed.Release()
		policy.Fail.Release()

		rd.progress[name] = NewProgress()
		rd.reports[name] = &WatchdogReport{
			PolicyName:     name,
			StartTime:      time.Now(),
			TimeLimit:      policy.GetBase(),
			FailureReasons: make([]string, 0),
		}
		fmt.Printf("[Reset] Policy '%s' reset\n", name)
	}
}

// resetPolicy releases all the call counts from all the policies
func (rd *Dog[T]) resetAllPolicies() {
	rd.mu.Lock()
	defer rd.mu.Unlock()

	for name, policy := range rd.policy {
		policy.Release()
		policy.Succeed.Release()
		policy.Fail.Release()

		rd.progress[name] = NewProgress()
		fmt.Printf("[Reset] Policy '%s' reset\n", name)
	}
}
