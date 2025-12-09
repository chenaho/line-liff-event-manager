package api

import (
	"log"
	"net/http"

	"event-manager/internal/models"
	"event-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type InteractionHandler struct {
	Service *service.InteractionService
}

func NewInteractionHandler(s *service.InteractionService) *InteractionHandler {
	return &InteractionHandler{Service: s}
}

type ActionRequest struct {
	Type    models.InteractionType `json:"type" binding:"required"`
	Payload map[string]interface{} `json:"payload"`
}

func (h *InteractionHandler) HandleAction(c *gin.Context) {
	eventID := c.Param("id")
	var req ActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[ACTION] Binding error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid := c.GetString("uid")
	log.Printf("[ACTION] Event: %s, Type: %s, User: %s, Payload: %+v", eventID, req.Type, uid, req.Payload)

	interaction := models.Interaction{
		UserID: uid,
		Type:   req.Type,
	}

	// Extract payload
	if val, ok := req.Payload["userDisplayName"].(string); ok {
		interaction.UserDisplayName = val
	}
	if val, ok := req.Payload["userPictureUrl"].(string); ok {
		interaction.UserPictureUrl = val
	}
	if val, ok := req.Payload["selectedOptions"].([]interface{}); ok {
		for _, v := range val {
			if s, ok := v.(string); ok {
				interaction.SelectedOptions = append(interaction.SelectedOptions, s)
			}
		}
	}
	if val, ok := req.Payload["count"].(float64); ok {
		interaction.Count = int(val)
	}
	if val, ok := req.Payload["note"].(string); ok {
		interaction.Note = val
	}
	if val, ok := req.Payload["content"].(string); ok {
		interaction.Content = val
	}

	log.Printf("[ACTION] Constructed interaction: %+v", interaction)

	if err := h.Service.HandleAction(c.Request.Context(), eventID, &interaction); err != nil {
		log.Printf("[ACTION] HandleAction failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[ACTION] Action successful")
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *InteractionHandler) UpdateRegistrationNote(c *gin.Context) {
	eventID := c.Param("id")
	recordID := c.Param("recordId")

	var req struct {
		Note string `json:"note"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	uid := c.GetString("uid")
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := h.Service.UpdateRegistrationNote(c.Request.Context(), eventID, recordID, uid, req.Note); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *InteractionHandler) UpdateMemoContent(c *gin.Context) {
	eventID := c.Param("id")
	recordID := c.Param("recordId")

	var req struct {
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid := c.GetString("uid")
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := h.Service.UpdateMemoContent(c.Request.Context(), eventID, recordID, uid, req.Content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *InteractionHandler) IncrementClapCount(c *gin.Context) {
	eventID := c.Param("id")
	recordID := c.Param("recordId")

	uid := c.GetString("uid")
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := h.Service.IncrementClapCount(c.Request.Context(), eventID, recordID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *InteractionHandler) GetEventStatus(c *gin.Context) {
	eventID := c.Param("id")
	status, err := h.Service.GetEventStatus(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}
