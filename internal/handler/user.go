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

// Package handler gin handler
package handler

// Author: Czy_4201b <speechlessmatt@qq.com>
// Created: 2026-01-21

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"noticat/internal/bridge"
	"noticat/internal/model"
	"noticat/pkg/global"
)

func SendCodeHandler(c *gin.Context) {
	reqCtx := c.Request.Context()
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "邮箱格式不正确"})
		return
	}

	emailErr := global.DB.Where("Email = ?", input.Email).First(&model.User{}).Error
	if emailErr == nil {
		c.JSON(400, gin.H{"error": "该用户名或邮箱已被占用"})
		return
	}
	if !errors.Is(emailErr, gorm.ErrRecordNotFound) {
		c.JSON(500, gin.H{"error": "数据库查询异常"})
		return
	}

	lockKey := "lock:send_code:" + input.Email

	exists, _ := global.RDB.Exists(reqCtx, lockKey).Result()
	if exists > 0 {
		c.JSON(429, gin.H{"error": "发送太频繁，请一分钟后再试"})
		return
	}

	// generate code
	code := fmt.Sprintf("%06d", rand.New(rand.NewSource(time.Now().UnixNano())).Intn(900000)+100000)
	codeKey := "code:" + input.Email
	err := global.RDB.Set(reqCtx, codeKey, code, 5*time.Minute).Err()
	if err != nil {
		c.JSON(500, gin.H{"error": "系统忙，无法生成验证码"})
		return
	}

	// set 60 freeze
	global.RDB.Set(reqCtx, lockKey, "1", time.Minute)

	fmt.Printf("目标：%s，验证码：%s\n", input.Email, code)

	err = bridge.SendMail(&bridge.SendOptions{
		SMTPServer:  global.SMTPSERVER,
		Account:     global.ACCOUNT,
		AuthCode:    global.AUTHCODE,
		Subject:     "[NotiCat]注册验证码：" + code,
		Body:        "注册验证码" + code,
		From:        global.ACCOUNT,
		To:          input.Email,
		Attachments: []string{},
	})
	if err != nil {
		log.Printf("发送邮件失败: %v", err)
		global.RDB.Del(reqCtx, codeKey)
		global.RDB.Del(reqCtx, lockKey)
		c.JSON(500, gin.H{"error": "邮件发送失败，请检查邮箱地址或稍后再试"})
		return
	}
	c.JSON(200, gin.H{"message": "验证码已发送"})
}

func RegisterHandler(c *gin.Context) {
	reqCtx := c.Request.Context()
	var input struct {
		Username string `json:"username" binding:"required,min=3,max=20,alphanum"`
		Password string `json:"password" binding:"required,min=6"`
		Email    string `json:"email" binding:"required,email"`
		Code     string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format error"})
		return
	}

	key := "code:" + input.Email
	storedCode, err := global.RDB.Get(reqCtx, key).Result()

	if err == redis.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证码已过期，请重新获取"})
		return
	} else if err != nil {
		c.JSON(500, gin.H{"error": "系统繁忙，请稍后再试"})
		return
	}

	if storedCode != input.Code {
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证码错误"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "fail to register"})
		return
	}

	newUser := model.User{
		Username: input.Username,
		Password: string(hashedPassword),
		Email:    string(input.Email),
	}
	if err := global.DB.Create(&newUser).Error; err != nil {
		c.JSON(400, gin.H{"error": "该用户名或邮箱已被占用"})
		return
	}

	global.RDB.Del(c.Request.Context(), key)

	c.JSON(200, gin.H{"message": "success"})
}

func LoginHandler(c *gin.Context) {
	var input struct {
		Account  string `json:"account" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "格式错误"})
		return
	}

	var user model.User

	result := global.DB.Where("username = ? OR email = ?", input.Account, input.Account).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "账户或密码错误"})
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "账户或密码错误"})
		return
	}

	token, err := GenerateToken(user.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "生成token失败"})
		return
	}

	c.JSON(200, gin.H{"message": "登录成功!", "token": token, "username": user.Username})
}

func GenerateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(global.JwtSecret)
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "log in first (missing Bearer token)!"})
			c.Abort()
			return
		}

		tokenString := authHeader[7:]

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "login first!"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return global.JwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "erro Auth"})
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		userID := claims["user_id"]

		var user model.User
		if err := global.DB.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user does not exist or has been deleted"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Set("username", user.Username)

		c.Next()
	}
}
