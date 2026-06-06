import { useMemo, useState } from 'react'
import { useEffect } from 'react'
import { audio } from '@/audio/AudioManager'
import { useTranslation } from 'react-i18next'
import type { StoryDefinition, StoryStep } from './story.types'

const HARBOR_BG = '/assets/story/chapter-1/newhaven_harbor.png'
const CECIL_SHY = '/assets/story/characters/cecil_shy.png'
const CECIL_FORMAL = '/assets/story/characters/cecil_formal.png'
const CECIL_SMILE = '/assets/story/characters/cecil_smile.png'
const BOAT = '/assets/story/props/arrival_boat.png'
const CLOUD_1 = '/assets/story/effects/cloud_soft_1.png'
const CLOUD_2 = '/assets/story/effects/cloud_soft_2.png'

interface StoryPlayerProps {
  story: StoryDefinition
  initialStepId?: string
  onComplete: () => void
  onProgress: (stepId: string) => void
  onSkip: () => void
}

export function StoryPlayer({ story, initialStepId, onComplete, onProgress, onSkip }: StoryPlayerProps) {
  const { t } = useTranslation()
  const [currentStepId, setCurrentStepId] = useState(initialStepId ?? story.firstStepId)

  const stepMap = useMemo(() => {
    return new Map(story.steps.map((step) => [step.id, step]))
  }, [story.steps])
  const step = stepMap.get(currentStepId) ?? story.steps[0]
  const isBlack = step.screen === 'black'
  const showPortrait =
    step.portrait === 'shadow' ||
    step.portrait === 'cecilShy' ||
    step.portrait === 'cecilFormal' ||
    step.portrait === 'cecilSmile'
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
    if (showChoices) return
    if (step.next) {
      setCurrentStepId(step.next)
      onProgress(step.next)
    } else {
      onComplete()
    }
  }

  const choose = (next: string) => {
    audio.playSfx('ui_confirm', { volume: 0.4 })
    setCurrentStepId(next)
    onProgress(next)
  }

  const textDisplay = st(step.id, step.text)
  const speakerDisplay = st(`${step.id}Speaker`, step.speaker)

  return (
    <section
      className={`story-root ${isBlack ? 'story-root-black' : 'story-root-harbor'}`}
      aria-label={story.title}
      onClick={advance}
    >
      <button type="button" className="story-skip-button" onClick={(event) => { event.stopPropagation(); onSkip() }}>
        Skip story
      </button>
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
        <BlackScreenStep step={step} text={textDisplay} onAdvance={advance} />
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
  const src = getCecilPortraitSrc(step.portrait)
  return (
    <div className={`story-portrait-wrap ${isShadow ? 'story-portrait-shadow' : ''}`}>
      <img
        className={`story-portrait story-portrait-${step.portrait ?? 'none'}`}
        src={src}
        alt=""
        draggable={false}
      />
    </div>
  )
}

function getCecilPortraitSrc(portrait: StoryStep['portrait']) {
  if (portrait === 'cecilShy') return CECIL_SHY
  if (portrait === 'cecilSmile') return CECIL_SMILE
  return CECIL_FORMAL
}

function BlackScreenStep({ step, text, onAdvance }: { step: StoryStep; text: string; onAdvance: () => void }) {
  if (step.kind === 'title') {
    return (
      <div className="story-black-content">
        <h1 className="story-black-title-text">{text}</h1>
        <button type="button" className="story-black-next" onClick={(event) => { event.stopPropagation(); onAdvance() }}>
          Continue
        </button>
      </div>
    )
  }
  return (
    <div className="story-black-content">
      <p key={step.id} className="story-black-narration">{text}</p>
      <button type="button" className="story-black-next" onClick={(event) => { event.stopPropagation(); onAdvance() }}>
        Continue
      </button>
    </div>
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
