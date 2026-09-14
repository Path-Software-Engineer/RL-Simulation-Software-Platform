import { describe, expect, it } from 'vitest'

import type { Transition } from '../app/types/api'
import { classifyRolloutRisk, summarizeRollout } from '../app/utils/rollout'

function transition(stepIndex: number, stepError: number, accumulatedError: number): Transition {
  return {
    id: `00000000-0000-4000-8000-${String(stepIndex).padStart(12, '0')}`,
    episodeId: '70000000-0000-4000-8000-000000000009',
    stepIndex,
    state: { row: 5, column: 0 },
    action: 'right',
    nextState: { row: 5, column: 1 },
    reward: -0.04,
    terminated: false,
    truncated: false,
    sampledAt: '2026-08-16T00:00:00Z',
    predictedState: { row: 5, column: 0 },
    predictedNextState: { row: 5, column: 1 },
    stepError,
    accumulatedError,
    modelVersion: 'empirical-action-delta/1.0.0'
  }
}

describe('world-model rollout projections', () => {
  it('retains first divergence, maximum and final accumulated error', () => {
    const summary = summarizeRollout([
      transition(0, 0, 0),
      transition(1, 1, 1),
      transition(2, 1, 2)
    ])

    expect(summary).toEqual({
      firstDivergenceStep: 2,
      maximumStepError: 1,
      accumulatedError: 2
    })
  })

  it('classifies risk without hiding the numeric threshold', () => {
    expect(classifyRolloutRisk(undefined)).toBe('Awaiting evidence')
    expect(classifyRolloutRisk(0.2)).toBe('Low observed drift')
    expect(classifyRolloutRisk(0.5)).toBe('Guarded')
    expect(classifyRolloutRisk(0.8)).toBe('High drift')
  })
})
