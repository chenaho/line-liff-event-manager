package api

import (
	"log"
	"net/http"

	"event-manager/internal/models"
	"event-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	Service *service.EventService
}

func NewEventHandler(s *service.EventService) *EventHandler {
	return &EventHandler{Service: s}
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
	var event models.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set CreatedBy from context (set by AuthMiddleware)
	uid, exists := c.Get("uid")
	if exists {
		event.CreatedBy = uid.(string)
	}

	createdEvent, err := h.Service.CreateEvent(c.Request.Context(), &event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdEvent)
}

func (h *EventHandler) GetEvent(c *gin.Context) {
	eventID := c.Param("id")
	event, err := h.Service.GetEvent(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	c.JSON(http.StatusOK, event)
}

type UpdateStatusRequest struct {
	IsActive bool `json:"isActive"`
}

func (h *EventHandler) UpdateEventStatus(c *gin.Context) {
	eventID := c.Param("id")
	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Service.UpdateEventStatus(c.Request.Context(), eventID, req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *EventHandler) UpdateEvent(c *gin.Context) {
	eventID := c.Param("id")
	var event models.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure eventID matches
	event.EventID = eventID

	updatedEvent, err := h.Service.UpdateEvent(c.Request.Context(), &event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedEvent)
}

func (h *EventHandler) ListEvents(c *gin.Context) {
	events, err := h.Service.ListEvents(c.Request.Context(), 20)
	if err != nil {
		log.Printf("[ListEvents] ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[ListEvents] Returning %d events", len(events))
	c.JSON(http.StatusOK, events)
}

type ArchiveEventRequest struct {
	IsArchived bool `json:"isArchived"`
}

func (h *EventHandler) ArchiveEvent(c *gin.Context) {
	eventID := c.Param("id")
	var req ArchiveEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Service.ArchiveEvent(c.Request.Context(), eventID, req.IsArchived); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated", "isArchived": req.IsArchived})
}
