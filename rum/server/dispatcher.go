// Package rum ....
// Flow ->
//
//				           {name : ["event1", "event2", "event3"]}
//			                                        V
//		                    Call the events as per the order & desc of the events.
//		                         Perform metric writing for each input.
//	                                Store completion event name & result.
package rum

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	rumdog "rum/app/dog"
	rumstack "rum/app/stack"
	"sync"
	"time"

	"github.com/avast/retry-go/v5"
)

// Dispatcher controls registered agent functions and their results
type Dispatcher[in, out any] struct {
	registry   map[string]IRegister[in, out]
	rinput     map[string]in
	Settings   Settings
	events     rumstack.Stack[string]
	result     map[string]*IDispatchResult
	isComplete map[string]bool
	metric     map[string]map[int]IAgentResp // name -> count -> resp
	wg         sync.WaitGroup
}

func NewDispatcher[in, out any](settings Settings) *Dispatcher[in, out] {
	return &Dispatcher[in, out]{
		registry:   make(map[string]IRegister[in, out]),
		rinput:     make(map[string]in),
		Settings:   settings,
		result:     make(map[string]*IDispatchResult),
		isComplete: make(map[string]bool),
		metric:     make(map[string]map[int]IAgentResp),
	}
}

type Settings struct {
	Base      time.Duration
	SleepTime time.Duration
	Dog       rumdog.Settings
}

// IAgentResp holds per-call metric data
type IAgentResp struct {
	Succeed *IMetricAgentSucceed `json:"succeed"`
	Fail    *IMetricAgentFail    `json:"fail"`
}

// get funcs

func (d *Dispatcher[in, out]) GetRegistry() map[string]IRegister[in, out] {
	return d.registry
}

func (d *Dispatcher[in, out]) GetEvents(limit int) []string {
	return d.events.Range(limit)
}

func (d *Dispatcher[in, out]) GetLatestRegistry() *string {
	return d.events.Latest()
}

func (d *Dispatcher[in, out]) GetResults(name string) *IDispatchResult {
	if _, ok := d.result[name]; !ok {
		for n := range d.registry {
			log.Println("names: ", n)
		}
		log.Println("dispatcher: not found name ", name)
		return nil
	}
	return d.result[name]
}

// GetMetric returns the latest metric entry for a named dispatch
func (d *Dispatcher[in, out]) GetMetric(name string) IAgentResp {
	return d.metric[name][d.metricCount(name)]
}

func (d *Dispatcher[in, out]) GetMetrics(name string) map[int]IAgentResp {
	return d.metric[name]
}

func (d *Dispatcher[in, out]) metricCount(name string) int {
	return len(d.metric[name])
}

// end

// set funcs

// Release deletes all the metrices
func (d *Dispatcher[in, out]) Release() {
	for r := range d.metric {
		delete(d.metric, r)
	}
}

func (d *Dispatcher[in, out]) Register(event string, fn IRegister[in, out]) {
	if _, ok := d.registry[event]; !ok {
		d.registry[event] = fn
		d.events.Push(event)
	}
}

func (d *Dispatcher[in, out]) Unregister(name string) {
	delete(d.registry, name)
	delete(d.rinput, name)
	delete(d.isComplete, name)
	delete(d.metric, name)
	delete(d.result, name)
	d.events.Erase(name)
}

// call invokes the named registered function and records its metric
func (d *Dispatcher[in, out]) call(ctx context.Context, name string, input in, policy *RetryPolicy) error {
	fn, ok := d.registry[name]
	if !ok {
		err := fmt.Errorf("service %s not found", name)
		d.writeMetric(name, IAgentResp{Fail: &IMetricAgentFail{At: time.Now(), Reason: err.Error()}})
		return err
	}

	max := 1
	interval := time.Duration(0)

	if policy != nil {
		max = policy.Max + 1 // +1 so Max=3 means 1 original + 3 retries
		interval = policy.Interval
	}

	log.Println("max: ", max)
	t := d.Settings.Base
	if t == 0 {
		t = 10 * time.Second
	}
	ts := d.Settings.SleepTime
	if ts == 0 {
		ts = 100 * time.Millisecond
	}

	// dog := rumdog.New[out](t)
	// dog.Watch()
	// defer dog.Shutdown()

	p := rumdog.NewPolicy[out](t)
	p.Name = name
	a := func() (*out, error) {
		time.Sleep(ts)
		resp, err := fn.Fn(ctx, input)
		return &resp, err
	}
	p.AddFunc(rumdog.Funcs[out]{
		Name: name,
		Fn:   &a,
	})

	attempt := 0
	ret := retry.New(retry.Attempts(uint(max)), retry.Delay(interval))

	err := ret.Do(func() error {
		attempt++

		if ctx.Err() != nil {
			return retry.Unrecoverable(ctx.Err())
		}

		cli := rumdog.NewClient[out](t)
		defer cli.Close()

		cli.DefinePolicy(name, ts).AddFuncWithReturn(name, a).Build()

		rep, err := cli.ExecuteAndReport(p.Name)

		if err == nil {
			inData, err := json.Marshal(input)
			if err != nil {
				return retry.Unrecoverable(err)
			}
			d.writeMetric(name, IAgentResp{
				Succeed: &IMetricAgentSucceed{
					TimeTaken:     rep.TotalDuration,
					AgentReply:    string(rep.Output),
					ClientRequest: string(inData),
				},
			})
			r := NewDispatchResult()
			r.IsReady = true
			r.Result = rep.Output
			d.handleOutput(name, r)
			d.handleComplete(name, true)
			d.handleInput(name, input)
			cli.Unregister(p.Name)
			return nil
		} else {
			cli.Unregister(p.Name)

			d.writeMetric(name, IAgentResp{
				Fail: &IMetricAgentFail{
					Reason: fmt.Sprintf("attempt %d: %s", attempt, err.Error()),
					At:     time.Now(),
				},
			})
			return err
		}
	})

	return err
}

func (d *Dispatcher[in, out]) writeMetric(name string, resp IAgentResp) {
	if _, ok := d.metric[name]; !ok {
		d.metric[name] = make(map[int]IAgentResp)
	}
	d.metric[name][d.metricCount(name)+1] = resp
}

func (d *Dispatcher[in, out]) handleOutput(name string, res *IDispatchResult) {
	d.result[name] = res
}
func (d *Dispatcher[in, out]) handleInput(name string, input in) {
	d.rinput[name] = input
}
func (d *Dispatcher[in, out]) handleComplete(name string, complete bool) {
	d.isComplete[name] = complete
}

// end
