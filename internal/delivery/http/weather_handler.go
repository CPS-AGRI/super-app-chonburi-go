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

	// Summary statistics for Web Admin
	group.Get("/stats", h.GetDashboardStats)

	// Radar & Satellite Frames / Tiles
	group.Get("/map/frames", h.GetMapFrames)
	group.Get("/tiles", h.GetMapFrames)

	// Cache invalidation
	group.Post("/flush-cache", h.FlushCache)
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
