import { useMemo, useState } from 'react'
import type { StoryDefinition, StoryStep } from './story.types'

const HARBOR_BG = '/assets/story/chapter-1/newhaven_harbor.png'
const CECIL_FULL = '/assets/story/characters/cecil_full.png'
const BOAT = '/assets/story/props/arrival_boat.png'
const CLOUD_1 = '/assets/story/effects/cloud_soft_1.png'
const CLOUD_2 = '/assets/story/effects/cloud_soft_2.png'

interface StoryPlayerProps {
  story: StoryDefinition
  onComplete: () => void
}

export function StoryPlayer({ story, onComplete }: StoryPlayerProps) {
  const [currentStepId, setCurrentStepId] = useState(story.firstStepId)

  const stepMap = useMemo(() => {
    return new Map(story.steps.map((step) => [step.id, step]))
  }, [story.steps])

  const step = stepMap.get(currentStepId) ?? story.steps[0]
  const isBlack = step.screen === 'black'
  const showPortrait = step.portrait === 'shadow' || step.portrait === 'cecil'
  const showBoat = step.boat === 'arriving' || step.boat === 'docked'
  const showChoices = step.kind === 'choice' && step.choices?.length

  const advance = () => {
    if (showChoices) return
    if (step.next) {
      setCurrentStepId(step.next)
      return
    }
    onComplete()
  }

  const choose = (next: string) => {
    setCurrentStepId(next)
  }

  return (
    <section
      className={`story-root ${isBlack ? 'story-root-black' : 'story-root-harbor'}`}
      aria-label={story.title}
      onClick={advance}
    >
      {!isBlack && (
        <>
          <img className="story-bg" src={HARBOR_BG} alt="" draggable={false} />
          <div className="story-sunwash" />
          <img className="story-cloud story-cloud-one" src={CLOUD_1} alt="" draggable={false} />
          <img className="story-cloud story-cloud-two" src={CLOUD_2} alt="" draggable={false} />
          {showBoat && (
            <img
              className={`story-boat ${step.boat === 'arriving' ? 'story-boat-arriving' : 'story-boat-docked'}`}
              src={BOAT}
              alt=""
              draggable={false}
            />
          )}
          {showPortrait && <CecilPortrait step={step} />}
        </>
      )}

      {step.locationTitle && (
        <div key={`${step.id}-location`} className="story-location-title">
          {step.locationTitle}
        </div>
      )}

      {isBlack ? (
        <BlackScreenStep step={step} />
      ) : (
        <DialoguePanel step={step} showChoices={!!showChoices} onChoose={choose} onAdvance={advance} />
      )}
    </section>
  )
}

function CecilPortrait({ step }: { step: StoryStep }) {
  return (
    <div className={`story-portrait-wrap ${step.portrait === 'shadow' ? 'story-portrait-shadow' : ''}`}>
      <img className="story-portrait" src={CECIL_FULL} alt="Cecil Ashwing" draggable={false} />
    </div>
  )
}

function BlackScreenStep({ step }: { step: StoryStep }) {
  const isTitle = step.kind === 'title'
  return (
    <div key={step.id} className="story-black-content">
      <p className={isTitle ? 'story-chapter-title' : 'story-intro-line'}>{step.text}</p>
      <span className="story-continue-hint">Continue</span>
    </div>
  )
}

function DialoguePanel({
  step,
  showChoices,
  onChoose,
  onAdvance,
}: {
  step: StoryStep
  showChoices: boolean
  onChoose: (next: string) => void
  onAdvance: () => void
}) {
  return (
    <div className={`story-dialogue ${step.kind === 'system' ? 'story-dialogue-system' : ''}`} onClick={(event) => event.stopPropagation()}>
      {step.speaker && <div className="story-speaker">{step.speaker}</div>}
      <p key={step.id} className="story-dialogue-text">{step.text}</p>
      {showChoices ? (
        <div className="story-choice-list">
          {step.choices?.map((choice) => (
            <button key={choice.label} type="button" className="story-choice" onClick={() => onChoose(choice.next)}>
              {choice.label}
            </button>
          ))}
        </div>
      ) : (
        <button type="button" className="story-next-button" onClick={onAdvance}>
          {step.actionLabel ?? 'Continue'}
        </button>
      )}
    </div>
  )
}
