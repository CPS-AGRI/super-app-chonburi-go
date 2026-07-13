package usecase

import (
	"context"
	"log"
	"sync"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type FCMPayload struct {
	Tokens     []string
	Title      string
	Body       string
	Data       map[string]string
	RetryCount int
}

type FCMWorkerPool struct {
	queue       chan FCMPayload
	workerCount int
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	isRunning   bool
	mu          sync.Mutex
	fcmClient   *messaging.Client
}

var GlobalFCMWorkerPool *FCMWorkerPool

func InitGlobalFCMWorkerPool(queueSize int, workerCount int) *FCMWorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &FCMWorkerPool{
		queue:       make(chan FCMPayload, queueSize),
		workerCount: workerCount,
		ctx:         ctx,
		cancel:      cancel,
	}

	opt := option.WithCredentialsFile("firebase-service-account.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("[FCM-WorkerPool] ข้อผิดพลาดร้ายแรง: ไม่สามารถเริ่มตัว Firebase App ได้ (%v)", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("[FCM-WorkerPool] ข้อผิดพลาดร้ายแรง: ไม่สามารถสร้าง Messaging Client ได้ (%v)", err)
	}

	pool.fcmClient = client
	log.Println("[FCM-WorkerPool] เชื่อมต่อกับ Google FCM API จริงสำเร็จ (Strict Live Mode)")

	GlobalFCMWorkerPool = pool
	pool.Start()
	return pool
}

func (p *FCMWorkerPool) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.isRunning {
		return
	}
	p.isRunning = true

	log.Printf("[FCM-WorkerPool] กำลังเปิดทำงานจำนวน %d Workers (Queue Size: %d)...", p.workerCount, cap(p.queue))

	for i := 1; i <= p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

func (p *FCMWorkerPool) worker(workerID int) {
	defer p.wg.Done()
	log.Printf("[FCM-Worker-%d] สแตนด์บายพร้อมประมวลผลงาน", workerID)

	for {
		select {
		case payload, ok := <-p.queue:
			if !ok {

				log.Printf("[FCM-Worker-%d] สิ้นสุดการทำงานเนื่องจากคิวปิดตัว", workerID)
				return
			}

			p.sendFCM(workerID, payload)

		case <-p.ctx.Done():

			log.Printf("[FCM-Worker-%d] ยกเลิกการทำงานตามระบบแจ้งเตือน", workerID)
			return
		}
	}
}

func (p *FCMWorkerPool) Submit(payload FCMPayload) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.isRunning {
		log.Println("[FCM-WorkerPool] ข้อผิดพลาด: พยายามส่งงานเข้าคิวแต่ระบบปิดใช้งานอยู่")
		return false
	}

	select {
	case p.queue <- payload:

		return true
	default:

		log.Println("[FCM-WorkerPool] คำเตือน: คิวหลักเต็มกำลังทำการรันแบบ Asynchronous Fallback...")
		go func() {

			p.sendFCM(-1, payload)
		}()
		return true
	}
}

func (p *FCMWorkerPool) sendFCM(workerID int, payload FCMPayload) {
	startTime := time.Now()

	for _, token := range payload.Tokens {
		msg := &messaging.Message{
			Token: token,
			Notification: &messaging.Notification{
				Title: payload.Title,
				Body:  payload.Body,
			},
			Data: payload.Data,
		}

		res, err := p.fcmClient.Send(p.ctx, msg)
		if err != nil {
			log.Printf("[FCM-Worker-%d] [LIVE-ERROR] ส่งล้มเหลว -> Token: [%s...], Error: %v",
				workerID, safeTokenPrefix(token), err)
		} else {
			log.Printf("[FCM-Worker-%d] [LIVE-SUCCESS] ส่งผลสำเร็จ -> MessageID: %s, Token: [%s...], หัวข้อ: %s, เวลาทำงาน: %v",
				workerID, res, safeTokenPrefix(token), payload.Title, time.Since(startTime))
		}
	}
}

func safeTokenPrefix(token string) string {
	if len(token) > 8 {
		return token[:8]
	}
	return token
}

func (p *FCMWorkerPool) Stop() {
	p.mu.Lock()
	if !p.isRunning {
		p.mu.Unlock()
		return
	}
	p.isRunning = false
	p.mu.Unlock()

	log.Println("[FCM-WorkerPool] กำลังทำความสะอาดคิวและสั่งหยุดการทำงาน (Graceful Shutdown)...")
	p.cancel()
	close(p.queue)
	p.wg.Wait()
	log.Println("[FCM-WorkerPool] หยุดการทำงานอย่างสมบูรณ์ ปราศจาก Goroutine leak")
}
