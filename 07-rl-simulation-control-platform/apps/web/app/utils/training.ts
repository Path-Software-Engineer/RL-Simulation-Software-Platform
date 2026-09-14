import type { MetricSample } from '../types/api'

export interface ActionSummary {
  label: string
  name: string
  value: number
  percent: number
}

export function metricPoints(samples: MetricSample[], name: string): Array<[number, number]> {
  return samples
    .filter(item => item.metric === name)
    .sort((left, right) => left.step - right.step)
    .map(item => [item.step, item.value])
}

export function latestMetricSample(samples: MetricSample[], name: string) {
  return samples.filter(item => item.metric === name).sort((left, right) => left.step - right.step).at(-1)
}

export function summarizeActions(samples: MetricSample[]): ActionSummary[] {
  const definitions = [
    ['Up', 'action_up'], ['Right', 'action_right'], ['Down', 'action_down'], ['Left', 'action_left']
  ] as const
  const values = definitions.map(([label, name]) => ({
    label,
    name,
    value: samples.filter(item => item.metric === name).reduce((total, item) => total + item.value, 0)
  }))
  const total = values.reduce((sum, item) => sum + item.value, 0)
  return values.map(item => ({ ...item, percent: total ? (item.value / total) * 100 : 0 }))
}
