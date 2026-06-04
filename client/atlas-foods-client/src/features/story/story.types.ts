export type StoryScreen = 'black' | 'harbor'
export type StoryPortrait = 'none' | 'shadow' | 'cecilShy' | 'cecilFormal' | 'cecilSmile'
export type StoryBoat = 'none' | 'arriving' | 'docked'
export type StoryStepKind = 'title' | 'narration' | 'dialogue' | 'choice' | 'system'

export interface StoryChoice {
  label: string
  next: string
}

export interface StoryStep {
  id: string
  kind: StoryStepKind
  screen: StoryScreen
  text: string
  next?: string
  speaker?: string
  portrait?: StoryPortrait
  boat?: StoryBoat
  locationTitle?: string
  choices?: StoryChoice[]
  actionLabel?: string
}

export interface StoryDefinition {
  id: string
  title: string
  firstStepId: string
  steps: StoryStep[]
}
