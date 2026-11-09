package http

import (
	"strconv"
	"time"

	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/internal/usecase"

	"github.com/gofiber/fiber/v3"
)

type WaterLevelHandler struct {
	useCase usecase.WaterLevelUseCase
}

func NewWaterLevelHandler(useCase usecase.WaterLevelUseCase) *WaterLevelHandler {
	return &WaterLevelHandler{
		useCase: useCase,
	}
}

// RegisterRoutes registers endpoints under router (typically with M2M or JWT protection)
func (h *WaterLevelHandler) RegisterRoutes(router fiber.Router) {
	registerRoutes := func(g fiber.Router) {
		g.Get("/devices", h.GetStations)
		g.Get("/stations", h.GetStations)
		g.Get("/devices/:serial", h.GetDeviceDetail)
		g.Get("/stations/:serial", h.GetDeviceDetail)
		g.Get("/devices/:serial/history", h.GetDeviceHistory)
		g.Get("/stations/:serial/history", h.GetDeviceHistory)
		g.Get("/devices/:serial/stats", h.GetDeviceHistory)
		g.Get("/devices/:serial/snapshots", h.GetSnapshots)
		g.Get("/stations/:serial/snapshots", h.GetSnapshots)
		g.Post("/devices/:serial/stream", h.InitStream)
		g.Post("/stations/:serial/stream", h.InitStream)
		g.Put("/devices/:serial/thresholds", h.UpdateDeviceThresholds)
		g.Put("/stations/:serial/thresholds", h.UpdateDeviceThresholds)
		g.Post("/devices/:serial/webrtc-offer", h.StartWebRTC)
		g.Post("/stations/:serial/webrtc-offer", h.StartWebRTC)
		g.Delete("/devices/:serial/stream", h.TerminateStream)
		g.Delete("/stations/:serial/stream", h.TerminateStream)
		g.Get("/alerts/active", h.GetActiveAlerts)

		// Alert Notification Settings & Card Templates
		g.Get("/alert-settings", h.GetAlertSettings)
		g.Post("/alert-settings", h.CreateAlertSetting)
		g.Put("/alert-settings/:id", h.UpdateAlertSetting)
		g.Delete("/alert-settings/:id", h.DeleteAlertSetting)
		g.Get("/card-templates", h.GetCardTemplates)
		g.Put("/card-templates", h.UpdateCardTemplates)
	}

	registerRoutes(router.Group("/water-level"))
	registerRoutes(router.Group("/flood"))
}

func (h *WaterLevelHandler) GetStations(c fiber.Ctx) error {
	ctx := c.Context()
	stations, err := h.useCase.GetStations(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"items":   stations,
		"total":   len(stations),
	})
}

func (h *WaterLevelHandler) GetDeviceDetail(c fiber.Ctx) error {
	ctx := c.Context()
	serial := c.Params("serial")
	if serial == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "serial is required",
		})
	}

	detail, err := h.useCase.GetDeviceDetail(ctx, serial)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    detail,
	})
}

func (h *WaterLevelHandler) GetDeviceHistory(c fiber.Ctx) error {
	ctx := c.Context()
	serial := c.Params("serial")
	if serial == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "serial is required",
		})
	}

	// Support explicit date-range (startDate/endDate or start/end in RFC3339)
	startDate := c.Query("startDate")
	if startDate == "" {
		startDate = c.Query("start")
	}
	endDate := c.Query("endDate")
	if endDate == "" {
		endDate = c.Query("end")
	}

	bkkLoc := time.FixedZone("Asia/Bangkok", 7*3600)
	if startDate != "" && len(startDate) == 10 {
		if t, err := time.ParseInLocation("2006-01-02", startDate, bkkLoc); err == nil {
			startDate = t.Format(time.RFC3339)
		}
	}
	if endDate != "" && len(endDate) == 10 {
		if t, err := time.ParseInLocation("2006-01-02", endDate, bkkLoc); err == nil {
			endDate = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second).Format(time.RFC3339)
		}
	}

	var history []domain.WaterHistoryPoint
	var err error

	if startDate != "" && endDate != "" {
		history, err = h.useCase.GetDeviceHistoryRange(ctx, serial, startDate, endDate)
	} else {
		timeframe := 168
		if tfStr := c.Query("timeframe"); tfStr != "" {
			if tf, err2 := strconv.Atoi(tfStr); err2 == nil && tf > 0 {
				timeframe = tf
			}
		}
		history, err = h.useCase.GetDeviceHistory(ctx, serial, timeframe)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"items":   history,
		"data":    history,
		"total":   len(history),
	})
}

func (h *WaterLevelHandler) GetSnapshots(c fiber.Ctx) error {
	ctx := c.Context()
	serial := c.Params("serial")
	if serial == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "serial is required",
		})
	}

	date := c.Query("date")
	if date == "" {
		date = c.Query("startDate")
	}
	logs, err := h.useCase.GetSnapshots(ctx, serial, date)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"items":   logs,
		"total":   len(logs),
	})
}

func (h *WaterLevelHandler) InitStream(c fiber.Ctx) error {
	ctx := c.Context()
	serial := c.Params("serial")
	if serial == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "serial is required",
		})
	}

	res, err := h.useCase.InitStream(ctx, serial)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(res)
}

func (h *WaterLevelHandler) TerminateStream(c fiber.Ctx) error {
	ctx := c.Context()
	serial := c.Params("serial")
	if serial == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "serial is required",
		})
	}

	err := h.useCase.TerminateStream(ctx, serial)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

func (h *WaterLevelHandler) StartWebRTC(c fiber.Ctx) error {
	ctx := c.Context()
	serial := c.Params("serial")
	if serial == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "serial is required",
		})
	}

	sdpOffer := string(c.Body())
	if sdpOffer == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "sdp offer body is required",
		})
	}

	sdpAnswer, err := h.useCase.RelayWebRTCOffer(ctx, serial, sdpOffer)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	c.Set("Content-Type", "application/sdp")
	return c.SendString(sdpAnswer)
}

func (h *WaterLevelHandler) GetActiveAlerts(c fiber.Ctx) error {
	ctx := c.Context()
	alerts, err := h.useCase.GetActiveAlerts(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"items":   alerts,
		"total":   len(alerts),
	})
}

func (h *WaterLevelHandler) UpdateDeviceThresholds(c fiber.Ctx) error {
	ctx := c.Context()
	serial := c.Params("serial")
	if serial == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "serial is required",
		})
	}

	var req domain.UpdateStationThresholdsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid request body: " + err.Error(),
		})
	}

	if err := h.useCase.UpdateStationThresholds(ctx, serial, &req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "thresholds updated successfully",
		"serial":  serial,
	})
}

// --- River / Water Level Alert Settings Handlers ---

func (h *WaterLevelHandler) GetAlertSettings(c fiber.Ctx) error {
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
	})
}

func (h *WaterLevelHandler) CreateAlertSetting(c fiber.Ctx) error {
	ctx := c.Context()
	var setting domain.WeatherAlertSetting
	if err := c.Bind().JSON(&setting); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if err := h.useCase.CreateAlertSetting(ctx, &setting); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    setting,
	})
}

func (h *WaterLevelHandler) UpdateAlertSetting(c fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "ID is required",
		})
	}
	var updates map[string]interface{}
	if err := c.Bind().JSON(&updates); err != nil {
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

func (h *WaterLevelHandler) DeleteAlertSetting(c fiber.Ctx) error {
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

func (h *WaterLevelHandler) GetCardTemplates(c fiber.Ctx) error {
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

func (h *WaterLevelHandler) UpdateCardTemplates(c fiber.Ctx) error {
	ctx := c.Context()
	var req struct {
		Severity      string `json:"severity"`
		BackgroundURL string `json:"background_url"`
		UpdatedBy     string `json:"updated_by"`
	}
	if err := c.Bind().JSON(&req); err != nil {
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

