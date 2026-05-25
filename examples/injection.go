package example

import (
	"context"
	"fmt"
	"log"
	"reflect"
	injection "rum/app/di"
	"time"
)

// DatabaseService is a dummy service simulating a database connection
type DatabaseService struct {
	ConnectionID string
}

func (db *DatabaseService) Query(query string) string {
	return fmt.Sprintf("Executing '%s' on connection %s", query, db.ConnectionID)
}

// UserService is a dummy service depending on DatabaseService
type UserService struct {
	DB *DatabaseService
}

func (us *UserService) GetUser(id int) string {
	return us.DB.Query(fmt.Sprintf("SELECT * FROM users WHERE id = %d", id))
}

// WorkerService is a dummy service simulating a pooled worker
type WorkerService struct {
	WorkerID int
}

// PlayInjection demonstrates how to use the dependency-injection package cleanly
func PlayInjection() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// exampleSingletonAndTransient(ctx)

	examplePooled(ctx)

	// exampleRebuild(ctx)
}

func exampleSingletonAndTransient(ctx context.Context) {
	client := injection.NewClient(ctx, "node-singleton")

	dbServiceType := reflect.TypeOf((*DatabaseService)(nil))
	userServiceType := reflect.TypeOf((*UserService)(nil))

	// Register Singleton
	client.AddSingleton(dbServiceType, injection.Factory{
		Fn: func(ctx context.Context, c *injection.Container) (any, error) {
			log.Println("Initializing Singleton DatabaseService...")
			return &DatabaseService{ConnectionID: "DB-CONN-001"}, nil
		},
	})

	// Register Transient
	client.AddTransient(userServiceType, injection.Factory{
		Fn: func(ctx context.Context, c *injection.Container) (any, error) {
			log.Println("Initializing Transient UserService...")
			dbAny, err := c.GetService(dbServiceType)
			if err != nil {
				return nil, fmt.Errorf("failed to get DatabaseService: %w", err)
			}
			return &UserService{DB: dbAny.(*DatabaseService)}, nil
		},
	})

	// Subscribe to the build status event via Chakra
	buildStatusCh := client.BuildStatus()
	defer client.CloseBuildStatus(buildStatusCh)

	go func() {
		if err := client.Build(ctx); err != nil {
			log.Fatalf("Build failed: %v", err)
		}
	}()

	log.Println("Waiting for services to be built via Chakra pub/sub...")
	status := <-buildStatusCh
	log.Printf("Received Chakra event: %s. Now we are ready to get services!\n", *status)

	log.Println("Retrieving UserService for Request 1...")
	u1, _ := client.GetService(userServiceType)
	log.Println("Request 1 Result:", u1.(*UserService).GetUser(42))

	log.Println("Retrieving UserService for Request 2...")
	u2, _ := client.GetService(userServiceType)
	log.Println("Request 2 Result:", u2.(*UserService).GetUser(99))

	client.Stop()
}

func examplePooled(ctx context.Context) {
	client := injection.NewClient(ctx, "node-pooled")
	workerServiceType := reflect.TypeOf((*WorkerService)(nil))

	workerCounter := 0
	client.AddPooled(workerServiceType, injection.Factory{
		Fn: func(ctx context.Context, c *injection.Container) (any, error) {
			workerCounter++
			log.Printf("Initializing Pooled WorkerService %d...\n", workerCounter)
			return &WorkerService{WorkerID: workerCounter}, nil
		},
	}, &injection.PoolConfig{
		MinConnections:    2,
		MaxConnections:    5,
		ConnectionTimeout: 2 * time.Second,
	})

	status := client.BuildStatus()
	defer client.CloseBuildStatus(status)

	go func() {
		if err := client.Build(ctx); err != nil {
			log.Fatalf("Build failed: %v", err)
		}
	}()

	log.Printf("Received Chakra event: %s. Now we are ready to get services!\n", *<-status)
	log.Println("Retrieving Pooled WorkerService 1...")
	w1, _ := client.GetService(workerServiceType)
	log.Printf("Got Worker %d from pool\n", w1.(*WorkerService).WorkerID)

	log.Println("Retrieving Pooled WorkerService 2...")
	w2, _ := client.GetService(workerServiceType)
	log.Printf("Got Worker %d from pool\n", w2.(*WorkerService).WorkerID)

	log.Println("Returning Worker 1 to pool...")
	client.ReturnPooledService(workerServiceType, w1)

	client.Stop()
}

func exampleRebuild(ctx context.Context) {
	client := injection.NewClient(ctx, "node-rebuild")
	dbServiceType := reflect.TypeOf((*DatabaseService)(nil))

	client.AddSingleton(dbServiceType, injection.Factory{
		Fn: func(ctx context.Context, c *injection.Container) (any, error) {
			log.Println("Initializing DB for Scale Event...")
			return &DatabaseService{ConnectionID: "DB-SCALE"}, nil
		},
	})

	go func() {
		if err := client.Build(ctx); err != nil {
			log.Fatalf("Build failed: %v", err)
		}
	}()

	status := client.BuildStatus()
	defer client.CloseBuildStatus(status)
	log.Printf("Received Chakra event: %s. Now we are ready to get services!\n", *<-status)

	log.Println("Triggering Rebuild (simulating scale event)...")
	if err := client.TriggerRebuild(); err != nil {
		log.Printf("Failed to trigger rebuild: %v\n", err)
	}

	time.Sleep(100 * time.Millisecond)

	client.Stop()
}
