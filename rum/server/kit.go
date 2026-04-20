package rum

// Flow:
// Bucket -> S3 bucket
// Model -> current Profile Model
//               Save Services
//       Read & Write to that services as per the descriptions.

import (
	rumstack "rum/app/stack"
	"fmt"
	"sync"
	"time"
)

const maxMetrics = 100

// Kit is the reusable container for a profile's model config, embedder,
// active/inactive services, and accumulated metrics.
type Kit[In, Out any] struct {
	mu sync.RWMutex
	// embd     *Embd
	Model    string
	Bucket   string
	isHybrid bool
	// genkit          *IGenKit
	activeService   map[string]*Service[In, Out]
	inactiveService map[string]*Service[In, Out]
	serviceStack    rumstack.Stack[string]
	Format          *TimeFormat
	Metrics         IMetric `json:"metrics"`
}

func NewKit[In, Out any](model string) *Kit[In, Out] {
	return &Kit[In, Out]{
		Model:           model,
		Metrics:         NewIMetric(),
		activeService:   make(map[string]*Service[In, Out]),
		inactiveService: make(map[string]*Service[In, Out]),
	}
}

// get funcs

func (k *Kit[In, Out]) GetBucket() string      { return k.Bucket }
func (k *Kit[In, Out]) IsHybrid() bool         { return k.isHybrid }
func (k *Kit[In, Out]) GetModel() string       { return k.Model }
func (k *Kit[In, Out]) GetFormat() *TimeFormat { return k.Format }
func (k *Kit[In, Out]) GetMetrics() IMetric    { return k.Metrics }

func (k *Kit[In, Out]) GetServices() map[string]*Service[In, Out] {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.activeService
}

func (k *Kit[In, Out]) GetService(key string) (*Service[In, Out], error) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	s, ok := k.activeService[key]
	if !ok {
		return nil, fmt.Errorf("service %q not found or inactive", key)
	}
	return s, nil
}

// end

// set funcs

func (k *Kit[In, Out]) SetBucket(bucket string) { k.Bucket = bucket }
func (k *Kit[In, Out]) SetMode(isHybrid bool)   { k.isHybrid = isHybrid }
func (k *Kit[In, Out]) SetFormat(f *TimeFormat) { k.Format = f }
func (k *Kit[In, Out]) SetModel(model string)   { k.Model = model }
func (k *Kit[In, Out]) SetMetrics(m IMetric)    { k.Metrics = m }

func (k *Kit[In, Out]) SetService(services map[string]*Service[In, Out]) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.activeService = services
	for key := range services {
		k.serviceStack.Push(key)
	}
}

func (k *Kit[In, Out]) PushService(name string, service *Service[In, Out]) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if _, ok := k.activeService[name]; !ok {
		k.serviceStack.Push(name)
	}
	k.activeService[name] = service
}

// GetServiceCollections returns active services in stack order
func (k *Kit[In, Out]) GetServiceCollections() []*Service[In, Out] {
	k.mu.RLock()
	defer k.mu.RUnlock()
	keys := k.serviceStack.Range(k.serviceStack.Len())
	out := make([]*Service[In, Out], 0, len(keys))
	for _, key := range keys {
		if svc, ok := k.activeService[key]; ok {
			out = append(out, svc)
		}
	}
	return out
}

func (k *Kit[In, Out]) GetActiveServiceKeys() []string {
	return k.serviceStack.Range(k.serviceStack.Len())
}

func (k *Kit[In, Out]) DeactivateService(key string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	svc, ok := k.activeService[key]
	if !ok {
		return fmt.Errorf("service %q not active", key)
	}
	k.inactiveService[key] = svc
	delete(k.activeService, key)
	k.serviceStack.Erase(key)
	return nil
}

func (k *Kit[In, Out]) ActivateService(key string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	svc, ok := k.inactiveService[key]
	if !ok {
		return fmt.Errorf("service %q not in inactive pool", key)
	}
	k.activeService[key] = svc
	delete(k.inactiveService, key)
	k.serviceStack.Push(key)
	return nil
}

func (k *Kit[In, Out]) RemoveService(key string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	delete(k.activeService, key)
	delete(k.inactiveService, key)
	k.serviceStack.Erase(key)
}

func (k *Kit[In, Out]) AddSucceedReport(report IMetricAgentSucceed) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.Metrics.AddSucceedReport(report)
}

func (k *Kit[In, Out]) AddFailReport(report IMetricAgentFail) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.Metrics.AddFailReport(report)
}

func (k *Kit[In, Out]) AddRemoveReport(t time.Time) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.Metrics.AddRemoveReport(t)
}

func (k *Kit[In, Out]) AddDeactiveReport(t time.Time) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.Metrics.AddDeactiveReport(t)
}

func (k *Kit[In, Out]) AddActivateReport(t time.Time) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.Metrics.AddActivateReport(t)
}

func (k *Kit[In, Out]) AddBudgetReport(b IMetricBudget) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.Metrics.AddBudget(b)
}

func (k *Kit[In, Out]) AddProfileReport(p IMetricProfile) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.Metrics.AddProfile(p)
}

func (k *Kit[In, Out]) AddRequestReport() {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.Metrics.AddRequest()
}

// end
