package pdf

import (
	"fmt"
	"os"
	"testing"
	"time"

	"super-app-chonburi-go/internal/domain"

	"github.com/google/uuid"
)

func TestGenerateReceiptPDF(t *testing.T) {
	err := os.MkdirAll("../../scratch", 0755)
	if err != nil {
		t.Fatalf("failed to create scratch directory: %v", err)
	}

	generator := NewReceiptGenerator()

	declHotel := &domain.TaxDeclaration{
		ID:                 uuid.MustParse("065f18d2-5500-415c-bf3c-5a67f72c5a92"),
		BusinessRegNumber:  "9999999",
		TaxType:            "hotel_fee",
		TaxMonth:           2,
		TaxYear:            2026,
		DeclarationVersion: 1,
		MonthlyRevenue:     70000,
		CalculatedTax:      350.00,
		FormFileURL:        "test.png",
		PayerEmail:         "phnjk2000@gmail.com",
		Ref1:               "123456701",
		Ref2:               "20260501",
		PaymentStatus:      "verified",
		PaidAmount:         floatPtr(350.00),
		PaidAt:             timePtr(time.Date(2026, 3, 6, 10, 0, 0, 0, time.UTC)),
		Business: &domain.TaxBusiness{
			NameTH:    "ธันย่า ชีวิว",
			OwnerName: stringPtr("ธันย่า ชีวิว"),
		},
	}

	pdfBytes, err := generator.GenerateReceiptPDF(declHotel)
	if err != nil {
		t.Fatalf("failed to generate hotel fee pdf: %v", err)
	}

	err = os.WriteFile("../../scratch/test_hotel_receipt.pdf", pdfBytes, 0644)
	if err != nil {
		t.Fatalf("failed to save hotel fee pdf: %v", err)
	}
	fmt.Println("SUCCESS: Saved scratch/test_hotel_receipt.pdf")

	declTobacco := &domain.TaxDeclaration{
		ID:                 uuid.MustParse("1a43627f-15ee-46c9-96a8-1472ff503c3a"),
		BusinessRegNumber:  "9999999",
		TaxType:            "tobacco_tax",
		TaxMonth:           1,
		TaxYear:            2026,
		DeclarationVersion: 1,
		MonthlyRevenue:     256077546,
		CalculatedTax:      2560775.46,
		FormFileURL:        "test.png",
		PayerEmail:         "phnjk2000@gmail.com",
		Ref1:               "999999903",
		Ref2:               "20260101",
		PaymentStatus:      "verified",
		PaidAmount:         floatPtr(2560775.46),
		PaidAt:             timePtr(time.Date(2026, 2, 19, 10, 0, 0, 0, time.UTC)),
		AuditNotes:         stringPtr("เช็คธนาคารดอยซ์แบงก์ สาขาสำนักงานใหญ่ เลขที่ 08329550 ลงวันที่ 17 กุมภาพันธ์ 2569"),
		Business: &domain.TaxBusiness{
			NameTH:    "บริษัท ซีพี ออลล์ จำกัด (มหาชน)",
			OwnerName: stringPtr("บริษัท ฟิลลิป มอร์ริส เทรดดิ้ง (ไทยแลนด์) จำกัด"),
		},
	}

	pdfBytes2, err := generator.GenerateReceiptPDF(declTobacco)
	if err != nil {
		t.Fatalf("failed to generate tobacco pdf: %v", err)
	}

	err = os.WriteFile("../../scratch/test_tobacco_receipt.pdf", pdfBytes2, 0644)
	if err != nil {
		t.Fatalf("failed to save tobacco pdf: %v", err)
	}
	fmt.Println("SUCCESS: Saved scratch/test_tobacco_receipt.pdf")

	declOilGas := &domain.TaxDeclaration{
		ID:                 uuid.MustParse("d80ff22b-eda4-4d30-8df9-7cbb4f0ee9cb"),
		BusinessRegNumber:  "765432102",
		TaxType:            "oil_gas_tax",
		TaxMonth:           2,
		TaxYear:            2026,
		DeclarationVersion: 1,
		MonthlyRevenue:     102150000,
		CalculatedTax:      4086.00,
		FormFileURL:        "test.png",
		PayerEmail:         "phnjk2000@gmail.com",
		Ref1:               "765432102",
		Ref2:               "20260201",
		PaymentStatus:      "verified",
		PaidAmount:         floatPtr(4086.00),
		PaidAt:             timePtr(time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)),
		Business: &domain.TaxBusiness{
			NameTH:    "บริษัท ปตท. น้ำมันและการค้าปลีก จำกัด (มหาชน)",
			OwnerName: stringPtr("บริษัท ปตท. น้ำมันและการค้าปลีก จำกัด (มหาชน)"),
		},
	}

	pdfBytes3, err := generator.GenerateReceiptPDF(declOilGas)
	if err != nil {
		t.Fatalf("failed to generate oil gas pdf: %v", err)
	}

	err = os.WriteFile("../../scratch/test_oil_gas_receipt.pdf", pdfBytes3, 0644)
	if err != nil {
		t.Fatalf("failed to save oil gas pdf: %v", err)
	}
	fmt.Println("SUCCESS: Saved scratch/test_oil_gas_receipt.pdf")
}

func floatPtr(v float64) *float64    { return &v }
func stringPtr(v string) *string     { return &v }
func timePtr(v time.Time) *time.Time { return &v }
