package main

// Author: Czy_4201b <speechlessmatt@qq.com>
// Created: 2026-01-16

import (
	"net/http"

	"noticat/internal/handler"
	"noticat/internal/meta"
	"noticat/internal/scheduler"
	"noticat/pkg/global"

	"github.com/gin-gonic/gin"
)

func main() {
	// Init Database
	global.InitInfrastructure()

	// Start scheduler
	scheduler.StartScheduler()

	r := gin.Default()

	meta.RegisterRoutes(r, "internal/meta/data")
	r.POST("/sendcode", handler.SendCodeHandler)
	r.POST("/register", handler.RegisterHandler)
	r.POST("/login", handler.LoginHandler)

	api := r.Group("/api")
	{
		api.Use(handler.AuthMiddleware())

		api.GET("/ping", func(c *gin.Context) {
			username, _ := c.Get("username")

			c.JSON(http.StatusOK, gin.H{
				"status":  "OK",
				"message": "Login Success",
				"username": username,
			})
		})

		api.POST("/testfetch", handler.TestFetchHandler)
		api.POST("/subscription", handler.CreateSubscriptionHandler)
		api.DELETE("/subscription/:id", handler.DeleteSubscriptionHandler)
		api.GET("/subscriptions", handler.GetSubscriptionsHandler)
		api.GET("/subscription/:id", handler.GetSubDetailHandler)
	}

	r.Run(":8080")
}
