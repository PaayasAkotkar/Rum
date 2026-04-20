package dog

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type test struct {
	Input string
}

func ExampleProperDurationTracking() {
	dog := New[test](10 * time.Second)
	dog.Watch()

	defer dog.Shutdown()

	policy := NewPolicy[test](1 * time.Second)
	policy.SetName("properTracking")

	policy.AddFunc(Funcs[test]{
		Name: "operation",
		Rank: 1,
		Void: func() error {
			time.Sleep(300 * time.Millisecond)
			return nil
		},
	})

	if err := dog.Register(policy); err != nil {
		panic(err)
	}

	if err := dog.ParkDog("properTracking"); err != nil {
		panic(err)
	}

	functionStartTime := time.Now()
	err := policy.Fn[0].Void()
	measuredDuration := time.Since(functionStartTime)

	if err != nil {
		dog.Bark(IBark{
			Policy: "properTracking",
			Reason: err.Error(),
		})
	} else {
		dog.Done(IDone{
			PolicyName:   "properTracking",
			FuncName:     "operation",
			Rank:         1,
			FuncDuration: measuredDuration,
		})
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		report := dog.Pakkun("properTracking")
		if report != nil && report.isReady {
			log.Println("report: ", report)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	wg.Wait()

}

func ExampleProperDurationReturnTypeTracking() {
	dog := New[test](10 * time.Second)
	dog.Watch()

	defer dog.Shutdown()

	policy := NewPolicy[test](1 * time.Second)
	policy.SetName("properTracking")

	policy.AddFunc(Funcs[test]{
		Name: "operation",
		Rank: 1,
		Fn: func() (*test, error) {
			time.Sleep(300 * time.Millisecond)
			return &test{Input: "succeed input"}, nil
		},
	})

	if err := dog.Register(policy); err != nil {
		panic(err)
	}

	if err := dog.ParkDog("properTracking"); err != nil {
		panic(err)
	}

	functionStartTime := time.Now()
	resp, err := policy.Fn[0].Fn()
	measuredDuration := time.Since(functionStartTime)
	log.Println("resp: ", resp)

	if err != nil {
		dog.Bark(IBark{
			Policy: "properTracking",
			Reason: err.Error(),
		})
	} else {
		dog.Done(IDone{
			PolicyName:   "properTracking",
			FuncName:     "operation",
			Rank:         1,
			FuncDuration: measuredDuration,
		})
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		report := dog.Pakkun("properTracking")
		if report != nil && report.isReady {
			log.Println("report: ", report)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	wg.Wait()

}

func ExampleProperDurationTrackingError() {
	dog := New[test](10 * time.Second)
	dog.Watch()

	defer dog.Shutdown()

	policy := NewPolicy[test](1 * time.Second)
	policy.SetName("properTracking")

	policy.AddFunc(Funcs[test]{
		Name: "operation",
		Rank: 1,
		Void: func() error {
			time.Sleep(300 * time.Millisecond)
			return errors.New("manually failing track")
		},
	})

	if err := dog.Register(policy); err != nil {
		panic(err)
	}

	if err := dog.ParkDog("properTracking"); err != nil {
		panic(err)
	}

	functionStartTime := time.Now()
	err := policy.Fn[0].Void()
	measuredDuration := time.Since(functionStartTime)

	if err != nil {
		dog.Bark(IBark{
			Policy: "properTracking",
			Reason: err.Error(),
		})
	} else {
		dog.Done(IDone{
			PolicyName:   "properTracking",
			FuncName:     "operation",
			Rank:         1,
			FuncDuration: measuredDuration,
		})
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		report := dog.Pakkun("properTracking")
		if report != nil && report.isReady {
			log.Println("report: ", report)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	wg.Wait()

}

func ExampleMultiplePoliciesConcurrent() {
	dog := New[test](10 * time.Second)
	dog.Watch()

	defer dog.Shutdown()

	var rankCounter int32 = 0

	policy1 := NewPolicy[test](500 * time.Millisecond)
	policy1.SetName("faShutdowns")

	fastRanks := make([]int, 3)
	for i := range 3 {
		rank := int(atomic.AddInt32(&rankCounter, 1))
		fastRanks[i] = rank

		policy1.AddFunc(Funcs[test]{
			Name: fmt.Sprintf("fast-%d", i+1),
			Rank: rank,
			Void: func() error {
				time.Sleep(100 * time.Millisecond)
				return nil
			},
		})
	}

	policy2 := NewPolicy[test](1 * time.Second)
	policy2.SetName("slowOps")

	slowRanks := make([]int, 3)
	for i := range 3 {
		rank := int(atomic.AddInt32(&rankCounter, 1))
		slowRanks[i] = rank

		policy2.AddFunc(Funcs[test]{
			Name: fmt.Sprintf("slow-%d", i+1),
			Rank: rank,
			Void: func() error {
				time.Sleep(300 * time.Millisecond)
				return nil
			},
		})
	}

	dog.Register(policy1)
	dog.Register(policy2)

	dog.ParkDog("faShutdowns")
	dog.ParkDog("slowOps")

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()

		for i := range 3 {

			var fn *Funcs[test]
			rank := fastRanks[i]
			for j := range policy1.Fn {
				if policy1.Fn[j].Rank == rank {
					log.Println("found")
					fn = &policy1.Fn[j]
					break
				}
			}

			if fn == nil {
				return
			}

			start := time.Now()
			err := fn.Void()
			duration := time.Since(start)

			if err != nil {
				dog.Bark(IBark{Policy: "faShutdowns", Reason: err.Error()})
			} else {
				log.Println("done")
				dog.Done(IDone{
					PolicyName:   "faShutdowns",
					FuncName:     fn.Name,
					Rank:         rank,
					FuncDuration: duration,
				})
			}
		}
	}()

	go func() {
		defer wg.Done()
		for i := range 3 {

			var fn *Funcs[test]
			rank := slowRanks[i]
			for j := range policy2.Fn {
				if policy2.Fn[j].Rank == rank {
					fn = &policy2.Fn[j]
					log.Println("found")
					break
				}
			}

			if fn == nil {
				continue
			}

			start := time.Now()
			err := fn.Void()
			duration := time.Since(start)

			if err != nil {
				dog.Bark(IBark{Policy: "slowOps", Reason: err.Error()})
			} else {
				log.Println("done")
				dog.Done(IDone{
					PolicyName:   "slowOps",
					FuncName:     fn.Name,
					Rank:         rank,
					FuncDuration: duration,
				})
			}

		}
	}()

	go func() {
		defer wg.Done()
		report := dog.Pakkun("faShutdowns")
		if report != nil && report.isReady {
			log.Println("report: ", report)
		}
	}()

	go func() {
		defer wg.Done()
		report := dog.Pakkun("slowOps")
		if report != nil && report.isReady {
			log.Println("report: ", report)
		}
	}()

	time.Sleep(200 * time.Millisecond)

	wg.Wait()
}
