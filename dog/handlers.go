package dog

import (
	"fmt"
	"log"
	"time"
)

// handleRegister processes policy registration
func (rd *Dog[T]) handleRegister(policy *Policy[T]) {
	// note: its been really important to focus on registeration than pulling out a pocket
	rd.mu.Lock()
	defer rd.mu.Unlock()

	if _, exists := rd.policy[policy.Name]; exists {
		log.Printf("⚠️  Policy %s already registered", policy.Name)
		return
	}

	log.Printf("📝 Registering policy: %s", policy.Name)

	rd.policy[policy.Name] = policy

	rd.lifecycle[policy.Name] = &PolicyLifecycle{
		Name:      policy.Name,
		State:     StateRegistered,
		Timestamp: time.Now(),
	}

	rd.progress[policy.Name] = NewProgress()
	rd.health[policy.Name] = &Health{}
	rd.metrics[policy.Name] = NewSystemMetrics(len(rd.policy))
	rd.durations[policy.Name] = 0

	log.Printf("✅ Policy registered: %s (timeout: %v, funcs: %d)",
		policy.Name, policy.GetBase(), len(policy.GetFunc()))

	// Signal registration complete
	// this kind of solves the problem that its actaully done than simply not registerating & moving on
	select {
	case rd.registeredCh <- policy.Name:
	default:
	}
}

// handleUnregister processes policy unregistration
func (rd *Dog[T]) handleUnregister(name string) {
	rd.mu.Lock()
	defer rd.mu.Unlock()

	if _, exists := rd.policy[name]; !exists {
		log.Printf("⚠️  Cannot unregister: policy %s not found", name)
		return
	}

	// Stop monitor if running
	if monitor, exists := rd.monitors[name]; exists {
		monitor.Stop()
		delete(rd.monitors, name)
	}

	// Cleanup
	delete(rd.policy, name)
	delete(rd.lifecycle, name)
	delete(rd.progress, name)
	delete(rd.health, name)
	delete(rd.metrics, name)
	delete(rd.reports, name)

	log.Printf("🗑️  Policy unregistered: %s", name)
}

// handleParkDog processes monitoring activation
func (rd *Dog[T]) handleParkDog(policyName string) {
	rd.mu.Lock()
	policy, pExists := rd.policy[policyName]
	lifecycle, lExists := rd.lifecycle[policyName]
	rd.mu.Unlock()

	if !pExists || !lExists {
		log.Printf("❌ Cannot park: policy %s not found", policyName)
		return
	}

	if !lifecycle.IsInState(StateRegistered) {
		log.Printf("❌ Cannot park: policy %s not in registered state (current: %s)",
			policyName, lifecycle.GetState())
		return
	}

	// Transition to monitoring
	lifecycle.SetState(StateMonitoring)

	log.Printf("🚀 Parking dog on policy: %s (monitoring %d functions)",
		policyName, len(policy.GetFunc()))

	rd.mu.Lock()
	rd.progress[policyName] = NewProgress()
	rd.reports[policyName] = &WatchdogReport{
		PolicyName:     policyName,
		StartTime:      time.Now(),
		TimeLimit:      policy.GetBase(),
		FailureReasons: make([]string, 0),
	}
	rd.mu.Unlock()

	// Start monitoring
	rd.wg.Add(1)
	go func() {
		defer rd.wg.Done()

		rd.mu.Lock()
		startTime := time.Now()
		if progress, ok := rd.progress[policyName]; ok {
			progress.StartedAtNano = startTime.UnixNano()
			progress.IsRunning = true
		}
		rd.mu.Unlock()

		log.Printf("⏱️  Started monitoring policy: %s", policyName)
	}()

	// Create and start monitor
	monitor := NewMonitorPolicy(rd.ctx)
	rd.mu.Lock()
	rd.monitors[policyName] = monitor
	rd.mu.Unlock()

	if err := monitor.Monitor(policyName,
		func() { rd.tickSinglePolicy(policyName) },
		rd.Settings.ShutdownTimeout); err != nil {
		log.Printf("❌ Monitor error for %s: %v", policyName, err)
		rd.mu.Lock()
		delete(rd.monitors, policyName)
		rd.mu.Unlock()
	}
}

// handleDone processes successful function completion
func (rd *Dog[T]) handleDone(done IDone) {
	rd.mu.Lock()
	defer rd.mu.Unlock()

	policy, exists := rd.policy[done.PolicyName]
	if !exists {
		log.Printf("❌ Done: policy %s not found", done.PolicyName)
		return
	}

	lifecycle, lExists := rd.lifecycle[done.PolicyName]

	log.Printf("✅ Completed: %s.%s (rank: %d, duration: %v)",
		done.PolicyName, done.FuncName, done.Rank, done.FuncDuration)

	// Sample final metrics
	if metrics, ok := rd.metrics[done.PolicyName]; ok {
		metrics.Sample()
	}

	// Update report
	rd.ensureReport(done.PolicyName, policy)

	if report, exists := rd.reports[done.PolicyName]; exists {
		report.addDuration(done.FuncDuration)
		report.calcAvg()
		report.updateMin(done.FuncDuration)
		report.updateMax(done.FuncDuration)
		report.Output = done.Output

		timeout := policy.GetBase()
		if done.FuncDuration < timeout {
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

		// Attach metrics snapshot
		if metrics, ok := rd.metrics[done.PolicyName]; ok {
			report.Metrics = metrics.GetSnapshot()
		}

		rd.cheetah.Publish(done.PolicyName, report)
	}

	policy.Call()
	if done.FuncDuration < policy.GetBase() {
		policy.Succeed.Call()
	} else {
		policy.Fail.Call()
	}

	newFns := make([]Funcs[T], 0, len(policy.Fn))
	for _, f := range policy.Fn {
		if f.Rank != done.Rank {
			newFns = append(newFns, f)
		}
	}
	policy.Fn = newFns

	if len(policy.Fn) == 0 {
		if p, ok := rd.progress[done.PolicyName]; ok {
			p.IsRunning = false
		}

		if lExists {
			lifecycle.SetState(StateCompleted)
		}

		if monitor, exists := rd.monitors[done.PolicyName]; exists {
			monitor.Stop()
		}

		log.Printf("🎉 All functions completed for policy: %s", done.PolicyName)
	}
}

// handleBark processes errors
func (rd *Dog[T]) handleBark(bark IBark) {
	rd.mu.Lock()
	defer rd.mu.Unlock()

	policy, exists := rd.policy[bark.Policy]
	if !exists {
		log.Printf("❌ Bark: policy %s not found", bark.Policy)
		return
	}

	lifecycle, lExists := rd.lifecycle[bark.Policy]

	log.Printf("🚨 Error in policy %s: %s (duration: %v)",
		bark.Policy, bark.Reason, bark.Duration)

	// Update state
	if p, ok := rd.progress[bark.Policy]; ok {
		p.IsRunning = false
	}

	if lExists {
		lifecycle.SetState(StateError)
	}

	// Stop monitor
	if monitor, exists := rd.monitors[bark.Policy]; exists {
		monitor.Stop()
	}

	policy.Call()
	policy.Fail.Call()

	rd.health[bark.Policy] = &Health{Danger: true, IsHealthy: false}

	// Update report
	rd.ensureReport(bark.Policy, policy)

	if report, exists := rd.reports[bark.Policy]; exists {
		report.fCount()
		report.pushFailureReason(bark.Reason)
		report.setLastError(fmt.Errorf("%s: %s", bark.Policy, bark.Reason))
		report.updateSRate()
		report.setStatus(rd.calculateStatus(report))

		// Attach metrics
		if metrics, ok := rd.metrics[bark.Policy]; ok {
			report.Metrics = metrics.GetSnapshot()
		}

		report.isReady = true
		rd.cheetah.Publish(bark.Policy, report)
	}
}

// handleReset resets a policy
func (rd *Dog[T]) handleReset(policyName string) {
	rd.mu.Lock()
	defer rd.mu.Unlock()

	policy, exists := rd.policy[policyName]
	if !exists {
		log.Printf("⚠️  Cannot reset: policy %s not found", policyName)
		return
	}

	policy.Release()
	policy.Succeed.Release()
	policy.Fail.Release()

	rd.progress[policyName] = NewProgress()
	rd.reports[policyName] = &WatchdogReport{
		PolicyName:     policyName,
		StartTime:      time.Now(),
		TimeLimit:      policy.GetBase(),
		FailureReasons: make([]string, 0),
	}

	if lifecycle, ok := rd.lifecycle[policyName]; ok {
		lifecycle.SetState(StateRegistered)
	}

	log.Printf("🔄 Policy reset: %s", policyName)
}

// handleResetAll resets all policies
func (rd *Dog[T]) handleResetAll() {
	rd.mu.Lock()
	defer rd.mu.Unlock()

	for name, policy := range rd.policy {
		policy.Release()
		policy.Succeed.Release()
		policy.Fail.Release()

		rd.progress[name] = NewProgress()

		if lifecycle, ok := rd.lifecycle[name]; ok {
			lifecycle.SetState(StateRegistered)
		}

		log.Printf("🔄 Policy reset: %s", name)
	}
}
