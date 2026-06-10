package pdf

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"super-app-chonburi-go/internal/domain"

	"github.com/jung-kurt/gofpdf/v2"
)

type ReceiptGenerator interface {
	GenerateReceiptPDF(decl *domain.TaxDeclaration) ([]byte, error)
}

type receiptGenerator struct {
	logoPath    string
	fontDir     string
	regularFont string
	boldFont    string
}

func NewReceiptGenerator() ReceiptGenerator {
	fontDir := "assets/fonts"
	logoPath := "assets/images/logo.png"

	if _, err := os.Stat(fontDir); os.IsNotExist(err) {
		if _, errParent := os.Stat("../../assets/fonts"); errParent == nil {
			fontDir = "../../assets/fonts"
			logoPath = "../../assets/images/logo.png"
		}
	}

	return &receiptGenerator{
		logoPath:    logoPath,
		fontDir:     fontDir,
		regularFont: "thsarabunnew.ttf",
		boldFont:    "thsarabunnewbd.ttf",
	}
}

func (g *receiptGenerator) GenerateReceiptPDF(decl *domain.TaxDeclaration) ([]byte, error) {

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 15, 20)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	pdf.AddUTF8Font("Sarabun", "", fmt.Sprintf("%s/%s", g.fontDir, g.regularFont))
	pdf.AddUTF8Font("Sarabun", "B", fmt.Sprintf("%s/%s", g.fontDir, g.boldFont))

	g.drawWatermarks(pdf)

	pdf.Image(g.logoPath, 88, 15, 34, 34, false, "PNG", 0, "")

	pdf.SetTextColor(50, 50, 50)

	pdf.SetFont("Sarabun", "B", 16)
	pdf.SetXY(20, 56)
	pdf.CellFormat(170, 8, "ใบเสร็จรับเงิน", "", 0, "C", false, 0, "")

	subHeaderY := 74.0

	pdf.SetFont("Sarabun", "B", 11)
	pdf.SetXY(20, subHeaderY)
	pdf.CellFormat(170, 6, "องค์การบริหารส่วนจังหวัดชลบุรี อำเภอเมืองชลบุรี จังหวัดชลบุรี", "", 0, "C", false, 0, "")

	receiptNo := getReceiptNumber(decl)

	paymentDate := time.Now()
	if decl.PaidAt != nil {
		paymentDate = *decl.PaidAt
	}
	dateStr := formatThaiDate(paymentDate)

	pdf.SetFont("Sarabun", "", 9)

	pdf.SetXY(135, 54)
	pdf.Cell(15, 5, "เลขที่")
	pdf.SetXY(150, 54)
	pdf.Cell(40, 5, receiptNo)

	pdf.SetXY(135, 60)
	pdf.Cell(15, 5, "วันที่")
	pdf.SetXY(150, 60)
	pdf.Cell(40, 5, dateStr)

	payerName := ""
	if decl.Business != nil {
		payerName = decl.Business.NameTH
		if decl.TaxType == "tobacco_tax" && decl.Business.OwnerName != nil && *decl.Business.OwnerName != "" {
			payerName = fmt.Sprintf("%s จ่ายโดย %s", decl.Business.NameTH, *decl.Business.OwnerName)
		}
	} else {
		payerName = decl.BusinessRegNumber
	}

	pdf.SetFont("Sarabun", "", 10)
	pdf.SetXY(20, 84)
	pdf.Cell(19, 6, "ได้รับเงินจาก  ")
	pdf.SetFont("Sarabun", "B", 10)
	pdf.Cell(140, 6, payerName)

	tableTop := 92.0
	headerHeight := 10.0
	itemHeight := 12.0
	totalHeight := 10.0
	tableBottom := tableTop + headerHeight + itemHeight + totalHeight

	pdf.SetFillColor(245, 245, 245)
	pdf.Rect(20, tableTop, 170, headerHeight, "F")

	pdf.SetDrawColor(180, 180, 180)
	pdf.SetLineWidth(0.2)

	pdf.Rect(20, tableTop, 170, headerHeight+itemHeight+totalHeight, "D")
	pdf.Line(20, tableTop+headerHeight, 190, tableTop+headerHeight)
	pdf.Line(20, tableTop+headerHeight+itemHeight, 190, tableTop+headerHeight+itemHeight)

	pdf.Line(35, tableTop, 35, tableTop+headerHeight+itemHeight)
	pdf.Line(130, tableTop, 130, tableBottom)
	pdf.Line(160, tableTop, 160, tableBottom)

	pdf.SetFont("Sarabun", "B", 11)
	pdf.SetXY(20, tableTop+2)
	pdf.CellFormat(15, 6, "ลำดับ", "", 0, "C", false, 0, "")
	pdf.CellFormat(95, 6, "รายการ", "", 0, "C", false, 0, "")
	pdf.CellFormat(30, 6, "จำนวนเงิน (บาท)", "", 0, "C", false, 0, "")
	pdf.CellFormat(30, 6, "หมายเหตุ", "", 0, "C", false, 0, "")

	itemY := tableTop + headerHeight
	pdf.SetFont("Sarabun", "", 11)

	pdf.SetXY(20, itemY+3)
	pdf.CellFormat(15, 6, "1", "", 0, "C", false, 0, "")

	taxTypeLabel := ""
	switch decl.TaxType {
	case "hotel_fee":
		taxTypeLabel = "ค่าธรรมเนียมบำรุง อบจ. จากผู้เข้าพักโรงแรม"
	case "tobacco_tax":
		taxTypeLabel = "ภาษีบำรุง อบจ.จากการค้ายาสูบ"
	case "oil_gas_tax":
		taxTypeLabel = "ภาษีบำรุง อบจ.จากการค้าน้ำมัน/ก๊าซ"
	default:
		taxTypeLabel = "ภาษี/ค่าธรรมเนียม"
	}
	pdf.SetXY(37, itemY+3)
	pdf.CellFormat(91, 6, taxTypeLabel, "", 0, "L", false, 0, "")

	amountStr := formatAmount(decl.CalculatedTax)
	pdf.SetXY(130, itemY+3)
	pdf.CellFormat(28, 6, amountStr, "", 0, "R", false, 0, "")

	thaiMonthShort := getThaiMonthNameShort(decl.TaxMonth)
	thaiYearShort := (decl.TaxYear + 543) % 100
	periodStr := fmt.Sprintf("%s %d", thaiMonthShort, thaiYearShort)
	pdf.SetXY(160, itemY+3)
	pdf.CellFormat(30, 6, periodStr, "", 0, "C", false, 0, "")

	totalY := tableTop + headerHeight + itemHeight
	pdf.SetFont("Sarabun", "", 11)
	pdf.SetXY(35, totalY+2)
	pdf.CellFormat(95, 6, "รวมเงิน   ", "", 0, "R", false, 0, "")
	pdf.SetXY(130, totalY+2)
	pdf.CellFormat(28, 6, amountStr, "", 0, "R", false, 0, "")

	bahtTextY := tableBottom
	pdf.SetFillColor(245, 245, 245)
	pdf.Rect(20, bahtTextY, 170, 8, "FD")

	bahtText := ConvertToThaiBaht(decl.CalculatedTax)
	pdf.SetXY(23, bahtTextY+1)
	pdf.SetFont("Sarabun", "", 11)
	pdf.Cell(16, 6, "ตัวอักษร  ")
	pdf.SetFont("Sarabun", "B", 11)
	pdf.Cell(140, 6, fmt.Sprintf("(%s)", bahtText))

	sigY := 142.0
	pdf.SetFont("Sarabun", "", 11)
	pdf.SetXY(20, sigY)
	pdf.Cell(50, 6, "ไว้เป็นการถูกต้องแล้ว")

	pdf.SetXY(105, sigY)
	pdf.CellFormat(85, 6, "ลงชื่อ ............................................................ ผู้รับเงิน", "", 0, "C", false, 0, "")
	pdf.SetFont("Sarabun", "", 11)
	pdf.SetXY(105, sigY+6)
	pdf.CellFormat(85, 6, "(นางสาวพรรณพิไล ไกรยบุตร)", "", 0, "C", false, 0, "")
	pdf.SetXY(105, sigY+12)
	pdf.CellFormat(85, 6, "เจ้าพนักงานการเงินและบัญชีชำนาญงาน", "", 0, "C", false, 0, "")

	g.drawPaymentDetailsBox(pdf, decl)

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to output pdf stream: %w", err)
	}

	return buf.Bytes(), nil
}

func (g *receiptGenerator) drawWatermarks(pdf *gofpdf.Fpdf) {
	pdf.SetTextColor(230, 230, 230)
	pdf.SetFont("Sarabun", "", 15.0)

	cols := []float64{15, 95, 175}
	rows := []float64{45, 105, 165, 225, 285}

	for _, x := range cols {
		for _, y := range rows {
			pdf.TransformBegin()
			pdf.TransformRotate(22, x, y)
			pdf.Text(x, y, "องค์การบริหารส่วนจังหวัดชลบุรี")
			pdf.Text(x, y+6.5, "Chonburi")
			pdf.TransformEnd()
		}
	}
}

func (g *receiptGenerator) drawPaymentDetailsBox(pdf *gofpdf.Fpdf, decl *domain.TaxDeclaration) {
	boxTop := 172.0
	boxHeight := 30.0
	pdf.SetFillColor(252, 252, 252)
	pdf.Rect(20, boxTop, 170, boxHeight, "FD")

	pdf.SetFont("Sarabun", "", 10)
	pdf.SetTextColor(70, 70, 70)

	isWalkIn := decl.KTBReconciliationRecordID == nil

	if isWalkIn {

		pdf.SetXY(23, boxTop+4)
		pdf.SetFont("Sarabun", "", 10)
		pdf.Cell(115, 5, "ใบเสร็จรับเงินฉบับนี้จะสมบูรณ์เมื่อธนาคารได้สั่งจ่ายเงินตามเช็ค/แคชเชียร์เช็ค/ตั๋วแลกเงิน ตามรายละเอียดดังนี้")

		pdf.SetFont("Sarabun", "", 10)
		paymentDetails := "เงินสดรับชำระหน้าเคาน์เตอร์กองคลัง อบจ. ชลบุรี"
		if decl.AuditNotes != nil && strings.Contains(*decl.AuditNotes, "เช็ค") {
			paymentDetails = *decl.AuditNotes
		}

		pdf.SetXY(23, boxTop+10)
		if strings.Contains(paymentDetails, "เลขที่") {
			parts := strings.SplitN(paymentDetails, "เลขที่", 2)
			line1 := strings.TrimSpace(parts[0])
			line2 := "เลขที่ " + strings.TrimSpace(parts[1])

			pdf.Cell(115, 5, line1)
			pdf.SetXY(23, boxTop+16)
			pdf.Cell(115, 5, line2)
		} else {
			pdf.Cell(115, 5, paymentDetails)
		}
	} else {

		pdf.SetXY(23, boxTop+4)
		pdf.Cell(115, 5, "โอนเงินเข้าบัญชีธนาคารกรุงไทย จำกัด (มหาชน) สาขาบางปลาสร้อย")

		payDate := time.Now()
		if decl.PaidAt != nil {
			payDate = *decl.PaidAt
		}
		payDateStr := formatThaiDate(payDate)

		pdf.SetXY(23, boxTop+10)
		pdf.Cell(115, 5, fmt.Sprintf("เลขที่บัญชี 228-6-03209-2 วันที่ %s", payDateStr))
	}

	amountStr := formatAmount(decl.CalculatedTax)
	pdf.SetXY(140, boxTop+10)
	pdf.CellFormat(10, 5, ":", "", 0, "C", false, 0, "")
	pdf.CellFormat(36, 5, fmt.Sprintf("%s บาท", amountStr), "", 0, "R", false, 0, "")

	pdf.SetFont("Sarabun", "", 10)
	pdf.SetXY(130, boxTop+22)
	pdf.CellFormat(20, 5, "รวม :", "", 0, "R", false, 0, "")
	pdf.CellFormat(36, 5, fmt.Sprintf("%s บาท", amountStr), "", 0, "R", false, 0, "")
}

func getReceiptNumber(decl *domain.TaxDeclaration) string {

	var hash uint32 = 0
	idStr := decl.ID.String()
	for i := 0; i < len(idStr); i++ {
		hash = hash*31 + uint32(idStr[i])
	}
	num := (hash % 90000) + 10000
	thaiYear := (decl.TaxYear + 543) % 100
	return fmt.Sprintf("RCPT-%d/%02d", num, thaiYear)
}

func formatThaiDate(t time.Time) string {
	thaiMonth := getThaiMonthName(int(t.Month()))
	thaiYear := t.Year() + 543
	return fmt.Sprintf("%d %s %d", t.Day(), thaiMonth, thaiYear)
}

func getThaiMonthName(m int) string {
	months := []string{"", "มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน", "กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม"}
	if m >= 1 && m <= 12 {
		return months[m]
	}
	return ""
}

func getThaiMonthNameShort(m int) string {
	months := []string{"", "ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.", "ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค."}
	if m >= 1 && m <= 12 {
		return months[m]
	}
	return ""
}

func formatAmount(val float64) string {

	integerPart := int64(val)
	fractionPart := int64(math.Round((val - float64(integerPart)) * 100))

	str := fmt.Sprintf("%d", integerPart)
	length := len(str)
	var parts []string

	for i := length; i > 0; i -= 3 {
		start := i - 3
		if start < 0 {
			start = 0
		}
		parts = append([]string{str[start:i]}, parts...)
	}

	formattedInt := strings.Join(parts, ",")
	return fmt.Sprintf("%s.%02d", formattedInt, fractionPart)
}
