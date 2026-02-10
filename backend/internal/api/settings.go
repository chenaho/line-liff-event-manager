package api

import (
	"net/http"

	"event-manager/internal/service"

	"github.com/gin-gonic/gin"
)

// SettingsHandler handles settings API requests
type SettingsHandler struct {
	Service *service.SettingsService
}

// NewSettingsHandler creates a new SettingsHandler
func NewSettingsHandler(service *service.SettingsService) *SettingsHandler {
	return &SettingsHandler{Service: service}
}

// GetSettings returns all settings
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	settings, err := h.Service.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

// UpdateSetting updates a single setting
func (h *SettingsHandler) UpdateSetting(c *gin.Context) {
	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Service.Set(c.Request.Context(), req.Key, req.Value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Setting updated"})
}
