package http

import "time"

// CreateNewsPayload คือข้อมูลสำหรับสร้างข่าวพร้อมกับการแจ้งเตือนในคำขอเดียว
type CreateNewsPayload struct {
	Title         string    `json:"title"`
	DescriptionTh *string   `json:"description_th"`
	DescriptionEn *string   `json:"description_en"`
	Type          string    `json:"type"`
	Priority      string    `json:"priority"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	Status        string    `json:"status"`
	Images        []struct {
		Url string `json:"url"`
	} `json:"images"`
}

// CreateNotificationRequest คือ DTO สำหรับรับ input จาก HTTP สำหรับสร้างการแจ้งเตือน
// รองรับ 3 modes:
//  1. text-only: ไม่มี PublicRelationId และ CreateNews → ส่งข้อความเท่านั้น
//  2. create-news: มี CreateNews → สร้างข่าวใหม่พร้อมกัน (atomic transaction)
//  3. link-news: มี PublicRelationId → เชื่อมโยงกับข่าวที่มีอยู่แล้ว
type CreateNotificationRequest struct {
	Title            string             `json:"title"`
	Description      *string            `json:"description"`
	SendDate         *string            `json:"send_date"`
	Type             string             `json:"type"`
	Status           string             `json:"status"`
	PublicRelationId *string            `json:"public_relation_id"`
	CreateNews       *CreateNewsPayload `json:"create_news"`
}
