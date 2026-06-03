# AI 绘画提示词模板

## 建筑透明素材模板

Use case: game asset sprite
Asset type: transparent building sprite for a 2.5D idle tycoon map game
Primary request: Create an isolated {building_name} building sprite, {level_description}, state: {state_description}.
Scene/backdrop: Perfectly flat solid {key_color} chroma-key background for background removal.
Subject: One readable 2.5D isometric game building for the production chain: {industry_chain}. The building id is {building_id}, level {level}.
Style/medium: Warm handcrafted 2.5D strategy game art, cozy business management game, clear silhouette, slight outline, readable at small size.
Composition/framing: Centered single building, generous padding, three-quarter isometric top-down angle, no cropped edges.
Lighting/mood: Warm afternoon light, friendly but productive.
Color palette: Warm cream, moss green, amber, teal accents, natural wood/brick/metal depending on level.
Materials/textures: {materials}
Constraints: Transparent-ready sprite source. Background must be one uniform {key_color}; no shadows, gradients, floor plane, texture, or reflections on the background. No text, no logo, no watermark.
Avoid: photorealism, literal farm field scene, stock market UI, dark cyberpunk, cluttered background.

## 建筑动画 spritesheet 模板

Use case: game asset spritesheet
Asset type: horizontal spritesheet for a 2.5D idle tycoon building
Primary request: Create a {frame_count}-frame horizontal spritesheet of {building_name}, level {level}, animation state {state}.
Scene/backdrop: Perfectly flat solid {key_color} chroma-key background for background removal.
Subject: Same building repeated across frames with subtle animation variation only: {animation_detail}.
Composition/framing: Horizontal strip, exactly {frame_count} evenly spaced frames, each frame centered and same scale, no cropping.
Constraints: Keep building identity identical across frames. No text, no watermark. Uniform chroma-key background only.
Avoid: changing camera angle, changing building design between frames, complex background, large motion blur.

## 地图背景模板

Use case: game background concept
Asset type: 16:9 map background for game frontend
Primary request: Create a warm 2.5D production map background for an idle resource management game.
Scene/backdrop: Expandable business park map with roads, empty plots, soft terrain, small district markers, and room for building sprites.
Subject: A playable map backdrop, not a completed illustration overloaded with unique buildings.
Style/medium: Polished strategy game map, clean readable shapes, cozy tycoon mood.
Composition/framing: 16:9 desktop layout, center-left clear map area, subtle roads and plots, room for UI overlays.
Constraints: No embedded UI text, no watermark, no stock charts, no realistic farm photo look.

## 头像模板

Use case: game character portrait
Asset type: transparent advisor avatar
Primary request: Create an isolated friendly business advisor avatar for a cozy industry-chain tycoon game.
Scene/backdrop: Perfectly flat solid {key_color} chroma-key background for background removal.
Subject: Bust portrait of a helpful operations manager, warm confident expression, business-casual outfit with subtle food/production motif.
Style/medium: 2D stylized game portrait, clean silhouette, expressive face, readable at small UI size.
Composition/framing: Centered bust, shoulders visible, generous padding.
Constraints: No text, no logo, no watermark, uniform chroma-key background only.
