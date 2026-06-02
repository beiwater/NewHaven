import { Container, Sprite, type Texture } from 'pixi.js'

/**
 * Create the map background from the actual art asset.
 * Scales to fill the canvas while preserving aspect ratio.
 */
export function createMapBackground(texture: Texture, canvasWidth: number, canvasHeight: number): Container {
  const bg = new Container()
  const sprite = new Sprite(texture)
  sprite.anchor.set(0, 0)

  // Scale to cover the canvas
  const scaleX = canvasWidth / texture.width
  const scaleY = canvasHeight / texture.height
  sprite.scale.set(Math.max(scaleX, scaleY))

  // Center the image if it's larger than the canvas
  if (scaleX > scaleY) {
    sprite.x = 0
    sprite.y = (canvasHeight - texture.height * sprite.scale.y) / 2
  } else {
    sprite.y = 0
    sprite.x = (canvasWidth - texture.width * sprite.scale.x) / 2
  }

  bg.addChild(sprite)
  return bg
}
