// Pruebas END-TO-END (TP7): un navegador real usa el FRONT de QA como un usuario.
// Front servido por nginx + API + MySQL de verdad, sin mocks. Cada test limpia lo que crea.
import { expect, test } from '@playwright/test'

const USUARIO = process.env.E2E_USER || 'admin'
const CLAVE = process.env.E2E_PASS || 'password123' // credenciales de prueba del seed (README)

test.beforeEach(async ({ page }) => {
  // la app usa alert()/confirm(): los aceptamos para que no frenen el flujo
  page.on('dialog', d => d.accept())
  await page.goto('/')
  await page.getByLabel('Usuario').fill(USUARIO)
  await page.getByLabel('Contraseña').fill(CLAVE)
  await page.getByRole('button', { name: 'Ingresar' }).click()
  await expect(page.getByRole('heading', { name: /Actividades Disponibles/ })).toBeVisible()
})

async function completarActividad (page, titulo, cupo) {
  await page.getByRole('button', { name: 'Crear actividad' }).click()
  await page.getByLabel('Título').fill(titulo)
  await page.getByLabel('Descripción').fill('creada por la prueba e2e')
  await page.getByLabel('Cupo máximo').fill(String(cupo))
  await page.getByLabel('Categoría').fill('Prueba')
  await page.getByLabel('Nombre del instructor').fill('Prueba')
  await page.getByRole('button', { name: 'Guardar actividad' }).click()
}

async function borrarDesdeAdmin (page, titulo) {
  await page.goto('/editarEliminarActividad')
  await page.getByRole('button', { name: `Eliminar ${titulo}` }).click()
  await expect(page.getByText(titulo)).toHaveCount(0)
}

test('crear una actividad la muestra en el listado, y eliminarla la saca', async ({ page }) => {
  const titulo = `e2e ${Date.now()}` // único: no choca con otra corrida

  await completarActividad(page, titulo, 10)

  await expect(page).toHaveURL(/\/home$/)
  await expect(page.getByRole('heading', { name: titulo })).toBeVisible() // el dato que ESTE test creó

  await borrarDesdeAdmin(page, titulo)
  await page.goto('/home')
  await expect(page.getByText(titulo)).toHaveCount(0) // y el borrado, comprobado
})

test('un cupo inválido muestra el error y no crea nada', async ({ page }) => {
  const titulo = `e2e ${Date.now()} cupo cero`

  await completarActividad(page, titulo, 0)

  await expect(page.getByRole('alert')).toContainText('cupo') // el usuario ve el error…
  await page.goto('/home')
  await expect(page.getByRole('heading', { name: /Actividades Disponibles/ })).toBeVisible()
  await expect(page.getByText(titulo)).toHaveCount(0) // …y en el listado no apareció nada
})

test('un socio se inscribe, la ve en Mis actividades, y se desinscribe', async ({ page }) => {
  const titulo = `e2e ${Date.now()} inscripcion`
  await completarActividad(page, titulo, 10) // la actividad propia de este test
  await expect(page.getByRole('heading', { name: titulo })).toBeVisible()

  await page.getByRole('button', { name: `Inscribirse a ${titulo}` }).click()
  await expect(page.getByRole('button', { name: `Desinscribirse de ${titulo}` })).toBeVisible()

  await page.getByRole('button', { name: 'Mis actividades' }).click()
  await expect(page.getByRole('heading', { name: titulo })).toBeVisible() // aparece en SUS actividades

  await page.getByRole('button', { name: `Desinscribirse de ${titulo}` }).click()
  await expect(page.getByText(titulo)).toHaveCount(0) // y la desinscripción, comprobada

  await borrarDesdeAdmin(page, titulo) // limpia la actividad que creó
})
