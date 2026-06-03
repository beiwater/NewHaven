# 第一条产业链设计

## 产业链

`小麦 -> 面粉 -> 面包 -> 简餐`

这条链条足够短，适合验证地图经营、建筑升级、生产等待、收取反馈、市场价格展示和订单交付。

## 建筑列表

| building_id | 名称 | 作用 | 推荐占地 |
| --- | --- | --- | --- |
| grain_plot | 粮食地块 | 产出小麦 | 1x1 |
| mill_house | 磨坊 | 小麦加工为面粉 | 2x1 |
| bakery_shop | 面包坊 | 面粉加工为面包 | 2x2 |
| meal_kiosk | 简餐摊/小餐馆 | 面包加工/售卖为简餐 | 2x2 |

## 等级差异

### lv1

低成本、手工感、木材、布棚、少量工具。
轮廓简单，适合新手阶段。

### lv2

砖墙、固定招牌、更好的设备、更干净的工作区。
体积和细节明显增加。

### lv3

自动化设备、金属结构、灯牌、管线、传送带或机械臂。
必须一眼看出比 lv1/lv2 更高级。

## 状态差异

| state | 视觉表现 | 用途 |
| --- | --- | --- |
| idle | 轻微待机，无强烈特效 | 可启动生产 |
| working | 烟、齿轮、灯光、传送带、蒸汽 | 正在生产 |
| ready | 金色光圈、跳动收取标识、产物气泡 | 可收取 |
| construction | 脚手架、施工布、锤子、半透明进度感 | 建造或升级中 |

## 第一批素材优先级

优先生成：

- `grain_plot_lv1_idle`
- `grain_plot_lv1_working`
- `grain_plot_lv1_ready`
- `mill_house_lv1_idle`
- `bakery_shop_lv1_idle`
- `meal_kiosk_lv1_idle`
- `map_background_v1`
- `advisor_avatar_v1`

后续再补 lv2/lv3 和完整动画帧。
