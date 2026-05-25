// Package main showcase the framework
package main

import (
	example "rum/app/examples"
	"sync"
)

var wg sync.WaitGroup

func main() {
	example.PlayBasicRumExample()
	// example.PlayAdvancedRumExample()
	// example.PlayTrain()
	// example.PlayRum()
	// example.PlayRumDI()
	// example.PlayRumV2()
	// example.PlayDog()
	// example.PlayInjection()
}
