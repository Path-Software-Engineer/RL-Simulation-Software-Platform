import type { Environment, Episode, MetricSample, Policy, ProblemDetails, TrainingRun, Transition } from '~/types/api'

export function useControlApi() {
  const config = useRuntimeConfig()
  const base = String(config.public.apiBase).replace(/\/$/, '')
  const operatorToken = useState<string>('operator-token', () => '')

  function authenticate(value: string) {
    const normalized = value.trim()
    if (!/^[A-Za-z0-9._~-]{16,128}$/.test(normalized)) {
      throw new Error('The operator token must contain 16-128 protocol-safe characters.')
    }
    operatorToken.value = normalized
    if (import.meta.client) sessionStorage.setItem('rl-operator-token', normalized)
  }

  function restoreSession() {
    if (!import.meta.client) return false
    const stored = sessionStorage.getItem('rl-operator-token') ?? ''
    if (!stored) return false
    try {
      authenticate(stored)
      return true
    } catch {
      clearSession()
      return false
    }
  }

  function clearSession() {
    operatorToken.value = ''
    if (import.meta.client) sessionStorage.removeItem('rl-operator-token')
  }

  async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
    if (!operatorToken.value) throw new Error('Operator authentication is required.')
    const response = await fetch(`${base}${path}`, {
      ...options,
      headers: {
        Accept: 'application/json',
        ...options.headers,
        Authorization: `Bearer ${operatorToken.value}`
      }
    })
    if (!response.ok) {
      const fallback: ProblemDetails = { title: `Request failed with ${response.status}`, status: response.status }
      const problem = await response.json().catch(() => fallback) as ProblemDetails
      throw new Error(problem.detail || problem.title)
    }
    if (response.status === 204) return undefined as T
    return await response.json() as T
  }

  return {
    operatorToken,
    authenticate,
    restoreSession,
    clearSession,
    listEnvironments: () => request<Environment[]>('/api/v1/environments'),
    listPolicies: () => request<Policy[]>('/api/v1/policies'),
    listRuns: () => request<{ items: TrainingRun[] }>('/api/v1/training-runs?limit=20'),
    createRun: (environmentId: string, policyId: string) => request<TrainingRun>('/api/v1/training-runs', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Idempotency-Key': crypto.randomUUID() },
      body: JSON.stringify({ environmentId, policyId, seed: 7, maxSteps: 64 })
    }),
    getRun: (id: string) => request<TrainingRun>(`/api/v1/training-runs/${id}`),
    controlRun: (id: string, action: 'pause' | 'resume' | 'cancel') => request<TrainingRun>(`/api/v1/training-runs/${id}/${action}`, {
      method: 'POST', headers: { 'Idempotency-Key': crypto.randomUUID() }
    }),
    listEpisodes: (id: string) => request<Episode[]>(`/api/v1/training-runs/${id}/episodes?limit=100`),
    listMetrics: (id: string, metric = '') => request<MetricSample[]>(`/api/v1/training-runs/${id}/metrics?limit=200${metric ? `&metric=${encodeURIComponent(metric)}` : ''}`),
    listTransitions: (id: string) => request<Transition[]>(`/api/v1/episodes/${id}/transitions?limit=200`),
    createFeedback: (id: string, category: string, note: string) => request<void>(`/api/v1/episodes/${id}/feedback`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Idempotency-Key': crypto.randomUUID() },
      body: JSON.stringify({ category, note })
    })
  }
}
