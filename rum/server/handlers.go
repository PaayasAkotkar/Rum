package rum

import (
	"log"
	"runtime/debug"
	"time"
)

func (r *Rum[In, Out]) handleDispatch(seq ISequence[In], kit *Kit[In, Out]) {
	svc := kit.read()
	if svc == nil {
		log.Println("post error empty kit")
		return
	}

	f := svc.GetFormat()
	if f != nil && f.Call > 0 {
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("PANIC RECOVERED in handleDispatch goroutine: %v\n%s", r, debug.Stack())
				}
			}()
			select {
			case <-time.After(f.Call):
				r.write(seq, svc)
			case <-r.ctx.Done():
			}
		}()
		return
	}
	r.write(seq, svc)
}

// func (r *Rum[In, Out]) handleServiceFormat(profile ISequence[In], svc *Service[In, Out]) {
// 	f := svc.GetFormat()
// 	if f == nil {
// 		return
// 	}

// 	x := ISequence[In]{Name: profile.Name, Rank: profile.Rank, Service: svc.GetName()}

// 	if f.ShouldRemove() {
// 		r.onRemoveService(x.Name, x.Service)
// 	} else if f.ShouldDeactivate() {
// 		r.onDeactivateService(x.Name, x.Service)
// 		if sleep := f.GetActivateTime(); sleep != nil {
// 			r.wg.Add(1)
// 			go func() {
// 				defer r.wg.Done()
// 				defer func() {
// 					if r := recover(); r != nil {
// 						log.Printf("PANIC RECOVERED in handleServiceFormat goroutine: %v\n%s", r, debug.Stack())
// 					}
// 				}()
// 				time.Sleep(*sleep)
// 				r.onActivateProfile(x.Name)
// 			}()
// 		}
// 	}
// }

// func (r *Rum[In, Out]) handleProfileFormat(profile ISequence[In], kit *Kit[In, Out]) {
// 	f := kit.GetFormat()
// 	if f == nil {
// 		return
// 	}
// 	x := ISequence[In]{Name: profile.Name, Rank: profile.Rank}

// 	if f.ShouldRemove() {
// 		r.onRemoveProfile(x.Name)
// 	} else if f.ShouldDeactivate() {
// 		r.onDeactivateProfile(x.Name)
// 		if sleep := f.GetActivateTime(); sleep != nil {
// 			r.wg.Add(1)
// 			go func() {
// 				defer r.wg.Done()
// 				defer func() {
// 					if r := recover(); r != nil {
// 						log.Printf("PANIC RECOVERED in handleProfileFormat goroutine: %v\n%s", r, debug.Stack())
// 					}
// 				}()
// 				time.Sleep(*sleep)
// 				r.activateProfile <- ILinks[In, Out]{Links: []ILink[In, Out]{{Seq: x}}, Clean: true}
// 			}()
// 		}
// 	}
// }
