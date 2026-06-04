你现在在 New Haven 项目根目录中工作。这是一个多人网页经济经营游戏，目录大致为：

* `backend/`：Go API server
* `client/`：React + PixiJS 前端
* `assets/`：游戏素材资源
* `docs/`：设计文档
* `scripts/`：工具脚本

任务：把游戏音效系统接入前端，不要改动核心经济逻辑，不要破坏现有 UI 和游戏流程。

目标：为游戏加入统一的音频管理系统，支持 UI 音效、金币/交易音效、建筑音效、生产音效、餐厅音效、研究/高管音效、通知音效、环境音和 BGM。

请先检查当前前端技术栈和已有资源路径。如果项目已经使用 PixiJS sound 插件、Howler.js 或其他音频库，优先复用现有方案；如果没有音频库，优先使用浏览器原生 Web Audio API 或 HTMLAudioElement 实现，不要随便引入很重的依赖。若确实需要新增依赖，请先说明理由，并选择轻量、稳定、适合网页游戏的方案。

请完成以下内容：

1. 创建音频资源目录结构

在 `assets/audio/` 或前端静态资源目录中建立类似结构：

```text
assets/audio/
  sfx/
    ui/
    money/
    market/
    building/
    farm/
    production/
    restaurant/
    warehouse/
    research/
    social/
    system/
  ambience/
  bgm/
  LICENSES/
```

如果项目已有 public/static 资源目录，请按项目实际规范放置，但要保持结构清晰。

2. 创建音频 manifest

创建一个统一的音频配置文件，例如：

```text
client/src/audio/audioManifest.ts
```

或：

```text
client/src/assets/audioManifest.ts
```

里面定义所有音效 key 和路径。先不要假设所有文件都已经存在，允许缺失文件时静默失败并打印 warning。

需要支持这些 key：

```ts
ui_button_click
ui_button_hover
ui_confirm
ui_cancel
ui_error
ui_disabled
ui_panel_open
ui_panel_close
ui_tab_switch
ui_popup
ui_drag_start
ui_drag_drop
ui_scroll_tick

money_coin_gain
money_coin_spend
money_big_profit
money_loss

market_buy
market_sell
market_order_created
market_order_filled
market_price_up
market_price_down
market_volatility_alert
contract_signed
debt_issued
futures_open
futures_close

build_place
build_preview_move
build_confirm
build_construction_start
build_construction_complete
build_upgrade
build_repair
build_demolish
land_unlock
road_place

farm_plant_seed
farm_water_crop
farm_harvest
farm_crop_ready
barn_animal_feed
barn_collect_milk
barn_collect_eggs
resource_pickup

mill_start
mill_complete
kitchen_chop
kitchen_pan_sizzle
kitchen_recipe_complete
bakery_oven_start
bakery_bread_ready
cafe_coffee_pour
cafe_espresso
food_packaged

inventory_open
inventory_sort
warehouse_store
warehouse_takeout
truck_depart
truck_arrive
ship_dock
delivery_complete

restaurant_open
customer_enter
menu_set
dish_served
dish_sold
restaurant_revenue_collect
occupancy_up
occupancy_down
restaurant_level_up

research_start
research_progress_tick
research_complete
tech_unlock
executive_hire
executive_level_up
skill_assign
buff_activate
buff_expire

quest_accept
quest_complete
achievement_unlock
leaderboard_rank_up
chat_send
chat_receive
player_join
player_leave
trade_request
trade_accepted
trade_rejected

day_start
day_end
season_change
save_success
autosave
notification_general
warning_soft
critical_alert

amb_harbor_day
amb_harbor_night
amb_market
amb_restaurant
amb_kitchen
amb_farm
amb_barn
amb_warehouse
amb_cafe
amb_rain_town

bgm_main_menu
bgm_harbor_town
bgm_market
bgm_restaurant
bgm_farm
bgm_night
```

3. 实现 AudioManager

创建：

```text
client/src/audio/AudioManager.ts
```

要求：

* 支持 `init()`
* 支持 `unlockAudio()`，处理浏览器自动播放限制，第一次点击页面后解锁音频
* 支持 `preloadSfx(keys?: string[])`
* 支持 `playSfx(key, options?)`
* 支持 `playMusic(key, options?)`
* 支持 `stopMusic()`
* 支持 `fadeMusicTo(key, durationMs)`
* 支持 `setMasterVolume(value)`
* 支持 `setSfxVolume(value)`
* 支持 `setMusicVolume(value)`
* 支持 `mute()` / `unmute()` / `toggleMute()`
* 支持本地保存设置到 `localStorage`
* 缺失音频文件时不能报错中断游戏
* 同一个短音效短时间内重复播放时要做 throttle，避免疯狂点击导致爆音
* 支持随机轻微 pitch/volume variation，让金币、点击、收获这种声音不要每次完全一样
* 环境音和 BGM 要 loop
* UI 短音效不要 loop

4. 创建 React Hook

创建：

```text
client/src/audio/useAudio.ts
```

提供：

```ts
const { playSfx, playMusic, stopMusic, setMasterVolume, setSfxVolume, setMusicVolume, muted, toggleMute } = useAudio()
```

如果项目状态管理已经有 Zustand/Redux/Context，请接入现有方案；否则创建轻量 AudioProvider。

5. 接入全局 UI

在 App 初始化处调用 AudioManager.init()，并在第一次用户点击、键盘输入或触摸时调用 unlockAudio()。

在设置页面加入音频设置：

* 总音量 Master Volume
* 音效音量 SFX Volume
* 音乐音量 Music Volume
* 静音 Mute
* 测试按钮：播放 `ui_confirm`

如果当前还没有设置页面，则创建一个简单的音频设置组件，放到现有 Settings / Options 页面中。

6. 把音效接入游戏事件

请搜索现有前端代码，把音效加到合适位置。不要为了加音效重写业务逻辑，只在事件成功后调用 playSfx。

建议映射如下：

UI：

* 按钮点击：`ui_button_click`
* 悬停：`ui_button_hover`，注意不要太频繁，可只给主要按钮
* 确认：`ui_confirm`
* 取消/返回：`ui_cancel`
* 错误/余额不足：`ui_error`
* 弹窗打开：`ui_popup`
* 面板打开：`ui_panel_open`
* 面板关闭：`ui_panel_close`
* tab 切换：`ui_tab_switch`

经济：

* 获得金币/收入结算：`money_coin_gain`
* 花钱：`money_coin_spend`
* 大额盈利：`money_big_profit`
* 亏损：`money_loss`

市场交易：

* 买入：`market_buy`
* 卖出：`market_sell`
* 创建订单：`market_order_created`
* 订单成交：`market_order_filled`
* 价格上涨：`market_price_up`
* 价格下跌：`market_price_down`
* 剧烈波动：`market_volatility_alert`
* 合同签署：`contract_signed`
* 期货开仓：`futures_open`
* 期货平仓：`futures_close`

建筑：

* 放置建筑：`build_place`
* 确认建造：`build_confirm`
* 开始施工：`build_construction_start`
* 建筑完成：`build_construction_complete`
* 建筑升级：`build_upgrade`
* 建筑维修：`build_repair`
* 拆除建筑：`build_demolish`
* 解锁土地：`land_unlock`

农场/生产：

* 播种：`farm_plant_seed`
* 浇水：`farm_water_crop`
* 收获：`farm_harvest`
* 作物成熟：`farm_crop_ready`
* 喂养动物：`barn_animal_feed`
* 收牛奶：`barn_collect_milk`
* 收鸡蛋：`barn_collect_eggs`
* 资源拾取：`resource_pickup`

加工：

* 磨坊开始：`mill_start`
* 磨坊完成：`mill_complete`
* 厨房加工：`kitchen_chop`
* 煎炒：`kitchen_pan_sizzle`
* 菜品完成：`kitchen_recipe_complete`
* 烘焙开始：`bakery_oven_start`
* 面包出炉：`bakery_bread_ready`
* 咖啡制作：`cafe_espresso`
* 食品打包：`food_packaged`

库存/物流：

* 打开库存：`inventory_open`
* 整理库存：`inventory_sort`
* 入库：`warehouse_store`
* 出库：`warehouse_takeout`
* 货车出发：`truck_depart`
* 货车到达：`truck_arrive`
* 船靠港：`ship_dock`
* 物流完成：`delivery_complete`

餐厅：

* 餐厅开门：`restaurant_open`
* 顾客进店：`customer_enter`
* 设置菜单：`menu_set`
* 上菜：`dish_served`
* 售出菜品：`dish_sold`
* 收取餐厅收入：`restaurant_revenue_collect`
* 上座率上升：`occupancy_up`
* 上座率下降：`occupancy_down`
* 餐厅升级：`restaurant_level_up`

研究/高管：

* 开始研究：`research_start`
* 研究完成：`research_complete`
* 科技解锁：`tech_unlock`
* 招募高管：`executive_hire`
* 高管升级：`executive_level_up`
* 分配技能点：`skill_assign`
* 加成激活：`buff_activate`
* 加成结束：`buff_expire`

任务/社交：

* 接任务：`quest_accept`
* 任务完成：`quest_complete`
* 解锁成就：`achievement_unlock`
* 排名上升：`leaderboard_rank_up`
* 发送聊天：`chat_send`
* 收到聊天：`chat_receive`
* 玩家上线：`player_join`
* 玩家离线：`player_leave`
* 交易请求：`trade_request`
* 交易接受：`trade_accepted`
* 交易拒绝：`trade_rejected`

系统：

* 新一天：`day_start`
* 日结算：`day_end`
* 季节切换：`season_change`
* 保存成功：`save_success`
* 自动保存：`autosave`
* 普通通知：`notification_general`
* 轻度警告：`warning_soft`
* 重要警告：`critical_alert`

7. 接入 BGM 和环境音

根据场景切换音乐：

* 主菜单：`bgm_main_menu`
* 海港主城：`bgm_harbor_town`
* 市场页面：`bgm_market`
* 餐厅页面：`bgm_restaurant`
* 农场页面：`bgm_farm`
* 夜晚/结算界面：`bgm_night`

根据地图或页面切换环境音：

* 海港白天：`amb_harbor_day`
* 海港夜晚：`amb_harbor_night`
* 市场：`amb_market`
* 餐厅：`amb_restaurant`
* 厨房：`amb_kitchen`
* 农场：`amb_farm`
* 牧场：`amb_barn`
* 仓库：`amb_warehouse`
* 咖啡馆：`amb_cafe`
* 雨天：`amb_rain_town`

BGM 切换要淡入淡出，不要突然断开。环境音音量默认比 BGM 小。

8. 音量默认值

默认配置：

```ts
masterVolume = 0.8
sfxVolume = 0.75
musicVolume = 0.35
ambienceVolume = 0.25
muted = false
```

UI 点击音效要小，不要刺耳。金币、升级、成就可以稍微大一点。

9. 资源命名规范

如果音频文件不存在，先创建 manifest 占位，不要强行生成假文件。
请把预期文件命名为：

```text
assets/audio/sfx/ui/ui_button_click.wav
assets/audio/sfx/money/money_coin_gain.wav
assets/audio/sfx/market/market_buy.wav
assets/audio/sfx/building/build_place.wav
assets/audio/ambience/amb_harbor_day.ogg
assets/audio/bgm/bgm_harbor_town.ogg
```

短音效优先 `.wav`，长 BGM / ambience 优先 `.ogg`。

10. 增加开发辅助功能

在开发模式下提供一个简单 SoundTestPanel，可以列出所有 audio key，并点击播放。
只在 development 环境显示，不要影响正式版。

11. 代码质量要求

* TypeScript 类型清晰
* 不要到处散落字符串，统一从 audio manifest 取 key
* 不要因为音频加载失败导致游戏白屏
* 不要重复创建大量 audio 对象造成内存泄漏
* BGM 同一首不要重复播放
* 快速点击时要限制同一音效重复播放频率
* 代码要有简短注释
* 保持现有代码风格
* 完成后运行 lint/build/test，如果项目有这些命令
* 最后输出修改文件清单和使用说明

12. 最终交付

请实现完整音频系统，并在最后说明：

* 新增了哪些文件
* 修改了哪些文件
* 音频资源应该放在哪里
* 如何添加新音效
* 如何在代码里播放音效
* 如何调整音量
* 如何测试音效是否正常
