package rum

import (
	"log"
	rumpaint "rum/app/paint"
	"runtime/debug"
)

// Hub all gRPC-triggered events are dispatched here sequentially.
// write() goroutines call handle funcs directly for internal lifecycle events,
// so they never block waiting on these channels.
func (r *Rum[In, Out]) Hub() {
	log.Println("listening 🐴")

	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("PANIC RECOVERED in Hub: %v\n%s", r, debug.Stack())
				}
			}()
			select {

			case token := <-r.post:
				log.Println("in listening post")
				for _, l := range token.Links {
					r.onPost(l.Seq)
				}

				// profile-modfiy

			case token := <-r.deactivateProfile:
				for _, l := range token.Links {
					r.onDeactivateProfile(l.Seq.Name)
				}

			case token := <-r.activateProfile:
				for _, l := range token.Links {
					r.onActivateProfile(l.Seq.Name)
				}

			case token := <-r.deleteProfile:
				for _, l := range token.Links {
					r.onRemoveProfile(l.Seq.Name)
				}

				// end

				// service-modify
			case token := <-r.deleteService:
				for _, l := range token.Links {
					r.onRemoveService(l.Seq.Name, l.Seq.Service)
				}

			case token := <-r.activateService:
				for _, l := range token.Links {
					r.onActivateService(l.Seq.Name, l.Seq.Service)
				}

			case token := <-r.deactivateService:
				for _, l := range token.Links {
					r.onDeactivateService(l.Seq.Name, l.Seq.Service)
				}
				// end
			case <-r.ctx.Done():
				r.wg.Wait()
				t := rumpaint.Header(`
██████╗░██╗░░░██╗███╗░░░███╗
██╔══██╗██║░░░██║████╗░████║
██████╔╝██║░░░██║██╔████╔██║
██╔══██╗██║░░░██║██║╚██╔╝██║
██║░░██║╚██████╔╝██║░╚═╝░██║
╚═╝░░╚═╝░╚═════╝░╚═╝░░░░░╚═╝
	`)
				log.Println(t)
				return
			}
		}()
	}
}
