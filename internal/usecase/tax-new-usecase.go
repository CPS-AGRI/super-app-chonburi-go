package usecase

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/mail"
	"super-app-chonburi-go/pkg/qr"

	"github.com/google/uuid"
)

type taxNewUseCase struct {
	repo       domain.TaxNewRepository
	mailSender mail.EmailSender
	billerID   string
}

// NewTaxNewUseCase creates a new usecase instance for the self-declaration tax module
func NewTaxNewUseCase(repo domain.TaxNewRepository, mailSender mail.EmailSender, billerID string) domain.TaxNewUseCase {
	if billerID == "" {
		billerID = "099400016485800" // Default Chonburi PAO Biller ID
	}
	return &taxNewUseCase{
		repo:       repo,
		mailSender: mailSender,
		billerID:   billerID,
	}
}

func (u *taxNewUseCase) GetBusiness(regNumber string) (*domain.TaxBusinessDTO, error) {
	business, err := u.repo.GetBusinessByRegNumber(regNumber)
	if err != nil {
		return nil, err
	}
	if business == nil {
		return nil, errors.New("business not found")
	}

	rate, err := u.repo.GetActiveTaxRate(business.TaxType)
	if err != nil {
		return nil, err
	}
	rateValue := 0.0
	rateUnit := "percentage"
	if rate != nil {
		rateValue = rate.RateValue
		rateUnit = rate.RateUnit
	}

	return &domain.TaxBusinessDTO{
		BusinessRegNumber: business.BusinessRegNumber,
		NameTH:            business.NameTH,
		TaxType:           business.TaxType,
		TaxRate:           rateValue,
		RateUnit:          rateUnit,
	}, nil
}

func (u *taxNewUseCase) DeclareTax(req domain.DeclareTaxRequest) (*domain.DeclareTaxResponse, error) {
	business, err := u.repo.GetBusinessByRegNumber(req.BusinessRegNumber)
	if err != nil {
		return nil, err
	}
	if business == nil {
		return nil, errors.New("business not found")
	}

	rate, err := u.repo.GetActiveTaxRate(business.TaxType)
	if err != nil {
		return nil, err
	}
	if rate == nil {
		return nil, errors.New("active tax rate not found for tax type " + business.TaxType)
	}

	// 1. Calculate Tax
	calculatedTax := req.MonthlyRevenue * rate.RateValue
	if rate.RateUnit == "percentage" {
		calculatedTax = req.MonthlyRevenue * (rate.RateValue / 100.0)
	}

	// 2. Determine Declaration Version
	version, err := u.repo.GetLatestDeclarationVersion(req.BusinessRegNumber, business.TaxType, req.TaxMonth, req.TaxYear)
	if err != nil {
		return nil, err
	}
	newVersion := version + 1

	// 3. Generate Ref1 & Ref2
	// Ref1: RegNumber (7 digits) + TaxTypeCode (2 digits)
	var typeCode string
	switch business.TaxType {
	case "hotel_fee":
		typeCode = "01"
	case "oil_gas_tax":
		typeCode = "02"
	case "tobacco_tax":
		typeCode = "03"
	default:
		typeCode = "00"
	}
	ref1 := fmt.Sprintf("%s%s", business.BusinessRegNumber, typeCode)

	// Ref2: Year (YYYY) + Month (MM) + Version (2 digits)
	ref2 := fmt.Sprintf("%04d%02d%02d", req.TaxYear, req.TaxMonth, newVersion)

	// 4. Generate PromptPay QR Content
	qrContent, err := qr.GeneratePromptPayBillPayment(u.billerID, ref1, ref2, calculatedTax)
	if err != nil {
		return nil, fmt.Errorf("failed to generate promptpay QR: %w", err)
	}

	// 5. Create Tax Declaration Record
	declaration := &domain.TaxDeclaration{
		ID:                 uuid.New(),
		BusinessID:         business.ID,
		BusinessRegNumber:  business.BusinessRegNumber,
		TaxType:            business.TaxType,
		TaxMonth:           req.TaxMonth,
		TaxYear:            req.TaxYear,
		DeclarationVersion: newVersion,
		MonthlyRevenue:     req.MonthlyRevenue,
		VolumeUnits:        req.VolumeUnits,
		CalculatedTax:      calculatedTax,
		FormFileURL:        req.FormFileURL,
		PayerEmail:         req.PayerEmail,
		PayerPhone:         &req.PayerPhone,
		Ref1:               ref1,
		Ref2:               ref2,
		QRCodeContent:      &qrContent,
		PaymentStatus:      "pending",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err = u.repo.CreateDeclaration(declaration)
	if err != nil {
		return nil, err
	}

	return &domain.DeclareTaxResponse{
		DeclarationID: declaration.ID,
		CalculatedTax: declaration.CalculatedTax,
		Ref1:          declaration.Ref1,
		Ref2:          declaration.Ref2,
		QRCodeContent: *declaration.QRCodeContent,
		PaymentStatus: declaration.PaymentStatus,
	}, nil
}

func (u *taxNewUseCase) GetDeclaration(id uuid.UUID) (*domain.TaxDeclaration, error) {
	return u.repo.GetDeclarationByID(id)
}

func (u *taxNewUseCase) UploadKTBFile(filename string, fileContent []byte, adminID uuid.UUID) (*domain.KTBReconciliationResponse, error) {
	// Initialize Batch Record (will write to database later after calculating totals)
	batch := &domain.BankReconciliationBatch{
		ID:           uuid.New(),
		Filename:     filename,
		UploadedBy:   adminID,
		RecordCount:  0,
		MatchedCount: 0,
		TotalAmount:  0.0,
		CreatedAt:    time.Now(),
	}

	// Parse Fixed-Width file content line by line
	lines := strings.Split(string(fileContent), "\n")
	var records []*domain.BankReconciliationRecord

	for _, line := range lines {
		// Clean line endings
		line = strings.TrimRight(line, "\r\n")
		if len(line) < 120 {
			continue // Skip invalid length lines (standard layout requires 120 bytes)
		}

		recordType := line[0:2]
		if recordType != "02" {
			continue // Skip Header (01) or Total (09) records
		}

		batch.RecordCount++

		// Extract fields based on specs
		payDateStr := line[2:10] // YYYYMMDD
		ref1 := strings.TrimSpace(line[16:36])
		ref2 := strings.TrimSpace(line[36:56])
		amountStr := line[56:69] // 13 digits, last 2 are decimal places

		// Parse Payment Date
		payDate, err := time.Parse("20060102", payDateStr)
		if err != nil {
			payDate = time.Now()
		}

		// Parse Amount
		amountInt, err := strconv.ParseInt(amountStr, 10, 64)
		amount := float64(amountInt) / 100.0
		if err != nil {
			amount = 0.0
		}

		batch.TotalAmount += amount

		record := &domain.BankReconciliationRecord{
			ID:          uuid.New(),
			BatchID:     batch.ID,
			Ref1:        ref1,
			Ref2:        ref2,
			Amount:      amount,
			PaymentDate: payDate,
			RawLine:     line,
			IsMatched:   false,
			CreatedAt:   time.Now(),
		}

		records = append(records, record)
	}

	// Identify matches to get the final MatchedCount before writing the batch
	for _, record := range records {
		decl, err := u.repo.GetDeclarationByRefs(record.Ref1, record.Ref2)
		if err == nil && decl != nil {
			if decl.CalculatedTax == record.Amount {
				record.IsMatched = true
				batch.MatchedCount++
			}
		}
	}

	// Save Batch Summary with correct final counts and totals
	err := u.repo.CreateReconciliationBatch(batch)
	if err != nil {
		return nil, err
	}

	// Save Reconciliation Records and Update Declarations
	for _, record := range records {
		err := u.repo.CreateReconciliationRecord(record)
		if err != nil {
			continue
		}

		if record.IsMatched {
			decl, err := u.repo.GetDeclarationByRefs(record.Ref1, record.Ref2)
			if err == nil && decl != nil {
				decl.PaymentStatus = "paid"
				decl.PaidAmount = &record.Amount
				decl.PaidAt = &record.PaymentDate
				decl.KTBReconciliationRecordID = &record.ID
				decl.UpdatedAt = time.Now()

				err = u.repo.UpdateDeclaration(decl)
				if err == nil {
					// Send Payment Email Confirmation
					u.sendPaymentSuccessEmail(decl)
				}
			}
		}
	}

	return &domain.KTBReconciliationResponse{
		BatchID:          batch.ID,
		Filename:         batch.Filename,
		TotalRecords:     batch.RecordCount,
		MatchedRecords:   batch.MatchedCount,
		UnmatchedRecords: batch.RecordCount - batch.MatchedCount,
		TotalAmount:      batch.TotalAmount,
	}, nil
}

func (u *taxNewUseCase) sendPaymentSuccessEmail(decl *domain.TaxDeclaration) {
	subject := fmt.Sprintf("ยืนยันการชำระเงินภาษี/ค่าธรรมเนียม อบจ. ชลบุรี - %s", decl.Business.NameTH)
	
	thaiMonth := getThaiMonthName(decl.TaxMonth)
	thaiYear := decl.TaxYear + 543 // Buddhist Era

	body := fmt.Sprintf(`
	<html>
	<body style="font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; color: #333; line-height: 1.6;">
		<div style="max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #e0e0e0; border-radius: 8px;">
			<div style="text-align: center; margin-bottom: 20px;">
				<h2 style="color: #1a73e8; margin-top: 10px;">ใบเสร็จรับเงิน / ยืนยันการชำระเงิน</h2>
				<p style="color: #666; font-size: 14px;">องค์การบริหารส่วนจังหวัดชลบุรี</p>
			</div>
			<hr style="border: 0; border-top: 1px solid #eee; margin: 20px 0;">
			<p>เรียน ผู้เสียภาษี,</p>
			<p>ระบบงานภาษีและค่าธรรมเนียม อบจ. ชลบุรี ได้รับเงินโอนค่าภาษี/ค่าธรรมเนียมของท่านเรียบร้อยแล้ว รายละเอียดดังนี้:</p>
			<table style="width: 100%%; border-collapse: collapse; margin: 20px 0;">
				<tr>
					<td style="padding: 8px 0; font-weight: bold; width: 180px;">รหัสสถานประกอบการ:</td>
					<td style="padding: 8px 0;">%s</td>
				</tr>
				<tr>
					<td style="padding: 8px 0; font-weight: bold;">ชื่อสถานประกอบการ:</td>
					<td style="padding: 8px 0;">%s</td>
				</tr>
				<tr>
					<td style="padding: 8px 0; font-weight: bold;">ประเภทภาษี/ค่าธรรมเนียม:</td>
					<td style="padding: 8px 0;">%s</td>
				</tr>
				<tr>
					<td style="padding: 8px 0; font-weight: bold;">รอบภาษีประจำเดือน:</td>
					<td style="padding: 8px 0;">%s %d (เวอร์ชันการยื่นที่ %d)</td>
				</tr>
				<tr>
					<td style="padding: 8px 0; font-weight: bold;">ยอดเงินที่ชำระ:</td>
					<td style="padding: 8px 0; font-size: 18px; color: #2e7d32; font-weight: bold;">%s บาท</td>
				</tr>
				<tr>
					<td style="padding: 8px 0; font-weight: bold;">วันที่ชำระเงิน:</td>
					<td style="padding: 8px 0;">%s</td>
				</tr>
				<tr>
					<td style="padding: 8px 0; font-weight: bold;">เลขอ้างอิง Ref 1 / Ref 2:</td>
					<td style="padding: 8px 0; font-family: monospace;">%s / %s</td>
				</tr>
			</table>
			<div style="background-color: #f1f8e9; border-left: 4px solid #8bc34a; padding: 15px; border-radius: 4px; margin-top: 20px;">
				<p style="margin: 0; font-size: 14px; color: #33691e;"><strong>หมายเหตุ:</strong> เอกสารใบเสร็จอย่างเป็นทางการของ อบจ. ชลบุรี จะถูกส่งมอบให้ท่านทางไปรษณีย์หรือช่องทางที่ท่านลงทะเบียนไว้ต่อไป</p>
			</div>
			<hr style="border: 0; border-top: 1px solid #eee; margin: 20px 0;">
			<p style="font-size: 12px; color: #999; text-align: center;">นี่เป็นอีเมลอัตโนมัติ กรุณาอย่าตอบกลับอีเมลฉบับนี้</p>
		</div>
	</body>
	</html>`, 
		decl.BusinessRegNumber, 
		decl.Business.NameTH, 
		getTaxTypeNameTH(decl.TaxType), 
		thaiMonth, 
		thaiYear, 
		decl.DeclarationVersion,
		formatWithCommas(decl.CalculatedTax), 
		decl.PaidAt.Format("02/01/2006 15:04:05"), 
		decl.Ref1, 
		decl.Ref2,
	)

	if err := u.mailSender.SendHTML([]string{decl.PayerEmail}, subject, body); err != nil {
		log.Printf("ERROR: Failed to send tax payment success email to %s: %v", decl.PayerEmail, err)
	}
}

func (u *taxNewUseCase) UploadElaasFile(filename string, fileContent []byte, adminID uuid.UUID) (int, error) {
	reader := csv.NewReader(bytes.NewReader(fileContent))
	// Read header row
	_, err := reader.Read()
	if err != nil {
		return 0, err
	}

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, err
		}

		if len(record) < 4 {
			continue
		}

		dateStr := record[0]
		taxType := record[1]
		amountStr := record[2]
		transCountStr := record[3]

		summaryDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			continue
		}

		transCount, err := strconv.Atoi(transCountStr)
		if err != nil {
			continue
		}

		summary := &domain.ElaasDailySummary{
			ID:               uuid.New(),
			SummaryDate:      summaryDate,
			TaxType:          taxType,
			TotalAmount:      amount,
			TransactionCount: transCount,
			Filename:         filename,
			UploadedBy:       adminID,
			CreatedAt:        time.Now(),
		}

		err = u.repo.UpsertElaasSummary(summary)
		if err == nil {
			count++
		}
	}

	return count, nil
}

func (u *taxNewUseCase) GetDashboard(startDateStr, endDateStr string) (*domain.DashboardSummaryResponse, error) {
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		// Default to start of current month
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		endDate = time.Now()
	}

	return u.repo.GetDashboardSummary(startDate, endDate)
}

func (u *taxNewUseCase) ImportBusinesses(fileContent []byte) (*domain.ImportBusinessesResponse, error) {
	reader := csv.NewReader(bytes.NewReader(fileContent))
	
	// Read header
	_, err := reader.Read()
	if err != nil {
		return nil, err
	}

	resp := &domain.ImportBusinessesResponse{}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			resp.Failed++
			continue
		}

		if len(record) < 4 {
			resp.Failed++
			continue
		}

		regNum := record[0]
		nameTH := record[1]
		nameEN := record[2]
		taxType := record[3]

		var ownerName, ownerID, contactPhone, contactEmail, address string
		if len(record) > 4 { ownerName = record[4] }
		if len(record) > 5 { ownerID = record[5] }
		if len(record) > 6 { contactPhone = record[6] }
		if len(record) > 7 { contactEmail = record[7] }
		if len(record) > 8 { address = record[8] }

		business := &domain.TaxBusiness{
			ID:                  uuid.New(),
			BusinessRegNumber:   regNum,
			NameTH:              nameTH,
			TaxType:             taxType,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		if nameEN != "" { business.NameEN = &nameEN }
		if ownerName != "" { business.OwnerName = &ownerName }
		if ownerID != "" { business.OwnerIdentityNumber = &ownerID }
		if contactPhone != "" { business.ContactPhone = &contactPhone }
		if contactEmail != "" { business.ContactEmail = &contactEmail }
		if address != "" { business.AddressDetail = &address }

		// Check if exists
		existing, err := u.repo.GetBusinessByRegNumber(regNum)
		if err == nil && existing != nil {
			business.ID = existing.ID
			business.CreatedAt = existing.CreatedAt
			err = u.repo.UpsertBusiness(business)
			if err == nil {
				resp.Updated++
			} else {
				resp.Failed++
			}
		} else {
			err = u.repo.UpsertBusiness(business)
			if err == nil {
				resp.Inserted++
			} else {
				resp.Failed++
			}
		}
	}

	return resp, nil
}

func (u *taxNewUseCase) UpdateAuditStatus(id uuid.UUID, status string, notes string, adminID uuid.UUID) error {
	decl, err := u.repo.GetDeclarationByID(id)
	if err != nil {
		return err
	}
	if decl == nil {
		return errors.New("declaration not found")
	}

	// Walk-in handling: if the declaration was pending and is now verified, it means a walk-in payment was made.
	if decl.PaymentStatus == "pending" && status == "verified" {
		now := time.Now()
		decl.PaidAt = &now
		decl.PaidAmount = &decl.CalculatedTax
	}

	decl.PaymentStatus = status
	decl.AuditedBy = &adminID
	decl.AuditNotes = &notes
	decl.UpdatedAt = time.Now()
	
	err = u.repo.UpdateDeclaration(decl)
	if err != nil {
		return err
	}

	if status == "audit_failed" {
		u.sendAuditFailedEmail(decl, notes)
	}

	return nil
}

func (u *taxNewUseCase) sendAuditFailedEmail(decl *domain.TaxDeclaration, notes string) {
	subject := fmt.Sprintf("แจ้งผลการตรวจสอบการยื่นแบบภาษี/ค่าธรรมเนียม (ตรวจสอบไม่ผ่าน) - %s", decl.Business.NameTH)
	
	thaiMonth := getThaiMonthName(decl.TaxMonth)
	thaiYear := decl.TaxYear + 543 // Buddhist Era

	body := fmt.Sprintf(`
	<html>
	<body style="font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; color: #333; line-height: 1.6;">
		<div style="max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #e0e0e0; border-radius: 8px;">
			<div style="text-align: center; margin-bottom: 20px;">
				<h2 style="color: #d32f2f; margin-top: 10px;">แจ้งผลการตรวจสอบไม่ผ่าน</h2>
				<p style="color: #666; font-size: 14px;">องค์การบริหารส่วนจังหวัดชลบุรี</p>
			</div>
			<hr style="border: 0; border-top: 1px solid #eee; margin: 20px 0;">
			<p>เรียน ผู้เสียภาษี,</p>
			<p>เจ้าหน้าที่ได้ทำการตรวจสอบแบบแสดงยอดรายได้ที่ท่านยื่นผ่านแอปพลิเคชันสำหรับ:</p>
			<table style="width: 100%%; border-collapse: collapse; margin: 20px 0;">
				<tr>
					<td style="padding: 8px 0; font-weight: bold; width: 180px;">รหัสสถานประกอบการ:</td>
					<td style="padding: 8px 0;">%s</td>
				</tr>
				<tr>
					<td style="padding: 8px 0; font-weight: bold;">ชื่อสถานประกอบการ:</td>
					<td style="padding: 8px 0;">%s</td>
				</tr>
				<tr>
					<td style="padding: 8px 0; font-weight: bold;">ประเภทภาษี/ค่าธรรมเนียม:</td>
					<td style="padding: 8px 0;">%s</td>
				</tr>
				<tr>
					<td style="padding: 8px 0; font-weight: bold;">รอบภาษีประจำเดือน:</td>
					<td style="padding: 8px 0;">%s %d (เวอร์ชันการยื่นที่ %d)</td>
				</tr>
				<tr>
					<td style="padding: 8px 0; font-weight: bold; color: #d32f2f;">สาเหตุ/รายละเอียดที่ไม่ผ่าน:</td>
					<td style="padding: 8px 0; font-weight: bold; color: #d32f2f;">%s</td>
				</tr>
			</table>
			<div style="background-color: #ffebee; border-left: 4px solid #f44336; padding: 15px; border-radius: 4px; margin-top: 20px;">
				<p style="margin: 0; font-size: 14px; color: #c62828;"><strong>สิ่งที่ต้องดำเนินการ:</strong></p>
				<p style="margin: 5px 0 0 0; font-size: 14px; color: #c62828;">กรุณาทำรายการยื่นแบบแสดงรายได้เพิ่มเติม (Supplementary Declaration) ผ่านแอปพลิเคชันมือถือเพื่อยื่นแสดงยอดที่ถูกต้อง และชำระเงินส่วนต่างภาษีที่ค้างให้เรียบร้อย</p>
			</div>
			<hr style="border: 0; border-top: 1px solid #eee; margin: 20px 0;">
			<p style="font-size: 12px; color: #999; text-align: center;">นี่เป็นอีเมลอัตโนมัติ กรุณาอย่าตอบกลับอีเมลฉบับนี้</p>
		</div>
	</body>
	</html>`, 
		decl.BusinessRegNumber, 
		decl.Business.NameTH, 
		getTaxTypeNameTH(decl.TaxType), 
		thaiMonth, 
		thaiYear, 
		decl.DeclarationVersion,
		notes,
	)

	if err := u.mailSender.SendHTML([]string{decl.PayerEmail}, subject, body); err != nil {
		log.Printf("ERROR: Failed to send audit failed email to %s: %v", decl.PayerEmail, err)
	}
}

// Helpers
func getThaiMonthName(m int) string {
	months := []string{"", "มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน", "กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม"}
	if m >= 1 && m <= 12 {
		return months[m]
	}
	return ""
}

func getTaxTypeNameTH(t string) string {
	switch t {
	case "hotel_fee":
		return "ค่าธรรมเนียมบำรุง อบจ. จากผู้เข้าพักโรงแรม"
	case "oil_gas_tax":
		return "ภาษีบำรุง อบจ.จากการค้าน้ำมัน/ก๊าซ"
	case "tobacco_tax":
		return "ภาษีบำรุง อบจ.จากการค้ายาสูบ"
	default:
		return "ภาษี/ค่าธรรมเนียม"
	}
}

func formatWithCommas(val float64) string {
	parts := strings.Split(fmt.Sprintf("%.2f", val), ".")
	intPart := parts[0]
	decPart := parts[1]

	var result []string
	length := len(intPart)
	for i, char := range intPart {
		if i > 0 && (length-i)%3 == 0 && intPart[i-1] != '-' {
			result = append(result, ",")
		}
		result = append(result, string(char))
	}
	return strings.Join(result, "") + "." + decPart
}

func (u *taxNewUseCase) ListDeclarations(taxType, status, search string, limit, offset int) ([]domain.TaxDeclaration, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return u.repo.ListDeclarations(taxType, status, search, limit, offset)
}


