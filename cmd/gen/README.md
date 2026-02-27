# 🐱 NotiCat Server

<p align="center">
  <img src="https://img.shields.io/badge/version-0.1.2-blue?style=for-the-badge" alt="version">
  <img src="https://img.shields.io/badge/build-2026.02.11-green?style=for-the-badge" alt="build">
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

- **需要配置**：订阅时请在「额外信息」中填入UP主的UID（例如 `1636034895`）
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

### 🔢 CMathcClient (`cmathc`)
**大学生数学竞赛网新闻动态抓取**

关注数学竞赛吗？如果关注数学建模竞赛或者大学生数学竞赛都可以订阅哦

- **无需额外配置**：不需要任何额外配置
- **无需凭证**：不需要登录任何账号

### 🆚 NuedcClient (`nuedc`)
**全国大学生电子设计竞赛网相关通知抓取**

虽然学校通知一般会发的，不过如果关注电赛的话可以订阅哦

- **无需额外配置**：不需要任何额外配置
- **无需凭证**：不需要登录任何账号

### 👩‍💼 BVFClient (`bvf`)
**志愿北京志愿项目抓取**

做志愿必备工具，抓取志愿北京的志愿项目，为社会奉献哦

- **需要配置**：订阅时请在「额外信息」中填入筛选后的网址哦（例如 `https://www.bv2008.cn/app/opp/list.php?tag=&area=2464&area2=2516&state=2&scope=&obj=&time_start=&time_end=&name=&members=&mode=list`）,不填URL默认抓取的是主页志愿项目
- **无需凭证**：不需要登录任何账号

---

## 服务信息

- **版本**：`0.1.2`
- **构建时间**：`2026-02-11`
- **维护者**：`edbinmatt`

---

<p align="center">
  <sub>🐾 NotiCat - 让通知自己找上门</sub>
</p>
