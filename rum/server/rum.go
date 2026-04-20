package rum

import (
	"context"
	"log"
	rumchakra "rum/app/chakra"
	rumrpc "rum/app/misc/rum"
	rumpaint "rum/app/paint"
	"strings"
	"sync"
)

// Rum implements the core design of server

type Rum[In, Out any] struct {
	rumrpc.UnimplementedOnRumServiceServer

	// light *Light[In, IDispatchResult]
	chakra *rumchakra.Chakra[IDispatchResult]
	store  *RumStore[In, Out]

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
		chakra:            rumchakra.New[IDispatchResult](),
		post:              make(chan ILinks[In, Out]),
		deleteService:     make(chan ILinks[In, Out]),
		activateService:   make(chan ILinks[In, Out]),
		deactivateService: make(chan ILinks[In, Out]),
		deleteProfile:     make(chan ILinks[In, Out]),
		activateProfile:   make(chan ILinks[In, Out]),
		deactivateProfile: make(chan ILinks[In, Out]),
		ctx:               ctx,
	}
}

// Paper monitors the profile and returns the result of the serivce of created event
func (r *Rum[In, Out]) Paper(profile ISequence[In]) *IDispatchResult {
	log.Println("in tick fetch")
	return r.tickFetch(profile)
}

func (r *Rum[In, Out]) tickFetch(profile ISequence[In]) *IDispatchResult {
	// Use only Name+Rank for pub/sub and profile lookup — Input is a pointer
	// whose address won't match between subscriber and publisher.

	// ch := r.light.Subscribe(key)
	// defer r.light.Unsub(key, ch)
	ch := r.chakra.Subscribe(profile.Name)
	defer r.chakra.Kai(profile.Name, ch)
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
