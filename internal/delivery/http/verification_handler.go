package http

import (
	"strconv"
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type VerificationHandler struct {
	useCase domain.AdminVerificationUseCase
}

func NewVerificationHandler(useCase domain.AdminVerificationUseCase) *VerificationHandler {
	return &VerificationHandler{useCase: useCase}
}

func (h *VerificationHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/verifications", jwtutil.RequireAuth())
	group.Get("/", h.GetVerifications)
	group.Get("/:id", h.GetVerificationByID)
	group.Post("/approve", h.Approve)
	group.Post("/reject", h.Reject)
}

func (h *VerificationHandler) GetVerifications(c fiber.Ctx) error {
	pageStr := c.Query("page_number")
	limitStr := c.Query("page_size")
	search := c.Query("search")
	status := c.Query("verification_status")

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	query := domain.VerificationQuery{
		PageNumber:         page,
		PageSize:           limit,
		Search:             search,
		VerificationStatus: status,
	}

	resp, err := h.useCase.GetVerifications(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(resp)
}

func (h *VerificationHandler) GetVerificationByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id format",
		})
	}

	item, err := h.useCase.GetVerificationByID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if item == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "verification request not found",
		})
	}

	return c.JSON(item)
}

func (h *VerificationHandler) Approve(c fiber.Ctx) error {
	adminID, err := GetAdminIDFromClaims(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req domain.ApproveVerificationRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.UserID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user_id is required",
		})
	}

	err = h.useCase.ApproveVerification(&req, adminID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "verification approved successfully",
	})
}

func (h *VerificationHandler) Reject(c fiber.Ctx) error {
	adminID, err := GetAdminIDFromClaims(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req domain.RejectVerificationRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.UserID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user_id is required",
		})
	}
	if req.Reason == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "reason is required",
		})
	}

	err = h.useCase.RejectVerification(req.UserID, req.Reason, adminID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "verification rejected successfully",
	})
}
