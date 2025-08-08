// main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"collision_app_go/config"
	"collision_app_go/internal/handler"
	"collision_app_go/internal/repository"
	"collision_app_go/internal/service"
	"collision_app_go/utils"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v\n", err)
	}
	poolCfg := config.LoadPoolConfig()
	utils.Infof("Database config loaded: %s", cfg.ConnectionString())
	utils.Infof("Connection pool config: Min=%d, Max=%d", poolCfg.MinConns, poolCfg.MaxConns)

	// 2. Connect to database (using connection pool)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Configure the pool
	poolConfig, err := pgxpool.ParseConfig(cfg.ConnectionString())
	if err != nil {
		utils.Errorf("Failed to parse pool config: %v\n", err)
		log.Fatalf("Failed to parse pool config: %v\n", err)
	}
	poolConfig.MinConns = poolCfg.MinConns
	poolConfig.MaxConns = poolCfg.MaxConns
	// You can configure other pool options here like MaxConnLifetime, HealthCheckPeriod etc.

	dbpool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		utils.Errorf("Unable to create connection pool: %v\n", err)
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}
	defer dbpool.Close()

	// Test the connection
	if err := dbpool.Ping(ctx); err != nil {
		utils.Errorf("Unable to ping database: %v\n", err)
		log.Fatalf("Unable to ping database: %v\n", err)
	}
	utils.Info("Successfully connected to the database!")

	// 3. Initialize layers
	buildingRepo := repository.NewBuildingRepository(dbpool)
	collisionService := service.NewCollisionService(buildingRepo)
	buildingsService := service.NewBuildingsService(buildingRepo)
	handler := handler.NewHandler(collisionService, buildingsService)

	// 4. Setup Gin router
	// gin.SetMode(gin.ReleaseMode) // Uncomment for production
	r := gin.Default()

	// Define routes
	api := r.Group("/api/v1")
	{
		api.GET("/collision_info", handler.CollisionInfo)
		api.POST("/insert_buildings_info", handler.InsertBuildingsInfo)
		api.POST("/update_buildings_info", handler.UpdateBuildingsInfo)
		// Add more routes here...
	}

	// 5. Start server in a goroutine
	srvAddr := ":8800" // Or get from config
	server := &http.Server{
		Addr:    srvAddr,
		Handler: r,
	}

	// Channel to listen for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		utils.Infof("Starting server on %s...\n", srvAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Errorf("listen: %s\n", err)
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 6. Wait for interrupt signal
	<-quit
	utils.Info("Shutting down server...")

	// 7. Gracefully shutdown the server with a timeout
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		utils.Errorf("Server forced to shutdown:", err)
		log.Fatal("Server forced to shutdown:", err)
	}

	utils.Info("Server exiting")
	fmt.Println("Server exiting")
}
