package handlers

import (
	"net/http"

	"victor-tool-collection/backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DeviceLabelHandler groups device-label endpoints.
type DeviceLabelHandler struct {
	DB *gorm.DB
}

// UpsertLabelRequest is the body for creating/updating a label.
type UpsertLabelRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
	Label    string `json:"label" binding:"required"`
}

// ListLabels returns all device labels.
// GET /api/device-labels
func (h *DeviceLabelHandler) ListLabels(c *gin.Context) {
	var labels []models.DeviceLabel
	h.DB.Order("updated_at desc").Find(&labels)
	c.JSON(http.StatusOK, gin.H{"ok": true, "data": labels})
}

// UpsertLabel creates or updates a label for a device.
// POST /api/device-labels
func (h *DeviceLabelHandler) UpsertLabel(c *gin.Context) {
	var req UpsertLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ok": false, "error": "device_id 和 label 不能为空"})
		return
	}

	var existing models.DeviceLabel
	result := h.DB.Where("device_id = ?", req.DeviceID).First(&existing)
	if result.Error != nil {
		// Create new
		label := models.DeviceLabel{
			DeviceID: req.DeviceID,
			Label:    req.Label,
		}
		if err := h.DB.Create(&label).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "创建标签失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "data": label})
	} else {
		// Update existing
		existing.Label = req.Label
		if err := h.DB.Save(&existing).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "更新标签失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "data": existing})
	}
}

// DeleteLabel deletes a device label by device_id.
// DELETE /api/device-labels/:device_id
func (h *DeviceLabelHandler) DeleteLabel(c *gin.Context) {
	deviceID := c.Param("device_id")
	if deviceID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ok": false, "error": "device_id 不能为空"})
		return
	}

	result := h.DB.Where("device_id = ?", deviceID).Delete(&models.DeviceLabel{})
	if result.RowsAffected == 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"ok": false, "error": "标签不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "标签已删除"})
}
