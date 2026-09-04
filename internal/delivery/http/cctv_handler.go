package http

import (
	"log"
	"strconv"
	"strings"
	"time"

	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type CCTVHandler struct {
	useCase domain.AdminCCTVUseCase
}

// NewCCTVHandler creates a new CCTVHandler.
func NewCCTVHandler(useCase domain.AdminCCTVUseCase) *CCTVHandler {
	return &CCTVHandler{useCase: useCase}
}

// RegisterRoutes registers Admin API CCTV endpoints.
func (h *CCTVHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/cctv", jwtutil.RequireAuth())
	group.Post("/", h.CreateCCTV)
	group.Get("/", h.GetCCTVs)
	group.Delete("/:id", h.DeleteCCTV)
	group.Get("/requests", h.GetRequests)
	group.Put("/requests/:id/approve", h.ApproveRequest)
	group.Put("/requests/:id/reject", h.RejectRequest)
	group.Post("/logs", h.CreateViewLog)
	group.Get("/logs/recent", h.GetRecentLogs)
	group.Get("/logs/summary", h.GetUserSummaryLogs)
}

type createCCTVInput struct {
	Name           string  `json:"name"`
	Location       string  `json:"location"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	StreamURL      string  `json:"stream_url"`
	PermissionType string  `json:"permission_type"` // 'all', 'staff_only'
	Status         string  `json:"status"`          // 'online', 'offline'
}

func (h *CCTVHandler) CreateCCTV(c fiber.Ctx) error {
	adminIDStr, _ := GetAdminIDFromClaims(c)

	var input createCCTVInput
	if err := c.Bind().JSON(&input); err != nil {
		return ErrorResponse(c, "invalid request body", fiber.StatusBadRequest)
	}

	status := "ONLINE"
	if strings.ToUpper(input.Status) == "OFFLINE" {
		status = "OFFLINE"
	}

	accessLevel := "PUBLIC"
	if input.PermissionType == "staff_only" {
		accessLevel = "STAFF_ONLY"
	}

	cam := domain.CCTV{
		ID:          uuid.New(),
		Name:        input.Name,
		Address:     input.Location,
		Latitude:    input.Latitude,
		Longitude:   input.Longitude,
		StreamURL:   input.StreamURL,
		Status:      status,
		AccessLevel: accessLevel,
	}

	if adminIDStr != "" {
		cam.CreatorID = &adminIDStr
	}

	if err := h.useCase.CreateCCTV(&cam); err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusBadRequest)
	}

	return SuccessResponse(c, cam, fiber.StatusCreated)
}

type cctvListResponse struct {
	ID             string  `json:"id"`
	UUID           string  `json:"uuid"`
	Name           string  `json:"name"`
	Location       string  `json:"location"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	StreamURL      string  `json:"stream_url"`
	Status         string  `json:"status"`          // 'online', 'offline'
	AccessLevel    string  `json:"access_level"`    // 'PUBLIC', 'STAFF_ONLY'
	PermissionType string  `json:"permission_type"` // 'all', 'staff_only'
	CreatedBy      *string `json:"created_by,omitempty"`
	DeletedBy      *string `json:"deleted_by,omitempty"`
	SnapshotURL    string  `json:"snapshot_url"`
}

func (h *CCTVHandler) GetCCTVs(c fiber.Ctx) error {
	pageStr := c.Query("page_number")
	limitStr := c.Query("page_size")
	name := c.Query("name")

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

	query := domain.CCTVQuery{
		PageNumber: page,
		PageSize:   limit,
		Name:       name,
	}

	resp, err := h.useCase.GetCCTVs(query)
	if err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusInternalServerError)
	}

	// Map to UI-friendly structure
	mappedItems := make([]cctvListResponse, len(resp.Items))
	for i, item := range resp.Items {
		status := "online"
		if strings.ToUpper(item.Status) == "OFFLINE" {
			status = "offline"
		}

		permType := "all"
		if item.AccessLevel == "STAFF_ONLY" {
			permType = "staff_only"
		}

		var createdByVal *string
		if item.Creator != nil {
			fullName := item.Creator.Name + " " + item.Creator.LastName
			createdByVal = &fullName
		} else if item.CreatorID != nil {
			createdByVal = item.CreatorID
		}

		var deletedByVal *string
		if item.Deleter != nil {
			fullName := item.Deleter.Name + " " + item.Deleter.LastName
			deletedByVal = &fullName
		} else if item.DeleterID != nil {
			deletedByVal = item.DeleterID
		}

		mappedItems[i] = cctvListResponse{
			ID:             item.Name, // Bind "id" to readable Name (e.g. CCTV01) for UI compatibility
			UUID:           item.ID.String(),
			Name:           item.Name,
			Location:       item.Address,
			Latitude:       item.Latitude,
			Longitude:      item.Longitude,
			StreamURL:      item.StreamURL,
			Status:         status,
			AccessLevel:    item.AccessLevel,
			PermissionType: permType,
			CreatedBy:      createdByVal,
			DeletedBy:      deletedByVal,
			SnapshotURL:    item.SnapshotURL,
		}
	}

	return c.JSON(fiber.Map{
		"items":       mappedItems,
		"total_items": resp.TotalItems,
		"page_number": resp.PageNumber,
		"total_pages": resp.TotalPages,
	})
}

type cctvRequestResponse struct {
	ID             string  `json:"id"`
	CameraID       string  `json:"camera_id"`
	CameraName     string  `json:"camera_name"`
	RequesterName  string  `json:"requester_name"`
	RequesterPhone string  `json:"requester_phone"`
	RequestedDate  string  `json:"requested_date"` // DD/MM/YYYY
	TimeRange      string  `json:"time_range"`      // HH:MM - HH:MM
	Reason         string  `json:"reason"`
	DocumentURL    *string `json:"document_url"`
	Status         string  `json:"status"` // 'pending', 'approved', 'rejected'
	VideoURL       *string `json:"video_url,omitempty"`
	RejectedReason *string `json:"rejected_reason,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

func (h *CCTVHandler) GetRequests(c fiber.Ctx) error {
	pageStr := c.Query("page_number")
	limitStr := c.Query("page_size")

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

	query := domain.CCTVRequestQuery{
		PageNumber: page,
		PageSize:   limit,
	}

	resp, err := h.useCase.GetCCTVRequests(query)
	if err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusInternalServerError)
	}

	mappedItems := make([]cctvRequestResponse, len(resp.Items))
	for i, item := range resp.Items {
		cameraName := item.CCTV.Name
		if cameraName == "" {
			cameraName = "CCTV Device"
		}

		requesterName := ""
		requesterPhone := ""

		// If user information is preloaded, fetch full name and phone number dynamically
		if item.User != nil && item.User.Information != nil {
			if item.User.Information.Name != "" {
				requesterName = item.User.Information.Name + " " + item.User.Information.LastName
			}
			if item.User.Information.Phone != "" {
				requesterPhone = item.User.Information.Phone
			}
		}

		status := strings.ToLower(item.Status)

		var docURL *string
		if item.EvidenceFileURL != "" {
			docURL = &item.EvidenceFileURL
		}

		timeRange := item.StartTime + " - " + item.EndTime

		mappedItems[i] = cctvRequestResponse{
			ID:             item.ID.String(),
			CameraID:       cameraName,
			CameraName:     cameraName,
			RequesterName:  requesterName,
			RequesterPhone: requesterPhone,
			RequestedDate:  item.IncidentDate.Format("02/01/2006"), // format as DD/MM/YYYY
			TimeRange:      timeRange,
			Reason:         item.Reason,
			DocumentURL:    docURL,
			Status:         status,
			VideoURL:       item.ResponseFileURL,
			RejectedReason: item.RejectReason,
			CreatedAt:      item.CreatedAt.Format(time.RFC3339),
		}
	}

	return c.JSON(fiber.Map{
		"items":       mappedItems,
		"total_items": resp.TotalItems,
		"page_number": resp.PageNumber,
		"total_pages": resp.TotalPages,
	})
}

type approveReqBody struct {
	ResponseFileURL string `json:"response_file_url"`
}

func (h *CCTVHandler) ApproveRequest(c fiber.Ctx) error {
	adminIDStr, err := GetAdminIDFromClaims(c)
	if err != nil {
		return ErrorResponse(c, "unauthorized", fiber.StatusUnauthorized)
	}

	adminID, err := uuid.Parse(adminIDStr)
	if err != nil {
		return ErrorResponse(c, "invalid admin id format", fiber.StatusBadRequest)
	}

	idStr := c.Params("id")
	reqID, err := uuid.Parse(idStr)
	if err != nil {
		return ErrorResponse(c, "invalid request id format", fiber.StatusBadRequest)
	}

	var body approveReqBody
	if err := c.Bind().JSON(&body); err != nil {
		return ErrorResponse(c, "invalid request body", fiber.StatusBadRequest)
	}

	if body.ResponseFileURL == "" {
		return ErrorResponse(c, "response_file_url is required", fiber.StatusBadRequest)
	}

	if err := h.useCase.ApproveRequest(reqID, body.ResponseFileURL, adminID); err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusInternalServerError)
	}

	return SuccessResponse(c, fiber.Map{"message": "request approved successfully"})
}

type rejectReqBody struct {
	RejectReason string `json:"reject_reason"`
}

func (h *CCTVHandler) RejectRequest(c fiber.Ctx) error {
	adminIDStr, err := GetAdminIDFromClaims(c)
	if err != nil {
		return ErrorResponse(c, "unauthorized", fiber.StatusUnauthorized)
	}

	adminID, err := uuid.Parse(adminIDStr)
	if err != nil {
		return ErrorResponse(c, "invalid admin id format", fiber.StatusBadRequest)
	}

	idStr := c.Params("id")
	reqID, err := uuid.Parse(idStr)
	if err != nil {
		return ErrorResponse(c, "invalid request id format", fiber.StatusBadRequest)
	}

	var body rejectReqBody
	if err := c.Bind().JSON(&body); err != nil {
		return ErrorResponse(c, "invalid request body", fiber.StatusBadRequest)
	}

	if body.RejectReason == "" {
		return ErrorResponse(c, "reject_reason is required", fiber.StatusBadRequest)
	}

	if err := h.useCase.RejectRequest(reqID, body.RejectReason, adminID); err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusInternalServerError)
	}

	return SuccessResponse(c, fiber.Map{"message": "request rejected successfully"})
}

func (h *CCTVHandler) DeleteCCTV(c fiber.Ctx) error {
	adminIDStr, err := GetAdminIDFromClaims(c)
	if err != nil {
		return ErrorResponse(c, "unauthorized", fiber.StatusUnauthorized)
	}

	idStr := c.Params("id")
	cctvID, err := uuid.Parse(idStr)
	if err != nil {
		return ErrorResponse(c, "invalid camera id format", fiber.StatusBadRequest)
	}

	if err := h.useCase.DeleteCCTV(cctvID, adminIDStr); err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusInternalServerError)
	}

	return SuccessResponse(c, fiber.Map{"message": "camera deleted successfully"})
}

func (h *CCTVHandler) CreateViewLog(c fiber.Ctx) error {
	log.Printf("[CCTV CreateViewLog] Received body: %s", string(c.Body()))

	adminIDStr, err := GetAdminIDFromClaims(c)
	if err != nil || adminIDStr == "" {
		log.Printf("[CCTV CreateViewLog] Auth error: %v, adminIDStr=%s", err, adminIDStr)
		return ErrorResponse(c, "unauthorized", fiber.StatusUnauthorized)
	}

	adminID, err := uuid.Parse(adminIDStr)
	if err != nil {
		log.Printf("[CCTV CreateViewLog] UUID parse error for adminID=%s: %v", adminIDStr, err)
		return ErrorResponse(c, "invalid user id in claims", fiber.StatusBadRequest)
	}

	var input domain.CreateCCTVLogInput
	if err := c.Bind().JSON(&input); err != nil {
		log.Printf("[CCTV CreateViewLog] JSON bind error: %v", err)
		return ErrorResponse(c, "invalid request body", fiber.StatusBadRequest)
	}

	if input.CCTVID == "" {
		return ErrorResponse(c, "cctv_id is required", fiber.StatusBadRequest)
	}

	clientIP := c.IP()
	userAgent := c.Get("User-Agent")

	if err := h.useCase.RecordViewLog(input, adminID, clientIP, userAgent); err != nil {
		log.Printf("[CCTV CreateViewLog] RecordViewLog returned error: %v", err)
		return ErrorResponse(c, err.Error(), fiber.StatusInternalServerError)
	}

	log.Printf("[CCTV CreateViewLog] Successfully recorded log for CCTV=%s, User=%s", input.CCTVID, adminID)
	return SuccessResponse(c, fiber.Map{"message": "log recorded successfully"})
}

func (h *CCTVHandler) GetRecentLogs(c fiber.Ctx) error {
	pageStr := c.Query("page_number")
	limitStr := c.Query("page_size")

	page := 1
	limit := 15

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	query := domain.CCTVLogQuery{
		PageNumber: page,
		PageSize:   limit,
	}

	if cctvIDStr := c.Query("cctv_id"); cctvIDStr != "" {
		if cID, err := uuid.Parse(cctvIDStr); err == nil {
			query.CCTVID = &cID
		}
	}

	resp, err := h.useCase.GetRecentLogs(query)
	if err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{
		"items":       resp.Items,
		"total_items": resp.TotalItems,
		"page_number": resp.PageNumber,
		"total_pages": resp.TotalPages,
	})
}

func (h *CCTVHandler) GetUserSummaryLogs(c fiber.Ctx) error {
	pageStr := c.Query("page_number")
	limitStr := c.Query("page_size")

	page := 1
	limit := 15

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	query := domain.CCTVLogQuery{
		PageNumber: page,
		PageSize:   limit,
		Search:     c.Query("search"),
	}

	resp, err := h.useCase.GetUserSummaryLogs(query)
	if err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{
		"items":       resp.Items,
		"total_items": resp.TotalItems,
		"page_number": resp.PageNumber,
		"total_pages": resp.TotalPages,
	})
}

