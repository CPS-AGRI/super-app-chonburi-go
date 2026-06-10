package firebase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type FCMMessage struct {
	RegistrationIDs []string        `json:"registration_ids"`
	Notification    FCMNotification `json:"notification"`
}

type FCMNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Sound string `json:"sound"`
}

func SendPushNotification(tokens []string, title string, body string) {
	if len(tokens) == 0 {
		log.Println("⚠️ FCM: No tokens provided for user, skipping push notification.")
		return
	}

	serverKey := os.Getenv("FCM_SERVER_KEY")
	if serverKey == "" {
		log.Printf("📱 [FCM MOCK PUSH] Sending to %d devices:\n   Tokens: %v\n   Title: %s\n   Body: %s\n", len(tokens), tokens, title, body)
		return
	}

	url := "https://fcm.googleapis.com/fcm/send"
	payload := FCMMessage{
		RegistrationIDs: tokens,
		Notification: FCMNotification{
			Title: title,
			Body:  body,
			Sound: "default",
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("❌ FCM: Failed to marshal payload: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Printf("❌ FCM: Failed to create request: %v\n", err)
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("key=%s", serverKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ FCM: Failed to send request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("✅ FCM: Push notification sent successfully to %d devices.\n", len(tokens))
	} else {
		log.Printf("❌ FCM: Server returned error status: %s\n", resp.Status)
	}
}
