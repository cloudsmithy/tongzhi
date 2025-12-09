# 微信通知系统

基于微信测试号的消息推送系统，支持多模板管理、动态字段和 Webhook API 调用。

## 功能特性

- 🔔 微信模板消息推送
- � 多模板管理理（支持多个微信模板）
- � 动e态 keyword 字段（first, keyword1, keyword2...）
- �  多接收者管理
- 🔗 Webhook API 支持
- � Token 认证
- ⚡ 并发消息发送
- 🛡️ 请求频率限制

## 技术栈

- **后端**: Go + Gin + SQLite
- **前端**: React + TypeScript + Vite

## 快速开始

### 1. 配置微信测试号

1. 访问 [微信公众平台测试号](https://mp.weixin.qq.com/debug/cgi-bin/sandbox?t=sandbox/login)
2. 获取 AppID 和 AppSecret
3. 创建消息模板，例如：
   ```
   {{first.DATA}}
   订单编号：{{keyword1.DATA}}
   金额：{{keyword2.DATA}}
   {{remark.DATA}}
   ```

### 2. 启动后端

```bash
cd backend
cp .env.example .env
# 编辑 .env 填入配置（或在 Web 界面配置）

go run main.go
```

后端运行在 `http://localhost:8080`

### 3. 启动前端

```bash
cd frontend
npm install
npm run dev
```

前端运行在 `http://localhost:5173`

## 使用方法

### Web 界面

1. 访问 `http://localhost:5173`
2. 在设置页面：
   - 配置微信测试号信息（AppID、AppSecret）
   - 添加消息模板（模板标识 + 微信模板ID）
   - 添加接收者（需要用户关注测试号获取 OpenID）
3. 在首页：
   - 选择模板
   - 填写模板字段（first, keyword1, keyword2...）
   - 选择接收者发送

### Webhook API

```bash
# 生成 Token（在设置页面操作）

# 发送消息
curl -X POST http://localhost:8080/webhook/send \
  -H "Authorization: Bearer <your-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "templateKey": "订单通知",
    "keywords": {
      "first": "您有一条新订单",
      "keyword1": "订单号123456",
      "keyword2": "￥100.00",
      "remark": "感谢您的支持"
    },
    "recipientIds": [1, 2]
  }'
```

参数说明：
- `templateKey` - 模板标识（必填，在设置页面添加的模板标识）
- `keywords` - 模板字段（必填，key-value 格式，对应微信模板的字段）
- `recipientIds` - 接收者 ID 列表（可选，不传则发送给所有接收者）

## 项目结构

```
├── backend/
│   ├── handlers/     # HTTP 处理器
│   ├── middleware/   # 中间件（CORS、认证、限流）
│   ├── models/       # 数据模型
│   ├── repository/   # 数据库操作
│   ├── services/     # 业务逻辑
│   └── config/       # 配置管理
├── frontend/
│   ├── src/
│   │   ├── components/  # React 组件
│   │   ├── pages/       # 页面
│   │   ├── services/    # API 调用
│   │   └── types/       # TypeScript 类型
│   └── ...
└── README.md
```

## License

MIT
