package example

import (
	"fmt"
	"log"
	"rum/app/dog"
	"time"
)

func PlayDog() {
	ExampleClientBasic()
}

// ExampleClientBasic demonstrates the simplified client API
func ExampleClientBasic() {
	client := dog.NewClient[struct{}](10 * time.Second)
	defer client.Close()

	_, err := client.DefinePolicy("quickOp", 1*time.Second).
		AddFunc("operation1", func() error {
			time.Sleep(200 * time.Millisecond)
			return nil
		}).
		AddFunc("operation2", func() error {
			time.Sleep(300 * time.Millisecond)
			return nil
		}).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	if err := client.LazyExecuteAndReport("quickOp"); err != nil {
		log.Fatal(err)
	}

	log.Println("✅ Done!")
}

// ExampleClientMultiple demonstrates multiple policies
func ExampleClientMultiple() {
	client := dog.NewClient[struct{}](10 * time.Second)
	defer client.Close()

	_, err := client.DefinePolicy("fastOps", 500*time.Millisecond).
		AddFunc("fast1", func() error {
			time.Sleep(100 * time.Millisecond)
			return nil
		}).
		AddFunc("fast2", func() error {
			time.Sleep(150 * time.Millisecond)
			return nil
		}).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.DefinePolicy("slowOps", 1*time.Second).
		AddFunc("slow1", func() error {
			time.Sleep(400 * time.Millisecond)
			return nil
		}).
		AddFunc("slow2", func() error {
			time.Sleep(500 * time.Millisecond)
			return nil
		}).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	reports, err := client.ExecuteMultiple("fastOps", "slowOps")
	if err != nil {
		log.Fatal(err)
	}

	for name, report := range reports {
		log.Printf("\n%s Report:\n", name)
		log.Println(&dog.FormattedReport{Report: report})
	}

	log.Println("✅ All done!")
}

// ExampleClientWithData demonstrates functions that return data
func ExampleClientWithData() {
	type Result struct {
		Value string
	}

	client := dog.NewClient[Result](10 * time.Second)
	defer client.Close()

	_, err := client.DefinePolicy("dataOp", 1*time.Second).
		AddFuncWithReturn("getData", func() (*Result, error) {
			time.Sleep(300 * time.Millisecond)
			return &Result{Value: "success"}, nil
		}).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	report, err := client.ExecuteAndReport("dataOp")
	if err != nil {
		log.Fatal(err)
	}

	log.Println(&dog.FormattedReport{Report: report})
	log.Printf("Output: %s\n", string(report.Output))
}

// ExampleClientRepeated demonstrates repeated execution with reset
func ExampleClientRepeated() {
	client := dog.NewClient[struct{}](10 * time.Second)
	defer client.Close()

	_, err := client.DefinePolicy("repeated", 1*time.Second).
		AddFunc("op", func() error {
			time.Sleep(300 * time.Millisecond)
			return nil
		}).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	for i := 1; i <= 3; i++ {
		log.Printf("\n--- Run %d ---\n", i)

		if err := client.LazyExecuteAndReport("repeated"); err != nil {
			log.Fatal(err)
		}

		if i < 3 {
			if err := client.Reset("repeated"); err != nil {
				log.Fatal(err)
			}
		}
	}
}

// ExampleClientWithMetrics demonstrates accessing metrics
func ExampleClientWithMetrics() {
	client := dog.NewClient[struct{}](10 * time.Second)
	defer client.Close()

	_, err := client.DefinePolicy("cpuIntensive", 2*time.Second).
		AddFunc("heavyOp", func() error {
			sum := 0
			for i := 0; i < 100000000; i++ {
				sum += i
			}
			time.Sleep(500 * time.Millisecond)
			return nil
		}).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	if err := client.LazyExecuteAndReport("cpuIntensive"); err != nil {
		log.Fatal(err)
	}

	metrics := client.GetMetrics("cpuIntensive")
	if metrics != nil {
		log.Printf("\nMetrics Summary:\n")
		log.Printf("  CPU Usage: %.1f%%\n", metrics.CPUUsage)
		log.Printf("  Memory: %.2f MB\n", metrics.AllocMB)
		log.Printf("  Max Memory Seen: %.2f MB\n", metrics.MaxMemorySeenMB)
		log.Printf("  CPU Health: %s\n", metrics.GetCPUHealth())
		log.Printf("  Memory Health: %s\n", metrics.GetMemoryHealth())
		log.Printf("  Thermal Level: %s\n", metrics.GetThermalHealth())
	}
}

// ExampleClientListAndInfo demonstrates listing policies and getting info
func ExampleClientListAndInfo() {
	client := dog.NewClient[struct{}](10 * time.Second)
	defer client.Close()

	for i := 1; i <= 3; i++ {
		name := fmt.Sprintf("policy%d", i)
		_, err := client.DefinePolicy(name, 1*time.Second).
			AddFunc("op", func() error {
				time.Sleep(100 * time.Millisecond)
				return nil
			}).
			Build()
		if err != nil {
			log.Fatal(err)
		}
	}

	policies := client.ListPolicies()
	log.Printf("Registered policies: %v\n", policies)

	for _, policyName := range policies {
		progress := client.GetProgress(policyName)
		metrics := client.GetMetrics(policyName)

		log.Printf("\n%s:\n", policyName)
		if progress != nil {
			log.Printf("  Progress: %d%%\n", progress.GetCompletion())
		}
		if metrics != nil {
			log.Printf("  CPU: %.1f%%\n", metrics.CPUUsage)
		}
	}
}
