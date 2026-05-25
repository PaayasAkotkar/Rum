package example

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reflect"
	injection "rum/app/di"
	rumrpc "rum/app/misc/rum"
	"rum/app/rum/client"
	rum "rum/app/rum/server"
	"sync"
	"syscall"
	"time"
)

// GreeterService is a dummy service that greets users
type GreeterService struct {
	Prefix string
}

func (s *GreeterService) Greet(name string) string {
	return fmt.Sprintf("%s, %s!", s.Prefix, name)
}

func PlayRumDI() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	addr := "localhost:9301"
	profName := "di_profile"

	store := rum.NewRumStore[Req, *Resp](ctx)
	rumServer := rum.New(ctx, store)
	rumClient := rum.NewClient(rumServer)

	service := rum.NewService[Req, *Resp](ctx, rum.Settings{}, "di-ser")
	dispatch := rum.NewDispatcher[Req, *Resp](rum.Settings{})

	rumx := rum.New(ctx, store)

	// --- DI Integration
	greeterType := reflect.TypeOf((*GreeterService)(nil))
	rumx.DI.AddSingleton(greeterType, injection.Factory{
		Fn: func(ctx context.Context, c *injection.Container) (any, error) {
			log.Println("[DI] Initializing GreeterService...")
			return &GreeterService{Prefix: "Welcome to Rum Server with DI"}, nil
		},
	})

	// Build the DI container
	statusCh := rumx.DI.BuildStatus()
	go func() {
		if err := rumx.DI.Build(ctx); err != nil {
			log.Fatalf("[DI] Build failed: %v", err)
		}
	}()

	status := <-statusCh
	log.Printf("[DI] Build status: %s", *status)
	rumx.DI.CloseBuildStatus(statusCh)

	var register rum.IRegister[Req, *Resp]
	register.Fn = func(ctx context.Context, req Req) (*Resp, error) {
		query := "default query"
		if req.Query != nil {
			query = *req.Query
		}

		name := "Anonymous"
		if req.Name != nil {
			name = *req.Name
		}

		// Resolve the dependency from rumx.DI inside the handler
		svc, err := rumx.DI.GetService(greeterType)
		var greeting string
		if err != nil {
			log.Printf("Failed to resolve GreeterService: %v", err)
			greeting = "Hello (without DI)"
		} else {
			greeting = svc.(*GreeterService).Greet(name)
		}

		var resp = Resp{
			Info: fmt.Sprintf("%s - You asked: %s", greeting, query),
		}
		return &resp, nil
	}

	dispatch.Register("di-reply", register)
	service.SetDispatch(dispatch)

	seq := rum.ISequence[Req]{Name: profName, Rank: 1}

	rumClient.CreateProfile(seq, 10*time.Second).
		PushService("di-service", service).
		Build()

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		rumx.Serve(ctx, rum.RumServer{
			Network: "tcp",
			Address: addr,
		})
	}()

	go func() {
		defer wg.Done()
		res := <-rumx.Poll(seq)

		if res.IsReady {
			log.Println("metric: ", res.Metric.JSON())
			log.Println("result: ", string(res.Result))
		}
	}()

	go func() {
		defer wg.Done()
		time.Sleep(time.Second * 3)
		var req Req
		id := "u123"
		name := "Paayas"
		query := "How does DI work in Rum?"
		req.ID = &id
		req.Name = &name
		req.Query = &query
		req.Profile = seq.Name
		parcel, err := json.Marshal(req)
		if err != nil {
			log.Println("marshal error", err)
			return
		}
		post := rumrpc.IPost{
			Profile: &rumrpc.ISequence{
				Name:  seq.Name,
				Rank:  int32(seq.Rank),
				Input: parcel,
			},
			Push: true,
		}
		if err := client.POST(addr, []*rumrpc.IPost{&post}); err != nil {
			log.Println("error: ", err)
		}
	}()

	wg.Wait()
}
