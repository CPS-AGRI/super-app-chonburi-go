package usecase

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/database"
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
	deadTokenCh chan string
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
		deadTokenCh: make(chan string, 1000),
		workerCount: workerCount,
		ctx:         ctx,
		cancel:      cancel,
	}

	var client *messaging.Client
	// Check if firebase-service-account.json exists
	if _, err := os.Stat("firebase-service-account.json"); err == nil {
		opt := option.WithCredentialsFile("firebase-service-account.json")
		app, err := firebase.NewApp(ctx, nil, opt)
		if err != nil {
			log.Printf("[FCM-WorkerPool] ⚠️ คำเตือน: ไม่สามารถเริ่มตัว Firebase App ได้ (%v)", err)
		} else {
			c, err := app.Messaging(ctx)
			if err != nil {
				log.Printf("[FCM-WorkerPool] ⚠️ คำเตือน: ไม่สามารถสร้าง Messaging Client ได้ (%v)", err)
			} else {
				client = c
				log.Println("[FCM-WorkerPool] เชื่อมต่อกับ Google FCM API จริงสำเร็จ (Strict Live Mode)")
			}
		}
	} else {
		log.Println("[FCM-WorkerPool] ⚠️ คำเตือน: ไม่พบไฟล์ firebase-service-account.json, ระบบจะทำงานในโหมด MOCK PUSH (Dry-Run)")
	}

	pool.fcmClient = client

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

	log.Printf("[FCM-WorkerPool] กำลังเปิดทำงานจำนวน %d Workers (Queue Size: %d, DeadToken Channel: %d)...",
		p.workerCount, cap(p.queue), cap(p.deadTokenCh))

	p.wg.Add(1)
	go p.deadTokenPurger()

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

func (p *FCMWorkerPool) deadTokenPurger() {
	defer p.wg.Done()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	var batch []string
	seen := make(map[string]bool)

	for {
		select {
		case <-p.ctx.Done():
			if len(batch) > 0 {
				p.flushDeadTokens(batch)
			}
			return
		case token, ok := <-p.deadTokenCh:
			if !ok {
				if len(batch) > 0 {
					p.flushDeadTokens(batch)
				}
				return
			}
			if !seen[token] {
				seen[token] = true
				batch = append(batch, token)
			}
			if len(batch) >= 100 {
				p.flushDeadTokens(batch)
				batch = batch[:0]
				seen = make(map[string]bool)
			}
		case <-ticker.C:
			if len(batch) > 0 {
				p.flushDeadTokens(batch)
				batch = batch[:0]
				seen = make(map[string]bool)
			}
		}
	}
}

func (p *FCMWorkerPool) flushDeadTokens(tokens []string) {
	if len(tokens) == 0 || database.DB == nil {
		return
	}
	db := database.DB

	// ลบ token ที่ตายแล้วออกจาก user_fcm_tokens
	if err := db.Where("token IN ?", tokens).Delete(&domain.UserFCMToken{}).Error; err != nil {
		log.Printf("[FCM-DeadTokenPurger] ไม่สามารถลบ Dead Token จาก user_fcm_tokens: %v", err)
	}

	// ลบ token ที่ตายแล้วออกจาก module_device_tokens
	if err := db.Where("token IN ?", tokens).Delete(&domain.ModuleDeviceToken{}).Error; err != nil {
		log.Printf("[FCM-DeadTokenPurger] ไม่สามารถลบ Dead Token จาก module_device_tokens: %v", err)
	}

	log.Printf("[FCM-DeadTokenPurger] 🧹 ลบ Invalid/Unregistered Token สำเร็จ %d รายการ", len(tokens))
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
		// ป้องกัน OOM: Bounded Backpressure โดย Graceful Drop พร้อมบันทึก Log ชัดเจน
		log.Printf("[FCM-WorkerPool] ⚠️ คำเตือน: คิวหลักเต็ม (capacity=%d) — Drop payload: หัวข้อ='%s', จำนวนอุปกรณ์=%d",
			cap(p.queue), payload.Title, len(payload.Tokens))
		return false
	}
}

func isDeadToken(err error) bool {
	if err == nil {
		return false
	}
	if messaging.IsUnregistered(err) || messaging.IsRegistrationTokenNotRegistered(err) || messaging.IsInvalidArgument(err) {
		return true
	}
	errStr := err.Error()
	return strings.Contains(errStr, "registration-token-not-registered") ||
		strings.Contains(errStr, "UNREGISTERED") ||
		strings.Contains(errStr, "invalid-argument") ||
		strings.Contains(errStr, "INVALID_ARGUMENT")
}

func (p *FCMWorkerPool) sendFCM(workerID int, payload FCMPayload) {
	startTime := time.Now()

	if p.fcmClient == nil {
		log.Printf("[FCM-Worker-%d] 📱 [FCM MOCK PUSH] ส่งแจ้งเตือนจำลองสำเร็จ (Mock Mode) -> หัวข้อ: %s, เนื้อหา: %s, จำนวนอุปกรณ์: %d",
			workerID, payload.Title, payload.Body, len(payload.Tokens))
		return
	}

	for _, token := range payload.Tokens {
		msg := &messaging.Message{
			Token: token,
			Notification: &messaging.Notification{
				Title: payload.Title,
				Body:  payload.Body,
			},
			Android: &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					ChannelID: "default",
					Sound:     "default",
					Priority:  messaging.PriorityHigh,
				},
			},
			APNS: &messaging.APNSConfig{
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{
						Sound: "default",
					},
				},
			},
			Data: payload.Data,
		}

		res, err := p.fcmClient.Send(p.ctx, msg)
		if err != nil {
			log.Printf("[FCM-Worker-%d] ❌ [LIVE-ERROR] ส่งล้มเหลว -> Token: [%s...], Error: %v",
				workerID, safeTokenPrefix(token), err)

			if isDeadToken(err) {
				// ส่งเข้าคิว Dead Token Purger แบบ non-blocking
				select {
				case p.deadTokenCh <- token:
				default:
				}
			} else if payload.RetryCount > 0 {
				// Retry เฉพาะ Transient Error
				retryPayload := payload
				retryPayload.Tokens = []string{token}
				retryPayload.RetryCount--
				go func() {
					time.Sleep(2 * time.Second)
					p.Submit(retryPayload)
				}()
			}
		} else {
			log.Printf("[FCM-Worker-%d] ✅ [LIVE-SUCCESS] ส่งผลสำเร็จ -> MessageID: %s, Token: [%s...], หัวข้อ: %s, เวลาทำงาน: %v",
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
	close(p.deadTokenCh)
	p.wg.Wait()
	log.Println("[FCM-WorkerPool] หยุดการทำงานอย่างสมบูรณ์ ปราศจาก Goroutine leak")
}
