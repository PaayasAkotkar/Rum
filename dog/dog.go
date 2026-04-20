// Package dog implements the timeout stragety management
// note: without unique rank id it wont work
package dog

import (
	"context"
	"fmt"
	"log"
	rumchakra "rum/app/chakra"
	rumpaint "rum/app/paint"
	"strconv"
	"strings"
	"sync"
	"time"

	"charm.land/lipgloss/v2"
)

// Dog provides a robust and clean timeout management system
type Dog[T any] struct {
	mu sync.RWMutex

	// Core configuration
	base time.Duration // Base fallback timeout

	chakra *rumchakra.Chakra[WatchdogReport]

	// Policy management
	policy map[string]*Policy[T] // name -> ...

	// Duration tracking per policy
	durations map[string]time.Duration // policy -> latest duration

	// Progress per policy
	progress map[string]*ExeProgress // policy -> ...

	// Health per policy
	health map[string]*Health // policy -> ...

	// Report storage
	reports map[string]*WatchdogReport // policy -> ...

	register   chan *Policy[T] // Register new policy
	unregister chan string     // Unregister policy
	parkDog    chan string     // Start watching policy
	done       chan IDone      // Signal function done
	bark       chan IBark      // Error occurred
	reset      chan string     // Reset policy stats
	resetAll   chan bool       // Reset all policies
	stopCh     chan struct{}   // Stop watchdog
	doneCh     chan struct{}

	// Dedicated fast tickers per active policy
	// tickers map[string]chan struct{}
	monitors map[string]*MonitorPolicy

	// Settings
	Settings *Settings

	// Control & Context
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
}

// IDone represents a completed function execution with unique tracking
type IDone struct {
	PolicyName   string        // Policy identifier
	FuncName     string        // Function name
	Rank         int           // Function rank
	ExecutionID  string        // OPTIONAL: Unique execution ID
	FuncDuration time.Duration // Client-measured duration for this func
}

// IBark represents an error/event
type IBark struct {
	Reason   string
	Policy   string
	Time     time.Time
	Duration time.Duration
}

// New creates a new Dog instance with base timeout
func New[T any](base time.Duration) *Dog[T] {
	return NewWithContext[T](context.Background(), base)
}

// NewWithContext creates a new Dog with context support
func NewWithContext[T any](ctx context.Context, base time.Duration) *Dog[T] {
	if base == 0 {
		panic("base timeout cannot be 0")
	}

	ctx, cancel := context.WithCancel(ctx)

	rd := &Dog[T]{
		base:      base,
		policy:    make(map[string]*Policy[T]),
		durations: make(map[string]time.Duration),
		progress:  make(map[string]*ExeProgress),
		health:    make(map[string]*Health),
		reports:   make(map[string]*WatchdogReport),
		// tickers:    make(map[string]chan struct{}),
		monitors:   make(map[string]*MonitorPolicy),
		register:   make(chan *Policy[T], 100),
		unregister: make(chan string, 100),
		parkDog:    make(chan string, 100),
		done:       make(chan IDone, 1000),
		bark:       make(chan IBark, 1000),
		reset:      make(chan string, 100),
		resetAll:   make(chan bool, 10),
		stopCh:     make(chan struct{}),
		doneCh:     make(chan struct{}),
		chakra:     rumchakra.New[WatchdogReport](),
		Settings:   DefaultSettings(),
		ctx:        ctx,
		cancel:     cancel,
	}
	t := rumpaint.Header(`

██████╗░░█████╗░░██████╗░
██╔══██╗██╔══██╗██╔════╝░
██║░░██║██║░░██║██║░░██╗░
██║░░██║██║░░██║██║░░╚██╗
██████╔╝╚█████╔╝╚██████╔╝
╚═════╝░░╚════╝░░╚═════╝░
			`)
	log.Println(t)

	return rd
}

func (rd *Dog[T]) Watch() {
	rd.wg.Add(1)

	go rd.watchDog()
}

// registerPolicy adds a new policy
func (rd *Dog[T]) registerPolicy(policy *Policy[T]) {
	rd.mu.Lock()
	defer rd.mu.Unlock()

	rd.policy[policy.Name] = policy
	rd.progress[policy.Name] = NewProgress()
	rd.health[policy.Name] = &Health{}

	title := "Registration Succeed 😄"
	desc := []string{
		fmt.Sprintf("Name: %s", policy.Name),
		fmt.Sprintf("Funcs to track: %d", len(policy.GetFunc())),
	}

	for _, fn := range policy.GetFunc() {
		desc = append(desc, fmt.Sprintf("- %s (rank: %d)", fn.Name, fn.Rank))
	}
	t := rumpaint.Card(title, strings.Join(desc, ", "))
	log.Println(t)
}

// unregisterPolicy removes a policy
func (rd *Dog[T]) unregisterPolicy(name string) {
	rd.mu.Lock()
	defer rd.mu.Unlock()

	// if stopChan, exists := rd.tickers[name]; exists {
	// 	close(stopChan)
	// 	delete(rd.tickers, name)
	// }

	if _, exists := rd.policy[name]; exists {
		delete(rd.policy, name)
		delete(rd.progress, name)
		delete(rd.health, name)
		delete(rd.reports, name)
		fmt.Printf("[Unregister] Policy '%s' cleaned up\n", name)
	}
}

// getTimeFloat returns the conversion based on settings
func (rd *Dog[T]) getTimeFloat(t time.Duration) IConv {
	switch true {
	case rd.Settings.ConvDurationInHours:
		return durationToHour(t)
	case rd.Settings.ConvDurationInMins:
		return durationToMin(t)
	case rd.Settings.ConvDurationInSecs:
		return durationToSec(t)
	case rd.Settings.ConvDurationInMiliSecs:
		return durationToMS(t)
	}
	return IConv{Conv: 0, Unit: "nil"}
}

// GenerateReport generates a formatted report for a policy
func (rd *Dog[T]) generateReport(name string) {

	report, exists := rd.reports[name]
	_, policyExists := rd.policy[name]

	t := rumpaint.Title("RUM WATCH DOG REPORT")
	log.Println(t)
	if !exists || !policyExists {
		log.Printf("No report found for policy '%s'", name)
		t := rumpaint.Title("REPORT DONE 😃")
		log.Println(t)
		return
	}

	if rd.Settings == nil {
		rd.Settings = DefaultSettings()
	}

	headers := []string{"ID", "Policy Name", "Time Limit", "Status", "Report Period"}
	data := [][]string{
		{
			"1",
			report.PolicyName,
			fmt.Sprintf("%v %s", strconv.FormatFloat(rd.getTimeFloat(report.TimeLimit).Conv, 'f', 2, 64), rd.getTimeFloat(report.TimeLimit).Unit),
			strings.ToUpper(report.Status),
			report.EndTime.String(),
		},
	}
	tx := rumpaint.Table("report", headers, data)
	lipgloss.Println(tx)
	headers = []string{"ID", "Total Executions", "Passed (< limit)", "Exceeded (≥ limit)", "Success Rate", "Failure Rate"}
	data = [][]string{
		{
			"1",
			strconv.FormatInt(report.ExecutionCount.Load(), 10),
			strconv.FormatFloat(report.getPassedProgress(), 'f', 2, 64) + "%",
			strconv.FormatFloat(report.getExceedProgress(), 'f', 2, 64) + "%",
			strconv.FormatFloat(report.SuccessRate*100, 'f', 2, 64) + "%",
			strconv.FormatFloat((1-report.SuccessRate)*100, 'f', 2, 64) + "%",
		},
	}
	tx = rumpaint.Table("Execution Statistics", headers, data)
	lipgloss.Println(tx)

	headers = []string{"ID", "Limit", "Min Duration", "Max Duration", "Avg Duration", "Total Duration"}
	data = [][]string{
		{
			"1",
			fmt.Sprintf("%v %s", strconv.FormatFloat(rd.getTimeFloat(report.TimeLimit).Conv, 'f', 2, 64), rd.getTimeFloat(report.TimeLimit).Unit),
			fmt.Sprintf("%v %s", strconv.FormatFloat(rd.getTimeFloat(report.MinDuration).Conv, 'f', 2, 64), rd.getTimeFloat(report.MinDuration).Unit),
			fmt.Sprintf("%v %s", strconv.FormatFloat(rd.getTimeFloat(report.MaxDuration).Conv, 'f', 2, 64), rd.getTimeFloat(report.MaxDuration).Unit),
			fmt.Sprintf("%v %s", strconv.FormatFloat(rd.getTimeFloat(report.AvgDuration).Conv, 'f', 2, 64), rd.getTimeFloat(report.AvgDuration).Unit),
			fmt.Sprintf("%v %s", strconv.FormatFloat(rd.getTimeFloat(report.TotalDuration).Conv, 'f', 2, 64), rd.getTimeFloat(report.TotalDuration).Unit),
		},
	}
	tx = rumpaint.Table("Time Analysis", headers, data)
	lipgloss.Println(tx)

	if len(report.FailureReasons) > 0 {
		title := "Failure Reasons"
		reasons := strings.Join(report.FailureReasons, ", ")
		t := rumpaint.Card(title, reasons)
		log.Println(t)
	}

	t = rumpaint.Title("REPORT DONE 😃")
	log.Println(t)
}
