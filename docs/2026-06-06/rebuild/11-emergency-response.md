# 突发事件处理手册 — NewHaven 项目重建期间

**版本**: 1.0  
**日期**: 2026-06-06  
**范围**: 重建期间（Phase 1 → Phase 7）所有可能出现的突发事件应对方案

---

## 0. 总原则

```txt
发现异常 → 立即停止当前操作 → 评估影响范围 → 选择应对策略 → 修复 → 复盘
```

**不要**：
- 在不确定影响范围的情况下继续推进
- 在没有测试的情况下直接修复
- 为了赶进度跳过回滚步骤
- 隐藏问题（没有坏消息只有坏消息被延迟）

---

## 1. 突发事件分类矩阵

| 类别 | ID | 事件 | 严重度 | 响应时间 | 响应人 |
|------|----|------|--------|---------|-------|
| 构建 | E-BUILD-01 | `go build` 失败 | Critical | 立即 | Backend |
| 构建 | E-BUILD-02 | 前端 `npm run build` 失败 | Critical | 立即 | Frontend |
| 构建 | E-BUILD-03 | `go test ./...` 失败 | High | 1 小时内 | QA + 责任人 |
| 经济 | E-ECON-01 | 公式测试失败（数值变化） | Critical | 立即 | Game Economy |
| 经济 | E-ECON-02 | 生产/市场收益异常 | Critical | 立即 | Game Economy |
| API | E-API-01 | 旧 API 响应格式改变 | Critical | 立即 | Backend |
| API | E-API-02 | 前端调用 API 返回 404/500 | Critical | 立即 | Backend + Frontend |
| 数据 | E-DATA-01 | 数据库 migration 导致数据丢失 | Critical | 立即 | Database |
| 数据 | E-DATA-02 | 静态 JSON 加载失败 | High | 1 小时内 | Backend |
| 安全 | E-SEC-01 | 认证绕过或 token 泄露 | Critical | 立即 | Backend |
| 安全 | E-SEC-02 | 反作弊被绕过 | High | 2 小时内 | Backend |
| 实时 | E-REALTIME-01 | WebSocket 推送错误状态 | Medium | 4 小时内 | Backend |
| 调度 | E-SCHED-01 | Scheduler tick 未执行 | High | 1 小时内 | Backend |
| 调度 | E-SCHED-02 | Bot 市场行为异常 | Medium | 4 小时内 | Backend |
| 流程 | E-PROC-01 | PR 合入了未批准的代码 | Medium | 审查后决定 | Architect |
| 流程 | E-PROC-02 | 发现之前合并的 PR 有隐藏 bug | High | 视影响范围 | 责任人 |

---

## 2. 各类突发事件应对流程

### 2.1 E-BUILD-01: `go build` 失败

**症状**: 
- `go build ./cmd/simapi/` 返回非零退出码
- CI 构建失败

**可能原因**:
- 编译错误（类型不匹配、导入循环、未使用的变量）
- go.mod 依赖版本冲突
- Go 版本不兼容

**应对流程**:

```
1. 立即查看编译错误信息
   → go build ./cmd/simapi/ 2>&1
   
2. 如果是最近 PR 引入:
   → git log --oneline -5
   → 找到引入问题的 commit
   
3. 修复优先级:
   a) 如果是简单编译错误（类型名拼错、import 路径不对）
      → 修复并自测
   b) 如果是 go.mod 冲突
      → go mod tidy 后重试
      → 如果仍有问题，检查 go.sum 是否需要更新
   c) 如果是导入循环
      → 检查新增的 interface 或类型是否放错包
      → 将共享类型移到更低层包或新建包
   d) 如果是 Go 版本问题
      → 检查 go.mod 中的 go 版本（当前 1.25）
      → 检查是否使用了 1.26+ 才支持的特性

4. 如果 30 分钟内无法修复:
   → 回滚最后一个 PR
   → 在独立分支上慢慢修
```

**验证**: `go build ./cmd/simapi/` → exit 0  
**回滚**: `git revert <commit-hash>` 或 `git reset --hard HEAD~1`

---

### 2.2 E-BUILD-03: `go test ./...` 失败

**症状**: 测试失败，可能是已有测试失败或新增测试失败

**可能原因**:
- 业务逻辑被非预期修改
- 测试用例依赖于特定状态（时间、随机数）
- 新增代码与现有测试假设冲突
- 公式变动导致 golden test 失败

**应对流程**:

```
1. 运行失败测试，确认失败模式:
   → go test -v -run <TestName> ./package/path/
   
2. 判断失败原因:
   a) 公式测试数值变化 → E-ECON-01 流程
   b) 业务逻辑测试失败:
      → 检查最近修改是否改变了行为
      → 如果是预期行为变更 → 更新测试 + 文档说明
      → 如果是非预期变更 → 修复代码
   c) 测试本身问题:
      → 测试依赖外部状态（时间、随机）→ 加 mock
      → 测试太脆弱 → 加强测试弹性

3. 如果确定是测试需要更新（行为变了且认可）:
   → 更新测试
   → 确保 Formula golden test 也同步更新
   → 在 PR 描述中注明"行为变更"

4. 如果是不确定的测试失败:
   → 不要跳过测试
   → 不要注释测试
   → 找 QA/Architect 确认
```

**验证**: `go test ./...` → all pass  
**注意**: formula 测试失败必须走 E-ECON-01，不可直接更新测试值。

---

### 2.3 E-ECON-01: 公式测试失败（数值变化）

**这是最严重的事件之一**。

**症状**: formula_test.go 中的 golden test 数值与预期不符

**可能原因**:
- 公式实现被修改
- 公式输入参数默认值被修改
- `game.json` 中的经济参数被修改
- 公式调用的常量被修改（如 `BondFaceValue`、`RetailZor`）
- `SetBondFaceValue` 未在 main.go 中调用

**应对流程**:

```
1. 立即停止当前重建工作

2. 定位变更来源:
   → git diff formula/  # 检查公式代码
   → git diff configs/game.json  # 检查经济参数
   → git diff model/types.go  # 检查常量或默认值

3. 判断是有意还是无意变更:
   a) 有意变更（设计决策）:
      → 必须确认是否经过 Game Economy 角色批准
      → 更新 game-formulas-v2.md
      → 更新 golden test 基准值
      → 在 PR 描述中突出标注"经济公式变更"
      
   b) 无意变更（bug）:
      → 立即修复
      → 检查是否有其他公式受影响
      → 确认当前 master 上的行为未受影响

4. 验证修复:
   → go test ./internal/formula/ -v
   → 对比变化前后的公式输出
```

**验证**: 
- `go test ./internal/formula/` → all pass
- 变更必须有文档记录（更新 game-formulas-v2.md）
- 如果是意外变更，还要确认没有其他模块已依赖了错误值

---

### 2.4 E-API-01: 旧 API 响应格式改变

**症状**: 前端页面显示异常、字段缺失、类型错误

**可能原因**:
- handler 返回的 `map[string]any` 中某个 key 被改名
- domain model 字段名被修改导致 JSON 序列化结果变化
- Domain model 内部结构被直接序列化为 API 响应
- DTO 转换层未正确处理旧字段
- 字段名从 snake_case 变为 camelCase 或反过来

**应对流程**:

```
1. 确定受影响的前端页面:
   → 检查该 API 在前端的所有调用点
   → grep -r "api/xxx" client/src/
   
2. 回滚或修复:
   a) 如果是 DTO 转换遗漏字段:
      → 补充 DTO 字段映射
      → 加 contract test 验证新老响应字段集合一致
   b) 如果是 domain model 字段名改了但 DTO 没同步:
      → 在 DTO 中保留旧字段名（json:"old_name"）
      → 同时支持新旧两个字段（兼容期）
   c) 如果是无意中改了 JSON tag:
      → 恢复旧 JSON tag
   
3. 验证:
   → 在测试中对比新老 API 响应的 JSON key 集合
   → 使用 shadow mode 对比实际响应
```

**验证**: 新旧 API 响应的 JSON key 集合相同（新响应可以多字段但不能少字段）

---

### 2.5 E-DATA-01: 数据库 Migration 导致数据丢失

**症状**: 
- 数据查询返回空或错误
- 数据完整性约束违反
- 玩家反馈"我的建筑/库存/钱不见了"

**可能原因**:
- Migration `down` 删除了不应该删的表或列
- 数据迁移脚本遗漏了部分数据
- Transaction 未正确处理导致部分写入
- JSONB 到规范化表的转换过程中字段映射丢失

**应对流程**:

```
1. 立即停止所有 migration 操作
2. 如果有备份，从备份恢复:
   → 检查备份时间点
   → 确认备份覆盖范围（全量/增量）
   
3. 如果没有备份，判断可否回滚 migration:
   → goose down 或 migrate down
   → 检查 down migration 是否完整

4. 确定影响范围:
   → 哪些表/列受影响
   → 哪些用户受影响
   → 是否有经济损失（需要补偿）

5. 修复方案:
   a) 如果是字段映射丢失:
      → 写数据修复 migration
      → 从源表恢复数据
   b) 如果是 schema 设计问题:
      → 回滚到上一个稳定版本
      → 重新设计 migration 顺序
```

**验证**: 数据完整性检查（行数、关联约束、业务断言）  
**预防**: 每次 migration 前都要 `SELECT COUNT(*)` 备份关键表行数

---

### 2.6 E-SEC-01: 认证绕过或 Token 泄露

**症状**: 非授权用户访问了授权资源、异常登录行为

**可能原因**:
- JWT 验证中间件被移除或 bypass
- JWT signing key 被硬编码泄露
- middleware chain 顺序错误（auth 被放在 recovery 外面）
- withAuth 包装器被某个新路由跳过

**应对流程**:

```
1. 立即检查所有路由的认证注册:
   → 对比当前路由注册 vs 路由清单
   → 确认是否有路由忘记用 withAuth 包装
   
2. 检查 middleware chain:
   → main.go 中的 handler 链
   → chi 中间件顺序（如果是 Phase 3+）

3. 如果是 middleware 问题:
   → 修复 middleware chain
   → 确认 Recovery → RequestID → Logger → CORS → Auth 顺序
   
4. 如果是 JWT signing key 泄露:
   → 立即轮换 key
   → 通知所有玩家重新登录
   → 废弃所有旧 token
```

**验证**: 所有非认证端点（`/healthz`, `/readyz`）可以访问；所有其他端点返回 401

---

### 2.7 E-SCHED-01: Scheduler Tick 未执行

**症状**:
- 市场没有周期性刷新
- 债券利息未结算
- Bot 市场流动性不足
- 政府订单未授标

**可能原因**:
- Scheduler 未启动或过早退出
- Scheduler 中的某个 tick action panic 导致后续 action 不执行
- Scheduler 被重构时误删了某个 action
- 配置变更导致 tick interval 变为 0 或负数

**应对流程**:

```
1. 检查 scheduler 运行状态:
   → 查看日志中是否有 tick 日志
   → 检查 scheduler.Start() 是否被调用
   
2. 如果是 panic 导致链中断:
   → 每个 tick action 必须用 recover 包裹
   → 单个 action 失败不应影响其他 action
   
3. 修复:
   a) 缺少 recover:
      → 给每个 action 加 defer recover
   b) action 被误删:
      → 恢复 action 调用
   c) 配置问题:
      → 检查 game.json 中的 scheduler 参数
```

**验证**: 日志中出现规律的 tick 日志，每个 action 都有执行记录

---

## 3. 紧急回滚流程

当发现突发事件无法在限定时间内修复时，执行回滚。

### 3.1 代码回滚

```bash
# 回滚整个 PR
git revert <merge-commit-hash>
git push origin main

# 或者回滚到特定 commit
git reset --hard <last-known-good-commit>
git push --force-with-lease origin main
```

### 3.2 Migration 回滚

```bash
# goose 回滚
goose down

# 回滚到特定版本
goose down-to <version>
```

### 3.3 路由回滚

如果使用了 chi bridge：

```bash
# 将所有域切回 old 模式
export SIM_API_ROUTE_SYSTEM=old
export SIM_API_ROUTE_AUTH=old
export SIM_API_ROUTE_COMPANY=old
# ... 全部设为 old
重启后端
```

### 3.4 前端回滚

```bash
# 切换到上一个部署版本
git checkout <previous-deploy-tag>

# 或者如果只是 API 连接问题，切到旧 API
# 修改 .env 中的 API_BASE_URL 回到旧后端
```

---

## 4. 重点防范时间段

以下时间段是突发事件高发期，需要特别警惕：

| 时间段 | 风险 | 应对 |
|--------|------|------|
| Phase 1 → 2 切换（chi 迁移时） | 路由没配好、旧路由丢失 | 用 script/audit-routes.sh 对比前后路由清单 |
| Phase 2 （OpenAPI 引入后） | DTO 字段映射遗漏 | 全量 contract test + shadow mode |
| Phase 3 （service 切分时） | 旧 facade 未正确转发 | 每个域拆分后 immediate shadow test |
| Phase 4 （数据库 migration） | 数据丢失 | migration 前备份、验证 up/down 对称 |
| Phase 5 （前端重构） | API 调用断裂 | 逐个页面验证，不要批量切换 |
| Phase 6 （WebSocket 上线） | 状态不一致 | WS 事件 + REST 双重验证 |
| Phase 7 （Formula 迁移） | 经济数值漂移 | 全量 golden test 回归 |

---

## 5. 复盘模板

每次突发事件处理后，必须填写复盘记录：

```markdown
## 突发事件复盘

**事件 ID**: E-XXX-XXX
**日期**: 2026-06-XX
**发现人**: 
**响应时间**: 

### 事件描述

（发生了什么）

### 根因

（为什么发生）

### 影响范围

（哪些用户 / 功能 / 数据受影响）

### 处理过程

（按时间线记录做了什么）

### 修复措施

（具体改了什么文件 / 配置）

### 预防措施

（如何防止再次发生）

### 验收

（证明修复有效的证据）
```

---

## 6. 联系人和值班表

| 角色 | 负责人 | 紧急联系电话 |
|------|--------|-------------|
| Architect | （项目所有者） | — |
| Backend | （项目所有者） | — |
| Database | （项目所有者） | — |
| Frontend | （项目所有者） | — |
| QA | （项目所有者） | — |
| Game Economy | （项目所有者） | — |

*单人开发模式下，所有角色由同一个人承担，但突发事件处理必须按流程走，不能跳步骤。*

---

*本文档是 NewHaven 重建计划的一部分。突发事件处理后必须更新本文档（添加新的应对场景）。*
