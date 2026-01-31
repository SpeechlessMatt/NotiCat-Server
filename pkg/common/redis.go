package common

// Author: Czy_4201b <speechlessmatt@qq.com>
// Created: 2026-01-21

import (
	"github.com/redis/go-redis/v9"
	"context"
)

// Gemini: lua protection
const luaReleaseLock = `
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
else
    return 0
end
`

// SafeReleaseLock 安全地释放 Redis 分布式锁
func SafeReleaseLock(rdb *redis.Client, key, value string) {
	// 释放锁通常使用 context.Background()
	// 确保即使原始请求已断开，清理工作也能完成
	rdb.Eval(context.Background(), luaReleaseLock, []string{key}, value)
}
