import { Container, Sprite, Graphics, Text, TextStyle, type Texture } from 'pixi.js'

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

  const bw = 100
  const bh = 80

  // Selection glow (behind the sprite)
  if (options.isSelected) {
    const glow = new Graphics()
    glow.ellipse(bw / 2, bh + 14, 60, 16)
    glow.fill({ color: 0xfbbf24, alpha: 0.35 })

    const glowLarge = new Graphics()
    glowLarge.ellipse(bw / 2, bh + 14, 80, 24)
    glowLarge.fill({ color: 0xfbbf24, alpha: 0.12 })
    node.addChild(glowLarge)
    node.addChild(glow)
  }

  // Building sprite
  const sprite = new Sprite(spriteTexture)
  sprite.anchor.set(0.5, 1)
  sprite.x = bw / 2
  sprite.y = bh
  sprite.scale.set(0.4) // Scale down for the game map
  node.addChild(sprite)

  // Level badge
  const lvlBg = new Graphics()
  lvlBg.circle(bw - 12, -8, 11)
  lvlBg.fill({ color: 0x5c3d2e })
  lvlBg.stroke({ width: 1.5, color: 0xf5e6c8 })
  node.addChild(lvlBg)

  const lvlText = new Text({
    text: `${options.level}`,
    style: new TextStyle({
      fontSize: 10,
      fontWeight: 'bold',
      fill: 0xf5e6c8,
      fontFamily: 'Inter, system-ui, sans-serif',
    }),
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
    const timerText = new Text({
      text: options.status === 'ready' ? '✅ Collect' : '⏳ In Progress',
      style: timerStyle,
    })
    timerText.anchor.set(0.5, 0)
    timerText.x = bw / 2
    timerText.y = bh + 28 + (options.isSelected ? 14 : 0)
    node.addChild(timerText)
  }

  node.on('pointerdown', onClick)

  return node
}
