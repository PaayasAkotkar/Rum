package rum

import "time"

// TimeFormat controls the lifecycle of a service after its dispatches complete.
// Priority: Remove > Deactivate
// ActivateAfter: if Deactivate is true and ActivateAfter is set,
// the service will be re-activated after the duration (acts as sleep/wake)
type TimeFormat struct {
	Call          time.Duration
	ActivateAfter *time.Duration
	Deactivate    bool
	Remove        bool
	Retry         *RetryPolicy // nil = no retry
}

func NewTimeFormat() *TimeFormat {
	return &TimeFormat{}
}

// set funcs

func (t *TimeFormat) SetCallTime(d time.Duration) {
	t.Call = d
}
func (t *TimeFormat) SetActivateTime(d time.Duration) {
	t.ActivateAfter = &d
}
func (t *TimeFormat) SetDeactivate(b bool) {
	t.Deactivate = b
}
func (t *TimeFormat) SetRemove(b bool) {
	t.Remove = b
}
func (t *TimeFormat) SetRetry(r *RetryPolicy) {
	t.Retry = r
}

// end

// get funcs

func (t *TimeFormat) GetCallTime() time.Duration {
	return t.Call
}
func (t *TimeFormat) GetActivateTime() *time.Duration {
	return t.ActivateAfter
}
func (t *TimeFormat) ShouldDeactivate() bool {
	return t.Deactivate
}
func (t *TimeFormat) ShouldRemove() bool {
	return t.Remove
}

// end

type RetryPolicy struct {
	Max      int           // how many times to retry
	Interval time.Duration // wait between retries
}

func NewRetryPolicy(max int, interval time.Duration) *RetryPolicy {
	return &RetryPolicy{
		Max:      max,
		Interval: interval,
	}
}

// set funcs

func (r *RetryPolicy) SetMaxRetry(max int) {
	r.Max = max
}
func (r *RetryPolicy) SetRetryInterval(interval time.Duration) {
	r.Interval = interval
}

// end

// get funcs

func (r *RetryPolicy) GetMaxRetry() int {
	return r.Max
}
func (r *RetryPolicy) GetRetryInterval() time.Duration {
	return r.Interval
}
