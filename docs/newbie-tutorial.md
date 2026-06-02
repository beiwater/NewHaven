# 新手教程：Go Sim API

## 1. 这是什么

这是 Sim Companies 游戏的 Go 语言复刻版 API 后端。它实现了经济模拟的核心系统：市场交易、生产制造、债券金融、政府合同等。

## 2. 快速启动

```bash
# 确保已安装 Go 1.22+
go version

# 启动服务（默认监听 127.0.0.1:8088，全内存运行）
go run ./cmd/simapi/
```

看到 `sim api on http://127.0.0.1:8088` 就起来了。

## 3. 验证服务是否正常

```bash
# 健康检查
curl http://127.0.0.1:8088/healthz

# 就绪检查
curl http://127.0.0.1:8088/readyz

# CSRF 令牌（所有写操作需要带上）
curl http://127.0.0.1:8088/api/csrf/
```

## 4. 典型请求流程

### 4.1 查公司信息

```bash
curl http://127.0.0.1:8088/api/v3/companies/1234567/
```

返回公司概况、资金、等级、库存、建筑列表。

### 4.2 看有什么资源

```bash
curl http://127.0.0.1:8088/api/v3/resources/1/
```

资源 ID 1=电力, 2=水, 8=牛肉, 9=汉堡。返回资源名称、类别、基础价格。

### 4.3 看市场行情

```bash
# 48 小时 K 线数据
curl http://127.0.0.1:8088/api/v3/market-ticker/8/

# 当前订单簿（含机器人挂单）
curl http://127.0.0.1:8088/api/v3/market/8/0/
```

第三个路径段是品质参数（0=普通品质）。

### 4.4 挂单买入

```bash
curl -X POST http://127.0.0.1:8088/api/v2/market-order/ \
  -H 'Content-Type: application/json' \
  -d '{"resourceId":8,"kind":1,"quality":0,"quantity":10,"price":9.5}'
```

- `kind`: 0=卖出, 1=买入
- `price`: 单价。系统会校验 tick size（最小价格变动单位），不满足会拒绝。
- 成交后收取 `SIM_API_FEE_PCT`（默认 4%）手续费。

### 4.5 吃单（市价买入）

```bash
curl -X POST http://127.0.0.1:8088/api/v2/market-order/take/ \
  -H 'Content-Type: application/json' \
  -d '{"resource":8,"quantity":5,"quality":0,"maxPrice":15}'
```

系统从最低价的卖单开始匹配，逐层吃掉直到 quantity 满足为止，`maxPrice` 限制最高接受单价。

### 4.6 取消挂单

```bash
curl -X DELETE http://127.0.0.1:8088/api/v2/market-order/cancel/{订单ID}/
```

### 4.7 启动生产

```bash
curl -X POST http://127.0.0.1:8088/api/v1/buildings/b-1/busy/ \
  -H 'Content-Type: application/json' \
  -d '{"kind":8,"amount":10,"estimatedSecondsToFinish":60}'
```

- `b-1`: 建筑 ID，可以通过 `GET /api/v2/companies/me/buildings/` 查到
- `kind`: 要生产的资源 ID
- `amount`: 数量
- `estimatedSecondsToFinish`: 生产耗时（秒），系统按此计算完成时间

### 4.8 查看生产任务

```bash
curl http://127.0.0.1:8088/api/v2/production/jobs/
```

返回所有生产任务及状态（`running` / `ready` / `claimed`）。任务完成后调用：

```bash
curl -X POST http://127.0.0.1:8088/api/v2/production/claim/{jobId}/
```

### 4.9 查账

```bash
curl http://127.0.0.1:8088/api/v2/companies/me/income-statement/
curl http://127.0.0.1:8088/api/v2/companies/me/balance-sheet/
curl http://127.0.0.1:8088/api/v2/companies/me/cashflow-statement/
curl http://127.0.0.1:8088/api/v2/companies/me/cashflow/recent/
```

财务报表由事件账本实时生成，反映公司成立至今的累计数据。

## 5. 配置

所有参数通过环境变量设置。最常用的：

| 环境变量 | 默认值 | 作用 |
|---|---|---|
| `SIM_API_ADDR` | `127.0.0.1:8088` | 监听地址 |
| `SIM_API_START_MONEY` | `200000` | 初始资金 |
| `SIM_API_START_LEVEL` | `42` | 初始等级 |
| `SIM_API_FEE_PCT` | `0.04` | 交易手续费率 |
| `SIM_API_DATABASE_URL` | (空) | PostgreSQL 连接字符串，不设则全内存运行 |
| `SIM_API_CSRF_TOKEN` | `dev-csrf-token` | CSRF 令牌 |
| `SIM_API_COMPANY_ID` | `1234567` | 默认公司 ID |
| `SIM_API_COMPANY_NAME` | `Example Company Inc` | 默认公司名称 |
| `SIM_API_BOT1_ID` | `900001` | 机器人 1 ID |
| `SIM_API_BOT2_ID` | `900002` | 机器人 2 ID |
| `SIM_API_BOT_MONEY` | `5000000` | 机器人初始资金 |
| `SIM_API_BOT_LEVEL` | `99` | 机器人等级 |
| `SIM_API_ADMIN_BASE` | `1.35` | 行政管理费基数 |
| `SIM_API_BOND_FACE` | `5000` | 债券面值 |
| `SIM_API_MAX_LEDGER` | `5000` | 账本最大条目数 |
| `SIM_API_PROD_MOD` | `1.02` | 生产系数 |
| `SIM_API_BOT_CYCLE_AMP` | `0.06` | 机器人报价周期波动幅度 |
| `SIM_API_BOT_BASE` | `8.0` | 机器人报价基价 |

完整列表见 `internal/config/config.go`。

## 6. 数据模型

所有核心类型定义在 `internal/model/types.go`：

| 类型 | 字段 | 说明 |
|---|---|---|
| `Company` | `ID, Name, Money, Level, Inventory, PlacedBuildings` | 公司实体，含资金、库存、已摆放建筑 |
| `MarketOrder` | `ID, ResourceID, Kind, Price, Quality, Quantity, Remaining, CompanyID, CreatedAt` | 市场挂单，`Kind=0` 卖出 / `Kind=1` 买入 |
| `Trade` | `ID, ResourceID, Quality, Quantity, Price, BuyOrderID, SellOrderID, CreatedAt` | 成交记录，关联买卖双方订单 |
| `ProductionJob` | `ID, BuildingID, ResourceID, Amount, Input, Output, StartedAt, CompletesAt, Status, Meta` | 生产任务，状态机：`running → ready → claimed` |
| `GameState` | `Company, Companies, Orders, Trades, CSRFToken, Bonds, Ledger, ...` | 全局服务状态容器，内存中持有全部数据 |

## 7. 架构分层

```
cmd/simapi/main.go          ← 90 行：加载配置、数据、创建服务、启动 HTTP
    ↓
internal/handler/            ← HTTP 层：路由注册 + JSON 请求/响应编解码
    ↓
internal/service/            ← 业务逻辑：市场撮合、生产调度、财务报表、债券合约
    ↓
internal/storage/            ← 持久化层：PostgreSQL (pgx/v5)，可选，不设则全内存
    ↓
internal/model/              ← 领域类型定义
internal/data/               ← 静态数据加载（物品、配方、等级表）
internal/config/             ← 环境变量配置
internal/formula/            ← 可执行经济公式（产量、价格、债券利率计算）
internal/middleware/         ← HTTP 中间件：CORS、日志、Recovery、RequestID
```

**数据流示意：**

```
curl → middleware（日志/CSRF 校验）→ handler（解析路径/JSON）→ service（业务处理）→ storage（若配 DB）→ 返回 JSON
```

主公司 + 两台机器人启动时自动初始化并预挂买卖单。所有数据生命周期同服务进程（除非配置了 PostgreSQL）。

## 8. 下一步

- 读 `game-wiki.md` 了解每个子系统的详细设计
- 读 `README.md` 了解完整 API 端点清单
- 读 `ECONOMY_SYSTEM.md` 了解经济参数设计
- 需要扩展 API：在 `internal/handler/` 下新建文件，在 `Register()` 里挂上路由即可
