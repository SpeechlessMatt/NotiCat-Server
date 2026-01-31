package common

// Author: Czy_4201b <speechlessmatt@qq.com>
// Created: 2026-01-20

import (
	"path/filepath"
	"regexp"
)

func CleanFileName(title string) string {
	// AI真好用
	// 1. 移除非法字符（把不是字母、数字、点、横杠、中文的都换成下划线）
	// 这样可以防止路径注入，也能防止 Windows 下文件名非法
	reg := regexp.MustCompile(`[\\/:*?"<>|]`)
	safeTitle := reg.ReplaceAllString(title, "_")

	// 2. 使用 filepath.Base 强制只取最后一部分
	// 防止类似 "../../../config.yaml" 这种攻击
	fileName := filepath.Base(safeTitle)

	// 3. 处理极端情况：如果 title 全是非法字符导致变成空或点
	if fileName == "." || fileName == "/" || fileName == "" {
		return "unnamed_file"
	}
	return fileName
}
