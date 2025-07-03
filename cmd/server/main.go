package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"integration-test-example/internal/handler"
	"integration-test-example/internal/middleware"
	"integration-test-example/internal/repository"
	"integration-test-example/internal/service"
	"integration-test-example/pkg/config"
	"integration-test-example/pkg/database"
	redisClient "integration-test-example/pkg/redis"
	"log"
	"net/http"
	"time"
)

func main() {
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatal("Fail to load config:", err)
	}
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatal("Fail to connect db:", err)
	}

	rdb, err := redisClient.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	log.Println(cfg)

	r := gin.Default()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	rateLimiter := middleware.NewRateLimiter(rdb, 120, time.Minute)
	r.Use(rateLimiter.RateLimit())

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "server is running",
		})
	})

	api := r.Group("/api/v1")
	{
		todoRepo := repository.NewTodoRepository(db)
		todoSerivce := service.NewTodoService(todoRepo)
		todoHandler := handler.NewTodoHandler(todoSerivce)
		todos := api.Group("/todos")

		todos.POST("", todoHandler.CreateTodo)
		todos.GET("", todoHandler.GetTodos)
		todos.GET("/:id", todoHandler.GetTodo)
		todos.PUT("/:id", todoHandler.UpdateTodo)
		todos.DELETE("/:id", todoHandler.DeleteTodo)
	}

	log.Printf("Server starting on port %d", cfg.Server.Port)
	if err := r.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
