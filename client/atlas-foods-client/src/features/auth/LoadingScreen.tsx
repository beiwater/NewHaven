import { useEffect, useRef, useState } from 'react'

interface LoadingScreenProps {
  statusText?: string
  onFinished: () => void
  minDurationMs?: number
}

export function LoadingScreen({
  statusText = '四处奔走 筹措物资中',
  onFinished,
  minDurationMs = 3000,
}: LoadingScreenProps) {
  const [visible, setVisible] = useState(true)
  const spriteRef = useRef<HTMLDivElement>(null)
  const frameRef = useRef(0)
  const timerRef = useRef<ReturnType<typeof setInterval>>()

  useEffect(() => {
    const sprite = spriteRef.current
    if (!sprite) return

    const FRAMES = [0, 1, 2, 3]
    const RATE = 500
    const COLS = 2

    function getFrameSize() {
      return { w: sprite.clientWidth, h: sprite.clientHeight }
    }

    function show(n: number) {
      const { w, h } = getFrameSize()
      const idx = FRAMES[n]
      const row = Math.floor(idx / COLS)
      const col = idx % COLS
      sprite.style.backgroundSize = `${w * 2}px ${h * 2}px`
      sprite.style.backgroundPosition = `${-col * w}px ${-row * h}px`
    }

    show(0)
    timerRef.current = setInterval(() => {
      frameRef.current = (frameRef.current + 1) % FRAMES.length
      show(frameRef.current)
    }, RATE)

    function onResize() {
      show(frameRef.current)
    }
    window.addEventListener('resize', onResize)

    // Minimum duration timer
    const minTimer = setTimeout(() => {
      setVisible(false)
      onFinished()
    }, minDurationMs)

    return () => {
      clearInterval(timerRef.current)
      clearTimeout(minTimer)
      window.removeEventListener('resize', onResize)
    }
  }, [minDurationMs, onFinished])

  if (!visible) return null

  return (
    <div className="fixed inset-0 z-[9999] bg-gradient-to-br from-amber-950 via-amber-900 to-stone-950 flex items-center justify-center">
      <div className="text-center">
        {/* Sprite */}
        <div
          ref={spriteRef}
          className="mx-auto w-[180px] h-[180px] sm:w-[200px] sm:h-[200px]"
          style={{
            background: 'url(/assets/running_sprite_sheet_transparen_4t.png) no-repeat',
            backgroundPosition: '0 0',
            backgroundSize: '400px 400px',
          }}
        />

        {/* Status text */}
        <p className="mt-7 text-sm sm:text-base text-amber-200/80 tracking-[2px] animate-pulse">
          {statusText}
        </p>

        {/* Progress bar */}
        <div className="mx-auto mt-3 w-[180px] sm:w-[200px] h-1 bg-amber-900/50 rounded-sm relative overflow-hidden">
          <div className="absolute top-0 left-[-32px] w-8 h-1 bg-amber-400 rounded-sm animate-[flowRight_1.6s_ease-in-out_infinite]" />
        </div>
      </div>

      {/* Inject the flowRight keyframe */}
      <style>{`
        @keyframes flowRight {
          0%   { left: -32px; }
          100% { left: 200px; }
        }
        @media (max-width: 640px) {
          .flow-track { width: 160px; }
          @keyframes flowRight {
            0%   { left: -32px; }
            100% { left: 160px; }
          }
        }
      `}</style>
    </div>
  )
}
