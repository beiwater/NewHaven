# NewHaven 最新 API 文档

本文档面向当前统一的最新 API，按业务模块拆分为多个独立 Markdown 文件，便于快速查找、复制示例和后续维护。

> 说明：本文档不直接展示后端的 `/api/v1`、`/api/v2`、`/api/v3` 前缀。
> 统一采用虚拟路由标识：`/V1/` 表示非加密 / 公共接口，`/E1/` 表示加密 / 需要登录的接口。
> 实际后端路径以源码和 OpenAPI 为准。

## 文档结构

- [`00-conventions.md`](./00-conventions.md): 通用约定、认证、响应信封、错误码
- [`01-user-company.md`](./01-user-company.md): 用户 / 公司
- [`02-market-trading.md`](./02-market-trading.md): 市场 / 交易 / 行情
- [`03-buildings.md`](./03-buildings.md): 建筑 / 货架
- [`04-warehouse.md`](./04-warehouse.md): 仓库
- [`05-production.md`](./05-production.md): 生产

## 适用范围

本套文档优先覆盖以下核心模块：

- 用户
- 公司
- 交易
- 市场
- 建筑
- 仓库
- 生产

## 来源说明

- 示例请求 / 响应优先参考 [`heritage/API-EXAMPLES.md`](../../heritage/API-EXAMPLES.md)
- 端点清单参考 [`heritage/API-QUICKREF.md`](../../heritage/API-QUICKREF.md)
- schema 与字段定义参考 [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml)

## 阅读建议

1. 先看 [`00-conventions.md`](./00-conventions.md) 统一请求和响应格式
2. 再按业务模块进入对应文档
3. 若要复制 JSON，直接使用每个接口下的请求 / 响应示例

