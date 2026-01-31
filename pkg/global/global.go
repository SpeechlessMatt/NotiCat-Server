package global

// Author: Czy_4201b <speechlessmatt@qq.com>
// Created: 2026-01-21

import (
	"context"
    "gorm.io/gorm"
    "github.com/redis/go-redis/v9"
)

var (
    DB  *gorm.DB
    RDB *redis.Client
	Ctx = context.Background()
)
