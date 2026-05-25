// package dog

// import (
// 	"fmt"
// 	"log"
// 	rumpaint "rum/app/paint"
// )

// func (rd *Dog[T]) watchDog() {
// 	log.Println("watch dog  ....")
// 	for {
// 		fmt.Println("[watchDog] Waiting for message...")

// 		select {
// 		case policy := <-rd.register:
// 			rd.registerPolicy(policy)

// 		case name := <-rd.unregister:
// 			rd.unregisterPolicy(name)

// 		case policyName := <-rd.parkDog:
// 			rd.monitorPolicy(policyName)

// 		case done := <-rd.done:
// 			rd.processDone(done)

// 		case bark := <-rd.bark:
// 			rd.processBark(bark)

// 		case policyName := <-rd.reset:
// 			rd.resetPolicy(policyName)

// 		case <-rd.resetAll:
// 			rd.resetAllPolicies()

// 		case <-rd.stopCh:
// 			x := rumpaint.Title("Shutting down... ")
// 			log.Println(x)
// 			t := rumpaint.Header(`

// ██████╗░░█████╗░░██████╗░
// ██╔══██╗██╔══██╗██╔════╝░
// ██║░░██║██║░░██║██║░░██╗░
// ██║░░██║██║░░██║██║░░╚██╗
// ██████╔╝╚█████╔╝╚██████╔╝
// ╚═════╝░░╚════╝░░╚═════╝░
// 			`)
// 			log.Println(t)
// 			return

// 		case <-rd.ctx.Done():
// 			x := rumpaint.Title("context done... ")
// 			log.Println(x)
// 			t := rumpaint.Header(`

// ██████╗░░█████╗░░██████╗░
// ██╔══██╗██╔══██╗██╔════╝░
// ██║░░██║██║░░██║██║░░██╗░
// ██║░░██║██║░░██║██║░░╚██╗
// ██████╔╝╚█████╔╝╚██████╔╝
// ╚═════╝░░╚════╝░░╚═════╝░
// 			`)
// 			log.Println(t)
// 			return
// 		}
// 	}
// }

package dog

import (
	"log"
)

// watchDog is the main event loop for the Dog watchdog
func (rd *Dog[T]) watchDog() {
	defer rd.wg.Done()

	log.Println("🐕 Watchdog started...")

	for {
		select {
		// Registration handling
		case policy := <-rd.register:
			rd.handleRegister(policy)

		// Unregistration handling
		case name := <-rd.unregister:
			rd.handleUnregister(name)

		// Park/Monitor handling
		case policyName := <-rd.parkDog:
			rd.handleParkDog(policyName)

		// Summon/Execute handling
		case policyName := <-rd.summonCh:
			rd.handleSummon(policyName)

		// Completion handling
		case done := <-rd.done:
			rd.handleDone(done)

		// Error handling
		case bark := <-rd.bark:
			rd.handleBark(bark)

		// Reset handling
		case policyName := <-rd.reset:
			rd.handleReset(policyName)

		// Reset all handling
		case <-rd.resetAll:
			rd.handleResetAll()

		// Shutdown
		case <-rd.stopCh:
			log.Println("🛑 Watchdog shutting down...")
			return

		case <-rd.ctx.Done():
			log.Println("🛑 Context cancelled, shutting down...")
			return
		}
	}
}
