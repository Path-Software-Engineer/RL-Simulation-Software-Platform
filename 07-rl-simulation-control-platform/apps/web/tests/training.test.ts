import { describe, expect, it } from 'vitest'

import type { MetricSample } from '../app/types/api'
import { latestMetricSample, metricPoints, summarizeActions } from '../app/utils/training'

function sample(metric: string, value: number, step: number): MetricSample {
  return {
    runId: '33333333-3333-4333-8333-333333333302',
    metric,
    value,
    unit: 'count',
    step,
    sampledAt: `2026-08-16T00:00:${String(step).padStart(2, '0')}Z`
  }
}

describe('DQN dashboard metric projections', () => {
  it('orders chart points and resolves the latest sample by episode', () => {
    const samples = [sample('epsilon', 0.5, 2), sample('epsilon', 1, 1)]

    expect(metricPoints(samples, 'epsilon')).toEqual([[1, 1], [2, 0.5]])
    expect(latestMetricSample(samples, 'epsilon')?.value).toBe(0.5)
  })

  it('aggregates the persisted action distribution without hiding exact counts', () => {
    const samples = [
      sample('action_up', 2, 1),
      sample('action_up', 3, 2),
      sample('action_right', 5, 1),
      sample('action_down', 0, 1),
      sample('action_left', 0, 1)
    ]

    expect(summarizeActions(samples)).toEqual([
      { label: 'Up', name: 'action_up', value: 5, percent: 50 },
      { label: 'Right', name: 'action_right', value: 5, percent: 50 },
      { label: 'Down', name: 'action_down', value: 0, percent: 0 },
      { label: 'Left', name: 'action_left', value: 0, percent: 0 }
    ])
  })
})
