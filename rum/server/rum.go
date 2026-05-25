package rum

import (
	"context"
	"log"
	cheetah "rum/app/cheetah"
	injection "rum/app/di"
	rumrpc "rum/app/misc/rum"
	rumpaint "rum/app/paint"
	"strings"
	"sync"
)

// Rum implements the core design of server

type Rum[In, Out any] struct {
	rumrpc.UnimplementedOnRumServiceServer

	DI *injection.Client

	// light *Light[In, IDispatchResult]
	cheetah *cheetah.Cheetah[IDispatchResult]
	store   *RumStore[In, Out]

	post              chan ILinks[In, Out]
	deleteService     chan ILinks[In, Out]
	activateService   chan ILinks[In, Out]
	deactivateService chan ILinks[In, Out]
	deleteProfile     chan ILinks[In, Out]
	activateProfile   chan ILinks[In, Out]
	deactivateProfile chan ILinks[In, Out]

	ctx context.Context
	mu  sync.Mutex
	wg  sync.WaitGroup
	sb  strings.Builder
}

func New[In, Out any](ctx context.Context, store *RumStore[In, Out]) *Rum[In, Out] {
	t := rumpaint.Header(`
██████╗░██╗░░░██╗███╗░░░███╗
██╔══██╗██║░░░██║████╗░████║
██████╔╝██║░░░██║██╔████╔██║
██╔══██╗██║░░░██║██║╚██╔╝██║
██║░░██║╚██████╔╝██║░╚═╝░██║
╚═╝░░╚═╝░╚═════╝░╚═╝░░░░░╚═╝
	`)
	log.Println(t)
	return &Rum[In, Out]{
		store: store,
		// light:             NewLight[In, IDispatchResult](),
		cheetah:           cheetah.New[IDispatchResult](),
		post:              make(chan ILinks[In, Out]),
		deleteService:     make(chan ILinks[In, Out]),
		activateService:   make(chan ILinks[In, Out]),
		deactivateService: make(chan ILinks[In, Out]),
		deleteProfile:     make(chan ILinks[In, Out]),
		activateProfile:   make(chan ILinks[In, Out]),
		deactivateProfile: make(chan ILinks[In, Out]),
		ctx:               ctx,
		DI:                injection.NewClient(ctx, "rum-server"),
	}
}

func (r *Rum[In, Out]) fetch(profile ISequence[In]) *IDispatchResult {

	// ch := r.light.Subscribe(key)
	// defer r.light.Unsub(key, ch)
	ch := r.cheetah.Subscribe(profile.Name)
	defer r.cheetah.Unsubscribe(profile.Name, ch)
	for {
		select {
		case <-r.ctx.Done():
			return nil

		case result := <-ch:
			if result != nil && result.IsReady {
				return result
			}
		}
	}
}

func (r *Rum[In, Out]) tickFetchPoll(profile ISequence[In]) <-chan *IDispatchResult {
	// Use only Name+Rank for pub/sub and profile lookup
	ch := r.cheetah.Subscribe(profile.Name)

	// Unsubscribe when the server context is cancelled
	go func() {
		<-r.ctx.Done()
		r.cheetah.Unsubscribe(profile.Name, ch)
	}()

	return ch
}
