// Package scheduler scheduler
package scheduler

// Author: Czy_4201b <speechlessmatt@qq.com>
// Created: 2026-01-22

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"noticat/internal/model"
	"noticat/internal/service"
	"noticat/pkg/global"
)

var shanghaiLoc *time.Location

func StartScheduler() {
	var err error
	shanghaiLoc, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Fatalf("无法加载时区: %v", err)
	}

	c := cron.New(
		cron.WithLocation(shanghaiLoc),
		cron.WithChain(cron.Recover(cron.DefaultLogger)),
	)

	_, err = c.AddFunc("@every 30m", func() {
		log.Println("[Scheduler] 🔔 触发整点扫描，开始派发任务...")
		DispatchAllTasks()
	})
	if err != nil {
		log.Println("[Scheduler] 任务派发过程出现异常，跳过异常")
	}

	c.Start()
	log.Println("[Scheduler] 🚀 调度服务已上线，运行频率：每30分钟/次")
}

func DispatchAllTasks() {
	var tasks []model.FetchTask
	if err := global.DB.Find(&tasks); err != nil {
		log.Printf("[Scheduler] 数据库繁忙: %v", err)
		return
	}

	now := time.Now().In(shanghaiLoc)
	log.Printf("[Scheduler] ⏰ Cron 触发 | time=%s | unix=%d", now.Format("2006-01-02 15:04:05"), now.Unix())
	log.Printf("[Scheduler] 本次共发现 %d 个待执行任务", len(tasks))

	// use a buffered channel as a semaphore to limit maximum concurrent tasks
	sem := make(chan struct{}, 3)

	for _, task := range tasks {
		sem <- struct{}{}

		go func(t model.FetchTask) {
			log.Printf("[Worker] 正在处理任务: (ID: %d)", t.ID)

			defer func() { <-sem }()
			safeExecute(t.ID)

			log.Printf("[Worker] 任务执行完毕: %d", t.ID)
		}(task)

		time.Sleep(2 * time.Second)
	}
}

func safeExecute(taskID uint) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Panic] 任务 %d 执行时崩溃: %v", taskID, r)
		}
	}()

	service.DispatchMail(taskID)
}
