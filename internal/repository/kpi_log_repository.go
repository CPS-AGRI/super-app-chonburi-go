package repository

import (
	"fmt"
	"math"
	"strings"
	"time"

	"super-app-chonburi-go/internal/domain"

	"gorm.io/gorm"
)

type kpiLogRepository struct {
	db *gorm.DB
}

func NewKPILogRepository(db *gorm.DB) domain.KPILogRepository {
	_ = db.AutoMigrate(&domain.KPIThresholdConfig{})
	return &kpiLogRepository{db: db}
}

func (r *kpiLogRepository) Create(log *domain.KPILog) error {
	return r.db.Create(log).Error
}

func (r *kpiLogRepository) CreateBatch(logs []domain.KPILog) error {
	if len(logs) == 0 {
		return nil
	}
	return r.db.Create(&logs).Error
}

func (r *kpiLogRepository) GetKPISummary(filter domain.KPIQueryFilter) (*domain.KPISummaryResponse, error) {
	if filter.ModuleKey == domain.ModuleKeyPublicRelations {
		var totalPublishedNews int64
		pubQuery := r.db.Table("module_public_relations").
			Where("LOWER(status) IN ('published', 'active')")
		if filter.UserID != "" {
			pubQuery = pubQuery.Where("admin_user_id::text = ?", filter.UserID)
		}
		if filter.DepartmentID != "" {
			pubQuery = pubQuery.Where("admin_user_id IN (SELECT admin_id FROM admin_departments WHERE department_id::text = ?)", filter.DepartmentID)
		} else if len(filter.DepartmentIDs) > 0 {
			pubQuery = pubQuery.Where("admin_user_id IN (SELECT admin_id FROM admin_departments WHERE department_id::text IN (?))", filter.DepartmentIDs)
		}
		if filter.StartDate != nil {
			pubQuery = pubQuery.Where("created_date >= ?", filter.StartDate)
		}
		if filter.EndDate != nil {
			pubQuery = pubQuery.Where("created_date <= ?", filter.EndDate)
		}
		_ = pubQuery.Count(&totalPublishedNews).Error

		type viewsResult struct {
			TotalViews int64 `gorm:"column:total_views"`
		}
		var vr viewsResult
		viewQuery := r.db.Table("module_public_relation_visitor_count vc").
			Select("COALESCE(SUM(vc.count), 0) as total_views").
			Joins("JOIN module_public_relations pr ON pr.id = vc.module_public_relation_id")
		if filter.UserID != "" {
			viewQuery = viewQuery.Where("pr.admin_user_id::text = ?", filter.UserID)
		}
		if filter.DepartmentID != "" {
			viewQuery = viewQuery.Where("pr.admin_user_id IN (SELECT admin_id FROM admin_departments WHERE department_id::text = ?)", filter.DepartmentID)
		} else if len(filter.DepartmentIDs) > 0 {
			viewQuery = viewQuery.Where("pr.admin_user_id IN (SELECT admin_id FROM admin_departments WHERE department_id::text IN (?))", filter.DepartmentIDs)
		}
		if filter.StartDate != nil {
			viewQuery = viewQuery.Where("pr.created_date >= ?", filter.StartDate)
		}
		if filter.EndDate != nil {
			viewQuery = viewQuery.Where("pr.created_date <= ?", filter.EndDate)
		}
		_ = viewQuery.Scan(&vr).Error

		var totalLikes int64
		likeQuery := r.db.Table("module_public_relation_likes l").
			Joins("JOIN module_public_relations pr ON pr.id = l.module_public_relation_id")
		if filter.UserID != "" {
			likeQuery = likeQuery.Where("pr.admin_user_id::text = ?", filter.UserID)
		}
		if filter.DepartmentID != "" {
			likeQuery = likeQuery.Where("pr.admin_user_id IN (SELECT admin_id FROM admin_departments WHERE department_id::text = ?)", filter.DepartmentID)
		} else if len(filter.DepartmentIDs) > 0 {
			likeQuery = likeQuery.Where("pr.admin_user_id IN (SELECT admin_id FROM admin_departments WHERE department_id::text IN (?))", filter.DepartmentIDs)
		}
		if filter.StartDate != nil {
			likeQuery = likeQuery.Where("pr.created_date >= ?", filter.StartDate)
		}
		if filter.EndDate != nil {
			likeQuery = likeQuery.Where("pr.created_date <= ?", filter.EndDate)
		}
		_ = likeQuery.Count(&totalLikes).Error

		var totalComments int64
		commentQuery := r.db.Table("module_public_relation_comments c").
			Joins("JOIN module_public_relations pr ON pr.id = c.module_public_relation_id").
			Where("c.status != 'hidden'")
		if filter.UserID != "" {
			commentQuery = commentQuery.Where("pr.admin_user_id::text = ?", filter.UserID)
		}
		if filter.DepartmentID != "" {
			commentQuery = commentQuery.Where("pr.admin_user_id IN (SELECT admin_id FROM admin_departments WHERE department_id::text = ?)", filter.DepartmentID)
		} else if len(filter.DepartmentIDs) > 0 {
			commentQuery = commentQuery.Where("pr.admin_user_id IN (SELECT admin_id FROM admin_departments WHERE department_id::text IN (?))", filter.DepartmentIDs)
		}
		if filter.StartDate != nil {
			commentQuery = commentQuery.Where("pr.created_date >= ?", filter.StartDate)
		}
		if filter.EndDate != nil {
			commentQuery = commentQuery.Where("pr.created_date <= ?", filter.EndDate)
		}
		_ = commentQuery.Count(&totalComments).Error

		var totalBroadcasts int64
		broadcastQuery := r.db.Table("module_public_relation_notifications n").
			Where("LOWER(n.status) IN ('published', 'active', 'sent')")
		if filter.UserID != "" {
			broadcastQuery = broadcastQuery.Where("n.admin_user_id::text = ?", filter.UserID)
		}
		if filter.DepartmentID != "" {
			broadcastQuery = broadcastQuery.Where("n.admin_user_id IN (SELECT admin_id FROM admin_departments WHERE department_id::text = ?)", filter.DepartmentID)
		} else if len(filter.DepartmentIDs) > 0 {
			broadcastQuery = broadcastQuery.Where("n.admin_user_id IN (SELECT admin_id FROM admin_departments WHERE department_id::text IN (?))", filter.DepartmentIDs)
		}
		if filter.StartDate != nil {
			broadcastQuery = broadcastQuery.Where("n.created_date >= ?", filter.StartDate)
		}
		if filter.EndDate != nil {
			broadcastQuery = broadcastQuery.Where("n.created_date <= ?", filter.EndDate)
		}
		_ = broadcastQuery.Count(&totalBroadcasts).Error

		var avgViews float64
		if totalPublishedNews > 0 {
			avgViews = math.Round((float64(vr.TotalViews)/float64(totalPublishedNews))*10) / 10
		}

		var engagementRate float64
		if vr.TotalViews > 0 {
			engagementRate = math.Round((float64(totalLikes+totalComments)/float64(vr.TotalViews))*10000) / 100
		}

		var totalEvents int64
		eventQuery := r.db.Table("kpi_activity_logs").Where("module_key = ?", domain.ModuleKeyPublicRelations)
		if filter.UserID != "" {
			eventQuery = eventQuery.Where("actor_id::text = ?", filter.UserID)
		}
		if filter.DepartmentID != "" {
			eventQuery = eventQuery.Where("department_id::text = ?", filter.DepartmentID)
		} else if len(filter.DepartmentIDs) > 0 {
			eventQuery = eventQuery.Where("department_id::text IN (?)", filter.DepartmentIDs)
		}
		if filter.StartDate != nil {
			eventQuery = eventQuery.Where("created_date >= ?", filter.StartDate)
		}
		if filter.EndDate != nil {
			eventQuery = eventQuery.Where("created_date <= ?", filter.EndDate)
		}
		_ = eventQuery.Count(&totalEvents).Error

		thresholds, _ := r.GetThresholds(domain.ModuleKeyPublicRelations)

		return &domain.KPISummaryResponse{
			TotalEvents:         totalEvents,
			TotalPublishedNews:  totalPublishedNews,
			TotalNewsViews:      vr.TotalViews,
			AvgViewsPerNews:     avgViews,
			TotalNewsLikes:      totalLikes,
			TotalNewsComments:   totalComments,
			TotalPushBroadcasts: totalBroadcasts,
			EngagementRate:      engagementRate,
			Thresholds:          thresholds,
		}, nil
	}

	if filter.ModuleKey == domain.ModuleKeyCCTV {
		type cameraSummaryRaw struct {
			TotalCameras      int64 `gorm:"column:total_cameras"`
			OnlineCameras     int64 `gorm:"column:online_cameras"`
			OfflineCameras    int64 `gorm:"column:offline_cameras"`
			PublicCameraCount int64 `gorm:"column:public_cameras"`
			StaffCameraCount  int64 `gorm:"column:staff_cameras"`
		}
		var camRaw cameraSummaryRaw
		camQuery := r.db.Table("module_cctv").
			Select(`
				COUNT(*) as total_cameras,
				COUNT(CASE WHEN status = 'ONLINE' THEN 1 END) as online_cameras,
				COUNT(CASE WHEN status = 'OFFLINE' THEN 1 END) as offline_cameras,
				COUNT(CASE WHEN access_level = 'PUBLIC' THEN 1 END) as public_cameras,
				COUNT(CASE WHEN access_level = 'STAFF_ONLY' THEN 1 END) as staff_cameras
			`).
			Where("deleted_at IS NULL")

		if filter.UserID != "" {
			camQuery = camQuery.Where("created_by::text = ?", filter.UserID)
		}
		_ = camQuery.Scan(&camRaw).Error

		var uptimeRate float64
		if camRaw.TotalCameras > 0 {
			uptimeRate = math.Round((float64(camRaw.OnlineCameras)/float64(camRaw.TotalCameras))*1000) / 10
		}

		type viewLogsSummaryRaw struct {
			TotalSessions    int64 `gorm:"column:total_sessions"`
			TotalDurationSec int64 `gorm:"column:total_duration_sec"`
			CitizenSessions  int64 `gorm:"column:citizen_sessions"`
			StaffSessions    int64 `gorm:"column:staff_sessions"`
		}
		var logRaw viewLogsSummaryRaw
		logQuery := r.db.Table("module_cctv_view_logs l").
			Select(`
				COUNT(*) as total_sessions,
				COALESCE(SUM(l.duration_seconds), 0) as total_duration_sec,
				COUNT(CASE WHEN l.policy_type = 'UserPolicy' THEN 1 END) as citizen_sessions,
				COUNT(CASE WHEN l.policy_type = 'AdminPolicy' THEN 1 END) as staff_sessions
			`)

		if filter.UserID != "" {
			logQuery = logQuery.Where("l.cctv_id IN (SELECT id FROM module_cctv WHERE created_by::text = ?)", filter.UserID)
		}
		if filter.StartDate != nil {
			logQuery = logQuery.Where("l.started_at >= ?", filter.StartDate)
		}
		if filter.EndDate != nil {
			logQuery = logQuery.Where("l.started_at <= ?", filter.EndDate)
		}
		_ = logQuery.Scan(&logRaw).Error

		var totalWatchMin int64
		if logRaw.TotalDurationSec > 0 {
			totalWatchMin = logRaw.TotalDurationSec / 60
			if totalWatchMin < 1 {
				totalWatchMin = 1
			}
		}

		type topCameraRow struct {
			ID               string `gorm:"column:id"`
			Name             string `gorm:"column:name"`
			Address          string `gorm:"column:address"`
			Status           string `gorm:"column:status"`
			AccessLevel      string `gorm:"column:access_level"`
			TotalViews       int64  `gorm:"column:total_views"`
			TotalDurationSec int64  `gorm:"column:total_duration_sec"`
		}
		var topRows []topCameraRow
		topCamQuery := r.db.Table("module_cctv c").
			Select(`
				c.id::text as id,
				c.name,
				COALESCE(c.address, '') as address,
				c.status,
				c.access_level,
				COUNT(l.id) as total_views,
				COALESCE(SUM(l.duration_seconds), 0) as total_duration_sec
			`).
			Joins("LEFT JOIN module_cctv_view_logs l ON l.cctv_id = c.id").
			Where("c.deleted_at IS NULL")

		if filter.UserID != "" {
			topCamQuery = topCamQuery.Where("c.created_by::text = ?", filter.UserID)
		}
		if filter.StartDate != nil {
			topCamQuery = topCamQuery.Where("l.started_at IS NULL OR l.started_at >= ?", filter.StartDate)
		}
		if filter.EndDate != nil {
			topCamQuery = topCamQuery.Where("l.started_at IS NULL OR l.started_at <= ?", filter.EndDate)
		}
		_ = topCamQuery.
			Group("c.id, c.name, c.address, c.status, c.access_level").
			Order("total_views DESC, total_duration_sec DESC").
			Limit(5).
			Scan(&topRows).Error

		topCameras := make([]domain.KPICCTVTopCamera, len(topRows))
		for i, rItem := range topRows {
			mins := rItem.TotalDurationSec / 60
			topCameras[i] = domain.KPICCTVTopCamera{
				ID:               rItem.ID,
				Name:             rItem.Name,
				Address:          rItem.Address,
				Status:           rItem.Status,
				AccessLevel:      rItem.AccessLevel,
				TotalViews:       rItem.TotalViews,
				TotalDurationMin: mins,
			}
		}

		thresholds, _ := r.GetThresholds(domain.ModuleKeyCCTV)

		return &domain.KPISummaryResponse{
			TotalCameras:          camRaw.TotalCameras,
			OnlineCameras:         camRaw.OnlineCameras,
			OfflineCameras:        camRaw.OfflineCameras,
			CameraUptimeRate:      uptimeRate,
			TotalViewSessions:     logRaw.TotalSessions,
			TotalWatchDurationMin: totalWatchMin,
			CitizenViewCount:      logRaw.CitizenSessions,
			StaffViewCount:        logRaw.StaffSessions,
			PublicCameraCount:     camRaw.PublicCameraCount,
			StaffCameraCount:      camRaw.StaffCameraCount,
			TopCameras:            topCameras,
			Thresholds:            thresholds,
		}, nil
	}

	if filter.ModuleKey == domain.ModuleKeyUserVerify {
		type verifySummaryRaw struct {
			TotalRequests  int64   `gorm:"column:total_requests"`
			ApprovedCount  int64   `gorm:"column:approved_count"`
			RejectedCount  int64   `gorm:"column:rejected_count"`
			PendingCount   int64   `gorm:"column:pending_count"`
			AvgDurationSec float64 `gorm:"column:avg_duration_sec"`
		}
		var raw verifySummaryRaw
		q := r.db.Table("user_informations").
			Select(`
				COUNT(*) as total_requests,
				COUNT(CASE WHEN verification_status = 'verified' THEN 1 END) as approved_count,
				COUNT(CASE WHEN verification_status = 'rejected' THEN 1 END) as rejected_count,
				COUNT(CASE WHEN verification_status = 'pending' THEN 1 END) as pending_count,
				COALESCE(AVG(CASE 
					WHEN verification_status IN ('verified', 'rejected') 
					     AND COALESCE(verified_date, updated_date) > created_date 
					THEN EXTRACT(EPOCH FROM (COALESCE(verified_date, updated_date) - created_date)) 
				END), 0) as avg_duration_sec
			`)

		if filter.UserID != "" {
			q = q.Where("updated_by = ?", filter.UserID)
		}
		if filter.DepartmentID != "" {
			q = q.Where("updated_by IN (SELECT admin_id::text FROM admin_departments WHERE department_id::text = ?)", filter.DepartmentID)
		} else if len(filter.DepartmentIDs) > 0 {
			q = q.Where("updated_by IN (SELECT admin_id::text FROM admin_departments WHERE department_id::text IN (?))", filter.DepartmentIDs)
		}
		if filter.StartDate != nil {
			q = q.Where("created_date >= ?", filter.StartDate)
		}
		if filter.EndDate != nil {
			q = q.Where("created_date <= ?", filter.EndDate)
		}
		_ = q.Scan(&raw).Error

		var approvalRate float64
		reviewedCount := raw.ApprovedCount + raw.RejectedCount
		if reviewedCount > 0 {
			approvalRate = math.Round((float64(raw.ApprovedCount)/float64(reviewedCount))*1000) / 10
		}

		avgDurationMin := math.Round((raw.AvgDurationSec/60.0)*10) / 10

		thresholds, _ := r.GetThresholds(domain.ModuleKeyUserVerify)

		return &domain.KPISummaryResponse{
			TotalVerificationRequests:  raw.TotalRequests,
			ApprovedVerifications:      raw.ApprovedCount,
			RejectedVerifications:      raw.RejectedCount,
			PendingVerifications:       raw.PendingCount,
			VerificationApprovalRate:   approvalRate,
			AvgVerificationDurationMin: avgDurationMin,
			TotalReviewedCases:         reviewedCount,
			Thresholds:                 thresholds,
		}, nil
	}

	query := r.db.Table("kpi_activity_logs")

	if filter.ModuleKey != "" {
		query = query.Where("module_key = ?", filter.ModuleKey)
	}
	if filter.DepartmentID != "" {
		query = query.Where("department_id = ?", filter.DepartmentID)
	} else if len(filter.DepartmentIDs) > 0 {
		query = query.Where("department_id IN (?)", filter.DepartmentIDs)
	}
	if filter.UserID != "" {
		query = query.Where("(target_user_id = ? OR actor_id = ?)", filter.UserID, filter.UserID)
	}
	if filter.StartDate != nil {
		query = query.Where("created_date >= ?", filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_date <= ?", filter.EndDate)
	}

	type summaryRaw struct {
		TotalEvents           int64   `gorm:"column:total_events"`
		TotalResolvedCases    int64   `gorm:"column:total_resolved_cases"`
		TotalDispatched       int64   `gorm:"column:total_dispatched"`
		AvgMTTRHours          float64 `gorm:"column:avg_mttr_hours"`
		AvgDispatchHours      float64 `gorm:"column:avg_dispatch_hours"`
		TotalSLABreachedCases int64   `gorm:"column:total_sla_breached_cases"`
		SLAComplianceRate     float64 `gorm:"column:sla_compliance_rate"`
		AvgCitizenRating      float64 `gorm:"column:avg_citizen_rating"`
		TotalDisputes         int64   `gorm:"column:total_disputes"`
	}

	var raw summaryRaw

	thresholds, _ := r.GetThresholds(filter.ModuleKey)
	resSeconds := int64(thresholds.ResolutionSLAHours * 3600)

	selectClause := fmt.Sprintf(`
		COUNT(*) as total_events,
		COUNT(CASE WHEN event_code = 'COMPLAINT_STAFF_RESOLVED' THEN 1 END) as total_resolved_cases,
		COUNT(DISTINCT CASE WHEN event_code IN ('COMPLAINT_CENTER_FORWARDED', 'COMPLAINT_MANAGER_ASSIGNED') THEN entity_id END) as total_dispatched,
		COALESCE(AVG(CASE WHEN event_code = 'COMPLAINT_STAFF_RESOLVED' AND step_duration_seconds > 0 THEN step_duration_seconds END) / 3600.0, 0) as avg_mttr_hours,
		COALESCE(AVG(CASE WHEN event_code IN ('COMPLAINT_CENTER_FORWARDED', 'COMPLAINT_MANAGER_ASSIGNED') AND step_duration_seconds > 0 THEN step_duration_seconds END) / 3600.0, 0) as avg_dispatch_hours,
		COUNT(CASE WHEN event_code = 'COMPLAINT_STAFF_RESOLVED' AND (is_sla_breached = true OR step_duration_seconds > %d) THEN 1 END) as total_sla_breached_cases,
		CASE 
			WHEN COUNT(CASE WHEN event_code = 'COMPLAINT_STAFF_RESOLVED' THEN 1 END) > 0 
			THEN ROUND((1.0 - (COUNT(CASE WHEN event_code = 'COMPLAINT_STAFF_RESOLVED' AND (is_sla_breached = true OR step_duration_seconds > %d) THEN 1 END)::numeric / COUNT(CASE WHEN event_code = 'COMPLAINT_STAFF_RESOLVED' THEN 1 END)::numeric)) * 100.0, 2)
			ELSE 100.0 
		END as sla_compliance_rate,
		COALESCE(AVG(CASE WHEN event_code = 'COMPLAINT_CITIZEN_RATED' AND (metadata->>'rating') ~ '^[0-9]+$' THEN (metadata->>'rating')::numeric END), 0) as avg_citizen_rating,
		COUNT(CASE WHEN event_code = 'COMPLAINT_CITIZEN_DISPUTED' THEN 1 END) as total_disputes
	`, resSeconds, resSeconds)

	if err := query.Select(selectClause).Scan(&raw).Error; err != nil {
		return nil, err
	}

	// Active stalled open complaints in module_complaints that have exceeded resolution SLA
	activeBreachedCount := int64(0)
	if filter.ModuleKey == "" || filter.ModuleKey == domain.ModuleKeyComplaint {
		activeQuery := r.db.Table("module_complaints").
			Where("status NOT IN ('completed', 'rejected')").
			Where("EXTRACT(EPOCH FROM (NOW() - created_date)) > ?", resSeconds)
		if filter.DepartmentID != "" {
			activeQuery = activeQuery.Where("department_id = ?", filter.DepartmentID)
		} else if len(filter.DepartmentIDs) > 0 {
			activeQuery = activeQuery.Where("department_id IN (?)", filter.DepartmentIDs)
		}
		if filter.StartDate != nil {
			activeQuery = activeQuery.Where("created_date >= ?", filter.StartDate)
		}
		if filter.EndDate != nil {
			activeQuery = activeQuery.Where("created_date <= ?", filter.EndDate)
		}
		_ = activeQuery.Count(&activeBreachedCount).Error
	}

	// Synchronize Citizen Rating & Disputes from module_complaint_rating_histories
	if filter.ModuleKey == "" || filter.ModuleKey == domain.ModuleKeyComplaint {
		type ratingStats struct {
			AvgRating     float64 `gorm:"column:avg_rating"`
			TotalDisputes int64   `gorm:"column:total_disputes"`
		}
		var rStats ratingStats
		ratingQuery := r.db.Table("module_complaint_rating_histories h").
			Select("COALESCE(ROUND(AVG(h.rating_score)::numeric, 2), 0) as avg_rating, COUNT(CASE WHEN h.is_disputed THEN 1 END) as total_disputes")

		if filter.DepartmentID != "" {
			ratingQuery = ratingQuery.Where("h.department_id::text = ?", filter.DepartmentID)
		} else if len(filter.DepartmentIDs) > 0 {
			ratingQuery = ratingQuery.Where("h.department_id::text IN (?)", filter.DepartmentIDs)
		}
		if filter.UserID != "" {
			ratingQuery = ratingQuery.Where("h.assignee_id::text = ?", filter.UserID)
		}
		if filter.StartDate != nil {
			ratingQuery = ratingQuery.Where("h.created_date >= ?", filter.StartDate)
		}
		if filter.EndDate != nil {
			ratingQuery = ratingQuery.Where("h.created_date <= ?", filter.EndDate)
		}

		if err := ratingQuery.Scan(&rStats).Error; err == nil && (rStats.AvgRating > 0 || rStats.TotalDisputes > 0) {
			raw.AvgCitizenRating = rStats.AvgRating
			raw.TotalDisputes = rStats.TotalDisputes
		}
	}

	totalBreached := raw.TotalSLABreachedCases + activeBreachedCount
	totalRelevantCases := raw.TotalResolvedCases + activeBreachedCount
	slaComplianceRate := raw.SLAComplianceRate
	if totalRelevantCases > 0 {
		compliantCases := raw.TotalResolvedCases - raw.TotalSLABreachedCases
		if compliantCases < 0 {
			compliantCases = 0
		}
		slaComplianceRate = math.Round((float64(compliantCases)/float64(totalRelevantCases))*1000) / 10.0
	}

	res := &domain.KPISummaryResponse{
		TotalEvents:           raw.TotalEvents,
		TotalResolvedCases:    raw.TotalResolvedCases,
		TotalDispatched:       raw.TotalDispatched,
		AvgMTTRHours:          raw.AvgMTTRHours,
		AvgMTTRDays:           raw.AvgMTTRHours / 24.0,
		AvgDispatchHours:      raw.AvgDispatchHours,
		SLAComplianceRate:     slaComplianceRate,
		TotalSLABreachedCases: totalBreached,
		AvgCitizenRating:      raw.AvgCitizenRating,
		TotalDisputes:         raw.TotalDisputes,
		Thresholds:            thresholds,
	}

	return res, nil
}

func (r *kpiLogRepository) GetStaffLeaderboard(filter domain.KPIQueryFilter) ([]domain.KPIStaffLeaderboardItem, error) {
	if filter.ModuleKey == domain.ModuleKeyPublicRelations {
		var prWhereClauses []string
		var prArgs []interface{}
		var notifWhereClauses []string
		var notifArgs []interface{}

		if filter.StartDate != nil {
			prWhereClauses = append(prWhereClauses, "pr.created_date >= ?")
			prArgs = append(prArgs, filter.StartDate)
			notifWhereClauses = append(notifWhereClauses, "n.created_date >= ?")
			notifArgs = append(notifArgs, filter.StartDate)
		}
		if filter.EndDate != nil {
			prWhereClauses = append(prWhereClauses, "pr.created_date <= ?")
			prArgs = append(prArgs, filter.EndDate)
			notifWhereClauses = append(notifWhereClauses, "n.created_date <= ?")
			notifArgs = append(notifArgs, filter.EndDate)
		}

		prDateCond := "1=1"
		if len(prWhereClauses) > 0 {
			prDateCond = strings.Join(prWhereClauses, " AND ")
		}

		notifDateCond := "1=1"
		if len(notifWhereClauses) > 0 {
			notifDateCond = strings.Join(notifWhereClauses, " AND ")
		}

		adminFilterSQL := ""
		var adminFilterArgs []interface{}
		if filter.UserID != "" {
			adminFilterSQL += " AND a.id::text = ?"
			adminFilterArgs = append(adminFilterArgs, filter.UserID)
		}
		if filter.DepartmentID != "" {
			adminFilterSQL += " AND EXISTS (SELECT 1 FROM admin_departments ad WHERE ad.admin_id = a.id AND ad.department_id::text = ?)"
			adminFilterArgs = append(adminFilterArgs, filter.DepartmentID)
		} else if len(filter.DepartmentIDs) > 0 {
			adminFilterSQL += " AND EXISTS (SELECT 1 FROM admin_departments ad WHERE ad.admin_id = a.id AND ad.department_id::text IN (?))"
			adminFilterArgs = append(adminFilterArgs, filter.DepartmentIDs)
		}

		sql := fmt.Sprintf(`
			WITH pr_stats AS (
				SELECT 
					pr.admin_user_id,
					COUNT(pr.id) as published_count,
					COALESCE(SUM(vc.count), 0) as total_views,
					COALESCE(SUM(lk.like_count), 0) as total_likes,
					COALESCE(SUM(cm.comment_count), 0) as total_comments
				FROM module_public_relations pr
				LEFT JOIN module_public_relation_visitor_count vc ON vc.module_public_relation_id = pr.id
				LEFT JOIN (
					SELECT module_public_relation_id, COUNT(*) as like_count 
					FROM module_public_relation_likes 
					GROUP BY module_public_relation_id
				) lk ON lk.module_public_relation_id = pr.id
				LEFT JOIN (
					SELECT module_public_relation_id, COUNT(*) as comment_count 
					FROM module_public_relation_comments 
					WHERE status != 'hidden' 
					GROUP BY module_public_relation_id
				) cm ON cm.module_public_relation_id = pr.id
				WHERE LOWER(pr.status) IN ('published', 'active')
				  AND %s
				GROUP BY pr.admin_user_id
			),
			notif_stats AS (
				SELECT 
					n.admin_user_id,
					COUNT(n.id) as broadcast_count
				FROM module_public_relation_notifications n
				WHERE LOWER(n.status) IN ('published', 'active', 'sent')
				  AND %s
				GROUP BY n.admin_user_id
			)
			SELECT 
				a.id::text as user_id,
				COALESCE(a.name, 'เจ้าหน้าที่') as name,
				COALESCE(a.last_name, '') as last_name,
				COALESCE(ar.type, '') as role_type,
				COALESCE(a.position, '') as position,
				COALESCE((
					SELECT d2.name 
					FROM admin_departments ad 
					JOIN departments d2 ON d2.id = ad.department_id 
					WHERE ad.admin_id = a.id 
					LIMIT 1
				), '-') as department_name,
				COALESCE(ps.published_count, 0) as published_news_count,
				COALESCE(ps.total_views, 0) as total_views_generated,
				COALESCE(ps.total_likes, 0) as total_likes_received,
				COALESCE(ps.total_comments, 0) as total_comments_received,
				COALESCE(ns.broadcast_count, 0) as push_broadcast_count
			FROM admin_users a
			LEFT JOIN admin_roles ar ON ar.id = a.role_id
			LEFT JOIN pr_stats ps ON ps.admin_user_id = a.id
			LEFT JOIN notif_stats ns ON ns.admin_user_id = a.id
			WHERE (COALESCE(ps.published_count, 0) > 0 OR COALESCE(ns.broadcast_count, 0) > 0)
			%s
			ORDER BY published_news_count DESC, total_views_generated DESC
			LIMIT 50
		`, prDateCond, notifDateCond, adminFilterSQL)

		allArgs := append(prArgs, notifArgs...)
		allArgs = append(allArgs, adminFilterArgs...)

		var items []domain.KPIStaffLeaderboardItem
		if err := r.db.Raw(sql, allArgs...).Scan(&items).Error; err != nil {
			return nil, err
		}
		if items == nil {
			items = []domain.KPIStaffLeaderboardItem{}
		}
		return items, nil
	}

	if filter.ModuleKey == domain.ModuleKeyCCTV {
		var logDateCond string
		var logArgs []interface{}
		if filter.StartDate != nil {
			logDateCond += " AND started_at >= ?"
			logArgs = append(logArgs, filter.StartDate)
		}
		if filter.EndDate != nil {
			logDateCond += " AND started_at <= ?"
			logArgs = append(logArgs, filter.EndDate)
		}

		adminFilterSQL := ""
		var adminFilterArgs []interface{}
		if filter.UserID != "" {
			adminFilterSQL += " AND a.id::text = ?"
			adminFilterArgs = append(adminFilterArgs, filter.UserID)
		}
		if filter.DepartmentID != "" {
			adminFilterSQL += " AND EXISTS (SELECT 1 FROM admin_departments ad WHERE ad.admin_id = a.id AND ad.department_id::text = ?)"
			adminFilterArgs = append(adminFilterArgs, filter.DepartmentID)
		} else if len(filter.DepartmentIDs) > 0 {
			adminFilterSQL += " AND EXISTS (SELECT 1 FROM admin_departments ad WHERE ad.admin_id = a.id AND ad.department_id::text IN (?))"
			adminFilterArgs = append(adminFilterArgs, filter.DepartmentIDs)
		}

		sql := fmt.Sprintf(`
			WITH camera_stats AS (
				SELECT 
					created_by::text as admin_id,
					COUNT(id) as total_cameras,
					COUNT(CASE WHEN status = 'ONLINE' THEN 1 END) as online_cameras,
					COUNT(CASE WHEN status = 'OFFLINE' THEN 1 END) as offline_cameras
				FROM module_cctv
				WHERE deleted_at IS NULL AND created_by IS NOT NULL
				GROUP BY created_by
			),
			surveillance_stats AS (
				SELECT 
					user_id::text as admin_id,
					COUNT(id) as total_sessions,
					COALESCE(SUM(duration_seconds) / 60, 0) as total_duration_min
				FROM module_cctv_view_logs
				WHERE policy_type = 'AdminPolicy' %s
				GROUP BY user_id
			)
			SELECT 
				a.id::text as user_id,
				COALESCE(a.name, 'เจ้าหน้าที่') as name,
				COALESCE(a.last_name, '') as last_name,
				COALESCE(ar.type, '') as role_type,
				COALESCE(a.position, '') as position,
				COALESCE((
					SELECT d2.name 
					FROM admin_departments ad 
					JOIN departments d2 ON d2.id = ad.department_id 
					WHERE ad.admin_id = a.id 
					LIMIT 1
				), '-') as department_name,
				COALESCE(cs.total_cameras, 0) as managed_cameras_count,
				COALESCE(cs.online_cameras, 0) as online_cameras_count,
				CASE 
					WHEN COALESCE(cs.total_cameras, 0) > 0 
					THEN ROUND((COALESCE(cs.online_cameras, 0)::numeric / cs.total_cameras::numeric) * 100.0, 1)
					ELSE 100.0 
				END as camera_health_rate,
				COALESCE(ss.total_sessions, 0) as surveillance_sessions,
				COALESCE(ss.total_duration_min, 0) as surveillance_duration_min
			FROM admin_users a
			LEFT JOIN admin_roles ar ON ar.id = a.role_id
			LEFT JOIN camera_stats cs ON cs.admin_id = a.id::text
			LEFT JOIN surveillance_stats ss ON ss.admin_id = a.id::text
			WHERE (COALESCE(cs.total_cameras, 0) > 0 OR COALESCE(ss.total_sessions, 0) > 0)
			%s
			ORDER BY managed_cameras_count DESC, surveillance_sessions DESC
			LIMIT 50
		`, logDateCond, adminFilterSQL)

		allArgs := append(logArgs, adminFilterArgs...)
		var items []domain.KPIStaffLeaderboardItem
		if err := r.db.Raw(sql, allArgs...).Scan(&items).Error; err != nil {
			return nil, err
		}
		if items == nil {
			items = []domain.KPIStaffLeaderboardItem{}
		}
		return items, nil
	}

	if filter.ModuleKey == domain.ModuleKeyUserVerify {
		var verifyDateCond string
		var verifyArgs []interface{}
		if filter.StartDate != nil {
			verifyDateCond += " AND (created_date >= ?)"
			verifyArgs = append(verifyArgs, filter.StartDate)
		}
		if filter.EndDate != nil {
			verifyDateCond += " AND (created_date <= ?)"
			verifyArgs = append(verifyArgs, filter.EndDate)
		}

		adminFilterSQL := ""
		var adminFilterArgs []interface{}
		if filter.UserID != "" {
			adminFilterSQL += " AND a.id::text = ?"
			adminFilterArgs = append(adminFilterArgs, filter.UserID)
		}
		if filter.DepartmentID != "" {
			adminFilterSQL += " AND EXISTS (SELECT 1 FROM admin_departments ad WHERE ad.admin_id = a.id AND ad.department_id::text = ?)"
			adminFilterArgs = append(adminFilterArgs, filter.DepartmentID)
		} else if len(filter.DepartmentIDs) > 0 {
			adminFilterSQL += " AND EXISTS (SELECT 1 FROM admin_departments ad WHERE ad.admin_id = a.id AND ad.department_id::text IN (?))"
			adminFilterArgs = append(adminFilterArgs, filter.DepartmentIDs)
		}

		sql := fmt.Sprintf(`
			WITH verify_stats AS (
				SELECT 
					updated_by as admin_id,
					COUNT(CASE WHEN verification_status = 'verified' THEN 1 END) as verified_count,
					COUNT(CASE WHEN verification_status = 'rejected' THEN 1 END) as rejected_count,
					COUNT(CASE WHEN verification_status IN ('verified', 'rejected') THEN 1 END) as total_audited,
					COALESCE(AVG(CASE 
						WHEN verification_status IN ('verified', 'rejected') 
						     AND COALESCE(verified_date, updated_date) > created_date 
						THEN EXTRACT(EPOCH FROM (COALESCE(verified_date, updated_date) - created_date)) / 60.0 
					END), 0) as avg_processing_min
				FROM user_informations
				WHERE updated_by IS NOT NULL AND updated_by != '' %s
				GROUP BY updated_by
			)
			SELECT 
				a.id::text as user_id,
				COALESCE(a.name, 'เจ้าหน้าที่') as name,
				COALESCE(a.last_name, '') as last_name,
				COALESCE(ar.type, '') as role_type,
				COALESCE(a.position, '') as position,
				COALESCE((
					SELECT d2.name 
					FROM admin_departments ad 
					JOIN departments d2 ON d2.id = ad.department_id 
					WHERE ad.admin_id = a.id 
					LIMIT 1
				), '-') as department_name,
				COALESCE(vs.verified_count, 0) as verified_count,
				COALESCE(vs.rejected_count, 0) as rejected_count,
				COALESCE(vs.total_audited, 0) as total_audited_count,
				CASE 
					WHEN COALESCE(vs.total_audited, 0) > 0 
					THEN ROUND((COALESCE(vs.verified_count, 0)::numeric / vs.total_audited::numeric) * 100.0, 1)
					ELSE 0 
				END as approval_percentage,
				ROUND(COALESCE(vs.avg_processing_min, 0)::numeric, 1) as avg_processing_min
			FROM admin_users a
			LEFT JOIN admin_roles ar ON ar.id = a.role_id
			JOIN verify_stats vs ON vs.admin_id = a.id::text
			WHERE vs.total_audited > 0
			%s
			ORDER BY total_audited_count DESC, approval_percentage DESC
			LIMIT 50
		`, verifyDateCond, adminFilterSQL)

		allArgs := append(verifyArgs, adminFilterArgs...)

		var items []domain.KPIStaffLeaderboardItem
		if err := r.db.Raw(sql, allArgs...).Scan(&items).Error; err != nil {
			return nil, err
		}
		if items == nil {
			items = []domain.KPIStaffLeaderboardItem{}
		}
		return items, nil
	}

	whereSQL := "WHERE k.target_user_id IS NOT NULL AND COALESCE(ar.type, '') NOT IN ('SuperAdmin', 'Executives') AND COALESCE(a.position, '') NOT ILIKE '%superadmin%' AND COALESCE(a.position, '') NOT ILIKE '%ผู้บริหาร%'"
	args := []interface{}{}

	if filter.ModuleKey != "" {
		whereSQL += " AND k.module_key = ?"
		args = append(args, filter.ModuleKey)
	}
	if filter.DepartmentID != "" {
		whereSQL += " AND k.department_id = ?"
		args = append(args, filter.DepartmentID)
	} else if len(filter.DepartmentIDs) > 0 {
		whereSQL += " AND k.department_id IN (?)"
		args = append(args, filter.DepartmentIDs)
	}
	if filter.UserID != "" {
		whereSQL += " AND k.target_user_id = ?"
		args = append(args, filter.UserID)
	}
	if filter.StartDate != nil {
		whereSQL += " AND k.created_date >= ?"
		args = append(args, filter.StartDate)
	}
	if filter.EndDate != nil {
		whereSQL += " AND k.created_date <= ?"
		args = append(args, filter.EndDate)
	}

	thresholds, _ := r.GetThresholds(filter.ModuleKey)
	resSeconds := int64(thresholds.ResolutionSLAHours * 3600)

	sql := fmt.Sprintf(`
		SELECT 
			k.target_user_id::text as user_id,
			COALESCE(a.name, 'เจ้าหน้าที่') as name,
			COALESCE(a.last_name, '') as last_name,
			COALESCE(ar.type, '') as role_type,
			COALESCE(a.position, '') as position,
			COALESCE((
				SELECT d2.name 
				FROM admin_departments ad 
				JOIN departments d2 ON d2.id = ad.department_id 
				WHERE ad.admin_id = k.target_user_id 
				LIMIT 1
			), MAX(d.name), '-') as department_name,
			COUNT(DISTINCT CASE WHEN k.event_code IN ('COMPLAINT_MANAGER_ASSIGNED', 'COMPLAINT_STAFF_ACCEPTED', 'COMPLAINT_STAFF_RESOLVED') THEN k.entity_id END) as total_assigned,
			COUNT(DISTINCT CASE WHEN k.event_code = 'COMPLAINT_STAFF_RESOLVED' THEN k.entity_id END) as total_resolved,
			COALESCE(ROUND((AVG(CASE WHEN k.event_code = 'COMPLAINT_STAFF_RESOLVED' AND k.step_duration_seconds > 0 THEN k.step_duration_seconds END) / 3600.0)::numeric, 1), 0) as avg_mttr_hours,
			CASE 
				WHEN COUNT(CASE WHEN k.event_code = 'COMPLAINT_STAFF_RESOLVED' THEN 1 END) > 0 
				THEN ROUND((1.0 - (COUNT(CASE WHEN k.event_code = 'COMPLAINT_STAFF_RESOLVED' AND (k.is_sla_breached = true OR k.step_duration_seconds > %d) THEN 1 END)::numeric / COUNT(CASE WHEN k.event_code = 'COMPLAINT_STAFF_RESOLVED' THEN 1 END)::numeric)) * 100.0, 1)
				ELSE 100.0 
			END as sla_compliance_rate,
			COALESCE(ROUND(AVG(CASE WHEN k.event_code = 'COMPLAINT_CITIZEN_RATED' AND (k.metadata->>'rating') ~ '^[0-9]+$' THEN (k.metadata->>'rating')::numeric END), 2), 0) as avg_rating,
			COUNT(CASE WHEN k.event_code = 'COMPLAINT_CITIZEN_DISPUTED' THEN 1 END) as dispute_count
		FROM kpi_activity_logs k
		LEFT JOIN admin_users a ON a.id = k.target_user_id
		LEFT JOIN admin_roles ar ON ar.id = a.role_id
		LEFT JOIN departments d ON d.id = k.department_id
		%s
		GROUP BY k.target_user_id, a.name, a.last_name, ar.type, a.position
		ORDER BY total_resolved DESC, avg_mttr_hours ASC
		LIMIT 50
	`, resSeconds, whereSQL)

	var items []domain.KPIStaffLeaderboardItem
	if err := r.db.Raw(sql, args...).Scan(&items).Error; err != nil {
		return nil, err
	}

	// Synchronize rating and dispute metrics from module_complaint_rating_histories for each staff member
	if len(items) > 0 && (filter.ModuleKey == "" || filter.ModuleKey == domain.ModuleKeyComplaint) {
		type staffRatingRow struct {
			AssigneeID   string  `gorm:"column:assignee_id"`
			AvgRating    float64 `gorm:"column:avg_rating"`
			DisputeCount int64   `gorm:"column:dispute_count"`
		}
		var ratingRows []staffRatingRow
		rhQuery := r.db.Table("module_complaint_rating_histories").
			Select("assignee_id, COALESCE(ROUND(AVG(rating_score)::numeric, 2), 0) as avg_rating, COUNT(CASE WHEN is_disputed THEN 1 END) as dispute_count").
			Where("assignee_id IS NOT NULL")

		if filter.DepartmentID != "" {
			rhQuery = rhQuery.Where("department_id::text = ?", filter.DepartmentID)
		} else if len(filter.DepartmentIDs) > 0 {
			rhQuery = rhQuery.Where("department_id::text IN (?)", filter.DepartmentIDs)
		}
		if filter.StartDate != nil {
			rhQuery = rhQuery.Where("created_date >= ?", filter.StartDate)
		}
		if filter.EndDate != nil {
			rhQuery = rhQuery.Where("created_date <= ?", filter.EndDate)
		}

		rhQuery = rhQuery.Group("assignee_id")
		if err := rhQuery.Scan(&ratingRows).Error; err == nil {
			ratingMap := make(map[string]staffRatingRow)
			for _, rRow := range ratingRows {
				ratingMap[rRow.AssigneeID] = rRow
			}
			for i := range items {
				if rData, exists := ratingMap[items[i].UserID]; exists {
					if rData.AvgRating > 0 {
						items[i].AvgRating = rData.AvgRating
					}
					items[i].DisputeCount = rData.DisputeCount
				}
			}
		}
	}

	return items, nil
}

func (r *kpiLogRepository) GetManagerScorecard(filter domain.KPIQueryFilter) ([]domain.KPIManagerScorecardItem, error) {
	// Build ON clauses for kpi_activity_logs so that inactive supervisors are NOT excluded if they haven't dispatched
	kpiJoinConditions := "AND k.event_code IN ('COMPLAINT_CENTER_FORWARDED', 'COMPLAINT_MANAGER_ASSIGNED', 'COMPLAINT_DEPT_RETURNED')"
	kpiArgs := []interface{}{}

	if filter.ModuleKey != "" {
		kpiJoinConditions += " AND k.module_key = ?"
		kpiArgs = append(kpiArgs, filter.ModuleKey)
	}
	if filter.DepartmentID != "" {
		kpiJoinConditions += " AND k.department_id = ?"
		kpiArgs = append(kpiArgs, filter.DepartmentID)
	} else if len(filter.DepartmentIDs) > 0 {
		kpiJoinConditions += " AND k.department_id IN (?)"
		kpiArgs = append(kpiArgs, filter.DepartmentIDs)
	}
	if filter.StartDate != nil {
		kpiJoinConditions += " AND k.created_date >= ?"
		kpiArgs = append(kpiArgs, filter.StartDate)
	}
	if filter.EndDate != nil {
		kpiJoinConditions += " AND k.created_date <= ?"
		kpiArgs = append(kpiArgs, filter.EndDate)
	}

	// Build WHERE clause for admin_users: Exclude SuperAdmin and Executives, include Managers
	whereSQL := "WHERE (ar.type = 'Managers' OR (ar.type IS NULL AND a.position ILIKE '%หัวหน้า%')) AND COALESCE(ar.type, '') NOT IN ('SuperAdmin', 'Executives') AND COALESCE(a.position, '') NOT ILIKE '%superadmin%' AND COALESCE(a.position, '') NOT ILIKE '%ผู้บริหาร%'"
	whereArgs := []interface{}{}

	if filter.UserID != "" {
		whereSQL += " AND a.id = ?"
		whereArgs = append(whereArgs, filter.UserID)
	}
	if filter.DepartmentID != "" {
		whereSQL += " AND (EXISTS (SELECT 1 FROM admin_departments ad_dept WHERE ad_dept.admin_id = a.id AND ad_dept.department_id = ?) OR EXISTS (SELECT 1 FROM kpi_activity_logs k2 WHERE k2.actor_id = a.id AND k2.department_id = ?))"
		whereArgs = append(whereArgs, filter.DepartmentID, filter.DepartmentID)
	} else if len(filter.DepartmentIDs) > 0 {
		whereSQL += " AND (EXISTS (SELECT 1 FROM admin_departments ad_dept WHERE ad_dept.admin_id = a.id AND ad_dept.department_id IN (?)) OR EXISTS (SELECT 1 FROM kpi_activity_logs k2 WHERE k2.actor_id = a.id AND k2.department_id IN (?)))"
		whereArgs = append(whereArgs, filter.DepartmentIDs, filter.DepartmentIDs)
	}

	sql := fmt.Sprintf(`
		SELECT 
			a.id::text as manager_id,
			COALESCE(a.name, 'หัวหน้างาน') as name,
			COALESCE(a.last_name, '') as last_name,
			COALESCE(ar.type, 'Managers') as role_type,
			COALESCE(a.position, 'หัวหน้างาน') as position,
			COALESCE((
				SELECT d2.name 
				FROM admin_departments ad 
				JOIN departments d2 ON d2.id = ad.department_id 
				WHERE ad.admin_id = a.id 
				LIMIT 1
			), '-') as department_name,
			COUNT(DISTINCT CASE WHEN k.event_code IN ('COMPLAINT_CENTER_FORWARDED', 'COMPLAINT_MANAGER_ASSIGNED') THEN k.entity_id END) as total_dispatched,
			COUNT(DISTINCT CASE WHEN k.event_code IN ('COMPLAINT_CENTER_FORWARDED', 'COMPLAINT_MANAGER_ASSIGNED') AND EXISTS (
				SELECT 1 FROM module_complaints mc WHERE mc.id = k.entity_id AND mc.status = 'completed'
			) THEN k.entity_id END) as total_resolved_from_dispatched,
			CASE 
				WHEN COUNT(DISTINCT CASE WHEN k.event_code IN ('COMPLAINT_CENTER_FORWARDED', 'COMPLAINT_MANAGER_ASSIGNED') THEN k.entity_id END) > 0
				THEN ROUND((COUNT(DISTINCT CASE WHEN k.event_code IN ('COMPLAINT_CENTER_FORWARDED', 'COMPLAINT_MANAGER_ASSIGNED') AND EXISTS (
					SELECT 1 FROM module_complaints mc WHERE mc.id = k.entity_id AND mc.status = 'completed'
				) THEN k.entity_id END)::numeric / COUNT(DISTINCT CASE WHEN k.event_code IN ('COMPLAINT_CENTER_FORWARDED', 'COMPLAINT_MANAGER_ASSIGNED') THEN k.entity_id END)::numeric) * 100.0, 1)
				ELSE 0.0
			END as completion_rate,
			COALESCE(ROUND((AVG(CASE WHEN k.event_code IN ('COMPLAINT_CENTER_FORWARDED', 'COMPLAINT_MANAGER_ASSIGNED') AND k.step_duration_seconds > 0 THEN k.step_duration_seconds END) / 3600.0)::numeric, 1), 0) as avg_dispatch_hours,
			0 as unassigned_count,
			COUNT(CASE WHEN k.event_code = 'COMPLAINT_DEPT_RETURNED' THEN 1 END) as rejected_return_count
		FROM admin_users a
		LEFT JOIN admin_roles ar ON ar.id = a.role_id
		LEFT JOIN kpi_activity_logs k ON k.actor_id = a.id %s
		%s
		GROUP BY a.id, a.name, a.last_name, ar.type, a.position
		ORDER BY total_dispatched DESC, total_resolved_from_dispatched DESC
		LIMIT 50
	`, kpiJoinConditions, whereSQL)

	allArgs := append(kpiArgs, whereArgs...)

	var items []domain.KPIManagerScorecardItem
	if err := r.db.Raw(sql, allArgs...).Scan(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func (r *kpiLogRepository) getPRStaffDetail(userID string, filter domain.KPIQueryFilter) (*domain.KPIStaffDetailResponse, error) {
	type adminInfo struct {
		Name           string `gorm:"column:name"`
		LastName       string `gorm:"column:last_name"`
		DepartmentName string `gorm:"column:department_name"`
		RoleType       string `gorm:"column:role_type"`
		Position       string `gorm:"column:position"`
	}
	var info adminInfo
	adminSQL := `
		SELECT 
			a.name, 
			a.last_name, 
			COALESCE(ar.type, '') as role_type,
			COALESCE(a.position, '') as position,
			COALESCE(d.name, '-') as department_name
		FROM admin_users a
		LEFT JOIN admin_roles ar ON ar.id = a.role_id
		LEFT JOIN admin_departments ad ON ad.admin_id = a.id
		LEFT JOIN departments d ON d.id = ad.department_id
		WHERE a.id = ?
		LIMIT 1
	`
	_ = r.db.Raw(adminSQL, userID).Scan(&info).Error

	var prDateCond string
	var prArgs []interface{}
	prArgs = append(prArgs, userID)
	if filter.StartDate != nil {
		prDateCond += " AND pr.created_date >= ?"
		prArgs = append(prArgs, filter.StartDate)
	}
	if filter.EndDate != nil {
		prDateCond += " AND pr.created_date <= ?"
		prArgs = append(prArgs, filter.EndDate)
	}

	type prSummaryStats struct {
		PublishedCount int64 `gorm:"column:published_count"`
		TotalViews     int64 `gorm:"column:total_views"`
		TotalLikes     int64 `gorm:"column:total_likes"`
		TotalComments  int64 `gorm:"column:total_comments"`
	}
	var prStats prSummaryStats
	prSQL := fmt.Sprintf(`
		SELECT 
			COUNT(DISTINCT pr.id) as published_count,
			COALESCE(SUM(vc.count), 0) as total_views,
			COALESCE(SUM(lk.likes_count), 0) as total_likes,
			COALESCE(SUM(cm.comments_count), 0) as total_comments
		FROM module_public_relations pr
		LEFT JOIN module_public_relation_visitor_count vc ON vc.module_public_relation_id = pr.id
		LEFT JOIN (
			SELECT module_public_relation_id, COUNT(*) as likes_count
			FROM module_public_relation_likes
			GROUP BY module_public_relation_id
		) lk ON lk.module_public_relation_id = pr.id
		LEFT JOIN (
			SELECT module_public_relation_id, COUNT(*) as comments_count
			FROM module_public_relation_comments
			GROUP BY module_public_relation_id
		) cm ON cm.module_public_relation_id = pr.id
		WHERE pr.admin_user_id = ? %s
	`, prDateCond)
	_ = r.db.Raw(prSQL, prArgs...).Scan(&prStats).Error

	var notifCount int64
	var notifDateCond string
	notifArgs := []interface{}{userID}
	if filter.StartDate != nil {
		notifDateCond += " AND created_date >= ?"
		notifArgs = append(notifArgs, filter.StartDate)
	}
	if filter.EndDate != nil {
		notifDateCond += " AND created_date <= ?"
		notifArgs = append(notifArgs, filter.EndDate)
	}
	notifSQL := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM module_public_relation_notifications 
		WHERE admin_user_id = ? %s
	`, notifDateCond)
	_ = r.db.Raw(notifSQL, notifArgs...).Scan(&notifCount).Error

	type recentNewsRow struct {
		ID          string    `gorm:"column:id"`
		Title       string    `gorm:"column:title"`
		Type        string    `gorm:"column:type"`
		Status      string    `gorm:"column:status"`
		Views       int64     `gorm:"column:views"`
		Likes       int64     `gorm:"column:likes"`
		Comments    int64     `gorm:"column:comments"`
		CreatedDate time.Time `gorm:"column:created_date"`
	}
	var recentRows []recentNewsRow
	recentNewsSQL := fmt.Sprintf(`
		SELECT 
			pr.id::text as id,
			COALESCE(pr.title, '-') as title,
			COALESCE(pr.type, 'general') as type,
			COALESCE(pr.status, 'published') as status,
			COALESCE(vc.count, 0) as views,
			COALESCE(lk.likes_count, 0) as likes,
			COALESCE(cm.comments_count, 0) as comments,
			pr.created_date
		FROM module_public_relations pr
		LEFT JOIN module_public_relation_visitor_count vc ON vc.module_public_relation_id = pr.id
		LEFT JOIN (
			SELECT module_public_relation_id, COUNT(*) as likes_count
			FROM module_public_relation_likes
			GROUP BY module_public_relation_id
		) lk ON lk.module_public_relation_id = pr.id
		LEFT JOIN (
			SELECT module_public_relation_id, COUNT(*) as comments_count
			FROM module_public_relation_comments
			GROUP BY module_public_relation_id
		) cm ON cm.module_public_relation_id = pr.id
		WHERE pr.admin_user_id = ? %s
		ORDER BY pr.created_date DESC
		LIMIT 20
	`, prDateCond)
	_ = r.db.Raw(recentNewsSQL, prArgs...).Scan(&recentRows).Error

	recentNews := make([]domain.KPIPRRecentNewsItem, 0, len(recentRows))
	for _, row := range recentRows {
		recentNews = append(recentNews, domain.KPIPRRecentNewsItem{
			ID:          row.ID,
			Title:       row.Title,
			Type:        row.Type,
			Status:      row.Status,
			Views:       row.Views,
			Likes:       row.Likes,
			Comments:    row.Comments,
			CreatedDate: row.CreatedDate,
		})
	}

	return &domain.KPIStaffDetailResponse{
		UserID:                userID,
		Name:                  info.Name,
		LastName:              info.LastName,
		DepartmentName:        info.DepartmentName,
		RoleType:              info.RoleType,
		Position:              info.Position,
		PublishedNewsCount:    prStats.PublishedCount,
		TotalViewsGenerated:   prStats.TotalViews,
		TotalLikesReceived:    prStats.TotalLikes,
		TotalCommentsReceived: prStats.TotalComments,
		PushBroadcastCount:    notifCount,
		RecentNews:            recentNews,
		RecentCases:           []domain.KPIPersonalCaseItem{},
	}, nil
}

func (r *kpiLogRepository) getCCTVStaffDetail(userID string, filter domain.KPIQueryFilter) (*domain.KPIStaffDetailResponse, error) {
	type adminInfo struct {
		Name           string `gorm:"column:name"`
		LastName       string `gorm:"column:last_name"`
		RoleType       string `gorm:"column:role_type"`
		Position       string `gorm:"column:position"`
		DepartmentName string `gorm:"column:department_name"`
	}
	var info adminInfo
	adminSQL := `
		SELECT 
			a.name, 
			a.last_name, 
			COALESCE(ar.type, '') as role_type,
			COALESCE(a.position, '') as position,
			COALESCE(d.name, '-') as department_name
		FROM admin_users a
		LEFT JOIN admin_roles ar ON ar.id = a.role_id
		LEFT JOIN admin_departments ad ON ad.admin_id = a.id
		LEFT JOIN departments d ON d.id = ad.department_id
		WHERE a.id = ?
		LIMIT 1
	`
	_ = r.db.Raw(adminSQL, userID).Scan(&info).Error

	type camStats struct {
		TotalCameras  int64 `gorm:"column:total_cameras"`
		OnlineCameras int64 `gorm:"column:online_cameras"`
	}
	var cs camStats
	_ = r.db.Table("module_cctv").
		Select("COUNT(*) as total_cameras, COUNT(CASE WHEN status = 'ONLINE' THEN 1 END) as online_cameras").
		Where("deleted_at IS NULL AND created_by::text = ?", userID).
		Scan(&cs).Error

	var healthRate float64
	if cs.TotalCameras > 0 {
		healthRate = math.Round((float64(cs.OnlineCameras)/float64(cs.TotalCameras))*1000) / 10
	}

	type survStats struct {
		TotalSessions int64 `gorm:"column:total_sessions"`
		TotalDuration int64 `gorm:"column:total_duration"`
	}
	var ss survStats
	survQuery := r.db.Table("module_cctv_view_logs").
		Select("COUNT(*) as total_sessions, COALESCE(SUM(duration_seconds)/60, 0) as total_duration").
		Where("user_id::text = ? AND policy_type = 'AdminPolicy'", userID)
	if filter.StartDate != nil {
		survQuery = survQuery.Where("started_at >= ?", filter.StartDate)
	}
	if filter.EndDate != nil {
		survQuery = survQuery.Where("started_at <= ?", filter.EndDate)
	}
	_ = survQuery.Scan(&ss).Error

	type managedCamRow struct {
		ID          string `gorm:"column:id"`
		Name        string `gorm:"column:name"`
		Address     string `gorm:"column:address"`
		Status      string `gorm:"column:status"`
		AccessLevel string `gorm:"column:access_level"`
		ViewCount   int64  `gorm:"column:view_count"`
	}
	var camRows []managedCamRow
	_ = r.db.Table("module_cctv c").
		Select("c.id::text as id, c.name, COALESCE(c.address, '') as address, c.status, c.access_level, COUNT(l.id) as view_count").
		Joins("LEFT JOIN module_cctv_view_logs l ON l.cctv_id = c.id").
		Where("c.deleted_at IS NULL AND c.created_by::text = ?", userID).
		Group("c.id, c.name, c.address, c.status, c.access_level").
		Order("c.created_at DESC").
		Scan(&camRows).Error

	managedCameras := make([]domain.KPICCTVManagedCamera, len(camRows))
	for i, c := range camRows {
		managedCameras[i] = domain.KPICCTVManagedCamera{
			ID:          c.ID,
			Name:        c.Name,
			Address:     c.Address,
			Status:      c.Status,
			AccessLevel: c.AccessLevel,
			ViewCount:   c.ViewCount,
		}
	}

	type survLogRow struct {
		ID              string    `gorm:"column:id"`
		CameraName      string    `gorm:"column:camera_name"`
		StartedAt       time.Time `gorm:"column:started_at"`
		DurationSeconds int       `gorm:"column:duration_seconds"`
		DeviceType      string    `gorm:"column:device_type"`
	}
	var logRows []survLogRow
	survLogQuery := r.db.Table("module_cctv_view_logs l").
		Select("l.id::text as id, COALESCE(c.name, 'กล้องวงจรปิด') as camera_name, l.started_at, l.duration_seconds, COALESCE(l.device_type, 'Desktop') as device_type").
		Joins("LEFT JOIN module_cctv c ON c.id = l.cctv_id").
		Where("l.user_id::text = ? AND l.policy_type = 'AdminPolicy'", userID)
	if filter.StartDate != nil {
		survLogQuery = survLogQuery.Where("l.started_at >= ?", filter.StartDate)
	}
	if filter.EndDate != nil {
		survLogQuery = survLogQuery.Where("l.started_at <= ?", filter.EndDate)
	}
	_ = survLogQuery.Order("l.started_at DESC").Limit(20).Scan(&logRows).Error

	survLogs := make([]domain.KPICCTVSurveillanceLog, len(logRows))
	for i, l := range logRows {
		mins := l.DurationSeconds / 60
		if mins < 1 && l.DurationSeconds > 0 {
			mins = 1
		}
		survLogs[i] = domain.KPICCTVSurveillanceLog{
			ID:              l.ID,
			CameraName:      l.CameraName,
			StartedAt:       l.StartedAt.Format("02/01/2006 15:04"),
			DurationSeconds: l.DurationSeconds,
			DurationText:    fmt.Sprintf("%d นาที", mins),
			DeviceType:      l.DeviceType,
		}
	}

	return &domain.KPIStaffDetailResponse{
		UserID:                  userID,
		Name:                    info.Name,
		LastName:                info.LastName,
		DepartmentName:          info.DepartmentName,
		RoleType:                info.RoleType,
		Position:                info.Position,
		ManagedCamerasCount:     cs.TotalCameras,
		OnlineCamerasCount:      cs.OnlineCameras,
		CameraHealthRate:        healthRate,
		SurveillanceSessions:    ss.TotalSessions,
		SurveillanceDurationMin: ss.TotalDuration,
		ManagedCameras:          managedCameras,
		RecentSurveillanceLogs:  survLogs,
		RecentCases:             []domain.KPIPersonalCaseItem{},
	}, nil
}

func (r *kpiLogRepository) getVerifyStaffDetail(userID string, filter domain.KPIQueryFilter) (*domain.KPIStaffDetailResponse, error) {
	type adminInfo struct {
		Name           string `gorm:"column:name"`
		LastName       string `gorm:"column:last_name"`
		RoleType       string `gorm:"column:role_type"`
		Position       string `gorm:"column:position"`
		DepartmentName string `gorm:"column:department_name"`
	}
	var info adminInfo
	adminSQL := `
		SELECT 
			a.name, 
			a.last_name, 
			COALESCE(ar.type, '') as role_type,
			COALESCE(a.position, '') as position,
			COALESCE(d.name, '-') as department_name
		FROM admin_users a
		LEFT JOIN admin_roles ar ON ar.id = a.role_id
		LEFT JOIN admin_departments ad ON ad.admin_id = a.id
		LEFT JOIN departments d ON d.id = ad.department_id
		WHERE a.id::text = ?
		LIMIT 1
	`
	_ = r.db.Raw(adminSQL, userID).Scan(&info).Error

	type verifyStats struct {
		VerifiedCount  int64   `gorm:"column:verified_count"`
		RejectedCount  int64   `gorm:"column:rejected_count"`
		TotalAudited   int64   `gorm:"column:total_audited"`
		AvgDurationSec float64 `gorm:"column:avg_duration_sec"`
	}
	var vs verifyStats
	statQuery := r.db.Table("user_informations").
		Select(`
			COUNT(CASE WHEN verification_status = 'verified' THEN 1 END) as verified_count,
			COUNT(CASE WHEN verification_status = 'rejected' THEN 1 END) as rejected_count,
			COUNT(CASE WHEN verification_status IN ('verified', 'rejected') THEN 1 END) as total_audited,
			COALESCE(AVG(CASE 
				WHEN verification_status IN ('verified', 'rejected') 
				     AND COALESCE(verified_date, updated_date) > created_date 
				THEN EXTRACT(EPOCH FROM (COALESCE(verified_date, updated_date) - created_date)) 
			END), 0) as avg_duration_sec
		`).
		Where("updated_by = ?", userID)

	if filter.StartDate != nil {
		statQuery = statQuery.Where("created_date >= ?", filter.StartDate)
	}
	if filter.EndDate != nil {
		statQuery = statQuery.Where("created_date <= ?", filter.EndDate)
	}
	_ = statQuery.Scan(&vs).Error

	var approvalPct float64
	if vs.TotalAudited > 0 {
		approvalPct = math.Round((float64(vs.VerifiedCount)/float64(vs.TotalAudited))*1000) / 10
	}
	avgProcessingMin := math.Round((vs.AvgDurationSec/60.0)*10) / 10

	type caseRow struct {
		UserID          string    `gorm:"column:user_id"`
		Prefix          string    `gorm:"column:prefix"`
		Name            string    `gorm:"column:name"`
		LastName        string    `gorm:"column:last_name"`
		Status          string    `gorm:"column:verification_status"`
		District        string    `gorm:"column:district"`
		Subdistrict     string    `gorm:"column:subdistrict"`
		RejectionReason *string   `gorm:"column:rejection_reason"`
		CreatedDate     time.Time `gorm:"column:created_date"`
		VerifiedDate    time.Time `gorm:"column:verified_date"`
	}
	var rows []caseRow
	caseQuery := r.db.Table("user_informations").
		Select(`
			user_id::text as user_id,
			COALESCE(prefix, '') as prefix,
			COALESCE(name, '') as name,
			COALESCE(last_name, '') as last_name,
			verification_status,
			COALESCE(district, '') as district,
			COALESCE(subdistrict, '') as subdistrict,
			rejection_reason,
			created_date,
			COALESCE(verified_date, updated_date) as verified_date
		`).
		Where("updated_by = ?", userID)

	if filter.StartDate != nil {
		caseQuery = caseQuery.Where("created_date >= ?", filter.StartDate)
	}
	if filter.EndDate != nil {
		caseQuery = caseQuery.Where("created_date <= ?", filter.EndDate)
	}
	_ = caseQuery.Order("COALESCE(verified_date, updated_date) DESC").Limit(20).Scan(&rows).Error

	recentCases := make([]domain.KPIVerifyRecentCaseItem, len(rows))
	for i, rItem := range rows {
		durationSec := int64(rItem.VerifiedDate.Sub(rItem.CreatedDate).Seconds())
		durationMin := durationSec / 60
		if durationMin < 1 && durationSec > 0 {
			durationMin = 1
		}
		var reason string
		if rItem.RejectionReason != nil {
			reason = *rItem.RejectionReason
		}
		fullName := strings.TrimSpace(fmt.Sprintf("%s%s %s", rItem.Prefix, rItem.Name, rItem.LastName))
		if fullName == "" {
			fullName = "ประชาชนผู้ลงทะเบียน"
		}
		recentCases[i] = domain.KPIVerifyRecentCaseItem{
			UserID:          rItem.UserID,
			CitizenName:     fullName,
			Status:          rItem.Status,
			District:        rItem.District,
			Subdistrict:     rItem.Subdistrict,
			RejectionReason: reason,
			CreatedDate:     rItem.CreatedDate,
			VerifiedDate:    rItem.VerifiedDate,
			DurationMin:     durationMin,
		}
	}

	return &domain.KPIStaffDetailResponse{
		UserID:             userID,
		Name:               info.Name,
		LastName:           info.LastName,
		DepartmentName:     info.DepartmentName,
		RoleType:           info.RoleType,
		Position:           info.Position,
		VerifiedCount:      vs.VerifiedCount,
		RejectedCount:      vs.RejectedCount,
		TotalAuditedCount:  vs.TotalAudited,
		ApprovalPercentage: approvalPct,
		AvgProcessingMin:   avgProcessingMin,
		RecentVerifyCases:  recentCases,
		RecentCases:        []domain.KPIPersonalCaseItem{},
	}, nil
}

func (r *kpiLogRepository) GetStaffDetail(userID string, filter domain.KPIQueryFilter) (*domain.KPIStaffDetailResponse, error) {
	if filter.ModuleKey == domain.ModuleKeyPublicRelations {
		return r.getPRStaffDetail(userID, filter)
	}
	if filter.ModuleKey == domain.ModuleKeyCCTV {
		return r.getCCTVStaffDetail(userID, filter)
	}
	if filter.ModuleKey == domain.ModuleKeyUserVerify {
		return r.getVerifyStaffDetail(userID, filter)
	}

	type adminInfo struct {
		Name           string `gorm:"column:name"`
		LastName       string `gorm:"column:last_name"`
		DepartmentName string `gorm:"column:department_name"`
	}
	var info adminInfo
	adminSQL := `
		SELECT 
			a.name, 
			a.last_name, 
			COALESCE(d.name, '-') as department_name
		FROM admin_users a
		LEFT JOIN admin_departments ad ON ad.admin_id = a.id
		LEFT JOIN departments d ON d.id = ad.department_id
		WHERE a.id = ?
		LIMIT 1
	`
	_ = r.db.Raw(adminSQL, userID).Scan(&info).Error

	whereSQL := "WHERE (k.target_user_id = ? OR k.actor_id = ?)"
	args := []interface{}{userID, userID}

	if filter.ModuleKey != "" {
		whereSQL += " AND k.module_key = ?"
		args = append(args, filter.ModuleKey)
	}
	if filter.StartDate != nil {
		whereSQL += " AND k.created_date >= ?"
		args = append(args, filter.StartDate)
	}
	if filter.EndDate != nil {
		whereSQL += " AND k.created_date <= ?"
		args = append(args, filter.EndDate)
	}

	thresholds, _ := r.GetThresholds(filter.ModuleKey)
	resolutionCutoff := int64(thresholds.ResolutionSLAHours * 3600)

	type personalMetrics struct {
		TotalAssigned     int64   `gorm:"column:total_assigned"`
		TotalResolved     int64   `gorm:"column:total_resolved"`
		AvgMTTRHours      float64 `gorm:"column:avg_mttr_hours"`
		SLAComplianceRate float64 `gorm:"column:sla_compliance_rate"`
		TotalSLABreached  int64   `gorm:"column:total_sla_breached"`
		AvgRating         float64 `gorm:"column:avg_rating"`
		DisputeCount      int64   `gorm:"column:dispute_count"`
	}
	var metrics personalMetrics
	metricsSQL := fmt.Sprintf(`
		SELECT 
			COUNT(DISTINCT CASE WHEN k.event_code IN ('COMPLAINT_MANAGER_ASSIGNED', 'COMPLAINT_STAFF_ACCEPTED', 'COMPLAINT_STAFF_RESOLVED') THEN k.entity_id END) as total_assigned,
			COUNT(DISTINCT CASE WHEN k.event_code = 'COMPLAINT_STAFF_RESOLVED' THEN k.entity_id END) as total_resolved,
			COALESCE(ROUND((AVG(CASE WHEN k.event_code = 'COMPLAINT_STAFF_RESOLVED' AND k.step_duration_seconds > 0 THEN k.step_duration_seconds END) / 3600.0)::numeric, 1), 0) as avg_mttr_hours,
			CASE 
				WHEN COUNT(CASE WHEN k.event_code = 'COMPLAINT_STAFF_RESOLVED' THEN 1 END) > 0 
				THEN ROUND((1.0 - (COUNT(CASE WHEN k.event_code = 'COMPLAINT_STAFF_RESOLVED' AND (k.is_sla_breached = true OR (k.step_duration_seconds > %d AND k.step_duration_seconds > 0)) THEN 1 END)::numeric / COUNT(CASE WHEN k.event_code = 'COMPLAINT_STAFF_RESOLVED' THEN 1 END)::numeric)) * 100.0, 1)
				ELSE 100.0 
			END as sla_compliance_rate,
			COUNT(CASE WHEN k.event_code = 'COMPLAINT_STAFF_RESOLVED' AND (k.is_sla_breached = true OR (k.step_duration_seconds > %d AND k.step_duration_seconds > 0)) THEN 1 END) as total_sla_breached,
			COALESCE(ROUND(AVG(CASE WHEN k.event_code = 'COMPLAINT_CITIZEN_RATED' AND (k.metadata->>'rating') ~ '^[0-9]+$' THEN (k.metadata->>'rating')::numeric END), 2), 0) as avg_rating,
			COUNT(CASE WHEN k.event_code = 'COMPLAINT_CITIZEN_DISPUTED' THEN 1 END) as dispute_count
		FROM kpi_activity_logs k
		%s
	`, resolutionCutoff, resolutionCutoff, whereSQL)
	_ = r.db.Raw(metricsSQL, args...).Scan(&metrics).Error

	// Synchronize rating and dispute count from module_complaint_rating_histories
	if filter.ModuleKey == "" || filter.ModuleKey == domain.ModuleKeyComplaint {
		type personalRatingStats struct {
			AvgRating    float64 `gorm:"column:avg_rating"`
			DisputeCount int64   `gorm:"column:dispute_count"`
		}
		var pStats personalRatingStats
		pQuery := r.db.Table("module_complaint_rating_histories").
			Select("COALESCE(ROUND(AVG(rating_score)::numeric, 2), 0) as avg_rating, COUNT(CASE WHEN is_disputed THEN 1 END) as dispute_count").
			Where("assignee_id::text = ?", userID)
		if filter.StartDate != nil {
			pQuery = pQuery.Where("created_date >= ?", filter.StartDate)
		}
		if filter.EndDate != nil {
			pQuery = pQuery.Where("created_date <= ?", filter.EndDate)
		}
		if err := pQuery.Scan(&pStats).Error; err == nil {
			if pStats.AvgRating > 0 {
				metrics.AvgRating = pStats.AvgRating
			}
			if pStats.DisputeCount > 0 {
				metrics.DisputeCount = pStats.DisputeCount
			}
		}
	}

	type caseRaw struct {
		EntityID            string    `gorm:"column:entity_id"`
		DocumentNumber      string    `gorm:"column:document_number"`
		EventCode           string    `gorm:"column:event_code"`
		FromStatus          string    `gorm:"column:from_status"`
		ToStatus            string    `gorm:"column:to_status"`
		StepDurationSeconds int64     `gorm:"column:step_duration_seconds"`
		SLATargetSeconds    int64     `gorm:"column:sla_target_seconds"`
		IsSLABreached       bool      `gorm:"column:is_sla_breached"`
		BreachDuration      int64     `gorm:"column:breach_duration_seconds"`
		CreatedDate         time.Time `gorm:"column:created_date"`
		Metadata            string    `gorm:"column:metadata"`
	}
	var rawCases []caseRaw
	staffCasesWhereSQL := whereSQL + " AND k.event_code IN ('COMPLAINT_STAFF_RESOLVED', 'COMPLAINT_STAFF_ACCEPTED', 'COMPLAINT_STAFF_ACKNOWLEDGED', 'COMPLAINT_STAFF_ACTION', 'COMPLAINT_STAFF_PROGRESS', 'COMPLAINT_STAFF_REJECTED')"
	casesSQL := fmt.Sprintf(`
		SELECT 
			COALESCE(k.entity_id::text, '') as entity_id,
			COALESCE(k.document_number, '-') as document_number,
			k.event_code,
			CASE 
				WHEN k.event_code = 'COMPLAINT_STAFF_RESOLVED' THEN COALESCE(NULLIF(k.from_status, ''), 'in_progress')
				WHEN k.event_code IN ('COMPLAINT_STAFF_ACKNOWLEDGED', 'COMPLAINT_STAFF_PROGRESS', 'COMPLAINT_STAFF_ACTION') THEN COALESCE(NULLIF(k.from_status, ''), 'in_progress')
				ELSE COALESCE(k.from_status, '')
			END as from_status,
			CASE 
				WHEN k.event_code = 'COMPLAINT_STAFF_RESOLVED' THEN COALESCE(NULLIF(k.to_status, ''), 'completed')
				WHEN k.event_code IN ('COMPLAINT_STAFF_ACKNOWLEDGED', 'COMPLAINT_STAFF_PROGRESS', 'COMPLAINT_STAFF_ACTION') THEN COALESCE(NULLIF(k.to_status, ''), 'in_progress')
				ELSE COALESCE(k.to_status, '')
			END as to_status,
			k.step_duration_seconds,
			%d as sla_target_seconds,
			(k.is_sla_breached = true OR (k.event_code = 'COMPLAINT_STAFF_RESOLVED' AND k.step_duration_seconds > %d)) as is_sla_breached,
			CASE 
				WHEN (k.event_code = 'COMPLAINT_STAFF_RESOLVED' AND k.step_duration_seconds > %d) THEN (k.step_duration_seconds - %d)
				ELSE k.breach_duration_seconds
			END as breach_duration_seconds,
			k.created_date,
			COALESCE(k.metadata::text, '{}') as metadata
		FROM kpi_activity_logs k
		%s
		ORDER BY k.created_date DESC
		LIMIT 20
	`, resolutionCutoff, resolutionCutoff, resolutionCutoff, resolutionCutoff, staffCasesWhereSQL)
	_ = r.db.Raw(casesSQL, args...).Scan(&rawCases).Error

	recentCases := make([]domain.KPIPersonalCaseItem, len(rawCases))
	for i, c := range rawCases {
		recentCases[i] = domain.KPIPersonalCaseItem{
			EntityID:            c.EntityID,
			DocumentNumber:      c.DocumentNumber,
			EventCode:           c.EventCode,
			FromStatus:          c.FromStatus,
			ToStatus:            c.ToStatus,
			StepDurationSeconds: c.StepDurationSeconds,
			SLATargetSeconds:    c.SLATargetSeconds,
			IsSLABreached:       c.IsSLABreached,
			BreachDuration:      c.BreachDuration,
			CreatedDate:         c.CreatedDate,
			Metadata:            c.Metadata,
		}
	}

	return &domain.KPIStaffDetailResponse{
		UserID:            userID,
		Name:              info.Name,
		LastName:          info.LastName,
		DepartmentName:    info.DepartmentName,
		TotalAssigned:     metrics.TotalAssigned,
		TotalResolved:     metrics.TotalResolved,
		AvgMTTRHours:      metrics.AvgMTTRHours,
		SLAComplianceRate: metrics.SLAComplianceRate,
		TotalSLABreached:  metrics.TotalSLABreached,
		AvgRating:         metrics.AvgRating,
		DisputeCount:      metrics.DisputeCount,
		RecentCases:       recentCases,
	}, nil
}

func (r *kpiLogRepository) GetManagerDetail(userID string, filter domain.KPIQueryFilter) (*domain.KPIManagerDetailResponse, error) {
	type managerInfo struct {
		Name           string `gorm:"column:name"`
		LastName       string `gorm:"column:last_name"`
		Position       string `gorm:"column:position"`
		DepartmentName string `gorm:"column:department_name"`
	}
	var info managerInfo
	adminSQL := `
		SELECT 
			a.name, 
			a.last_name, 
			COALESCE(a.position, 'หัวหน้างาน') as position,
			COALESCE((
				SELECT d2.name 
				FROM admin_departments ad 
				JOIN departments d2 ON d2.id = ad.department_id 
				WHERE ad.admin_id = a.id 
				LIMIT 1
			), '-') as department_name
		FROM admin_users a
		WHERE a.id = ?
		LIMIT 1
	`
	_ = r.db.Raw(adminSQL, userID).Scan(&info).Error

	whereSQL := "WHERE k.actor_id = ? AND k.event_code IN ('COMPLAINT_MANAGER_ASSIGNED', 'COMPLAINT_CENTER_FORWARDED', 'COMPLAINT_DEPT_RETURNED')"
	args := []interface{}{userID}

	if filter.ModuleKey != "" {
		whereSQL += " AND k.module_key = ?"
		args = append(args, filter.ModuleKey)
	}
	if filter.DepartmentID != "" {
		whereSQL += " AND (k.department_id = ? OR EXISTS (SELECT 1 FROM admin_departments ad2 WHERE ad2.admin_id = ? AND ad2.department_id = ?))"
		args = append(args, filter.DepartmentID, userID, filter.DepartmentID)
	}
	if filter.StartDate != nil {
		whereSQL += " AND k.created_date >= ?"
		args = append(args, filter.StartDate)
	}
	if filter.EndDate != nil {
		whereSQL += " AND k.created_date <= ?"
		args = append(args, filter.EndDate)
	}

	thresholds, _ := r.GetThresholds(filter.ModuleKey)
	dispatchCutoff := int64(thresholds.DispatchSLAHours * 3600)

	type managerMetrics struct {
		TotalDispatched             int64   `gorm:"column:total_dispatched"`
		TotalResolvedFromDispatched int64   `gorm:"column:total_resolved_from_dispatched"`
		CompletionRate              float64 `gorm:"column:completion_rate"`
		AvgDispatchHours            float64 `gorm:"column:avg_dispatch_hours"`
		SLAComplianceRate           float64 `gorm:"column:sla_compliance_rate"`
		TotalSLABreached            int64   `gorm:"column:total_sla_breached"`
		RejectedReturnCount         int64   `gorm:"column:rejected_return_count"`
	}
	var metrics managerMetrics
	metricsSQL := fmt.Sprintf(`
		SELECT 
			COUNT(DISTINCT CASE WHEN k.event_code IN ('COMPLAINT_CENTER_FORWARDED', 'COMPLAINT_MANAGER_ASSIGNED') THEN k.entity_id END) as total_dispatched,
			COUNT(DISTINCT CASE WHEN k.event_code IN ('COMPLAINT_CENTER_FORWARDED', 'COMPLAINT_MANAGER_ASSIGNED') AND EXISTS (
				SELECT 1 FROM module_complaints mc WHERE mc.id = k.entity_id AND mc.status = 'completed'
			) THEN k.entity_id END) as total_resolved_from_dispatched,
			CASE 
				WHEN COUNT(DISTINCT CASE WHEN k.event_code IN ('COMPLAINT_CENTER_FORWARDED', 'COMPLAINT_MANAGER_ASSIGNED') THEN k.entity_id END) > 0 
				THEN ROUND((COUNT(DISTINCT CASE WHEN k.event_code IN ('COMPLAINT_CENTER_FORWARDED', 'COMPLAINT_MANAGER_ASSIGNED') AND EXISTS (
					SELECT 1 FROM module_complaints mc WHERE mc.id = k.entity_id AND mc.status = 'completed'
				) THEN k.entity_id END)::numeric / COUNT(DISTINCT CASE WHEN k.event_code IN ('COMPLAINT_CENTER_FORWARDED', 'COMPLAINT_MANAGER_ASSIGNED') THEN k.entity_id END)::numeric) * 100.0, 1)
				ELSE 0.0 
			END as completion_rate,
			COALESCE(ROUND((AVG(CASE WHEN k.event_code IN ('COMPLAINT_CENTER_FORWARDED', 'COMPLAINT_MANAGER_ASSIGNED') AND k.step_duration_seconds > 0 THEN k.step_duration_seconds END) / 3600.0)::numeric, 1), 0) as avg_dispatch_hours,
			CASE 
				WHEN COUNT(CASE WHEN k.event_code IN ('COMPLAINT_CENTER_FORWARDED', 'COMPLAINT_MANAGER_ASSIGNED') THEN 1 END) > 0 
				THEN ROUND((1.0 - (COUNT(CASE WHEN k.event_code IN ('COMPLAINT_CENTER_FORWARDED', 'COMPLAINT_MANAGER_ASSIGNED') AND (k.is_sla_breached = true OR (k.step_duration_seconds > %d AND k.step_duration_seconds > 0)) THEN 1 END)::numeric / COUNT(CASE WHEN k.event_code IN ('COMPLAINT_CENTER_FORWARDED', 'COMPLAINT_MANAGER_ASSIGNED') THEN 1 END)::numeric)) * 100.0, 1)
				ELSE 100.0 
			END as sla_compliance_rate,
			COUNT(CASE WHEN k.event_code IN ('COMPLAINT_CENTER_FORWARDED', 'COMPLAINT_MANAGER_ASSIGNED') AND (k.is_sla_breached = true OR (k.step_duration_seconds > %d AND k.step_duration_seconds > 0)) THEN 1 END) as total_sla_breached,
			COUNT(CASE WHEN k.event_code = 'COMPLAINT_DEPT_RETURNED' THEN 1 END) as rejected_return_count
		FROM kpi_activity_logs k
		%s
	`, dispatchCutoff, dispatchCutoff, whereSQL)
	_ = r.db.Raw(metricsSQL, args...).Scan(&metrics).Error

	type dispatchCaseRaw struct {
		EntityID            string    `gorm:"column:entity_id"`
		DocumentNumber      string    `gorm:"column:document_number"`
		Title               string    `gorm:"column:title"`
		CategoryName        string    `gorm:"column:category_name"`
		TargetUserName      string    `gorm:"column:target_user_name"`
		DepartmentName      string    `gorm:"column:department_name"`
		EventCode           string    `gorm:"column:event_code"`
		StepDurationSeconds int64     `gorm:"column:step_duration_seconds"`
		SLATargetSeconds    int64     `gorm:"column:sla_target_seconds"`
		IsSLABreached       bool      `gorm:"column:is_sla_breached"`
		BreachDuration      int64     `gorm:"column:breach_duration_seconds"`
		CreatedDate         time.Time `gorm:"column:created_date"`
		Reason              string    `gorm:"column:reason"`
	}
	var rawCases []dispatchCaseRaw
	casesSQL := fmt.Sprintf(`
		SELECT 
			COALESCE(k.entity_id::text, '') as entity_id,
			COALESCE(k.document_number, '-') as document_number,
			COALESCE(c.description, '-') as title,
			COALESCE(mt.name_th, '-') as category_name,
			COALESCE(NULLIF(TRIM(CONCAT(target_admin.name, ' ', target_admin.last_name)), ''), 'เจ้าหน้าที่') as target_user_name,
			COALESCE(d.name, '-') as department_name,
			k.event_code,
			k.step_duration_seconds,
			%d as sla_target_seconds,
			(k.is_sla_breached = true OR (k.event_code IN ('COMPLAINT_CENTER_FORWARDED', 'COMPLAINT_MANAGER_ASSIGNED') AND k.step_duration_seconds > %d)) as is_sla_breached,
			CASE 
				WHEN (k.step_duration_seconds > %d) THEN (k.step_duration_seconds - %d)
				ELSE k.breach_duration_seconds
			END as breach_duration_seconds,
			k.created_date,
			COALESCE(k.metadata->>'reason', '') as reason
		FROM kpi_activity_logs k
		LEFT JOIN module_complaints c ON c.id = k.entity_id
		LEFT JOIN module_types mt ON mt.id = c.module_type_id
		LEFT JOIN admin_users target_admin ON target_admin.id = k.target_user_id
		LEFT JOIN departments d ON d.id = COALESCE(k.department_id, c.department_id)
		%s
		ORDER BY k.created_date DESC
		LIMIT 30
	`, dispatchCutoff, dispatchCutoff, dispatchCutoff, dispatchCutoff, whereSQL)
	_ = r.db.Raw(casesSQL, args...).Scan(&rawCases).Error

	recentCases := make([]domain.KPIManagerDispatchCaseItem, len(rawCases))
	for i, c := range rawCases {
		recentCases[i] = domain.KPIManagerDispatchCaseItem{
			EntityID:            c.EntityID,
			DocumentNumber:      c.DocumentNumber,
			Title:               c.Title,
			CategoryName:        c.CategoryName,
			TargetUserName:      c.TargetUserName,
			DepartmentName:      c.DepartmentName,
			EventCode:           c.EventCode,
			StepDurationSeconds: c.StepDurationSeconds,
			SLATargetSeconds:    c.SLATargetSeconds,
			IsSLABreached:       c.IsSLABreached,
			BreachDuration:      c.BreachDuration,
			CreatedDate:         c.CreatedDate,
			Reason:              c.Reason,
		}
	}

	return &domain.KPIManagerDetailResponse{
		ManagerID:           userID,
		Name:                info.Name,
		LastName:            info.LastName,
		DepartmentName:      info.DepartmentName,
		Position:            info.Position,
		TotalDispatched:             metrics.TotalDispatched,
		TotalResolvedFromDispatched: metrics.TotalResolvedFromDispatched,
		CompletionRate:              metrics.CompletionRate,
		AvgDispatchHours:            metrics.AvgDispatchHours,
		SLAComplianceRate:           metrics.SLAComplianceRate,
		TotalSLABreached:            metrics.TotalSLABreached,
		RejectedReturnCount:         metrics.RejectedReturnCount,
		RecentCases:                 recentCases,
		Thresholds:                  thresholds,
	}, nil
}

func (r *kpiLogRepository) GetThresholds(moduleKey string) (*domain.KPIThresholdConfig, error) {
	if moduleKey == "" {
		moduleKey = domain.ModuleKeyComplaint
	}
	var config domain.KPIThresholdConfig
	err := r.db.Where("module_key = ?", moduleKey).First(&config).Error
	if err != nil {
		defaultConfig := domain.KPIThresholdConfig{
			ModuleKey:                   moduleKey,
			ResolutionSLAHours:          72,
			DispatchSLAHours:            24,
			ForwardSLAHours:             24,
			ManagerHighCompletionRate:   80,
			ManagerMediumCompletionRate: 50,
			SLAExcellentRate:            90,
			SLAGoodRate:                 80,
			UpdatedDate:                 time.Now().UTC(),
		}
		_ = r.db.Create(&defaultConfig)
		return &defaultConfig, nil
	}
	if config.ManagerHighCompletionRate <= 0 {
		config.ManagerHighCompletionRate = 80
	}
	if config.ManagerMediumCompletionRate <= 0 {
		config.ManagerMediumCompletionRate = 50
	}
	return &config, nil
}

func (r *kpiLogRepository) UpdateThresholds(moduleKey string, req domain.UpdateKPIThresholdRequest, adminID string) (*domain.KPIThresholdConfig, error) {
	if moduleKey == "" {
		moduleKey = domain.ModuleKeyComplaint
	}
	if req.ResolutionSLAHours <= 0 {
		req.ResolutionSLAHours = 72
	}
	if req.DispatchSLAHours <= 0 {
		req.DispatchSLAHours = 24
	}
	if req.ForwardSLAHours <= 0 {
		req.ForwardSLAHours = 24
	}
	if req.ManagerHighCompletionRate <= 0 {
		req.ManagerHighCompletionRate = 80
	}
	if req.ManagerMediumCompletionRate <= 0 {
		req.ManagerMediumCompletionRate = 50
	}
	if req.SLAExcellentRate <= 0 {
		req.SLAExcellentRate = 90
	}
	if req.SLAGoodRate <= 0 {
		req.SLAGoodRate = 80
	}

	var config domain.KPIThresholdConfig
	err := r.db.Where("module_key = ?", moduleKey).First(&config).Error
	if err != nil {
		config = domain.KPIThresholdConfig{
			ModuleKey:                   moduleKey,
			ResolutionSLAHours:          req.ResolutionSLAHours,
			DispatchSLAHours:            req.DispatchSLAHours,
			ForwardSLAHours:             req.ForwardSLAHours,
			ManagerHighCompletionRate:   req.ManagerHighCompletionRate,
			ManagerMediumCompletionRate: req.ManagerMediumCompletionRate,
			SLAExcellentRate:            req.SLAExcellentRate,
			SLAGoodRate:                 req.SLAGoodRate,
			UpdatedBy:                   adminID,
			UpdatedDate:                 time.Now().UTC(),
		}
		if createErr := r.db.Create(&config).Error; createErr != nil {
			return nil, createErr
		}
		return &config, nil
	}

	config.ResolutionSLAHours = req.ResolutionSLAHours
	config.DispatchSLAHours = req.DispatchSLAHours
	config.ForwardSLAHours = req.ForwardSLAHours
	config.ManagerHighCompletionRate = req.ManagerHighCompletionRate
	config.ManagerMediumCompletionRate = req.ManagerMediumCompletionRate
	config.SLAExcellentRate = req.SLAExcellentRate
	config.SLAGoodRate = req.SLAGoodRate
	config.UpdatedBy = adminID
	config.UpdatedDate = time.Now().UTC()

	if saveErr := r.db.Save(&config).Error; saveErr != nil {
		return nil, saveErr
	}

	return &config, nil
}

func (r *kpiLogRepository) GetBreachedCases(filter domain.KPIQueryFilter) ([]domain.KPIBreachedCaseItem, error) {
	if filter.ModuleKey == domain.ModuleKeyPublicRelations || filter.ModuleKey == domain.ModuleKeyCCTV || filter.ModuleKey == domain.ModuleKeyUserVerify {
		return []domain.KPIBreachedCaseItem{}, nil
	}

	thresholds, _ := r.GetThresholds(filter.ModuleKey)
	resSeconds := int64(thresholds.ResolutionSLAHours * 3600)
	if resSeconds <= 0 {
		resSeconds = 72 * 3600
	}

	var results []domain.KPIBreachedCaseItem

	// 1. Query Active stalled cases in module_complaints
	if filter.ModuleKey == "" || filter.ModuleKey == domain.ModuleKeyComplaint {
		activeWhere := "WHERE c.status NOT IN ('completed', 'rejected') AND EXTRACT(EPOCH FROM (NOW() - c.created_date)) > ?"
		activeArgs := []interface{}{resSeconds}

		if filter.DepartmentID != "" {
			activeWhere += " AND c.department_id = ?"
			activeArgs = append(activeArgs, filter.DepartmentID)
		} else if len(filter.DepartmentIDs) > 0 {
			activeWhere += " AND c.department_id IN (?)"
			activeArgs = append(activeArgs, filter.DepartmentIDs)
		}
		if filter.UserID != "" {
			activeWhere += " AND (c.assignee_id = ? OR c.assigner_id = ?)"
			activeArgs = append(activeArgs, filter.UserID, filter.UserID)
		}
		if filter.StartDate != nil {
			activeWhere += " AND c.created_date >= ?"
			activeArgs = append(activeArgs, filter.StartDate)
		}
		if filter.EndDate != nil {
			activeWhere += " AND c.created_date <= ?"
			activeArgs = append(activeArgs, filter.EndDate)
		}

		activeSQL := fmt.Sprintf(`
			SELECT 
				c.id::text as id,
				COALESCE(c.document_id, '-') as document_number,
				COALESCE(c.description, '-') as title,
				COALESCE(mt.name_th, '-') as category_name,
				COALESCE(c.department_id::text, '') as department_id,
				COALESCE(d.name, '-') as department_name,
				COALESCE(NULLIF(TRIM(CONCAT(a.name, ' ', a.last_name)), ''), 'รอมอบหมาย') as assignee_name,
				COALESCE(NULLIF(TRIM(CONCAT(mgr.name, ' ', mgr.last_name)), ''), '-') as assigner_name,
				c.status,
				c.created_date,
				NULL as resolved_date,
				EXTRACT(EPOCH FROM (NOW() - c.created_date))::bigint as duration_seconds,
				%d as sla_target_seconds,
				(EXTRACT(EPOCH FROM (NOW() - c.created_date)) - %d)::bigint as overdue_seconds,
				true as is_active
			FROM module_complaints c
			LEFT JOIN module_types mt ON mt.id = c.module_type_id
			LEFT JOIN departments d ON d.id = c.department_id
			LEFT JOIN admin_users a ON a.id = c.assignee_id
			LEFT JOIN admin_users mgr ON mgr.id = c.assigner_id
			%s
			ORDER BY c.created_date ASC
			LIMIT 100
		`, resSeconds, resSeconds, activeWhere)

		var activeCases []domain.KPIBreachedCaseItem
		if err := r.db.Raw(activeSQL, activeArgs...).Scan(&activeCases).Error; err == nil {
			results = append(results, activeCases...)
		}
	}

	// 2. Query Resolved cases that breached SLA in kpi_activity_logs
	logWhere := "WHERE k.event_code = 'COMPLAINT_STAFF_RESOLVED' AND (k.is_sla_breached = true OR k.step_duration_seconds > ?)"
	logArgs := []interface{}{resSeconds}

	if filter.ModuleKey != "" {
		logWhere += " AND k.module_key = ?"
		logArgs = append(logArgs, filter.ModuleKey)
	}
	if filter.DepartmentID != "" {
		logWhere += " AND (k.department_id = ? OR c.department_id = ?)"
		logArgs = append(logArgs, filter.DepartmentID, filter.DepartmentID)
	} else if len(filter.DepartmentIDs) > 0 {
		logWhere += " AND (k.department_id IN (?) OR c.department_id IN (?))"
		logArgs = append(logArgs, filter.DepartmentIDs, filter.DepartmentIDs)
	}
	if filter.UserID != "" {
		logWhere += " AND (k.target_user_id = ? OR k.actor_id = ? OR c.assignee_id = ?)"
		logArgs = append(logArgs, filter.UserID, filter.UserID, filter.UserID)
	}
	if filter.StartDate != nil {
		logWhere += " AND k.created_date >= ?"
		logArgs = append(logArgs, filter.StartDate)
	}
	if filter.EndDate != nil {
		logWhere += " AND k.created_date <= ?"
		logArgs = append(logArgs, filter.EndDate)
	}

	logSQL := fmt.Sprintf(`
		SELECT 
			c.id::text as id,
			COALESCE(c.document_id, '-') as document_number,
			COALESCE(c.description, '-') as title,
			COALESCE(mt.name_th, '-') as category_name,
			COALESCE(COALESCE(k.department_id, c.department_id)::text, '') as department_id,
			COALESCE(d.name, '-') as department_name,
			COALESCE(NULLIF(TRIM(CONCAT(a.name, ' ', a.last_name)), ''), 'เจ้าหน้าที่') as assignee_name,
			COALESCE(NULLIF(TRIM(CONCAT(mgr.name, ' ', mgr.last_name)), ''), '-') as assigner_name,
			c.status,
			c.created_date,
			k.created_date as resolved_date,
			k.step_duration_seconds as duration_seconds,
			%d as sla_target_seconds,
			CASE 
				WHEN (k.step_duration_seconds > %d) THEN (k.step_duration_seconds - %d)
				ELSE k.breach_duration_seconds
			END as overdue_seconds,
			false as is_active
		FROM kpi_activity_logs k
		JOIN module_complaints c ON c.id = k.entity_id
		LEFT JOIN module_types mt ON mt.id = c.module_type_id
		LEFT JOIN departments d ON d.id = COALESCE(k.department_id, c.department_id)
		LEFT JOIN admin_users a ON a.id = COALESCE(k.actor_id, c.assignee_id)
		LEFT JOIN admin_users mgr ON mgr.id = c.assigner_id
		%s
		ORDER BY k.created_date DESC
		LIMIT 100
	`, resSeconds, resSeconds, resSeconds, logWhere)

	var resolvedCases []domain.KPIBreachedCaseItem
	if err := r.db.Raw(logSQL, logArgs...).Scan(&resolvedCases).Error; err == nil {
		results = append(results, resolvedCases...)
	}

	if results == nil {
		results = []domain.KPIBreachedCaseItem{}
	}
	return results, nil
}

func (r *kpiLogRepository) GetCategoryStats(filter domain.KPIQueryFilter) ([]domain.KPICategoryStatItem, error) {
	if filter.ModuleKey == domain.ModuleKeyPublicRelations {
		var whereClauses []string
		var args []interface{}

		if filter.UserID != "" {
			whereClauses = append(whereClauses, "pr.admin_user_id::text = ?")
			args = append(args, filter.UserID)
		}
		if filter.DepartmentID != "" {
			whereClauses = append(whereClauses, "pr.admin_user_id IN (SELECT admin_id FROM admin_departments WHERE department_id::text = ?)")
			args = append(args, filter.DepartmentID)
		} else if len(filter.DepartmentIDs) > 0 {
			whereClauses = append(whereClauses, "pr.admin_user_id IN (SELECT admin_id FROM admin_departments WHERE department_id::text IN (?))")
			args = append(args, filter.DepartmentIDs)
		}
		if filter.StartDate != nil {
			whereClauses = append(whereClauses, "pr.created_date >= ?")
			args = append(args, filter.StartDate)
		}
		if filter.EndDate != nil {
			whereClauses = append(whereClauses, "pr.created_date <= ?")
			args = append(args, filter.EndDate)
		}

		whereSQL := ""
		if len(whereClauses) > 0 {
			whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
		}

		querySQL := fmt.Sprintf(`
			SELECT 
				COALESCE(pr.type, 'Information') as category_name,
				COUNT(*) as total_count,
				COUNT(CASE WHEN LOWER(pr.status) IN ('published', 'active') THEN 1 END) as resolved_count,
				COUNT(CASE WHEN LOWER(pr.status) = 'draft' THEN 1 END) as pending_count
			FROM module_public_relations pr
			%s
			GROUP BY pr.type
			ORDER BY total_count DESC
		`, whereSQL)

		var items []domain.KPICategoryStatItem
		if err := r.db.Raw(querySQL, args...).Scan(&items).Error; err != nil {
			return nil, err
		}

		totalAll := int64(0)
		for _, item := range items {
			totalAll += item.TotalCount
		}

		for i := range items {
			if totalAll > 0 {
				items[i].Percentage = math.Round((float64(items[i].TotalCount)/float64(totalAll))*1000) / 10
			}
		}

		if items == nil {
			items = []domain.KPICategoryStatItem{}
		}

		return items, nil
	}

	if filter.ModuleKey == domain.ModuleKeyCCTV {
		type cctvCategoryRow struct {
			CategoryName  string `gorm:"column:category_name"`
			TotalCount    int64  `gorm:"column:total_count"`
			ResolvedCount int64  `gorm:"column:resolved_count"`
			PendingCount  int64  `gorm:"column:pending_count"`
		}
		var rows []cctvCategoryRow
		q := r.db.Table("module_cctv").
			Select(`
				CASE 
					WHEN access_level = 'PUBLIC' THEN 'กล้องเพื่อประชาชน (Public)'
					ELSE 'กล้องความมั่นคง (Staff Only)'
				END as category_name,
				COUNT(*) as total_count,
				COUNT(CASE WHEN status = 'ONLINE' THEN 1 END) as resolved_count,
				COUNT(CASE WHEN status = 'OFFLINE' THEN 1 END) as pending_count
			`).
			Where("deleted_at IS NULL")

		if filter.UserID != "" {
			q = q.Where("created_by = ?", filter.UserID)
		}

		_ = q.Group("access_level").Order("total_count DESC").Scan(&rows).Error

		totalAll := int64(0)
		for _, rItem := range rows {
			totalAll += rItem.TotalCount
		}

		items := make([]domain.KPICategoryStatItem, len(rows))
		for i, rItem := range rows {
			var pct float64
			if totalAll > 0 {
				pct = math.Round((float64(rItem.TotalCount)/float64(totalAll))*1000) / 10
			}
			items[i] = domain.KPICategoryStatItem{
				CategoryName:  rItem.CategoryName,
				TotalCount:    rItem.TotalCount,
				ResolvedCount: rItem.ResolvedCount,
				PendingCount:  rItem.PendingCount,
				Percentage:    pct,
			}
		}

		return items, nil
	}

	if filter.ModuleKey == domain.ModuleKeyUserVerify {
		type districtCategoryRow struct {
			DistrictName  string `gorm:"column:district_name"`
			TotalCount    int64  `gorm:"column:total_count"`
			VerifiedCount int64  `gorm:"column:verified_count"`
			PendingCount  int64  `gorm:"column:pending_count"`
		}
		var rows []districtCategoryRow
		q := r.db.Table("user_informations").
			Select(`
				COALESCE(NULLIF(TRIM(district), ''), 'ไม่ระบุอำเภอ') as district_name,
				COUNT(*) as total_count,
				COUNT(CASE WHEN verification_status = 'verified' THEN 1 END) as verified_count,
				COUNT(CASE WHEN verification_status = 'pending' THEN 1 END) as pending_count
			`)

		if filter.UserID != "" {
			q = q.Where("updated_by = ?", filter.UserID)
		}
		if filter.DepartmentID != "" {
			q = q.Where("updated_by IN (SELECT admin_id::text FROM admin_departments WHERE department_id::text = ?)", filter.DepartmentID)
		} else if len(filter.DepartmentIDs) > 0 {
			q = q.Where("updated_by IN (SELECT admin_id::text FROM admin_departments WHERE department_id::text IN (?))", filter.DepartmentIDs)
		}
		if filter.StartDate != nil {
			q = q.Where("created_date >= ?", filter.StartDate)
		}
		if filter.EndDate != nil {
			q = q.Where("created_date <= ?", filter.EndDate)
		}

		_ = q.Group("COALESCE(NULLIF(TRIM(district), ''), 'ไม่ระบุอำเภอ')").Order("total_count DESC").Scan(&rows).Error

		totalAll := int64(0)
		for _, rItem := range rows {
			totalAll += rItem.TotalCount
		}

		items := make([]domain.KPICategoryStatItem, len(rows))
		for i, rItem := range rows {
			var pct float64
			if totalAll > 0 {
				pct = math.Round((float64(rItem.TotalCount)/float64(totalAll))*1000) / 10
			}
			items[i] = domain.KPICategoryStatItem{
				CategoryName:  rItem.DistrictName,
				TotalCount:    rItem.TotalCount,
				ResolvedCount: rItem.VerifiedCount,
				PendingCount:  rItem.PendingCount,
				Percentage:    pct,
			}
		}

		return items, nil
	}

	whereClauses := []string{"c.status != 'draft'"}
	args := []interface{}{}

	if filter.DepartmentID != "" {
		whereClauses = append(whereClauses, "c.department_id::text = ?")
		args = append(args, filter.DepartmentID)
	} else if len(filter.DepartmentIDs) > 0 {
		whereClauses = append(whereClauses, "c.department_id::text IN (?)")
		args = append(args, filter.DepartmentIDs)
	}

	if filter.UserID != "" {
		whereClauses = append(whereClauses, "(c.assignee_id::text = ? OR c.assigner_id::text = ?)")
		args = append(args, filter.UserID, filter.UserID)
	}

	if filter.StartDate != nil {
		whereClauses = append(whereClauses, "c.created_date >= ?")
		args = append(args, filter.StartDate)
	}

	if filter.EndDate != nil {
		whereClauses = append(whereClauses, "c.created_date <= ?")
		args = append(args, filter.EndDate)
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	querySQL := fmt.Sprintf(`
		SELECT 
			COALESCE(c.module_type_id::text, '') as module_type_id,
			COALESCE(NULLIF(TRIM(mt.name_th), ''), 'อื่นๆ') as category_name,
			COUNT(*) as total_count,
			COUNT(CASE WHEN c.status = 'completed' THEN 1 END) as resolved_count,
			COUNT(CASE WHEN c.status NOT IN ('completed', 'rejected') THEN 1 END) as pending_count
		FROM module_complaints c
		LEFT JOIN module_types mt ON mt.id = c.module_type_id
		%s
		GROUP BY c.module_type_id, mt.name_th
		ORDER BY total_count DESC
	`, whereSQL)

	var items []domain.KPICategoryStatItem
	if err := r.db.Raw(querySQL, args...).Scan(&items).Error; err != nil {
		return nil, err
	}

	totalAll := int64(0)
	for _, item := range items {
		totalAll += item.TotalCount
	}

	for i := range items {
		if totalAll > 0 {
			items[i].Percentage = math.Round((float64(items[i].TotalCount)/float64(totalAll))*1000) / 10
		}
	}

	if items == nil {
		items = []domain.KPICategoryStatItem{}
	}

	return items, nil
}

func (r *kpiLogRepository) GetDataSharingLogs(module string, startDate, endDate string, limit int) ([]domain.DataSharingLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var logs []domain.DataSharingLog
	query := r.db.Table("dashboard_data_sharing_logs")
	if module == domain.ModuleKeyFlood {
		query = query.Where("category IN (?)", []string{"river_sensors", "cctv_streams"})
	} else if module == domain.ModuleKeyWeather {
		query = query.Where("category IN (?)", []string{"weather_sensors", "radar_tiles"})
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate+" 00:00:00")
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}
	err := query.
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	if err != nil {
		return []domain.DataSharingLog{}, err
	}
	if logs == nil {
		logs = []domain.DataSharingLog{}
	}
	return logs, nil
}

func (r *kpiLogRepository) GetDataSharingSummary(module string, startDate, endDate string) (*domain.DataSharingSummaryResponse, error) {
	type overallStats struct {
		TotalQueries    int64   `gorm:"column:total_queries"`
		TotalDataPoints int64   `gorm:"column:total_data_points"`
		AvgLatency      float64 `gorm:"column:avg_latency"`
		TodayQueries    int64   `gorm:"column:today_queries"`
		SuccessQueries  int64   `gorm:"column:success_queries"`
	}
	var stats overallStats
	baseQuery := r.db.Table("dashboard_data_sharing_logs")
	if module == domain.ModuleKeyFlood {
		baseQuery = baseQuery.Where("category IN (?)", []string{"river_sensors", "cctv_streams"})
	} else if module == domain.ModuleKeyWeather {
		baseQuery = baseQuery.Where("category IN (?)", []string{"weather_sensors", "radar_tiles"})
	}
	if startDate != "" {
		baseQuery = baseQuery.Where("created_at >= ?", startDate+" 00:00:00")
	}
	if endDate != "" {
		baseQuery = baseQuery.Where("created_at <= ?", endDate+" 23:59:59")
	}
	_ = baseQuery.
		Select(`
			COUNT(*)::int AS total_queries,
			COALESCE(SUM(data_points), 0)::bigint AS total_data_points,
			COALESCE(AVG(duration_ms), 0)::numeric(10, 1) AS avg_latency,
			COUNT(CASE WHEN created_at >= CURRENT_DATE THEN 1 END)::int AS today_queries,
			COUNT(CASE WHEN status != 'failed' THEN 1 END)::int AS success_queries
		`).
		Scan(&stats).Error

	type catRow struct {
		Category        string  `gorm:"column:category"`
		CategoryLabel   string  `gorm:"column:category_label"`
		TotalRecords    int64   `gorm:"column:total_records"`
		TotalDataPoints int64   `gorm:"column:total_data_points"`
		AvgDuration     float64 `gorm:"column:avg_duration"`
	}
	var catRows []catRow
	catQuery := r.db.Table("dashboard_data_sharing_logs")
	if module == domain.ModuleKeyFlood {
		catQuery = catQuery.Where("category IN (?)", []string{"river_sensors", "cctv_streams"})
	} else if module == domain.ModuleKeyWeather {
		catQuery = catQuery.Where("category IN (?)", []string{"weather_sensors", "radar_tiles"})
	}
	if startDate != "" {
		catQuery = catQuery.Where("created_at >= ?", startDate+" 00:00:00")
	}
	if endDate != "" {
		catQuery = catQuery.Where("created_at <= ?", endDate+" 23:59:59")
	}
	_ = catQuery.
		Select(`
			category,
			category_label,
			COUNT(*) as total_records,
			COALESCE(SUM(data_points), 0) as total_data_points,
			COALESCE(AVG(duration_ms), 0) as avg_duration
		`).
		Group("category, category_label").
		Scan(&catRows).Error

	breakdown := make([]domain.DataSharingCategoryStat, len(catRows))
	for i, c := range catRows {
		breakdown[i] = domain.DataSharingCategoryStat{
			Category:        c.Category,
			CategoryLabel:   c.CategoryLabel,
			TotalRecords:    c.TotalRecords,
			TotalDataPoints: c.TotalDataPoints,
			AvgDurationMs:   math.Round(c.AvgDuration*10) / 10,
		}
	}

	recentLogs, _ := r.GetDataSharingLogs(module, startDate, endDate, 20)

	sharedRiverNodes := 61
	sharedWeatherNodes := 165
	if module == domain.ModuleKeyFlood {
		sharedWeatherNodes = 0
	} else if module == domain.ModuleKeyWeather {
		sharedRiverNodes = 0
	}

	uptimePercent := 100.0
	if stats.TotalQueries > 0 {
		uptimePercent = math.Round((float64(stats.SuccessQueries)/float64(stats.TotalQueries))*10000) / 100
	}

	return &domain.DataSharingSummaryResponse{
		TotalQueries:        stats.TotalQueries,
		TotalDataPoints:     stats.TotalDataPoints,
		AvgLatencyMs:        math.Round(stats.AvgLatency*10) / 10,
		TodayQueries:        stats.TodayQueries,
		SystemUptimePercent: uptimePercent,
		ActiveStreams:       61,
		SharedRiverNodes:    sharedRiverNodes,
		SharedWeatherNodes:  sharedWeatherNodes,
		CategoryBreakdown:   breakdown,
		RecentLogs:          recentLogs,
	}, nil
}

