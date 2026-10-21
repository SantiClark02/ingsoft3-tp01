import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { configDefaults } from 'vitest/config'

export default defineConfig({
  plugins: [react()],
  server: {
    // Replica en desarrollo lo que nginx hace en producción,
    // para que /api funcione igual con y sin Docker.
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
  },
  test: {
    exclude: [...configDefaults.exclude, 'e2e/**'], // los specs de Playwright no son de vitest
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html', 'json-summary'],
      include: ['src/lib/**'],          // sólo la lógica: los componentes se verifican e2e (TP7)
      exclude: ['src/lib/**/*.test.js'],
      thresholds: { lines: 80, branches: 80 },
    },
  },
})
