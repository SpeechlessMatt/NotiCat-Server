package global

// Author: Czy_4201b <speechlessmatt@qq.com>
// Created: 2026-01-21

import (
	"log"
	"os"
)

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Printf("[Warning] 环境变量 %s 没有设置", key)
	return fallback
}

var (
	JwtSecret  = []byte(getEnv("NOTICAT_JWT_SECRET", "NotiCat"))

	SMTPSERVER = getEnv("NOTICAT_SMTP_SERVER", "163")
	ACCOUNT    = getEnv("NOTICAT_EMAIL_ACCOUNT", "")
	AUTHCODE   = getEnv("NOTICAT_EMAIL_AUTHCODE", "")

	RedisAddr = getEnv("REDIS_ADDR", "localhost:6379")
	AppPort = getEnv("APP_PORT", "8080")
)

