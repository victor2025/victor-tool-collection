package main

import (
	"log"
	"os"
	"time"

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

	// 历史明文密码升级为 bcrypt（幂等，已是哈希时直接跳过）
	if err := handlers.MigrateAdminPassword(db); err != nil {
		log.Printf("MigrateAdminPassword: %v", err)
	}

	// 清理过期会话
	db.Where("expires_at < ?", time.Now()).Delete(&models.Session{})

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
	}

	// 乐谱写操作：每次都要密码（不依赖登录会话）
	scoreWrite := router.Group("/api")
	scoreWrite.Use(handlers.WriteAuth(db))
	{
		scoreWrite.POST("/scores", scoreHandler.Create)
		scoreWrite.PATCH("/scores/:id", scoreHandler.UpdateScore)
		scoreWrite.DELETE("/scores/:id", scoreHandler.Delete)
		scoreWrite.POST("/scores/:id/images", scoreHandler.AddImages)
		scoreWrite.DELETE("/scores/images/:imageId", scoreHandler.DeleteImage)
	}

	// 只监听回环地址：外部统一走 nginx 反代，避免绕过限流/安全头直连 8003。
	// 需要对外监听时可用 VTC_LISTEN 覆盖（例如 0.0.0.0:8003，仅用于排障）。
	addr := "127.0.0.1:" + cfg.ServerPort
	if v := os.Getenv("VTC_LISTEN"); v != "" {
		addr = v
	}
	log.Printf("Listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Run: %v", err)
	}
}

func seedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&models.Admin{}).Count(&count)
	if count > 0 {
		return
	}

	// 初始密码可用 VTC_INITIAL_PASSWORD 指定，默认 admin888（文档记录，不写日志）
	initial := os.Getenv("VTC_INITIAL_PASSWORD")
	if initial == "" {
		initial = "admin888"
	}
	hashed, err := handlers.HashPassword(initial)
	if err != nil {
		log.Fatalf("seedAdmin: %v", err)
	}
	db.Create(&models.Admin{Password: hashed})
	log.Println("默认管理员已创建（密码以 bcrypt 存储），请尽快在 /admin 修改")
}
