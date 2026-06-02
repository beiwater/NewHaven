# UI 图标 Sheet 切分说明

来源文件：`art-src/ui-icons/source/ui_icon_sheet_v1.png`

用户原始文件：`E:/Downloads/ChatGPT Image 2026年6月1日 18_26_29.png`

## 切分方式

原图尺寸为 `1254x1254`，内容为标准 `4x4` 图标 sheet。

切分规则：

- 从左到右、从上到下编号。
- 每行 4 个图标，共 4 行。
- 每个 cell 约为 `313.5x313.5`。
- 使用按比例网格切分，避免因为 1254 不能被 4 整除导致像素偏移。
- 每个图标输出两类文件：
  - `slices/*_cell.png`：保留紫色背景的 128x128 原始切片。
  - `transparent/*.png`：自动去紫底后的 128x128 透明版。

注意：原图背景是紫色渐变，不是纯色 chroma-key，因此透明版属于自动去底结果，需要人工抽检边缘。正式使用前，如果发现紫边，需要重新生成纯色背景版或手工修边。

## 图标语义映射

| 序号 | 行列 | 文件 ID | 中文语义 | 用途 |
| --- | --- | --- | --- | --- |
| 01 | R1C1 | `icon_coin_v1` | 金币/硬币 | 现金、货币、奖励 |
| 02 | R1C2 | `icon_cash_v1` | 现金钞票 | 大额资金、收入 |
| 03 | R1C3 | `icon_level_badge_v1` | 等级徽章 | 公司等级、建筑等级 |
| 04 | R1C4 | `icon_xp_v1` | 经验 XP | 玩家经验、升级进度 |
| 05 | R2C1 | `icon_upgrade_v1` | 升级箭头 | 建筑升级、科技升级 |
| 06 | R2C2 | `icon_collect_v1` | 收取/收菜 | 领取产物、收获奖励 |
| 07 | R2C3 | `icon_timer_v1` | 计时器 | 生产倒计时、冷却 |
| 08 | R2C4 | `icon_market_v1` | 市场摊位 | 市场、价格、交易 |
| 09 | R3C1 | `icon_contract_v1` | 合同订单 | 订单、合约、政府合同 |
| 10 | R3C2 | `icon_warehouse_v1` | 仓库箱子 | 仓库、库存、物流 |
| 11 | R3C3 | `icon_energy_v1` | 能量闪电 | 体力、加速、boost |
| 12 | R3C4 | `icon_builder_v1` | 安全帽/施工 | 建造、施工、升级中 |
| 13 | R4C1 | `icon_wheat_resource_v1` | 小麦资源 | 小麦资源、小麦价格 |
| 14 | R4C2 | `icon_factory_v1` | 工厂加工 | 加工、生产、制造 |
| 15 | R4C3 | `icon_restaurant_v1` | 餐厅销售 | 餐饮、销售终端 |
| 16 | R4C4 | `icon_refresh_v1` | 刷新循环 | 刷新、循环、重置 |

## 输出目录

- 原 sheet：`art-src/ui-icons/source/ui_icon_sheet_v1.png`
- 保留紫底切片：`art-src/ui-icons/slices/`
- 透明 128 图标：`art-src/ui-icons/transparent/`
- JSON manifest：`art-src/ui-icons/docs/ui_icon_sheet_v1_manifest.json`

## 使用建议

优先在 UI 中使用 `transparent/*.png`。
如果发现透明版边缘有紫色残留，先使用 `slices/*_cell.png` 做占位，后续重新生成纯色背景或真正透明版。

这些图标适合：

- 顶部资源栏
- 底部价格条
- 建筑详情面板
- 生产气泡
- 订单奖励
- 新手引导
- 地图状态标记
