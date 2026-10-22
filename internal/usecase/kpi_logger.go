package usecase

import (
	"log"
	"sync"
	"time"

	"super-app-chonburi-go/internal/domain"
)

type asyncKPILogger struct {
	repo       domain.KPILogRepository
	queue      chan *domain.KPILog
	wg         sync.WaitGroup
	stopSignal chan struct{}
	once       sync.Once
}

const (
	defaultQueueSize   = 10000
	defaultBatchSize   = 50
	defaultFlushPeriod = 1 * time.Second
	defaultWorkerCount = 3
)

// NewAsyncKPILogger initializes an ultra high-performance non-blocking event worker pool
func NewAsyncKPILogger(repo domain.KPILogRepository) domain.KPILogger {
	logger := &asyncKPILogger{
		repo:       repo,
		queue:      make(chan *domain.KPILog, defaultQueueSize),
		stopSignal: make(chan struct{}),
	}

	for i := 0; i < defaultWorkerCount; i++ {
		logger.wg.Add(1)
		go logger.workerLoop(i + 1)
	}

	return logger
}

func (l *asyncKPILogger) LogEvent(event *domain.KPILog) {
	if event == nil {
		return
	}
	if event.CreatedDate.IsZero() {
		event.CreatedDate = time.Now().UTC()
	} else {
		event.CreatedDate = event.CreatedDate.UTC()
	}
	if event.Metadata == "" {
		event.Metadata = "{}"
	}

	select {
	case l.queue <- event:
	default:
		log.Printf("⚠️ [KPILogger] Buffer queue full! Dropped event: %s (Entity: %s)", event.EventCode, event.EntityID)
	}
}

func (l *asyncKPILogger) workerLoop(workerID int) {
	defer l.wg.Done()

	ticker := time.NewTicker(defaultFlushPeriod)
	defer ticker.Stop()

	batch := make([]domain.KPILog, 0, defaultBatchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := l.repo.CreateBatch(batch); err != nil {
			log.Printf("❌ [KPILogger-Worker %d] Failed to batch insert %d logs: %v", workerID, len(batch), err)
		}
		batch = make([]domain.KPILog, 0, defaultBatchSize)
	}

	for {
		select {
		case <-l.stopSignal:
			// Drain remaining queue before shutdown
			for len(l.queue) > 0 {
				item := <-l.queue
				batch = append(batch, *item)
				if len(batch) >= defaultBatchSize {
					flush()
				}
			}
			flush()
			return

		case item := <-l.queue:
			batch = append(batch, *item)
			if len(batch) >= defaultBatchSize {
				flush()
			}

		case <-ticker.C:
			flush()
		}
	}
}

func (l *asyncKPILogger) Close() {
	l.once.Do(func() {
		close(l.stopSignal)
		l.wg.Wait()
		close(l.queue)
	})
}
