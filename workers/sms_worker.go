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
	TypeSmsSend     = "sms:send"
	TypeSmsBulkSend = "sms:bulk_send"
)

// SmsSendPayload contains the data for sending a single SMS
type SmsSendPayload struct {
	Message    string `json:"message"`
	Recipient  string `json:"recipient"`
	OwnerType  string `json:"owner_type"`
	OwnerID    int64  `json:"owner_id"`
	SenderName string `json:"sender_name"`
	Free       bool   `json:"free"` // If true, don't deduct from package
}

// SmsBulkSendPayload contains the data for sending bulk SMS
type SmsBulkSendPayload struct {
	Message    string   `json:"message"`
	Recipients []string `json:"recipients"`
	OwnerType  string   `json:"owner_type"`
	OwnerID    int64    `json:"owner_id"`
	SenderName string   `json:"sender_name"`
	Free       bool     `json:"free"`
}

// SmsWorker handles background SMS processing
type SmsWorker struct {
	client    *asynq.Client
	server    *asynq.Server
	stopChan  chan struct{}
	isRunning bool
}

// NewSmsWorker creates a new SMS worker
func NewSmsWorker() *SmsWorker {
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})

	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 5,
			Queues: map[string]int{
				"sms":     6,
				"default": 3,
				"low":     1,
			},
		},
	)

	return &SmsWorker{
		client:   client,
		server:   server,
		stopChan: make(chan struct{}),
	}
}

// EnqueueSmsSend enqueues a single SMS send task
func (w *SmsWorker) EnqueueSmsSend(payload SmsSendPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeSmsSend, data)

	info, err := w.client.Enqueue(task,
		asynq.Queue("sms"),
		asynq.MaxRetry(3),
		asynq.Timeout(30*time.Second),
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Printf("Enqueued SMS send task: id=%s queue=%s recipient=%s", info.ID, info.Queue, payload.Recipient)
	return nil
}

// EnqueueSmsBulkSend enqueues a bulk SMS send task
func (w *SmsWorker) EnqueueSmsBulkSend(payload SmsBulkSendPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeSmsBulkSend, data)

	info, err := w.client.Enqueue(task,
		asynq.Queue("sms"),
		asynq.MaxRetry(3),
		asynq.Timeout(5*time.Minute), // Longer timeout for bulk
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Printf("Enqueued bulk SMS send task: id=%s queue=%s recipients=%d", info.ID, info.Queue, len(payload.Recipients))
	return nil
}

// SmsSendHandler handles the SMS send task
type SmsSendHandler struct {
	ProcessFunc func(payload SmsSendPayload) error
}

func (h *SmsSendHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload SmsSendPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("Processing SMS send to: %s at %s", payload.Recipient, time.Now().Format(time.RFC3339))

	if h.ProcessFunc != nil {
		if err := h.ProcessFunc(payload); err != nil {
			log.Printf("SMS send failed for %s: %v", payload.Recipient, err)
			return err
		}
	}

	log.Printf("SMS send completed for: %s", payload.Recipient)
	return nil
}

// SmsBulkSendHandler handles the bulk SMS send task
type SmsBulkSendHandler struct {
	ProcessFunc func(payload SmsBulkSendPayload) error
}

func (h *SmsBulkSendHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload SmsBulkSendPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("Processing bulk SMS send to %d recipients at %s", len(payload.Recipients), time.Now().Format(time.RFC3339))

	if h.ProcessFunc != nil {
		if err := h.ProcessFunc(payload); err != nil {
			log.Printf("Bulk SMS send failed: %v", err)
			return err
		}
	}

	log.Printf("Bulk SMS send completed for %d recipients", len(payload.Recipients))
	return nil
}

// StartWorker starts the asynq worker server
func (w *SmsWorker) StartWorker(sendFunc func(SmsSendPayload) error, bulkSendFunc func(SmsBulkSendPayload) error) error {
	mux := asynq.NewServeMux()
	mux.Handle(TypeSmsSend, &SmsSendHandler{ProcessFunc: sendFunc})
	mux.Handle(TypeSmsBulkSend, &SmsBulkSendHandler{ProcessFunc: bulkSendFunc})

	log.Println("Starting SMS worker...")
	return w.server.Run(mux)
}

// Close closes the worker connections
func (w *SmsWorker) Close() error {
	if w.isRunning {
		close(w.stopChan)
		w.isRunning = false
		log.Println("SMS worker stopped")
	}

	if err := w.client.Close(); err != nil {
		log.Printf("Error closing asynq client: %v", err)
	}

	w.server.Shutdown()
	log.Println("SMS asynq server shutdown complete")
	return nil
}

// GetClient returns the asynq client for enqueueing tasks
func (w *SmsWorker) GetClient() *asynq.Client {
	return w.client
}
