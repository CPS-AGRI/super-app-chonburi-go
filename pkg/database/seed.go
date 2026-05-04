package database

import (
	"log"
	"time"

	"super-app-chonburi-go/internal/domain"
	"github.com/google/uuid"
)

func Seed() {
	if DB == nil {
		return
	}

	log.Println("Seeding Database...")

	// 1. Delete old test permissions
	DB.Exec("DELETE FROM system_permissions WHERE module = 'COMPLAINT_MODULE' OR id = 'complaint' OR module = 'complaint'")

	// 2. Insert new permissions matching user's exact specification
	permissions := []domain.SystemPermission{
		{
			ID:          "COMP_ACCIDENT",
			Module:      "COMPLAINT_MODULE",
			NameTh:      "แจ้งเหตุอุบัติเหตุทางถนน",
			Description: "แจ้งเหตุอุบัติเหตุทางถนน",
			CreatedBy:   "seed",
			UpdatedBy:   "seed",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "COMP_AIDS_PATIENT",
			Module:      "COMPLAINT_MODULE",
			NameTh:      "ผู้ป่วยเอดส์",
			Description: "ผู้ป่วยเอดส์",
			CreatedBy:   "seed",
			UpdatedBy:   "seed",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "COMP_CAR_BREAKDOWN",
			Module:      "COMPLAINT_MODULE",
			NameTh:      "แจ้งเหตุ / ขอความช่วยเหลือรถเสีย",
			Description: "แจ้งเหตุ / ขอความช่วยเหลือรถเสีย",
			CreatedBy:   "seed",
			UpdatedBy:   "seed",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "COMP_CULTURE",
			Module:      "COMPLAINT_MODULE",
			NameTh:      "งานประเพณี วัฒนธรรม ศาสนา",
			Description: "งานประเพณี วัฒนธรรม ศาสนา",
			CreatedBy:   "seed",
			UpdatedBy:   "seed",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "COMP_DISABLED_ALLOWANCE",
			Module:      "COMPLAINT_MODULE",
			NameTh:      "เบี้ยยังชีพผู้พิการ",
			Description: "เบี้ยยังชีพผู้พิการ",
			CreatedBy:   "seed",
			UpdatedBy:   "seed",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, p := range permissions {
		if err := DB.Create(&p).Error; err != nil {
			log.Printf("Warning: Failed to create permission %s: %v\n", p.ID, err)
		} else {
			log.Printf("Created system permission %s\n", p.ID)
		}
	}

	// 3. Clear existing complaints to have a clean mock
	DB.Exec("DELETE FROM module_complaints")
	DB.Exec("DELETE FROM module_complaint_user_informations")
	DB.Exec("DELETE FROM module_complaint_images")

	// 4. Insert 4 diverse mockup complaints to test every state perfectly
	complaints := []domain.Complaint{
		{
			ID:            uuid.New(),
			DocumentID:    "CMP-20260504-001",
			PermissionID:  "COMP_ACCIDENT",
			Latitude:      "13.3611",
			Longitude:     "100.9847",
			GoogleMapsUrl: "https://www.google.com/maps?q=13.3611,100.9847",
			Description:   "มีขยะสะสมส่งกลิ่นเหม็นบริเวณซอย 4 ช่วยมาเก็บด้วยครับ",
			Status:        domain.ComplaintStatusReceived,
			CreatedBy:     "citizen",
			UpdatedBy:     "citizen",
			CreatedAt:     time.Now().Add(-24 * time.Hour),
			UpdatedAt:     time.Now().Add(-24 * time.Hour),
			UserInformation: &domain.ComplaintUserInformation{
				ID:             uuid.New(),
				Prefix:         "นาย",
				Name:           "สมชาย",
				LastName:       "ใจดี",
				Phone:          "0812345678",
				IdentityNumber: "1100223344556",
			},
			Images: []domain.ComplaintImage{
				{
					ID:       uuid.New(),
					URL:      "https://images.unsplash.com/photo-1532996122724-e3c354a0b15b?q=80&w=600",
					Sequence: 1,
				},
			},
		},
		{
			ID:            uuid.New(),
			DocumentID:    "CMP-20260504-003",
			PermissionID:  "COMP_CAR_BREAKDOWN",
			Latitude:      "13.3620",
			Longitude:     "100.9855",
			GoogleMapsUrl: "https://www.google.com/maps?q=13.3620,100.9855",
			Description:   "ท่อระบายน้ำอุดตัน น้ำขังรอการระบาย",
			Status:        domain.ComplaintStatusCompleted,
			CreatedBy:     "citizen",
			UpdatedBy:     "citizen",
			CreatedAt:     time.Now().Add(-72 * time.Hour),
			UpdatedAt:     time.Now().Add(-2 * time.Hour),
			UserInformation: &domain.ComplaintUserInformation{
				ID:             uuid.New(),
				Prefix:         "นาย",
				Name:           "สมศักดิ์",
				LastName:       "คงดี",
				Phone:          "0855554444",
				IdentityNumber: "3210987654321",
			},
			Images: []domain.ComplaintImage{
				{
					ID:       uuid.New(),
					URL:      "https://images.unsplash.com/photo-1581093458791-9f3c3250bbd1?q=80&w=600",
					Sequence: 1,
				},
			},
		},
		{
			ID:            uuid.New(),
			DocumentID:    "CMP-20260504-005",
			PermissionID:  "COMP_DISABLED_ALLOWANCE",
			Latitude:      "13.3630",
			Longitude:     "100.9865",
			GoogleMapsUrl: "https://www.google.com/maps?q=13.3630,100.9865",
			Description:   "มีปัญหาเรื่องทางลาดของผู้พิการในเขตชุมชน",
			Status:        domain.ComplaintStatusReceived,
			CreatedBy:     "citizen",
			UpdatedBy:     "citizen",
			CreatedAt:     time.Now().Add(-2 * time.Hour),
			UpdatedAt:     time.Now().Add(-2 * time.Hour),
			UserInformation: &domain.ComplaintUserInformation{
				ID:             uuid.New(),
				Prefix:         "นาย",
				Name:           "สมพงษ์",
				LastName:       "คงเจริญ",
				Phone:          "0844445555",
				IdentityNumber: "3322114455667",
			},
			Images: []domain.ComplaintImage{
				{
					ID:       uuid.New(),
					URL:      "https://images.unsplash.com/photo-1590086782957-93c06ef21604?q=80&w=600",
					Sequence: 1,
				},
			},
		},
		{
			ID:            uuid.New(),
			DocumentID:    "CMP-20260504-006",
			PermissionID:  "COMP_ACCIDENT",
			Latitude:      "13.3635",
			Longitude:     "100.9870",
			GoogleMapsUrl: "https://www.google.com/maps?q=13.3635,100.9870",
			Description:   "อุบัติเหตุเฉี่ยวชนบริเวณสี่แยกไฟแดง",
			Status:        domain.ComplaintStatusReceived,
			CreatedBy:     "citizen",
			UpdatedBy:     "citizen",
			CreatedAt:     time.Now().Add(-5 * time.Hour),
			UpdatedAt:     time.Now().Add(-5 * time.Hour),
			UserInformation: &domain.ComplaintUserInformation{
				ID:             uuid.New(),
				Prefix:         "นาง",
				Name:           "สมฤดี",
				LastName:       "มีทรัพย์",
				Phone:          "0811223344",
				IdentityNumber: "4455663322119",
			},
			Images: []domain.ComplaintImage{
				{
					ID:       uuid.New(),
					URL:      "https://images.unsplash.com/photo-1532996122724-e3c354a0b15b?q=80&w=600",
					Sequence: 1,
				},
			},
		},
		{
			ID:            uuid.New(),
			DocumentID:    "CMP-20260504-007",
			PermissionID:  "COMP_CAR_BREAKDOWN",
			Latitude:      "13.3640",
			Longitude:     "100.9875",
			GoogleMapsUrl: "https://www.google.com/maps?q=13.3640,100.9875",
			Description:   "รถกระบะยางแตกขวางทางเข้าหมู่บ้าน",
			Status:        domain.ComplaintStatusReceived,
			CreatedBy:     "citizen",
			UpdatedBy:     "citizen",
			CreatedAt:     time.Now().Add(-8 * time.Hour),
			UpdatedAt:     time.Now().Add(-8 * time.Hour),
			UserInformation: &domain.ComplaintUserInformation{
				ID:             uuid.New(),
				Prefix:         "นาย",
				Name:           "สมรักษ์",
				LastName:       "ขวัญใจ",
				Phone:          "0899998888",
				IdentityNumber: "5566778899112",
			},
			Images: []domain.ComplaintImage{
				{
					ID:       uuid.New(),
					URL:      "https://images.unsplash.com/photo-1517486808906-6ca8b3f04846?q=80&w=600",
					Sequence: 1,
				},
			},
		},
	}

	for i := range complaints {
		if complaints[i].UserInformation != nil {
			complaints[i].UserInformation.ComplaintID = complaints[i].ID
		}

		var existing domain.Complaint
		if err := DB.Where("document_id = ?", complaints[i].DocumentID).First(&existing).Error; err == nil {
			if complaints[i].UserInformation != nil {
				complaints[i].UserInformation.ComplaintID = existing.ID
				var existingUser domain.ComplaintUserInformation
				if err := DB.Where("complaint_id = ?", existing.ID).First(&existingUser).Error; err != nil {
					DB.Create(complaints[i].UserInformation)
					log.Printf("Added missing user info for existing complaint: %s\n", complaints[i].DocumentID)
				}
			}
		} else {
			if err := DB.Create(&complaints[i]).Error; err != nil {
				log.Printf("Warning: Failed to seed complaint %s: %v\n", complaints[i].DocumentID, err)
			} else {
				log.Printf("Seeded mock complaint: %s (Status: %s)\n", complaints[i].DocumentID, complaints[i].Status)
			}
		}
	}

	log.Println("✅ 5 Mockup permissions and 4 complaints seeded successfully")
}
