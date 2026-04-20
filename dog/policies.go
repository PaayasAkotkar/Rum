package dog

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// Policy represents a timeout policy with its tracked functions
type Policy[T any] struct {
	Name      string        // Policy identifier
	Base      time.Duration // Base timeout limit
	Fn        []Funcs[T]    // Functions to track
	callsMade atomic.Int64  // Total calls
	Succeed   IPolicy       // Success tracking
	Fail      IPolicy       // Failure tracking
}

// NewPolicy creates a new policy
func NewPolicy[T any](base time.Duration) *Policy[T] {
	return &Policy[T]{
		Base:    base,
		Fn:      make([]Funcs[T], 0),
		Succeed: IPolicy{},
		Fail:    IPolicy{},
	}
}

func (p *Policy[T]) Continue() bool {

	return len(p.Fn) > 0
}

func (p *Policy[T]) SetName(name string) {
	p.Name = name
}

func (p *Policy[T]) GetName() string {
	return p.Name
}

func (p *Policy[T]) SetBase(base time.Duration) {
	p.Base = base
}

func (p *Policy[T]) GetBase() time.Duration {
	return p.Base
}

func (p *Policy[T]) AddFunc(fn Funcs[T]) {
	log.Printf("[Policy] func with name %s added to the monitor", fn.Name)
	p.Fn = append(p.Fn, fn)
}

func (p *Policy[T]) SetFunc(fns []Funcs[T]) {
	p.Fn = fns
}

func (p *Policy[T]) GetFunc() []Funcs[T] {
	return p.Fn
}

func (p *Policy[T]) Call() {
	p.callsMade.Add(1)
}

func (p *Policy[T]) TotalCalls() int64 {
	return p.callsMade.Load()
}

func (p *Policy[T]) Release() {
	p.callsMade.Store(0)
}

// Funcs represents a tracked function
type Funcs[T any] struct {
	Name string
	Rank int // Distinguisher for multiple calls
	Fn   func() (*T, error)
	Void func() error
}

// IPolicy tracks success or failure metrics
type IPolicy struct {
	callsMade atomic.Int64
	Reason    string
	TimeTaken time.Duration
}

func (i *IPolicy) Call() {
	i.callsMade.Add(1)
}

func (i *IPolicy) WriteReason(reason string) {
	i.Reason = reason
}

func (i *IPolicy) SetTimeTaken(t time.Duration) {
	i.TimeTaken = t
}

func (i *IPolicy) TotalCalls() int64 {
	return i.callsMade.Load()
}

func (i *IPolicy) Release() {
	i.callsMade.Store(0)
}

// Health represents health status
type Health struct {
	IsHealthy bool // true if healthy
	Silent    bool // true if all good (progress >= 75%)
	Mid       bool // true if in middle zone (30-75%)
	Danger    bool // true if in danger zone (< 30%)
}

// ExeProgress tracks real-time progress with FIXED timestamp field
type ExeProgress struct {
	StartedAtNano int64
	ToComplete    time.Duration // Percentage (0-100)
	Health        Health        // Current health status
	IsRunning     bool          // True if currently tracking
}

// NewProgress creates new progress tracker
func NewProgress() *ExeProgress {
	return &ExeProgress{
		StartedAtNano: 0,
		ToComplete:    0,
		Health:        Health{},
		IsRunning:     false,
	}
}

func (e *ExeProgress) SetCompletion(percent time.Duration) {
	if percent > 100 {
		percent = 100
	}
	e.ToComplete = percent
}

func (e *ExeProgress) GetCompletion() time.Duration {
	return e.ToComplete
}

func (e *ExeProgress) SetHealth(h Health) {
	e.Health = h
}

func (e *ExeProgress) GetHealth() Health {
	return e.Health
}

type MonitorPolicy struct {
	policy    string
	interval  time.Duration
	fn        func()
	stopChan  chan struct{}
	done      chan struct{}
	isRunning bool
	StopTime  time.Duration
	ctx       context.Context
	mu        sync.Mutex
	wg        sync.WaitGroup
}

func NewMonitorPolicy(ctx context.Context) *MonitorPolicy {
	return &MonitorPolicy{
		ctx:      ctx,
		stopChan: make(chan struct{}),
		done:     make(chan struct{}),
	}
}

func (m *MonitorPolicy) PushPolicy(name string) {
	m.policy = name
}

func (m *MonitorPolicy) Monitor(name string, fn func(), ti time.Duration) error {

	if m.IsRunning() {
		err := fmt.Errorf("policy %s already running", name)
		return err
	}

	if ti == 0 {
		return fmt.Errorf("interval cannot be %v", ti)
	}
	if fn == nil {
		return fmt.Errorf("no func to record")
	}

	m.interval = ti
	m.policy = name
	m.fn = fn

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.hub()
	}()

	return nil
}

func (m *MonitorPolicy) hub() error {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.fn()
		case <-m.stopChan:
			err := m.stop()
			return err
		case <-m.ctx.Done():
			m.isRunning = false
			return nil
		}
	}
}
func (m *MonitorPolicy) Stop() {
	var x struct{}
	m.stopChan <- x
}

func (m *MonitorPolicy) stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.isRunning {
		return nil
	}

	go func() {
		m.wg.Wait()
		close(m.done)
	}()
	t := m.StopTime
	if t == 0 {
		t = 1 * time.Second
	}
	select {
	case <-m.done:
		return nil
	case <-time.After(t):
		return fmt.Errorf("stop time exceeded for policy %s", m.policy)
	}
}

func (m *MonitorPolicy) SetStopTimeDuration(t time.Duration) {
	m.StopTime = t
}

func (m *MonitorPolicy) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isRunning
}
func (m *MonitorPolicy) GetPolicyName() string {
	return m.policy
}
func (m *MonitorPolicy) GetInterval() time.Duration {
	return m.interval
}

// Settings holds configuration
type Settings struct {
	EnableAvg bool // incase of base duration you can set avg of that

	// timeouts -> custom way to handle api timeout stragtey
	RegisterationTimeout   time.Duration
	UnregisterationTimeout time.Duration
	ParkdogTimeout         time.Duration
	ProcessDoneTimeout     time.Duration
	BarkTimeout            time.Duration
	ResetCallsTimeout      time.Duration
	ResetAllCallsTimeout   time.Duration
	ShutdownTimeout        time.Duration

	TickInterval    time.Duration
	ReportRetention time.Duration
	MaxHistorySize  int

	ShowReport             bool // showcases the report
	ConvDurationInHours    bool
	ConvDurationInMins     bool
	ConvDurationInSecs     bool
	ConvDurationInMiliSecs bool
}

// DefaultSettings returns default settings
func DefaultSettings() *Settings {
	return &Settings{
		EnableAvg:              true,
		ConvDurationInMiliSecs: true,
		ShowReport:             true,
		TickInterval:           100 * time.Millisecond,
		MaxHistorySize:         100,
		ReportRetention:        24 * time.Hour,
		RegisterationTimeout:   0,
		UnregisterationTimeout: 0,
		ParkdogTimeout:         0,
		ProcessDoneTimeout:     0,
		BarkTimeout:            0,
		ResetCallsTimeout:      0,
		ResetAllCallsTimeout:   0,
		ShutdownTimeout:        0,
	}
}

// ILog represents a log entry
type ILog struct {
	Policy  string
	Fail    *IPolicy
	Succeed *IPolicy
}

type IConv struct {
	Conv float64
	Unit string
}

func durationToHour(t time.Duration) IConv {
	return IConv{Conv: t.Hours(), Unit: "hr"}
}

func durationToMin(t time.Duration) IConv {
	return IConv{Conv: t.Minutes(), Unit: "min"}
}

func durationToSec(t time.Duration) IConv {
	return IConv{Conv: t.Seconds(), Unit: "sec"}
}

func durationToMS(t time.Duration) IConv {
	return IConv{Conv: float64(t / time.Millisecond), Unit: "ms"}
}

func (rd *Dog[T]) bin(req IDone) {
	p, exists := rd.policy[req.PolicyName]
	if !exists {
		return
	}

	newFns := make([]Funcs[T], 0, len(p.Fn))
	for _, r := range p.Fn {
		if r.Rank != req.Rank {
			newFns = append(newFns, r)
		}
	}
	p.Fn = newFns
}

// UniqueRank returns atomic id for the func rank
// func UniqueRank(i int) int {
// 	var rank int32 = 0
// 	return int(atomic.AddInt32(&rank, int32(i)))
// }
