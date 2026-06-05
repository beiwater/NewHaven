import { useEffect, useState } from 'react'
import { formatDuration } from '@/game/executives'

export function useTrainingNow(active: boolean) {
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    if (!active) return
    const id = window.setInterval(() => setNow(Date.now()), 1000)
    return () => window.clearInterval(id)
  }, [active])

  return now
}

export function formatTrainingRemaining(trainingEndTime: string | undefined, now: number) {
  if (!trainingEndTime) return 'Training...'

  const endTime = new Date(trainingEndTime).getTime()
  if (!Number.isFinite(endTime)) return 'Training...'

  const remainingSeconds = Math.max(0, Math.ceil((endTime - now) / 1000))
  return remainingSeconds > 0 ? formatDuration(remainingSeconds) : 'Complete'
}
