package main

import (
	"event-manager/internal/api"
	"event-manager/internal/repository"
	"event-manager/internal/service"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize Repository
	firestoreRepo, err := repository.NewFirestoreRepository() // Renamed repo to firestoreRepo
	if err != nil {
		log.Fatalf("Failed to initialize Firestore: %v", err)
	}
	defer firestoreRepo.Close() // Renamed

	// Initialize services
	eventService := service.NewEventService(firestoreRepo) // Passed firestoreRepo

	// Initialize cache service with 30-second TTL
	cacheService := service.NewCacheService(30 * time.Second) // New cache service initialization

	interactionService := service.NewInteractionService(firestoreRepo, cacheService) // Passed firestoreRepo and cacheService
	authService := service.NewAuthService(firestoreRepo)                             // Passed firestoreRepo and reordered

	// Initialize Handlers
	authHandler := api.NewAuthHandler(authService)
	eventHandler := api.NewEventHandler(eventService)
	interactionHandler := api.NewInteractionHandler(interactionService)

	r := gin.Default()

	// CORS Middleware (Basic)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/api/health", func(c *gin.Context) {
		// Check required environment variables
		jwtSecret := os.Getenv("JWT_SECRET")
		lineChannelID := os.Getenv("LINE_CHANNEL_ID")
		adminList := os.Getenv("ADMIN_LIST")

		envStatus := gin.H{
			"JWT_SECRET":      jwtSecret != "",
			"LINE_CHANNEL_ID": lineChannelID != "",
			"ADMIN_LIST":      adminList != "",
		}

		allConfigured := jwtSecret != "" && lineChannelID != "" && adminList != ""

		c.JSON(http.StatusOK, gin.H{
			"status":         "ok",
			"env_configured": allConfigured,
			"env_status":     envStatus,
		})
	})

	// Auth Routes
	r.POST("/api/auth/login", authHandler.Login)

	// Protected Routes Group
	apiGroup := r.Group("/api")
	apiGroup.Use(api.AuthMiddleware())
	{
		apiGroup.POST("/events", api.AdminMiddleware(), eventHandler.CreateEvent)
		apiGroup.PUT("/events/:id", api.AdminMiddleware(), eventHandler.UpdateEvent)
		apiGroup.PUT("/events/:id/status", api.AdminMiddleware(), eventHandler.UpdateEventStatus)
		apiGroup.GET("/events", api.AdminMiddleware(), eventHandler.ListEvents) // Admin list

		// Public/User Event Routes (some might need auth, some public)
		// Spec says "GET /api/events/{id}" is for user interaction
		apiGroup.GET("/events/:id", eventHandler.GetEvent)
		apiGroup.GET("/events/:id/status", interactionHandler.GetEventStatus)
		apiGroup.POST("/events/:id/action", interactionHandler.HandleAction)
		apiGroup.PATCH("/events/:id/records/:recordId/note", interactionHandler.UpdateRegistrationNote)
		apiGroup.PATCH("/events/:id/records/:recordId/content", interactionHandler.UpdateMemoContent)
		apiGroup.POST("/events/:id/records/:recordId/clap", interactionHandler.IncrementClapCount)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
