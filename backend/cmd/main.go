package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"event-manager/internal/api"
	"event-manager/internal/repository"
	"event-manager/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load database configuration from environment
	dbConfig := repository.LoadConfigFromEnv()
	log.Printf("Using database type: %s", dbConfig.Type)

	// Initialize repositories based on configuration
	repos, err := repository.NewRepositories(dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize repositories: %v", err)
	}
	defer repos.Close()

	// Initialize cache service with 30-second TTL
	cacheService := service.NewCacheService(30 * time.Second)

	// Initialize services
	eventService := service.NewEventService(repos.Events)
	authService := service.NewAuthService(repos.Users)

	// InteractionService needs FirestoreInteractionRepository for transaction support
	// This will be abstracted further in Phase 2 when PostgreSQL support is added
	firestoreInteractionRepo, ok := repos.Interactions.(*repository.FirestoreInteractionRepository)
	if !ok {
		log.Fatal("InteractionService currently requires FirestoreInteractionRepository")
	}
	interactionService := service.NewInteractionService(firestoreInteractionRepo, repos.Events, repos.Users, cacheService)

	// Initialize Handlers
	authHandler := api.NewAuthHandler(authService)
	eventHandler := api.NewEventHandler(eventService)
	interactionHandler := api.NewInteractionHandler(interactionService)

	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "db_type": dbConfig.Type})
	})

	// Auth Routes
	r.POST("/auth/login", authHandler.Login)

	// Protected Routes
	apiGroup := r.Group("/api")
	apiGroup.Use(api.AuthMiddleware())
	{
		// Events
		apiGroup.POST("/events", eventHandler.CreateEvent)
		apiGroup.GET("/events", eventHandler.ListEvents)
		apiGroup.GET("/events/:id", eventHandler.GetEvent)
		apiGroup.GET("/events/:id/status", interactionHandler.GetEventStatus)
		apiGroup.POST("/events/:id/action", interactionHandler.HandleAction)
		apiGroup.PUT("/events/:id/status", eventHandler.UpdateEventStatus)
		apiGroup.PUT("/events/:id", eventHandler.UpdateEvent)

		// Interaction updates
		apiGroup.PATCH("/events/:id/records/:recordId/note", interactionHandler.UpdateRegistrationNote)
		apiGroup.PATCH("/events/:id/records/:recordId/content", interactionHandler.UpdateMemoContent)
		apiGroup.POST("/events/:id/records/:recordId/clap", interactionHandler.IncrementClapCount)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	r.Run(":" + port)
}
