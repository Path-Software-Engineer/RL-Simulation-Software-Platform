<script setup lang="ts">
import type { Environment, Episode, MetricSample, Policy, TrainingRun, Transition } from '~/types/api'

const api = useControlApi()
const environments = ref<Environment[]>([])
const policies = ref<Policy[]>([])
const selectedEnvironmentId = ref('')
const selectedPolicyId = ref('')
const run = ref<TrainingRun | null>(null)
const episodes = ref<Episode[]>([])
const transitions = ref<Transition[]>([])
const metrics = ref<MetricSample[]>([])
const currentTransition = ref<Transition>()
const phase = ref<'locked' | 'loading' | 'ready' | 'working' | 'error'>('locked')
const error = ref('')
const accessToken = ref('')
const feedbackCategory = ref('clear')
const feedbackNote = ref('')
const feedbackState = ref<'idle' | 'sending' | 'sent'>('idle')
let poller: ReturnType<typeof setInterval> | undefined

const environment = computed(() => environments.value.find(item => item.id === selectedEnvironmentId.value))
const policy = computed(() => policies.value.find(item => item.id === selectedPolicyId.value))
const latestEpisode = computed(() => episodes.value[0])
const activeRunId = computed(() => run.value?.id ?? null)
const isTerminal = computed(() => run.value ? ['succeeded', 'failed', 'cancelled'].includes(run.value.status) : false)
const canPause = computed(() => run.value?.status === 'running')
const canResume = computed(() => run.value?.status === 'paused')
const canCancel = computed(() => run.value && ['queued', 'running', 'pausing', 'paused'].includes(run.value.status))

const { state: streamState, resync: resyncStream } = useRunStream(activeRunId, api.operatorToken, refreshRun)

async function unlockWorkspace() {
  error.value = ''
  try {
    api.authenticate(accessToken.value)
    accessToken.value = ''
    await loadWorkspace()
    if (phase.value === 'error') {
      const message = error.value
      api.clearSession()
      phase.value = 'locked'
      error.value = message
    }
  } catch (reason) {
    api.clearSession()
    phase.value = 'locked'
    error.value = reason instanceof Error ? reason.message : 'Operator authentication failed.'
  }
}

function lockWorkspace() {
  api.clearSession()
  run.value = null
  environments.value = []
  policies.value = []
  episodes.value = []
  transitions.value = []
  metrics.value = []
  phase.value = 'locked'
  error.value = ''
}

async function loadWorkspace() {
  phase.value = 'loading'
  error.value = ''
  try {
    const [environmentItems, policyItems, runPage] = await Promise.all([
      api.listEnvironments(), api.listPolicies(), api.listRuns()
    ])
    environments.value = environmentItems
    policies.value = policyItems
    selectedEnvironmentId.value ||= environmentItems[0]?.id ?? ''
    selectedPolicyId.value ||= policyItems[0]?.id ?? ''
    run.value = runPage.items[0] ?? null
    if (run.value) {
      selectedEnvironmentId.value = run.value.environmentId
      selectedPolicyId.value = run.value.policyId
      await refreshRun()
    }
    phase.value = 'ready'
  } catch (reason) {
    error.value = reason instanceof Error ? reason.message : 'The platform could not load the workspace.'
    phase.value = 'error'
  }
}

async function refreshRun() {
  if (!run.value) return
  try {
    const confirmedRun = await api.getRun(run.value.id)
    run.value = confirmedRun
    const [episodeItems, metricItems] = await Promise.all([
      api.listEpisodes(confirmedRun.id), api.listMetrics(confirmedRun.id)
    ])
    episodes.value = episodeItems
    metrics.value = metricItems
    if (episodeItems[0]) {
      transitions.value = await api.listTransitions(episodeItems[0].id)
      currentTransition.value = transitions.value[0]
    }
  } catch (reason) {
    error.value = reason instanceof Error ? reason.message : 'Run evidence could not be synchronized.'
    throw reason
  }
}

async function startRun() {
  if (!environment.value || !policy.value) return
  phase.value = 'working'
  error.value = ''
  episodes.value = []
  transitions.value = []
  metrics.value = []
  try {
    run.value = await api.createRun(environment.value.id, policy.value.id)
    phase.value = 'ready'
    await refreshRun()
  } catch (reason) {
    error.value = reason instanceof Error ? reason.message : 'The run could not be started.'
    phase.value = 'error'
  }
}

async function control(action: 'pause' | 'resume' | 'cancel') {
  if (!run.value) return
  try {
    run.value = await api.controlRun(run.value.id, action)
    await refreshRun()
  } catch (reason) {
    error.value = reason instanceof Error ? reason.message : 'The control command was rejected.'
  }
}

async function submitFeedback() {
  if (!latestEpisode.value || !feedbackNote.value.trim()) return
  feedbackState.value = 'sending'
  try {
    await api.createFeedback(latestEpisode.value.id, feedbackCategory.value, feedbackNote.value.trim())
    feedbackState.value = 'sent'
    feedbackNote.value = ''
  } catch (reason) {
    error.value = reason instanceof Error ? reason.message : 'Feedback could not be recorded.'
    feedbackState.value = 'idle'
  }
}

onMounted(() => {
  if (api.restoreSession()) void loadWorkspace()
  poller = setInterval(() => {
    if (run.value && !isTerminal.value) void refreshRun().catch(() => undefined)
  }, 2000)
})
onBeforeUnmount(() => { if (poller) clearInterval(poller) })
</script>

<template>
  <div class="app-frame">
    <a class="skip-link" href="#main">Skip to workspace</a>
    <aside class="sidebar">
      <div class="brand"><span class="brand-mark" aria-hidden="true">V</span><div><strong>Vector</strong><small>RL evidence lab</small></div></div>
      <nav aria-label="Sprint navigation">
        <a href="#experiment" aria-current="page"><span>01</span>Experiment</a>
        <a href="#environment"><span>02</span>Environment</a>
        <a href="#episode"><span>03</span>Episode</a>
        <a href="#evidence"><span>04</span>Evidence</a>
      </nav>
      <div class="boundary-card"><p class="eyebrow">Evidence boundary</p><strong>Controlled tabular policy</strong><small>No arbitrary environments, uploads or Python execution.</small></div>
    </aside>

    <main id="main">
      <header class="topbar"><div><small>PROJECT 07 / SPRINT 01</small><strong>Gridworld Agent Visualizer</strong></div><div class="topbar-actions"><div class="topbar-state"><span :data-state="streamState" />{{ phase === 'locked' ? 'locked' : streamState }}</div><button v-if="phase !== 'locked'" class="text-button" type="button" @click="lockWorkspace">Lock</button></div></header>

      <section id="experiment" class="hero">
        <div><p class="eyebrow">Observed reinforcement learning</p><h1>Follow the policy.<br><span>Read every consequence.</span></h1><p class="hero-copy">Run a registered tabular agent in a controlled Gridworld, then inspect the exact state, action, reward and terminal signal persisted by the platform.</p></div>
        <div class="hero-orbit" aria-hidden="true"><i /><i /><b>Q</b></div>
      </section>

      <section v-if="error" class="notice error" role="alert"><strong>Vector could not complete the request.</strong><span>{{ error }}</span><button v-if="phase !== 'locked'" type="button" @click="loadWorkspace">Retry</button></section>
      <section v-if="phase === 'locked'" class="auth-panel" aria-labelledby="auth-title">
        <div><p class="eyebrow">Local operator boundary</p><h2 id="auth-title">Unlock the controlled workspace.</h2><p>Enter the operator token configured in <code>.env</code>. It stays in this browser tab and is never written to application logs.</p></div>
        <form @submit.prevent="unlockWorkspace"><label>Operator token<input v-model="accessToken" type="password" minlength="16" maxlength="128" pattern="[A-Za-z0-9._~-]+" autocomplete="current-password" required></label><button class="button primary" type="submit">Authenticate <span>→</span></button></form>
      </section>
      <section v-if="phase === 'loading'" class="loading-panel" aria-live="polite"><span class="loader" />Reading the registered environment and policies…</section>

      <template v-else-if="environment && policy">
        <section id="environment" class="control-panel">
          <div><p class="eyebrow">Controlled experiment</p><h2>Choose registered evidence.</h2><p>Every option is resolved by the Go API and backed by versioned database records.</p></div>
          <div class="field-grid">
            <label>Environment<select v-model="selectedEnvironmentId" :disabled="phase === 'working'"><option v-for="item in environments" :key="item.id" :value="item.id">{{ item.name }} · v{{ item.version }}</option></select></label>
            <label>Policy<select v-model="selectedPolicyId" :disabled="phase === 'working'"><option v-for="item in policies" :key="item.id" :value="item.id">{{ item.algorithm }} · v{{ item.version }}</option></select></label>
            <button class="button primary" type="button" :disabled="phase === 'working'" @click="startRun">{{ phase === 'working' ? 'Queueing run…' : 'Run controlled episode' }} <span>→</span></button>
          </div>
        </section>

        <section v-if="run" class="run-bar" aria-live="polite">
          <div><p class="eyebrow">Confirmed run</p><strong>{{ run.id.slice(0, 8) }}</strong><StatusPill :status="run.status" /></div>
          <div class="run-controls">
            <button v-if="canPause" class="button secondary" type="button" @click="control('pause')">Pause runner</button>
            <button v-if="canResume" class="button secondary" type="button" @click="control('resume')">Resume runner</button>
            <button v-if="canCancel" class="button danger" type="button" @click="control('cancel')">Cancel run</button>
            <button v-if="streamState === 'unavailable'" class="button secondary" type="button" @click="resyncStream">Resync with REST</button>
          </div>
        </section>

        <section id="episode" class="workspace-grid">
          <div class="grid-card"><div class="section-heading"><div><p class="eyebrow">Policy overlay</p><h2>{{ environment.name }}</h2></div><span>{{ environment.rows }}×{{ environment.columns }}</span></div><GridworldGrid :environment="environment" :policy="policy" :transition="currentTransition" /></div>
          <aside class="evidence-card">
            <p class="eyebrow">Run evidence</p>
            <h2 v-if="latestEpisode">{{ latestEpisode.terminalReason.replaceAll('_', ' ') }}</h2><h2 v-else>Awaiting a completed episode.</h2>
            <dl><div><dt>Total reward</dt><dd>{{ latestEpisode ? latestEpisode.totalReward.toFixed(2) : '—' }}</dd></div><div><dt>Transitions</dt><dd>{{ latestEpisode?.stepCount ?? '—' }}</dd></div><div><dt>Collisions</dt><dd>{{ latestEpisode?.collisions ?? '—' }}</dd></div><div><dt>Terminal</dt><dd>{{ latestEpisode?.status ?? '—' }}</dd></div></dl>
            <div class="artifact"><small>POLICY ARTIFACT</small><strong>{{ policy.algorithm }} / {{ policy.version }}</strong><code>{{ policy.sha256.slice(0, 16) }}…</code><p>{{ policy.provenance }}</p></div>
          </aside>
        </section>

        <EpisodePlayer v-if="transitions.length" :transitions="transitions" @change="currentTransition = $event" />

        <section id="evidence" class="evidence-grid">
          <article><p class="eyebrow">Reward semantics</p><h3>Reward is a teaching signal.</h3><p>Step {{ environment.rewardMap.step }}, collision {{ environment.rewardMap.collision }}, goal +{{ environment.rewardMap.goal }}. It is not a safety guarantee.</p></article>
          <article><p class="eyebrow">Temporal samples</p><h3>{{ metrics.length }} persisted metrics</h3><ul><li v-for="item in metrics" :key="`${item.metric}-${item.sampledAt}`"><span>{{ item.metric }}</span><strong>{{ item.value }} {{ item.unit }}</strong></li></ul></article>
          <form v-if="latestEpisode" @submit.prevent="submitFeedback"><p class="eyebrow">Episode annotation</p><h3>Record human feedback.</h3><label>Category<select v-model="feedbackCategory"><option value="clear">Clear</option><option value="unexpected">Unexpected</option><option value="loop">Loop</option><option value="collision">Collision</option><option value="other">Other</option></select></label><label>Note<textarea v-model="feedbackNote" maxlength="500" required placeholder="What did you observe?" /></label><button class="button secondary" type="submit" :disabled="feedbackState === 'sending'">{{ feedbackState === 'sent' ? 'Feedback recorded' : 'Save annotation' }}</button></form>
        </section>
      </template>
    </main>
  </div>
</template>
