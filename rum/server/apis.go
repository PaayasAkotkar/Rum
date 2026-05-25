package rum

import "log"

// Poll monitors the profile and returns a channel of results
// never closes
func (r *Rum[In, Out]) Poll(profile ISequence[In]) <-chan *IDispatchResult {
	log.Println("in tick fetch")
	return r.tickFetchPoll(profile)
}

// Paper fetches single request
func (r *Rum[In, Out]) Paper(profile ISequence[In]) *IDispatchResult {
	return r.fetch(profile)
}
