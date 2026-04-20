package dog

import (
	"fmt"
	"log"
	rumpaint "rum/app/paint"
)

func (rd *Dog[T]) watchDog() {
	log.Println("watch dog  ....")
	for {
		fmt.Println("[watchDog] Waiting for message...")

		select {
		case policy := <-rd.register:
			rd.registerPolicy(policy)

		case name := <-rd.unregister:
			rd.unregisterPolicy(name)

		case policyName := <-rd.parkDog:
			rd.monitorPolicy(policyName)

		case done := <-rd.done:
			rd.processDone(done)

		case bark := <-rd.bark:
			rd.processBark(bark)

		case policyName := <-rd.reset:
			rd.resetPolicy(policyName)

		case <-rd.resetAll:
			rd.resetAllPolicies()

		case <-rd.stopCh:
			x := rumpaint.Title("Shutting down... ")
			log.Println(x)
			t := rumpaint.Header(`

██████╗░░█████╗░░██████╗░
██╔══██╗██╔══██╗██╔════╝░
██║░░██║██║░░██║██║░░██╗░
██║░░██║██║░░██║██║░░╚██╗
██████╔╝╚█████╔╝╚██████╔╝
╚═════╝░░╚════╝░░╚═════╝░
			`)
			log.Println(t)
			return

		case <-rd.ctx.Done():
			x := rumpaint.Title("context done... ")
			log.Println(x)
			t := rumpaint.Header(`

██████╗░░█████╗░░██████╗░
██╔══██╗██╔══██╗██╔════╝░
██║░░██║██║░░██║██║░░██╗░
██║░░██║██║░░██║██║░░╚██╗
██████╔╝╚█████╔╝╚██████╔╝
╚═════╝░░╚════╝░░╚═════╝░
			`)
			log.Println(t)
			return
		}
	}
}
