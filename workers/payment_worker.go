package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hibiken/asynq"
)

const (
	TypePaymentProcess = "payment:process"
)

// PaymentProcessPayload contains the data for processing a payment
type PaymentProcessPayload struct {
	PaymentID uint `json:"payment_id"`
}

// PaymentWorker handles background payment processing
type PaymentWorker struct {
	client    *asynq.Client
	server    *asynq.Server
	stopChan  chan struct{} // Channel to signal status checker to stop
	isRunning bool
}

// NewPaymentWorker creates a new payment worker
func NewPaymentWorker() *PaymentWorker {
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})

	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	return &PaymentWorker{
		client:   client,
		server:   server,
		stopChan: make(chan struct{}),
	}
}

// EnqueuePaymentProcess enqueues a payment processing task
func (w *PaymentWorker) EnqueuePaymentProcess(paymentID uint) error {
	payload, err := json.Marshal(PaymentProcessPayload{PaymentID: paymentID})
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypePaymentProcess, payload)

	// Process after 2 seconds delay (like Ruby's sleep 2)
	info, err := w.client.Enqueue(task,
		asynq.ProcessIn(2*time.Second),
		asynq.Queue("critical"),
		asynq.MaxRetry(3),
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Printf("Enqueued payment processing task: id=%s queue=%s", info.ID, info.Queue)
	return nil
}

// HandlePaymentProcess handles the payment processing task
type PaymentProcessHandler struct {
	ProcessFunc func(paymentID uint) error
}

func (h *PaymentProcessHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload PaymentProcessPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("Starting payment processing for ID: %d at %s", payload.PaymentID, time.Now().Format(time.RFC3339))

	if h.ProcessFunc != nil {
		if err := h.ProcessFunc(payload.PaymentID); err != nil {
			log.Printf("Payment processing failed for ID %d: %v", payload.PaymentID, err)
			return err
		}
	}

	log.Printf("Payment processing completed for ID: %d", payload.PaymentID)
	return nil
}

// StartWorker starts the asynq worker server
func (w *PaymentWorker) StartWorker(processFunc func(paymentID uint) error) error {
	mux := asynq.NewServeMux()
	mux.Handle(TypePaymentProcess, &PaymentProcessHandler{ProcessFunc: processFunc})

	log.Println("Starting payment worker...")
	return w.server.Run(mux)
}

// Close closes the worker connections and stops the status checker
func (w *PaymentWorker) Close() error {
	// Signal status checker to stop
	if w.isRunning {
		close(w.stopChan)
		w.isRunning = false
		log.Println("Status checker stopped")
	}

	// Close asynq client
	if err := w.client.Close(); err != nil {
		log.Printf("Error closing asynq client: %v", err)
	}

	// Shutdown asynq server
	w.server.Shutdown()
	log.Println("Asynq server shutdown complete")
	return nil
}

// GetClient returns the asynq client for enqueueing tasks
func (w *PaymentWorker) GetClient() *asynq.Client {
	return w.client
}

// StatusCheckerFunc is the function type for checking payment statuses
type StatusCheckerFunc func() error

// PaymentProcessorFunc is the function type for processing created payments
type PaymentProcessorFunc func() error

// StartPaymentProcessor starts a background goroutine that processes "created" payments every 3 seconds
// This sends payments to Hubtel and changes their status from "created" to "pending"
// Set ENABLE_PAYMENT_PROCESSOR=true in .env to enable this feature
func (w *PaymentWorker) StartPaymentProcessor(processFunc PaymentProcessorFunc) {
	// Check if payment processor is enabled via environment variable
	enableProcessor := os.Getenv("ENABLE_PAYMENT_PROCESSOR")
	if enableProcessor != "true" && enableProcessor != "1" {
		log.Println("Payment processor is DISABLED. Set ENABLE_PAYMENT_PROCESSOR=true to enable.")
		return
	}

	// Mark as running
	w.isRunning = true

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		log.Println("Payment processor is ENABLED (runs every 3 seconds)...")

		// Run immediately on start, then every 3 seconds
		log.Println("Running initial payment processor...")
		if err := processFunc(); err != nil {
			log.Printf("Error processing created payments: %v", err)
		}

		for {
			select {
			case <-w.stopChan:
				log.Println("Payment processor received stop signal")
				return
			case <-ticker.C:
				log.Println("Processing created payments...")
				if err := processFunc(); err != nil {
					log.Printf("Error processing created payments: %v", err)
				}
			}
		}
	}()
}

// StartStatusChecker starts a background goroutine that checks pending payment statuses every 20 seconds
// Set ENABLE_PAYMENT_STATUS_CHECKER=true in .env to enable this feature
func (w *PaymentWorker) StartStatusChecker(checkFunc StatusCheckerFunc) {
	// Check if status checker is enabled via environment variable
	enableChecker := os.Getenv("ENABLE_PAYMENT_STATUS_CHECKER")
	if enableChecker != "true" && enableChecker != "1" {
		log.Println("Payment status checker is DISABLED. Set ENABLE_PAYMENT_STATUS_CHECKER=true to enable.")
		return
	}

	// Mark as running
	w.isRunning = true

	go func() {
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()

		log.Println("Payment status checker is ENABLED (runs every 10 seconds)...")

		// Run immediately on start, then every 20 seconds
		log.Println("Running initial payment status check...")
		if err := checkFunc(); err != nil {
			log.Printf("Error checking payment statuses: %v", err)
		}

		for {
			select {
			case <-w.stopChan:
				log.Println("Status checker received stop signal")
				return
			case <-ticker.C:
				log.Println("Checking pending payment statuses...")
				if err := checkFunc(); err != nil {
					log.Printf("Error checking payment statuses: %v", err)
				}
			}
		}
	}()
}
