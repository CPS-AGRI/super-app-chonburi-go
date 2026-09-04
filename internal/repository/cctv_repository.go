package repository

import (
	"fmt"
	"strings"
	"time"

	"super-app-chonburi-go/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type cctvRepository struct {
	db *gorm.DB
}

// NewCCTVRepository creates a new AdminCCTVRepository.
func NewCCTVRepository(db *gorm.DB) domain.AdminCCTVRepository {
	return &cctvRepository{db: db}
}

func (r *cctvRepository) Create(cctv *domain.CCTV) error {
	return r.db.Create(cctv).Error
}

func (r *cctvRepository) GetPaginated(query domain.CCTVQuery) (*domain.PaginatedCCTVResponse, error) {
	var items []domain.CCTV
	var total int64

	db := r.db.Model(&domain.CCTV{})
	if query.Name != "" {
		db = db.Where("name ILIKE ?", "%"+query.Name+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (query.PageNumber - 1) * query.PageSize
	err := db.Preload("Creator").Preload("Deleter").Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&items).Error
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(query.PageSize) - 1) / int64(query.PageSize))
	if totalPages < 1 {
		totalPages = 1
	}

	return &domain.PaginatedCCTVResponse{
		Items:      items,
		TotalItems: total,
		PageNumber: query.PageNumber,
		TotalPages: totalPages,
	}, nil
}

func (r *cctvRepository) GetRequestsPaginated(query domain.CCTVRequestQuery) (*domain.PaginatedCCTVRequestResponse, error) {
	var items []domain.CCTVRequest
	var total int64

	db := r.db.Model(&domain.CCTVRequest{})

	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (query.PageNumber - 1) * query.PageSize
	// Preload CCTV, User, and User.Information to prevent N+1 query loops.
	err := db.Preload("CCTV").Preload("User").Preload("User.Information").Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&items).Error
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(query.PageSize) - 1) / int64(query.PageSize))
	if totalPages < 1 {
		totalPages = 1
	}

	return &domain.PaginatedCCTVRequestResponse{
		Items:      items,
		TotalItems: total,
		PageNumber: query.PageNumber,
		TotalPages: totalPages,
	}, nil
}

func (r *cctvRepository) GetRequestByID(id uuid.UUID) (*domain.CCTVRequest, error) {
	var item domain.CCTVRequest
	err := r.db.Preload("CCTV").Preload("User").Preload("User.Information").Where("id = ?", id).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *cctvRepository) UpdateRequest(req *domain.CCTVRequest) error {
	return r.db.Save(req).Error
}

func (r *cctvRepository) Delete(id uuid.UUID, adminID string) error {
	if err := r.db.Model(&domain.CCTV{}).Where("id = ?", id).Update("deleted_by", adminID).Error; err != nil {
		return err
	}
	return r.db.Where("id = ?", id).Delete(&domain.CCTV{}).Error
}

func (r *cctvRepository) CreateViewLog(logItem *domain.CCTVViewLog) error {
	err := r.db.Omit("CCTV", "Admin", "User").Create(logItem).Error
	if err != nil {
		fmt.Printf("[CCTV Repository CreateViewLog] DB error: %v\n", err)
	}
	return err
}

func (r *cctvRepository) GetRecentLogs(query domain.CCTVLogQuery) (*domain.PaginatedCCTVRecentLogResponse, error) {
	var items []domain.CCTVViewLog
	var total int64

	db := r.db.Model(&domain.CCTVViewLog{})
	if query.CCTVID != nil {
		db = db.Where("cctv_id = ?", *query.CCTVID)
	}
	if query.UserID != nil {
		db = db.Where("user_id = ?", *query.UserID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if query.PageNumber < 1 {
		query.PageNumber = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 15
	}

	offset := (query.PageNumber - 1) * query.PageSize
	err := db.Preload("CCTV").
		Order("started_at DESC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, len(items))
	for i, item := range items {
		userIDs[i] = item.UserID.String()
	}

	var admins []domain.Admin
	if len(userIDs) > 0 {
		r.db.Where("id IN ?", userIDs).Find(&admins)
	}
	adminMap := make(map[string]string)
	for _, a := range admins {
		fullName := strings.TrimSpace(a.Name + " " + a.LastName)
		if fullName == "" {
			fullName = a.Name
		}
		adminMap[a.ID] = fullName
	}

	var users []domain.AppUser
	if len(userIDs) > 0 {
		r.db.Preload("Information").Where("id IN ?", userIDs).Find(&users)
	}
	userMap := make(map[string]string)
	for _, u := range users {
		if u.Information != nil {
			fullName := strings.TrimSpace(u.Information.Name + " " + u.Information.LastName)
			if fullName != "" {
				userMap[u.ID.String()] = fullName
			}
		}
	}

	mappedItems := make([]domain.CCTVRecentLogItem, len(items))
	for i, item := range items {
		uIDStr := item.UserID.String()
		userName := "ผู้ใช้งานทั่วไป"
		if adminName, ok := adminMap[uIDStr]; ok && adminName != "" {
			userName = adminName
		} else if appUserName, ok := userMap[uIDStr]; ok && appUserName != "" {
			userName = appUserName
		}

		camName := "CAM"
		if item.CCTV != nil && item.CCTV.Name != "" {
			camName = item.CCTV.Name
		}

		mins := item.DurationSeconds / 60
		if mins < 1 && item.DurationSeconds > 0 {
			mins = 1
		}
		durationText := fmt.Sprintf("%d นาที", mins)

		device := item.DeviceType
		if device == "" {
			device = "Desktop"
		}

		policy := item.PolicyType
		if policy == "" {
			policy = "AdminPolicy"
		}

		mappedItems[i] = domain.CCTVRecentLogItem{
			ID:              item.ID.String(),
			UserName:        userName,
			PolicyType:      policy,
			CameraName:      camName,
			ViewedTime:      item.StartedAt.Format("15:04:05"),
			DurationText:    durationText,
			DurationSeconds: item.DurationSeconds,
			DeviceType:      device,
			StartedAt:       item.StartedAt.Format(time.RFC3339),
		}
	}

	totalPages := int((total + int64(query.PageSize) - 1) / int64(query.PageSize))
	if totalPages < 1 {
		totalPages = 1
	}

	return &domain.PaginatedCCTVRecentLogResponse{
		Items:      mappedItems,
		TotalItems: total,
		PageNumber: query.PageNumber,
		TotalPages: totalPages,
	}, nil
}

type userSummaryAggregateRow struct {
	UserID               uuid.UUID `gorm:"column:user_id"`
	PolicyType           string    `gorm:"column:policy_type"`
	LastViewedAt         time.Time `gorm:"column:last_viewed_at"`
	TotalDurationSeconds int64     `gorm:"column:total_duration_seconds"`
	TotalSessions        int64     `gorm:"column:total_sessions"`
}

func (r *cctvRepository) GetUserSummaryLogs(query domain.CCTVLogQuery) (*domain.PaginatedCCTVUserSummaryResponse, error) {
	if query.PageNumber < 1 {
		query.PageNumber = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 15
	}

	// 1. Count distinct users
	var totalDistinctUsers int64
	err := r.db.Model(&domain.CCTVViewLog{}).
		Select("COUNT(DISTINCT user_id)").
		Scan(&totalDistinctUsers).Error
	if err != nil {
		return nil, err
	}

	// 2. Fetch aggregated user summary rows
	offset := (query.PageNumber - 1) * query.PageSize
	var rows []userSummaryAggregateRow
	err = r.db.Model(&domain.CCTVViewLog{}).
		Select("user_id, policy_type, MAX(started_at) as last_viewed_at, SUM(duration_seconds) as total_duration_seconds, COUNT(id) as total_sessions").
		Group("user_id, policy_type").
		Order("last_viewed_at DESC").
		Offset(offset).
		Limit(query.PageSize).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return &domain.PaginatedCCTVUserSummaryResponse{
			Items:      []domain.CCTVUserSummaryItem{},
			TotalItems: totalDistinctUsers,
			PageNumber: query.PageNumber,
			TotalPages: 1,
		}, nil
	}

	// 3. Collect user UUIDs to batch query names
	userIDs := make([]string, len(rows))
	for i, row := range rows {
		userIDs[i] = row.UserID.String()
	}

	var admins []domain.Admin
	r.db.Where("id IN ?", userIDs).Find(&admins)
	adminMap := make(map[string]string)
	for _, a := range admins {
		fullName := strings.TrimSpace(a.Name + " " + a.LastName)
		if fullName == "" {
			fullName = a.Name
		}
		adminMap[a.ID] = fullName
	}

	var users []domain.AppUser
	r.db.Preload("Information").Where("id IN ?", userIDs).Find(&users)
	userMap := make(map[string]string)
	for _, u := range users {
		if u.Information != nil {
			fullName := strings.TrimSpace(u.Information.Name + " " + u.Information.LastName)
			if fullName != "" {
				userMap[u.ID.String()] = fullName
			}
		}
	}

	// 4. Map aggregated rows into response items
	mappedItems := make([]domain.CCTVUserSummaryItem, len(rows))
	for i, row := range rows {
		uIDStr := row.UserID.String()
		name := "ผู้ใช้งานทั่วไป"
		if adminName, ok := adminMap[uIDStr]; ok && adminName != "" {
			name = adminName
		} else if appUserName, ok := userMap[uIDStr]; ok && appUserName != "" {
			name = appUserName
		}

		totalMins := int(row.TotalDurationSeconds / 60)
		if totalMins < 1 && row.TotalDurationSeconds > 0 {
			totalMins = 1
		}

		policy := row.PolicyType
		if policy == "" {
			policy = "AdminPolicy"
		}

		mappedItems[i] = domain.CCTVUserSummaryItem{
			UserID:           uIDStr,
			UserName:         name,
			PolicyType:       policy,
			LastViewedText:   formatThaiDateTime(row.LastViewedAt),
			TotalMinutes:     totalMins,
			TotalMinutesText: fmt.Sprintf("%d นาที", totalMins),
			TimeLimitText:    "ไม่จำกัด",
			Status:           "ปกติ",
		}
	}

	totalPages := int((totalDistinctUsers + int64(query.PageSize) - 1) / int64(query.PageSize))
	if totalPages < 1 {
		totalPages = 1
	}

	return &domain.PaginatedCCTVUserSummaryResponse{
		Items:      mappedItems,
		TotalItems: totalDistinctUsers,
		PageNumber: query.PageNumber,
		TotalPages: totalPages,
	}, nil
}

func formatThaiDateTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	thaiMonths := []string{
		"", "ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.",
		"ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค.",
	}
	yearBE := (t.Year() + 543) % 100 // e.g. 2026 + 543 = 2569 -> 69
	month := int(t.Month())
	monthStr := ""
	if month >= 1 && month <= 12 {
		monthStr = thaiMonths[month]
	}
	return fmt.Sprintf("%d %s %02d %02d:%02d", t.Day(), monthStr, yearBE, t.Hour(), t.Minute())
}

