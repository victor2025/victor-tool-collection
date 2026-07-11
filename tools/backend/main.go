package main

import (
	"log"
	"victor-tool-collection/backend/config"
	"victor-tool-collection/backend/database"
	"victor-tool-collection/backend/handlers"
	"victor-tool-collection/backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config.Load: %v", err)
	}

	db, err := database.New(cfg.DBType, cfg.DSN)
	if err != nil {
		log.Fatalf("database.New: %v", err)
	}

	// Auto-migrate all models (including new Session)
	if err := db.AutoMigrate(&models.Visit{}, &models.Admin{}, &models.Session{}); err != nil {
		log.Fatalf("AutoMigrate: %v", err)
	}
	log.Println("Migration completed")

	// Seed default admin if none exists
	seedAdmin(db)

	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(handlers.CORSMiddleware())

	authHandler := &handlers.AuthHandler{DB: db}
	visitHandler := &handlers.VisitHandler{DB: db}

	// Public routes
	router.POST("/api/login", authHandler.Login)
	router.POST("/api/logout", authHandler.Logout)
	router.GET("/api/check-session", authHandler.CheckSession)
	router.POST("/api/visit", visitHandler.LogVisit)
	router.GET("/api/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// Protected routes
	protected := router.Group("/api")
	protected.Use(handlers.AuthRequired(db))
	{
		protected.POST("/change-password", authHandler.ChangePassword)
		protected.GET("/stats", visitHandler.GetStats)
		protected.GET("/visits", visitHandler.GetVisits)
	}

	log.Printf("Listening on :%s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Run: %v", err)
	}
}

func seedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&models.Admin{}).Count(&count)
	if count > 0 {
		return
	}
	db.Create(&models.Admin{Password: "admin888"})
	log.Println("Default admin created (password: admin888)")
}
