import { useState, type ReactNode } from 'react'
import { useCompany, useUpdateStoryProgress, type StoryProgress } from '@/api/company.api'
import { useUIStore } from '@/store/ui.store'
import { StoryPlayer } from '@/features/story/StoryPlayer'
import { chapter1ArrivalStory } from '@/features/story/chapter1Arrival.story'

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

export function StoryGate({ children }: { children: ReactNode }) {
  const { data: companyData, isLoading } = useCompany()
  const updateProgress = useUpdateStoryProgress()
  const [closedThisSession, setClosedThisSession] = useState(false)
  const storyProgress = readStoryProgress(companyData?.preferences?.storyProgress)
  const shouldShowStory = !closedThisSession &&
    (storyProgress?.status === 'not_started' || storyProgress?.status === 'in_progress')

  if (isLoading && !companyData) {
    return (
      <div className="flex h-screen w-screen items-center justify-center bg-amber-950 text-sm font-bold text-amber-100">
        Loading story...
      </div>
    )
  }

  const closeStory = (status: 'in_progress' | 'completed' | 'skipped') => {
    setClosedThisSession(true)
    useUIStore.getState().setActiveView('map')
    updateProgress.mutate({
      storyId: chapter1ArrivalStory.id,
      stepId: storyProgress?.stepId ?? chapter1ArrivalStory.firstStepId,
      status,
    })
  }

  // When the last story step is reached, set in_progress (not completed)
  // so the backend auto-completes on first production claim.
  const handleLastStep = () => {
    closeStory('in_progress')
  }

  if (shouldShowStory) {
    return (
      <StoryPlayer
        story={chapter1ArrivalStory}
        initialStepId={storyProgress?.stepId}
        onProgress={(stepId) => updateProgress.mutate({
          storyId: chapter1ArrivalStory.id,
          stepId,
          status: 'in_progress',
        })}
        onComplete={() => handleLastStep()}
        onSkip={() => closeStory('skipped')}
      />
    )
  }

  return <>{children}</>
}

function readStoryProgress(value: unknown): StoryProgress | undefined {
  if (!isRecord(value) || !isRecord(value.chapter1Arrival)) return undefined
  const status = value.chapter1Arrival.status
  const stepId = value.chapter1Arrival.stepId
  if (
    (status === 'not_started' || status === 'in_progress' || status === 'completed' || status === 'skipped') &&
    typeof stepId === 'string'
  ) {
    return { status, stepId }
  }
  return undefined
}
