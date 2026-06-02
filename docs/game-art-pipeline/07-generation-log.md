# 素材生成日志

生成日期：2026-06-01

## 本批素材

### grain_plot_lv1_idle

用途：第一条产业链的起始建筑，表示小麦/粮食生产节点。

文件：

- 原始 chroma-key：`art-src/generated/grain_plot_lv1_idle_chromakey.png`
- 透明完整画布：`art-src/transparent/grain_plot_lv1_idle.png`
- 透明裁切版：`art-src/transparent/grain_plot_lv1_idle_trimmed.png`

质量记录：

- 抠图成功，边缘基本干净。
- 作为建筑素材非常漂亮，可读性强。
- 但对 `lv1` 来说略显豪华，后续批量生成 lv1 时应降低复杂度，让 lv2/lv3 有升级空间。

### map_background_v1

用途：地图式主界面背景概念图。

文件：

- `art-src/backgrounds/map_background_v1.png`

质量记录：

- 适合表达“地图样式经营界面”。
- 有大量可放置地块、道路和扩展区域。
- 后续如果要用于真实前端，需要再生成一版更干净的 `clean_plots`，减少固定建筑，方便叠加建筑精灵。

### avatar_advisor_v1

用途：新手引导/经营顾问头像。

文件：

- 原始 chroma-key：`art-src/generated/avatar_advisor_v1_chromakey.png`
- 透明完整画布：`art-src/avatars/avatar_advisor_v1.png`
- 透明裁切版：`art-src/avatars/avatar_advisor_v1_trimmed.png`

质量记录：

- 抠图成功，头发边缘有少量半透明像素，整体可用。
- 风格偏精致角色头像，可作为 UI 顾问。
- 后续如果游戏整体更偏欧美经营风，可以再生成一版更朴素、更“经营经理”的头像。

## 下一批建议

为了统一风格，下一批不要立刻生成全部动画，先补齐同等级建筑静态图：

- `mill_house_lv1_idle`
- `bakery_shop_lv1_idle`
- `meal_kiosk_lv1_idle`

然后做一次风格评审。
如果风格统一，再生成：

- lv2/lv3 静态升级版
- working spritesheet
- ready spritesheet
- construction spritesheet

## 透明处理记录

使用 chroma-key 抠图策略：

- `grain_plot` 使用 `#ff00ff`
- `avatar` 使用 `#00ff00`

处理后再按 alpha 边界裁切，并保留 32px 安全边距。

## 第二批补齐产业链节点

生成日期：2026-06-01

本批补齐了加工、深加工、销售节点，使第一条产业链不再只有生产节点。

### mill_house_lv1_idle

用途：加工节点，小麦/粮食 -> 面粉。

文件：

- 原始 chroma-key：`art-src/generated/mill_house_lv1_idle_chromakey.png`
- 透明完整画布：`art-src/transparent/mill_house_lv1_idle.png`
- 透明裁切版：`art-src/transparent/mill_house_lv1_idle_trimmed.png`

质量记录：

- 抠图成功。
- 形象明确，是磨坊/加工节点。
- 比较适合作为地图上的 2x1 或 2x2 建筑。

### bakery_shop_lv1_idle

用途：深加工节点，面粉 -> 面包。

文件：

- 原始 chroma-key：`art-src/generated/bakery_shop_lv1_idle_chromakey.png`
- 透明完整画布：`art-src/transparent/bakery_shop_lv1_idle.png`
- 透明裁切版：`art-src/transparent/bakery_shop_lv1_idle_trimmed.png`

质量记录：

- 抠图成功。
- 烤炉、面包、柜台都清楚。
- 略像成熟店铺，后续如果严格区分 lv1/lv2，可以让 lv1 更简陋。

### meal_kiosk_lv1_idle

用途：销售终端，面包/简餐 -> 现金收入。

文件：

- 原始 chroma-key：`art-src/generated/meal_kiosk_lv1_idle_chromakey.png`
- 透明完整画布：`art-src/transparent/meal_kiosk_lv1_idle.png`
- 透明裁切版：`art-src/transparent/meal_kiosk_lv1_idle_trimmed.png`

质量记录：

- 抠图成功。
- 销售/终端节点语义明确。
- 有招牌和菜单板但无可读文字，可用；正式版可减少招牌复杂度。

## 第三批补齐资源/中间产物图标

生成日期：2026-06-01

本批补齐产业链中间产物，让玩家能看到建筑之间流动的货物。

### resource_items_chain_v1

用途：价格条、仓库、生产气泡、订单奖励、地图飘字。

文件：

- 原始 chroma-key：`art-src/generated/resource_items_chain_v1_chromakey.png`
- 透明 sheet：`art-src/items/resource_items_chain_v1_sheet.png`
- 小麦图标：`art-src/items/item_wheat_v1.png`
- 面粉图标：`art-src/items/item_flour_v1.png`
- 面包图标：`art-src/items/item_bread_v1.png`
- 简餐图标：`art-src/items/item_meal_v1.png`

质量记录：

- 四个图标均已抠透明。
- 每个最终图标统一压到 256x256 透明 PNG。
- 风格与建筑素材一致，适合 UI 和地图气泡。
- 面粉、面包、小麦识别度很好；简餐图标偏三明治/便当，但能表达终端商品。
