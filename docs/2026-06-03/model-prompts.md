# Model Prompts - 2026-06-03

这些提示词用于让模型按任务包逐步操作。使用时把对应提示词复制给模型，并附上必要的文件路径或代码仓库上下文。

通用要求：

- 先读代码，再判断，不要只按文档猜。
- 保留现有代码风格和目录结构。
- 不要实现暂缓清单中的功能。
- 不要引入工人管理、RecipeWorkload、全局 UnitsSold、全局 PriceAcceptance。
- 涉及用户可见文案时，`SimBoost` 统一显示为 `Power-up`。
- 所有修改必须给出验证方式。

## Prompt 1 - 代码现状审计（运行次数1次）

```text
你是一个资深全栈工程师。请审计当前 NewHaven 代码库，目标是把 2026-06-03 的计划落到真实代码状态上。

请先读取：
- docs/2026-06-03/source-digest.md
- docs/2026-06-03/new-plan.md
- 前端 src 目录
- 后端 internal 目录
- config、formula、storage、test 相关目录

任务：
1. 梳理前端路由、导航、API hook、feature 目录结构。
2. 梳理后端 handler、service、storage、formula、middleware 的结构。
3. 对 Research、Financial、Executive、Chat、Power-up、Leaderboard 分别判断：已完成、部分完成、缺失、接口不匹配。
4. 核查 BuildView 建筑市场 API 当前实际调用路径，并判断应使用 /api/v2/buildings/ 还是 /api/v2/buildings/market/。
5. 核查经济公式是否符合“建筑产量 + 成本指数 + 饱和度控价 + 自动交易机器人”的极简口径。
6. 输出一个任务风险表，按 P0/P1/P2 排序。

限制：
- 这一步只做审计，不改代码。
- 不要实现 Bond、Auction、Aerospace、WebSocket、资源图标补齐。

交付：
- 审计报告。
- 建议的下一步最小实现任务。
- 发现的接口或公式差异。
```

## Prompt 2 - Research Page 实现（运行次数1次）

```text
你是当前代码库的前端实现工程师。请基于审计结果实现 Research 页面。

背景：
- Research 是优先功能。
- 侧边栏已有入口，但 activeView === 'research' 时页面为空或缺失。
- 研究系统应对齐 docs/2026-06-03/source-digest.md 中的 Research Curve：研究节点、RequiredProductGroup、BaseFoodResearchQty、FloatingFoodResearchQty、FoodResearchPrice、MarketPurchaseCost、CashFee、TotalResearchCost。

需要接入 API：
- GET /api/v2/research/
- POST /api/v2/research/start/
- GET /api/v2/research/progress/
- POST /api/v2/research/complete/

实现要求：
1. 先读现有 feature、API client、hook、页面布局、导航写法。
2. 按现有模式创建或补齐 ResearchPage。
3. 页面需要包含：研究项目列表、消耗展示、开始按钮、当前进度、剩余时间、完成/领取按钮。
4. 处理 loading、empty、error、操作中状态。
5. 不新增特殊图片资产，可用现有 icon。
6. 不做 WebSocket，进度用轮询或手动刷新，按现有代码习惯选择。

验收：
- 点击 Research 导航能看到页面。
- 能加载研究项目。
- 能开始、查看进度、完成研究。
- 错误状态不会破坏布局。
- 提供测试或手动验证步骤。
```

## Prompt 3 - Executive Page 实现（运行次数1次）

```text
请实现 Executive 页面，只做高管市场、招募、培训、详情，不做挖角和报价管理。

背景：
- 高管系统必须对齐 docs/2026-06-03/source-digest.md 中的 Executive Curve。
- 高管加成和全局口碑应该相加，不要相乘。
- UI 需要展示高管工资、生产加成、销售加成、管理折扣、训练成本和等级阶段。

需要接入 API：
- POST /api/v2/executives/search/
- POST /api/v2/executives/recruit/
- POST /api/v2/executives/train/{id}/
- GET /api/v3/executives/{id}/

实现要求：
1. 先读取现有 API client、feature 页面、卡片/表格/弹窗组件风格。
2. 新增或补齐 src/features/executives/。
3. 页面包含：搜索/筛选、候选高管列表、已招募高管、详情区域、训练操作。
4. 挖角和报价管理入口不要做，避免半成品。
5. 所有操作要有 loading、error、成功刷新。

验收：
- 能搜索可用高管。
- 能招募高管。
- 能训练已招募高管。
- 能查看高管详情。
- 高管数值文案与经济模型一致。
```

## Prompt 4 - Financial Pages 实现（运行次数1次）

```text
请实现 Financial 页面或财务面板，优先让玩家能观察经营结果。

需要接入 API：
- GET /api/v2/companies/me/income-statement/
- GET /api/v2/companies/me/balance-sheet/
- GET /api/v2/companies/me/cashflow-statement/
- GET /api/v2/companies/me/cashflow/recent/
- GET /api/v3/companies/me/past-finances/

实现要求：
1. 先确认现有导航和 TopBar 资金显示逻辑。
2. 可以选择 TopBar 资金点击打开财务面板，或新增 Financial 导航项；选择要符合现有 UI。
3. 展示损益表、资产负债表、现金流量表、最近流水、历史净利润趋势。
4. 历史趋势如现有图表库可用就用图，否则先用清晰表格。
5. 处理 loading、empty、error。

验收：
- 玩家能从主界面进入财务视图。
- 5 类财务数据均可读取并展示。
- 最近流水最多展示合理数量，避免页面过长。
- 提供验证步骤。
```

## Prompt 5 - Chat API 接入（运行次数1次）

```text
请把 ChatPanel 从本地 mock 数据切换到后端 API。

需要接入 API：
- GET /api/messages/
- POST /api/messages/
- POST /api/v2/message/{id}/read/
- GET /api/v2/chatroom/
- GET /api/v2/contacts/

要求：
1. 先读现有 ChatPanel、mock 数据来源、API client/hook 模式。
2. 尽量保留现有 UI，不做大改版。
3. 新增或复用 hooks：加载消息、发送消息、标记已读、加载聊天室、加载联系人。
4. 消息发送后要刷新或乐观更新，按现有项目风格选择。
5. 处理失败、空消息、联系人为空。

验收：
- 不再依赖 mock 消息作为主要数据源。
- 能发送消息。
- 能标记已读。
- 聊天室和联系人可加载。
```

## Prompt 6 - Power-up UI（运行次数1次）

```text
请实现 Power-up UI。注意：后端 API 可能叫 simboost，但用户可见名称必须是 Power-up，不要显示 SimBoost。

需要接入 API：
- GET /api/v2/players/simboosts/
- POST /api/v2/players/simboosts-use/
- GET /api/v2/players/simboosts-use/ 或按后端实际定义读取活跃状态

要求：
1. 先确认 TopBar 和现有弹出面板/按钮模式。
2. 在 TopBar 增加一个简洁入口，点击打开 Power-up 面板。
3. 面板展示可用 Power-up、剩余次数、效果、活跃状态和剩余时间。
4. 使用后刷新状态。
5. 文案不得出现 SimBoost。

验收：
- TopBar 能打开 Power-up 面板。
- 能查看可用道具。
- 能使用道具。
- 活跃状态可见。
```

## Prompt 7 - Auth 与多 Player 后端（运行次数1次）

```text
请实现真实认证和多 Player 隔离。先读代码，不要直接重写认证系统。

背景：
- 当前文档指出登录/注册可能用 username 做 key，dev mode 固定 dev-player，token 可能是 base64(username) 伪造。
- 目标是密码哈希、签名 JWT、多玩家隔离。

重点文件：
- internal/handler/auth.go
- internal/service/service.go
- internal/middleware/auth.go
- model 和 storage 中与 player/company 相关文件

要求：
1. 审计当前认证实现。
2. 使用安全密码哈希。
3. 使用签名 JWT，并从 token 解析 playerId。
4. dev mode 与真实认证边界清晰。
5. 注册、登录、鉴权、多玩家隔离都要有测试。

验收：
- 非 dev 模式不再固定 dev-player。
- token 不能由 base64(username) 伪造。
- 不同玩家的数据不能串。
```

## Prompt 8 - Postgres 持久化审计与补齐（运行次数1次）

```text
请审计并补齐 Postgres 持久化。

重点文件：
- internal/storage/storage.go
- internal/storage/postgres.go
- scheduler 保存逻辑
- model 目录

任务：
1. 对比内存 storage 和 Postgres storage 的接口覆盖。
2. 找出未实现或行为不一致的方法。
3. 补齐核心 CRUD。
4. 确认首次运行建表或迁移策略。
5. 确认 Scheduler 的 SaveAll 或等价逻辑在 Postgres 模式下合理。

验收：
- 内存模式和 Postgres 模式核心行为一致。
- 玩家、公司、建筑、库存、市场订单、财务记录至少覆盖。
- 有 storage 层测试。
```

## Prompt 9 - Leaderboard API 与页面（运行次数1）

```text
请新增排行榜功能，包括后端 API 和前端页面。

需求：
- 按净资产、等级或关键经营指标排名。
- 支持分页。
- 只展示允许公开的玩家/公司字段。
- 前端需要一个独立页面或清晰入口。

实现要求：
1. 先读现有 company/player/finance 数据结构。
2. 设计 GET /api/v2/leaderboard/ 或符合项目版本习惯的路径。
3. 支持 sort 参数，例如 net_worth、level、profit。
4. 前端展示排名、公司/玩家名、关键指标。
5. 添加后端测试和前端验证步骤。

验收：
- API 返回稳定分页结果。
- 排名计算正确。
- 前端能切换排序或至少展示默认净资产排名。
```

## Prompt 10 - 经济公式对齐与数值回归（运行次数1）

```text
请审计并对齐经济公式，但不要盲目大改。先输出差异，再实现必要修改。

必须读取：
- docs/2026-06-03/source-digest.md
- formula/production.go
- formula/bonds.go
- service/market_competition.go
- service/government.go
- service/market_trade.go
- config/game.json 或相关配置

检查点：
1. 生产是否直接由 BuildingBaseOutputPerHour_Lv1 驱动。
2. 是否删除或避免使用 WorkerCount、RecipeWorkload、全局 UnitsSold、全局 PriceAcceptance。
3. 成本是否由 LaborCostIndex、MaterialCostIndex、EnergyCostIndex 控制。
4. 饱和度是否按 ProductGroup 计算并影响 EffectivePrice。
5. SaturationPriceMultiplier 是否按 CLAMP(0.70, 1.10, 1 - MAX(0, saturation - 1) * SaturationK) 或等价逻辑实现。
6. 高管加成和全局口碑是否相加。
7. 自动交易机器人是否有价差、库存限制、预算限制、参与限制。
8. 等级甜蜜点后是否有管理成本收敛。
9. 债券 DailyBondInterest 的 50x 系数是否合理，并给出保留或调整理由。

输出：
- 先给差异清单。
- 再给建议修改。
- 经确认或在任务允许时再实现。

验收：
- 有数值回归测试样例。
- 修改后的 Lv1 建筑利润应接近目标每小时利润 6000 的平衡口径。
- 机器人不能制造无风险套利。
```

## Prompt 11 - 测试扩容

```text
请按优先级扩容测试，不做新功能。

优先模块：
- formula/
- service/market_trade
- service/bond
- service/government
- service/production
- auth
- storage
- anticheat
- aml

要求：
1. 先统计当前测试文件和覆盖范围。
2. 优先给纯函数和关键服务路径补测试。
3. 每个新增测试都要有明确业务断言，不要只测不 panic。
4. 经济公式测试要包含固定数值样例，防止以后调参误伤。
5. 如果某模块需要重构才能测，先说明，不要大范围重构。

验收：
- 新增测试可运行。
- 覆盖 happy path、边界、失败路径。
- 输出测试命令和结果。
```

