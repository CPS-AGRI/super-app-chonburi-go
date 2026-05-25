package http

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"
	"super-app-chonburi-go/pkg/pdf"
	"super-app-chonburi-go/pkg/storage"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// TaxNewHandler handles requests for the self-declaration tax module
type TaxNewHandler struct {
	uc      domain.TaxNewUseCase
	storage storage.StorageProvider
}

// NewTaxNewHandler creates a new TaxNewHandler instance
func NewTaxNewHandler(uc domain.TaxNewUseCase, storage storage.StorageProvider) *TaxNewHandler {
	return &TaxNewHandler{
		uc:      uc,
		storage: storage,
	}
}

// RegisterRoutes registers the handlers to the fiber router
func (h *TaxNewHandler) RegisterRoutes(router fiber.Router) {
	// Mobile and public endpoints
	mobileGroup := router.Group("/tax-new")
	mobileGroup.Get("/business/:reg_number", h.GetBusiness)
	mobileGroup.Post("/declare", h.DeclareTax)
	mobileGroup.Get("/declare/:id", h.GetDeclaration)
	mobileGroup.Get("/declare/:id/receipt/pdf", h.DownloadReceiptPDF)
	mobileGroup.Post("/upload", h.UploadFile)

	// Admin endpoints
	adminGroup := router.Group("/admin/tax-new", jwtutil.RequireAuth())
	adminGroup.Post("/reconciliation/upload", h.UploadKTBReconciliation)
	adminGroup.Post("/elaas/upload", h.UploadElaasSummary)
	adminGroup.Get("/dashboard/summary", h.GetDashboardSummary)
	adminGroup.Get("/declarations", h.ListDeclarations)
	adminGroup.Get("/declare/:id/receipt/pdf", h.DownloadReceiptPDF)
	adminGroup.Post("/businesses/import", h.ImportBusinesses)
	adminGroup.Post("/declare/:id/audit-status", h.UpdateAuditStatus)
}


// GetBusiness searches for business by registration number
func (h *TaxNewHandler) GetBusiness(c fiber.Ctx) error {
	regNumber := c.Params("reg_number")
	if regNumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "registration number is required"})
	}

	business, err := h.uc.GetBusiness(regNumber)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    business,
	})
}

// DeclareTax submits a new declaration
func (h *TaxNewHandler) DeclareTax(c fiber.Ctx) error {
	var req domain.DeclareTaxRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body: " + err.Error()})
	}

	resp, err := h.uc.DeclareTax(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// GetDeclaration gets status of declaration
func (h *TaxNewHandler) GetDeclaration(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid declaration ID"})
	}

	declaration, err := h.uc.GetDeclaration(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if declaration == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "declaration not found"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"declaration_id":   declaration.ID,
			"business_name":    declaration.Business.NameTH,
			"tax_month":        declaration.TaxMonth,
			"tax_year":         declaration.TaxYear,
			"calculated_tax":   declaration.CalculatedTax,
			"payment_status":   declaration.PaymentStatus,
			"paid_at":          declaration.PaidAt,
			"ref1":             declaration.Ref1,
			"ref2":             declaration.Ref2,
			"qr_code_content":  declaration.QRCodeContent,
			"monthly_revenue":  declaration.MonthlyRevenue,
		},
	})
}

// UploadFile handles PDF/image uploads
func (h *TaxNewHandler) UploadFile(c fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to read file from form: " + err.Error()})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to open file: " + err.Error()})
	}
	defer file.Close()

	ext := filepath.Ext(fileHeader.Filename)
	uniqueFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	fileURL, err := h.storage.Upload(file, uniqueFilename)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to upload file: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"file_url": fileURL,
	})
}

// UploadKTBReconciliation handles KTB Reconciliation text file import
func (h *TaxNewHandler) UploadKTBReconciliation(c fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to read file from form: " + err.Error()})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to open file: " + err.Error()})
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to read file: " + err.Error()})
	}

	adminIDStr, err := GetAdminIDFromClaims(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	adminID := uuid.MustParse(adminIDStr)

	resp, err := h.uc.UploadKTBFile(fileHeader.Filename, fileBytes, adminID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// UploadElaasSummary handles e-LAAS Daily Summary CSV import
func (h *TaxNewHandler) UploadElaasSummary(c fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to read file from form: " + err.Error()})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to open file: " + err.Error()})
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to read file: " + err.Error()})
	}

	adminIDStr, err := GetAdminIDFromClaims(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	adminID := uuid.MustParse(adminIDStr)

	importedCount, err := h.uc.UploadElaasFile(fileHeader.Filename, fileBytes, adminID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"records_imported": importedCount,
		},
	})
}

// GetDashboardSummary retrieves the admin dashboard report summary
func (h *TaxNewHandler) GetDashboardSummary(c fiber.Ctx) error {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	resp, err := h.uc.GetDashboard(startDateStr, endDateStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ImportBusinesses imports registration records from CSV
func (h *TaxNewHandler) ImportBusinesses(c fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to read file from form: " + err.Error()})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to open file: " + err.Error()})
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to read file: " + err.Error()})
	}

	resp, err := h.uc.ImportBusinesses(fileBytes)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// UpdateAuditStatus changes status to audit_failed or verified
func (h *TaxNewHandler) UpdateAuditStatus(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid declaration ID"})
	}

	var req struct {
		Status string `json:"status" validate:"required"` // e.g. "audit_failed", "verified"
		Notes  string `json:"notes"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	adminIDStr, err := GetAdminIDFromClaims(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	adminID := uuid.MustParse(adminIDStr)

	err = h.uc.UpdateAuditStatus(id, req.Status, req.Notes, adminID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Audit status updated successfully",
	})
}

// ListDeclarations retrieves list of declarations with filters
func (h *TaxNewHandler) ListDeclarations(c fiber.Ctx) error {
	taxType := c.Query("tax_type")
	status := c.Query("payment_status")
	search := c.Query("search")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit := 50
	if limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil {
			limit = val
		}
	}

	offset := 0
	if offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil {
			offset = val
		}
	}

	declarations, total, err := h.uc.ListDeclarations(taxType, status, search, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"declarations": declarations,
			"total":        total,
		},
	})
}

// DownloadReceiptPDF generates and downloads the A4 PDF receipt for a declaration
func (h *TaxNewHandler) DownloadReceiptPDF(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid declaration ID"})
	}

	declaration, err := h.uc.GetDeclaration(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if declaration == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "declaration not found"})
	}

	// Verify that payment status is completed (paid or verified)
	if declaration.PaymentStatus != "paid" && declaration.PaymentStatus != "verified" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ยังไม่ได้ชำระเงิน"})
	}

	generator := pdf.NewReceiptGenerator()
	pdfBytes, err := generator.GenerateReceiptPDF(declaration)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate receipt PDF: " + err.Error()})
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=receipt-%s.pdf", declaration.Ref1))

	return c.Send(pdfBytes)
}


