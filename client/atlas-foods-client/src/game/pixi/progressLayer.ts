import { Container, Graphics } from 'pixi.js'
import type { ProductionJob } from '@/game/types'

/**
 * Draw circular progress indicators for production buildings.
 * Each ring shows the elapsed/remaining ratio for a running job.
 */
export function createProgressRing(
  x: number,
  y: number,
  radius: number,
  progress: number, // 0.0 - 1.0
): Container {
  const ring = new Container()
  ring.x = x
  ring.y = y

  // Background ring
  const bgRing = new Graphics()
  bgRing.arc(0, 0, radius, 0, Math.PI * 2)
  bgRing.stroke({ width: 3, color: 0x5c3d2e, alpha: 0.3 })
  ring.addChild(bgRing)

  // Progress arc (clockwise from top)
  if (progress > 0) {
    const startAngle = -Math.PI / 2
    const endAngle = startAngle + Math.PI * 2 * Math.min(1, progress)
    const progressRing = new Graphics()
    progressRing.arc(0, 0, radius, startAngle, endAngle)
    progressRing.stroke({
      width: 3,
      color: progress >= 1 ? 0x4caf50 : 0x3b82f6,
    })
    ring.addChild(progressRing)
  }

  return ring
}

/**
 * Map building IDs to their progress state.
 * Call this to update progress rings each frame/tick.
 */
export function updateProgressRings(
  container: Container,
  jobs: ProductionJob[],
): void {
  // Remove old rings
  container.removeChildren()

  for (const job of jobs) {
    if (job.status !== 'running') continue

    const now = Date.now()
    const started = new Date(job.startedAt).getTime()
    const completesAt = new Date(job.completesAt).getTime()
    const total = completesAt - started
    const elapsed = now - started
    const progress = total > 0 ? Math.min(1, elapsed / total) : 0

    // Position needs to align with building positions
    // In production, match to building coordinates
    const ring = createProgressRing(40, 30, 30, progress)
    container.addChild(ring)
  }
}
