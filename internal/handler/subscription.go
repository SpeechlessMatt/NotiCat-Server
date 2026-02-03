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
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"noticat/internal/bridge"
	"noticat/internal/model"
	"noticat/internal/service"
	"noticat/pkg/common"
	"noticat/pkg/global"
)

func CreateSubscriptionHandler(c *gin.Context) {
	type FilterInput struct {
		Type       string `json:"type" binding:"required,oneof=regex keyword"`
		Pattern    string `json:"pattern" binding:"required"`
		IgnoreCase bool   `json:"ignore_case"`
	}

	var input struct {
		SubscriptionID int            `json:"subscription_id" binding:"required"`
		Client         string         `json:"client" binding:"required"`
		Credentials    map[string]any `json:"credentials"`
		Extra          map[string]any `json:"extra"`
		Filters        []FilterInput  `json:"filters" binding:"omitempty,dive"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format error"})
		return
	}

	// fmt.Printf("%#v\n", input.Filters)

	// check client
	rawClient := strings.ToLower(input.Client)
	clientType := bridge.Client(rawClient)
	if !clientType.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的client"})
		return
	}

	// normalizeJSON: map -> string
	credsStr := common.NormalizeJSON(input.Credentials)
	extraStr := common.NormalizeJSON(input.Extra)
	data := fmt.Sprintf("%s|%s|%s", rawClient, credsStr, extraStr)
	logicHash := fmt.Sprintf("%x", sha256.Sum256([]byte(data)))

	// try to lock
	lockValue := uuid.New().String()
	lockKey := fmt.Sprintf("lock:sub:%s", logicHash)

	// set lock
	ok, err := global.RDB.SetNX(c.Request.Context(), lockKey, lockValue, 10*time.Second).Result()
	if err != nil {
		log.Printf("Redis 锁异常: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统繁忙，请稍后再试"})
		return
	}
	if !ok {
		c.JSON(http.StatusConflict, gin.H{"error": "同一个任务正在处理中，请勿重复点击"})
		return
	}
	defer common.SafeReleaseLock(global.RDB, lockKey, lockValue)

	// check userID
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "身份类型错误"})
		return
	}

	// try to compile filter
	if len(input.Filters) > 0 {
		for _, f := range input.Filters {
			if _, regErr := common.NewFilter(f.Pattern, f.Type == "regex", f.IgnoreCase); regErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("过滤器正则表达式格式错误: %s", f.Pattern),
				})
				return
			}
		}
	}

	// fast path
	// if subscription - task all exists -> skip try fetch
	fastPathErr := global.DB.Transaction(func(tx *gorm.DB) error {
		var existingTask model.FetchTask
		if err := tx.Where(&model.FetchTask{LogicHash: logicHash}).First(&existingTask).Error; err != nil {
			return err
		}

		var sub model.UserSubscription
		err := tx.Where(&model.UserSubscription{
			TaskID: existingTask.ID,
			UserID: userID,
		}).First(&sub).Error
		if err != nil {
			return err
		}
		// filter
		if len(input.Filters) > 0 {

			// input.SubscriptionID != -1 (means that user is modify the old subscription)
			// if subscription_id (upload from user) == id in database
			// (means that user is to change the filters rather than extra or credentials
			// if user change extra or credentials: subscription_id != sub.ID
			// so it will go to 'append filters'
			if input.SubscriptionID != -1 && input.SubscriptionID == int(sub.ID) {
				if err := tx.Unscoped().Where("subscription_id = ?", sub.ID).Delete(&model.SubscriptionFilter{}).Error; err != nil {
					log.Printf("清理旧过滤规则失败: %v", err)
					return err
				}
				var newFilters []model.SubscriptionFilter
				for _, f := range input.Filters {
					newFilters = append(newFilters, model.SubscriptionFilter{
						SubscriptionID: sub.ID,
						Type:           f.Type,
						Pattern:        f.Pattern,
						IgnoreCase:     f.IgnoreCase,
					})
				}
				if len(newFilters) > 0 {
					if err := tx.Create(&newFilters).Error; err != nil {
						return err
					}
				}
				return nil
			}

			// append filters
			for _, f := range input.Filters {
				newFilter := model.SubscriptionFilter{
					SubscriptionID: sub.ID,
					Pattern:        f.Pattern,
				}
				err := tx.Where(&newFilter).Assign(map[string]any{
					"type":        f.Type,
					"ignore_case": f.IgnoreCase,
				}).FirstOrCreate(&newFilter).Error
				if err != nil {
					log.Printf("保存过滤规则失败：%v", err)
					return err
				}
			}
		}
		return nil
	})
	if fastPathErr == nil {
		c.JSON(http.StatusOK, gin.H{"message": "规则已保存"})
		return
	}

	// fastPathErr: if database err? -> return
	if !errors.Is(fastPathErr, gorm.ErrRecordNotFound) {
		log.Printf("数据库异常，拦截 Fetch 流程: %v", fastPathErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统繁忙，请稍后再试"})
		return
	}

	// try to fetch
	_, notices, err := service.FetchByConfig(rawClient, credsStr, extraStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法连接源站，请检查凭证或配置"})
		return
	}

	if len(notices) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "源站目前没有任何内容，无法初始化订阅"})
		return
	}

	var finalTaskID uint
	var finalSubscriptionID uint
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		// find or create FetchTask (soft delete->restore with Assign)
		var task model.FetchTask
		err := tx.Unscoped().Where(&model.FetchTask{
			LogicHash: logicHash,
		}).Attrs(model.FetchTask{
			Client:      rawClient,
			Credentials: credsStr,
			Extra:       extraStr,
		}).Assign(map[string]any{
			"deleted_at": nil,
		}).FirstOrCreate(&task).Error
		if err != nil {
			log.Printf("任务同步失败:  %v", err)
			return err
		}

		// task_id
		finalTaskID = task.ID

		// find or create UserSubscription (soft delete->restore with Assign)
		var sub model.UserSubscription
		err = tx.Unscoped().Where(&model.UserSubscription{
			UserID: userID,
			TaskID: task.ID,
		}).Assign(map[string]any{
			"deleted_at": nil,
		}).FirstOrCreate(&sub).Error
		if err != nil {
			log.Printf("订阅失败：%v", err)
			return err
		}

		finalSubscriptionID = sub.ID

		// filter
		if len(input.Filters) > 0 {
			for _, f := range input.Filters {
				newFilter := model.SubscriptionFilter{
					SubscriptionID: sub.ID,
					Type:           f.Type,
					Pattern:        f.Pattern,
					IgnoreCase:     f.IgnoreCase,
				}
				err := tx.Where(&newFilter).FirstOrCreate(&newFilter).Error
				if err != nil {
					log.Printf("保存过滤规则失败：%v", err)
					return err
				}
			}
		}

		// init dispatch
		success := 0
		for _, notice := range notices {
			un := model.UserNotice{
				UserID:      sub.UserID,
				Client:      rawClient,
				ContentHash: notice.ContentHash(),
			}

			result := tx.Where(&un).FirstOrCreate(&un)
			if result.Error != nil {
				log.Printf("写入 UserNotice 失败: %v", result.Error)
				continue
			}
			success += 1
		}
		if success == 0 {
			log.Printf("任务 %d 失败：数据库无法写入", sub.TaskID)
			return fmt.Errorf("任务 %d 失败：数据库无法写入", sub.TaskID)
		}

		return nil
	})
	if err != nil {
		log.Printf("[API] 订阅事务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "订阅执行失败，数据已安全回滚",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "订阅成功", "task_id": finalTaskID, "subscription_id": finalSubscriptionID})
}

func DeleteSubscriptionHandler(c *gin.Context) {
	// get userID
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "身份类型错误"})
		return
	}

	subscriptionID := c.Param("id")

	err := global.DB.Transaction(func(tx *gorm.DB) error {
		var sub model.UserSubscription
		if err := tx.Where("id = ? AND user_id = ?", subscriptionID, userID).First(&sub).Error; err != nil {
			log.Printf("错误: %v", err)
			return fmt.Errorf("订阅不存在")
		}
		taskID := sub.TaskID

		// delete filters
		if err := tx.Unscoped().Where("subscription_id = ?", sub.ID).Delete(&model.SubscriptionFilter{}).Error; err != nil {
			log.Printf("错误: %v", err)
			return fmt.Errorf("无法删除filters")
		}

		// delete subscription
		if err := tx.Delete(&sub).Error; err != nil {
			log.Printf("错误: %v", err)
			return fmt.Errorf("无法删除订阅")
		}

		var count int64
		if err := tx.Model(&model.UserSubscription{}).Where("task_id = ?", taskID).Count(&count).Error; err != nil {
			log.Printf("错误: %v", err)
			return fmt.Errorf("无法统计任务订阅数")
		}

		if count == 0 {
			// delete task
			if err := tx.Delete(&model.FetchTask{}, taskID).Error; err != nil {
				log.Printf("错误: %v", err)
				return fmt.Errorf("无法删除任务")
			}
			log.Printf("任务 %d 已无订阅者，已彻底移除", taskID)
		}

		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "删除执行失败，数据已安全回滚",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func GetSubscriptionsHandler(c *gin.Context) {
	// get userID
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "身份类型错误"})
		return
	}

	var subs []model.UserSubscription

	if err := global.DB.Preload("Task", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "client")
	}).Where("user_id = ?", userID).Find(&subs).Error; err != nil {
		log.Printf("查询订阅列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	type SubscriptionResponse struct {
		ID     uint   `json:"subscription_id"`
		Client string `json:"client"`
	}

	resp := make([]SubscriptionResponse, 0, len(subs))
	for _, s := range subs {
		client := "Unknow Task"

		if s.Task.ID != 0 {
			client = s.Task.Client
		}

		resp = append(resp, SubscriptionResponse{
			ID:     s.ID,
			Client: client,
		})
	}

	c.JSON(http.StatusOK, resp)
}

func GetSubDetailHandler(c *gin.Context) {
	// get userID
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "身份类型错误"})
		return
	}

	subscriptionID := c.Param("id")

	var sub model.UserSubscription
	if err := global.DB.
		Preload("Task").
		Preload("Filters").
		Where("id = ? AND user_id = ?", subscriptionID, userID).
		First(&sub).
		Error; err != nil {

		log.Printf("查询订阅列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	extra, err := common.ParseJSONString(sub.Task.Extra)
	if err != nil {
		log.Printf("[CRITICAL] Task ID %d Extra 字段 JSON 损坏: %v", sub.TaskID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	credentials, err := common.ParseJSONString(sub.Task.Credentials)
	if err != nil {
		log.Printf("[CRITICAL] Task ID %d Credentials 字段 JSON 损坏: %v", sub.TaskID, err)
		log.Printf("json格式转换失败：%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          sub.ID,
		"client":      sub.Task.Client,
		"extra":       extra,
		"credentials": credentials,
		"filters":     sub.Filters,
	})
}
