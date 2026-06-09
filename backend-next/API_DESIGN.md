# New Haven — API Design (v1)

Base URL: `/A1`

All endpoints return JSON.  
Auth: Bearer JWT in `Authorization` header (except auth endpoints).  
Pagination: `?page=1&per_page=20`.  
Errors: `{ "error": { "code": "NOT_FOUND", "message": "..." } }`.

---

## 1. Auth

### POST /A1/auth/register

Create a new account. Returns a JWT token directly on success (no OTP unless configured).

```
Request:
{
  "email": "string",
  "password": "string",
  "displayName": "string",
  "companyName": "string",
  "inviteCode": "string|null"
}

Response 201:
{
  "token": "jwt...",
  "user": { "id": "uuid", "email": "...", "displayName": "..." }
}

Response 200 (if OTP required):
{
  "step": "otp",
  "message": "Verification code sent to email"
}
```

### POST /A1/auth/verify-otp

```
Request:
{
  "email": "string",
  "otpCode": "string"
}

Response 200:
{
  "token": "jwt...",
  "user": { ... }
}
```

### POST /A1/auth/resend-otp

```
Request:
{
  "email": "string"
}

Response 200:
{ "message": "Code resent" }
```

### POST /A1/auth/login

```
Request:
{
  "email": "string",
  "password": "string"
}

Response 200:
{
  "token": "jwt...",
  "user": { "id": "uuid", "email": "...", "displayName": "...", "avatar": "..." }
}
```

### POST /A1/auth/logout

Invalidates the current token.

```
Headers: Authorization: Bearer <token>

Response 200:
{ "message": "ok" }
```

### POST /A1/auth/refresh

```
Request:
{
  "refreshToken": "string"
}

Response 200:
{
  "token": "jwt...",
  "refreshToken": "string"
}
```

### POST /A1/auth/forgot-password

```
Request:
{
  "email": "string"
}

Response 200:
{ "message": "If account exists, reset link sent" }
```

### POST /A1/auth/reset-password

```
Request:
{
  "resetToken": "string",
  "newPassword": "string"
}

Response 200:
{ "message": "Password updated" }
```

### GET /A1/auth/me

```
Headers: Authorization: Bearer <token>

Response 200:
{
  "id": "uuid",
  "email": "...",
  "displayName": "...",
  "avatar": "...",
  "createdAt": "iso8601"
}
```

---

## 2. Player

### GET /A1/player

Current player state — used by TopBar, SettingsDrawer.

```
Response 200:
{
  "id": "uuid",
  "name": "Captain Mochi",
  "avatar": "🧑🍳",
  "level": 5,
  "xp": 320,
  "xpMax": 500,
  "cash": 12500,
  "companyName": "Mochi Foods Co.",
  "notifications": 3,
  "boosts": ["🌾 Spring Boost +15%"]
}
```

### PATCH /A1/player

Update profile.

```
Request:
{
  "displayName": "string",
  "avatar": "string",
  "companyName": "string"
}

Response 200: { ...player }
```

### GET /A1/player/profile

Aggregated profile page data.

```
Response 200:
{
  "player": { ...same as GET /A1/player },
  "stats": {
    "totalBuildings": 12,
    "totalProduced": 4520,
    "totalEarned": 87500,
    "totalOrders": 68,
    "achievementsUnlocked": 7
  },
  "recentActivity": [
    { "type": "production", "text": "Farm produced 120 Wheat", "time": "2m ago" },
    { "type": "order", "text": "Sell order filled: 30 Bread", "time": "15m ago" }
  ],
  "buildings": [ ... ],
  "inventory": [ ... ],
  "executives": [ ... ],
  "research": [ ... ],
  "finance": { ... }
}
```

---

## 3. Buildings

### GET /A1/buildings

Building catalog — static definitions for the Build shop.

```
Response 200:
[
  {
    "id": "farm",
    "name": "Farm",
    "icon": "🌾",
    "description": "Grows wheat & corn",
    "category": "production",
    "price": 500,
    "produces": ["wheat", "corn"],
    "unlockLevel": 1,
    "upgrades": [
      { "level": 1, "cost": 300, "effect": "+20% yield" }
    ]
  },
  ...
]
```

### GET /A1/player/buildings

Player's built buildings on the map.

```
Response 200:
[
  {
    "id": "uuid",
    "buildingId": "farm",
    "plotId": "harbor-03",
    "name": "Farm",
    "icon": "🌾",
    "level": 2,
    "status": "idle" | "producing" | "ready",
    "production": {
      "resource": "wheat",
      "startedAt": "iso8601",
      "completesAt": "iso8601",
      "progress": 0.65
    } | null,
    "assignedExecutive": { "id": "uuid", "name": "...", "avatar": "..." } | null
  },
  ...
]
```

### POST /A1/player/buildings

Build a new building on a plot.

```
Request:
{
  "buildingId": "farm",
  "plotId": "harbor-03"
}

Response 201:
{ "id": "uuid", ...building }
```

### DELETE /A1/player/buildings/:id

Demolish a building.

```
Response 200:
{ "message": "demolished" }
```

### PATCH /A1/player/buildings/:id/move

Move building to another plot.

```
Request:
{
  "newPlotId": "harbor-07"
}

Response 200:
{ ...building }
```

### GET /A1/player/buildings/:id

Detailed building state — drives BuildingDetailPanel.

```
Response 200:
{
  "id": "uuid",
  "buildingId": "farm",
  "plotId": "harbor-03",
  "level": 2,
  "status": "idle" | "producing" | "ready",
  "production": { ... } | null,
  "availableCrops": [
    { "id": "wheat", "icon": "🌾", "name": "Wheat", "duration": 1800, "yield": 100 }
  ],
  "assignedExecutive": { ... } | null,
  "upgradeCost": 800,
  "upgradeEffects": "+20% yield"
}
```

---

## 4. Regions & Plots

### GET /A1/regions

All regions with plot states.

```
Response 200:
[
  {
    "id": "harbor",
    "name": "New Harbor",
    "icon": "⚓",
    "emoji": "🌊",
    "description": "Coastal harbor town...",
    "bonus": "+5% Trading Speed",
    "bonusColor": "text-blue-600",
    "suggestedBuildings": ["Farm", "Market Stall", "Trading Hub", "Warehouse"],
    "locked": false,
    "unlockRequirement": null,
    "slotsUsed": 4,
    "slotsTotal": 12,
    "plots": [
      {
        "id": "harbor-01",
        "x": 0, "y": 0,
        "state": "occupied" | "available" | "locked",
        "building": { ... } | null,
        "lockedReason": null
      },
      ...
    ]
  },
  ...
]
```

---

## 5. Market

### GET /A1/market/prices

Current market snapshot — for PriceTicker + MarketPage resource list.

```
Response 200:
[
  {
    "resource": "wheat",
    "price": 12.5,
    "change24h": 2.3,
    "high24h": 14.0,
    "low24h": 11.0,
    "volume24h": 4520,
    "buyPrice": 12.8,
    "sellPrice": 12.2,
    "inventory": 340
  },
  ...
]
```

### GET /A1/market/prices/:resource?range=1h|6h|12h|24h|48h|7d

Historical price data for charts.

```
Response 200:
{
  "resource": "wheat",
  "range": "24h",
  "points": [
    { "time": "iso8601", "price": 12.3, "volume": 120 },
    { "time": "iso8601", "price": 12.8, "volume": 85 },
    ...
  ]
}
```

### GET /A1/market/orderbook/:resource

Buy/sell order book and depth.

```
Response 200:
{
  "resource": "wheat",
  "bids": [
    { "price": 12.0, "quantity": 200, "total": 2400 },
    { "price": 11.8, "quantity": 150, "total": 1770 },
    ...
  ],
  "asks": [
    { "price": 12.6, "quantity": 180, "total": 2268 },
    { "price": 12.8, "quantity": 300, "total": 3840 },
    ...
  ]
}
```

### GET /A1/market/orders?status=open|filled|cancelled

User's own orders.

```
Response 200:
[
  {
    "id": "uuid",
    "resource": "wheat",
    "type": "buy" | "sell",
    "price": 12.0,
    "quantity": 100,
    "filled": 40,
    "status": "open" | "filled" | "cancelled" | "partial",
    "createdAt": "iso8601"
  },
  ...
]
```

### POST /A1/market/orders

Place an order.

```
Request:
{
  "resource": "wheat",
  "type": "buy" | "sell",
  "price": 12.0,
  "quantity": 100
}

Response 201:
{ "id": "uuid", ...order }
```

### DELETE /A1/market/orders/:id

Cancel an open order.

```
Response 200:
{ "message": "cancelled" }
```

---

## 6. Warehouse

### GET /A1/warehouse

Current inventory.

```
Response 200:
{
  "capacity": { "used": 340, "total": 500 },
  "items": [
    { "resource": "wheat", "quantity": 120, "quality": 92 },
    { "resource": "flour", "quantity": 45, "quality": 88 },
    ...
  ]
}
```

### POST /A1/warehouse/transfer

Move items (e.g., from warehouse to market listing).

```
Request:
{
  "resource": "wheat",
  "quantity": 50,
  "destination": "market" | "production"
}

Response 200:
{ ...updated warehouse }
```

---

## 7. Finance

### GET /A1/finance/summary

Financial dashboard data.

```
Response 200:
{
  "cash": 12500,
  "revenue": { "today": 3200, "week": 18400, "month": 72000 },
  "expenses": { "today": 1200, "week": 8100, "month": 34000 },
  "assets": {
    "buildings": 22000,
    "inventory": 12000,
    "cash": 8000
  },
  "profitMargin": 0.34
}
```

### GET /A1/finance/cashflow?period=week|month

Daily/weekly cashflow data for chart.

```
Response 200:
{
  "period": "week",
  "data": [
    { "date": "2026-06-02", "income": 2100, "expenses": 1200 },
    { "date": "2026-06-03", "income": 1800, "expenses": 900 },
    ...
  ]
}
```

---

## 8. Production

### GET /A1/production/jobs?status=active|completed

Production jobs across all buildings.

```
Response 200:
[
  {
    "id": "uuid",
    "buildingId": "uuid",
    "buildingName": "Farm (Plot #1)",
    "buildingIcon": "🌾",
    "resource": "wheat",
    "quantity": 120,
    "quality": 92,
    "status": "in_progress" | "ready" | "claimed",
    "startedAt": "iso8601",
    "completesAt": "iso8601",
    "completedAt": "iso8601|null"
  },
  ...
]
```

### POST /A1/player/buildings/:id/produce

Start a production job.

```
Request:
{
  "recipe": "wheat",
  "quantity": 100
}

Response 201:
{ "job": { ...job } }
```

### POST /A1/production/jobs/:id/claim

Claim completed production.

```
Response 200:
{
  "job": { ...job, status: "claimed" },
  "itemsAdded": { "wheat": 120 }
}
```

---

## 9. Chat

### WebSocket /A1/ws?token=<jwt>

Real-time messaging. Protocol: JSON messages over WebSocket.

```
→ Client sends:
{
  "type": "join_channel",
  "channelId": "general"
}

← Server sends:
{
  "type": "message",
  "id": "msg-123",
  "channelId": "general" | "dm:user-456",
  "from": { "id": "uuid", "name": "...", "avatar": "..." },
  "content": "Hello!",
  "timestamp": "iso8601"
}

← Server sends (market tick):
{
  "type": "price_update",
  "resource": "wheat",
  "price": 12.8,
  "change24h": 2.3
}

← Server sends (notification):
{
  "type": "notification",
  "id": "uuid",
  "category": "production_ready" | "order_filled" | "contract_expiring",
  "message": "Farm #1 is ready to harvest!",
  "data": { "buildingId": "..." }
}
```

### GET /A1/chat/channels

Chat channel list.

```
Response 200:
[
  { "id": "general", "label": "General", "emoji": "💬", "unread": 3 },
  { "id": "sales", "label": "Sales", "emoji": "🏷️", "unread": 0 },
  { "id": "help", "label": "Help", "emoji": "❓", "unread": 1 }
]
```

### GET /A1/chat/channels/:id/messages?before=<cursor>&limit=50

Channel message history.

```
Response 200:
{
  "messages": [
    { "id": "uuid", "from": { ... }, "content": "...", "timestamp": "iso8601" }
  ],
  "hasMore": true
}
```

### GET /A1/chat/dms

DM contact list.

```
Response 200:
[
  {
    "id": "uuid",
    "name": "Atlas Trading Bot",
    "avatar": "🤖",
    "role": "Trade Assistant",
    "online": true,
    "lastMessage": "Wheat is rising!",
    "lastMessageTime": "2m ago",
    "unread": 1
  },
  ...
]
```

### GET /A1/chat/dms/:id/messages?before=<cursor>&limit=50

DM message history.

```
Same shape as channel messages.
```

### POST /A1/chat/messages

Send a message (fallback when WebSocket not connected).

```
Request:
{
  "channelId": "general" | null,
  "recipientId": "uuid" | null,
  "content": "Hello!"
}

Response 201:
{ "id": "uuid", ... }
```

---

## 10. Contracts

### GET /A1/contracts?status=active|completed|available

```
Response 200:
[
  {
    "id": "uuid",
    "client": "Harbor Inn 🏨",
    "items": [
      { "resource": "bread", "quantity": 30 },
      { "resource": "butter", "quantity": 10 }
    ],
    "reward": 1500,
    "deadline": "iso8601",
    "status": "active" | "completed" | "available",
    "progress": 0.6,
    "penalty": 300
  },
  ...
]
```

### POST /A1/contracts/:id/accept

```
Response 200:
{ ...contract, status: "active" }
```

### POST /A1/contracts/:id/fulfill

Deliver contract items from warehouse.

```
Response 200:
{
  "contract": { ...contract, status: "completed" },
  "reward": 1500,
  "itemsDeducted": { "bread": 30, "butter": 10 }
}
```

---

## 11. Research

### GET /A1/research/nodes

Research tree.

```
Response 200:
[
  {
    "id": "efficient_farming",
    "name": "Efficient Farming",
    "icon": "🌱",
    "description": "Farm production +15%",
    "cost": { "cash": 2000, "researchPoints": 100 },
    "duration": 3600,
    "status": "locked" | "available" | "in_progress" | "completed",
    "progress": 0 | 0.5,
    "prerequisites": [],
    "category": "farming" | "processing" | "commerce" | "logistics",
    "tier": 1
  },
  ...
]
```

### POST /A1/research/:id/start

Begin researching a node.

```
Response 200:
{ ...node, status: "in_progress", startedAt: "iso8601", completesAt: "iso8601" }
```

### POST /A1/research/:id/claim

Claim completed research.

```
Response 200:
{ ...node, status: "completed" }
```

---

## 12. Executives

### GET /A1/executives?status=available|hired

```
Response 200:
[
  {
    "id": "uuid",
    "name": "Chef Marco",
    "avatar": "👨🍳",
    "role": "Production Manager",
    "skill": "+20% Kitchen speed",
    "salary": 500,
    "status": "available" | "hired",
    "assignedTo": { "buildingId": "uuid", "buildingName": "Kitchen" } | null,
    "morale": 78,
    "contractEnd": "iso8601"
  },
  ...
]
```

### POST /A1/executives/:id/hire

```
Request:
{
  "salary": 500
}

Response 200:
{ ...executive, status: "hired" }
```

### POST /A1/executives/:id/assign

Assign to a building.

```
Request:
{
  "buildingId": "uuid"
}
```

### POST /A1/executives/:id/fire

```
Response 200:
{ ...executive, status: "available" }
```

---

## 13. Leaderboard

### GET /A1/leaderboard?period=all_time|weekly|monthly

```
Response 200:
{
  "period": "all_time",
  "myRank": 42,
  "entries": [
    { "rank": 1, "playerName": "FoodKing", "avatar": "👑", "level": 50, "score": 125000, "companyName": "Royal Foods" },
    { "rank": 2, "playerName": "FarmQueen", "avatar": "👩🌾", "level": 48, "score": 118000, "companyName": "Green Acres" },
    ...
  ]
}
```

---

## 14. Achievements

### GET /A1/achievements

```
Response 200:
[
  {
    "id": "first_farm",
    "name": "First Farm",
    "icon": "🌾",
    "description": "Build your first farm",
    "reward": { "cash": 500, "xp": 100 },
    "progress": 1,
    "maxProgress": 1,
    "status": "locked" | "completed" | "claimed"
  },
  ...
]
```

### POST /A1/achievements/:id/claim

```
Response 200:
{
  "achievement": { ...achievement, status: "claimed" },
  "reward": { "cash": 500, "xp": 100 }
}
```

---

## 15. Collection

### GET /A1/collection

Claimable items across all buildings.

```
Response 200:
{
  "claimableCount": 3,
  "items": [
    {
      "id": "uuid",
      "buildingName": "Farm (Plot #1)",
      "buildingIcon": "🌾",
      "resource": "wheat",
      "quantity": 120,
      "quality": 92,
      "completedAt": "2m ago",
      "status": "ready" | "claimed"
    },
    ...
  ]
}
```

### POST /A1/collection/claim-all

Claim all ready items.

```
Response 200:
{
  "claimed": [ { "resource": "wheat", "quantity": 120 }, ... ],
  "totalItems": 340
}
```

---

## 16. Wiki (Static Data)

### GET /A1/wiki/resources

Resource encyclopedia.

```
Response 200:
[
  {
    "id": "wheat",
    "icon": "🌾",
    "name": "Wheat",
    "category": "Raw" | "Processed" | "Finished",
    "tier": 1,
    "basePrice": 12.5,
    "producedBy": ["Farm"],
    "usedIn": ["Flour", "Bread"],
    "tradable": true,
    "tip": "Foundation of your food chain."
  },
  ...
]
```

### GET /A1/wiki/buildings

Building encyclopedia.

```
Response 200:
[
  {
    "id": "farm",
    "icon": "🌾",
    "name": "Farm",
    "cost": 500,
    "difficulty": "Beginner",
    "region": "Inland Estate",
    "produces": ["Grain", "Sugar", "Vegetables"],
    "mechanic": "Crop selection + seasonal bonuses",
    "upgradeEffect": "Unlock rarer crops and larger batches",
    "description": "The backbone of your food empire."
  },
  ...
]
```

### GET /A1/wiki/guides

FAQ / getting started guides.

```
Response 200:
{
  "sections": [
    {
      "title": "Getting Started",
      "articles": [
        { "question": "How do I build my first farm?", "answer": "..." },
        ...
      ]
    },
    ...
  ]
}
```

---

## WebSocket Events Summary

| Direction | Event | Description |
|---|---|---|
| → | `join_channel` | Join a chat channel |
| → | `leave_channel` | Leave a chat channel |
| → | `send_message` | Send a chat message |
| ← | `message` | New chat message |
| ← | `price_update` | Real-time market price change |
| ← | `order_filled` | User's market order matched |
| ← | `notification` | System notification |
| ← | `production_ready` | Building finished production |

---

## Error Codes

| Code | HTTP | Meaning |
|---|---|---|
| `INVALID_CREDENTIALS` | 401 | Wrong email/password |
| `TOKEN_EXPIRED` | 401 | JWT expired, refresh needed |
| `FORBIDDEN` | 403 | No permission |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Duplicate / state conflict |
| `VALIDATION_ERROR` | 422 | Invalid request body |
| `RATE_LIMITED` | 429 | Too many requests |
| `INSUFFICIENT_FUNDS` | 422 | Not enough cash |
| `INSUFFICIENT_INVENTORY` | 422 | Not enough items |
| `BUILDING_BUSY` | 409 | Building already producing |
| `PREREQUISITES_NOT_MET` | 422 | Research/building locked |

# Appendix: backend-next Implementation Status

> 对比基准：client-next 页面需求 vs backend-next 现有 handler。
> - ✅ = 已实现
> - ⚠️ = 已有但路径/协议/字段不一致
> - ❌ = 未实现

| 模块 | API端点 | backend-next | 说明 |
|---|---|---|---|
| **Auth** | | | |
| | `POST /A1/auth/register` | ⚠️ `POST /api/register` | 后端用 username，前端发 email；缺 OTP |
| | `POST /A1/auth/login` | ⚠️ `POST /api/login` | 同上 |
| | `POST /A1/auth/logout` | ❌ | 需要新增 |
| | `POST /A1/auth/refresh` | ❌ | 需要新增 |
| | `GET /A1/auth/me` | ❌ | 需要新增 |
| | `POST /A1/auth/forgot-password` | ❌ | 需要新增 |
| | `POST /A1/auth/reset-password` | ❌ | 需要新增 |
| | `POST /A1/auth/verify-otp` | ❌ | 需要新增 |
| | `POST /A1/auth/resend-otp` | ❌ | 需要新增 |
| **Player** | | | |
| | `GET /A1/player` | ❌ | TopBar/Settings 核心数据 |
| | `PATCH /A1/player` | ❌ | 资料修改 |
| | `GET /A1/player/profile` | ❌ | Profile 页聚合 |
| **Buildings** | | | |
| | `GET /A1/buildings` | ⚠️ 目录数据已加载(catalog) | 只缺 Handler 暴露 |
| | `GET /A1/player/buildings` | ⚠️ `GET /api/v3/companies/me/buildings/` | 路径差异，shape 需对齐 |
| | `POST /A1/player/buildings` | ❌ | 新建建筑 |
| | `DELETE /A1/player/buildings/:id` | ❌ | 拆除 |
| | `PATCH /A1/player/buildings/:id/move` | ❌ | 移动 |
| | `GET /A1/player/buildings/:id` | ❌ | 建筑详情 |
| **Regions** | | | |
| | `GET /A1/regions` | ❌ | 地图 + 地块 |
| **Market** | | | |
| | `GET /A1/market/prices` | ⚠️ `GET /api/v3/resources/` | 路径差异 |
| | `GET /A1/market/prices/:resource` | ❌ | 历史 K 线 |
| | `GET /A1/market/orderbook/:resource` | ⚠️ `GET /api/v3/market-depth/{rid}/{quality}/` | 路径差异 |
| | `GET /A1/market/orders` | ❌ | 我的订单列表 |
| | `POST /A1/market/orders` | ⚠️ `POST /api/v2/market-order/` | 路径差异 |
| | `DELETE /A1/market/orders/:id` | ⚠️ `DELETE /api/v2/market-order/cancel/{orderId}/` | 路径差异 |
| | `POST /A1/market/take` | ⚠️ `POST /api/v2/market-order/take/` | 路径差异 |
| **Warehouse** | | | |
| | `GET /A1/warehouse` | ⚠️ `GET /api/v2/companies/me/warehouse/` | 路径差异 |
| | `POST /A1/warehouse/transfer` | ❌ | 调拨物品 |
| **Finance** | | | |
| | `GET /A1/finance/summary` | ⚠️ 拆在 income/balance/cashflow 三个端点 | 需聚合 |
| | `GET /A1/finance/cashflow` | ⚠️ `GET /api/v2/companies/me/cashflow/recent/` | 路径差异 |
| **Production** | | | |
| | `GET /A1/production/jobs` | ⚠️ `GET /api/v2/production/jobs/` | 路径差异 |
| | `GET /A1/production/claimable` | ⚠️ `GET /api/v2/production/claimable/` | 路径差异 |
| | `POST /A1/player/buildings/:id/produce` | ⚠️ `POST /api/v2/production/start/` | 路径差异 |
| | `POST /A1/production/jobs/:id/claim` | ⚠️ `POST /api/v2/production/claim/{jobId}/` | 路径差异 |
| **Chat** | | | |
| | `WebSocket /A1/ws` | ❌ | tech stack 有 coder/websocket 但未实现 |
| | `GET /A1/chat/channels` | ❌ |  |
| | `GET /A1/chat/channels/:id/messages` | ❌ |  |
| | `GET /A1/chat/dms` | ❌ |  |
| | `GET /A1/chat/dms/:id/messages` | ❌ |  |
| | `POST /A1/chat/messages` | ❌ |  |
| **Contracts** | | | |
| | `GET /A1/contracts` | ❌ |  |
| | `POST /A1/contracts/:id/accept` | ❌ |  |
| | `POST /A1/contracts/:id/fulfill` | ❌ |  |
| **Research** | | | |
| | `GET /A1/research/nodes` | ❌ | domain 有 Project 类型，缺 endpoints |
| | `POST /A1/research/:id/start` | ❌ |  |
| | `POST /A1/research/:id/claim` | ❌ |  |
| **Executives** | | | |
| | `GET /A1/executives` | ❌ |  |
| | `POST /A1/executives/:id/hire` | ❌ |  |
| | `POST /A1/executives/:id/assign` | ❌ |  |
| | `POST /A1/executives/:id/fire` | ❌ |  |
| **Leaderboard** | | | |
| | `GET /A1/leaderboard` | ❌ |  |
| **Achievements** | | | |
| | `GET /A1/achievements` | ❌ |  |
| | `POST /A1/achievements/:id/claim` | ❌ |  |
| **Collection** | | | |
| | `GET /A1/collection` | ⚠️ 可复用 `/production/claimable/` |  |
| | `POST /A1/collection/claim-all` | ❌ |  |
| **Wiki** | | | |
| | `GET /A1/wiki/resources` | ❌ | 已有 catalog 数据，加 handler 暴露 |
| | `GET /A1/wiki/buildings` | ❌ | 同上 |
| | `GET /A1/wiki/guides` | ❌ | 纯静态 FAQ |
| **Bonds (后端有，客户端无)** | | | |
| | `GET /A1/bonds` | ✅  | 后端已实现，前端无对应页面 |
| | `POST /A1/bonds` | ✅  |  |
| | `GET /A1/bonds/{id}` | ✅  |  |
| | `GET /A1/bonds/owned` | ✅  |  |
| | `GET /A1/bonds/sold` | ✅  |  |

## 迁移优先级

| 轮次 | 内容 | 预估 |
|---|---|---|
| 1 | router prefix 从 `/api`/`/api/v2`/`/api/v3` 统一改为 `/A1`，shape 对齐 | 0.5d |
| 2 | Auth 补 OTP/email/me/refresh/forgot/reset，后端 LoginRequest 改 email | 1d |
| 3 | 新增 Player / Regions / Chat (WebSocket) | 2d |
| 4 | 新增 Contracts / Research (由 domain 已有类型补齐) / Executives | 1.5d |
| 5 | 新增 Leaderboard / Achievements / Wiki / Collection claim-all | 1d |
| 合计 | | ~6d |
