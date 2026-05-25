// package dog

// import (
// 	"sync"
// 	"sync/atomic"
// 	"time"
// )

// // WatchdogReport is the final report for a policy
// type WatchdogReport struct {
// 	mu                                                   sync.Mutex
// 	PolicyName                                           string
// 	StartTime, EndTime                                   time.Time
// 	ExecutionCount                                       atomic.Int64
// 	PassedCount                                          atomic.Int64 // Completed within timeout
// 	ExceededCount                                        atomic.Int64 // Exceeded timeout
// 	isReady                                              bool
// 	TotalDuration, AvgDuration, MinDuration, MaxDuration time.Duration
// 	SuccessCount, FailureCount                           atomic.Int64
// 	SuccessRate                                          float64
// 	TimeLimit                                            time.Duration
// 	FailureReasons                                       []string
// 	LastError                                            error
// 	Status                                               string // "healthy", "warning", "critical"
// 	Output                                               []byte
// }

// func (w *WatchdogReport) IsReady() bool {
// 	return w.isReady
// }

// func (w *WatchdogReport) exCount() {
// 	w.ExecutionCount.Add(1)
// }

// func (w *WatchdogReport) eCount() {

// 	w.ExecutionCount.Add(1)
// 	w.ExceededCount.Add(1)
// }

// // fCount increments the failure count
// func (w *WatchdogReport) fCount() {

// 	w.ExecutionCount.Add(1)
// 	w.FailureCount.Add(1)
// }

// // pCount increments the passed count
// func (w *WatchdogReport) pCount() {

// 	w.PassedCount.Add(1)
// }

// // sCount increments the success count
// func (w *WatchdogReport) sCount() {

// 	w.SuccessCount.Add(1)
// }

// // addDuration adds a duration to total
// func (w *WatchdogReport) addDuration(dur time.Duration) {

// 	w.TotalDuration += dur
// }

// // calcAvg calculates average duration
// func (w *WatchdogReport) calcAvg() {

// 	if w.ExecutionCount.Load() > 0 {
// 		w.AvgDuration = w.TotalDuration / time.Duration(w.ExecutionCount.Load())
// 	}
// }

// // updateMin updates the minimum duration if applicable
// func (w *WatchdogReport) updateMin(duration time.Duration) {

// 	if duration < w.MinDuration || w.MinDuration == 0 {
// 		w.MinDuration = duration
// 	}
// }

// // updateMax updates the maximum duration if applicable
// func (w *WatchdogReport) updateMax(duration time.Duration) {

// 	if duration > w.MaxDuration {
// 		w.MaxDuration = duration
// 	}
// }

// // updateSRate calculates and updates the success rate
// func (w *WatchdogReport) updateSRate() {

// 	if w.ExecutionCount.Load() > 0 {
// 		w.SuccessRate = float64(w.SuccessCount.Load()) / float64(w.ExecutionCount.Load())
// 	}
// }

// // setEndTime sets the end time of the report period
// func (w *WatchdogReport) setEndTime(e time.Time) {

// 	w.EndTime = e
// }

// // setStatus sets the health status
// func (w *WatchdogReport) setStatus(status string) {

// 	w.Status = status
// }

// // pushFailureReason adds a failure reason to the list
// func (w *WatchdogReport) pushFailureReason(reason string) {

// 	w.FailureReasons = append(w.FailureReasons, reason)
// }

// // setLastError sets the last error encountered
// func (w *WatchdogReport) setLastError(err error) {

// 	w.LastError = err
// }

// // getPassedProgress returns the percentage of passed executions
// func (w *WatchdogReport) getPassedProgress() float64 {

// 	if w.ExecutionCount.Load() == 0 {
// 		return 0
// 	}
// 	return float64(w.PassedCount.Load()) / float64(w.ExecutionCount.Load()) * 100
// }

// // getExceedProgress returns the percentage of exceeded executions
// func (w *WatchdogReport) getExceedProgress() float64 {

// 	if w.ExecutionCount.Load() == 0 {
// 		return 0
// 	}
// 	return float64(w.ExceededCount.Load()) / float64(w.ExecutionCount.Load()) * 100
// }

// func (w *WatchdogReport) getReport() *WatchdogReport {
// 	if w == nil {
// 		panic("nill report")
// 	}
// 	return w
// }

// func (w *WatchdogReport) resetAll() {
// 	w.ExecutionCount.Store(0)
// 	w.PassedCount.Store(0)
// 	w.ExceededCount.Store(0)
// }

package dog

import (
	"encoding/json"
	"fmt"
	"log"
	rumpaint "rum/app/paint"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// WatchdogReport is the final report for a policy execution
type WatchdogReport struct {
	mu                                                   sync.Mutex
	PolicyName                                           string
	StartTime, EndTime                                   time.Time
	ExecutionCount                                       atomic.Int64
	PassedCount                                          atomic.Int64
	ExceededCount                                        atomic.Int64
	isReady                                              bool
	TotalDuration, AvgDuration, MinDuration, MaxDuration time.Duration
	SuccessCount, FailureCount                           atomic.Int64
	SuccessRate                                          float64
	TimeLimit                                            time.Duration
	FailureReasons                                       []string
	LastError                                            error
	Status                                               string
	Output                                               []byte
	Metrics                                              *SystemMetrics
}

// Report metadata methods

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

func (w *WatchdogReport) fCount() {
	w.ExecutionCount.Add(1)
	w.FailureCount.Add(1)
}

func (w *WatchdogReport) pCount() {
	w.PassedCount.Add(1)
}

func (w *WatchdogReport) sCount() {
	w.SuccessCount.Add(1)
}

// Duration tracking methods
func (w *WatchdogReport) addDuration(dur time.Duration) {
	w.TotalDuration += dur
}

func (w *WatchdogReport) calcAvg() {
	if w.ExecutionCount.Load() > 0 {
		w.AvgDuration = w.TotalDuration / time.Duration(w.ExecutionCount.Load())
	}
}

func (w *WatchdogReport) updateMin(duration time.Duration) {
	if duration < w.MinDuration || w.MinDuration == 0 {
		w.MinDuration = duration
	}
}

func (w *WatchdogReport) updateMax(duration time.Duration) {
	if duration > w.MaxDuration {
		w.MaxDuration = duration
	}
}

func (w *WatchdogReport) updateSRate() {
	if w.ExecutionCount.Load() > 0 {
		w.SuccessRate = float64(w.SuccessCount.Load()) / float64(w.ExecutionCount.Load())
	}
}

// Status methods
func (w *WatchdogReport) setEndTime(e time.Time) {
	w.EndTime = e
}

func (w *WatchdogReport) setStatus(status string) {
	w.Status = status
}

func (w *WatchdogReport) pushFailureReason(reason string) {
	w.FailureReasons = append(w.FailureReasons, reason)
}

func (w *WatchdogReport) setLastError(err error) {
	w.LastError = err
}

// Progress methods
func (w *WatchdogReport) getPassedProgress() float64 {
	if w.ExecutionCount.Load() == 0 {
		return 0
	}
	return float64(w.PassedCount.Load()) / float64(w.ExecutionCount.Load()) * 100
}

func (w *WatchdogReport) getExceedProgress() float64 {
	if w.ExecutionCount.Load() == 0 {
		return 0
	}
	return float64(w.ExceededCount.Load()) / float64(w.ExecutionCount.Load()) * 100
}

func (w *WatchdogReport) resetAll() {
	w.ExecutionCount.Store(0)
	w.PassedCount.Store(0)
	w.ExceededCount.Store(0)
}

// FormattedReport generates a pretty-printed report
type FormattedReport struct {
	Report *WatchdogReport
}

// Display showscase the report in terminal gui
func (fr *FormattedReport) Display() {
	if fr.Report == nil {
		return
	}

	r := fr.Report
	// var sb strings.Builder

	// Header
	// sb.WriteString("\n")
	// sb.WriteString("╔════════════════════════════════════════════════════════════════╗\n")
	// sb.WriteString(fmt.Sprintf("║  🐕 WATCHDOG REPORT: %s\n", padRight(r.PolicyName, 45)))
	// sb.WriteString("╠════════════════════════════════════════════════════════════════╣\n")

	// Basic Info
	// sb.WriteString("║ EXECUTION INFO:\n")
	// sb.WriteString(fmt.Sprintf("║   • Policy Name:      %s\n", r.PolicyName))
	// sb.WriteString(fmt.Sprintf("║   • Status:           %s\n", getStatusEmoji(r.Status)+r.Status))
	// sb.WriteString(fmt.Sprintf("║   • Start Time:       %s\n", r.StartTime.Format("2006-01-02 15:04:05")))
	// sb.WriteString(fmt.Sprintf("║   • End Time:         %s\n", r.EndTime.Format("2006-01-02 15:04:05")))
	// sb.WriteString(fmt.Sprintf("║   • Duration:         %v\n", r.EndTime.Sub(r.StartTime)))
	// sb.WriteString("║\n")
	headers := []string{"policy-name", "status", "start-time", "end-time", "duration"}
	data := [][]string{{r.PolicyName, r.Status, r.StartTime.Format("2006-01-02 15:04:05"), r.EndTime.Format("2006-01-02 15:04:05"), r.EndTime.Sub(r.StartTime).String()}}
	t := rumpaint.Table("execution info", headers, data)
	log.Println(t)

	// Execution Statistics
	// sb.WriteString("║ EXECUTION STATISTICS:\n")
	totalExec := r.ExecutionCount.Load()
	passedExec := r.PassedCount.Load()
	exceededExec := r.ExceededCount.Load()
	failureExec := r.FailureCount.Load()

	// sb.WriteString(fmt.Sprintf("║   • Total Executions:%s %d\n", padLeft("", 10), totalExec))
	// sb.WriteString(fmt.Sprintf("║   • Passed (< limit): %s %d (%.1f%%)\n",
	// 	padLeft("", 10), passedExec, r.getPassedProgress()))
	// sb.WriteString(fmt.Sprintf("║   • Exceeded (≥ limit):%s %d (%.1f%%)\n",
	// 	padLeft("", 9), exceededExec, r.getExceedProgress()))
	// sb.WriteString(fmt.Sprintf("║   • Failed:           %s %d (%.1f%%)\n",
	// 	padLeft("", 10), failureExec, (1-r.SuccessRate)*100))
	// sb.WriteString(fmt.Sprintf("║   • Success Rate:     %s %.2f%%\n",
	// 	padLeft("", 10), r.SuccessRate*100))
	// sb.WriteString("║\n")
	headers = []string{"total-ex", fmt.Sprintf("passed (<limit) %d", passedExec), fmt.Sprintf("exceeded (≥limit) %d", exceededExec), fmt.Sprintf("failed (%d)", failureExec), "success-rate"}
	data = [][]string{
		{
			fmt.Sprintf("%d", totalExec),
			fmt.Sprintf("%.1f%%", r.getPassedProgress()),
			fmt.Sprintf("%.1f%%", r.getExceedProgress()),
			fmt.Sprintf("%d", failureExec),
			fmt.Sprintf("%.2f%%", (1-r.SuccessRate)*100),
		},
	}
	t = rumpaint.Table("execution statistics", headers, data)
	log.Println(t)
	// Time Analysis
	// sb.WriteString("║ TIME ANALYSIS:\n")
	// sb.WriteString(fmt.Sprintf("║   • Time Limit:       %s %v\n", padLeft("", 10), r.TimeLimit))
	// sb.WriteString(fmt.Sprintf("║   • Min Duration:     %s %v\n", padLeft("", 10), r.MinDuration))
	// sb.WriteString(fmt.Sprintf("║   • Max Duration:     %s %v\n", padLeft("", 10), r.MaxDuration))
	// sb.WriteString(fmt.Sprintf("║   • Avg Duration:     %s %v\n", padLeft("", 10), r.AvgDuration))
	// sb.WriteString(fmt.Sprintf("║   • Total Duration:   %s %v\n", padLeft("", 10), r.TotalDuration))
	// sb.WriteString("║\n")
	headers = []string{"time-limit", "min-duration", "max-duration", "avg-duration", "total-duration"}
	data = [][]string{{r.TimeLimit.String(), r.MinDuration.String(), r.MaxDuration.String(), r.AvgDuration.String(), r.TotalDuration.String()}}
	t = rumpaint.Table("time analysis", headers, data)
	log.Println(t)
	// System Metrics
	if r.Metrics != nil {
		snapshot := r.Metrics.GetSnapshot()

		// sb.WriteString("║ SYSTEM METRICS:\n")
		// sb.WriteString("║   CPU:\n")
		// sb.WriteString("║   CPU Name:    " + snapshot.CPUName + "\n")
		// sb.WriteString(fmt.Sprintf("║     • Usage:         %.1f%%\n", snapshot.CPUUsage))
		// sb.WriteString(fmt.Sprintf("║     • Temperature:   %.1f°C\n", snapshot.CPUTemp))
		// sb.WriteString(fmt.Sprintf("║     • Health:        %s\n", snapshot.GetCPUHealth()))
		// sb.WriteString("║   Memory:\n")
		// sb.WriteString(fmt.Sprintf("║     • Alloc:         %.2f MB\n", snapshot.AllocMB))
		// sb.WriteString(fmt.Sprintf("║     • Max Seen:      %.2f MB\n", snapshot.MaxMemorySeenMB))
		// sb.WriteString(fmt.Sprintf("║     • Percent:       %.1f%%\n", snapshot.MemoryPercent))
		// sb.WriteString(fmt.Sprintf("║     • Health:        %s\n", snapshot.GetMemoryHealth()))
		// sb.WriteString("║   GPU:\n")
		// sb.WriteString(fmt.Sprintf("║     • Name:          %s\n", snapshot.GPUName))
		// sb.WriteString(fmt.Sprintf("║     • Usage:         %.1f%%\n", snapshot.GPUUsage))
		// sb.WriteString(fmt.Sprintf("║     • Temperature:   %.1f°C\n", snapshot.GPUTemp))
		// sb.WriteString(fmt.Sprintf("║     • Health:        %s\n", snapshot.GetGPUHealth()))
		// sb.WriteString("║   Thermal:\n")
		// sb.WriteString(fmt.Sprintf("║     • Level:         %s\n", snapshot.ThermalLevel))
		// sb.WriteString(fmt.Sprintf("║     • Throttled:     %v\n", snapshot.CPUThrottled))
		// sb.WriteString(fmt.Sprintf("║   Goroutines:       %d\n", snapshot.GoroutineCount))
		// sb.WriteString("║\n")
		headers = []string{"cpu-name", "usage", "temparature", "health"}
		data = [][]string{{snapshot.CPUName, fmt.Sprintf("%.1f%%", snapshot.CPUUsage), fmt.Sprintf("%.1f°C", snapshot.CPUTemp), snapshot.GetCPUHealth()}}
		t = rumpaint.Table("cpu", headers, data)
		log.Println(t)
		headers = []string{"alloc", "max-seen", "percent", "health"}
		data = [][]string{{fmt.Sprintf("%.2f MB", snapshot.AllocMB), fmt.Sprintf("%.2f MB", snapshot.MaxMemorySeenMB), fmt.Sprintf("%.1f%%", snapshot.MemoryPercent), snapshot.GetMemoryHealth()}}
		t = rumpaint.Table("memory", headers, data)
		log.Println(t)
		headers = []string{"name", "usage", "temparature", "health"}
		data = [][]string{{snapshot.GPUName, fmt.Sprintf("%.1f%%", snapshot.GPUUsage), fmt.Sprintf("%.1f°C", snapshot.GPUTemp), snapshot.GetGPUHealth()}}
		t = rumpaint.Table("gpu", headers, data)
		log.Println(t)
		headers = []string{"level", "throttled"}
		data = [][]string{{snapshot.ThermalLevel, fmt.Sprintf("%v", snapshot.CPUThrottled)}}
		t = rumpaint.Table("thermal", headers, data)
		log.Println(t)
		headers = []string{"goroutines"}
		data = [][]string{{fmt.Sprintf("%d", snapshot.GoroutineCount)}}
		t = rumpaint.Table("goroutines", headers, data)
		log.Println(t)
	}

	// Failure Reasons
	if len(r.FailureReasons) > 0 {
		reasons := []string{}
		for i, reason := range r.FailureReasons {
			reasons = append(reasons, fmt.Sprintf("   %d. %s\n", i+1, reason))
		}
		t = rumpaint.Table("failure reasons", []string{"reason"}, [][]string{reasons})
		log.Println(t)
	}

	// Footer
	// sb.WriteString("╚════════════════════════════════════════════════════════════════╝\n")

	// return sb.String()
}

// Helper functions for formatting
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func padLeft(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}

func getStatusEmoji(status string) string {
	switch status {
	case "healthy":
		return "✅ "
	case "warning":
		return "⚠️  "
	case "critical":
		return "🔴 "
	case "pending":
		return "⏳ "
	default:
		return "❓ "
	}
}

// GenerateReport generates and logs a formatted report
func (rd *Dog[T]) generateReport(name string) {
	rd.mu.RLock()
	report, exists := rd.reports[name]
	rd.mu.RUnlock()

	if !exists {
		log.Printf("❌ No report found for policy '%s'", name)
		return
	}

	formatted := &FormattedReport{Report: report}
	formatted.Display()

}

// Helper for serializing outputs
func serializeOutput(resp interface{}) []byte {
	// Try to serialize as JSON
	data, err := json.Marshal(resp)
	if err == nil {
		return data
	}

	// Fallback to string representation
	return []byte(fmt.Sprintf("%v", resp))
}
