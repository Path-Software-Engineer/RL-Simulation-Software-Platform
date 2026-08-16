<script setup lang="ts">
import type { Coordinate, Environment, Transition } from '~/types/api'

const props = defineProps<{
  environment: Environment
  transition?: Transition
}>()

const cells = computed(() => Array.from(
  { length: props.environment.rows * props.environment.columns },
  (_, index) => ({ row: Math.floor(index / props.environment.columns), column: index % props.environment.columns })
))

function same(left: Coordinate | undefined, right: Coordinate) {
  return left?.row === right.row && left.column === right.column
}

function isObstacle(cell: Coordinate) {
  return props.environment.obstacles.some(item => same(item, cell))
}

function cellLabel(cell: Coordinate, marker: Coordinate | undefined, kind: string) {
  const details = [`row ${cell.row}`, `column ${cell.column}`]
  if (isObstacle(cell)) details.push('obstacle')
  if (same(props.environment.goal, cell)) details.push('goal')
  if (same(marker, cell)) details.push(kind)
  return details.join(', ')
}
</script>

<template>
  <div class="model-inputs"><span>Real input <strong v-if="transition">({{ transition.state.row }}, {{ transition.state.column }})</strong><strong v-else>—</strong></span><span>Model input <strong v-if="transition?.predictedState">({{ transition.predictedState.row }}, {{ transition.predictedState.column }})</strong><strong v-else>—</strong></span></div>
  <div class="model-comparison">
    <section>
      <div><small>Expected next state</small><strong v-if="transition">({{ transition.nextState.row }}, {{ transition.nextState.column }})</strong><strong v-else>—</strong></div>
      <ol class="model-grid" role="grid" aria-label="Expected next-state grid">
        <li v-for="cell in cells" :key="`actual-${cell.row}-${cell.column}`" role="gridcell" :aria-label="cellLabel(cell, transition?.nextState, 'expected state')" :class="{ obstacle: isObstacle(cell), goal: same(environment.goal, cell), actual: same(transition?.nextState, cell) }"><span v-if="same(transition?.nextState, cell)">E</span></li>
      </ol>
    </section>
    <section>
      <div><small>Predicted next state</small><strong v-if="transition?.predictedNextState">({{ transition.predictedNextState.row }}, {{ transition.predictedNextState.column }})</strong><strong v-else>—</strong></div>
      <ol class="model-grid" role="grid" aria-label="Predicted next-state grid">
        <li v-for="cell in cells" :key="`predicted-${cell.row}-${cell.column}`" role="gridcell" :aria-label="cellLabel(cell, transition?.predictedNextState, 'predicted state')" :class="{ obstacle: isObstacle(cell), goal: same(environment.goal, cell), predicted: same(transition?.predictedNextState, cell) }"><span v-if="same(transition?.predictedNextState, cell)">P</span></li>
      </ol>
    </section>
  </div>
</template>
