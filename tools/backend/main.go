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
	cfg := config.Load()

	db, err := database.New(cfg.DBType, cfg.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&models.Visit{}, &models.Admin{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migration completed")

	initAdmin(db, cfg.AdminPassword)

	router := gin.Default()
	router.Use(handlers.CORSMiddleware())

	authHandler := &handlers.AuthHandler{DB: db}
	router.POST("/api/login", authHandler.Login)
	router.POST("/api/check-session", authHandler.CheckSession)

	protected := router.Group("/api")
	protected.Use(handlers.AuthMiddleware())
	{
		protected.POST("/change-password", authHandler.ChangePassword)
	}

	visitHandler := &handlers.VisitHandler{DB: db}
	router.POST("/api/visit", visitHandler.LogVisit)
	router.GET("/api/stats", handlers.AuthMiddleware(), visitHandler.GetStats)
	router.GET("/api/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	router.GET("/api/myip", func(c *gin.Context) {
		ip := c.GetHeader("X-Real-IP")
		if ip == "" {
			ip = c.GetHeader("X-Forwarded-For")
		}
		if ip == "" {
			ip = c.ClientIP()
		}
		c.JSON(200, gin.H{"ip": ip})
	})

	log.Printf("Server starting on port %s\n", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initAdmin(db *gorm.DB, password string) {
	var count int64
	db.Model(&models.Admin{}).Count(&count)
	if count > 0 {
		log.Println("Admin already exists")
		return
	}
	db.Create(&models.Admin{Password: password})
	log.Println("Admin user created")
}
