// Package model model
package model

// Author: Czy_4201b <speechlessmatt@qq.com>
// Created: 2026-01-22

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"unique;not null" json:"username"`
	Password string `gorm:"not null" json:"password"`
	Email    string `gorm:"unique;not null" json:"email"`
}

type UserSubscription struct {
	gorm.Model
	UserID  uint                 `gorm:"uniqueIndex:idx_user_task"`
	User    User                 `gorm:"foreignKey:UserID"`
	TaskID  uint                 `gorm:"not null;uniqueIndex:idx_user_task"`
	Task    FetchTask            `gorm:"foreignKey:TaskID"`
	Filters []SubscriptionFilter `gorm:"foreignKey:SubscriptionID"`
}

type SubscriptionFilter struct {
	gorm.Model
	SubscriptionID uint   `gorm:"index:idx_sub_pattern"`
	Type           string `json:"type"`
	Pattern        string `gorm:"index:idx_sub_pattern" json:"pattern"`
	IgnoreCase     bool   `json:"ignore_case"`
}

type UserNotice struct {
	gorm.Model
	UserID      uint   `gorm:"uniqueIndex:idx_user_content"`
	Client      string `gorm:"uniqueIndex:idx_user_content"`
	ContentHash string `gorm:"uniqueIndex:idx_user_content"`
}

type FetchTask struct {
	gorm.Model
	LogicHash   string `gorm:"unique"`
	Client      string `gorm:"not null"`
	Credentials string
	Extra       string
	LastFetchAt time.Time
}
