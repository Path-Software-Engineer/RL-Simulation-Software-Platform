import type { Transition } from '../types/api'

export function summarizeRollout(transitions: Transition[]) {
  const first = transitions.find(item => (item.stepError ?? 0) > 0)
  return {
    firstDivergenceStep: first ? first.stepIndex + 1 : undefined,
    maximumStepError: transitions.reduce(
      (maximum, item) => Math.max(maximum, item.stepError ?? 0),
      0
    ),
    accumulatedError: transitions.at(-1)?.accumulatedError
  }
}

export function classifyRolloutRisk(value: number | undefined) {
  if (value === undefined) return 'Awaiting evidence'
  if (value >= 0.75) return 'High drift'
  if (value >= 0.25) return 'Guarded'
  return 'Low observed drift'
}
