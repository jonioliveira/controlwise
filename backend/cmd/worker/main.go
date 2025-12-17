package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/controlewise/backend/internal/config"
	"github.com/controlewise/backend/internal/database"
	"github.com/controlewise/backend/internal/jobs"
	"github.com/controlewise/backend/internal/workflow"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration: ", err)
	}

	// Initialize database
	db, err := database.NewPostgres(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Configure Asynq with Redis
	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	// Create Asynq client for enqueuing jobs
	client := asynq.NewClient(redisOpt)
	defer client.Close()

	// Create workflow engine
	engine := workflow.NewEngine(db, client)

	// Create Asynq server
	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 10,
			// Specify multiple queues with different priorities
			Queues: map[string]int{
				"critical": 6, // High priority - immediate notifications
				"default":  3, // Normal priority - scheduled tasks
				"low":      1, // Low priority - batch operations
			},
			// Retry configuration
			RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
				return time.Duration(n) * time.Minute // Exponential backoff
			},
			// Error handler
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Printf("[Worker] Error processing task %s: %v", task.Type(), err)
			}),
		},
	)

	// Initialize handlers with workflow engine
	handlers := jobs.NewHandlers(db, engine)

	// Create mux for routing tasks to handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(jobs.TypeSendNotification, handlers.HandleSendNotification)
	mux.HandleFunc(jobs.TypeExecuteTrigger, handlers.HandleExecuteTrigger)
	mux.HandleFunc(jobs.TypeCheckTimeTriggers, handlers.HandleCheckTimeTriggers)

	// Start scheduler for periodic tasks
	scheduler := asynq.NewScheduler(redisOpt, nil)

	// Schedule time trigger check every minute
	_, err = scheduler.Register("* * * * *", asynq.NewTask(jobs.TypeCheckTimeTriggers, nil))
	if err != nil {
		log.Fatal("Failed to register scheduled task: ", err)
	}

	// Start scheduler in goroutine
	go func() {
		if err := scheduler.Run(); err != nil {
			log.Printf("Scheduler error: %v", err)
		}
	}()

	// Start worker in goroutine
	go func() {
		log.Println("Asynq worker starting...")
		if err := srv.Run(mux); err != nil {
			log.Fatal("Worker failed to start:", err)
		}
	}()

	log.Println("Workflow worker is running. Press Ctrl+C to exit.")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down worker...")
	srv.Shutdown()
	scheduler.Shutdown()
	log.Println("Worker exited")
}
