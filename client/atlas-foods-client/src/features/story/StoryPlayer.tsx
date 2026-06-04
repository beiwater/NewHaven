import { useMemo, useState } from 'react'
import { useEffect } from 'react'
import { audio } from '@/audio/AudioManager'
import { useTranslation } from 'react-i18next'
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
  const { t } = useTranslation()
  const [currentStepId, setCurrentStepId] = useState(story.firstStepId)

  const stepMap = useMemo(() => {
    return new Map(story.steps.map((step) => [step.id, step]))
  }, [story.steps])
  const step = stepMap.get(currentStepId) ?? story.steps[0]
  const isBlack = step.screen === 'black'
  const showPortrait = step.portrait === 'shadow' || step.portrait === 'cecil'
  const showBoat = step.boat === 'arriving' || step.boat === 'docked'
  const showChoices = step.kind === 'choice' && step.choices?.length

  const chapter = story.id === 'chapter1Arrival' ? 'chapter1' : story.id
  const st = (key: string, fallback?: string) => t(`story.${chapter}.${key}`, fallback ?? '')

  // Play BGM while story is shown
  useEffect(() => {
    audio.playMusic('bgm_main_menu')
    audio.playAmbience('amb_harbor_day')
  }, [])
  const advance = () => {
    audio.playSfx('ui_confirm', { volume: 0.4 })
    if (step.next) {
      setCurrentStepId(step.next)
    } else {
      onComplete()
    }
  }

  const choose = (next: string) => {
    audio.playSfx('ui_confirm', { volume: 0.4 })
    setCurrentStepId(next)
  }

  const textDisplay = st(step.id, step.text)
  const speakerDisplay = st(`${step.id}Speaker`, step.speaker)

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
          {st(step.id, step.locationTitle)}
        </div>
      )}

      {isBlack ? (
        <BlackScreenStep step={step} text={textDisplay} />
      ) : (
        <DialoguePanel
          step={step}
          showChoices={!!showChoices}
          onChoose={choose}
          onAdvance={advance}
          text={textDisplay}
          speaker={speakerDisplay}
          t={st}
        />
      )}
    </section>
  )
}

function CecilPortrait({ step }: { step: StoryStep }) {
  const isShadow = step.portrait === 'shadow'
  return (
    <img
      className={`story-portrait ${isShadow ? 'story-portrait-shadow' : 'story-portrait-reveal'}`}
      src={CECIL_FULL}
      alt=""
      draggable={false}
    />
  )
}

function BlackScreenStep({ step, text }: { step: StoryStep; text: string }) {
  if (step.kind === 'title') {
    return (
      <div className="story-black-title">
        <h1 className="story-black-title-text">{text}</h1>
      </div>
    )
  }
  return (
    <p key={step.id} className="story-black-narration">
      {text}
    </p>
  )
}

function DialoguePanel({
  step,
  showChoices,
  onChoose,
  onAdvance,
  text,
  speaker,
  t: st,
}: {
  step: StoryStep
  showChoices: boolean
  onChoose: (next: string) => void
  onAdvance: () => void
  text: string
  speaker: string
  t: (key: string, fallback?: string) => string
}) {
  return (
    <div className={`story-dialogue ${step.kind === 'system' ? 'story-dialogue-system' : ''}`} onClick={(event) => event.stopPropagation()}>
      {speaker && <div className="story-speaker">{speaker}</div>}
      <p key={step.id} className="story-dialogue-text">{text}</p>
      {showChoices ? (
        <div className="story-choice-list">
          {step.choices?.map((choice, index) => (
            <button key={choice.label} type="button" className="story-choice" onClick={() => onChoose(choice.next)}>
              {st(`${step.id}${String.fromCharCode(65 + index)}`, choice.label)}
            </button>
          ))}
        </div>
      ) : (
        <button type="button" className="story-next-button" onClick={onAdvance}>
          {step.actionLabel ? st(step.id, step.actionLabel) : st('continue', 'Continue')}
        </button>
      )}
    </div>
  )
}
