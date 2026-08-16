<script setup lang="ts">
import type { Coordinate, Environment, Policy, Transition } from '~/types/api'
import { actionArrow, coordinateKey } from '~/utils/gridworld'

const props = defineProps<{ environment: Environment; policy: Policy; transition?: Transition }>()
const key = (point: Coordinate) => coordinateKey(point)
const obstacleKeys = computed(() => new Set(props.environment.obstacles.map(key)))
const agent = computed(() => props.transition?.nextState ?? props.environment.start)

function cellKind(row: number, column: number) {
  const cell = `${row},${column}`
  if (obstacleKeys.value.has(cell)) return 'obstacle'
  if (key(props.environment.goal) === cell) return 'goal'
  if (key(agent.value) === cell) return 'agent'
  if (key(props.environment.start) === cell) return 'start'
  return 'open'
}

function label(row: number, column: number) {
  const kind = cellKind(row, column)
  const action = props.policy.actionByState[`${row},${column}`]
  return `Row ${row + 1}, column ${column + 1}: ${kind}${action && action !== 'blocked' ? `, policy ${action}` : ''}`
}

function policyArrow(row: number, column: number) {
  const action = props.policy.actionByState[`${row},${column}`]
  return action && action !== 'blocked' ? actionArrow[action] ?? '' : ''
}
</script>

<template>
  <figure class="grid-figure">
    <div class="gridworld" :style="{ '--columns': environment.columns }" role="grid" :aria-label="`${environment.name}, ${environment.rows} by ${environment.columns}`">
      <template v-for="row in environment.rows" :key="row">
        <div v-for="column in environment.columns" :key="`${row}-${column}`" class="grid-cell" :data-kind="cellKind(row - 1, column - 1)" role="gridcell" tabindex="0" :aria-label="label(row - 1, column - 1)">
          <span v-if="cellKind(row - 1, column - 1) === 'agent'" class="agent-mark" aria-hidden="true" />
          <span v-else-if="cellKind(row - 1, column - 1) === 'goal'" class="goal-mark" aria-hidden="true">G</span>
          <span v-else-if="policyArrow(row - 1, column - 1)" class="policy-arrow" aria-hidden="true">{{ policyArrow(row - 1, column - 1) }}</span>
        </div>
      </template>
    </div>
    <figcaption class="legend"><span><i data-kind="agent" />Agent</span><span><i data-kind="goal" />Goal</span><span><i data-kind="obstacle" />Obstacle</span><span>Arrows: registered policy</span></figcaption>
  </figure>
</template>
