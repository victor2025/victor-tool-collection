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

	// Auto-migrate all models
	if err := db.AutoMigrate(&models.Visit{}, &models.Admin{}, &models.Session{}, &models.DeviceLabel{}, &models.Score{}, &models.ScoreImage{}); err != nil {
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
	deviceLabelHandler := &handlers.DeviceLabelHandler{DB: db}
	systemHandler := &handlers.SystemHandler{}
	scoreHandler := &handlers.ScoreHandler{DB: db, Dir: cfg.ScoresDir}

	// Public routes
	router.POST("/api/login", authHandler.Login)
	router.POST("/api/logout", authHandler.Logout)
	router.GET("/api/check-session", authHandler.CheckSession)
	router.POST("/api/visit", visitHandler.LogVisit)
	router.GET("/api/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	// 乐谱夹：查看公开，写入需要登录
	router.GET("/api/scores", scoreHandler.List)
	router.GET("/api/scores/types", scoreHandler.Types)

	// Protected routes
	protected := router.Group("/api")
	protected.Use(handlers.AuthRequired(db))
	{
		protected.POST("/change-password", authHandler.ChangePassword)
		protected.GET("/stats", visitHandler.GetStats)
		protected.GET("/visits", visitHandler.GetVisits)
		protected.GET("/device-labels", deviceLabelHandler.ListLabels)
		protected.POST("/device-labels", deviceLabelHandler.UpsertLabel)
		protected.DELETE("/device-labels/:device_id", deviceLabelHandler.DeleteLabel)
		protected.GET("/system", systemHandler.GetSystem)
		protected.POST("/scores", scoreHandler.Create)
		protected.PATCH("/scores/:id", scoreHandler.UpdateScore)
		protected.DELETE("/scores/:id", scoreHandler.Delete)
		protected.POST("/scores/:id/images", scoreHandler.AddImages)
		protected.DELETE("/scores/images/:imageId", scoreHandler.DeleteImage)
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
