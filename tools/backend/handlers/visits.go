package handlers

import (
	"net/http"
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
	IP   string `json:"ip"`
	Tool string `json:"tool"`
}

// LogVisit records an IP + tool visit.
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

	visit := models.Visit{
		IP:        req.IP,
		Tool:      req.Tool,
		UserAgent: c.GetHeader("User-Agent"),
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
	IPDetails  []IPDetail       `json:"ip_details,omitempty"`
}

// IPDetail holds per-IP tool counts.
type IPDetail struct {
	IP        string           `json:"ip"`
	ToolCount map[string]int64 `json:"tool_count"`
	Total     int64            `json:"total"`
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

	// Hourly distribution (DB-agnostic: fetch all timestamps, group in Go)
	var allVisits []models.Visit
	h.DB.Model(&models.Visit{}).Select("visited_at").Find(&allVisits)

	hd := make(map[int]int64, 24)
	for _, v := range allVisits {
		h := v.VisitedAt.Hour()
		hd[h]++
	}

	resp := StatsResponse{
		PerTool:    pt,
		UniqueIPs:  uniqueIPs,
		HourlyDist: hd,
		Total:      total,
	}

	// Optional IP details (per-IP tool breakdown)
	if showDetail {
		var ipTools []struct {
			IP    string
			Tool  string
			Count int64
		}
		h.DB.Model(&models.Visit{}).
			Select("ip, tool, count(*) as count").
			Group("ip, tool").
			Order("ip").
			Find(&ipTools)

		ipMap := make(map[string]*IPDetail)
		for _, r := range ipTools {
			if _, ok := ipMap[r.IP]; !ok {
				ipMap[r.IP] = &IPDetail{
					IP:        r.IP,
					ToolCount: make(map[string]int64),
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


