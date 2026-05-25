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
	"syscall"
	"time"
)

// PlayBasicRumExample demonstrates a simple Rum V2 setup with a single profile
func PlayBasicRumExample() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	const (
		serverAddr  = "localhost:9303"
		profileName = "basic_profile"
	)

	store := rum.NewRumStore[SimpleRequest, *SimpleResponse](ctx)
	rumServer := rum.New(ctx, store)
	rumClient := rum.NewClient(rumServer)

	seq := rum.ISequence[SimpleRequest]{
		Name: profileName,
		Rank: 1,
	}

	rumClient.CreateProfile(seq, 10*time.Second).
		RegisterDispatch(
			ctx,
			"user-service",
			"get-user",
			rum.Settings{},
			handleGetUser,
		).
		Build()

	if err := rumClient.BuildAll(); err != nil {
		log.Fatalf("Failed to build profiles: %v", err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		conf := rum.RumServer{
			Network: "tcp",
			Address: serverAddr,
		}
		rumClient.Run(ctx, conf)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		result := <-rumClient.Server().Poll(seq)

		if result.IsReady {
			log.Printf("✓ Profile '%s' completed successfully", profileName)
			log.Printf("  Metric: %s", result.Metric.JSON())
			log.Printf("  Result: %s", string(result.Result))
		} else {
			log.Printf("✗ Profile '%s' failed or timed out", profileName)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(3 * time.Second)

		req := SimpleRequest{
			UserID: "baba_96",
			Email:  "baba@example.com",
		}

		if err := sendRequest(serverAddr, profileName, seq.Rank, req); err != nil {
			log.Printf("Failed to send request: %v", err)
		}
	}()

	wg.Wait()
}

// handleGetUser is the event handler for user-service.get-user
func handleGetUser(ctx context.Context, req SimpleRequest) (*SimpleResponse, error) {
	resp := &SimpleResponse{
		UserID: req.UserID,
		Status: "success",
		Message: fmt.Sprintf(
			"User %s (%s) retrieved successfully",
			req.UserID,
			req.Email,
		),
		Timestamp: time.Now().Unix(),
	}
	return resp, nil
}

// sendRequest marshals and sends a request to the Rum server
func sendRequest(addr, profileName string, rank int, req SimpleRequest) error {
	parcel, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	post := rumrpc.IPost{
		Profile: &rumrpc.ISequence{
			Name:  profileName,
			Rank:  int32(rank),
			Input: parcel,
		},
		Push: true,
	}

	return client.POST(addr, []*rumrpc.IPost{&post})
}

// SimpleRequest represents a basic user request
type SimpleRequest struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// SimpleResponse represents a basic response from the service
type SimpleResponse struct {
	UserID    string `json:"user_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}
