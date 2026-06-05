# Database Design v1 — Phase 1

> **Phase 1 设计文档，仅用于规划。不生成迁移文件。**
> 目标：为 NewHaven 后端从内存态切换到 PostgreSQL 持久化提供完整的数据模型参考。
> 当前基线：pgx/v5，JSONB 全量快照，无正式迁移工具。
> 推荐迁移工具：`github.com/pressly/goose/v3`（Phase 2 引入）。

---

## 设计原则

1. **存储即适配器，不主导模型** — `model` 层无 database tag；storage adapter 负责行与 struct 的转换。
2. **版本化快照优先** — 全量 `game_state` snapshot 在迁移过渡期作为兜底，正式迁移后逐步淘汰。
3. **分布式友好主键** — 所有业务实体使用 BIGINT 或 UUID（`gen_random_uuid()`），不依赖自增作为业务标识。
4. **追加写 = 不可变日志** — market_trades、ledger_entries、chat_messages 仅 INSERT，无 UPDATE/DELETE。
5. **静态数据独立 schema** — `*_catalog` 表为只读种子数据，代码仓库管理，不通过应用层写入。
6. **事务边界在服务层** — 游戏逻辑的 ACID 由 service 层编排，一个 service method = 一个 db transaction。
7. **时间由服务层注入** — `created_at` `updated_at` 等时间字段由应用层 `time.Now()` 提供，允许测试中 mock。

---

## 域划分

| 域 | 描述 | 涉及表 |
|------|-------------|----------|
| auth | 认证与鉴权 | players, auth_sessions |
| company | 公司与建筑 | companies, company_buildings |
| catalog | 静态配置 | building_catalog, resource_catalog, recipe_catalog |
| warehouse | 库存 | warehouse_items |
| production | 生产 | production_jobs |
| market | 市场 | market_orders, market_trades, market_tickers |
| finance | 财务 | ledger_entries, financial_snapshots, bonds, bond_holdings |
| research | 科研 | research_nodes, company_research |
| executives | 高管 | executive_catalog, company_executives |
| government | 政府订单 | government_orders, government_bids |
| social | 社交 | chat_messages, notifications |

---

## 1. 表定义

### 1.1 players — 玩家账户

| 项目 | 值 |
|------|--------|
| **用途** | 存储玩家登录凭证与基本资料，每个自然人对应用户表一行。 |
| **域归属** | auth |
| **主键** | `id BIGSERIAL` |
| **重要字段** | `username VARCHAR(64) NOT NULL UNIQUE` — 登录用户名，大小写敏感<br>`password_hash VARCHAR(255) NOT NULL` — bcrypt hash<br>`display_name VARCHAR(64) NOT NULL` — 游戏内显示名<br>`gender VARCHAR(8)` — 可选，`male`/`female`/`other`<br>`email VARCHAR(255)` — 可选，预留密码找回<br>`created_at TIMESTAMPTZ NOT NULL` |
| **外键** | 无 |
| **重要索引** | `UNIQUE(username)`，`UNIQUE(email)` 若 email 非空 |
| **事务说明** | 注册时 INSERT players + 创建默认 company 应在一个事务内完成 |
| **历史/审计** | 无硬删除需求；可加 `deleted_at TIMESTAMPTZ` 软删除标记 |
| **迁移风险** | 低。与现有 `model.Player` 结构基本匹配。密码 hash 字段从现有内存状态迁移需注意 bcrypt cost 一致。 |

```sql
CREATE TABLE players (
    id           BIGSERIAL PRIMARY KEY,
    username     VARCHAR(64) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(64) NOT NULL,
    gender       VARCHAR(8),
    email        VARCHAR(255),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_players_username ON players(username) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_players_email ON players(email) WHERE email IS NOT NULL AND deleted_at IS NULL;
```

---

### 1.2 auth_sessions — 登录会话

| 项目 | 值 |
|------|--------|
| **用途** | 跟踪玩家活跃会话，支持 token 撤销和刷新。 |
| **域归属** | auth |
| **主键** | `id BIGSERIAL` |
| **重要字段** | `player_id BIGINT NOT NULL` — 关联玩家<br>`token_hash VARCHAR(255) NOT NULL` — token 的 SHA-256 hash，不存原文<br>`expires_at TIMESTAMPTZ NOT NULL` — 过期时间，查询时过滤<br>`created_at TIMESTAMPTZ NOT NULL DEFAULT now()` |
| **外键** | `player_id REFERENCES players(id) ON DELETE CASCADE` |
| **重要索引** | `INDEX(player_id)` — 查某玩家的活跃会话<br>`INDEX(token_hash)` — 查 token 对应会话 |
| **事务说明** | 登录时 INSERT session，注销时 DELETE，无需跨表事务 |
| **历史/审计** | 无，过期 session 可定期清理或忽略 |
| **迁移风险** | 低。当前 JWT 无状态；引入 sessions 表需要同步修改 auth middleware 做数据库验证。 |

```sql
CREATE TABLE auth_sessions (
    id         BIGSERIAL PRIMARY KEY,
    player_id  BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_auth_sessions_player ON auth_sessions(player_id);
CREATE INDEX idx_auth_sessions_token ON auth_sessions(token_hash);
```

---

### 1.3 companies — 公司

| 项目 | 值 |
|------|--------|
| **用途** | 游戏中每个玩家的核心实体，持有资金、等级、经验值及偏好设置。 |
| **域归属** | company |
| **主键** | `id BIGSERIAL` |
| **重要字段** | `player_id BIGINT NOT NULL` — 所有者<br>`company_name VARCHAR(128) NOT NULL`<br>`money DECIMAL(20,2) NOT NULL DEFAULT 0` — 当前资金，两位小数<br>`level INT NOT NULL DEFAULT 1`<br>`xp BIGINT NOT NULL DEFAULT 0`<br>`preferences JSONB` — 用户偏好，如自动续购开关<br>`created_at TIMESTAMPTZ NOT NULL` |
| **外键** | `player_id REFERENCES players(id) ON DELETE CASCADE` |
| **重要索引** | `UNIQUE(player_id)` — 一个玩家一个公司（当前设计），`INDEX(company_name)` |
| **事务说明** | `money` 为游戏核心货币字段，所有资金变更须在事务内完成，配合 audit 行级校验（`balance_after >= 0`） |
| **历史/审计** | 资金变更由 ledger_entries 表记录；公司元信息变更推荐 `updated_at` 时间戳 |
| **迁移风险** | 中。现有 `model.Company.Money` 为 `float64`，迁移到 `DECIMAL(20,2)` 需注意浮点数精度截断。Inventory 等字段被拆入 warehouse_items。 |

```sql
CREATE TABLE companies (
    id           BIGSERIAL PRIMARY KEY,
    player_id    BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    company_name VARCHAR(128) NOT NULL,
    money        DECIMAL(20,2) NOT NULL DEFAULT 0,
    level        INT NOT NULL DEFAULT 1,
    xp           BIGINT NOT NULL DEFAULT 0,
    preferences  JSONB DEFAULT '{}',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_companies_player ON companies(player_id);
CREATE INDEX idx_companies_name ON companies(company_name);
```

---

### 1.4 building_catalog — 建筑配置（静态）

| 项目 | 值 |
|------|--------|
| **用途** | 定义游戏中所有可建造的建筑类型模板，种子数据由代码仓库维护。 |
| **域归属** | catalog |
| **主键** | `id INT PRIMARY KEY` |
| **重要字段** | `kind INT NOT NULL` — 建筑种类分类<br>`name VARCHAR(128) NOT NULL`<br>`type VARCHAR(32) NOT NULL` — 如 `production` `storage` `headquarters`<br>`base_cost DECIMAL(12,2) NOT NULL` — 基础建造费用<br>`base_output DECIMAL(12,2)` — 基础产出/效率<br>`config JSONB NOT NULL DEFAULT '{}'` — 扩展属性（占用格数、解锁条件等） |
| **外键** | 无 |
| **重要索引** | `UNIQUE(kind)` 若 kind 为业务标识，否则 `INDEX(kind)` |
| **事务说明** | 只读，无事务需求 |
| **历史/审计** | 无，静态数据在版本控制中，变更走迁移文件 |
| **迁移风险** | 低。纯 INSERT 种子数据。需与前端 `building_catalog` 对齐 ID。 |

```sql
CREATE TABLE building_catalog (
    id          INT PRIMARY KEY,
    kind        INT NOT NULL,
    name        VARCHAR(128) NOT NULL,
    type        VARCHAR(32) NOT NULL,
    base_cost   DECIMAL(12,2) NOT NULL,
    base_output DECIMAL(12,2),
    config      JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_building_catalog_kind ON building_catalog(kind);
```

---

### 1.5 company_buildings — 公司已建造建筑

| 项目 | 值 |
|------|--------|
| **用途** | 记录每个公司在地图上的建筑实例及状态。 |
| **域归属** | company |
| **主键** | `id BIGSERIAL` |
| **重要字段** | `company_id BIGINT NOT NULL`<br>`building_id INT NOT NULL` — 指向 building_catalog<br>`map_id INT NOT NULL` — 地图地块 ID<br>`slot_id INT` — 地块内的格子编号<br>`x INT NOT NULL, y INT NOT NULL` — 坐标冗余，方便空间查询<br>`level INT NOT NULL DEFAULT 1` — 建筑等级<br>`robot_count INT NOT NULL DEFAULT 0` — 分配的机器人数量<br>`created_at TIMESTAMPTZ NOT NULL` |
| **外键** | `company_id REFERENCES companies(id) ON DELETE CASCADE`<br>`building_id REFERENCES building_catalog(id)` |
| **重要索引** | `UNIQUE(company_id, map_id, slot_id)` — 同一地块同一格子不能重叠<br>`INDEX(company_id)` — 查某公司的所有建筑<br>`INDEX(company_id, building_id)` — 查某类建筑 |
| **事务说明** | 建造、升级、拆除在一个事务内完成，同时更新 company.money |
| **历史/审计** | 建造/升级/拆除记录可以写入 ledgers 或在独立 audit 表 |
| **迁移风险** | 中。现有 `model.Company.PlacedBuildings` 为 `[]map[string]any`，需拆为行级。 |

```sql
CREATE TABLE company_buildings (
    id          BIGSERIAL PRIMARY KEY,
    company_id  BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    building_id INT NOT NULL REFERENCES building_catalog(id),
    map_id      INT NOT NULL,
    slot_id     INT,
    x           INT NOT NULL,
    y           INT NOT NULL,
    level       INT NOT NULL DEFAULT 1,
    robot_count INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_cb_pos ON company_buildings(company_id, map_id, slot_id);
CREATE INDEX idx_cb_company ON company_buildings(company_id);
CREATE INDEX idx_cb_company_building ON company_buildings(company_id, building_id);
```

---

### 1.6 resource_catalog — 资源配置（静态）

| 项目 | 值 |
|------|--------|
| **用途** | 定义游戏中所有资源类型模板（食材、半成品、成品等），种子数据。 |
| **域归属** | catalog |
| **主键** | `id INT PRIMARY KEY` |
| **重要字段** | `db_letter VARCHAR(2) NOT NULL` — 数据库单字母缩写（兼容旧系统）<br>`name VARCHAR(128) NOT NULL`<br>`category VARCHAR(32) NOT NULL` — 分类如 `raw` `processed` `finished`<br>`tier INT NOT NULL` — 资源阶位，影响解锁和价格<br>`base_price DECIMAL(12,2) NOT NULL` — 系统基准价<br>`config JSONB NOT NULL DEFAULT '{}'` — 扩展属性（是否可交易、体积等） |
| **外键** | 无 |
| **重要索引** | `UNIQUE(db_letter)` |
| **事务说明** | 只读 |
| **历史/审计** | 无 |
| **迁移风险** | 低。现有 `decompiled/data/resources.json` 已有完整数据。 |

```sql
CREATE TABLE resource_catalog (
    id         INT PRIMARY KEY,
    db_letter  VARCHAR(2) NOT NULL,
    name       VARCHAR(128) NOT NULL,
    category   VARCHAR(32) NOT NULL,
    tier       INT NOT NULL,
    base_price DECIMAL(12,2) NOT NULL,
    config     JSONB NOT NULL DEFAULT '{}'
);

CREATE UNIQUE INDEX idx_rc_letter ON resource_catalog(db_letter);
```

---

### 1.7 warehouse_items — 库存

| 项目 | 值 |
|------|--------|
| **用途** | 每个公司对每种资源的持有数量，多行 = 多资源。 |
| **域归属** | warehouse |
| **主键** | `(company_id, resource_id)` 复合主键 |
| **重要字段** | `company_id BIGINT NOT NULL`<br>`resource_id INT NOT NULL`<br>`quantity INT NOT NULL DEFAULT 0` — 当前持有量，可为 0（保留行） |
| **外键** | `company_id REFERENCES companies(id) ON DELETE CASCADE`<br>`resource_id REFERENCES resource_catalog(id)` |
| **重要索引** | 复合主键自带；`INDEX(company_id)` |
| **事务说明** | 所有库存变更在事务内完成；`quantity` 不应低于 0，建议应用层或 `CHECK(quantity >= 0)` |
| **历史/审计** | 库存变更由 ledger_entries 的 metadata 记录资源变动明细 |
| **迁移风险** | 中。现有 `model.Company.Inventory` 为 `map[int]int`，需逐条插入。quality 相关字段（`QualityInventory`）未来扩展。 |

```sql
CREATE TABLE warehouse_items (
    company_id  BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    resource_id INT NOT NULL REFERENCES resource_catalog(id),
    quantity    INT NOT NULL DEFAULT 0,
    PRIMARY KEY (company_id, resource_id)
);
```

---

### 1.8 recipe_catalog — 配方配置（静态）

| 项目 | 值 |
|------|--------|
| **用途** | 定义从输入资源到输出资源的转换配方，每个配方对应一种建筑类型。 |
| **域归属** | catalog |
| **主键** | `id INT PRIMARY KEY` |
| **重要字段** | `output_resource_id INT NOT NULL` — 产出资源<br>`inputs JSONB NOT NULL` — `{"resource_id": quantity, ...}`，输入配方<br>`duration_base INT NOT NULL` — 基础生产时长（秒）<br>`building_kind INT NOT NULL` — 允许生产的建筑 kind |
| **外键** | `output_resource_id REFERENCES resource_catalog(id)`<br>`building_kind REFERENCES building_catalog(kind)` |
| **重要索引** | `INDEX(building_kind)` — 按建筑查可用配方<br>`INDEX(output_resource_id)` |
| **事务说明** | 只读 |
| **历史/审计** | 无 |
| **迁移风险** | 低。现有 `backend/decompiled/data/economy_model.json` 或公式文档中有配方定义。 |

```sql
CREATE TABLE recipe_catalog (
    id                 INT PRIMARY KEY,
    output_resource_id INT NOT NULL REFERENCES resource_catalog(id),
    inputs             JSONB NOT NULL,
    duration_base      INT NOT NULL,
    building_kind      INT NOT NULL REFERENCES building_catalog(kind)
);

CREATE INDEX idx_recipe_building ON recipe_catalog(building_kind);
CREATE INDEX idx_recipe_output ON recipe_catalog(output_resource_id);
```

---

### 1.9 production_jobs — 生产任务

| 项目 | 值 |
|------|--------|
| **用途** | 每个建筑的生产工单，记录开始时间、产出物及完成/领取状态。 |
| **域归属** | production |
| **主键** | `id BIGSERIAL` |
| **重要字段** | `company_id BIGINT NOT NULL`<br>`building_id BIGINT NOT NULL` — 指向 company_buildings<br>`resource_id INT NOT NULL` — 产出的资源类型<br>`quantity INT NOT NULL` — 产出数量<br>`started_at TIMESTAMPTZ NOT NULL`<br>`duration_seconds INT NOT NULL` — 开始时计算的持续时间（考虑 buff 后）<br>`completed BOOLEAN NOT NULL DEFAULT FALSE` — 是否完成（`started_at + duration_seconds <= now()`）<br>`claimed BOOLEAN NOT NULL DEFAULT FALSE` — 是否领取产出 |
| **外键** | `company_id REFERENCES companies(id) ON DELETE CASCADE`<br>`building_id REFERENCES company_buildings(id) ON DELETE CASCADE`<br>`resource_id REFERENCES resource_catalog(id)` |
| **重要索引** | `INDEX(company_id, completed, claimed)` — 查某公司未领取的任务<br>`INDEX(building_id)` — 查某建筑上的任务<br>`INDEX(completed) WHERE completed = FALSE` — 定时器扫描未完成任务 |
| **事务说明** | 开始生产：检查输入库存 → 扣除输入 → INSERT job → 状态变更。领取生产：检查 completed → UPDATE claimed → INSERT 产出到库存 → 加 XP。两者均在事务内完成。 |
| **历史/审计** | `claimed` 翻转不可逆，已领取不可撤回。所有资源变动反映在 ledger。 |
| **迁移风险** | 中。现有 `model.ProductionJob` 字段较多，需要映射。`CompletesAt` 由 `started_at + duration_seconds` 计算。 |

```sql
CREATE TABLE production_jobs (
    id               BIGSERIAL PRIMARY KEY,
    company_id       BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    building_id      BIGINT NOT NULL REFERENCES company_buildings(id) ON DELETE CASCADE,
    resource_id      INT NOT NULL REFERENCES resource_catalog(id),
    quantity         INT NOT NULL,
    started_at       TIMESTAMPTZ NOT NULL,
    duration_seconds INT NOT NULL,
    completed        BOOLEAN NOT NULL DEFAULT FALSE,
    claimed          BOOLEAN NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_pj_company_claim ON production_jobs(company_id, completed, claimed);
CREATE INDEX idx_pj_building ON production_jobs(building_id);
CREATE INDEX idx_pj_pending ON production_jobs(completed) WHERE completed = FALSE;
```

---

### 1.10 market_orders — 市场挂单

| 项目 | 值 |
|------|--------|
| **用途** | 玩家在市场发布的买卖委托，包含价格、数量和剩余未成交量。 |
| **域归属** | market |
| **主键** | `id BIGSERIAL` |
| **重要字段** | `company_id BIGINT NOT NULL`<br>`resource_id INT NOT NULL`<br>`is_buy BOOLEAN NOT NULL` — `TRUE`=买单，`FALSE`=卖单<br>`price DECIMAL(20,8) NOT NULL` — 单价，高精度对抗通账<br>`quantity INT NOT NULL` — 原始数量<br>`filled_quantity INT NOT NULL DEFAULT 0` — 已成交量<br>`status TEXT NOT NULL DEFAULT 'open'` — `open` / `filled` / `cancelled`<br>`created_at TIMESTAMPTZ NOT NULL` |
| **外键** | `company_id REFERENCES companies(id) ON DELETE CASCADE`<br>`resource_id REFERENCES resource_catalog(id)` |
| **重要索引** | `INDEX(resource_id, is_buy, price)` — 行情撮合查询<br>`INDEX(company_id)` — 查某公司所有挂单<br>`INDEX(status) WHERE status = 'open'` — 活跃挂单扫描 |
| **事务说明** | 创建挂单：买单检查公司资金并冻结/扣减，卖单检查库存并锁定。撮合：UPDATE 双方 order、INSERT trade、UPDATE 双方资金和库存、INSERT ledger 条目。撤销：原子状态变更 + 释放资金/库存。全部事务内完成。 |
| **历史/审计** | 已成交数量 `filled_quantity` 累加不可回退；market_trades 表提供逐笔明细。 |
| **迁移风险** | 高。market 是游戏核心，撮合逻辑重度依赖事务完整性。`price` 精度需与现有 `float64` 对比确认无截断。 |

```sql
CREATE TABLE market_orders (
    id              BIGSERIAL PRIMARY KEY,
    company_id      BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    resource_id     INT NOT NULL REFERENCES resource_catalog(id),
    is_buy          BOOLEAN NOT NULL,
    price           DECIMAL(20,8) NOT NULL,
    quantity        INT NOT NULL,
    filled_quantity INT NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'open',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ
);

CREATE INDEX idx_mo_book ON market_orders(resource_id, is_buy, price) WHERE status = 'open';
CREATE INDEX idx_mo_company ON market_orders(company_id);
```

---

### 1.11 market_trades — 市场成交（追加写）

| 项目 | 值 |
|------|--------|
| **用途** | 每笔撮合成交的不可变记录，用于行情计算和历史回溯。 |
| **域归属** | market |
| **主键** | `id BIGSERIAL` |
| **重要字段** | `buy_order_id BIGINT NOT NULL`<br>`sell_order_id BIGINT NOT NULL`<br>`resource_id INT NOT NULL`<br>`price DECIMAL(20,8) NOT NULL`<br>`quantity INT NOT NULL`<br>`fee DECIMAL(20,8) NOT NULL DEFAULT 0` — 交易手续费<br>`created_at TIMESTAMPTZ NOT NULL` |
| **外键** | `buy_order_id REFERENCES market_orders(id)`<br>`sell_order_id REFERENCES market_orders(id)`<br>`resource_id REFERENCES resource_catalog(id)` |
| **重要索引** | `INDEX(resource_id, created_at)` — K 线/成交量查询<br>`INDEX(buy_order_id)` / `INDEX(sell_order_id)` — 查某订单的成交明细 |
| **事务说明** | 仅在撮合事务中 INSERT，不可单独存在 |
| **历史/审计** | 追加写，永不 UPDATE/DELETE。可用于恢复 ticker 和做市商回测。 |
| **迁移风险** | 低。结构清晰，但与 `model.Trade` 的对齐需要确认 `Quality` 字段的走向。 |

```sql
CREATE TABLE market_trades (
    id            BIGSERIAL PRIMARY KEY,
    buy_order_id  BIGINT NOT NULL REFERENCES market_orders(id),
    sell_order_id BIGINT NOT NULL REFERENCES market_orders(id),
    resource_id   INT NOT NULL REFERENCES resource_catalog(id),
    price         DECIMAL(20,8) NOT NULL,
    quantity      INT NOT NULL,
    fee           DECIMAL(20,8) NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_mt_resource_time ON market_trades(resource_id, created_at);
CREATE INDEX idx_mt_buy_order ON market_trades(buy_order_id);
CREATE INDEX idx_mt_sell_order ON market_trades(sell_order_id);
```

---

### 1.12 market_tickers — 市场行情快照

| 项目 | 值 |
|------|--------|
| **用途** | 每种资源的最新市场行情概览，以 `resource_id` 为唯一维度，持续更新。 |
| **域归属** | market |
| **主键** | `resource_id INT PRIMARY KEY` |
| **重要字段** | `last_price DECIMAL(20,8) NOT NULL` — 最新成交价<br>`volume_24h BIGINT NOT NULL DEFAULT 0` — 24小时成交量<br>`high_24h DECIMAL(20,8) NOT NULL` — 24小时最高价<br>`low_24h DECIMAL(20,8) NOT NULL` — 24小时最低价<br>`updated_at TIMESTAMPTZ NOT NULL` |
| **外键** | `resource_id REFERENCES resource_catalog(id)` |
| **重要索引** | 主键足够 |
| **事务说明** | 每次成交后在撮合事务末尾 UPDATE。若使用物化刷新方式，则用独立定时任务。 |
| **历史/审计** | 当前表仅保持最新快照。如需历史 K 线，使用 `market_trades` 聚合或另建 bar 表。 |
| **迁移风险** | 低。24h 统计值的凌晨归零逻辑需要处理。现有 `GameState` 中有 `DailyTradeVolume` `DailyHighPrice` 等映射字段。 |

```sql
CREATE TABLE market_tickers (
    resource_id INT PRIMARY KEY REFERENCES resource_catalog(id),
    last_price  DECIMAL(20,8) NOT NULL,
    volume_24h  BIGINT NOT NULL DEFAULT 0,
    high_24h    DECIMAL(20,8) NOT NULL,
    low_24h     DECIMAL(20,8) NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

---

### 1.13 ledger_entries — 财务流水（追加写）

| 项目 | 值 |
|------|--------|
| **用途** | 每笔资金变动的不可变审计日志，用于对账、回放和财务报表。 |
| **域归属** | finance |
| **主键** | `id BIGSERIAL` |
| **重要字段** | `company_id BIGINT NOT NULL`<br>`kind VARCHAR(32) NOT NULL` — 分类，如 `market_sale` `production_reward` `bond_issue` `bond_interest` `tax` `order_refund`<br>`amount DECIMAL(20,2) NOT NULL` — 变动金额（正/负）<br>`direction TEXT NOT NULL` — `in` / `out`，冗余但便于查询<br>`balance_after DECIMAL(20,2) NOT NULL` — 变动后公司余额，审计用<br>`metadata JSONB` — 扩展信息：关联 order_id、resource_id、quantity 等<br>`created_at TIMESTAMPTZ NOT NULL` |
| **外键** | `company_id REFERENCES companies(id) ON DELETE CASCADE` |
| **重要索引** | `INDEX(company_id, created_at)` — 查某公司流水<br>`INDEX(kind, created_at)` — 查某类型的所有流水（全局审计） |
| **事务说明** | 每个涉及资金变动的业务事务末尾 INSERT 该表。不可单独存在外部队列写入可能。 |
| **历史/审计** | 追加写，永不 UPDATE/DELETE。`balance_after` 提供链式验证（当前 entry 的 balance_after = 上一 entry 的 balance_after + 本 entry 的 amount）。 |
| **迁移风险** | 中。现有 `model.LedgerEntry` 使用 float64 表示金额与 ID 为 string（可能是 UUID）。需统一为 DECIMAL 与 BIGSERIAL。 |

```sql
CREATE TABLE ledger_entries (
    id           BIGSERIAL PRIMARY KEY,
    company_id   BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    kind         VARCHAR(32) NOT NULL,
    amount       DECIMAL(20,2) NOT NULL,
    direction    TEXT NOT NULL,
    balance_after DECIMAL(20,2) NOT NULL,
    metadata     JSONB DEFAULT '{}',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_le_company ON ledger_entries(company_id, created_at);
CREATE INDEX idx_le_kind ON ledger_entries(kind, created_at);
```

---

### 1.14 financial_snapshots — 财务快照

| 项目 | 值 |
|------|--------|
| **用途** | 定期生成的财务报表快照，用于查看历史财务状态和排行榜指标。 |
| **域归属** | finance |
| **主键** | `id BIGSERIAL` |
| **重要字段** | `company_id BIGINT NOT NULL`<br>`period_start TIMESTAMPTZ NOT NULL` — 统计周期开始<br>`period_end TIMESTAMPTZ NOT NULL` — 统计周期结束<br>`income_statement JSONB NOT NULL` — 损益表，包含收入/支出分类汇总<br>`balance_sheet JSONB NOT NULL` — 资产负债表，包含现金、库存估值、建筑估值等 |
| **外键** | `company_id REFERENCES companies(id) ON DELETE CASCADE` |
| **重要索引** | `UNIQUE(company_id, period_start, period_end)`<br>`INDEX(company_id)` |
| **事务说明** | 由定时任务或管理员手动触发生成，不参与常规业务事务 |
| **历史/审计** | 追加写，每个周期生成一条新记录 |
| **迁移风险** | 低。Phase 1 仅设计，Phase 2 实现生成逻辑。 |

```sql
CREATE TABLE financial_snapshots (
    id              BIGSERIAL PRIMARY KEY,
    company_id      BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    period_start    TIMESTAMPTZ NOT NULL,
    period_end      TIMESTAMPTZ NOT NULL,
    income_statement JSONB NOT NULL,
    balance_sheet    JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_fs_company_period ON financial_snapshots(company_id, period_start, period_end);
```

---

### 1.15 bonds — 债券发行

| 项目 | 值 |
|------|--------|
| **用途** | 公司发行的债券，记录发行参数、当前状态和可流通数量。 |
| **域归属** | finance |
| **主键** | `id BIGSERIAL` |
| **重要字段** | `issuer_company_id BIGINT NOT NULL` — 发行人<br>`face_value DECIMAL(20,2) NOT NULL` — 每张面值<br>`interest_rate DECIMAL(5,4) NOT NULL` — 年化利率，如 0.05 = 5%<br>`total_quantity INT NOT NULL` — 总发行量<br>`issued_quantity INT NOT NULL DEFAULT 0` — 已售出量<br>`status TEXT NOT NULL DEFAULT 'issuing'` — `issuing` / `fully_issued` / `matured` / `defaulted`<br>`created_at TIMESTAMPTZ NOT NULL` |
| **外键** | `issuer_company_id REFERENCES companies(id) ON DELETE RESTRICT` |
| **重要索引** | `INDEX(issuer_company_id)` — 查某公司发行的所有债券<br>`INDEX(status)` |
| **事务说明** | 发行：INSERT bond + 发行人收款 + ledger entry。偿还利息：逐个持有者计算 + debit issuer + credit holder + ledger entries。全部在事务内。 |
| **历史/审计** | 债券表本身记录发行状态变化（issuing → fully_issued → matured）；利息支付由 ledger 和 bond_holdings 记录。 |
| **迁移风险** | 高。现有 `model.Bond` 结构差异较大（有 `OwnerCompanyID` 表示持有人，新设计拆入 bond_holdings）。`Interest` 为 float64，`Amount` 可能对应面值。需仔细映射。 |

```sql
CREATE TABLE bonds (
    id                 BIGSERIAL PRIMARY KEY,
    issuer_company_id  BIGINT NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    face_value         DECIMAL(20,2) NOT NULL,
    interest_rate      DECIMAL(5,4) NOT NULL,
    total_quantity     INT NOT NULL,
    issued_quantity    INT NOT NULL DEFAULT 0,
    status             TEXT NOT NULL DEFAULT 'issuing',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    matured_at         TIMESTAMPTZ
);

CREATE INDEX idx_bonds_issuer ON bonds(issuer_company_id);
CREATE INDEX idx_bonds_status ON bonds(status);
```

---

### 1.16 bond_holdings — 债券持有记录

| 项目 | 值 |
|------|--------|
| **用途** | 每个公司持有的债券数量记录，N:N 关系表。 |
| **域归属** | finance |
| **主键** | `(bond_id, company_id)` 复合主键 |
| **重要字段** | `bond_id BIGINT NOT NULL`<br>`company_id BIGINT NOT NULL` — 持有者<br>`quantity INT NOT NULL` — 持有张数<br>`purchased_at TIMESTAMPTZ NOT NULL` — 首次购买时间 |
| **外键** | `bond_id REFERENCES bonds(id) ON DELETE RESTRICT`<br>`company_id REFERENCES companies(id) ON DELETE CASCADE` |
| **重要索引** | 复合主键；`INDEX(company_id)` |
| **事务说明** | 购买债券：INSERT/UPDATE bond_holdings + debit buyer + credit issuer + ledger entries。利息结算：按 holding 逐条计算。全部事务内。 |
| **历史/审计** | `quantity` 累加变更通过 ledger_entries 的 metadata 可追溯。 |
| **迁移风险** | 中。持有 N:N 关系新设计，旧 `model.Bond` 直接带 `OwnerCompanyID` 即每行一个持有人。 |

```sql
CREATE TABLE bond_holdings (
    bond_id      BIGINT NOT NULL REFERENCES bonds(id) ON DELETE RESTRICT,
    company_id   BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    quantity     INT NOT NULL,
    purchased_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (bond_id, company_id)
);

CREATE INDEX idx_bh_company ON bond_holdings(company_id);
```

---

### 1.17 research_nodes — 科研节点配置（静态）

| 项目 | 值 |
|------|--------|
| **用途** | 定义科技树上的每个可研究节点，包含前置条件与效果。 |
| **域归属** | catalog |
| **主键** | `id INT PRIMARY KEY` |
| **重要字段** | `name VARCHAR(128) NOT NULL`<br>`category VARCHAR(32) NOT NULL` — 分类，如 `crop` `livestock` `logistics`<br>`cost DECIMAL(12,2) NOT NULL` — 研究消耗资金<br>`duration_seconds INT NOT NULL` — 研究基础时长<br>`prerequisites JSONB` — 前置节点 ID 列表 `[1, 2, 5]`<br>`effects JSONB NOT NULL` — 效果描述 `{"unlock_recipe": 12, "output_bonus": 0.1}` |
| **外键** | 无 |
| **重要索引** | `INDEX(category)` |
| **事务说明** | 只读 |
| **历史/审计** | 无 |
| **迁移风险** | 低。纯种子数据。需与前端 research tree 对齐。现有 `model.ResearchProject` 较简单，此表提供更灵活的结构。 |

```sql
CREATE TABLE research_nodes (
    id               INT PRIMARY KEY,
    name             VARCHAR(128) NOT NULL,
    category         VARCHAR(32) NOT NULL,
    cost             DECIMAL(12,2) NOT NULL,
    duration_seconds INT NOT NULL,
    prerequisites    JSONB DEFAULT '[]',
    effects          JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_rn_category ON research_nodes(category);
```

---

### 1.18 company_research — 公司研究进度

| 项目 | 值 |
|------|--------|
| **用途** | 记录每个公司对每个科研节点的研究进度和完成时间。 |
| **域归属** | research |
| **主键** | `(company_id, research_node_id)` 复合主键 |
| **重要字段** | `company_id BIGINT NOT NULL`<br>`research_node_id INT NOT NULL`<br>`started_at TIMESTAMPTZ NOT NULL` — 开始时间<br>`completed_at TIMESTAMPTZ` — 完成时间，NULL = 进行中 |
| **外键** | `company_id REFERENCES companies(id) ON DELETE CASCADE`<br>`research_node_id REFERENCES research_nodes(id)` |
| **重要索引** | 复合主键；`INDEX(company_id)`；`INDEX(company_id, completed_at) WHERE completed_at IS NULL` — 扫描进行中的研究 |
| **事务说明** | 开始研究：检查前置节点（company_research 中该节点 completed_at IS NOT NULL）→ 扣除资金 → INSERT company_research。完成研究：定时器或领取时 UPDATE completed_at + 应用 effects。 |
| **历史/审计** | 节点研究一旦完成不可撤销。资金消耗反映在 ledger。 |
| **迁移风险** | 中。现有 `model.ResearchProject` 是最小实现，新方案需要重建数据模型。前置条件验证由应用层完成。 |

```sql
CREATE TABLE company_research (
    company_id       BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    research_node_id INT NOT NULL REFERENCES research_nodes(id),
    started_at       TIMESTAMPTZ NOT NULL,
    completed_at     TIMESTAMPTZ,
    PRIMARY KEY (company_id, research_node_id)
);

CREATE INDEX idx_cr_in_progress ON company_research(company_id) WHERE completed_at IS NULL;
```

---

### 1.19 executive_catalog — 高管配置（静态）

| 项目 | 值 |
|------|--------|
| **用途** | 定义所有可招募的高管类型模板，包含基础属性、稀有度和招募成本。 |
| **域归属** | catalog |
| **主键** | `id INT PRIMARY KEY` |
| **重要字段** | `name VARCHAR(128) NOT NULL`<br>`rarity INT NOT NULL` — 稀有度，1=普通 2=稀有 3=史诗 4=传说<br>`base_skills JSONB NOT NULL` — 基础技能，如 `{"production_speed": 0.05, "cost_reduction": 0.02}`<br>`salary DECIMAL(12,2) NOT NULL` — 每周期薪资<br>`recruit_cost DECIMAL(12,2) NOT NULL` — 招募一次性费用 |
| **外键** | 无 |
| **重要索引** | `INDEX(rarity)` |
| **事务说明** | 只读 |
| **历史/审计** | 无 |
| **迁移风险** | 低。纯种子数据。现有 `model.Company.Executives` 为 `[]map[string]any`，需要摸底对齐。 |

```sql
CREATE TABLE executive_catalog (
    id            INT PRIMARY KEY,
    name          VARCHAR(128) NOT NULL,
    rarity        INT NOT NULL,
    base_skills   JSONB NOT NULL DEFAULT '{}',
    salary        DECIMAL(12,2) NOT NULL,
    recruit_cost  DECIMAL(12,2) NOT NULL
);

CREATE INDEX idx_ec_rarity ON executive_catalog(rarity);
```

---

### 1.20 company_executives — 公司已招募高管

| 项目 | 值 |
|------|--------|
| **用途** | 每个公司当前雇佣的高管实例及其成长状态。 |
| **域归属** | executives |
| **主键** | `(company_id, exec_catalog_id)` 复合主键（一个公司对同类高管只持有一份） |
| **重要字段** | `company_id BIGINT NOT NULL`<br>`exec_catalog_id INT NOT NULL`<br>`level INT NOT NULL DEFAULT 1` — 当前等级<br>`training_progress INT NOT NULL DEFAULT 0` — 培训进度 0-100<br>`hired_at TIMESTAMPTZ NOT NULL` |
| **外键** | `company_id REFERENCES companies(id) ON DELETE CASCADE`<br>`exec_catalog_id REFERENCES executive_catalog(id)` |
| **重要索引** | 复合主键；`INDEX(company_id)` |
| **事务说明** | 招募：扣除公司资金 → INSERT/UPDATE company_executives + ledger entry。培训/升级在事务内。 |
| **历史/审计** | 等级和培训进度变更由 ledger 记录花费。 |
| **迁移风险** | 中。新设计将高管的唯一的实例与 catalog 分离。现有数据需要摸底。 |

```sql
CREATE TABLE company_executives (
    company_id       BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    exec_catalog_id  INT NOT NULL REFERENCES executive_catalog(id),
    level            INT NOT NULL DEFAULT 1,
    training_progress INT NOT NULL DEFAULT 0,
    hired_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (company_id, exec_catalog_id)
);

CREATE INDEX idx_ce_company ON company_executives(company_id);
```

---

### 1.21 government_orders — 政府订单

| 项目 | 值 |
|------|--------|
| **用途** | 政府发布的采购订单，各公司可以通过 bidding 竞争供货。 |
| **域归属** | government |
| **主键** | `id BIGSERIAL` |
| **重要字段** | `resource_id INT NOT NULL`<br>`quantity INT NOT NULL` — 需求总量<br>`unit_price DECIMAL(20,2) NOT NULL` — 政府最高收购单价<br>`deadline TIMESTAMPTZ NOT NULL` — 交货截止<br>`status TEXT NOT NULL DEFAULT 'open'` — `open` / `awarded` / `delivered` / `expired`<br>`created_at TIMESTAMPTZ NOT NULL` |
| **外键** | `resource_id REFERENCES resource_catalog(id)` |
| **重要索引** | `INDEX(status)` → 活跃订单扫描<br>`INDEX(deadline)` → 到期检查 |
| **事务说明** | 创建：INSERT government_orders。评标：原子状态变更 + 选定 winning bid。交货：验证数量匹配 → 付款。 |
| **历史/审计** | status 变更记录在 metadata 或独立 event 表。 |
| **迁移风险** | 高。现有 `model.GovContract` 字段丰富（`DepositRate`, `Quality`, `WinningPrice` 等），新设计简化了合同模型。需要确认 Phase 1 是否保留所有旧字段。 |

```sql
CREATE TABLE government_orders (
    id          BIGSERIAL PRIMARY KEY,
    resource_id INT NOT NULL REFERENCES resource_catalog(id),
    quantity    INT NOT NULL,
    unit_price  DECIMAL(20,2) NOT NULL,
    deadline    TIMESTAMPTZ NOT NULL,
    status      TEXT NOT NULL DEFAULT 'open',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ
);

CREATE INDEX idx_go_status ON government_orders(status);
CREATE INDEX idx_go_deadline ON government_orders(deadline) WHERE status = 'open';
```

---

### 1.22 government_bids — 政府订单投标

| 项目 | 值 |
|------|--------|
| **用途** | 各公司对政府订单的投标记录。 |
| **域归属** | government |
| **主键** | `(order_id, company_id)` 复合主键 |
| **重要字段** | `order_id BIGINT NOT NULL`<br>`company_id BIGINT NOT NULL`<br>`unit_price DECIMAL(20,2) NOT NULL` — 投标单价<br>`quantity INT NOT NULL` — 投标供货量<br>`created_at TIMESTAMPTZ NOT NULL` |
| **外键** | `order_id REFERENCES government_orders(id) ON DELETE CASCADE`<br>`company_id REFERENCES companies(id) ON DELETE CASCADE` |
| **重要索引** | 复合主键；`INDEX(order_id)` |
| **事务说明** | 投标 INSERT；评标时读取 bids 表选出最优价格。 |
| **历史/审计** | 加入后不可修改，但公司可撤标（DELETE）。 |
| **迁移风险** | 低。新设计，模型简单。 |

```sql
CREATE TABLE government_bids (
    order_id   BIGINT NOT NULL REFERENCES government_orders(id) ON DELETE CASCADE,
    company_id BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    unit_price DECIMAL(20,2) NOT NULL,
    quantity   INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (order_id, company_id)
);

CREATE INDEX idx_gb_order ON government_bids(order_id);
```

---

### 1.23 chat_messages — 聊天消息（追加写）

| 项目 | 值 |
|------|--------|
| **用途** | 游戏内聊天频道消息的持久化存储。 |
| **域归属** | social |
| **主键** | `id BIGSERIAL` |
| **重要字段** | `company_id BIGINT NOT NULL` — 发送者公司<br>`sender_name VARCHAR(64) NOT NULL` — 发送时玩家显示名（冗余，避免 JOIN）<br>`content TEXT NOT NULL`<br>`channel TEXT NOT NULL` — 频道，如 `global` `trade` `system`<br>`created_at TIMESTAMPTZ NOT NULL` |
| **外键** | `company_id REFERENCES companies(id) ON DELETE CASCADE` |
| **重要索引** | `INDEX(channel, created_at)` — 按频道拉取消息<br>`INDEX(created_at)` — 全局按时间排序 |
| **事务说明** | 单行 INSERT，无需跨表 |
| **历史/审计** | 追加写，永不 UPDATE/DELETE。可设置保留策略（如仅保留最近 N 天）。 |
| **迁移风险** | 低。结构简单。`model.Message` 含 `Chatroom`（对应 channel）、`Body`、`From`/`FromID` 等。 |

```sql
CREATE TABLE chat_messages (
    id          BIGSERIAL PRIMARY KEY,
    company_id  BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    sender_name VARCHAR(64) NOT NULL,
    content     TEXT NOT NULL,
    channel     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_cm_channel ON chat_messages(channel, created_at DESC);
CREATE INDEX idx_cm_time ON chat_messages(created_at DESC);
```

---

### 1.24 notifications — 通知

| 项目 | 值 |
|------|--------|
| **用途** | 发送给玩家的系统通知，如订单成交、生产完成、政府订单公告等。 |
| **域归属** | social |
| **主键** | `id BIGSERIAL` |
| **重要字段** | `company_id BIGINT NOT NULL` — 目标公司<br>`kind VARCHAR(32) NOT NULL` — 类型，如 `market_trade` `production_ready` `bond_interest` `gov_contract`<br>`message TEXT NOT NULL` — 通知正文<br>`read BOOLEAN NOT NULL DEFAULT FALSE` — 是否已读<br>`created_at TIMESTAMPTZ NOT NULL` |
| **外键** | `company_id REFERENCES companies(id) ON DELETE CASCADE` |
| **重要索引** | `INDEX(company_id, read, created_at)` — 查某公司未读通知<br>`INDEX(company_id, created_at)` — 查某公司所有通知 |
| **事务说明** | 在业务事务中 INSERT 通知（与主操作在同一事务内），确保通知不丢失。标记已读为独立 UPDATE。 |
| **历史/审计** | 标记已读后不可逆转。可设置自动清理策略。 |
| **迁移风险** | 低。现有 `model.Notification` 结构简单（`Title`+`Body`），新设计使用 `kind`+`message`。 |

```sql
CREATE TABLE notifications (
    id         BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    kind       VARCHAR(32) NOT NULL,
    message    TEXT NOT NULL,
    read       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notif_unread ON notifications(company_id, read, created_at DESC) WHERE read = FALSE;
CREATE INDEX idx_notif_all ON notifications(company_id, created_at DESC);
```

---

## 2. 高风险事务分析

> 以下事务涉及多表互锁、分布式一致性语义（在单数据库内）以及业务完整性约束。
> 所有事务使用 pgx `BeginTx` 或 `pgxpool.BeginTx` 控制，确保 `ROLLBACK` 路径完备。

---

### 2.1 创建挂单（Create Market Order）

**上下文**：玩家下达买单或卖单，必须在 INSERT 前验证资源可用性并预扣。

**事务流程**：

```
BEGIN
  1. 读取 companies 表 FOR UPDATE（锁住公司行，防止并发修改资金）
  2. 如果是买单 (is_buy = TRUE)：
     - CHECK(company.money >= order.price * order.quantity)
     - UPDATE companies SET money = money - (price * quantity) WHERE id = $1
     - 插入 ledger_entries 记录预扣（kind='order_freeze', direction='out'）
  3. 如果是卖单 (is_buy = FALSE)：
     - SELECT quantity FROM warehouse_items WHERE company_id = $1 AND resource_id = $2 FOR UPDATE
     - CHECK(warehouse.quantity >= order.quantity)
     - 卖单在撮合前不扣库存（仅锁定不可用），或扣至"冻结"数量
     - 推荐做法：warehouse_items 增加 frozen_quantity 字段，或卖单时减少可用 quantity 并记录冻结
  4. INSERT INTO market_orders (status = 'open')
COMMIT
```

**风险点**：
- 并发下单导致资金 double-spend：`FOR UPDATE` 锁住 company 行可防。
- 多个卖单同一资源超卖：卖单创建时需要原子减少可用库存（或使用冻结机制）。推荐 warehouse 加 `frozen_quantity` 列，撮合时最终扣减 `quantity`。
- 资金精度：`DECIMAL(20,2)` 累加计算确保无浮点误差。
- **误操作上限**：此事务中资金已被扣除而非简单冻结；撤销挂单需要退款 + 审计（见 2.3）。

**表级锁需求**：`companies` 行锁，`warehouse_items` 行锁（卖单）。

---

### 2.2 撮合成交（Match Market Orders）

**上下文**：系统或做市商检测到买单价格 ≥ 卖单价格时触发撮合。

**事务流程**（批量撮合一个或多个配对）：

```
BEGIN
  1. SELECT buy_order FROM market_orders WHERE id = $buy_id FOR UPDATE
     SELECT sell_order FROM market_orders WHERE id = $sell_id FOR UPDATE
     （锁住两个订单行防止重复撮合）
  
  2. 计算可撮合数量：match_qty = MIN(buy.quantity - buy.filled_quantity, sell.quantity - sell.filled_quantity)
     CHECK(买价 >= 卖价)
     CHECK(match_qty > 0)
     CHECK(两个订单 status = 'open')
  
  3. UPDATE market_orders SET filled_quantity = filled_quantity + match_qty
     WHERE id IN ($buy_id, $sell_id)
     -- 若 filled_quantity = quantity，同时 SET status = 'filled'
  
  4. INSERT INTO market_trades (buy_order_id, sell_order_id, resource_id, price, quantity, fee)
     VALUES ($buy_id, $sell_id, $res_id, match_price, match_qty, fee_amt)
  
  5. 更新买家库存：
     UPSERT INTO warehouse_items (company_id, resource_id, quantity)
     VALUES ($buyer, $res_id, match_qty)
     ON CONFLICT UPDATE SET quantity = warehouse_items.quantity + match_qty
  
  6. 更新卖家库存/冻结：
     UPDATE warehouse_items SET quantity = quantity - match_qty
     WHERE company_id = $seller AND resource_id = $res_id
     -- 或者释放冻结 (quantity - frozen 模式)
  
  7. 资金流转：
     UPDATE companies SET money = money - (match_price * match_qty) WHERE id = $buyer
     UPDATE companies SET money = money + (match_price * match_qty) - fee_amt WHERE id = $seller
  
  8. INSERT INTO ledger_entries × 2（买家支出 + 卖家收入 + 手续费）
  
  9. INSERT INTO notifications × 2（通知买卖双方）
  
  10. UPDATE market_tickers（last_price, volume_24h 等）
COMMIT
```

**风险点**：
- **幽灵撮合**：订单被其他线程先撮合导致数量不足。`FOR UPDATE` 是必需且必须在事务开始时。
- **死锁风险**：同时撮合同组买卖订单从两端进入，可能 A 事务锁 buy→sell，B 事务锁 sell→buy。解决：固定锁顺序（如总是先锁 buy_order_id，再锁 sell_order_id）。
- **资金可用性**：买家在创建挂单时已预扣资金，撮合时无需再扣（退款只需更新 filled_quantity）。但若预扣模式改为只冻结不扣款（推荐），则在撮合时执行实际扣款。
- **库存负数**：UPSERT 买家库存使用 `ON CONFLICT`；卖家 `UPDATE quantity = quantity - match_qty` 后可用 CHECK 或应用层保证不越界。
- **大数据量**：高频撮合循环中批量操作应尽可能一条 SQL 完成。

---

### 2.3 撤销挂单（Cancel Order）

**上下文**：玩家撤销未完全成交的订单，需要退还剩余资金或库存。

**事务流程**：

```
BEGIN
  1. SELECT order FROM market_orders WHERE id = $id FOR UPDATE
     CHECK(status = 'open')
     remaining = quantity - filled_quantity
     CHECK(remaining > 0)
  
  2. UPDATE market_orders SET status = 'cancelled' WHERE id = $id
  
  3. 如果是买单：
     refund = price * remaining
     UPDATE companies SET money = money + refund WHERE id = $company
     INSERT INTO ledger_entries (kind='order_refund', direction='in', amount=refund, ...)
  
  4. 如果是卖单：
     UPDATE warehouse_items SET quantity = quantity + remaining
     WHERE company_id = $company AND resource_id = $res_id
     -- 如果是冻结模式，释放冻结而非加回 quantity
COMMIT
```

**风险点**：
- **重复撤销**：`status = 'cancelled'` 在 UPDATE 前检查，双重防止重复退款。
- **部分成交后撤销**：`remaining = quantity - filled_quantity` 正确计算未成交部分，不影响已成交。
- **资金返还顺序**：在 CREATE 时如果资金已扣除，取消时必须原路退还。若采用冻结模式（资金不扣减），则取消时无资金操作。

---

### 2.4 开始生产（Start Production）

**上下文**：玩家选择一个建筑和配方开始生产，需要扣除输入材料并生成 production_job。

**事务流程**：

```
BEGIN
  1. SELECT recipe FROM recipe_catalog WHERE id = $recipe_id
  2. SELECT building FROM company_buildings WHERE id = $bid FOR UPDATE
     CHECK(building 当前没有运行中的 production_job) -- 或检查 slot 容量
  
  3. 遍历 recipe.inputs：
     SELECT quantity FROM warehouse_items
     WHERE company_id = $cid AND resource_id = $input_res_id
     FOR UPDATE
     CHECK(quantity >= required_amount)
     UPDATE warehouse_items SET quantity = quantity - required_amount
  
  4. INSERT INTO production_jobs (company_id, building_id, resource_id, quantity,
     started_at = now(), duration_seconds = recipe.duration_base * (1 - speed_bonus),
     completed = FALSE, claimed = FALSE)
  
  5. （可选）UPDATE company_buildings SET 当前槽位状态为 "生产中" 或其他表示忙碌的方式
COMMIT
```

**风险点**：
- **并发同建筑多任务**：如果建筑支持多槽位并行生产，需使用 `slot_id` 粒度的互斥锁或在应用层检查。否可能重复开始生产。
- **输入材料被其他事务消耗**：`warehouse_items` 行锁 `FOR UPDATE` 防止超卖。
- **配方数据变化**：配方表为静态数据，但若有百分比 buff（建筑等级、高管技能），`duration_seconds` 应在应用层计算后写入。

---

### 2.5 领取生产产物（Claim Production）

**上下文**：玩家领取已完成的 production_job 的产物，获得资源和经验值。

**事务流程**：

```
BEGIN
  1. SELECT job FROM production_jobs WHERE id = $job_id FOR UPDATE
     CHECK(completed = TRUE)
     CHECK(claimed = FALSE)
     -- completed 由定时任务或按需查询确认：
     -- completed = (now() >= started_at + duration_seconds)
  
  2. UPDATE production_jobs SET claimed = TRUE WHERE id = $job_id
  
  3. UPSERT INTO warehouse_items (company_id, resource_id, quantity)
     VALUES ($cid, job.resource_id, job.quantity)
     ON CONFLICT UPDATE SET quantity = warehouse_items.quantity + job.quantity
  
  4. 计算 XP 奖励（基于配方、数量、建筑等级等，由 formula 包计算）
     UPDATE companies SET xp = xp + xp_reward, level = <计算新等级> WHERE id = $cid
  
  5. （可选）INSERT INTO ledger_entries (kind='production_claim', ...)
COMMIT
```

**风险点**：
- **双倍领取**：`claimed` 在 `FOR UPDATE` 保护下更新，防止并发领取同一 job。
- **completed 检查时序**：如果 `completed` 只是一个计算字段（非物理列），需要在应用层确保 `started_at + duration_seconds <= now()`。定时任务也可在完成时 UPDATE `completed = TRUE`。
- **XP 计算**：`level` 变更可能需要额外的 level 计算公式，涉及 `xpToNextLevel`。推荐放在 formula 层纯函数处理。

---

### 2.6 发行债券（Issue Bond）

**上下文**：公司发行债券，立即获得发行收入（等于面值 × 发行量），未来需付息还本。

**事务流程**：

```
BEGIN
  1. 检查发行人资质（等级、已有负债率等，未来扩展）
  
  2. INSERT INTO bonds (issuer_company_id, face_value, interest_rate,
     total_quantity, issued_quantity = 0, status = 'issuing')
  
  3. 初始发行时发行人自身可能认购一部分（或全量认购）：
     UPDATE companies SET money = money + (face_value * initial_issued_qty)
     WHERE id = $issuer
     UPDATE bonds SET issued_quantity = issued_quantity + initial_issued_qty
  
  4. INSERT INTO bond_holdings (bond_id, company_id = issuer, quantity = initial_issued_qty)
     -- 如果发行人自持
  
  5. INSERT INTO ledger_entries (kind='bond_issue', direction='in', ...)
COMMIT
```

**风险点**：
- **发行人信用风险**：Phase 1 可能不实现信用评级，但预留扩展点（在 companies 表加入 `debt_total` 追踪字段，或从 ledger 实时计算）。
- **债券 ID 暴露**：BIGSERIAL 作为内部 ID，对外接口应使用 UUID 或编码 ID。
- **发行后不可撤销**：如果事务中途失败（如资金更新后 bonds 表写入失败），ROLLBACK 确保一致性。

---

### 2.7 购买债券（Buy Bond）

**上下文**：公司 A 购买公司 B 发行的债券，资金从 A 流向 B。

**事务流程**：

```
BEGIN
  1. SELECT bond FROM bonds WHERE id = $bond_id FOR UPDATE
     CHECK(status IN ('issuing', 'fully_issued'))
     CHECK(bond.issued_quantity < bond.total_quantity)
  
  2. buy_qty = MIN(purchase_qty, bond.total_quantity - bond.issued_quantity)
  
  3. UPDATE companies SET money = money - (bond.face_value * buy_qty)
     WHERE id = $buyer AND money >= (bond.face_value * buy_qty)
     -- CHECK 资金充足（可加 FOR UPDATE 锁定 buyer 行）
  
  4. UPDATE companies SET money = money + (bond.face_value * buy_qty)
     WHERE id = bond.issuer_company_id
  
  5. UPDATE bonds SET issued_quantity = issued_quantity + buy_qty
     -- 若 issued_quantity = total_quantity，SET status = 'fully_issued'
  
  6. UPSERT INTO bond_holdings (bond_id, company_id, quantity)
     VALUES ($bond_id, $buyer, buy_qty)
     ON CONFLICT UPDATE SET quantity = bond_holdings.quantity + buy_qty
  
  7. INSERT INTO ledger_entries × 2（买家支出 + 发行人收入）
COMMIT
```

**风险点**：
- **自买自卖**：应允许发行人在二级市场回购自己债券，但需确保 business logic 允许（可能造成人为价格操纵）。
- **超额发行**：`issued_quantity < total_quantity` 在事务内校验防止超卖。
- **发行人禁售期**：未来可能增加 `matured_at` 前的限制。
- **资金同步扣减**：买家 `FOR UPDATE` 锁定，防止并发购买同一种债券造成 double-spend。

---

### 2.8 结算债券利息（Settle Bond Interest）

**上下文**：付息日系统对每支债券的所有持有者计算并支付利息，按比例分配。

**事务流程**：

```
BEGIN
  1. SELECT bond FROM bonds WHERE id = $bond_id FOR UPDATE
     CHECK(bond.status NOT IN ('matured', 'defaulted'))
  
  2. 利息计算：
     total_interest = bond.face_value * bond.interest_rate * elapsed_years
     -- 简化：每周期（游戏日）结算固定比例
     total_interest_payable = bond.face_value * bond.interest_rate / 365 * days_since_last_pay
     -- 或使用固定周期，比如每 7 天结算一次
  
  3. SELECT SUM(quantity) AS total_outstanding FROM bond_holdings WHERE bond_id = $bond_id
     -- 实际流通量
  
  4. 对于 bond_holdings 中每条记录：
     holding_interest = total_interest_payable * (holding.quantity / total_outstanding)
     
     -- Debit 发行人
     UPDATE companies SET money = money - holding_interest
     WHERE id = bond.issuer_company_id
     CHECK(money >= holding_interest) -- 如果资金不足，标记为违约（defaulted）
  
     -- Credit 持有人
     UPDATE companies SET money = money + holding_interest
     WHERE id = holding.company_id
  
     INSERT INTO ledger_entries × 2 (发行人支出 + 持有人收入)
  
  5. 若发行人资金不足，SET bond.status = 'defaulted'，记录违约。
     后续处理（罚息、资产冻结等）由业务逻辑处理。
  
  6. 若达到到期日：bond.status = 'matured'，归还面值逻辑类似。
COMMIT
```

**风险点**：
- **发行人资不抵债**：遍历 holdings 过程中发行人中途资金不足。解决方案：先计算 total_interest_payable，一次性扣除发行人，然后分发给持有人。不足则全部进入违约处理。
- **循环结算**：如果 A 公司持有 B 债券，B 公司持有 A 债券，且双方互为发行人，结算顺序需固定（如按 bond_id 排序）防止死锁。
- **大事务**：大量 holdings 时，单事务内遍历 INSERT 多条 ledger 可能导致行锁范围过大。考虑拆分为小批次（但需保证原子性，或使用补偿机制）。
- **浮点精度**：`interest_rate DECIMAL(5,4)` 与 `face_value DECIMAL(20,2)` 乘积需要 `DECIMAL(25,6)` 中间量确保精度不丢失。

---

### 2.9 政府合同评标（Gov Contract Award）

**上下文**：政府订单截至后，系统选择最合适的投标（通常是最低价）并授予合同。

**事务流程**：

```
BEGIN
  1. SELECT order FROM government_orders WHERE id = $order_id FOR UPDATE
     CHECK(status = 'open')
     CHECK(deadline <= now()) -- 已截止
  
  2. 读取所有投标：
     SELECT * FROM government_bids WHERE order_id = $order_id
     ORDER BY unit_price ASC, created_at ASC
     -- 最低价优先，同价先到先得
  
  3. 选择 winning_bid（第一名）
  
  4. UPDATE government_orders SET
     status = 'awarded',
     winner_company_id = winning_bid.company_id,
     winning_price = winning_bid.unit_price,
     awarded_at = now(),
     updated_at = now()
  
  5. （可选）通知中标者和未中标者：INSERT INTO notifications × N
  
  6. （可选）中标公司缴纳履约保证金（deposit）：
     UPDATE companies SET money = money - deposit
     WHERE id = winning_bid.company_id
     INSERT INTO ledger_entries
COMMIT
```

**风险点**：
- **两次评标**：`status = 'open'` 条件 + `FOR UPDATE` 防止同订单被两次评标。
- **投标数据在截止后仍然变化**：评标前应禁止新投标（截止检查 + status 保护）。
- **中标违约**：如果中标公司放弃（如资金不足），可选择第二顺位。此时需要事务回滚或补偿操作。

---

### 2.10 政府合同交货（Gov Contract Delivery）

**上下文**：中标公司完成订单，交付资源并获得合同款项。

**事务流程**：

```
BEGIN
  1. SELECT order FROM government_orders WHERE id = $order_id FOR UPDATE
     CHECK(status = 'awarded')
     CHECK(company_id = order.winner_company_id) -- 仅中标公司可交货
  
  2. SELECT quantity FROM warehouse_items WHERE company_id = $cid AND resource_id = $res_id
     FOR UPDATE
     CHECK(quantity >= order.quantity) -- 检查库存足够
  
  3. UPDATE warehouse_items SET quantity = quantity - order.quantity
     WHERE company_id = $cid AND resource_id = $res_id
  
  4. total_payment = winning_price * order.quantity
     UPDATE companies SET money = money + total_payment WHERE id = $cid
  
  5. UPDATE government_orders SET status = 'delivered', delivered_at = now(), updated_at = now()
  
  6. INSERT INTO ledger_entries (kind='gov_delivery', direction='in', amount=total_payment, ...)
  
  7. （可选）退还履约保证金 + 额外奖励或罚款
  
  8. INSERT INTO notifications (kind='gov_delivery', ...)
COMMIT
```

**风险点**：
- **分段交付**：当前设计不支持分批交付（总数量一次性扣除）。如果未来需要分批交付，需在 government_orders 增加 `delivered_quantity` 字段，并放宽 status 变化规则。
- **超时违约**：如果 `deadline` 已过但未交付，系统需要定时任务处理违约：`status = 'expired'`，扣除保证金等。此逻辑在事务外处理。
- **价格操纵**：`winning_price` 在评标时已锁定，交货时按该价格付款。不需要重新询价。

---

## 3. 索引策略总结

| 场景 | 索引模式 | 示例 |
|------|----------|----------|
| 按外键查询 | B-tree INDEX | `INDEX(company_id)` 几乎出现在每个业务表 |
| 按状态过滤活跃数据 | 部分索引 `WHERE status = 'open'` | market_orders, government_orders |
| 行情撮合查询 | 复合索引 (resource_id, is_buy, price) WHERE open | market_orders |
| 追加写时间序查询 | (channel, created_at DESC) | chat_messages, notifications |
| 空间查询（建筑位置） | (company_id, map_id, slot_id) UNIQUE | company_buildings |

---

## 4. 外键约束策略

| 策略 | 适用场景 |
|--------|-------------|
| `ON DELETE CASCADE` | 子表依赖父表存在且无独立业务含义：company_buildings、warehouse_items、production_jobs、chat_messages、notifications |
| `ON DELETE RESTRICT` | 子表有独立业务含义，不能因父表删除而无声消失：bonds（持有关系）、bond_holdings |
| `ON DELETE SET NULL` | 如未来引入软删除场景 |

> **注意**：在迁移过渡期（内存+数据库双写），外键约束可能导致一致性问题。建议在正式切换后才启用外键，或在应用层维护引用完整性。

---

## 5. 迁移策略与风险

| 表组 | 迁移复杂度 | 策略 |
|-------|-----------------|--------|
| 静态配置（4 表） | 低 | 直接 INSERT 种子数据，每次部署重新同步 |
| auth（2 表） | 中 | 密码 hash 可直接迁移；session 表从 JWT 无状态模式变更，需平滑过渡 |
| 核心业务（6 表） | 高 | companies、warehouse_items、company_buildings、production_jobs、market_orders、market_trades 需从内存 GameState 逐字段映射，建议 clone + diff 验证 |
| finance（4 表） | 中 | ledger_entries 和 bonds 结构变化较大，需业务逻辑调整配合 |
| research（2 表） | 中 | 新 research_nodes 需要将旧的 ResearchProject 映射到节点树 |
| government（2 表） | 高 | 现有 GovContract 模型简化，需确认字段剪裁范围 |
| social（2 表） | 低 | 结构接近，迁移简单 |

### 推荐迁移顺序

```
Phase 0: 静态配置（catalog 表） — 无风险，先行导入
Phase 1: players + companies + warehouse_items — 核心实体
Phase 2: market_orders + market_trades + market_tickers — 市场核心
Phase 3: production_jobs + building/recipe catalog — 生产核心
Phase 4: ledger_entries + bonds + bond_holdings — 财务
Phase 5: research + executives — 扩展系统
Phase 6: government_orders + government_bids — 政府订单
Phase 7: chat_messages + notifications + auth_sessions — 附属系统
```

每个阶段完成后运行全量 go test，并通过 QA 环境验证前端可正常交互。

---

## 6. 附录：在建表前需要确认的关键问题

1. **市场资金模式**：创建买单时是否立即扣款（pre-debit）还是仅冻结（reserve）？本设计推荐冻结模式，取消时无需退款操作。但冻结模式需要增加 `frozen_money` 或 `reserved_money` 字段。
2. **库存冻结**：卖单是否需要冻结库存？推荐类似资金冻结，warehouse_items 增加 `frozen_quantity`。
3. **精度统一**：所有金额字段使用 `DECIMAL(20,2)` 是否足够？市场价可能需要更高精度（如 `DECIMAL(20,8)`）。
4. **函数 vs 计算列**：`production_jobs.completed` 用 BOOLEAN 列（定时任务触发）还是生成列（`started_at + duration_seconds <= now()`）？当前设计推荐 BOOLEAN 列便于索引和查询。
5. **UUID vs BIGSERIAL**：对外暴露的 ID 使用 UUID 还是 BIGSERIAL 编码？本设计使用 BIGSERIAL 作为内部主键，对外接口建议用编码 ID 或独立 public_id 列。
