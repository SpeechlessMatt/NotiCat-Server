// Package global Init Database
package global

// Author: Czy_4201b <speechlessmatt@qq.com>
// Created: 2026-01-21

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"noticat/internal/model"
)

func InitInfrastructure() {
	// --- 1. 初始化 SQLite ---
	var err error
	DB, err = gorm.Open(sqlite.Open("noticat.db"), &gorm.Config{})
	if err != nil {
		panic("无法连接数据库: " + err.Error())
	}

	// 性能优化：开启 WAL 模式
	if sqlDB, err := DB.DB(); err == nil {
		sqlDB.Exec("PRAGMA journal_mode=WAL;")
	}

	// 自动迁移表结构
	DB.AutoMigrate(&model.User{}, &model.UserSubscription{}, &model.SubscriptionFilter{}, &model.UserNotice{}, &model.FetchTask{})

	// --- 2. 初始化 Redis ---
	RDB = redis.NewClient(&redis.Options{
		Addr:     RedisAddr,
		Password: "", // 如果没设密码就留空
		DB:       0,
	})

	// 测试 Redis 连通性
	timeoutCtx, cancel := context.WithTimeout(Ctx, 5*time.Second)
	defer cancel()
	if _, err := RDB.Ping(timeoutCtx).Result(); err != nil {
		panic("Redis 连接失败: " + err.Error())
	}

	fmt.Println("🚀 数据库与 Redis 初始化成功！")
}
