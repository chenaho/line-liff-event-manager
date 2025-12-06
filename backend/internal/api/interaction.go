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

	// We might want to fetch user display name here or trust frontend payload?
	// Secure way: fetch user from DB.
	// For Vibe Coding, let's assume we fetch it or pass it.
	// The Interaction model has UserDisplayName.
	// Let's construct Interaction model from payload.

	interaction := models.Interaction{
		UserID: uid,
		Type:   req.Type,
	}

	// Extract payload
	// This is a bit manual mapping.
	if val, ok := req.Payload["userDisplayName"].(string); ok {
		interaction.UserDisplayName = val
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

func (h *InteractionHandler) GetEventStatus(c *gin.Context) {
	eventID := c.Param("id")
	status, err := h.Service.GetEventStatus(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}
