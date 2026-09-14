<script setup lang="ts">
import type { Transition } from '~/types/api'

const props = defineProps<{ transitions: Transition[]; autoplayKey?: number }>()
const emit = defineEmits<{ change: [transition: Transition | undefined] }>()
const index = ref(0)
const speed = ref(800)
const playing = ref(false)
const currentTransition = computed(() => props.transitions[index.value])
let timer: ReturnType<typeof setInterval> | undefined

function publish() { emit('change', currentTransition.value) }
function step(delta: number) {
  index.value = Math.max(0, Math.min(index.value + delta, Math.max(props.transitions.length - 1, 0)))
  publish()
}
function stop() {
  if (timer) clearInterval(timer)
  timer = undefined
  playing.value = false
}
function start() {
  stop()
  playing.value = true
  timer = setInterval(() => {
    if (index.value >= props.transitions.length - 1) return stop()
    step(1)
  }, speed.value)
}
function toggle() { if (playing.value) return stop(); start() }
watch(() => props.transitions, () => { index.value = 0; stop(); publish() })
watch(() => props.autoplayKey, (key) => { if (key) { index.value = 0; publish(); start() } })
watch(speed, () => { if (playing.value) { stop(); toggle() } })
onBeforeUnmount(stop)
</script>

<template>
  <section class="player" aria-labelledby="player-title">
    <div><p class="eyebrow">Episode player</p><h2 id="player-title">Inspect each decision.</h2></div>
    <div class="player-controls">
      <button class="icon-button" type="button" :disabled="!transitions.length || index === 0" aria-label="Previous transition" @click="step(-1)">←</button>
      <button class="button secondary" type="button" :disabled="transitions.length < 2" @click="toggle">{{ playing ? 'Pause playback' : 'Play episode' }}</button>
      <button class="icon-button" type="button" :disabled="!transitions.length || index >= transitions.length - 1" aria-label="Next transition" @click="step(1)">→</button>
      <label>Speed<select v-model.number="speed"><option :value="1200">0.75×</option><option :value="800">1×</option><option :value="400">2×</option></select></label>
    </div>
    <div v-if="currentTransition" class="transition-readout" aria-live="polite">
      <span>Step <strong>{{ index + 1 }}/{{ transitions.length }}</strong></span>
      <span>Action <strong>{{ currentTransition.action }}</strong></span>
      <span>Reward <strong>{{ currentTransition.reward.toFixed(2) }}</strong></span>
      <span>Terminal <strong>{{ currentTransition.terminated ? 'yes' : currentTransition.truncated ? 'truncated' : 'no' }}</strong></span>
    </div>
  </section>
</template>
