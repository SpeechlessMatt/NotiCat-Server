// Copyright 2026 Czy_4201b
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

	r.Run(":" + global.AppPort)
}
