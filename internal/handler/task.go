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

package handler

// Author: Czy_4201b <speechlessmatt@qq.com>
// Created: 2026-01-22

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"noticat/internal/bridge"
	"noticat/internal/model"
	"noticat/internal/service"
	"noticat/pkg/global"
)

func TestFetchHandler(c *gin.Context) {
	var input struct {
		TaskID uint `json:"task_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "format error"})
		return
	}

	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}
	var userID uint
	if val, ok := userIDVal.(float64); ok {
		userID = uint(val)
	} else if val, ok := userIDVal.(uint); ok {
		userID = val
	} else {
		log.Printf("实际类型是: %T", userIDVal)
		c.JSON(500, gin.H{"error": "身份类型错误"})
		return
	}

	var userSubscription model.UserSubscription
	result := global.DB.Where("user_id = ? AND task_id = ?", userID, input.TaskID).First(&userSubscription)
	if result.Error != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "fail to fetch"})
		return
	}

	client, notices, err := service.FetchByTaskID(input.TaskID)
	if err != nil {
		log.Printf("TestFetch Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fail to fetch"})
		return
	}

	if len(notices) == 0 {
		c.JSON(http.StatusOK, gin.H{"error": "fetch empty"})
		return
	}

	// fetch the first notice to test
	notice := notices[0]

	var user model.User
	if err := global.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "fail to get user email"})
		return
	}

	err = bridge.SendMail(&bridge.SendOptions{
		SMTPServer:  global.SMTPSERVER,
		Account:     global.ACCOUNT,
		AuthCode:    global.AUTHCODE,
		Subject:     "[Noticat]Test Fetch",
		Body:        notice.Title,
		From:        global.ACCOUNT,
		To:          user.Email,
		Attachments: []string{},
	})
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "fail to send email to user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "测试通过",
		"client":  client,
		"sample":  notice,
	})
}
