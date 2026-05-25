package dog

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Client represents the simplified watchdog client API
// Client only needs to:
// 1. Create policies with functions
// 2. Register policies
// 3. Call Summon() - everything else is automatic
// 4. Retrieve reports
type Client[T any] struct {
	dog *Dog[T]
}

// NewClient creates a new  client wrapper
func NewClient[T any](baseTimeout time.Duration) *Client[T] {
	dog := New[T](baseTimeout)
	dog.Watch()

	return &Client[T]{
		dog: dog,
	}
}

func (c *Client[T]) Dog() *Dog[T] {
	return c.dog
}

// DefinePolicy creates a policy with a friendly API
func (c *Client[T]) DefinePolicy(name string, timeout time.Duration) *PolicyBuilder[T] {
	policy := NewPolicy[T](timeout)
	policy.SetName(name)

	return &PolicyBuilder[T]{
		client: c,
		policy: policy,
	}
}

// ExecuteAndReport executes all functions in a policy and returns the report
// This is the main API - client calls this and gets everything
func (c *Client[T]) ExecuteAndReport(policyName string) (*WatchdogReport, error) {
	if err := c.dog.Summon(policyName); err != nil {
		return nil, fmt.Errorf("failed to execute policy %s: %w", policyName, err)
	}

	report := c.dog.Pakkun(policyName)
	if report == nil {
		return nil, fmt.Errorf("no report available for policy %s", policyName)
	}

	// Check if the execution actually succeeded
	if report.LastError != nil {
		return report, report.LastError
	}

	return report, nil
}

// LazyExecuteAndReport is kind of simple to execute and logs the report
func (c *Client[T]) LazyExecuteAndReport(policyName string) error {
	report, err := c.ExecuteAndReport(policyName)
	if err != nil {
		return err
	}
	formatted := &FormattedReport{Report: report}
	formatted.Display()
	return nil
}

// ExecuteMultiple executes multiple policies concurrently and returns all reports
func (c *Client[T]) ExecuteMultiple(policyNames ...string) (map[string]*WatchdogReport, error) {
	results := make(map[string]*WatchdogReport)
	resultsMu := sync.Mutex{}

	var wg sync.WaitGroup
	errChan := make(chan error, len(policyNames))

	for _, name := range policyNames {
		wg.Add(1)
		go func(pName string) {
			defer wg.Done()

			report, err := c.ExecuteAndReport(pName)
			if err != nil {
				errChan <- err
				return
			}
			resultsMu.Lock()
			results[pName] = report
			resultsMu.Unlock()
		}(name)
	}
	wg.Wait()

	close(errChan)

	var errs []error
	for err := range errChan {
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return results, fmt.Errorf("errors during execution: %v", errs)
	}

	return results, nil
}

// GetMetrics retrieves metrics for a policy
func (c *Client[T]) GetMetrics(policyName string) *SystemMetrics {
	return c.dog.GetMetrics(policyName)
}

// GetProgress retrieves progress for a policy
func (c *Client[T]) GetProgress(policyName string) *ExeProgress {
	return c.dog.GetProgress(policyName)
}

// ListPolicies returns all registered policy names
func (c *Client[T]) ListPolicies() []string {
	return c.dog.GetAllPolicies()
}

// Reset resets a policy's counters
func (c *Client[T]) Reset(policyName string) error {
	return c.dog.Reset(policyName)
}

// ResetAll resets all policies
func (c *Client[T]) ResetAll() error {
	return c.dog.ResetAll()
}

// Close gracefully shuts down the client
func (c *Client[T]) Close() error {
	return c.dog.Shutdown()
}

// Unregister erases the policy
func (c *Client[T]) Unregister(policyName string) error {
	return c.dog.Unregister(policyName)
}

// PolicyBuilder provides a fluent API for building policies
type PolicyBuilder[T any] struct {
	client *Client[T]
	policy *Policy[T]
}

// AddFunc adds a function to the policy
// rank can be auto-generated, user just passes the function
func (pb *PolicyBuilder[T]) AddFunc(name string, fn func() error) *PolicyBuilder[T] {
	rank := len(pb.policy.Fn) + 1
	pb.policy.AddFunc(Funcs[T]{
		Name: name,
		Rank: rank,
		Void: &fn,
	})
	return pb
}

// AddFuncWithReturn adds a function that returns data
func (pb *PolicyBuilder[T]) AddFuncWithReturn(name string, fn func() (*T, error)) *PolicyBuilder[T] {
	rank := len(pb.policy.Fn) + 1
	pb.policy.AddFunc(Funcs[T]{
		Name: name,
		Rank: rank,
		Fn:   &fn,
	})
	return pb
}

// Build registers the policy and returns the client
func (pb *PolicyBuilder[T]) Build() (*Client[T], error) {
	if err := pb.client.dog.Register(pb.policy); err != nil {
		return nil, fmt.Errorf("failed to register policy %s: %w", pb.policy.Name, err)
	}
	if err := pb.client.dog.ParkDog(pb.policy.Name); err != nil {
		return nil, fmt.Errorf("failed to start monitoring policy %s: %w", pb.policy.Name, err)
	}

	log.Printf("✅ Policy ready: %s (timeout: %v, functions: %d)",
		pb.policy.Name, pb.policy.GetBase(), len(pb.policy.GetFunc()))

	return pb.client, nil
}
