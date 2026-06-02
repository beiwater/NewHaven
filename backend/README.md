# Go Sim API

Sim Companies 游戏经济的 Go API 复刻版。纯 Go 实现，零外部运行时依赖（PostgreSQL可选）。

## 快速开始

```bash
go run ./cmd/simapi/
# => sim api on http://127.0.0.1:8088
```

## 架构

```
cmd/simapi/main.go  — 30行编排，优雅关闭
├── internal/config/     — 27个环境变量
├── internal/middleware/ — Recovery, Logger, CORS, RequestID
├── internal/handler/    — HTTP层，10个域文件，typed DTO
├── internal/service/    — 业务逻辑，6个域文件，sync.Mutex防并发
├── internal/storage/    — PostgreSQL持久化(可选)
├── internal/model/      — 数据模型
├── internal/data/       — JSON数据加载
└── internal/formula/    — 纯函数公式
```

## 技术栈

|层|选型|
|---|---|
|语言|Go 1.22+|
|HTTP|标准库 `net/http.ServeMux`|
|存储|内存(默认) / PostgreSQL(可选 via `SIM_API_DATABASE_URL`)|
|数据库驱动|`github.com/jackc/pgx/v5`|
|Lint|`golangci-lint` (gocyclo, gocognit, errcheck, staticcheck)|

## 功能覆盖

|系统|端点数|状态|
|---|---|---|
|公司|15|✅|
|市场交易|7|✅ 含机器人流动性 + 价格撮合|
|生产制造|3|✅ 配方 + 计时 + 收货|
|债券金融|7|✅ 发行/购买/赎回/计息/评级|
|政府合同|5|✅ 投标/授标/交付/违约|
|高管|6|✅ 搜索/招募/培训/挖角|
|报社|4|✅ 文章CRUD|
|研发|4|✅ 研究项目 + 进度|
|航空|5|✅ 火箭项目 + 发射|
|等级经验|3|✅ XP + 升级 + 奖励|
|SimBoost|3|✅ 加速道具|
|建筑竞标|4|✅ 拍卖|
|财务|4|✅ 三大报表 + 流水|

## 配置

所有配置通过 `SIM_API_*` 环境变量：

```bash
# 启动并连接 PostgreSQL
export SIM_API_DATABASE_URL="postgres://user:pass@localhost:5432/simapi?sslmode=disable"
export SIM_API_START_MONEY=500000
export SIM_API_FEE_PCT=0.02
go run ./cmd/simapi/
```

完整配置列表见 `internal/config/config.go`。

## 项目结构

```
go-sim-api/
├── cmd/simapi/main.go         # 入口
├── internal/
│   ├── config/                # 配置
│   ├── middleware/             # HTTP中间件
│   ├── handler/               # HTTP处理器(typed DTO)
│   ├── service/               # 业务逻辑
│   ├── storage/               # 持久化
│   ├── model/                 # 数据类型
│   ├── data/                  # JSON数据加载
│   └── formula/               # 经济公式
├── docs/                      # 文档
├── decompiled/                # 反编译参考数据(只读)
└── go.mod
```

## 开发

```bash
# 检查代码质量
go vet ./...
golangci-lint run ./...

# 构建
go build ./...

# 运行(内存模式)
go run ./cmd/simapi/
```

## 分文件规约

- 单文件 ≤ 300 行
- handler 按域分文件（company.go, market.go, bond.go ...）
- service 按域分文件
- 公共类型在 model/types.go
- 纯函数在 formula/

## 数据来源

反编译资源数据位于 `decompiled/data/`，仅供复刻参考。

[MIT](../LICENSE)

