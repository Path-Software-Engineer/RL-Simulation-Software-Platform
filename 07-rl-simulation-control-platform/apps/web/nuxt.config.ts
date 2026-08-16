export default defineNuxtConfig({
  compatibilityDate: '2026-07-01',
  devtools: { enabled: false },
  css: ['~/assets/css/main.css'],
  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE ?? 'http://127.0.0.1:8080',
      wsBase: process.env.NUXT_PUBLIC_WS_BASE ?? 'ws://127.0.0.1:8080'
    }
  },
  typescript: { strict: true, typeCheck: true },
  app: {
    head: {
      title: 'Vector — Gridworld Agent Visualizer',
      meta: [
        { name: 'description', content: 'Inspect a real, controlled tabular reinforcement-learning episode.' },
        { name: 'theme-color', content: '#050914' }
      ]
    }
  }
})
