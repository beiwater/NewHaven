# 素材清单

## 已规划素材

### 建筑

- grain_plot lv1 idle / working / ready / construction
- grain_plot lv2 idle / working / ready / construction
- grain_plot lv3 idle / working / ready / construction
- mill_house lv1 idle / working / ready / construction
- mill_house lv2 idle / working / ready / construction
- mill_house lv3 idle / working / ready / construction
- bakery_shop lv1 idle / working / ready / construction
- bakery_shop lv2 idle / working / ready / construction
- bakery_shop lv3 idle / working / ready / construction
- meal_kiosk lv1 idle / working / ready / construction
- meal_kiosk lv2 idle / working / ready / construction
- meal_kiosk lv3 idle / working / ready / construction

### 资源/中间产物

- wheat：小麦，生产节点产物
- flour：面粉，加工节点产物
- bread：面包，深加工节点产物
- meal：简餐，销售/交付节点产物

### 背景

- map_background_v1：地图主背景概念
- map_background_clean_plots_v1：更适合放置建筑的干净地块版本

### 头像

- avatar_advisor_v1：经营顾问
- avatar_market_broker_v1：市场顾问
- avatar_engineer_v1：生产工程师

### UI 图标

- icon_collect
- icon_upgrade
- icon_market_price
- icon_contract
- icon_research

## 已生成素材

### 产业链 lv1 idle 静态节点

| 节点 | 文件 | 用途 |
| --- | --- | --- |
| 生产 | `art-src/transparent/grain_plot_lv1_idle_trimmed.png` | 小麦/粮食产出 |
| 加工 | `art-src/transparent/mill_house_lv1_idle_trimmed.png` | 小麦加工为面粉 |
| 深加工 | `art-src/transparent/bakery_shop_lv1_idle_trimmed.png` | 面粉加工为面包 |
| 销售 | `art-src/transparent/meal_kiosk_lv1_idle_trimmed.png` | 面包/简餐售卖变现 |

### 资源/中间产物小图

最终小图统一为 256x256 透明 PNG，适合价格条、仓库、生产气泡和订单奖励展示。

| 资源 | 文件 | 用途 |
| --- | --- | --- |
| wheat | `art-src/items/item_wheat_v1.png` | 小麦原料 |
| flour | `art-src/items/item_flour_v1.png` | 面粉中间产物 |
| bread | `art-src/items/item_bread_v1.png` | 面包中间产物/商品 |
| meal | `art-src/items/item_meal_v1.png` | 简餐终端商品 |

原始 2x2 图标 sheet：`art-src/items/resource_items_chain_v1_sheet.png`

### 背景和头像

| 类型 | 文件 | 用途 |
| --- | --- | --- |
| 地图背景 | `art-src/backgrounds/map_background_v1.png` | 地图式主界面概念 |
| 顾问头像 | `art-src/avatars/avatar_advisor_v1_trimmed.png` | 新手引导/经营顾问 |

## 下一批建议

下一批建议不要急着生成全部建筑，而是先补状态动画：

- `grain_plot_lv1_working_strip_6f`
- `grain_plot_lv1_ready_strip_4f`
- `mill_house_lv1_working_strip_6f`
- `bakery_shop_lv1_working_strip_6f`
- `meal_kiosk_lv1_ready_strip_4f`

如果要先看升级差异，则生成：

- `grain_plot_lv2_idle`
- `grain_plot_lv3_idle`
- `mill_house_lv2_idle`
- `mill_house_lv3_idle`

## UI 细节图标

用户提供了一张 4x4 图标 sheet，已切分为 16 个 128x128 图标。

文档：`art-src/ui-icons/docs/ui-icon-slicing.md`

原图：`art-src/ui-icons/source/ui_icon_sheet_v1.png`

透明版目录：`art-src/ui-icons/transparent/`

包含图标：

- `icon_coin_v1.png`
- `icon_cash_v1.png`
- `icon_level_badge_v1.png`
- `icon_xp_v1.png`
- `icon_upgrade_v1.png`
- `icon_collect_v1.png`
- `icon_timer_v1.png`
- `icon_market_v1.png`
- `icon_contract_v1.png`
- `icon_warehouse_v1.png`
- `icon_energy_v1.png`
- `icon_builder_v1.png`
- `icon_wheat_resource_v1.png`
- `icon_factory_v1.png`
- `icon_restaurant_v1.png`
- `icon_refresh_v1.png`
