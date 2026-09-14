<script setup lang="ts">
interface ChartSeries {
  name: string
  color: string
  values: Array<[number, number]>
}

const props = withDefaults(defineProps<{
  title: string
  yLabel: string
  series: ChartSeries[]
  xLabel?: string
}>(), { xLabel: 'episode' })

const width = 720
const height = 250
const padding = { left: 54, right: 18, top: 18, bottom: 38 }
const allPoints = computed(() => props.series.flatMap(item => item.values))
const hasData = computed(() => allPoints.value.length > 0)
const xMin = computed(() => Math.min(...allPoints.value.map(point => point[0])))
const xMax = computed(() => Math.max(...allPoints.value.map(point => point[0])))
const yMin = computed(() => Math.min(...allPoints.value.map(point => point[1])))
const yMax = computed(() => Math.max(...allPoints.value.map(point => point[1])))

function scale(value: number, minimum: number, maximum: number, start: number, end: number) {
  if (maximum === minimum) return (start + end) / 2
  return start + ((value - minimum) / (maximum - minimum)) * (end - start)
}

function polyline(values: Array<[number, number]>) {
  return values.map(([step, value]) => {
    const x = scale(step, xMin.value, xMax.value, padding.left, width - padding.right)
    const y = scale(value, yMin.value, yMax.value, height - padding.bottom, padding.top)
    return `${x.toFixed(1)},${y.toFixed(1)}`
  }).join(' ')
}

function format(value: number) {
  if (Math.abs(value) >= 100) return value.toFixed(0)
  if (Math.abs(value) >= 10) return value.toFixed(1)
  return value.toFixed(3)
}
</script>

<template>
  <div v-if="hasData" class="metric-chart" role="group" :aria-label="title">
    <div class="chart-legend" aria-hidden="true">
      <span v-for="item in series" :key="item.name"><i :style="{ backgroundColor: item.color }" />{{ item.name }}</span>
    </div>
    <svg :viewBox="`0 0 ${width} ${height}`" role="img" :aria-label="`${title}. ${yLabel} by ${xLabel}.`">
      <g class="chart-grid">
        <line v-for="index in 5" :key="index" :x1="padding.left" :x2="width - padding.right" :y1="padding.top + (index - 1) * ((height - padding.top - padding.bottom) / 4)" :y2="padding.top + (index - 1) * ((height - padding.top - padding.bottom) / 4)" />
      </g>
      <text class="axis-label" :x="padding.left" :y="height - 8">{{ xLabel }} {{ xMin }}</text>
      <text class="axis-label" text-anchor="end" :x="width - padding.right" :y="height - 8">{{ xLabel }} {{ xMax }}</text>
      <text class="axis-label" :x="8" :y="padding.top + 4">{{ format(yMax) }}</text>
      <text class="axis-label" :x="8" :y="height - padding.bottom">{{ format(yMin) }}</text>
      <polyline v-for="item in series" :key="item.name" :points="polyline(item.values)" :stroke="item.color" fill="none" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" />
    </svg>
    <details>
      <summary>View chart data table</summary>
      <div class="table-scroll">
        <table>
          <thead><tr><th>Series</th><th>{{ xLabel }}</th><th>{{ yLabel }}</th></tr></thead>
          <tbody>
            <template v-for="item in series" :key="item.name">
              <tr v-for="point in item.values" :key="`${item.name}-${point[0]}`"><td>{{ item.name }}</td><td>{{ point[0] }}</td><td>{{ format(point[1]) }}</td></tr>
            </template>
          </tbody>
        </table>
      </div>
    </details>
  </div>
  <div v-else class="chart-empty" role="status">No training samples have been persisted for this chart yet.</div>
</template>
