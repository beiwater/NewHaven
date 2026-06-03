import { Container, Sprite, Graphics, Text, TextStyle, type Texture } from 'pixi.js'

/** Uniform display size for all building sprites (before anchor offset) */
const BUILDING_DISPLAY_W = 80
const BUILDING_DISPLAY_H = 80

const labelStyle = new TextStyle({
  fontSize: 11,
  fontWeight: 'bold',
  fill: 0xffffff,
  fontFamily: 'Inter, system-ui, sans-serif',
  stroke: { color: 0x000000, width: 3 },
})

const timerStyle = new TextStyle({
  fontSize: 10,
  fill: 0xcccccc,
  fontFamily: 'Inter, system-ui, sans-serif',
  stroke: { color: 0x000000, width: 2 },
})

/** Create a building node with actual building sprite + selection glow + labels */
export function createBuildingNode(
  spriteTexture: Texture,
  options: {
    name: string
    level: number
    status?: string
    isSelected: boolean
  },
  onClick: () => void,
): Container {
  const node = new Container()
  node.eventMode = 'static'
  node.cursor = 'pointer'

  // Container width/height for layout
  const bw = 100
  const bh = 80

  // Selection glow (behind the sprite)
  if (options.isSelected) {
    const glow = new Graphics()
    glow.rect(0, 0, bw, bh)
    glow.fill({ color: 0xf5e6c8, alpha: 0.3 })
    glow.stroke({ width: 2, color: 0xf5a623 })
    node.addChild(glow)
  }

  // Building sprite — scale uniformly to fit display size regardless of source dimensions
  const sprite = new Sprite(spriteTexture)
  const texW = spriteTexture.width
  const texH = spriteTexture.height
  if (texW > 0 && texH > 0) {
    const scale = Math.min(BUILDING_DISPLAY_W / texW, BUILDING_DISPLAY_H / texH)
    sprite.scale.set(scale)
  }
  sprite.anchor.set(0.5, 1)
  sprite.x = bw / 2
  sprite.y = bh
  node.addChild(sprite)

  // Level badge
  const lvlBg = new Graphics()
  lvlBg.circle(bw - 12, -8, 11)
  lvlBg.fill({ color: 0x5c3d2e })
  lvlBg.stroke({ width: 1.5, color: 0xf5e6c8 })
  node.addChild(lvlBg)

  const lvlText = new Text({
    text: String(options.level),
    style: { fontSize: 10, fontWeight: 'bold', fill: 0xf5e6c8, fontFamily: 'Inter, system-ui, sans-serif' },
  })
  lvlText.anchor.set(0.5)
  lvlText.x = bw - 12
  lvlText.y = -8
  node.addChild(lvlText)

  // Name label
  const nameText = new Text({
    text: options.name,
    style: labelStyle,
  })
  nameText.anchor.set(0.5, 0)
  nameText.x = bw / 2
  nameText.y = bh + 10 + (options.isSelected ? 14 : 0)
  node.addChild(nameText)

  // Status / timer label
  if (options.status && options.status !== 'idle') {
    const statusText = new Text({
      text: options.status === 'running' ? '⏳' : '✅',
      style: timerStyle,
    })
    statusText.anchor.set(0, 0)
    statusText.x = 4
    statusText.y = 4
    node.addChild(statusText)
  }

  node.on('pointerdown', onClick)

  return node
}
