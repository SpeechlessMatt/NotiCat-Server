# 🐱 NotiCat Server

<p align="center">
  <img src="https://img.shields.io/badge/version-0.1.2-blue?style=for-the-badge" alt="version">
  <img src="https://img.shields.io/badge/build-2026.02.02-green?style=for-the-badge" alt="build">
</p>

<p align="center">
  <b>Notification bridge server</b><br>
  自动抓取各类站点更新，推送到你的客户端
</p>

---

## 支持的客户端

### 📺 BiliClient (`bili`)
**B站UP主动态监控**

盯着你喜欢的UP主，一旦发动态或视频就立刻通知你。

- **需要配置**：订阅时请在「额外信息」中填入UP主的个人空间动态页链接（例如 `https://space.bilibili.com/xxxx/dynamic`）
- **无需凭证**：不需要登录B站账号，公开链接即可抓取

### 🎓 BUPTClient (`bupt`)
**北邮信息门户通知抓取**

自动登录 `my.bupt.edu.cn`，抓取校内最新通知，配合正则筛选功能过滤你关心的内容（奖学金、选课、讲座等）。

- **需要凭证**：提供你的学号和密码（北邮统一认证）
- **无需额外配置**：订阅时无需填写额外字段

### 🏁 SaikrClient (`saikr`)
**赛氪赛事中心监控**

打听你关注的赛事，比如大学生英语竞赛，四六级哦

- **无需额外配置**：不需要任何额外配置
- **无需凭证**：不需要登录任何账号

---

## 服务信息

- **版本**：`0.1.2`
- **构建时间**：`2026-02-02`
- **维护者**：`edbinmatt`

---

<p align="center">
  <sub>🐾 NotiCat - 让通知自己找上门</sub>
</p>
