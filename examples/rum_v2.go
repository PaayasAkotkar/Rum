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

func PlayRumV2() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	addr := "localhost:9302"
	profName := "v2_profile"

	store := rum.NewRumStore[Req, *Resp](ctx)

	rumServer := rum.New(ctx, store)
	rumClient := rum.NewClient(rumServer)

	rumx := rum.New(ctx, store)

	seq := rum.ISequence[Req]{Name: profName, Rank: 1}

	rumClient.CreateProfile(seq, 10*time.Second).
		RegisterDispatch(ctx, "v2-service", "v2-reply", rum.Settings{},
			func(ctx context.Context, req Req) (*Resp, error) {
				query := "default query"
				if req.Query != nil {
					query = *req.Query
				}

				name := "Anonymous"
				if req.Name != nil {
					name = *req.Name
				}

				var resp = Resp{
					Info: fmt.Sprintf("Hello %s (V2 Builder API)! - You asked: %s", name, query),
				}
				return &resp, nil
			},
		)

	rumClient.BuildAll()
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
		query := "How does the V2 Builder work?"
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

		client.POST(addr, []*rumrpc.IPost{&post})
	}()

	wg.Wait()
}
