// Package service some services
package service

// Author: Czy_4201b <speechlessmatt@qq.com>
// Created: 2026-01-22

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"noticat/internal/bridge"
	"noticat/internal/model"
	"noticat/pkg/common"
	"noticat/pkg/global"
)

type FetchContext struct {
	Client   string
	Account  string
	Password string
	Extra    map[string]any
}

// DispatchMail send mail to user
func DispatchMail(taskID uint) {
	fetchCtx, notices, err := FetchByTaskID(taskID)
	if err != nil {
		log.Printf("任务 %d 失败: %v", taskID, err)
		return
	}

	if len(notices) == 0 {
		log.Printf("任务 %d 失败: 取到了空的Notices", taskID)
		return
	}

	// client check (just a check)
	if clientCheck := bridge.Client(strings.ToLower(fetchCtx.Client)); !clientCheck.IsValid() {
		return
	}

	// find user who need this task
	var subscriptions []model.UserSubscription
	err = global.DB.
		Preload("Filters").
		Preload("User").
		Where("task_id = ?", taskID).
		Find(&subscriptions).Error
	if err != nil {
		log.Printf("任务 %d 失败: %v", taskID, err)
		return
	}

	type SubscriptionWithFilters struct {
		sub     *model.UserSubscription
		filters []*common.StringFilter
	}

	var subsWithFilters []SubscriptionWithFilters

	// 2026/1/22
	// chatGPT optimize: compile first
	// Gemini optimize: use point
	for i := range subscriptions {
		filters := make([]*common.StringFilter, 0, len(subscriptions[i].Filters))
		for _, f := range subscriptions[i].Filters {
			sf, err := common.NewFilter(f.Pattern, f.Type == "regex", f.IgnoreCase)
			if err == nil {
				filters = append(filters, sf)
			}
		}

		subsWithFilters = append(subsWithFilters, SubscriptionWithFilters{
			sub:     &subscriptions[i],
			filters: filters,
		})
	}

	// for every users who subscript this noice
	for _, swf := range subsWithFilters {
		sub := swf.sub

		uid := sub.UserID
		user := sub.User
		activeFilters := swf.filters

		for _, notice := range notices {
			// use a closure to ensure 'defer' executes at the end of each iteration
			func() {
				// filter old notice and push new notice
				un := model.UserNotice{
					UserID:      uid,
					Client:      fetchCtx.Client,
					ContentHash: notice.ContentHash(),
				}

				result := global.DB.Where(&un).FirstOrCreate(&un)
				if result.Error != nil {
					log.Printf("写入 UserNotice 失败: %v", result.Error)
					return
				}

				// is New Notice?
				if result.RowsAffected > 0 {
					if len(activeFilters) > 0 {
						matched := false
						for _, sf := range activeFilters {
							if sf.Match(notice.Title) {
								matched = true
								break
							}
						}
						if !matched {
							return
						}
					}

					// try to fetch detail
					detail, err := bridge.FetchDetailFromPython(&bridge.DetailOptions{
						Client:   bridge.Client(fetchCtx.Client),
						Account:  fetchCtx.Account,
						Password: fetchCtx.Password,
						URL:      notice.URL,
						Extra:    fetchCtx.Extra,
					})
					if err != nil {
						// if non detail: just send title
						log.Printf("non detail: %v", err)
						err := bridge.SendMail(&bridge.SendOptions{
							SMTPServer:  global.SMTPSERVER,
							Account:     global.ACCOUNT,
							AuthCode:    global.AUTHCODE,
							Subject:     "[NotiCat]" + common.ShortenTitle(notice.Title),
							Body:        notice.Title,
							From:        global.ACCOUNT,
							To:          user.Email,
							Attachments: []string{},
						})
						if err != nil {
							log.Printf("发送邮件失败: %v", err)
							return
						}
						return
					}

					body := detail.Body

					cacheRoot := ".cache"
					if err := os.MkdirAll(cacheRoot, 0o755); err != nil {
						log.Printf("创建cache失败，下载失败: %v", err)

						// if non cache: just send title and body
						err := bridge.SendMail(&bridge.SendOptions{
							SMTPServer:  global.SMTPSERVER,
							Account:     global.ACCOUNT,
							AuthCode:    global.AUTHCODE,
							Subject:     "[NotiCat]" + common.ShortenTitle(notice.Title),
							Body:        body,
							From:        global.ACCOUNT,
							To:          user.Email,
							Attachments: []string{},
						})
						if err != nil {
							log.Printf("发送邮件失败: %v", err)
							return
						}
						return
					}

					cacheDir, err := os.MkdirTemp(cacheRoot, "noticat_")
					if err != nil {
						log.Printf("创建临时目录失败: %v", err)
						// if non cache: just send title and body
						err := bridge.SendMail(&bridge.SendOptions{
							SMTPServer:  global.SMTPSERVER,
							Account:     global.ACCOUNT,
							AuthCode:    global.AUTHCODE,
							Subject:     "[NotiCat]" + common.ShortenTitle(notice.Title),
							Body:        body,
							From:        global.ACCOUNT,
							To:          user.Email,
							Attachments: []string{},
						})
						if err != nil {
							log.Printf("发送邮件失败: %v", err)
							return
						}
						return
					}
					defer os.RemoveAll(cacheDir)

					// try to download attachments (limit size: 15MB)
					var downloadedPaths []string
					var errorHints []string
					limit := 15
					for _, attachment := range detail.Attachments {
						safeName := common.CleanFileName(attachment.Title)
						savePath := filepath.Join(cacheDir, safeName)

						err := bridge.DownloadFromPython(&bridge.DownloadOptions{
							Client:   bridge.Client(fetchCtx.Client),
							Account:  fetchCtx.Account,
							Password: fetchCtx.Password,
							URL:      attachment.URL,
							MaxSize:  limit,
							SavePath: savePath,
							Referer:  notice.URL,
							Extra:    fetchCtx.Extra,
						})
						if err != nil {
							hint := "缺失附件: " + attachment.Title
							errorHints = append(errorHints, hint)
						} else {
							downloadedPaths = append(downloadedPaths, savePath)
						}
					}

					if len(errorHints) > 0 {
						finalHint := strings.Join(errorHints, "\n")
						log.Println("下载摘要:\n", finalHint)

						body += "\n\n———\n附件下载提示：\n" + finalHint
					}

					err = bridge.SendMail(&bridge.SendOptions{
						SMTPServer:  global.SMTPSERVER,
						Account:     global.ACCOUNT,
						AuthCode:    global.AUTHCODE,
						Subject:     "[NotiCat]" + common.ShortenTitle(notice.Title),
						Body:        body,
						From:        global.ACCOUNT,
						To:          user.Email,
						Attachments: downloadedPaths,
					})
					if err != nil {
						log.Printf("发送邮件失败: %v", err)
						return
					}
				}
			}()
		}
	}
}

func FetchByTaskID(taskID uint) (*FetchContext, []bridge.Notice, error) {
	var task model.FetchTask
	if err := global.DB.First(&task, taskID).Error; err != nil {
		return nil, nil, err
	}

	// Inherit context
	fetchCtx, notices, err := FetchByConfig(task.Client, task.Credentials, task.Extra)
	if err != nil {
		return nil, nil, err
	}

	if err := global.DB.Model(&model.FetchTask{}).Where("id = ?", taskID).Update("last_fetch_at", time.Now()).Error; err != nil {
		log.Printf("Warning: 无法更新任务 %d 的抓取时间: %v", taskID, err)
	}

	return fetchCtx, notices, nil
}

func FetchByConfig(client string, credentials string, extra string) (*FetchContext, []bridge.Notice, error) {
	// client
	rawClient := strings.ToLower(client)
	clientType := bridge.Client(rawClient)
	if !clientType.IsValid() {
		return nil, nil, fmt.Errorf("数据库存储了错误的client: %s", client)
	}

	// credentials
	var creds map[string]any
	err := json.Unmarshal([]byte(credentials), &creds)
	if err != nil {
		return nil, nil, fmt.Errorf("credentials解析失败: %v", err)
	}
	account, _ := creds["account"].(string)
	password, _ := creds["password"].(string)

	// extra
	var ext map[string]any
	err = json.Unmarshal([]byte(extra), &ext)
	if err != nil {
		return nil, nil, fmt.Errorf("extra解析失败: %v", err)
	}

	notices, err := bridge.FetchFromPython(&bridge.FetchOptions{
		Client:   clientType,
		Account:  account,
		Password: password,
		Extra:    ext,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("python执行失败: %v", err)
	}

	return &FetchContext{
		Client:   rawClient,
		Account:  account,
		Password: password,
		Extra:    ext,
	}, notices, nil
}
