# NotiCat Server

<div align="center">

![License](https://img.shields.io/github/license/SpeechlessMatt/NotiCat-Server)
![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![Python Version](https://img.shields.io/badge/Python-3.13+-3776AB?logo=python)
![C++ Version](https://img.shields.io/badge/C++-17+-00599C?logo=c%2B%2B)

[NotiCat Android 客户端](https://github.com/SpeechlessMatt/NotiCat-Android) · [问题反馈](https://github.com/SpeechlessMatt/NotiCat-Server/issues)

</div>

## 🌟 项目简介

NotiCat Server 是一个高效的网页内容监控与通知聚合系统。它为没有主动通知功能的网站添加监控能力，通过智能去重和规则筛选，将用户感兴趣的内容通过邮件（或其他可替换的通道）实时推送给用户。

## ✨ 核心特性

- **多语言混合架构**：Go 作为主框架，C++ 处理邮件发送，Python 负责网页抓取
- **智能请求去重**：通过数据库聚合用户订阅，对同一资源只抓取一次，分发多个用户
- **高度解耦设计**：邮件模块和抓取模块均可独立替换
- **灵活规则筛选**：支持正则表达式和关键字过滤，精准匹配用户需求
- **便捷客户端扩展**：添加新网站支持仅需简单配置和 Python 脚本

## 🏗️ 系统架构

NotiCat-Server/
├── go/ # 主框架 (Gin + GORM)
├── cpp/ # 邮件发送模块 (libcurl)
└── scripts/ # Python 抓取脚本
└── clients/ # 各网站客户端实现


## 📦 快速开始

### 环境要求

> **注意**：请确保已安装以下依赖（版本要求可能随项目更新变化）

- **Go** 1.21+ (主框架)
- **Python** 3.13+ (网页抓取)
- **C++17** 兼容编译器 (邮件发送)
- **Redis** (任务队列与缓存)
- **Make** (构建工具)

### 构建与安装

```bash
# 克隆项目
git clone https://github.com/SpeechlessMatt/NotiCat-Server.git
cd NotiCat-Server

# 完整构建（编译所有模块）
make all

# 配置与代码生成（更新客户端支持）
make gem
```

## 🔧 配置说明

### 客户端配置文件 (clients.json)

位于 cmd/gem/clients.json，定义了所有支持的网站客户端：

```json
{
  "name": "NotiCat Server (Main)",
  "version": "0.1.2",
  "owner": "edbinmatt",
  "description": "Notification bridge server",
  "support_clients": [
    {
      "client": "bili",
      "name": "BiliClient",
      "url": "https://www.bilibili.com",
      "description": "B站UP主动态订阅",
      "credentials": [],
      "extra": [
        {
          "label": "URL",
          "api_key": "url"
        }
      ]
    }
  ]
}
```

**字段说明**：

- **client**: 客户端标识符（对应 Python 脚本）

- **name**: 显示名称

- **url**: 目标网站URL

- **description**: 功能描述

- **credentials**: 所需认证字段（如用户名/密码：["username", "password"]）

- **extra**: 所需额外参数，传递给 Python 脚本

## 🚀 添加新客户端

扩展 NotiCat 以支持新网站非常简单，只需两步：

### 步骤1：修改配置文件

在 clients.json 的 support_clients 数组中添加新条目：

```json
{
  "client": "example",
  "name": "ExampleClient",
  "url": "https://example.com",
  "description": "示例网站监控",
  "credentials": ["username", "password"],
  "extra": [
    {
      "label": "喜好",
      "api_key": "like"
    },
    {
      "label": "页码",
      "api_key": "page"
    }
  ]
}
```

### 步骤2：创建 Python 客户端

在 scripts/clients/ 目录下创建新的 Python 文件：

```python
from .base import BaseClient

class ExampleClient(BaseClient):
    # client_id 会自动从类名生成（移除"Client"并转为小写）
    # 即：ExampleClient -> "example"
    
    def __init__(self, username, password, extra) -> None:
        super().__init__(username=username, password=password, extra=extra)
    
    async def fetch(self) -> list:
        """实现抓取逻辑，返回消息列表"""
        # 您的抓取代码
        messages = []
        # ... 抓取逻辑
        return messages
```

### 步骤3：应用更改

项目根目录运行

```bash
# 运行 make gem 更新配置
make gem
```

完成！ 新的客户端已集成到系统中。Go 框架会自动调用：

```bash
python scripts/catcher.py example 用户名 密码 --extra ...
```

## 📡 运行机制

### 任务调度流程

1. 订阅聚合：系统收集所有用户对同一资源的订阅

2. 智能抓取：对每个资源只执行一次抓取操作

3. 规则过滤：根据用户设置的正则/关键字规则筛选内容

4. 分发推送：将匹配的内容通过邮件（或其他通道）发送给相应用户

### 模块调用关系

用户请求 → Go主框架 → 任务调度 → Python抓取 → C++邮件发送
                ↓
            数据库记录
                ↓
          用户规则匹配

## 🛠️ 开发与部署

### 开发环境

先不写

### 生产部署

先不写

## 🤝 贡献指南

我们欢迎各种形式的贡献！请参阅以下步骤：

1. Fork 本仓库

2. 创建功能分支 (git checkout -b feature/amazing-feature)

3. 提交更改 (git commit -m 'Add some amazing feature')

4. 推送分支 (git push origin feature/amazing-feature)

5. 开启 Pull Request

### 贡献类型

- 添加新的网站客户端

- 改进现有抓取逻辑

- 优化系统性能

- 修复 Bug

- 完善文档

## 📄 许可证

先不写

<div align="center"> <sub>Built with ❤️ by the NotiCat Team</sub> </div>

