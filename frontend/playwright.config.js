import { defineConfig } from '@playwright/test'

// Dos suites con la misma herramienta (TP7):
//  - e2e/api.spec.js         → integración: pedidos HTTP a la API de QA, sin navegador
//  - e2e/actividades.spec.js → e2e: un navegador real usando el front de QA
// Cada job del pipeline corre SU archivo por nombre.
export default defineConfig({
  testDir: './e2e',
  timeout: 60_000,             // tope de cada test: generoso (QA puede estar arrancando)
  expect: { timeout: 15_000 }, // tope de cada aserción (el default es 5 s)
  workers: 1,                  // QA es uno solo y compartido: un test por vez
  retries: 1,                  // absorbe una demora suelta; si pasa recién acá, sale «flaky»
  use: {
    baseURL: process.env.E2E_BASE_URL || 'http://localhost:3001', // el FRONT de QA
    trace: 'on-first-retry',          // traza del fallo (pestaña Network incluida)
    screenshot: 'only-on-failure',
  },
  reporter: [['html', { open: 'never' }], ['list'], ['json', { outputFile: 'playwright-resultados.json' }]],
})
