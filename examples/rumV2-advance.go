package example

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	rumrpc "rum/app/misc/rum"
	"rum/app/rum/client"
	rum "rum/app/rum/server"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// PlayAdvancedRumExample demonstrates Rum V2 with multiple profiles and comprehensive monitoring
func PlayAdvancedRumExample() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	const (
		serverAddr = "localhost:9304"
	)

	profiles := []ProfileConfig{
		{
			Name:    "auth_profile",
			Rank:    1,
			Timeout: 15 * time.Second,
		},
		{
			Name:    "data_profile",
			Rank:    2,
			Timeout: 20 * time.Second,
		},
	}

	store := rum.NewRumStore[OrderRequest, *OrderResponse](ctx)
	rumServer := rum.New(ctx, store)
	rumClient := rum.NewClient(rumServer)

	for _, prof := range profiles {
		seq := rum.ISequence[OrderRequest]{
			Name: prof.Name,
			Rank: prof.Rank,
		}

		buildProfile(ctx, rumClient, seq, prof.Timeout)
	}

	if err := rumClient.BuildAll(); err != nil {
		log.Fatalf("Failed to build profiles: %v", err)
	}

	log.Println("📦 Rum V2 Advanced Example - Multiple Profiles")
	log.Printf("   Server: %s\n", serverAddr)
	log.Printf("   Profiles: %d\n", len(profiles))

	var wg sync.WaitGroup
	var metrics MultiProfileMetrics

	wg.Add(1)
	go func() {
		defer wg.Done()
		conf := rum.RumServer{
			Network: "tcp",
			Address: serverAddr,
		}
		rumClient.Run(ctx, conf)
	}()

	for _, prof := range profiles {
		wg.Add(1)
		go monitorProfile(
			ctx,
			rumClient,
			prof,
			&metrics,
		)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		sendMultipleRequests(ctx, serverAddr, profiles)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				metrics.Print()
			}
		}
	}()

	wg.Wait()

	log.Println("\n📊 Final Metrics Summary")
	metrics.Print()
}

// ProfileConfig holds configuration for a profile
type ProfileConfig struct {
	Name    string
	Rank    int
	Timeout time.Duration
}

// buildProfile configures and builds a single profile
func buildProfile(
	ctx context.Context,
	client *rum.Client[OrderRequest, *OrderResponse],
	seq rum.ISequence[OrderRequest],
	timeout time.Duration,
) {
	builder := client.CreateProfile(seq, timeout)

	switch seq.Name {
	case "auth_profile":
		builder.
			RegisterDispatch(
				ctx,
				"auth-service",
				"authenticate",
				rum.Settings{},
				handleAuthenticate,
			).
			RegisterDispatch(
				ctx,
				"auth-service",
				"authorize",
				rum.Settings{},
				handleAuthorize,
			)

	case "data_profile":
		builder.
			RegisterDispatch(
				ctx,
				"order-service",
				"create-order",
				rum.Settings{},
				handleCreateOrder,
			).
			RegisterDispatch(
				ctx,
				"order-service",
				"validate-order",
				rum.Settings{},
				handleValidateOrder,
			).
			RegisterDispatch(
				ctx,
				"payment-service",
				"process-payment",
				rum.Settings{},
				handleProcessPayment,
			)
	}

	builder.Build()
}

// monitorProfile monitors a single profile and collects metrics
func monitorProfile(
	ctx context.Context,
	client *rum.Client[OrderRequest, *OrderResponse],
	prof ProfileConfig,
	metrics *MultiProfileMetrics,
) {
	seq := rum.ISequence[OrderRequest]{
		Name: prof.Name,
		Rank: prof.Rank,
	}
	startTime := time.Now()
	select {

	case result := <-client.Server().Poll(seq):
		duration := time.Since(startTime)

		if result.IsReady {
			log.Printf("✓ Profile '%s' completed in %v", prof.Name, duration)
			metrics.AddSuccess(prof.Name, duration, result.Metric.JSON())
		} else {
			log.Printf("✗ Profile '%s' failed after %v", prof.Name, duration)
			metrics.AddFailure(prof.Name, duration)
		}

	case <-ctx.Done():
		duration := time.Since(startTime)
		log.Printf("⏱ Profile '%s' cancelled after %v", prof.Name, duration)
		metrics.AddTimeout(prof.Name, duration)
	}
}

// sendMultipleRequests sends requests to different profiles at intervals
func sendMultipleRequests(
	ctx context.Context,
	addr string,
	profiles []ProfileConfig,
) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	requestCount := 0

	for i := 0; i < 3; i++ {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, prof := range profiles {
				req := generateOrderRequest()

				if err := sendOrderRequest(addr, prof, req); err != nil {
					log.Printf("Failed to send request to %s: %v", prof.Name, err)
				} else {
					requestCount++
					log.Printf("→ Sent request #%d to '%s' for order %s",
						requestCount, prof.Name, req.OrderID)
				}
			}
		}
	}
}

// sendOrderRequest marshals and sends an order request to the server
func sendOrderRequest(addr string, prof ProfileConfig, req OrderRequest) error {
	parcel, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	post := rumrpc.IPost{
		Profile: &rumrpc.ISequence{
			Name:  prof.Name,
			Rank:  int32(prof.Rank),
			Input: parcel,
		},
		Push: true,
	}

	return client.POST(addr, []*rumrpc.IPost{&post})
}

func handleAuthenticate(ctx context.Context, req OrderRequest) (*OrderResponse, error) {
	time.Sleep(100 * time.Millisecond)
	return &OrderResponse{
		OrderID: req.OrderID,
		Stage:   "authenticated",
		Status:  "success",
		Message: fmt.Sprintf("User %s authenticated successfully", req.UserID),
	}, nil
}

func handleAuthorize(ctx context.Context, req OrderRequest) (*OrderResponse, error) {
	time.Sleep(100 * time.Millisecond)
	return &OrderResponse{
		OrderID: req.OrderID,
		Stage:   "authorized",
		Status:  "success",
		Message: fmt.Sprintf("User %s is authorized", req.UserID),
	}, nil
}

func handleCreateOrder(ctx context.Context, req OrderRequest) (*OrderResponse, error) {
	time.Sleep(150 * time.Millisecond)
	return &OrderResponse{
		OrderID: req.OrderID,
		Stage:   "order_created",
		Status:  "success",
		Message: fmt.Sprintf("Order created with %d items", len(req.Items)),
	}, nil
}

func handleValidateOrder(ctx context.Context, req OrderRequest) (*OrderResponse, error) {
	time.Sleep(100 * time.Millisecond)
	return &OrderResponse{
		OrderID: req.OrderID,
		Stage:   "order_validated",
		Status:  "success",
		Message: "Order validation passed",
	}, nil
}

func handleProcessPayment(ctx context.Context, req OrderRequest) (*OrderResponse, error) {
	time.Sleep(200 * time.Millisecond)
	return &OrderResponse{
		OrderID: req.OrderID,
		Stage:   "payment_processed",
		Status:  "success",
		Message: fmt.Sprintf("Payment of $%.2f processed", req.Amount),
	}, nil
}

// OrderRequest represents an order processing request
type OrderRequest struct {
	OrderID string   `json:"order_id"`
	UserID  string   `json:"user_id"`
	Items   []string `json:"items"`
	Amount  float64  `json:"amount"`
}

// OrderResponse represents a response from the order processing pipeline
type OrderResponse struct {
	OrderID string `json:"order_id"`
	Stage   string `json:"stage"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ProfileMetrics tracks metrics for a single profile
type ProfileMetrics struct {
	Name       string
	Successes  int64
	Failures   int64
	Timeouts   int64
	TotalTime  time.Duration
	LastMetric string
}

// MultiProfileMetrics tracks metrics across multiple profiles
type MultiProfileMetrics struct {
	mu       sync.RWMutex
	profiles map[string]*ProfileMetrics
}

// AddSuccess records a successful profile execution
func (m *MultiProfileMetrics) AddSuccess(name string, duration time.Duration, metric string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.profiles == nil {
		m.profiles = make(map[string]*ProfileMetrics)
	}

	if m.profiles[name] == nil {
		m.profiles[name] = &ProfileMetrics{Name: name}
	}

	pm := m.profiles[name]
	atomic.AddInt64(&pm.Successes, 1)
	pm.TotalTime += duration
	pm.LastMetric = metric
}

// AddFailure records a failed profile execution
func (m *MultiProfileMetrics) AddFailure(name string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.profiles == nil {
		m.profiles = make(map[string]*ProfileMetrics)
	}

	if m.profiles[name] == nil {
		m.profiles[name] = &ProfileMetrics{Name: name}
	}

	pm := m.profiles[name]
	atomic.AddInt64(&pm.Failures, 1)
	pm.TotalTime += duration
}

// AddTimeout records a timed-out profile
func (m *MultiProfileMetrics) AddTimeout(name string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.profiles == nil {
		m.profiles = make(map[string]*ProfileMetrics)
	}

	if m.profiles[name] == nil {
		m.profiles[name] = &ProfileMetrics{Name: name}
	}

	pm := m.profiles[name]
	atomic.AddInt64(&pm.Timeouts, 1)
	pm.TotalTime += duration
}

// Print outputs the current metrics summary
func (m *MultiProfileMetrics) Print() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.profiles) == 0 {
		return
	}

	fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║           Profile Execution Metrics Summary                ║")
	fmt.Println("╠════════════════════════════════════════════════════════════╣")

	for _, pm := range m.profiles {
		total := pm.Successes + pm.Failures + pm.Timeouts
		if total == 0 {
			continue
		}

		successRate := float64(pm.Successes) / float64(total) * 100
		avgTime := pm.TotalTime / time.Duration(total)

		fmt.Printf("║ Profile: %-45s ║\n", pm.Name)
		fmt.Printf("║   Successes: %d | Failures: %d | Timeouts: %d         ║\n",
			pm.Successes, pm.Failures, pm.Timeouts)
		fmt.Printf("║   Success Rate: %.1f%% | Avg Time: %v              ║\n",
			successRate, avgTime)
		fmt.Println("╠════════════════════════════════════════════════════════════╣")
	}

	fmt.Println("╚════════════════════════════════════════════════════════════╝")
}

// generateOrderRequest creates a mock order request
func generateOrderRequest() OrderRequest {
	return OrderRequest{
		OrderID: fmt.Sprintf("ORD-%d", time.Now().UnixNano()%10000),
		UserID:  "user_abc",
		Items:   []string{"item1", "item2", "item3"},
		Amount:  299.99,
	}
}
