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

const harvestStyle = new TextStyle({
  fontSize: 10,
  fontWeight: 'bold',
  fill: 0x3f6212,
  fontFamily: 'Inter, system-ui, sans-serif',
})

/** Create a building node with actual building sprite + selection glow + labels */
export function createBuildingNode(
  spriteTexture: Texture,
  options: {
    name: string
    level: number
    status?: string
    progress?: number
    collectableAmount?: number
    pulse?: number
    resourceTexture?: Texture
    isSelected: boolean
  },
  onClick: () => void,
  onCollect?: () => void,
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

  if (options.status === 'running' || (options.collectableAmount ?? 0) > 0) {
    const workGlow = new Graphics()
    const alpha = 0.12 + (options.pulse ?? 0) * 0.08
    workGlow.ellipse(bw / 2, bh - 10, 46, 18)
    workGlow.fill({ color: (options.collectableAmount ?? 0) > 0 ? 0xb6d46a : 0xf5c16c, alpha })
    node.addChild(workGlow)
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
  if (options.status === 'running') {
    const breathe = 1 + (options.pulse ?? 0) * 0.025
    sprite.scale.x *= breathe
    sprite.scale.y *= 1 + (options.pulse ?? 0) * 0.015
  }
  node.addChild(sprite)

  if (options.status === 'running') {
    const seedA = new Graphics()
    seedA.circle(18, 52 - (options.pulse ?? 0) * 8, 2.5)
    seedA.fill({ color: 0x7aa34b, alpha: 0.65 })
    node.addChild(seedA)
    const seedB = new Graphics()
    seedB.circle(82, 44 + (options.pulse ?? 0) * 6, 2)
    seedB.fill({ color: 0xd9a441, alpha: 0.55 })
    node.addChild(seedB)
  }

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

  if (typeof options.progress === 'number' && options.status === 'running') {
    const progressBg = new Graphics()
    progressBg.roundRect(14, bh + 1, 72, 5, 3)
    progressBg.fill({ color: 0x5c3d2e, alpha: 0.24 })
    node.addChild(progressBg)

    const progressFill = new Graphics()
    progressFill.roundRect(14, bh + 1, 72 * Math.max(0.04, Math.min(1, options.progress)), 5, 3)
    progressFill.fill({ color: 0x7aa34b, alpha: 0.9 })
    node.addChild(progressFill)
  }

  if ((options.collectableAmount ?? 0) > 0) {
    const badge = new Container()
    badge.eventMode = 'static'
    badge.cursor = 'pointer'
    badge.x = bw - 18
    badge.y = 22

    const badgeBg = new Graphics()
    badgeBg.roundRect(-30, -16, 58, 28, 9)
    badgeBg.fill({ color: 0xfff7dc, alpha: 0.96 })
    badgeBg.stroke({ width: 1.5, color: 0x8f6b2d, alpha: 0.75 })
    badge.addChild(badgeBg)

    if (options.resourceTexture) {
      const icon = new Sprite(options.resourceTexture)
      const scale = Math.min(18 / Math.max(1, options.resourceTexture.width), 18 / Math.max(1, options.resourceTexture.height))
      icon.scale.set(scale)
      icon.anchor.set(0.5)
      icon.x = -17
      icon.y = -2
      badge.addChild(icon)
    }

    const amountText = new Text({
      text: String(options.collectableAmount),
      style: harvestStyle,
    })
    amountText.anchor.set(0, 0.5)
    amountText.x = -5
    amountText.y = -2
    badge.addChild(amountText)
    badge.on('pointerdown', (event) => {
      event.stopPropagation()
      onCollect?.()
    })
    node.addChild(badge)
  }

  node.on('pointerdown', onClick)

  return node
}
