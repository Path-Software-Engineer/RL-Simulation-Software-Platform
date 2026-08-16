export type Coordinate = { row: number; column: number }
export type RunStatus = 'draft' | 'queued' | 'running' | 'pausing' | 'paused' | 'cancelling' | 'cancelled' | 'succeeded' | 'failed'

export interface Environment {
  id: string
  slug: string
  name: string
  version: string
  rows: number
  columns: number
  start: Coordinate
  goal: Coordinate
  obstacles: Coordinate[]
  rewardMap: { step: number; collision: number; goal: number }
  observationSpace: string
  actionSpace: string[]
}

export interface Policy {
  id: string
  algorithm: 'q-learning' | 'sarsa' | 'dqn' | 'world-model'
  version: string
  artifactUri: string
  sha256: string
  provenance: string
  observationSpace: string
  actionSpace: string[]
  actionByState: Record<string, string>
  enabled: boolean
}

export interface TrainingRun {
  id: string
  environmentId: string
  policyId: string
  status: RunStatus
  attempt: number
  seed: number
  maxSteps: number
  terminalReason: string | null
  createdAt: string
  updatedAt: string
  startedAt?: string
  completedAt?: string
}

export interface Episode {
  id: string
  runId: string
  episodeNumber: number
  status: string
  totalReward: number
  stepCount: number
  collisions: number
  terminalReason: string
  startedAt: string
  completedAt: string
}

export interface Transition {
  id: string
  episodeId: string
  stepIndex: number
  state: Coordinate
  action: string
  nextState: Coordinate
  reward: number
  terminated: boolean
  truncated: boolean
  sampledAt: string
  predictedState?: Coordinate
  predictedNextState?: Coordinate
  stepError?: number
  accumulatedError?: number
  modelVersion?: string
}

export interface MetricSample {
  runId: string
  metric: string
  value: number
  unit: string
  step: number
  sampledAt: string
}

export interface ProblemDetails {
  title: string
  status: number
  detail?: string
  code?: string
  traceId?: string
}
