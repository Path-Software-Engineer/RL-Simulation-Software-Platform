export function useRunStream(
  runId: Ref<string | null>,
  operatorToken: Ref<string>,
  onConfirmedUpdate: () => Promise<void>
) {
  const config = useRuntimeConfig()
  const state = ref<'idle' | 'connecting' | 'connected' | 'reconnecting' | 'unavailable'>('idle')
  const attempt = ref(0)
  let socket: WebSocket | undefined
  let retryTimer: ReturnType<typeof setTimeout> | undefined
  let generation = 0

  function disconnect() {
    generation += 1
    if (retryTimer) clearTimeout(retryTimer)
    retryTimer = undefined
    socket?.close()
    socket = undefined
    state.value = 'idle'
  }

  async function resync() {
    try {
      await onConfirmedUpdate()
      if (socket?.readyState === WebSocket.OPEN) state.value = 'connected'
    } catch {
      state.value = 'unavailable'
    }
  }

  function connect() {
    disconnect()
    if (!runId.value || !operatorToken.value || !import.meta.client) return
    const connectionGeneration = generation
    state.value = attempt.value ? 'reconnecting' : 'connecting'
    const base = String(config.public.wsBase).replace(/\/$/, '')
    socket = new WebSocket(`${base}/ws/v1/runs/${runId.value}`, ['rl-run-v1', operatorToken.value])
    socket.addEventListener('open', () => {
      if (connectionGeneration !== generation) return
      attempt.value = 0
      state.value = 'connected'
      void resync()
    })
    socket.addEventListener('message', () => {
      if (connectionGeneration === generation) void resync()
    })
    socket.addEventListener('close', () => {
      if (connectionGeneration !== generation || !runId.value) return
      attempt.value += 1
      if (attempt.value > 5) {
        state.value = 'unavailable'
        return
      }
      state.value = 'reconnecting'
      retryTimer = setTimeout(connect, Math.min(1000 * 2 ** attempt.value, 12000))
    })
  }

  watch([runId, operatorToken], () => {
    attempt.value = 0
    connect()
  })
  onMounted(connect)
  onBeforeUnmount(disconnect)
  return { state, resync }
}
