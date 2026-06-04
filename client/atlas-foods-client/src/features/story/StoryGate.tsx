import { useState, type ReactNode } from 'react'
import { useCompany, useSavePreferences } from '@/api/company.api'
import { useUIStore } from '@/store/ui.store'
import { StoryPlayer } from '@/features/story/StoryPlayer'
import { chapter1ArrivalStory } from '@/features/story/chapter1Arrival.story'

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

export function StoryGate({ children }: { children: ReactNode }) {
  const { data: companyData, isLoading } = useCompany()
  const savePreferences = useSavePreferences()
  const [completedThisSession, setCompletedThisSession] = useState(false)
  const storyProgress = companyData?.preferences?.storyProgress
  const chapterCompleted =
    completedThisSession ||
    (isRecord(storyProgress) && storyProgress.chapter1Arrival === 'completed')

  if (isLoading && !companyData) {
    return (
      <div className="flex h-screen w-screen items-center justify-center bg-amber-950 text-sm font-bold text-amber-100">
        Loading story...
      </div>
    )
  }

  if (!chapterCompleted) {
    return (
      <StoryPlayer
        story={chapter1ArrivalStory}
        onComplete={() => {
          const currentProgress = isRecord(storyProgress) ? storyProgress : {}
          setCompletedThisSession(true)
          useUIStore.getState().setActiveView('map')
          savePreferences.mutate({
            storyProgress: {
              ...currentProgress,
              chapter1Arrival: 'completed',
            },
          })
        }}
      />
    )
  }

  return <>{children}</>
}
