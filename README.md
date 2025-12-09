<p align="center">
  <img src="https://img.icons8.com/color/96/weixing.png" alt="WeChat Logo" width="80"/>
</p>

<h1 align="center">📮 微信通知系统</h1>

<p align="center">
  <strong>基于微信测试号的消息推送系统</strong><br>
  支持多模板管理 · 动态字段 · Webhook API
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/React-18+-61DAFB?style=flat-square&logo=react" alt="React">
  <img src="https://img.shields.io/badge/TypeScript-5+-3178C6?style=flat-square&logo=typescript" alt="TypeScript">
  <img src="https://img.shields.io/badge/SQLite-3+-003B57?style=flat-square&logo=sqlite" alt="SQLite">
  <img src="https://img.shields.io/badge/License-MIT-green?style=flat-square" alt="License">
</p>

---

## ✨ 功能特性

<table>
  <tr>
    <td>🔔 <b>模板消息</b></td>
    <td>基于微信模板消息 API，支持富文本格式</td>
  </tr>
  <tr>
    <td>📋 <b>多模板管理</b></td>
    <td>创建和管理多个消息模板，灵活切换</td>
  </tr>
  <tr>
    <td>🔤 <b>动态字段</b></td>
    <td>支持 first、keyword1~n、remark 等动态字段</td>
  </tr>
  <tr>
    <td>👥 <b>接收者管理</b></td>
    <td>批量管理消息接收者，支持分组发送</td>
  </tr>
  <tr>
    <td>🔗 <b>Webhook API</b></td>
    <td>RESTful API 接口，轻松集成到任何系统</td>
  </tr>
  <tr>
    <td>🔑 <b>Token 认证</b></td>
    <td>安全的 Bearer Token 认证机制</td>
  </tr>
  <tr>
    <td>⚡ <b>并发发送</b></td>
    <td>高效的并发消息发送，支持批量操作</td>
  </tr>
  <tr>
    <td>🛡️ <b>频率限制</b></td>
    <td>内置请求频率限制，防止滥用</td>
  </tr>
</table>

---

## 🚀 快速开始

### 📝 1. 配置微信测试号

1. 访问 [微信公众平台测试号](https://mp.weixin.qq.com/debug/cgi-bin/sandbox?t=sandbox/login)
2. 扫码登录，获取 **AppID** 和 **AppSecret**
3. 新增测试模板，例如：

```
{{first.DATA}}
订单编号：{{keyword1.DATA}}
金额：{{keyword2.DATA}}
时间：{{keyword3.DATA}}
{{remark.DATA}}
```

### 🖥️ 2. 启动后端

```bash
cd backend
cp .env.example .env
# 编辑 .env 或在 Web 界面配置

go run main.go
```

> 🟢 后端运行在 `http://localhost:8080`

### 🎨 3. 启动前端

```bash
cd frontend
npm install
npm run dev
```

> 🟢 前端运行在 `http://localhost:5173`

---

## 📖 使用指南

### 🌐 Web 界面

| 步骤 | 操作 |
|------|------|
| 1️⃣ | 访问 `http://localhost:5173` |
| 2️⃣ | **设置页面** → 配置微信 AppID/AppSecret |
| 3️⃣ | **设置页面** → 添加消息模板 |
| 4️⃣ | **设置页面** → 添加接收者 OpenID |
| 5️⃣ | **Webhook 页面** → 生成 API Token |
| 6️⃣ | **首页** → 选择模板、填写内容、发送！ |

### 🔌 Webhook API

```bash
curl -X POST http://localhost:5173/api/webhook/send \
  -H "Authorization: Bearer <your-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "templateKey": "测试",
    "keywords": {
      "first": "您有一条新消息",
      "keyword1": "订单编号: 2024120901",
      "keyword2": "金额: ¥99.00",
      "keyword3": "时间: 2024-12-09 15:30",
      "remark": "点击查看详情"
    },
    "recipientIds": [1]
  }'
```

#### 📋 参数说明

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `templateKey` | string | ✅ | 模板名称（设置页面添加的） |
| `keywords` | object | ✅ | 模板字段，key-value 格式 |
| `recipientIds` | number[] | ❌ | 接收者 ID，不传则发送给所有人 |

---

## 📁 项目结构

```
📦 wechat-notification
├── 🔧 backend/
│   ├── handlers/      # HTTP 处理器
│   ├── middleware/    # 中间件
│   ├── models/        # 数据模型
│   ├── repository/    # 数据库操作
│   ├── services/      # 业务逻辑
│   └── config/        # 配置管理
│
├── 🎨 frontend/
│   └── src/
│       ├── components/   # React 组件
│       ├── pages/        # 页面
│       ├── services/     # API 调用
│       └── types/        # TypeScript 类型
│
└── 📄 README.md
```

---

## 📄 License

MIT © 2024

---

<p align="center">
  <sub>Made with ❤️ for WeChat developers</sub>
</p>
