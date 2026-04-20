package dog

import (
	"sync"
	"sync/atomic"
	"time"
)

// WatchdogReport is the final report for a policy
type WatchdogReport struct {
	mu                                                   sync.Mutex
	PolicyName                                           string
	StartTime, EndTime                                   time.Time
	ExecutionCount                                       atomic.Int64
	PassedCount                                          atomic.Int64 // Completed within timeout
	ExceededCount                                        atomic.Int64 // Exceeded timeout
	isReady                                              bool
	TotalDuration, AvgDuration, MinDuration, MaxDuration time.Duration
	SuccessCount, FailureCount                           atomic.Int64
	SuccessRate                                          float64
	TimeLimit                                            time.Duration
	FailureReasons                                       []string
	LastError                                            error
	Status                                               string // "healthy", "warning", "critical"
}

func (w *WatchdogReport) IsReady() bool {
	return w.isReady
}

func (w *WatchdogReport) exCount() {
	w.ExecutionCount.Add(1)
}

func (w *WatchdogReport) eCount() {

	w.ExecutionCount.Add(1)
	w.ExceededCount.Add(1)
}

// fCount increments the failure count
func (w *WatchdogReport) fCount() {

	w.ExecutionCount.Add(1)
	w.FailureCount.Add(1)
}

// pCount increments the passed count
func (w *WatchdogReport) pCount() {

	w.PassedCount.Add(1)
}

// sCount increments the success count
func (w *WatchdogReport) sCount() {

	w.SuccessCount.Add(1)
}

// addDuration adds a duration to total
func (w *WatchdogReport) addDuration(dur time.Duration) {

	w.TotalDuration += dur
}

// calcAvg calculates average duration
func (w *WatchdogReport) calcAvg() {

	if w.ExecutionCount.Load() > 0 {
		w.AvgDuration = w.TotalDuration / time.Duration(w.ExecutionCount.Load())
	}
}

// updateMin updates the minimum duration if applicable
func (w *WatchdogReport) updateMin(duration time.Duration) {

	if duration < w.MinDuration || w.MinDuration == 0 {
		w.MinDuration = duration
	}
}

// updateMax updates the maximum duration if applicable
func (w *WatchdogReport) updateMax(duration time.Duration) {

	if duration > w.MaxDuration {
		w.MaxDuration = duration
	}
}

// updateSRate calculates and updates the success rate
func (w *WatchdogReport) updateSRate() {

	if w.ExecutionCount.Load() > 0 {
		w.SuccessRate = float64(w.SuccessCount.Load()) / float64(w.ExecutionCount.Load())
	}
}

// setEndTime sets the end time of the report period
func (w *WatchdogReport) setEndTime(e time.Time) {

	w.EndTime = e
}

// setStatus sets the health status
func (w *WatchdogReport) setStatus(status string) {

	w.Status = status
}

// pushFailureReason adds a failure reason to the list
func (w *WatchdogReport) pushFailureReason(reason string) {

	w.FailureReasons = append(w.FailureReasons, reason)
}

// setLastError sets the last error encountered
func (w *WatchdogReport) setLastError(err error) {

	w.LastError = err
}

// getPassedProgress returns the percentage of passed executions
func (w *WatchdogReport) getPassedProgress() float64 {

	if w.ExecutionCount.Load() == 0 {
		return 0
	}
	return float64(w.PassedCount.Load()) / float64(w.ExecutionCount.Load()) * 100
}

// getExceedProgress returns the percentage of exceeded executions
func (w *WatchdogReport) getExceedProgress() float64 {

	if w.ExecutionCount.Load() == 0 {
		return 0
	}
	return float64(w.ExceededCount.Load()) / float64(w.ExecutionCount.Load()) * 100
}

func (w *WatchdogReport) getReport() *WatchdogReport {
	if w == nil {
		panic("nill report")
	}
	return w
}

func (w *WatchdogReport) resetAll() {
	w.ExecutionCount.Store(0)
	w.PassedCount.Store(0)
	w.ExceededCount.Store(0)
}
