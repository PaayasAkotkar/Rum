// Package dog ...
// The things with the apis is that
// it is created based on respecting the time
// each api has to executed under some circumstances
// new: returns the error if any
package dog

import (
	"fmt"
	"log"
	"time"
)

// Register pushes the profile in heap
func (rd *Dog[T]) Register(policy *Policy[T]) error {
	t := rd.Settings.RegisterationTimeout
	if t == 0 {
		t = 2 * time.Second
	}
	select {
	case rd.register <- policy:
		return nil
	case <-time.After(t):
		return fmt.Errorf("timeout registering policy %s", policy.Name)
	case <-rd.ctx.Done():
		return fmt.Errorf("rumdog context cancelled")
	}
}

// Unregister completey relaeses the profile
func (rd *Dog[T]) Unregister(name string) error {
	t := rd.Settings.UnregisterationTimeout
	if t == 0 {
		t = 1 * time.Second
	}
	select {
	case rd.unregister <- name:
		return nil
	case <-time.After(t):
		return fmt.Errorf("timeout unregistering policy %s", name)
	case <-rd.ctx.Done():
		return fmt.Errorf("rumdog context cancelled")
	}
}

// ParkDog signals to monitor the policyName for all the funcs
func (rd *Dog[T]) ParkDog(policyName string) error {
	t := rd.Settings.ParkdogTimeout
	if t == 0 {
		t = 1 * time.Second
	}
	select {
	case rd.parkDog <- policyName:
		return nil
	case <-time.After(t):
		return fmt.Errorf("timeout parking dog on policy %s", policyName)
	case <-rd.ctx.Done():
		return fmt.Errorf("rumdog context cancelled")
	}
}

// Done successfully performed the func
func (rd *Dog[T]) Done(done IDone) error {
	t := rd.Settings.ProcessDoneTimeout
	if t == 0 {
		t = 1 * time.Second
	}
	select {
	case rd.done <- done:
		return nil
	case <-time.After(t):
		return fmt.Errorf("timeout signaling done for policy %s", done.PolicyName)
	case <-rd.ctx.Done():
		return fmt.Errorf("rumdog context cancelled")
	}
}

// Bark signals an error during func
func (rd *Dog[T]) Bark(bark IBark) error {
	t := rd.Settings.BarkTimeout
	if t == 0 {
		t = 1 * time.Second
	}
	bark.Time = time.Now()
	select {
	case rd.bark <- bark:
		return nil
	case <-time.After(t):
		return fmt.Errorf("timeout barking for policy %s", bark.Policy)
	case <-rd.ctx.Done():
		return fmt.Errorf("rumdog context cancelled")
	}
}

// Reset reset the policy call
func (rd *Dog[T]) Reset(policyName string) error {
	t := rd.Settings.ResetCallsTimeout
	if t == 0 {
		t = 1 * time.Second
	}
	select {
	case rd.reset <- policyName:
		return nil
	case <-time.After(t):
		return fmt.Errorf("timeout resetting policy %s", policyName)
	case <-rd.ctx.Done():
		return fmt.Errorf("rumdog context cancelled")
	}
}

// ResetAll resets all the pollicies calls to zero
func (rd *Dog[T]) ResetAll() error {
	t := rd.Settings.ResetAllCallsTimeout
	if t == 0 {
		t = 1 * time.Second
	}
	select {
	case rd.resetAll <- true:
		return nil
	case <-time.After(t):
		return fmt.Errorf("timeout resetting all policies")
	case <-rd.ctx.Done():
		return fmt.Errorf("rumdog context cancelled")
	}
}

// Shutdown graceful shutdown
func (rd *Dog[T]) Shutdown() error {
	if rd.cancel != nil {
		rd.cancel()
	}

	rd.once.Do(func() {
		close(rd.stopCh)
	})

	t := rd.Settings.ShutdownTimeout
	if t == 0 {
		t = 5 * time.Second
	}

	go func() {
		rd.wg.Wait()
		close(rd.done)
	}()

	select {
	case <-rd.done:
		log.Println("Shutdown complete")
		for r := range rd.monitors {
			delete(rd.monitors, r)
		}
		return nil
	case <-time.After(t):
		return fmt.Errorf("graceful shutdown timed out after %v", t)
	}
}

// SetSettings custom setting
func (rd *Dog[T]) SetSettings(settings *Settings) {
	rd.mu.Lock()
	defer rd.mu.Unlock()
	rd.Settings = settings
}
