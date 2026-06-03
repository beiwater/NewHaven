# 素材规格

## 总原则

所有可放置建筑素材必须是透明背景 PNG。
素材必须按建筑、等级、状态命名，不允许随手命名。

## 推荐尺寸

| 类型 | 尺寸 | 说明 |
| --- | --- | --- |
| 单帧建筑 | 512x512 | 透明背景 PNG |
| 建筑动画帧 | 512x512 每帧 | 横向 spritesheet 或独立帧 |
| 地图背景 | 2048x1152 | 不透明背景图，可用于概念或主屏 |
| 头像 | 512x512 | 透明背景 PNG |
| UI 图标 | 256x256 | 透明背景 PNG |

## 动画规格

| state | 帧数 | FPS | 循环 |
| --- | --- | --- | --- |
| idle | 4 | 6 | 是 |
| working | 6 | 8 | 是 |
| ready | 4 | 6 | 是 |
| construction | 6 | 8 | 是 |

## 命名规范

建筑单帧：

`{building_id}_lv{level}_{state}.png`

建筑 spritesheet：

`{building_id}_lv{level}_{state}_strip_{frames}f.png`

头像：

`avatar_{role}_{variant}.png`

地图背景：

`map_background_{variant}.png`

示例：

- `grain_plot_lv1_idle.png`
- `grain_plot_lv1_working_strip_6f.png`
- `bakery_shop_lv3_ready_strip_4f.png`
- `avatar_advisor_v1.png`
- `map_background_v1.png`

## 目录规范

生成源文件和最终透明图分开保存。

- `art-src/generated/`：AI 原始生成图
- `art-src/transparent/`：抠好透明背景的最终 PNG
- `art-src/spritesheets/`：动画条图
- `art-src/backgrounds/`：地图背景和场景图
- `art-src/avatars/`：角色头像
- `art-src/prompts/`：实际使用过的提示词

未来接入前端时，再从 `art-src` 复制精选素材到正式 `src/assets`，不要让实验素材污染源码目录。

## 透明背景策略

优先生成纯色 chroma-key 背景，再本地抠透明。

默认 key color：`#00ff00`

如果素材本体含大量绿色，则改用：`#ff00ff`

生成提示中必须写明：

- perfectly flat solid chroma-key background
- no shadows on background
- no texture or gradient
- subject fully separated from background
- no text
- no watermark
