// Package rum...
// Grpc Serviers that just act as a gateawy nothing else basically triggers the channels which than triggers the hub
package rum

import (
	"context"
	"encoding/json"
	"log"
	rumrpc "rum/app/misc/rum"
	"runtime/debug"
	"time"
)

// POST publishes the paper
func (r *Rum[In, Out]) POST(ctx context.Context, in *rumrpc.IPostRequest) (*rumrpc.IPostResponse, error) {
	log.Println("in push")
	var s = make([]ILink[In, Out], 0, len(in.Post))

	for _, x := range in.Post {
		var input In
		if err := json.Unmarshal(x.Profile.Input, &input); err != nil {
			log.Println("unmarshal error:", err)
			continue
		}
		log.Println("unmarshalling: ", string(x.Profile.Input))
		log.Println("input:", input)

		s = append(s, ILink[In, Out]{
			Seq: ISequence[In]{
				Name:  x.Profile.Name,
				Rank:  int(x.Profile.Rank),
				Input: &input,
			},
		})
	}

	r.post <- ILinks[In, Out]{Links: s, Clean: true}

	return &rumrpc.IPostResponse{Succeed: true}, nil
}

// DELETE permanently deletes a profile
func (r *Rum[In, Out]) DELETE(ctx context.Context, in *rumrpc.IDeleteRequest) (*rumrpc.IDeleteResponse, error) {
	log.Println("in remove")
	var s = make([]ILink[In, Out], 0, len(in.Delete))

	for _, x := range in.Delete {

		s = append(s, ILink[In, Out]{
			Seq: ISequence[In]{
				Name:    x.Profile.Name,
				Service: x.Profile.Service,
				Rank:    int(x.Profile.Rank),
			},
		})
	}

	r.deleteProfile <- ILinks[In, Out]{Links: s, Clean: true}
	return &rumrpc.IDeleteResponse{Succeed: true}, nil
}

// ACTIVATE brings a parked profile (or re-enables it) back into service
func (r *Rum[In, Out]) ACTIVATE(ctx context.Context, in *rumrpc.IActivateRequest) (*rumrpc.IActivateResponse, error) {
	var s = make([]ILink[In, Out], 0, len(in.Activate))

	for _, x := range in.Activate {
		log.Println("unmarhalling: ", string(x.Profile.Input))

		s = append(s, ILink[In, Out]{
			Seq: ISequence[In]{
				Name:    x.Profile.Name,
				Service: x.Profile.Service,
				Rank:    int(x.Profile.Rank),
			},
		})
	}

	r.activateProfile <- ILinks[In, Out]{Links: s, Clean: true}
	return &rumrpc.IActivateResponse{Succeed: true}, nil
}

func (r *Rum[In, Out]) REMOVESERVICE(ctx context.Context, in *rumrpc.IRemoveServiceRequest) (*rumrpc.IRemoveServiceResponse, error) {
	var s = make([]ILink[In, Out], 0, len(in.Delete))

	for _, x := range in.Delete {
		var input In
		if err := json.Unmarshal(x.Profile.Input, &input); err != nil {
			log.Println("unmarshal error:", err)
			continue
		}
		log.Println("unmarhalling: ", string(x.Profile.Input))
		log.Println("input:", input)

		s = append(s, ILink[In, Out]{
			Seq: ISequence[In]{
				Name:  x.Profile.Name,
				Rank:  int(x.Profile.Rank),
				Input: &input,
			},
		})
	}

	r.deleteService <- ILinks[In, Out]{Links: s, Clean: true}

	return &rumrpc.IRemoveServiceResponse{Succeed: true}, nil
}

// DEACTIVATE parks a profile without deleting it
func (r *Rum[In, Out]) DEACTIVATE(ctx context.Context, in *rumrpc.IDeactivateRequest) (*rumrpc.IDeactivateResponse, error) {
	var s = make([]ILink[In, Out], 0, len(in.Deactivate))

	for _, x := range in.Deactivate {

		s = append(s,
			ILink[In, Out]{
				Seq: ISequence[In]{
					Name:    x.Profile.Name,
					Service: x.Profile.Service,
					Rank:    int(x.Profile.Rank),
				},
			},
		)
	}

	r.deactivateProfile <- ILinks[In, Out]{Links: s, Clean: true}
	return &rumrpc.IDeactivateResponse{Succeed: true}, nil
}

func (r *Rum[In, Out]) DEACTIVATESERVICE(ctx context.Context, in *rumrpc.IDeactivateServiceRequest) (*rumrpc.IDeactivateServiceResponse, error) {
	var s = make([]ILink[In, Out], 0, len(in.Delete))

	for _, x := range in.Delete {

		s = append(s,
			ILink[In, Out]{
				Seq: ISequence[In]{
					Name:    x.Profile.Name,
					Service: x.Profile.Service,
					Rank:    int(x.Profile.Rank),
				},
			},
		)
	}

	r.deactivateService <- ILinks[In, Out]{Links: s, Clean: true}
	return &rumrpc.IDeactivateServiceResponse{Succeed: true}, nil
}

func (r *Rum[In, Out]) ACTIVATESERVICE(ctx context.Context, in *rumrpc.IActivateServiceRequest) (*rumrpc.IActivateServiceResponse, error) {
	var s = make([]ILink[In, Out], 0, len(in.Delete))

	for _, x := range in.Delete {

		s = append(s,
			ILink[In, Out]{
				Seq: ISequence[In]{
					Name:    x.Profile.Name,
					Service: x.Profile.Service,
					Rank:    int(x.Profile.Rank),
				},
			},
		)
	}

	r.activateService <- ILinks[In, Out]{Links: s, Clean: true}
	return &rumrpc.IActivateServiceResponse{Succeed: true}, nil
}

func (r *Rum[In, Out]) onPost(seq ISequence[In]) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC RECOVERED in onPost: %v\n%s", r, debug.Stack())
		}
	}()
	kit, err := r.store.GetKit(seq.Name)
	if err != nil {
		log.Println("post error: ", err)
		return
	}

	r.handleDispatch(seq, kit)
}

func (r *Rum[In, Out]) onRemoveService(inprofile, service string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC RECOVERED in onRemoveService: %v\n%s", r, debug.Stack())
		}
	}()
	s, err := r.store.GetKit(inprofile)
	if err != nil {
		return
	}
	s.AddRemoveReport(time.Now())
	s.RemoveService(service)
}

func (r *Rum[In, Out]) onActivateService(inprofile, service string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC RECOVERED in onActivateService: %v\n%s", r, debug.Stack())
		}
	}()
	s, err := r.store.GetKit(inprofile)
	if err != nil {
		return
	}
	s.AddActivateReport(time.Now())
	s.ActivateService(service)
}

func (r *Rum[In, Out]) onDeactivateService(inprofile, service string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC RECOVERED in onDeactivateService: %v\n%s", r, debug.Stack())
		}
	}()
	s, err := r.store.GetKit(inprofile)
	if err != nil {
		return
	}
	s.AddDeactiveReport(time.Now())
	s.DeactivateService(service)
}

func (r *Rum[In, Out]) onActivateProfile(token string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC RECOVERED in onActivateProfile: %v\n%s", r, debug.Stack())
		}
	}()
	if s, err := r.store.GetKit(token); err == nil {
		s.AddActivateReport(time.Now())
	}
	r.store.profile.ActivateProfile(token)
}

func (r *Rum[In, Out]) onDeactivateProfile(token string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC RECOVERED in onDeactivateProfile: %v\n%s", r, debug.Stack())
		}
	}()
	if s, err := r.store.GetKit(token); err == nil {
		s.AddDeactiveReport(time.Now())
	}
	r.store.profile.DeactivateProfile(token)
}

func (r *Rum[In, Out]) onRemoveProfile(token string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC RECOVERED in onRemoveProfile: %v\n%s", r, debug.Stack())
		}
	}()
	if s, err := r.store.GetKit(token); err == nil {
		s.AddRemoveReport(time.Now())
	}
	r.store.profile.RemoveProfile(token)
}
