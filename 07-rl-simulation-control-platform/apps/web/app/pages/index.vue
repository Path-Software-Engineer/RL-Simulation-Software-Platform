<script setup lang="ts">
import type { Environment, Episode, MetricSample, Policy, TrainingRun, Transition } from '~/types/api'
import type { ActionSummary } from '~/utils/training'
import { classifyRolloutRisk, summarizeRollout } from '~/utils/rollout'
import { latestMetricSample as selectLatestMetric, metricPoints as selectMetricPoints, summarizeActions } from '~/utils/training'

const DQN_EPISODES = 40
const trainingMetricNames = [
  'episode_reward',
  'moving_average_reward',
  'epsilon',
  'loss',
  'episode_steps',
  'success_rate',
  'action_up',
  'action_right',
  'action_down',
  'action_left'
] as const
const worldMetricNames = [
  'training_examples',
  'prediction_error',
  'accumulated_error',
  'rollout_risk'
] as const

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
const runPolicy = computed(() => policies.value.find(item => item.id === run.value?.policyId))
const evidencePolicy = computed(() => runPolicy.value ?? policy.value)
const latestEpisode = computed(() => episodes.value.at(-1))
const activeRunId = computed(() => run.value?.id ?? null)
const isTerminal = computed(() => run.value ? ['succeeded', 'failed', 'cancelled'].includes(run.value.status) : false)
const hasActiveRun = computed(() => Boolean(run.value && !isTerminal.value))
const isDqnSelected = computed(() => policy.value?.algorithm === 'dqn')
const isDqnRun = computed(() => runPolicy.value?.algorithm === 'dqn')
const isWorldModelSelected = computed(() => policy.value?.algorithm === 'world-model')
const isWorldModelRun = computed(() => runPolicy.value?.algorithm === 'world-model')
const showTrainingDashboard = computed(() => isDqnRun.value || (!run.value && isDqnSelected.value))
const showWorldModelDashboard = computed(() => isWorldModelRun.value || (!run.value && isWorldModelSelected.value))
const canPause = computed(() => run.value?.status === 'running')
const canResume = computed(() => run.value?.status === 'paused')
const canCancel = computed(() => run.value && ['queued', 'running', 'pausing', 'paused'].includes(run.value.status))

const rewardSeries = computed(() => [
  { name: 'Episode reward', color: '#62e8b2', values: metricPoints('episode_reward') },
  { name: 'Moving average', color: '#6ea8fe', values: metricPoints('moving_average_reward') }
])
const epsilonSeries = computed(() => [
  { name: 'Epsilon', color: '#f6c85f', values: metricPoints('epsilon') }
])
const lossSeries = computed(() => [
  { name: 'MSE loss', color: '#f08ba6', values: metricPoints('loss') }
])
const rolloutErrorSeries = computed(() => [
  { name: 'Step error', color: '#f6c85f', values: metricPoints('prediction_error') },
  { name: 'Accumulated error', color: '#f08ba6', values: metricPoints('accumulated_error') }
])
const bestEpisode = computed(() => episodes.value.reduce<Episode | null>((best, item) => {
  if (!best || item.totalReward > best.totalReward) return item
  return best
}, null))
const latestMovingAverage = computed(() => latestMetric('moving_average_reward'))
const latestEpsilon = computed(() => latestMetric('epsilon'))
const latestLoss = computed(() => latestMetric('loss'))
const latestSuccessRate = computed(() => latestMetric('success_rate'))
const trainingProgress = computed(() => Math.min((episodes.value.length / DQN_EPISODES) * 100, 100))
const actionDistribution = computed(() => {
  return summarizeActions(metrics.value)
})
const dominantAction = computed(() => actionDistribution.value.reduce<ActionSummary | null>((best, item) => {
  if (!best || item.value > best.value) return item
  return best
}, null))
const worldTransitions = computed(() => transitions.value.filter(item => item.predictedNextState && item.stepError !== undefined))
const rolloutSummary = computed(() => summarizeRollout(worldTransitions.value))
const trainingExamples = computed(() => latestMetric('training_examples'))
const accumulatedError = computed(() => latestMetric('accumulated_error'))
const rolloutRisk = computed(() => latestMetric('rollout_risk'))
const riskLabel = computed(() => classifyRolloutRisk(rolloutRisk.value))
const compactMetrics = computed(() => {
  const names = isWorldModelRun.value ? worldMetricNames : trainingMetricNames
  return names.map(name => ({ name, sample: latestMetricSample(name) })).filter(item => item.sample)
})

const { state: streamState, resync: resyncStream } = useRunStream(activeRunId, api.operatorToken, refreshRun)

function metricPoints(name: string): Array<[number, number]> {
  return selectMetricPoints(metrics.value, name)
}

function latestMetricSample(name: string) {
  return selectLatestMetric(metrics.value, name)
}

function latestMetric(name: string) {
  return latestMetricSample(name)?.value
}

function displayMetric(value: number | undefined, digits = 3) {
  return value === undefined ? '—' : value.toFixed(digits)
}

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
    selectedPolicyId.value ||= policyItems.find(item => item.algorithm === 'world-model')?.id ?? policyItems[0]?.id ?? ''
    run.value = runPage.items[0] ?? null
    if (run.value) {
      selectedEnvironmentId.value = run.value.environmentId
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
    const activePolicy = policies.value.find(item => item.id === confirmedRun.policyId)
    const metricRequests = activePolicy?.algorithm === 'dqn'
      ? trainingMetricNames.map(name => api.listMetrics(confirmedRun.id, name))
      : [api.listMetrics(confirmedRun.id)]
    const [episodeItems, metricGroups] = await Promise.all([
      api.listEpisodes(confirmedRun.id), Promise.all(metricRequests)
    ])
    episodes.value = episodeItems
    metrics.value = metricGroups.flat().sort((left, right) => left.step - right.step || left.metric.localeCompare(right.metric))
    const newestEpisode = episodeItems.at(-1)
    if (newestEpisode) {
      const selectedStep = currentTransition.value?.stepIndex
      transitions.value = await api.listTransitions(newestEpisode.id)
      currentTransition.value = transitions.value.find(item => item.stepIndex === selectedStep) ?? transitions.value[0]
    } else {
      transitions.value = []
      currentTransition.value = undefined
    }
  } catch (reason) {
    error.value = reason instanceof Error ? reason.message : 'Run evidence could not be synchronized.'
    throw reason
  }
}

async function startRun() {
  if (!environment.value || !policy.value || hasActiveRun.value) return
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
  }, 1000)
})
onBeforeUnmount(() => { if (poller) clearInterval(poller) })
</script>

<template>
  <div class="app-frame">
    <a class="skip-link" href="#main">Skip to workspace</a>
    <aside class="sidebar">
      <div class="brand"><span class="brand-mark" aria-hidden="true">V</span><div><strong>Vector</strong><small>RL evidence lab</small></div></div>
      <nav aria-label="Sprint navigation">
        <a href="#experiment"><span>01</span>Experiment</a>
        <a href="#model" aria-current="page"><span>02</span>World model</a>
        <a href="#rollout"><span>03</span>Rollout</a>
        <a href="#evidence"><span>04</span>Evidence</a>
      </nav>
      <div class="boundary-card"><p class="eyebrow">Evidence boundary</p><strong>Bounded transition model</strong><small>Twelve registered examples and one fixed rollout plan. No uploads, hidden planning or arbitrary Python.</small></div>
    </aside>

    <main id="main">
      <header class="topbar"><div><small>PROJECT 07 / SPRINT 03</small><strong>World Model Rollout Viewer</strong></div><div class="topbar-actions"><div class="topbar-state"><span :data-state="streamState" />{{ phase === 'locked' ? 'locked' : streamState }}</div><button v-if="phase !== 'locked'" class="text-button" type="button" @click="lockWorkspace">Lock</button></div></header>

      <section id="experiment" class="hero model-hero">
        <div><p class="eyebrow">Model-based reinforcement learning observability</p><h1>See the imagined future.<br><span>Measure where it drifts.</span></h1><p class="hero-copy">Fit a registered empirical transition model, compare its autoregressive rollout with real Gridworld transitions and inspect error before trusting a plan.</p></div>
        <div class="hero-orbit" aria-hidden="true"><i /><i /><b>WM</b></div>
      </section>

      <section v-if="error" class="notice error" role="alert"><strong>Vector could not complete the request.</strong><span>{{ error }}</span><button v-if="phase !== 'locked'" type="button" @click="loadWorkspace">Retry</button></section>
      <section v-if="phase === 'locked'" class="auth-panel" aria-labelledby="auth-title">
        <div><p class="eyebrow">Local operator boundary</p><h2 id="auth-title">Unlock the controlled workspace.</h2><p>Enter the operator token configured in <code>.env</code>. It stays in this browser tab and is never written to application logs.</p></div>
        <form @submit.prevent="unlockWorkspace"><label>Operator token<input v-model="accessToken" type="password" minlength="16" maxlength="128" pattern="[A-Za-z0-9._~-]+" autocomplete="current-password" required></label><button class="button primary" type="submit">Authenticate <span>→</span></button></form>
      </section>
      <section v-if="phase === 'loading'" class="loading-panel" aria-live="polite"><span class="loader" />Reading registered environments, agents and training evidence…</section>

      <template v-else-if="environment && policy">
        <section class="control-panel">
          <div><p class="eyebrow">Controlled experiment</p><h2>Choose a registered execution profile.</h2><p>The world-model profile fits 12 transition examples and compares one fixed rollout with the real environment. DQN and tabular checkpoints remain available.</p></div>
          <div class="field-grid">
            <label>Environment<select v-model="selectedEnvironmentId" :disabled="phase === 'working' || hasActiveRun"><option v-for="item in environments" :key="item.id" :value="item.id">{{ item.name }} · v{{ item.version }}</option></select></label>
            <label>Agent profile<select v-model="selectedPolicyId" :disabled="phase === 'working' || hasActiveRun"><option v-for="item in policies" :key="item.id" :value="item.id">{{ item.algorithm }} · v{{ item.version }}</option></select></label>
            <button class="button primary" type="button" :disabled="phase === 'working' || hasActiveRun" @click="startRun">{{ phase === 'working' ? 'Queueing run…' : hasActiveRun ? 'Run in progress' : isWorldModelSelected ? 'Run model rollout' : isDqnSelected ? 'Start DQN training' : 'Run controlled episode' }} <span>→</span></button>
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

        <section v-if="showTrainingDashboard" id="training" class="training-dashboard" aria-labelledby="training-title">
          <div class="dashboard-heading"><div><p class="eyebrow">Persisted training telemetry</p><h2 id="training-title">DQN learning observatory</h2></div><div class="progress-copy"><strong>{{ episodes.length }} / {{ DQN_EPISODES }}</strong><span>episodes persisted</span></div></div>
          <div class="progress-track" role="progressbar" aria-label="DQN training progress" aria-valuemin="0" :aria-valuemax="DQN_EPISODES" :aria-valuenow="episodes.length"><span :style="{ width: `${trainingProgress}%` }" /></div>

          <div class="summary-grid">
            <article><small>Moving avg reward</small><strong>{{ displayMetric(latestMovingAverage, 2) }}</strong><span>last 10 episodes</span></article>
            <article><small>Best episode</small><strong>{{ bestEpisode ? bestEpisode.totalReward.toFixed(2) : '—' }}</strong><span>{{ bestEpisode ? `episode ${bestEpisode.episodeNumber}` : 'awaiting samples' }}</span></article>
            <article><small>Exploration ε</small><strong>{{ displayMetric(latestEpsilon) }}</strong><span>linear schedule</span></article>
            <article><small>Latest loss</small><strong>{{ displayMetric(latestLoss, 4) }}</strong><span>mean squared TD error</span></article>
            <article><small>Success rate</small><strong>{{ latestSuccessRate === undefined ? '—' : `${(latestSuccessRate * 100).toFixed(1)}%` }}</strong><span>goal reached / episodes</span></article>
            <article><small>Dominant action</small><strong>{{ dominantAction?.value ? dominantAction.label : '—' }}</strong><span>{{ dominantAction?.value ? `${dominantAction.percent.toFixed(1)}% of actions` : 'awaiting samples' }}</span></article>
          </div>

          <div class="chart-grid-layout">
            <article class="chart-card wide"><div><p class="eyebrow">Learning curve</p><h3>Reward by episode</h3></div><MetricChart title="Reward learning curve" y-label="reward" :series="rewardSeries" /></article>
            <article class="chart-card"><div><p class="eyebrow">Exploration schedule</p><h3>Epsilon decay</h3></div><MetricChart title="Epsilon schedule" y-label="epsilon" :series="epsilonSeries" /></article>
            <article class="chart-card"><div><p class="eyebrow">Optimization signal</p><h3>Training loss</h3></div><MetricChart title="DQN training loss" y-label="MSE" :series="lossSeries" /></article>
          </div>

          <article class="action-card">
            <div><p class="eyebrow">Action distribution</p><h3>Exploration and learned preference</h3><p>Counts aggregate every persisted DQN episode. Color is supported by labels and exact values.</p></div>
            <ul><li v-for="item in actionDistribution" :key="item.name"><span>{{ item.label }}</span><div><i :style="{ width: `${item.percent}%` }" /></div><strong>{{ item.value.toFixed(0) }} · {{ item.percent.toFixed(1) }}%</strong></li></ul>
          </article>
        </section>

        <section v-if="showWorldModelDashboard" id="model" class="world-model-dashboard" aria-labelledby="model-title">
          <div class="dashboard-heading"><div><p class="eyebrow">Persisted model evidence</p><h2 id="model-title">Expected state versus imagined state</h2></div><div class="progress-copy"><strong>{{ worldTransitions.length }} / 11</strong><span>rollout steps persisted</span></div></div>

          <div class="model-summary-grid">
            <article><small>Transition dataset</small><strong>{{ trainingExamples?.toFixed(0) ?? '—' }}</strong><span>versioned examples</span></article>
            <article><small>First divergence</small><strong>{{ rolloutSummary.firstDivergenceStep === undefined ? '—' : `step ${rolloutSummary.firstDivergenceStep}` }}</strong><span>expected ≠ predicted</span></article>
            <article><small>Accumulated error</small><strong>{{ displayMetric(accumulatedError, 0) }}</strong><span>Manhattan cell distance</span></article>
            <article><small>Planning risk</small><strong>{{ riskLabel }}</strong><span>{{ displayMetric(rolloutRisk, 3) }} mean error / step</span></article>
          </div>

          <div class="model-inspection-grid">
            <article class="comparison-card"><div><p class="eyebrow">Next-state prediction</p><h3>Step {{ currentTransition ? currentTransition.stepIndex + 1 : '—' }} · {{ currentTransition?.action ?? 'awaiting rollout' }}</h3><p>Expected state comes from the real environment. Predicted state comes from the persisted empirical rollout.</p></div><WorldModelComparison :environment="environment" :transition="currentTransition" /></article>
            <article class="chart-card"><div><p class="eyebrow">Error propagation</p><h3>Drift by rollout step</h3></div><MetricChart title="World-model error by rollout step" x-label="rollout step" y-label="cell error" :series="rolloutErrorSeries" /></article>
          </div>

          <div class="risk-grid">
            <article><small>Observed obstacle collisions</small><strong>{{ latestEpisode?.collisions ?? '—' }}</strong><p>The model has no obstacle representation, so collision behavior is not learned.</p></article>
            <article><small>Maximum step error</small><strong>{{ rolloutSummary.maximumStepError.toFixed(0) }} cells</strong><p>Exact Manhattan distance between the predicted and expected next states.</p></article>
            <article><small>Planning limitation</small><strong>Open-loop drift</strong><p>Each prediction feeds the next prediction. Error can persist even when the real path recovers.</p></article>
          </div>
        </section>

        <RolloutViewer v-if="isWorldModelRun && worldTransitions.length" id="rollout" :transitions="worldTransitions" @change="currentTransition = $event" />

        <section v-if="evidencePolicy && !isWorldModelRun" id="episode" class="workspace-grid">
          <div class="grid-card"><div class="section-heading"><div><p class="eyebrow">Latest episode trace</p><h2>{{ environment.name }}</h2></div><span>{{ environment.rows }}×{{ environment.columns }}</span></div><GridworldGrid :environment="environment" :policy="evidencePolicy" :transition="currentTransition" /></div>
          <aside class="evidence-card">
            <p class="eyebrow">Episode evidence</p>
            <h2 v-if="latestEpisode">{{ latestEpisode.terminalReason.replaceAll('_', ' ') }}</h2><h2 v-else>Awaiting a persisted episode.</h2>
            <dl><div><dt>Total reward</dt><dd>{{ latestEpisode ? latestEpisode.totalReward.toFixed(2) : '—' }}</dd></div><div><dt>Transitions</dt><dd>{{ latestEpisode?.stepCount ?? '—' }}</dd></div><div><dt>Collisions</dt><dd>{{ latestEpisode?.collisions ?? '—' }}</dd></div><div><dt>Episode</dt><dd>{{ latestEpisode?.episodeNumber ?? '—' }}</dd></div></dl>
            <div class="artifact"><small>REGISTERED AGENT ARTIFACT</small><strong>{{ evidencePolicy.algorithm }} / {{ evidencePolicy.version }}</strong><code>{{ evidencePolicy.sha256.slice(0, 16) }}…</code><p>{{ evidencePolicy.provenance }}</p></div>
          </aside>
        </section>

        <EpisodePlayer v-if="transitions.length && !isWorldModelRun" :transitions="transitions" @change="currentTransition = $event" />

        <section id="evidence" class="evidence-grid">
          <article><p class="eyebrow">Interpretation boundary</p><h3>{{ isWorldModelRun ? 'A rollout is not a safe plan.' : 'Deep does not mean stable.' }}</h3><p v-if="isWorldModelRun">The model omits obstacle structure and is evaluated on one registered action sequence. Completion does not establish planning reliability, generalization or safety.</p><p v-else>Reward and loss describe this seeded teaching run. They do not establish convergence, generalization, robustness or real-world safety.</p></article>
          <article><p class="eyebrow">Latest persisted samples</p><h3>{{ metrics.length }} metric points loaded</h3><ul><li v-for="item in compactMetrics" :key="item.name"><span>{{ item.name.replaceAll('_', ' ') }}</span><strong>{{ item.sample?.value.toFixed(3) }} {{ item.sample?.unit }}</strong></li></ul></article>
          <form v-if="latestEpisode" @submit.prevent="submitFeedback"><p class="eyebrow">Episode annotation</p><h3>Record human feedback.</h3><label>Category<select v-model="feedbackCategory"><option value="clear">Clear</option><option value="unexpected">Unexpected</option><option value="loop">Loop</option><option value="collision">Collision</option><option value="other">Other</option></select></label><label>Note<textarea v-model="feedbackNote" maxlength="500" required placeholder="What did you observe?" /></label><button class="button secondary" type="submit" :disabled="feedbackState === 'sending'">{{ feedbackState === 'sent' ? 'Feedback recorded' : 'Save annotation' }}</button></form>
        </section>
      </template>
    </main>
  </div>
</template>
