package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"super-app-chonburi-go/config"
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/internal/repository"
	"super-app-chonburi-go/internal/usecase"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	os.Setenv("DB_DSN", "postgres://uat_admin:C807BI%2B%2BwFizSBH%2FFkNgL7J3qW646npW@122.155.169.235:5489/uat_chonburi?sslmode=disable")
	cfg := config.LoadConfig()
	db, err := gorm.Open(postgres.Open(cfg.DBDsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	repo := repository.NewTaxNewRepository(db)
	// We pass a mock email sender that does nothing
	emailSender := &MockEmailSender{}
	uc := usecase.NewTaxNewUseCase(repo, emailSender, "099400016485800")

	// 1. Create a dummy declaration in pending status for walk-in test
	businessID := uuid.MustParse("959608f3-d198-4f13-809a-68dfa8b0b76f") // ร้านค้าส่งยาสูบชลบุรี
	declID := uuid.New()
	decl := &domain.TaxDeclaration{
		ID:                 declID,
		BusinessID:         businessID,
		BusinessRegNumber:  "9999999",
		TaxType:            "tobacco_tax",
		TaxMonth:           6, // June
		TaxYear:            2026,
		DeclarationVersion: 1,
		MonthlyRevenue:     30000,
		CalculatedTax:      300, // 1% of 30,000
		FormFileURL:        "http://localhost:8082/uploads/test.png",
		PayerEmail:         "phnjk2000@gmail.com",
		Ref1:               "999999903",
		Ref2:               "20260601",
		PaymentStatus:      "pending",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Clean up if already exists
	db.Exec("DELETE FROM tax_declarations WHERE tax_month = 6 AND tax_year = 2026 AND business_reg_number = '9999999'")

	if err := db.Create(decl).Error; err != nil {
		log.Fatalf("failed to create dummy declaration: %v", err)
	}
	fmt.Println("Created pending declaration for walk-in test.")

	// 2. Perform walk-in payment approval (admin changes status to 'verified')
	adminID := uuid.MustParse("a91e1b6a-5034-413a-b5ec-96197eb858a9")
	notes := "รับเงินสดชำระหน้าเคาน์เตอร์เรียบร้อย ออกใบเสร็จ e-LAAS RCPT-99999/69"
	err = uc.UpdateAuditStatus(declID, "verified", notes, adminID)
	if err != nil {
		log.Fatalf("failed to update audit status: %v", err)
	}
	fmt.Println("Walk-in payment approved successfully.")

	// 3. Verify that PaidAt, PaidAmount, AuditedBy, and AuditNotes are set in database
	var updatedDecl domain.TaxDeclaration
	if err := db.First(&updatedDecl, "id = ?", declID).Error; err != nil {
		log.Fatalf("failed to fetch updated declaration: %v", err)
	}

	fmt.Printf("\n--- Verified Declaration Status ---\n")
	fmt.Printf("PaymentStatus: %s\n", updatedDecl.PaymentStatus)
	fmt.Printf("PaidAmount: %v\n", *updatedDecl.PaidAmount)
	fmt.Printf("PaidAt: %v\n", *updatedDecl.PaidAt)
	fmt.Printf("AuditedBy: %v\n", *updatedDecl.AuditedBy)
	fmt.Printf("AuditNotes: %q\n", *updatedDecl.AuditNotes)

	// 4. Test Dashboard calculation
	startDate := time.Now().AddDate(0, 0, -1)
	endDate := time.Now().AddDate(0, 0, 1)
	summary, err := repo.GetDashboardSummary(startDate, endDate)
	if err != nil {
		log.Fatalf("failed to get dashboard summary: %v", err)
	}

	summaryJSON, _ := json.MarshalIndent(summary, "", "  ")
	fmt.Printf("\n--- Dashboard Summary (Including Walk-In verified declaration) ---\n")
	fmt.Println(string(summaryJSON))
}

type MockEmailSender struct{}

func (m *MockEmailSender) SendHTML(to []string, subject, body string) error {
	fmt.Printf("[MockEmailSender] Email to %v with subject %q sent successfully.\n", to, subject)
	return nil
}
