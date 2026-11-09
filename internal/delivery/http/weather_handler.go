package http

import (
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/internal/usecase"

	"github.com/gofiber/fiber/v3"
)

type WeatherHandler struct {
	useCase usecase.WeatherUseCase
}

func NewWeatherHandler(useCase usecase.WeatherUseCase) *WeatherHandler {
	return &WeatherHandler{
		useCase: useCase,
	}
}

// RegisterRoutes registers weather & radar map endpoints under the given router
func (h *WeatherHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/weather")

	// Station & Device endpoints
	group.Get("/stations", h.GetStations)
	group.Get("/devices", h.GetStations)
	group.Get("/stations/:serial", h.GetStationBySerial)
	group.Get("/devices/:serial", h.GetStationBySerial)
	group.Get("/stations/:serial/history", h.GetStationHistory)
	group.Get("/devices/:serial/history", h.GetStationHistory)

	// Summary statistics for Web Admin
	group.Get("/stats", h.GetDashboardStats)

	// Radar & Satellite Frames / Tiles
	group.Get("/map/frames", h.GetMapFrames)
	group.Get("/tiles", h.GetMapFrames)

	// Cache invalidation
	group.Post("/flush-cache", h.FlushCache)

	// Alert Settings CRUD (Real Database in PostgreSQL)
	group.Get("/alert-settings", h.GetAlertSettings)
	group.Post("/alert-settings", h.CreateAlertSetting)
	group.Put("/alert-settings/:id", h.UpdateAlertSetting)
	group.Delete("/alert-settings/:id", h.DeleteAlertSetting)

	// Card Templates (3 Severities)
	group.Get("/card-templates", h.GetCardTemplates)
	group.Put("/card-templates", h.UpdateCardTemplates)

	// Station Metadata & Custom Aliases
	group.Get("/stations/:serial/metadata", h.GetStationMetadata)
	group.Put("/stations/:serial/metadata", h.UpdateStationMetadata)

	// Test Alert Dispatch for Subdistricts
	group.Post("/test-alert", h.SendTestAlert)
}

func (h *WeatherHandler) GetStations(c fiber.Ctx) error {
	ctx := c.Context()
	scope := c.Query("scope", "all")
	adminZone := c.Query("admin_zone")
	district := c.Query("district")
	search := c.Query("search")
	status := c.Query("status")

	filter := domain.WeatherQueryFilter{
		Scope:     scope,
		AdminZone: adminZone,
		District:  district,
		Search:    search,
		Status:    status,
	}

	sensors, err := h.useCase.GetSensors(ctx, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	region := "จังหวัดชลบุรีทั้งหมด"
	if scope == "pattaya" {
		region = "เมืองพัทยาและพื้นที่โดยรอบ (165 จุดตรวจวัด)"
	}

	return c.JSON(fiber.Map{
		"success": true,
		"total":   len(sensors),
		"scope":   scope,
		"region":  region,
		"sensors": sensors,
		"items":   sensors, // Compatible with both Admin and Pattaya schemas
	})
}

func (h *WeatherHandler) GetStationBySerial(c fiber.Ctx) error {
	ctx := c.Context()
	serial := c.Params("serial")

	sensor, err := h.useCase.GetSensorBySerial(ctx, serial)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"sensor":  sensor,
		"data":    sensor,
	})
}

func (h *WeatherHandler) GetStationHistory(c fiber.Ctx) error {
	ctx := c.Context()
	serial := c.Params("serial")
	rangeType := c.Query("range", "hourly")
	date := c.Query("date")
	startDate := c.Query("startDate", c.Query("start_date"))
	endDate := c.Query("endDate", c.Query("end_date"))

	history, err := h.useCase.GetStationHistory(ctx, serial, rangeType, date, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"history": history,
		"items":   history,
		"total":   len(history),
	})
}

func (h *WeatherHandler) GetDashboardStats(c fiber.Ctx) error {
	ctx := c.Context()
	stats, err := h.useCase.GetDashboardStats(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"stats":   stats,
		"data":    stats,
	})
}

func (h *WeatherHandler) GetMapFrames(c fiber.Ctx) error {
	ctx := c.Context()
	variable := c.Query("variable", "rain")
	animType := c.Query("anim_type")
	if animType == "" {
		animType = c.Query("type", "nrt")
	}
	date := c.Query("date")

	frames, err := h.useCase.GetMapFrames(ctx, variable, animType, date)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"variable":  variable,
		"anim_type": animType,
		"date":      date,
		"total":     len(frames),
		"frames":    frames,
		"data": fiber.Map{
			"frames": frames,
		},
	})
}

func (h *WeatherHandler) FlushCache(c fiber.Ctx) error {
	ctx := c.Context()
	h.useCase.FlushCache(ctx)
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Weather sensors and frames cache flushed successfully",
	})
}

// --- Alert Settings Handlers ---

func (h *WeatherHandler) GetAlertSettings(c fiber.Ctx) error {
	ctx := c.Context()
	settings, err := h.useCase.GetAlertSettings(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    settings,
		"items":   settings,
		"total":   len(settings),
	})
}

func (h *WeatherHandler) CreateAlertSetting(c fiber.Ctx) error {
	ctx := c.Context()
	var item domain.WeatherAlertSetting
	if err := c.Bind().Body(&item); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}
	if err := h.useCase.CreateAlertSetting(ctx, &item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    item,
	})
}

func (h *WeatherHandler) UpdateAlertSetting(c fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "ID is required",
		})
	}
	var updates map[string]interface{}
	if err := c.Bind().Body(&updates); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}
	delete(updates, "id")
	delete(updates, "created_at")
	updated, err := h.useCase.UpdateAlertSetting(ctx, id, updates)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    updated,
	})
}

func (h *WeatherHandler) DeleteAlertSetting(c fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")
	if err := h.useCase.DeleteAlertSetting(ctx, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Setting deleted successfully",
	})
}

// --- Card Templates Handlers ---

func (h *WeatherHandler) GetCardTemplates(c fiber.Ctx) error {
	ctx := c.Context()
	templates, err := h.useCase.GetCardTemplates(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    templates,
	})
}

func (h *WeatherHandler) UpdateCardTemplates(c fiber.Ctx) error {
	ctx := c.Context()
	var req struct {
		Severity      string `json:"severity"`
		BackgroundURL string `json:"background_url"`
		UpdatedBy     string `json:"updated_by"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}
	if req.Severity == "" || req.BackgroundURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "severity and background_url are required",
		})
	}
	if err := h.useCase.UpdateCardTemplate(ctx, req.Severity, req.BackgroundURL, req.UpdatedBy); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Card template updated successfully",
	})
}

// --- Station Metadata & Custom Aliases Handlers ---

func (h *WeatherHandler) GetStationMetadata(c fiber.Ctx) error {
	ctx := c.Context()
	serial := c.Params("serial")
	meta, err := h.useCase.GetStationMetadata(ctx, serial)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Station metadata not found",
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    meta,
	})
}

func (h *WeatherHandler) UpdateStationMetadata(c fiber.Ctx) error {
	ctx := c.Context()
	serial := c.Params("serial")
	var req domain.WeatherStationMetadata
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}
	req.Serial = serial
	if err := h.useCase.UpsertStationMetadata(ctx, &req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    req,
		"message": "Station metadata updated successfully",
	})
}

// SendTestAlert handles manual/test weather alerts for citizen subscribers of a subdistrict.
func (h *WeatherHandler) SendTestAlert(c fiber.Ctx) error {
	ctx := c.Context()
	var req struct {
		Subdistrict string `json:"subdistrict"`
		District    string `json:"district"`
		Severity    string `json:"severity"` // "info", "warning", "danger"
		ImageURL    string `json:"image_url"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if req.Subdistrict == "" {
		req.Subdistrict = "อ่างศิลา"
	}
	if req.District == "" {
		req.District = "เมืองชลบุรี"
	}
	if req.Severity == "" {
		req.Severity = "info"
	}

	count, err := h.useCase.DispatchTestAlert(ctx, req.Subdistrict, req.District, req.Severity, req.ImageURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":     true,
		"subdistrict": req.Subdistrict,
		"district":    req.District,
		"severity":    req.Severity,
		"image_url":   req.ImageURL,
		"sent_to":     count,
		"message":     "ส่งข้อความแจ้งเตือนทดสอบสำเร็จ",
	})
}
