// Package rum ....
// flow:
// register -> Profile+ service -> save via name & perform modifcation
package rum

import (
	"cmp"
	"fmt"
	rumstack "rum/app/stack"
	"slices"
	"sync"
	"time"
)

// IProfile manages active and inactive Kits keyed by ISequence[In]
type IProfile[In, Out any] struct {
	mu              sync.Mutex
	activeProfile   map[string]map[ISequence[In]]*Kit[In, Out]
	inactiveProfile map[string]map[ISequence[In]]*Kit[In, Out]
	profileStack    rumstack.Stack[string]
}

func NewProfile[In, Out any]() *IProfile[In, Out] {
	return &IProfile[In, Out]{
		activeProfile:   make(map[string]map[ISequence[In]]*Kit[In, Out]),
		inactiveProfile: make(map[string]map[ISequence[In]]*Kit[In, Out]),
	}
}

// get funcs

func (r *IProfile[In, Out]) GetKit(name string) (*Kit[In, Out], error) {
	if kit, ok := r.activeProfile[name]; ok {
		for _, v := range kit {
			return v, nil
		}
	}
	return nil, fmt.Errorf("profile %v not found or inactive", name)
}

func (r *IProfile[In, Out]) Kits() []*Kit[In, Out] {
	keys := r.profileStack.Range(r.profileStack.Len())
	out := make([]*Kit[In, Out], 0, len(keys))
	for _, k := range keys {
		if kit, ok := r.activeProfile[k]; ok {
			for _, v := range kit {
				out = append(out, v)
			}
		}
	}
	return out
}

func (r *IProfile[In, Out]) ActiveProfileKeys() []string {
	return r.profileStack.Range(r.profileStack.Len())
}

// Sort returns services for a profile sorted by rank (ascending)
func (r *IProfile[In, Out]) Sort(name string) []*Service[In, Out] {
	kit, err := r.GetKit(name)
	if err != nil {
		return nil
	}
	svcs := kit.GetServices()
	out := make([]*Service[In, Out], 0, len(svcs))
	for _, s := range svcs {
		out = append(out, s)
	}
	slices.SortFunc(out, func(a, b *Service[In, Out]) int {
		return cmp.Compare(a.Rank, b.Rank)
	})
	return out
}

// end

// set funcs

func (r *IProfile[In, Out]) RegisterProfile(profile ISequence[In], base time.Duration, kit *Kit[In, Out]) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.activeProfile[profile.Name]; !ok {
		r.activeProfile[profile.Name] = make(map[ISequence[In]]*Kit[In, Out])
	}
	r.activeProfile[profile.Name][profile] = kit
	r.profileStack.Push(profile.Name)
}

// UpdateProfileServices either adds or replaces the old one
func (r *IProfile[In, Out]) UpdateProfileServices(profile ISequence[In], services map[string]*Service[In, Out]) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.activeProfile[profile.Name]; ok {
		var prf = r.activeProfile[profile.Name][profile]
		for n, s := range services {
			prf.PushService(n, s)
		}
	}
}

func (r *IProfile[In, Out]) PushProfile(profile ISequence[In], kit *Kit[In, Out]) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.activeProfile[profile.Name]; !ok {
		r.activeProfile[profile.Name] = make(map[ISequence[In]]*Kit[In, Out])
	}
	r.activeProfile[profile.Name][profile] = kit
	r.profileStack.Push(profile.Name)
}

func (r *IProfile[In, Out]) DeactivateProfile(key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k, ok := r.activeProfile[key]
	if !ok {
		return fmt.Errorf("profile %v not active", key)
	}
	r.inactiveProfile[key] = k
	delete(r.activeProfile, key)
	r.profileStack.Erase(key)
	return nil
}

func (r *IProfile[In, Out]) ActivateProfile(key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k, ok := r.inactiveProfile[key]
	if !ok {
		return fmt.Errorf("profile %v not in inactive pool", key)
	}
	r.activeProfile[key] = k
	delete(r.inactiveProfile, key)
	r.profileStack.Push(key)
	return nil
}

func (r *IProfile[In, Out]) RemoveProfile(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.activeProfile, key)
	delete(r.inactiveProfile, key)
	r.profileStack.Erase(key)
}

// end
