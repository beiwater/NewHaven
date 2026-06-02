import { Container, Graphics } from 'pixi.js'

/**
 * Highlight tiles / clickable areas on the map.
 * Draws a semi-transparent overlay with grid hints.
 */
export function createHighlightTile(
  x: number,
  y: number,
  width: number,
  height: number,
  color: number = 0xfbbf24,
  alpha: number = 0.25,
): Container {
  const tile = new Container()
  tile.x = x
  tile.y = y

  const g = new Graphics()
  g.rect(0, 0, width, height)
  g.fill({ color, alpha })
  g.stroke({ width: 1.5, color, alpha: 0.6 })
  tile.addChild(g)

  return tile
}

/**
 * Draw a dashed selection rectangle around the active building.
 */
export function createSelectionHighlight(
  x: number,
  y: number,
  width: number,
  height: number,
): Container {
  const sel = new Container()
  sel.x = x - 8
  sel.y = y - 24

  const g = new Graphics()
  g.roundRect(0, 0, width + 16, height + 36, 8)
  g.stroke({ width: 2, color: 0xfbbf24, alpha: 0.8 })
  g.fill({ color: 0xfbbf24, alpha: 0.08 })
  sel.addChild(g)

  return sel
}
