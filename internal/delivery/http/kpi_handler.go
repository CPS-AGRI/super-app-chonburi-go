package http

import (
	"strconv"
	"strings"
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"
	"time"

	"github.com/gofiber/fiber/v3"
)

type KPIHandler struct {
	uc domain.KPIUseCase
}

func NewKPIHandler(uc domain.KPIUseCase) *KPIHandler {
	return &KPIHandler{uc: uc}
}

func (h *KPIHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/kpi/data-sharing-logs", h.GetDataSharingLogs)
	router.Get("/kpi/data-sharing-summary", h.GetDataSharingSummary)

	protected := router.Group("/kpi", jwtutil.RequireAuth())
	protected.Get("/summary", h.GetSummary)
	protected.Get("/staff-leaderboard", h.GetStaffLeaderboard)
	protected.Get("/manager-scorecard", h.GetManagerScorecard)
	protected.Get("/staff-detail", h.GetStaffDetail)
	protected.Get("/manager-detail", h.GetManagerDetail)
	protected.Get("/breached-cases", h.GetBreachedCases)
	protected.Get("/category-stats", h.GetCategoryStats)
	protected.Get("/thresholds", h.GetThresholds)
	protected.Put("/thresholds", h.UpdateThresholds)
}

func parseKPIFilter(c fiber.Ctx) domain.KPIQueryFilter {
	filter := domain.KPIQueryFilter{
		ModuleKey:    c.Query("module_key", domain.ModuleKeyComplaint),
		DepartmentID: c.Query("department_id"),
		UserID:       c.Query("user_id"),
	}

	deptIDsStr := c.Query("department_ids")
	if deptIDsStr != "" {
		for _, s := range strings.Split(deptIDsStr, ",") {
			trimmed := strings.TrimSpace(s)
			if trimmed != "" {
				filter.DepartmentIDs = append(filter.DepartmentIDs, trimmed)
			}
		}
	}

	startDateStr := c.Query("start_date")
	if startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = &t
		} else if tRFC, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filter.StartDate = &tRFC
		}
	}

	endDateStr := c.Query("end_date")
	if endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// end of the day
			tEnd := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			filter.EndDate = &tEnd
		} else if tRFC, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filter.EndDate = &tRFC
		}
	}

	return filter
}

func (h *KPIHandler) GetSummary(c fiber.Ctx) error {
	filter := parseKPIFilter(c)

	summary, err := h.uc.GetSummary(filter)
	if err != nil {
		return ErrorResponse(c, "Failed to fetch KPI summary: "+err.Error())
	}

	return SuccessResponse(c, summary)
}

func (h *KPIHandler) GetStaffLeaderboard(c fiber.Ctx) error {
	filter := parseKPIFilter(c)

	leaderboard, err := h.uc.GetStaffLeaderboard(filter)
	if err != nil {
		return ErrorResponse(c, "Failed to fetch staff leaderboard: "+err.Error())
	}

	return SuccessResponse(c, leaderboard)
}

func (h *KPIHandler) GetManagerScorecard(c fiber.Ctx) error {
	filter := parseKPIFilter(c)

	scorecard, err := h.uc.GetManagerScorecard(filter)
	if err != nil {
		return ErrorResponse(c, "Failed to fetch manager scorecard: "+err.Error())
	}

	return SuccessResponse(c, scorecard)
}

func (h *KPIHandler) GetStaffDetail(c fiber.Ctx) error {
	userID := c.Query("user_id")
	if userID == "" {
		return ErrorResponse(c, "user_id is required")
	}

	filter := parseKPIFilter(c)

	detail, err := h.uc.GetStaffDetail(userID, filter)
	if err != nil {
		return ErrorResponse(c, "Failed to fetch staff detail: "+err.Error())
	}

	return SuccessResponse(c, detail)
}

func (h *KPIHandler) GetManagerDetail(c fiber.Ctx) error {
	userID := c.Query("user_id")
	if userID == "" {
		return ErrorResponse(c, "user_id is required")
	}

	filter := parseKPIFilter(c)

	detail, err := h.uc.GetManagerDetail(userID, filter)
	if err != nil {
		return ErrorResponse(c, "Failed to fetch manager detail: "+err.Error())
	}

	return SuccessResponse(c, detail)
}

func (h *KPIHandler) GetBreachedCases(c fiber.Ctx) error {
	filter := parseKPIFilter(c)
	cases, err := h.uc.GetBreachedCases(filter)
	if err != nil {
		return ErrorResponse(c, "Failed to fetch breached cases: "+err.Error())
	}
	return SuccessResponse(c, cases)
}

func (h *KPIHandler) GetCategoryStats(c fiber.Ctx) error {
	filter := parseKPIFilter(c)
	stats, err := h.uc.GetCategoryStats(filter)
	if err != nil {
		return ErrorResponse(c, "Failed to fetch category stats: "+err.Error())
	}
	return SuccessResponse(c, stats)
}

func (h *KPIHandler) GetThresholds(c fiber.Ctx) error {
	moduleKey := c.Query("module_key", domain.ModuleKeyComplaint)
	thresholds, err := h.uc.GetThresholds(moduleKey)
	if err != nil {
		return ErrorResponse(c, "Failed to fetch thresholds: "+err.Error())
	}
	return SuccessResponse(c, thresholds)
}

func (h *KPIHandler) UpdateThresholds(c fiber.Ctx) error {
	claims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if !ok || claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	role := strings.ToLower(claims.Role)
	if role != "executives" && role != "executive" && role != "superadmin" && role != "super_admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "เฉพาะผู้บริหารระดับสูง (Executives) หรือ SuperAdmin เท่านั้นที่สามารถกำหนดเกณฑ์ SLA ได้",
		})
	}

	var req domain.UpdateKPIThresholdRequest
	if err := c.Bind().Body(&req); err != nil {
		return ErrorResponse(c, "Invalid request body: "+err.Error())
	}

	if req.ModuleKey == "" {
		req.ModuleKey = domain.ModuleKeyComplaint
	}

	updated, err := h.uc.UpdateThresholds(req.ModuleKey, req, claims.ID)
	if err != nil {
		return ErrorResponse(c, "Failed to update thresholds: "+err.Error())
	}

	return SuccessResponse(c, updated)
}

func (h *KPIHandler) GetDataSharingLogs(c fiber.Ctx) error {
	module := c.Query("module")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	limitStr := c.Query("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	logs, err := h.uc.GetDataSharingLogs(module, startDate, endDate, limit)
	if err != nil {
		return ErrorResponse(c, "Failed to fetch data sharing logs: "+err.Error())
	}
	return c.JSON(fiber.Map{
		"success": true,
		"items":   logs,
		"total":   len(logs),
	})
}

func (h *KPIHandler) GetDataSharingSummary(c fiber.Ctx) error {
	module := c.Query("module")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	summary, err := h.uc.GetDataSharingSummary(module, startDate, endDate)
	if err != nil {
		return ErrorResponse(c, "Failed to fetch data sharing summary: "+err.Error())
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    summary,
	})
}

