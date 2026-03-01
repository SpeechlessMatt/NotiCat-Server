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

// Package meta say sth to user
package meta

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
)

func RegisterRoutes(r *gin.Engine, baseDir string) {
	r.GET("info", func(c *gin.Context) {
		data, err := os.ReadFile(filepath.Join(baseDir, "info.json"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "info not available"})
			return
		}
		c.Data(http.StatusOK, "application/json; charset=utf-8", data)
	})

	r.GET("readme", func(c *gin.Context) {
		md, err := os.ReadFile(filepath.Join(baseDir, "README.md"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "README not available"})
			return
		}

		html := markdown.ToHTML(md, nil, nil)

		css := `
			:root {
				--primary-color: #2563eb;
				--text-main: #1f2937;
				--text-muted: #4b5563;
				--bg-body: #f8fafc;
				--bg-card: #ffffff;
				--border-color: #e5e7eb;
				--code-bg: #f1f5f9;
			}

			* { box-sizing: border-box; }

			body { 
				font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; 
				line-height: 1.6; 
				margin: 0;
				padding: 0; 
				background-color: var(--bg-body); 
				color: var(--text-main);
				display: flex;
				justify-content: center;
			}

			/* 核心容器：移动端适配的关键 */
			.container {
				width: 100%;
				max-width: 800px;
				margin: 2rem 1rem;
				padding: 2.5rem;
				background: var(--bg-card);
				border-radius: 12px;
				box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
			}

			/* 移动端间距调整 */
			@media (max-width: 640px) {
				.container {
					margin: 0;
					padding: 1.5rem;
					border-radius: 0;
				}
				body { background-color: var(--bg-card); }
			}

			h1, h2, h3 { color: #111827; margin-top: 1.5em; font-weight: 700; }
			h1 { font-size: 2.25rem; border-bottom: 2px solid var(--border-color); padding-bottom: 0.5rem; }
			
			a { color: var(--primary-color); text-decoration: none; border-bottom: 1px transparent; transition: border 0.2s; }
			a:hover { border-bottom: 1px solid var(--primary-color); }

			/* 代码块美化 */
			code { 
				background: var(--code-bg); 
				padding: 0.2rem 0.4rem; 
				border-radius: 6px; 
				font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; 
				font-size: 0.9em;
				color: #eb5757;
			}
			
			pre { 
				background: #1e293b; /* 深色主题代码块更有质感 */
				color: #f8fafc;
				padding: 1.25rem; 
				border-radius: 8px; 
				overflow-x: auto; 
				line-height: 1.5;
			}
			
			pre code { 
				background: transparent; 
				color: inherit; 
				padding: 0; 
			}

			img { max-width: 100%; height: auto; border-radius: 8px; }
			
			blockquote {
				margin: 1rem 0;
				padding-left: 1rem;
				border-left: 4px solid var(--border-color);
				color: var(--text-muted);
				font-style: italic;
			}
		`
		page := []byte(`<!DOCTYPE html>
		<html lang="zh-CN">
		<head>
			<meta charset="utf-8">
			<meta name="viewport" content="width=device-width, initial-scale=1">
			<title>Server README</title>
			<style>` + css + `</style>
		</head>
		<body>
			<div class="container">` + string(html) + `</div>
		</body>
		</html>`)

		c.Data(http.StatusOK, "text/html; charset=utf-8", page)
	})
}
