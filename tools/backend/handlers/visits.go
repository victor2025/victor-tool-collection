package handlers

import (
	"net/http"
	"strconv"
	"time"

	"victor-tool-collection/backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// VisitHandler groups visit-related endpoints.
type VisitHandler struct {
	DB *gorm.DB
}

// LogVisitRequest is the JSON body for logging a visit.
type LogVisitRequest struct {
	IP       string `json:"ip"`
	Tool     string `json:"tool"`
	DeviceID string `json:"device_id"`
}

// LogVisit records an IP + tool visit with dedup (1s window).
// POST /api/visit
func (h *VisitHandler) LogVisit(c *gin.Context) {
	var req LogVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Tool = "unknown"
	}
	if req.IP == "" {
		req.IP = c.GetHeader("X-Real-IP")
		if req.IP == "" {
			req.IP = c.GetHeader("X-Forwarded-For")
		}
		if req.IP == "" {
			req.IP = c.ClientIP()
		}
	}

	// 同一 IP + 工具 + 设备 1 秒内去重
	var recent int64
	h.DB.Model(&models.Visit{}).
		Where("ip = ? AND tool = ? AND device_id = ? AND visited_at > ?", req.IP, req.Tool, req.DeviceID, time.Now().Add(-1*time.Second)).
		Count(&recent)
	if recent > 0 {
		c.JSON(http.StatusOK, gin.H{"message": "visit deduped"})
		return
	}

	visit := models.Visit{
		IP:        req.IP,
		Tool:      req.Tool,
		UserAgent: c.GetHeader("User-Agent"),
		DeviceID:  req.DeviceID,
		VisitedAt: time.Now(),
	}

	if err := h.DB.Create(&visit).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to log visit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "visit logged", "id": visit.ID})
}

// StatsResponse holds the aggregated statistics.
type StatsResponse struct {
	PerTool    map[string]int64 `json:"per_tool"`
	UniqueIPs  int64            `json:"unique_ips"`
	HourlyDist map[int]int64    `json:"hourly_distribution"`
	Total      int64            `json:"total"`
	DailyDist  map[string]int64 `json:"daily_distribution"`
	IPDetails  []IPDetail       `json:"ip_details,omitempty"`
}

// IPDetail holds per-IP tool counts.
type IPDetail struct {
	IP        string           `json:"ip"`
	ToolCount map[string]int64 `json:"tool_count"`
	Total     int64            `json:"total"`
	UserAgent string           `json:"user_agent,omitempty"`
}

// GetStats returns aggregated visit statistics.
// GET /api/stats?ip_detail=1
func (h *VisitHandler) GetStats(c *gin.Context) {
	showDetail := c.Query("ip_detail") == "1"

	// Total count
	var total int64
	h.DB.Model(&models.Visit{}).Count(&total)

	// Per-tool counts
	var perTool []struct {
		Tool  string
		Count int64
	}
	h.DB.Model(&models.Visit{}).
		Select("tool, count(*) as count").
		Group("tool").
		Find(&perTool)

	pt := make(map[string]int64, len(perTool))
	for _, r := range perTool {
		pt[r.Tool] = r.Count
	}

	// Unique IPs
	var uniqueIPs int64
	h.DB.Model(&models.Visit{}).
		Distinct("ip").
		Count(&uniqueIPs)

	// Hourly & daily distribution (DB-agnostic: fetch all timestamps, group in Go)
	var allVisits []models.Visit
	h.DB.Model(&models.Visit{}).Select("visited_at").Find(&allVisits)

	hd := make(map[int]int64, 24)
	dd := make(map[string]int64)
	for _, v := range allVisits {
		h := v.VisitedAt.Hour()
		hd[h]++
		day := v.VisitedAt.Format("2006-01-02")
		dd[day]++
	}

	resp := StatsResponse{
		PerTool:    pt,
		UniqueIPs:  uniqueIPs,
		HourlyDist: hd,
		DailyDist:  dd,
		Total:      total,
	}

	// Optional IP details (per-IP tool breakdown)
	if showDetail {
		var ipTools []struct {
			IP        string
			Tool      string
			Count     int64
			UserAgent string
		}
		h.DB.Model(&models.Visit{}).
			Select("ip, tool, count(*) as count, max(user_agent) as user_agent").
			Group("ip, tool").
			Order("ip").
			Find(&ipTools)

		ipMap := make(map[string]*IPDetail)
		for _, r := range ipTools {
			if _, ok := ipMap[r.IP]; !ok {
				ipMap[r.IP] = &IPDetail{
					IP:        r.IP,
					ToolCount: make(map[string]int64),
					UserAgent: r.UserAgent,
				}
			}
			ipMap[r.IP].ToolCount[r.Tool] = r.Count
			ipMap[r.IP].Total += r.Count
		}

		details := make([]IPDetail, 0, len(ipMap))
		for _, d := range ipMap {
			details = append(details, *d)
		}
		resp.IPDetails = details
	}

	c.JSON(http.StatusOK, resp)
}

// GetVisits returns paginated visit records.
// GET /api/visits?page=1&page_size=20&tool=xxx&start_date=2026-07-01&end_date=2026-07-20
func (h *VisitHandler) GetVisits(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	toolFilter := c.Query("tool")
	showLabel := c.Query("show_label") == "1"
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := h.DB.Model(&models.Visit{})
	if toolFilter != "" {
		query = query.Where("tool = ?", toolFilter)
	}
	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("visited_at >= ?", t)
		}
	}
	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			// Include the full end date (end of day)
			query = query.Where("visited_at < ?", t.Add(24*time.Hour))
		}
	}

	var total int64
	query.Count(&total)

	var visits []models.Visit
	query.Order("id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&visits)

	// Get available tools for filter dropdown
	var tools []string
	h.DB.Model(&models.Visit{}).Distinct("tool").Pluck("tool", &tools)

	// Resolve labels if requested
	var labels map[string]string
	if showLabel {
		var deviceLabels []models.DeviceLabel
		h.DB.Find(&deviceLabels)
		labels = make(map[string]string, len(deviceLabels))
		for _, dl := range deviceLabels {
			labels[dl.DeviceID] = dl.Label
		}
	}

	// Enrich visits with label info
	type VisitWithLabel struct {
		models.Visit
		Label string `json:"label,omitempty"`
	}
	data := make([]VisitWithLabel, len(visits))
	for i, v := range visits {
		data[i] = VisitWithLabel{Visit: v}
		if showLabel {
			data[i].Label = labels[v.DeviceID]
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":    true,
		"data":  data,
		"total": total,
		"page":  page,
		"size":  pageSize,
		"tools": tools,
	})
}



