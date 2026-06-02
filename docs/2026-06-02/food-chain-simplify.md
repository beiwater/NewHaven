# Food Chain Simplification Plan (v3)

**Date:** 2026-06-02  
**Goal:** 缩减产业链，聚焦食品，增加销售建筑和终端食品种类

---

## 约束条件

- 从地里到终端产品，不能超过 **4 个建筑**
- 要有 **2 步**、**3 步**、**4 步** 的食品
- 一个 **全能餐厅** 收集所有终端食品
- **4 个中间加工建筑** + **4 个销售建筑**

---

## 建筑总览 (10 种)

```
Farm ─→ Mill → Kitchen → Bakery ─→ Restaurant / Cafe
  │         │       │       │
  │         │       │       └── Food Truck
  │         │       │
  │         │       └── Deli
  │         │
  └── Barn ─┴── Market Stall
```

### 原材料建筑 (2)

| 建筑 | 产出 |
|------|------|
| **Farm** (农场) | Wheat, Vegetables, Fruits |
| **Barn** (牧场) | Eggs, Milk |

### 中间加工建筑 (4)

| 建筑 | 加工 | 最短步数 | 最长步数 |
|------|------|---------|---------|
| **Mill** (作坊) | Flour, Butter, Juice, Cheese | 2 步到终端 | — |
| **Kitchen** (厨房) | Dough, Salad, Omelette, Jam | 2 步 | 3 步 |
| **Bakery** (面包房) | Bread, Pizza, Cake | 4 步到顶 | 4 步 |
| **Deli** (熟食坊) | Sandwich, Ice Cream, Fruit Bowl | 3 步 | 4 步 |

### 销售建筑 (4)

| 建筑 | 卖什么 | 特点 |
|------|--------|------|
| **Restaurant** (全能餐厅) | 全部终端食品 | 高价 + 额外 XP，终端终点站 |
| **Market Stall** (集市摊位) | 原材料 + 加工品 | 薄利多销，周转快 |
| **Cafe** (咖啡馆) | Bread, Cake, Juice | 面包 + 蛋糕 + 果汁套餐，中端 |
| **Food Truck** (餐车) | Pizza, Omelette, Salad, Sandwich | 快餐，走量 |

---

## 资源体系 (19 种)

### 原材料 (Raw)

| ID | 名称 | 产自 |
|----|------|------|
| 1 | Wheat | Farm |
| 2 | Vegetables | Farm |
| 3 | Fruits | Farm |
| 4 | Eggs | Barn |
| 5 | Milk | Barn |

### 加工品 (Processed)

| ID | 名称 | 步数 | 配方 | 可销售于 |
|----|------|------|------|---------|
| 6 | Flour | 2 | Wheat | Market Stall |
| 7 | Butter | 2 | Milk | Market Stall / Kitchen |
| 8 | Juice | 2 | Fruits | Market Stall / Cafe |
| 9 | Cheese | 2 | Milk | Market Stall / Deli |
| 10 | Dough | 3 | Flour + Eggs | Kitchen → Bakery |
| 11 | Salad | 2 | Vegetables + Eggs | Market Stall / Food Truck |
| 12 | Omelette | 2 | Eggs + Vegetables | Market Stall / Food Truck |
| 13 | Jam | 2 | Fruits | Market Stall / Cafe |

### 终端食品 (Terminal)

| ID | 名称 | 步数 | 配方 | 销售建筑 |
|----|------|------|------|---------|
| 14 | Bread | 4 | Dough | Restaurant / Cafe |
| 15 | Pizza | 4 | Dough + Vegetables | Restaurant / Food Truck |
| 16 | Cake | 4 | Dough + Butter + Eggs | Restaurant / Cafe |
| 17 | Sandwich | 3 | Bread + Cheese + Vegetables | Restaurant / Deli / Food Truck |
| 18 | Ice Cream | 3 | Milk + Fruits | Restaurant / Deli |
| 19 | Fruit Bowl | 3 | Fruits + Jam | Restaurant / Deli / Cafe |

### 公用事业

| ID | 名称 |
|----|------|
| 20 | Power |
| 21 | Water |

---

## 生产链一览

```
                                ┌── Flour (2)
Farm ── Wheat ──→ Mill ────────┼── Juice (2) → Cafe / Market Stall
       │                       │
       │                       └── Butter (2) → Market Stall
       │
       ├── Vegetables ─────────→ Kitchen ── Salad (2) → Food Truck
       │                              │
       │                              ├── Omelette (2) → Food Truck
       ├── Fruits ──────────→ Mill ───┤
       │                              └── Jam (2) → Cafe / Market Stall
       │
       │                        Deli ── Sandwich (3) → Food Truck / Restaurant
       │                              ├── Ice Cream (3) → Restaurant
       │                              └── Fruit Bowl (3) → Cafe / Restaurant
Barn ── Eggs ────────────────→ Kitchen ── Dough (3)
       │                              │
       └── Milk ─────────────→ Mill ───┤
                                       │
                                  Bakery ── Bread (4) → Restaurant / Cafe
                                       ├── Pizza (4) → Restaurant / Food Truck
                                       └── Cake (4) → Restaurant / Cafe
```

**最长路径：** Farm → Mill → Kitchen → Bakery = **4 建筑** ✅  
**最短路径：** Farm → Kitchen = **Salad/Omelette (2 步)** ✅

---

## 终端食品去向矩阵

| 食品 | 餐厅 | 咖啡馆 | 餐车 | 熟食坊 | 集市 |
|------|------|--------|------|--------|------|
| Bread | ✅ | ✅ | | | |
| Pizza | ✅ | | ✅ | | |
| Cake | ✅ | ✅ | | | |
| Sandwich | ✅ | | ✅ | ✅ | |
| Ice Cream | ✅ | | | ✅ | |
| Fruit Bowl | ✅ | ✅ | | ✅ | |
| Flour | | | | | ✅ |
| Juice | | ✅ | | | ✅ |
| Salad | | | ✅ | | ✅ |
| Omelette | | | ✅ | | ✅ |
| Jam | | ✅ | | | ✅ |
| Cheese | | | | ✅ | ✅ |

---

## 需要修改的文件

### 后端数据

| 文件 | 改动 |
|------|------|
| `decompiled/data/resources.json` | 从 23+ 精简到 21 种 (含 Power/Water)，重写配方 |
| `decompiled/data/buildings.json` | 从 4 种扩到 10 种 |
| `decompiled/data/economy_model.json` | 简化 |
| `decompiled/data/resource_lookups.json` | 同步更新 |
| `configs/game.json` | `bot_resources` 更新，起步资金/订单数调整 |

### 后端代码

| 文件 | 改动 |
|------|------|
| `internal/service/building_shop.go` | `BuildingMarket()` 返回 10 种建筑 |
| `internal/service/service.go` | `ResourcesWithMarket()` 更新 |
| `internal/service/production.go` | `productionIDsForKind()` 映射更新 |

### 前端

| 文件 | 改动 |
|------|------|
| `src/game/resources.ts` | `FALLBACK_MARKET_RESOURCES` 精简到 21 种；`MARKET_GROUPS` 分 Raw / Processed / Terminal 三组；`resourceIcon()` 补充全部映射 |
| `src/game/GameCanvas.tsx` | `BUILDING_TEXTURES` 增加 kind 5-10 的 6 种新建筑 |
| `src/features/buildings/BuildingCard.tsx` | `BUILDING_PREVIEW` 同步增加 |
| `src/features/inventory/InventoryBar.tsx` | 更新 emoji/图标 |
| `src/features/market/MarketPage.tsx` | 分组 tab 更新 |

### 资产新增

| 新增资源图 | 新增建筑图 |
|-----------|-----------|
| vegetables | Deli (熟食坊) |
| eggs | Cafe (咖啡馆) |
| milk | Food Truck (餐车) |
| butter | (Restaurant 已有/可复用) |
| dough | (Market Stall 新建) |
| juice | |
| cheese | |
| salad | |
| omelette | |
| jam | |
| ice cream | |
| sandwich | |
| cake | |
| fruit bowl | |
