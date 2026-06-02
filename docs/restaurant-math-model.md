# 餐厅数学模型 v0

版本：`20260601T000132Z-e32490`

这份文档定义第一版餐厅经营的核心公式。目标不是复刻其他游戏，而是借鉴“菜单、员工、风格、评级、价格、12 小时周期”这些经营变量，做成适合格子农场 MVP 的可解释模型。

## 1. 核心周期

餐厅按固定周期结算：

```txt
cycleHours = 12
```

一个周期开始前，玩家可以配置：

```txt
menuConfig
menuPrice
staffAbility
restaurantStyle
```

周期结算时，后端计算：

```txt
economicCycle
restaurantRating
cycleCapacity
occupancyRate
soldCustomers
grossRevenue
operatingCost
netIncome
reputationChange
```

前端只展示预测和提交配置，不可信任前端提交的收益、库存消耗、售出人数或评级。

## 2. 菜单配置

第一版菜单有三类：

```txt
starter = 前菜
main    = 主菜
drink   = 饮品
```

开店条件：

```txt
每类至少选择 1 道菜
```

每道菜定义：

```ts
menuItem = {
  id,
  category,
  inputs,
  quality,
  ratingBonus
}
```

变量：

```txt
Q_i = 第 i 道菜品质，范围 0-100
B_i = 第 i 道菜评级加成
N   = 菜单上菜品数量
C   = 菜品类别数量，第一版最大为 3
```

平均菜单品质：

```txt
avgMenuQuality = sum(Q_i) / N
```

菜单评级加成：

```txt
menuBonus = sum(B_i)
```

多样性加成：

```txt
varietyBonus = max(0, N - 3) * 2
```

解释：

三类基础菜单没有额外多样性奖励。每多一道菜，餐厅评级小幅增加，但不会指数膨胀。

## 3. 员工能力

员工不是单纯美术配置，会影响评分和成本。

第一版员工类型：

```txt
trainee      = 培训生
professional = 专业员工
```

变量：

```txt
staffAbility      = 员工能力，范围 0-100
staffRatingBonus  = 员工带来的评级加成
staffWageCost     = 12 小时周期工资成本
```

建议第一版参数：

```txt
trainee:
  staffAbility = 35
  staffRatingBonus = 0
  staffWageCost = 30

professional:
  staffAbility = 80
  staffRatingBonus = 10
  staffWageCost = 150
```

后续可以把员工能力拆成：

```txt
serviceSkill
cookingSkill
managementSkill
```

但第一版先合并为一个 `staffAbility`，避免系统过早复杂。

## 4. 餐厅风格

餐厅风格决定座位数、运营成本和评级倾向。

第一版风格：

```txt
fast     = 快餐式
balanced = 家庭餐厅
luxury   = 豪华餐厅
```

变量：

```txt
styleSeats            = 周期最大接待人数
styleOperatingCost    = 12 小时固定运营成本
styleRatingBonus      = 风格评分加成
styleRatingMultiplier = 风格评分倍率
```

建议第一版参数：

```txt
fast:
  styleSeats = 24
  styleOperatingCost = 45
  styleRatingBonus = -6
  styleRatingMultiplier = 0.90

balanced:
  styleSeats = 18
  styleOperatingCost = 70
  styleRatingBonus = 0
  styleRatingMultiplier = 1.00

luxury:
  styleSeats = 10
  styleOperatingCost = 110
  styleRatingBonus = 14
  styleRatingMultiplier = 1.20
```

解释：

快餐式座位多、评级低、成本低；豪华式座位少、评级高、成本高。这样玩家可以走“低价高客流”或“高价高评分”的路线。

## 5. 餐厅评级

餐厅评级用于影响上座率。

变量：

```txt
restaurantRating = 餐厅评级，范围 0-100
restaurantLevel  = 餐厅等级
reputation       = 餐厅声誉，长期缓慢变化
```

公式：

```txt
rawRating =
  24
  + avgMenuQuality * 0.32
  + menuBonus
  + varietyBonus
  + staffRatingBonus
  + styleRatingBonus
  + reputation * 0.25
  + restaurantLevel * 2
```

最终评级：

```txt
restaurantRating =
  clamp(rawRating * styleRatingMultiplier, 0, 100)
```

设计原则：

```txt
菜品品质影响基础评级
菜单多样性提供小幅奖励
专业员工提高评级但增加成本
豪华风格提高评级但减少座位
声誉慢慢积累，不应短期暴涨
```

## 6. 菜单价格

菜单价格代表每个顾客支付的价格。

第一版价格范围：

```txt
60 <= menuPrice <= 350
```

变量：

```txt
recommendedPrice = 系统推荐价格
menuPrice        = 玩家设置价格
```

价格因子：

```txt
priceFactor =
  clamp(
    1 + (recommendedPrice / menuPrice - 1) * priceSensitivity,
    0.30,
    1.35
  )
```

解释：

价格越高，上座率越低；价格低于推荐价时能提高上座率，但上限被限制，避免无脑低价。

## 7. 周期容量与食物消耗

餐厅不是卖出多少才消耗多少。为了让菜单配置有成本，第一版采用“开一个周期即准备本周期食物”的规则。

每类菜品可支持的份数：

```txt
categoryCapacity(category) =
  sum(maxServings(menuItem in category))
```

单道菜可支持份数：

```txt
maxServings(menuItem) =
  min(floor(inventory[resource] / requiredAmount) for each input)
```

周期最大供给：

```txt
foodCapacity =
  min(
    categoryCapacity(starter),
    categoryCapacity(main),
    categoryCapacity(drink)
  )
```

周期接待容量：

```txt
cycleCapacity =
  min(styleSeats, foodCapacity)
```

食物消耗规则：

```txt
每个类别按 cycleCapacity 分摊到该类别已选择菜品
每道菜消耗 ceil(cycleCapacity / categoryItemCount) 份对应原料
```

解释：

如果一个类别选择了更多菜，该类别每道菜消耗量会降低，但总类别消耗仍服务于本周期容量。

## 8. 经济周期与上座率

### 8.1 七天经济周期

游戏存在全局经济周期，每 7 天切换一次：

```txt
cycleDays = 7
```

第一版周期：

```txt
steady        = 平稳期
boom          = 繁荣期
recession     = 衰退期
food_festival = 食品热潮
frugal_week   = 节俭周
```

每个周期提供：

```txt
demandMultiplier  = 需求倍率
priceSensitivity  = 价格敏感度
ratingSensitivity = 评分敏感度
volatility        = 周期波动幅度
```

当前经济周期必须显示在顶部栏。

### 8.2 波动因子

为了让餐厅盈利不是机械固定，每次结算加入可复现的周期波动因子：

```txt
cycleWave =
  sin(settlementIndex * 1.73 + reputation * 0.13)

volatilityFactor =
  clamp(1 + cycleWave * economicCycle.volatility, 0.85, 1.15)
```

这里不使用纯随机数，避免刷新或重复结算导致结果不可追踪。

### 8.3 上座率

变量：

```txt
baseDemand        = 基础需求
priceFactor       = 价格因子
ratingFactor      = 评级因子
reputationFactor  = 声誉因子
volatilityFactor  = 周期波动因子
```

第一版：

```txt
baseDemand = 0.60
```

评级因子：

```txt
ratingFactor =
  clamp(restaurantRating / 70 * ratingSensitivity, 0.25, 1.45)
```

声誉因子：

```txt
reputationFactor =
  1 + reputation / 150
```

上座率：

```txt
occupancyRate =
  clamp(
    baseDemand
    * demandMultiplier
    * priceFactor
    * ratingFactor
    * reputationFactor
    * volatilityFactor,
    0,
    1
  )
```

## 9. 售出人数与收入

售出顾客数：

```txt
soldCustomers =
  floor(cycleCapacity * occupancyRate)
```

毛收入：

```txt
grossRevenue =
  soldCustomers * menuPrice
```

运营成本：

```txt
operatingCost =
  staffWageCost + styleOperatingCost
```

净收入：

```txt
netIncome =
  max(0, grossRevenue - operatingCost)
```

第一版净收入不允许为负，避免新手一轮结算直接破产。后续真实验可改成允许亏损。

## 10. 声誉变化

声誉是慢变量，不直接由玩家输入。

第一版：

```txt
if occupancyRate >= 0.72:
  reputation += 2
else if occupancyRate < 0.35:
  reputation -= 1
else:
  reputation += 1
```

限制：

```txt
reputation = clamp(reputation, 0, 100)
```

后续可以加入：

```txt
关店惩罚
连续高评分奖励
连续缺货惩罚
顾客类型分层
市场竞争压力
```

## 11. 后端结算伪代码

```ts
validate(menu has starter/main/drink)
validate(menuPrice between 60 and 350)

rating = calculateRestaurantRating(menu, staff, style, reputation, level)
cycle = calculateRestaurantCycleCapacity(inventory, menu, style)

if cycle.capacity <= 0:
  reject("Not enough food")

economicCycle = getCurrentEconomicCycle(now)
volatility = calculateCycleVolatility(settlementIndex, reputation, economicCycle.volatility)
occupancy = calculateOccupancyRate(menuPrice, recommendedPrice, rating, reputation, economicCycle, volatility)
soldCustomers = floor(cycle.capacity * occupancy)
grossRevenue = soldCustomers * menuPrice
operatingCost = staff.wageCost + style.operatingCost
netIncome = max(0, grossRevenue - operatingCost)

consume(cycle.consumption)
cash += netIncome
reputation = updateReputation(reputation, occupancy)
saveSettlementLog()
```

## 12. 第一版暂不做

```txt
不做真实玩家间餐厅竞争
不做三类顾客群体拆分
不做 CMO 系统
不做关店评级惩罚
不做装修停工时间
不做负利润破产
不做完整品质批次库存
```

这些可以作为第二阶段扩展。第一版只需要让玩家理解：菜单、员工、风格、价格会共同影响评级、上座率和收入。
