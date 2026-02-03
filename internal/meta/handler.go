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
        body { font-family: "Helvetica Neue", Helvetica, Arial, sans-serif; line-height: 1.6; padding: 2rem; background: #f5f5f5; color: #333; }
        h1,h2,h3,h4 { color: #2c3e50; }
        code { background: #eee; padding: 2px 4px; border-radius: 3px; font-family: monospace; }
        pre { background: #eee; padding: 1rem; border-radius: 5px; overflow-x: auto; }
        a { color: #3498db; text-decoration: none; }
        a:hover { text-decoration: underline; }
        `

		page := []byte(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Server README</title>
	<style>` + css + `</style>
</head>
<body>` + string(html) + `</body>
</html>`)

		c.Data(http.StatusOK, "text/html; charset=utf-8", page)
	})
}
