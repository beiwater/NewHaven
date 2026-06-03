下面按原图 **1672 × 941 px** 来描述。尺寸是根据截图估算的，适合你后面做 UI 复刻、切图、前端布局参考，误差大概在 **±5～15 px**。

# 1. 整体布局结构

这是一张横屏经营模拟游戏 UI，整体可以分成五大层：

```txt
Root Canvas 1672×941
├── Top Header 顶部状态栏 1672×66
├── Main Body 主体区域 1672×774
│   ├── Left Sidebar 左侧导航栏 142×774
│   ├── Production Map 中央地图 1205×774
│   └── Right Detail Panel 右侧建筑详情面板 322×735
└── Bottom Price Bar 底部价格栏 1672×101
```

视觉上是经典的经营游戏布局：
**顶部管玩家状态，左侧管功能菜单，中间管地图操作，右侧管当前建筑详情，底部管市场行情。**

---

# 2. 顶部 Header 区域

顶部栏高度约 **66 px**，宽度占满全屏 **1672 px**。

```txt
Top Header
位置：x=0, y=0
大小：1672×66
```

内部从左到右分成几个信息块。

## 2.1 公司 Logo 与公司名区域

```txt
Company Block
位置：x=0, y=0
大小：约 373×66
```

里面嵌套：

```txt
Company Block
├── Logo Badge 约 105×66
│   └── 圆形农田图标 约 54×54
└── Company Text 约 260×60
    ├── Mellow Acres Co.
    └── 小字空白/副标题区域
```

左上角是一个圆形农场 Logo，背景是米黄色徽章。
右侧文字是 **Mellow Acres Co.**，字体较大，黑褐色，像乡村农场品牌名。

---

## 2.2 现金区域

```txt
Cash Block
位置：x=374, y=0
大小：约 326×66
```

内部：

```txt
Cash Block
├── 现金图标 约 42×36
├── Cash 文本
└── $248,920 数值
```

显示：

```txt
Cash $248,920
```

现金图标是绿色钞票堆，整体偏偏左。
这个区块背景是浅米色，与顶部栏其他区域用竖线分隔。

---

## 2.3 等级与经验条区域

```txt
Level Block
位置：x=700, y=0
大小：约 392×66
```

内部：

```txt
Level Block
├── 星星等级图标 约 42×42
├── Level 42 文本
├── XP 数字 18,560 / 25,000 XP
└── 经验进度条 约 220×9
```

显示：

```txt
Level 42
18,560 / 25,000 XP
```

经验条是横向蓝绿色条，背景是灰褐色槽。
星星图标在左侧，像等级徽章。

---

## 2.4 体力区域

```txt
Energy Block
位置：x=1092, y=0
大小：约 158×66
```

内部：

```txt
Energy Block
├── 闪电图标 约 26×34
├── 120 / 120
└── 绿色加号按钮 约 26×26
```

显示：

```txt
120 / 120
```

这是体力/能量系统。

---

## 2.5 人口/员工区域

```txt
Worker Block
位置：x=1250, y=0
大小：约 120×66
```

内部：

```txt
Worker Block
├── 双人头像图标 约 27×24
└── 8 / 10
```

代表当前员工、队列、人口或可用工人数量。

---

## 2.6 右上角功能按钮

```txt
Top Icons Block
位置：x=1490, y=0
大小：约 180×66
```

内部：

```txt
Top Icons Block
├── 设置齿轮 约 34×34
└── 通知铃铛 约 28×34
```

右上角有齿轮和铃铛。
图标比较小，留白很多，适合点击。

---

# 3. 左侧导航栏

左侧导航栏位于顶部栏下方，底部价格栏上方。

```txt
Left Sidebar
位置：x=0, y=66
大小：约 142×774
```

背景是浅米黄色面板，边框深棕色。
左栏里面是纵向菜单。

## 3.1 菜单按钮结构

每个菜单按钮大约：

```txt
Menu Item
大小：约 142×80
```

选中的 Map 按钮高度约 **68 px**，背景偏黄色。
其他按钮为浅米色背景，之间有细分隔线。

菜单层级：

```txt
Left Sidebar
├── Map Button 142×68
├── Build Button 142×80
├── Warehouse Button 142×80
├── Market Button 142×80
├── Contracts Button 142×80
├── Research Button 142×80
└── Company Reputation Card 约 116×95
```

---

## 3.2 Map 按钮

```txt
Map Button
位置：x=0, y=126
大小：约 142×68
```

内部：

```txt
Map Button
├── 地图图标 约 45×45
└── Map 文本
```

这是当前选中的页面，背景更亮。

---

## 3.3 Build 按钮

```txt
Build Button
位置：x=0, y=194
大小：约 142×80
```

图标是安全帽，文字是：

```txt
Build
```

---

## 3.4 Warehouse 按钮

```txt
Warehouse Button
位置：x=0, y=274
大小：约 142×80
```

图标是绿色仓库箱，文字是：

```txt
Warehouse
```

---

## 3.5 Market 按钮

```txt
Market Button
位置：x=0, y=354
大小：约 142×80
```

图标是购物车，文字是：

```txt
Market
```

---

## 3.6 Contracts 按钮

```txt
Contracts Button
位置：x=0, y=434
大小：约 142×80
```

图标是合同板夹，文字是：

```txt
Contracts
```

---

## 3.7 Research 按钮

```txt
Research Button
位置：x=0, y=514
大小：约 142×80
```

图标是绿色烧瓶，文字是：

```txt
Research
```

---

## 3.8 公司声望卡片

```txt
Company Reputation Card
位置：x=13, y=694
大小：约 116×92
```

内部：

```txt
Company Reputation Card
├── 标题 Company Reputation
├── 徽章图标
├── 等级盾牌 5
├── Trusted Partner 文本
└── 进度条
```

显示：

```txt
COMPANY REPUTATION
Trusted Partner
5
```

这是一个小型状态卡，深色背景，放在左下角。

---

# 4. 中央 Production Map 地图区域

中央地图是整张图的核心。

```txt
Production Map
位置：x=142, y=66
大小：约 1205×774
```

它是一个俯视角城市/农场工业园地图。底层是背景图，上面叠加建筑、标签、进度圈、按钮、可建地块等 UI。

层级大概是：

```txt
Production Map
├── 背景地图图像
│   ├── 河流
│   ├── 道路
│   ├── 树林
│   ├── 草地
│   └── 工业园区
├── 顶部标题牌 Production Map
├── 建筑节点 Layer
│   ├── 建筑插画
│   ├── 建筑名标签
│   ├── 时间标签
│   ├── 进度圆环
│   └── Collect / 状态按钮
├── 可建造地块 Layer
└── 扩展区域 Layer
```

---

## 4.1 Production Map 标题牌

```txt
Production Map Banner
位置：x=172, y=88
大小：约 318×50
```

文字：

```txt
Production Map
```

这是一个浅色横幅牌，有折角和阴影。
它悬浮在地图左上方，不占地图交互空间。

---

# 5. 地图上的建筑节点

每个建筑节点结构基本一致：

```txt
Building Node
├── Building Illustration 建筑插画
├── Name Tag 名称标签
│   ├── 小图标
│   ├── 建筑名
│   └── 时间/状态条
├── Circular Progress Ring 底部圆形进度环
└── Action Button 可选按钮
```

建筑节点不是规则网格，而是分布在地图道路之间。

---

## 5.1 Research Lab

```txt
Research Lab Node
位置：x=247, y=160
整体大小：约 130×110
```

标签：

```txt
Research Lab
2h 15
```

嵌套：

```txt
Research Lab
├── 实验室建筑 约 80×70
├── 蓝色圆形进度环 约 58×18
└── 标签 约 95×42
```

位于左上方靠河边。建筑是蓝色圆柱形实验室，有烧瓶图标。

---

## 5.2 Contract Office

```txt
Contract Office Node
位置：x=600, y=115
整体大小：约 130×115
```

标签：

```txt
Contract Office
1h 40m
```

结构：

```txt
Contract Office
├── 办公楼 约 105×85
├── 蓝色进度环 约 65×20
└── 标签 约 125×43
```

建筑是灰色办公楼，带烟囱和停车区。

---

## 5.3 Market Terminal

```txt
Market Terminal Node
位置：x=845, y=105
整体大小：约 125×120
```

标签：

```txt
Market Terminal
45m 10s
```

这是一个市场/交易终端建筑，屋顶有黑绿色显示屏。
底部有绿色进度环。

```txt
Market Terminal
├── 建筑 约 100×80
├── 绿色进度环 约 65×20
└── 标签 约 125×43
```

---

## 5.4 Corporate HQ

```txt
Corporate HQ Node
位置：x=1130, y=95
整体大小：约 135×130
```

标签：

```txt
Corporate HQ
MAX
```

嵌套：

```txt
Corporate HQ
├── 大型总部建筑 约 125×100
├── 水池/广场装饰
└── 标签 约 120×45
```

这是右上角的大楼，显示已经满级。
`MAX` 是绿色小标签。

---

## 5.5 Supplier Hub

```txt
Supplier Hub Node
位置：x=385, y=250
整体大小：约 150×120
```

标签：

```txt
Supplier Hub
3h 20m
```

建筑是一排货车、仓库和卸货区。
标签在建筑上方偏左。

```txt
Supplier Hub
├── 货运建筑/车辆 约 140×80
├── 标签 约 125×43
└── 蓝色小进度状态
```

---

## 5.6 Bakery 当前选中建筑

```txt
Bakery Node
位置：x=620, y=235
整体大小：约 145×145
```

这是当前被选中的建筑，因为它底部有明显黄色光圈。

标签：

```txt
Bakery
25m 30s
Collect
```

嵌套：

```txt
Bakery Node
├── 黄色选中光圈 约 105×36
├── Bakery 建筑插画 约 110×85
├── 建筑标签 约 120×74
│   ├── 面包图标
│   ├── Bakery
│   ├── 25m 30s
│   └── Collect 按钮
```

这个建筑位于地图中央偏上，右侧详情面板也显示它的信息，说明用户当前选中了 Bakery。

---

## 5.7 Meal Kitchen

```txt
Meal Kitchen Node
位置：x=850, y=277
整体大小：约 130×120
```

标签：

```txt
Meal Kitchen
1h 05m
```

建筑像小餐饮加工厨房，招牌写着 MEAL。
底部有蓝色进度环。

```txt
Meal Kitchen
├── 建筑 约 110×75
├── 进度环 约 65×20
└── 标签 约 115×43
```

---

## 5.8 Restaurant

```txt
Restaurant Node
位置：x=1080, y=355
整体大小：约 145×130
```

标签：

```txt
Restaurant
35m 50s
Collect
```

建筑是橙红色餐厅，有遮阳棚。
下方有蓝色环形进度条。

```txt
Restaurant
├── 餐厅建筑 约 125×85
├── 蓝色进度环 约 75×20
└── 标签卡 约 115×74
    ├── 餐盘图标
    ├── Restaurant
    ├── 时间
    └── Collect 按钮
```

---

## 5.9 Grain Mill

```txt
Grain Mill Node
位置：x=285, y=400
整体大小：约 150×125
```

标签：

```txt
Grain Mill
15m 0s
Collect
```

建筑是谷物磨坊/加工厂，有圆筒仓。
底部有黄色光圈。

```txt
Grain Mill
├── 建筑 约 135×85
├── 黄色进度环 约 100×30
└── 标签卡 约 120×74
```

---

## 5.10 Dairy Plant

```txt
Dairy Plant Node
位置：x=610, y=430
整体大小：约 145×120
```

标签：

```txt
Dairy Plant
50m 00s
```

建筑是乳制品工厂，有储罐和蓝色水滴/奶瓶图标。
底部是蓝色进度环。

```txt
Dairy Plant
├── 工厂建筑 约 130×85
├── 蓝色进度环 约 75×20
└── 标签 约 115×43
```

---

## 5.11 Meat Processing

```txt
Meat Processing Node
位置：x=875, y=495
整体大小：约 150×120
```

标签：

```txt
Meat Processing
1h 10m
```

建筑是肉类加工厂，标签左侧有红色肉块图标。
底部蓝色进度环。

```txt
Meat Processing
├── 建筑 约 135×85
├── 进度环 约 70×20
└── 标签 约 135×43
```

---

## 5.12 Logistics Depot

```txt
Logistics Depot Node
位置：x=1120, y=520
整体大小：约 160×120
```

标签：

```txt
Logistics Depot
40m 15s
```

建筑是物流仓库，周围有卡车和货物。
底部有蓝色进度环。

```txt
Logistics Depot
├── 仓库建筑 约 145×90
├── 蓝色进度环 约 75×20
└── 标签 约 135×43
```

---

## 5.13 Packaging Plant

```txt
Packaging Plant Node
位置：x=755, y=625
整体大小：约 145×125
```

标签：

```txt
Packaging Plant
20m 25s
Collect
```

建筑是包装工厂，有烟囱和仓库。
底部是黄色进度环。

```txt
Packaging Plant
├── 工厂建筑 约 125×85
├── 黄色进度环 约 95×25
└── 标签卡 约 125×74
```

---

## 5.14 Cold Storage

```txt
Cold Storage Node
位置：x=970, y=655
整体大小：约 150×125
```

标签：

```txt
Cold Storage
1h 30m
```

图标是雪花，建筑像冷库。
底部是蓝色进度环。

```txt
Cold Storage
├── 冷库建筑 约 135×90
├── 蓝色进度环 约 80×22
└── 标签 约 120×43
```

---

# 6. 地图上的空地与扩展区

## 6.1 右上 Available Plot

```txt
Available Plot 1
位置：x=1218, y=330
大小：约 125×100
```

虚线白色边框，内部文字：

```txt
Available Plot
Tap to Build
```

结构：

```txt
Available Plot
├── 白色虚线多边形边框
└── 中心说明文字
```

表示可建造地块。

---

## 6.2 右下 Available Plot

```txt
Available Plot 2
位置：x=1220, y=685
大小：约 120×120
```

同样是虚线白色地块，内部文字：

```txt
Available Plot
Tap to Build
```

---

## 6.3 左下 Expansion Zone

```txt
Expansion Zone
位置：x=255, y=625
大小：约 250×135
```

这是一块更大的虚线多边形区域。

内部：

```txt
Expansion Zone
├── 加号圆形按钮 约 34×34
├── Expansion Zone 文本
└── Unlock at Level 45 锁定提示
```

显示：

```txt
Expansion Zone
Unlock at Level 45
```

它不是普通建筑位，而是地图扩建区域。

---

# 7. 右侧 Bakery 详情面板

右侧详情面板是当前选中建筑的信息窗口。

```txt
Right Detail Panel
位置：x=1348, y=76
大小：约 315×734
```

背景是浅米色羊皮纸风格，深色描边，圆角约 **8～12 px**。

层级：

```txt
Bakery Detail Panel
├── Header
│   ├── Bakery 标题
│   ├── Level 6
│   └── Close Button
├── Building Preview
├── Produces Section
├── Production Progress Section
├── Output Section
├── Main Action Buttons
├── Upgrade Section
└── Active Jobs Section
```

---

## 7.1 面板 Header

```txt
Panel Header
位置：x=1350, y=78
大小：约 312×75
```

内部：

```txt
Header
├── Bakery 标题
├── Level 6
└── 关闭 X 按钮
```

标题：

```txt
Bakery
Level 6
```

右上角关闭按钮：

```txt
Close Button
位置：x=1624, y=84
大小：约 31×31
```

---

## 7.2 建筑预览图

```txt
Building Preview
位置：x=1412, y=145
大小：约 125×95
```

显示一个小型 3D Bakery 建筑，有黄色底座阴影。
预览图居中放在标题下面。

---

## 7.3 Produces 区域

```txt
Produces Section
位置：x=1366, y=254
大小：约 270×37
```

显示：

```txt
Produces Bread
```

其中 **Bread** 是红褐色强调文字，旁边有面包图标。

嵌套：

```txt
Produces Section
├── Produces 文本
├── Bread 文字
└── 面包图标 约 28×18
```

---

## 7.4 Production Progress 区域

```txt
Progress Section
位置：x=1366, y=300
大小：约 280×70
```

包含：

```txt
Production Progress
25m 30s / 1h 00m
横向进度条
面包图标
```

进度条大小约：

```txt
Progress Bar
位置：x=1368, y=356
大小：约 226×15
```

绿色进度填充大概占 **42%**。
右侧有一个面包图标，大小约 **34×23**。

---

## 7.5 Output 区域

```txt
Output Section
位置：x=1365, y=385
大小：约 280×55
```

显示：

```txt
Output
320 / 720
```

左边有面包图标。
右侧有一个箱子图标按钮。

```txt
Output Section
├── 面包图标 约 36×24
├── 320 / 720 数值
└── 箱子按钮 约 38×38
```

箱子按钮可能表示库存、产物详情或产出容量。

---

## 7.6 Collect / Start 按钮区

```txt
Action Buttons Row
位置：x=1372, y=430
大小：约 267×50
```

内部两个按钮：

```txt
Collect Button
位置：x=1372, y=430
大小：约 128×47

Start Button
位置：x=1510, y=430
大小：约 128×47
```

Collect 是绿色按钮。
Start 是蓝绿色按钮。
两个按钮宽度接近，中间间隔约 **10 px**。

---

## 7.7 Upgrade 区域

```txt
Upgrade Section
位置：x=1365, y=498
大小：约 280×60
```

内部：

```txt
Upgrade Section
├── 绿色上箭头按钮 约 36×36
├── Upgrade 标题
├── Increase output capacity. 说明
└── $36,000 按钮
```

价格按钮：

```txt
Upgrade Price Button
位置：x=1555, y=506
大小：约 82×34
```

显示：

```txt
$36,000
```

---

## 7.8 Active Jobs 区域

```txt
Active Jobs Section
位置：x=1365, y=575
大小：约 280×170
```

标题行：

```txt
Active Jobs      2 / 2
```

说明当前队列满了。

里面两个任务：

### Bread Production

```txt
Job Row 1
位置：x=1370, y=620
大小：约 265×40
```

嵌套：

```txt
Bread Production Row
├── 面包图标 约 32×22
├── Bread Production 文本
├── 时间 1h 00m
├── 进度条 约 160×8
└── 红色取消按钮 约 27×27
```

---

### Flour Production

```txt
Job Row 2
位置：x=1370, y=668
大小：约 265×40
```

嵌套：

```txt
Flour Production Row
├── 麦穗图标 约 32×32
├── Flour Production 文本
├── 时间 45m 00s
├── 进度条 约 160×8
└── 红色取消按钮 约 27×27
```

---

## 7.9 View All Jobs 按钮

```txt
View All Jobs Button
位置：x=1428, y=716
大小：约 136×38
```

显示：

```txt
View All Jobs
```

是浅棕色按钮，位于 Active Jobs 底部居中。

---

# 8. 底部 Latest Prices 市场价格栏

底部价格栏横跨全屏。

```txt
Bottom Price Bar
位置：x=0, y=839
大小：1672×102
```

层级：

```txt
Bottom Price Bar
├── Latest Prices 标题区
├── 商品价格卡片列表
│   ├── Wheat
│   ├── Vegetable
│   ├── Meat
│   ├── Bread
│   └── Meal
└── View Market 按钮
```

背景是浅米色，与整体纸质 UI 风格一致。

---

## 8.1 Latest Prices 标题区

```txt
Latest Prices Block
位置：x=0, y=839
大小：约 220×102
```

内部：

```txt
Latest Prices
├── 上涨图表图标 约 45×45
└── Latest Prices 文本
```

文字较大，位于左下方。

---

## 8.2 Wheat 价格卡

```txt
Wheat Card
位置：x=220, y=839
大小：约 235×102
```

内部：

```txt
Wheat Card
├── 麦穗图标 约 50×65
├── Wheat 名称
├── $4.20 价格
├── +2.1% 涨幅
└── 绿色迷你折线图
```

显示：

```txt
Wheat
$4.20
+2.1%
```

---

## 8.3 Vegetable 价格卡

```txt
Vegetable Card
位置：x=455, y=839
大小：约 230×102
```

内部：

```txt
Vegetable
$5.10
-0.8%
```

蔬菜图标是绿色包菜。
涨跌幅为红色，右侧有红色迷你折线图。

---

## 8.4 Meat 价格卡

```txt
Meat Card
位置：x=685, y=839
大小：约 250×102
```

显示：

```txt
Meat
$9.80
+4.3%
```

图标是肉排。
右侧绿色小折线图。

---

## 8.5 Bread 价格卡

```txt
Bread Card
位置：x=935, y=839
大小：约 245×102
```

显示：

```txt
Bread
$11.50
+1.2%
```

图标是面包。
右侧绿色折线图。

---

## 8.6 Meal 价格卡

```txt
Meal Card
位置：x=1180, y=839
大小：约 330×102
```

显示：

```txt
Meal
$28.40
-1.5%
```

图标是一盘餐食。
右侧红色折线图。

---

## 8.7 View Market 按钮

```txt
View Market Button
位置：x=1522, y=865
大小：约 128×47
```

显示：

```txt
View Market >
```

按钮是浅米色描边风格，右侧有箭头。

---

# 9. UI 风格总结

这张图的设计风格是：

```txt
经营模拟游戏
农场 + 工厂 + 交易市场
俯视角地图
纸质面板 UI
暖色系木质/羊皮纸边框
半写实卡通建筑
大量小型状态标签
```

核心视觉层级非常清楚：

```txt
第一层：背景地图
第二层：建筑插画
第三层：建筑状态标签
第四层：当前选中建筑详情面板
第五层：顶部/左侧/底部固定 UI
```

从玩法上看，它表达的是：

```txt
玩家在生产地图中管理建筑
每个建筑有生产时间、产出和队列
选中建筑后右侧显示详细生产信息
底部显示市场价格变化
左侧切换系统模块
顶部显示公司状态和资源
```

最重要的嵌套逻辑是：

```txt
Root
├── 固定 HUD：顶部资源、左侧菜单、底部行情
├── 主交互区：Production Map
│   ├── 地图背景
│   ├── 建筑节点
│   ├── 建筑状态标签
│   └── 可扩展地块
└── 当前选中对象详情：Bakery Panel
    ├── 建筑基础信息
    ├── 生产进度
    ├── 当前产出
    ├── 操作按钮
    ├── 升级入口
    └── 生产队列
```

这套布局非常适合你做网页经营游戏：
**主地图负责沉浸感，右侧面板负责精确操作，底部价格栏负责经济系统，左侧菜单负责功能入口。**
