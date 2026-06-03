import { Application } from 'pixi.js'

export async function createGameApp(container: HTMLElement): Promise<Application> {
  const app = new Application()

  await app.init({
    resizeTo: container,
    backgroundColor: 0xe8dcc8,
    antialias: true,
    resolution: window.devicePixelRatio || 1,
    autoDensity: true,
  })

  container.appendChild(app.canvas as HTMLCanvasElement)

  return app
}
