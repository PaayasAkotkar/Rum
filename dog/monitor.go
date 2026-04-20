package dog

import (
	"fmt"
	"log"
	rumpaint "rum/app/paint"
	"time"
)

// monitorPolicy monitors all functions in a policy
func (rd *Dog[T]) monitorPolicy(policyName string) {
	log.Println("[monitorPolicy]....")
	rd.mu.Lock()
	defer rd.mu.Unlock()

	policy, exists := rd.policy[policyName]
	for p := range rd.policy {
		log.Println("policies: ", p)
	}

	if !exists {
		fmt.Printf("[ParkDog] Policy '%s' not found\n", policyName)
		return
	}

	// if ch, alreadyTicking := rd.tickers[policyName]; alreadyTicking {
	// 	fmt.Printf("[ParkDog] Policy '%s' already being monitored\n", policyName)
	// 	close(ch)
	// 	return
	// }

	if _, alreadyMonitoring := rd.monitors[policyName]; alreadyMonitoring {
		fmt.Printf("[ParkDog] Policy '%s' already being monitored\n", policyName)
		return
	}

	desc := fmt.Sprintf("Watching policy %s with %d functions\n", policyName, len(policy.GetFunc()))
	t := rumpaint.Card("ParkDog", desc)
	log.Println(t)
	rd.progress[policyName] = NewProgress()

	rd.reports[policyName] = &WatchdogReport{
		PolicyName:     policyName,
		StartTime:      time.Now(),
		TimeLimit:      policy.GetBase(),
		FailureReasons: make([]string, 0),
	}

	rd.wg.Add(1)
	go func() {
		defer rd.wg.Done()

		// Start time tracking
		startTime := time.Now()
		rd.recordStart(policyName, startTime)

		if p, ok := rd.progress[policyName]; ok {
			p.IsRunning = true
		}
	}()
	monitor := NewMonitorPolicy(rd.ctx)
	rd.monitors[policyName] = monitor
	if err := monitor.Monitor(policyName, func() { rd.tickSinglePolicy(policyName) }, rd.Settings.ShutdownTimeout); err != nil {
		delete(rd.monitors, policyName)
		return
	}

}

// tickSinglePolicy checks timeout for a specific policy
func (rd *Dog[T]) tickSinglePolicy(policyName string) {
	// rd.mu.Lock()
	// defer rd.mu.Unlock()

	progress, pExists := rd.progress[policyName]
	policy, policyExists := rd.policy[policyName]

	if !pExists || !policyExists || !progress.IsRunning || progress.StartedAtNano == 0 {
		return
	}

	start := time.Unix(0, progress.StartedAtNano)
	duration := time.Since(start)
	timeout := policy.GetBase()

	var percent uint64
	if timeout > 0 {

		percent = uint64((float64(duration) / float64(timeout)) * 100)
	}

	if duration >= timeout {
		rd.processBark(IBark{
			Reason:   "Timeout Exceeded",
			Policy:   policyName,
			Time:     time.Now(),
			Duration: duration,
		})
	} else {
		rd.updateProgress(policyName, percent)
	}
}

// processDone triggers the writes the report & triggers pakkun
func (rd *Dog[T]) processDone(done IDone) {
	rd.mu.Lock()
	defer rd.mu.Unlock()

	policyName := done.PolicyName

	policy, exists := rd.policy[policyName]
	if !exists {
		fmt.Printf("[processDone] ❌ Policy %s doesn't exist\n", policyName)
		return
	}

	var totalDuration time.Duration
	if progress, ok := rd.progress[policyName]; ok && progress.StartedAtNano > 0 {
		start := time.Unix(0, progress.StartedAtNano)
		totalDuration = time.Since(start)
	}

	// Determine which duration to use
	funcDuration := done.FuncDuration
	if funcDuration == 0 {
		funcDuration = totalDuration
	}

	rd.ensureReport(policyName, policy)

	if report, exists := rd.reports[policyName]; exists {
		report.addDuration(funcDuration)
		report.calcAvg()
		report.updateMin(funcDuration)
		report.updateMax(funcDuration)

		timeout := policy.GetBase()
		if funcDuration < timeout {
			report.pCount()
			report.sCount()
		} else {
			report.eCount()
			report.fCount()
		}

		report.exCount()
		report.updateSRate()
		report.setStatus(rd.calculateStatus(report))
		report.setEndTime(time.Now())
		report.isReady = true
		rd.chakra.KageBunshinNoJutsu(policyName, report)

	}

	policy.Call()
	timeout := policy.GetBase()
	if funcDuration < timeout {
		policy.Succeed.Call()
	} else {
		policy.Fail.Call()
	}

	var percent uint64
	if timeout > 0 {
		percent = uint64((float64(funcDuration) / float64(timeout)) * 100)
	}
	if percent > 100 {
		percent = 100
	}

	if progress, ok := rd.progress[policyName]; ok {
		progress.SetCompletion(time.Duration(percent))
		health := rd.genHealth(percent)
		progress.SetHealth(health)
		rd.health[policyName] = &health
	}

	rd.bin(done)

	if len(policy.Fn) > 0 {
		return
	}

	if p, ok := rd.progress[policyName]; ok {
		p.IsRunning = false
	}

	// if stopChan, exists := rd.tickers[policyName]; exists {
	// 	close(stopChan)
	// 	delete(rd.tickers, policyName)
	// }

	if monitor, exists := rd.monitors[done.PolicyName]; exists {
		monitor.Stop()
	}

	rd.durations[policyName] = totalDuration
}

// processBark handles errors
func (rd *Dog[T]) processBark(bark IBark) {
	rd.mu.Lock()
	defer rd.mu.Unlock()

	policy, exists := rd.policy[bark.Policy]

	if !exists {
		return
	}

	if p, ok := rd.progress[bark.Policy]; ok {
		p.IsRunning = false
	}

	// Stop ticker
	// if stopChan, exists := rd.tickers[bark.Policy]; exists {
	// 	close(stopChan)
	// 	delete(rd.tickers, bark.Policy)
	// }
	if monitor, exists := rd.monitors[bark.Policy]; exists {
		monitor.Stop()
	}
	policy.Call()
	policy.Fail.Call()

	health := Health{Danger: true, IsHealthy: false}
	rd.health[bark.Policy] = &health

	rd.ensureReport(bark.Policy, policy)

	if report, exists := rd.reports[bark.Policy]; exists {
		report.fCount()
		report.pushFailureReason(bark.Reason)
		e := fmt.Errorf("%s: %s", bark.Policy, bark.Reason)
		report.setLastError(e)
		report.updateSRate()
		report.setStatus(rd.calculateStatus(report))
		report.isReady = true
		rd.chakra.KageBunshinNoJutsu(bark.Policy, report)
	}

	fmt.Printf("[Bark] Policy '%s': ERROR - %s (duration: %v)\n",
		bark.Policy, bark.Reason, bark.Duration)
}

// ensureReport inits the report
func (rd *Dog[T]) ensureReport(policyName string, policy *Policy[T]) {

	if _, exists := rd.reports[policyName]; !exists {
		x := &WatchdogReport{
			PolicyName:     policyName,
			StartTime:      time.Now(),
			TimeLimit:      policy.GetBase(),
			FailureReasons: make([]string, 0),
			MinDuration:    time.Duration(int64(^uint64(0) >> 1)), // Max int64
		}
		rd.reports[policyName] = x
	}

}
