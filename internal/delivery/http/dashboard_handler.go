package http

import (
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"
	"time"

	"github.com/gofiber/fiber/v3"
)

type DashboardHandler struct {
	uc domain.DashboardUseCase
}

func NewDashboardHandler(uc domain.DashboardUseCase) *DashboardHandler {
	return &DashboardHandler{uc: uc}
}

func (h *DashboardHandler) RegisterRoutes(router fiber.Router) {
	dashboard := router.Group("/dashboard")
	protected := dashboard.Group("", jwtutil.RequireAuth())
	
	protected.Get("/back-office", h.GetBackOffice)
	protected.Post("/seed", h.SeedMockData)
}

func (h *DashboardHandler) GetBackOffice(c fiber.Ctx) error {
	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if !ok || userClaims == nil {
		return ErrorResponse(c, "Unauthorized", fiber.StatusUnauthorized)
	}

	filter := domain.DashboardFilter{}

	startDateStr := c.Query("start_date")
	if startDateStr != "" {
		if t, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filter.StartDate = &t
		}
	}

	endDateStr := c.Query("end_date")
	if endDateStr != "" {
		if t, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filter.EndDate = &t
		}
	}

	result, err := h.uc.GetBackOffice(filter)
	if err != nil {
		return ErrorResponse(c, "Failed to load dashboard data")
	}

	return SuccessResponse(c, result)
}

func (h *DashboardHandler) SeedMockData(c fiber.Ctx) error {
	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if !ok || userClaims == nil {
		return ErrorResponse(c, "Unauthorized", fiber.StatusUnauthorized)
	}

	if err := h.uc.SeedMockData(""); err != nil {
		return ErrorResponse(c, "Failed to seed mock data: "+err.Error())
	}

	return SuccessResponse(c, fiber.Map{"message": "Mock data seeded successfully!"})
}
