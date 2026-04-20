// Package main showcase the framework
// note: make sure to pass the minIO id of access id and pass or access key in the compose.yml
package main

import (
	"sync"
)

var wg sync.WaitGroup

func main() {
	playRum()
	// playDog()
	// playHybridSearch()
	// playSearch()
}
