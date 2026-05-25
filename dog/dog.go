// Package dog implements the timeout stragety management
// note: without unique rank id it wont work
package dog

import (
	"context"
	"fmt"
	"log"
	cheetah "rum/app/cheetah"
	rumpaint "rum/app/paint"
	"strings"
	"sync"
	"time"
)

// // Dog provides a robust and clean timeout management system
// type Dog[T any] struct {
// 	mu sync.RWMutex

// 	// Core configuration
// 	base time.Duration // Base fallback timeout

// 	cheetah  *rumcheetah .cheetah [WatchdogReport]

// 	// Policy management
// 	policy map[string]*Policy[T] // name -> ...

// 	// Duration tracking per policy
// 	durations map[string]time.Duration // policy -> latest duration

// 	// Progress per policy
// 	progress map[string]*ExeProgress // policy -> ...

// 	// Health per policy
// 	health map[string]*Health // policy -> ...

// 	// Report storage
// 	reports map[string]*WatchdogReport // policy -> ...

// 	register   chan *Policy[T] // Register new policy
// 	unregister chan string     // Unregister policy
// 	parkDog    chan string     // Start watching policy
// 	done       chan IDone      // Signal function done
// 	bark       chan IBark      // Error occurred
// 	reset      chan string     // Reset policy stats
// 	resetAll   chan bool       // Reset all policies
// 	stopCh     chan struct{}   // Stop watchdog
// 	doneCh     chan struct{}
// 	summonCh   chan string
// 	// Dedicated fast tickers per active policy
// 	// tickers map[string]chan struct{}
// 	monitors map[string]*MonitorPolicy

// 	// Settings
// 	Settings *Settings

// 	// Control & Context
// 	wg     sync.WaitGroup
// 	ctx    context.Context
// 	cancel context.CancelFunc
// 	once   sync.Once
// }

// IDone represents a completed function execution with unique tracking
// type IDone struct {
// 	PolicyName   string        // Policy identifier
// 	FuncName     string        // Function name
// 	Rank         int           // Function rank
// 	ExecutionID  string        // OPTIONAL: Unique execution ID
// 	FuncDuration time.Duration // Client-measured duration for this func
// 	Output       []byte
// }

// // IBark represents an error/event
// type IBark struct {
// 	Reason   string
// 	Policy   string
// 	Time     time.Time
// 	Duration time.Duration
// }

// // New creates a new Dog instance with base timeout
// func New[T any](base time.Duration) *Dog[T] {
// 	return NewWithContext[T](context.Background(), base)
// }

// // NewWithContext creates a new Dog with context support
// func NewWithContext[T any](ctx context.Context, base time.Duration) *Dog[T] {
// 	if base == 0 {
// 		panic("base timeout cannot be 0")
// 	}

// 	ctx, cancel := context.WithCancel(ctx)

// 	rd := &Dog[T]{
// 		base:      base,
// 		policy:    make(map[string]*Policy[T]),
// 		summonCh:  make(chan string, 100),
// 		durations: make(map[string]time.Duration),
// 		progress:  make(map[string]*ExeProgress),
// 		health:    make(map[string]*Health),
// 		reports:   make(map[string]*WatchdogReport),
// 		// tickers:    make(map[string]chan struct{}{}),
// 		monitors:   make(map[string]*MonitorPolicy),
// 		register:   make(chan *Policy[T], 100),
// 		unregister: make(chan string, 100),
// 		parkDog:    make(chan string, 100),
// 		done:       make(chan IDone, 1000),
// 		bark:       make(chan IBark, 1000),
// 		reset:      make(chan string, 100),
// 		resetAll:   make(chan bool, 10),
// 		stopCh:     make(chan struct{}),
// 		doneCh:     make(chan struct{}),
// 		cheetah :     rumcheetah .New[WatchdogReport](),
// 		Settings:   DefaultSettings(),
// 		ctx:        ctx,
// 		cancel:     cancel,
// 	}
// 	t := rumpaint.Header(`

// ██████╗░░█████╗░░██████╗░
// ██╔══██╗██╔══██╗██╔════╝░
// ██║░░██║██║░░██║██║░░██╗░
// ██║░░██║██║░░██║██║░░╚██╗
// ██████╔╝╚█████╔╝╚██████╔╝
// ╚═════╝░░╚════╝░░╚═════╝░
// 			`)
// 	log.Println(t)

// 	return rd
// }

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

// // GenerateReport generates a formatted report for a policy
// func (rd *Dog[T]) generateReport(name string) {

// 	report, exists := rd.reports[name]
// 	_, policyExists := rd.policy[name]

// 	t := rumpaint.Title("RUM WATCH DOG REPORT")
// 	log.Println(t)
// 	if !exists || !policyExists {
// 		log.Printf("No report found for policy '%s'", name)
// 		t := rumpaint.Title("REPORT DONE 😃")
// 		log.Println(t)
// 		return
// 	}

// 	if rd.Settings == nil {
// 		rd.Settings = DefaultSettings()
// 	}

// 	headers := []string{"ID", "Policy Name", "Time Limit", "Status", "Report Period"}
// 	data := [][]string{
// 		{
// 			"1",
// 			report.PolicyName,
// 			fmt.Sprintf("%v %s", strconv.FormatFloat(rd.getTimeFloat(report.TimeLimit).Conv, 'f', 2, 64), rd.getTimeFloat(report.TimeLimit).Unit),
// 			strings.ToUpper(report.Status),
// 			report.EndTime.String(),
// 		},
// 	}
// 	tx := rumpaint.Table("report", headers, data)
// 	lipgloss.Println(tx)
// 	headers = []string{"ID", "Total Executions", "Passed (< limit)", "Exceeded (≥ limit)", "Success Rate", "Failure Rate"}
// 	data = [][]string{
// 		{
// 			"1",
// 			strconv.FormatInt(report.ExecutionCount.Load(), 10),
// 			strconv.FormatFloat(report.getPassedProgress(), 'f', 2, 64) + "%",
// 			strconv.FormatFloat(report.getExceedProgress(), 'f', 2, 64) + "%",
// 			strconv.FormatFloat(report.SuccessRate*100, 'f', 2, 64) + "%",
// 			strconv.FormatFloat((1-report.SuccessRate)*100, 'f', 2, 64) + "%",
// 		},
// 	}
// 	tx = rumpaint.Table("Execution Statistics", headers, data)
// 	lipgloss.Println(tx)

// 	headers = []string{"ID", "Limit", "Min Duration", "Max Duration", "Avg Duration", "Total Duration"}
// 	data = [][]string{
// 		{
// 			"1",
// 			fmt.Sprintf("%v %s", strconv.FormatFloat(rd.getTimeFloat(report.TimeLimit).Conv, 'f', 2, 64), rd.getTimeFloat(report.TimeLimit).Unit),
// 			fmt.Sprintf("%v %s", strconv.FormatFloat(rd.getTimeFloat(report.MinDuration).Conv, 'f', 2, 64), rd.getTimeFloat(report.MinDuration).Unit),
// 			fmt.Sprintf("%v %s", strconv.FormatFloat(rd.getTimeFloat(report.MaxDuration).Conv, 'f', 2, 64), rd.getTimeFloat(report.MaxDuration).Unit),
// 			fmt.Sprintf("%v %s", strconv.FormatFloat(rd.getTimeFloat(report.AvgDuration).Conv, 'f', 2, 64), rd.getTimeFloat(report.AvgDuration).Unit),
// 			fmt.Sprintf("%v %s", strconv.FormatFloat(rd.getTimeFloat(report.TotalDuration).Conv, 'f', 2, 64), rd.getTimeFloat(report.TotalDuration).Unit),
// 		},
// 	}
// 	tx = rumpaint.Table("Time Analysis", headers, data)
// 	lipgloss.Println(tx)

// 	if len(report.FailureReasons) > 0 {
// 		title := "Failure Reasons"
// 		reasons := strings.Join(report.FailureReasons, ", ")
// 		t := rumpaint.Card(title, reasons)
// 		log.Println(t)
// 	}

// 	t = rumpaint.Title("REPORT DONE 😃")
// 	log.Println(t)
// }

// PolicyState represents the lifecycle state of a policy
type PolicyState string

const (
	StateUnregistered PolicyState = "unregistered"
	StateRegistered   PolicyState = "registered"
	StateMonitoring   PolicyState = "monitoring"
	StateCompleted    PolicyState = "completed"
	StateError        PolicyState = "error"
)

// PolicyLifecycle tracks the state of a policy
type PolicyLifecycle struct {
	Name      string
	State     PolicyState
	mu        sync.RWMutex
	Timestamp time.Time
}

func (pl *PolicyLifecycle) GetState() PolicyState {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	return pl.State
}

func (pl *PolicyLifecycle) SetState(state PolicyState) {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	pl.State = state
	pl.Timestamp = time.Now()
}

func (pl *PolicyLifecycle) IsInState(state PolicyState) bool {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	return pl.State == state
}

// Dog provides a robust timeout management system
type Dog[T any] struct {
	mu sync.RWMutex

	// Core configuration
	base time.Duration

	cheetah *cheetah.Cheetah[WatchdogReport]

	// Policy management
	policy    map[string]*Policy[T]       // name -> policy
	lifecycle map[string]*PolicyLifecycle // name -> state tracker

	// Duration tracking per policy
	durations map[string]time.Duration

	// Progress per policy
	progress map[string]*ExeProgress

	// Health per policy
	health map[string]*Health

	// System metrics per policy
	metrics map[string]*SystemMetrics

	// Report storage
	reports map[string]*WatchdogReport

	// Channels
	register     chan *Policy[T]
	unregister   chan string
	parkDog      chan string
	done         chan IDone
	bark         chan IBark
	reset        chan string
	resetAll     chan bool
	stopCh       chan struct{}
	doneCh       chan struct{}
	summonCh     chan string
	registeredCh chan string // Signals when registration is complete

	// Monitors
	monitors map[string]*MonitorPolicy

	// Settings
	Settings *Settings

	// Control & Context
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
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
		base:         base,
		policy:       make(map[string]*Policy[T]),
		lifecycle:    make(map[string]*PolicyLifecycle),
		durations:    make(map[string]time.Duration),
		progress:     make(map[string]*ExeProgress),
		health:       make(map[string]*Health),
		metrics:      make(map[string]*SystemMetrics),
		reports:      make(map[string]*WatchdogReport),
		monitors:     make(map[string]*MonitorPolicy),
		register:     make(chan *Policy[T], 100),
		unregister:   make(chan string, 100),
		parkDog:      make(chan string, 100),
		done:         make(chan IDone, 1000),
		bark:         make(chan IBark, 1000),
		reset:        make(chan string, 100),
		resetAll:     make(chan bool, 10),
		stopCh:       make(chan struct{}),
		doneCh:       make(chan struct{}),
		summonCh:     make(chan string, 100),
		registeredCh: make(chan string, 100),
		cheetah:      cheetah.New[WatchdogReport](),
		Settings:     DefaultSettings(),
		ctx:          ctx,
		cancel:       cancel,
	}

	printHeader()
	return rd
}

// Helper function to print header
func printHeader() {
	header := `
██████╗░░█████╗░░██████╗░
██╔══██╗██╔══██╗██╔════╝░
██║░░██║██║░░██║██║░░██╗░
██║░░██║██║░░██║██║░░╚██╗
██████╔╝╚█████╔╝╚██████╔╝
╚═════╝░░╚════╝░░╚═════╝░
`
	log.Println(header)
}
