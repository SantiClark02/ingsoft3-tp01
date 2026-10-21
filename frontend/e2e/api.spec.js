// Pruebas de INTEGRACIÓN (TP7): le hablan a la API de QA ya desplegada, con su
// MySQL de verdad detrás. Sin navegador y sin ningún doble en el medio.
import { expect, test } from '@playwright/test'

// La API tiene su propia dirección (NO es la del front). Sin barra al final.
const API = process.env.API_BASE_URL || 'http://localhost:8081'
const USUARIO = process.env.E2E_USER || 'admin'
const CLAVE = process.env.E2E_PASS || 'password123' // credenciales de prueba del seed (README)

async function loginAdmin (request) {
  const r = await request.post(`${API}/users/login`, { data: { username: USUARIO, password: CLAVE } })
  expect(r.status()).toBe(200)
  return (await r.json()).token
}

async function listar (request) {
  const r = await request.get(`${API}/activities`)
  expect(r.status()).toBe(200)
  return (await r.json()) || []
}

function actividad (titulo, extra = {}) {
  return {
    titulo,
    descripcion: 'creada por la prueba de integración',
    cupo_max: 10,
    nombre_categoria: 'Prueba',
    nombre_instructor: 'Prueba',
    activo: true,
    horario: { dia_semana: 'Lun', hora_inicio: '10:00', duracion_min: 60 },
    ...extra,
  }
}

test('el alta guarda en la base de verdad, y el borrado la saca', async ({ request }) => {
  const token = await loginAdmin(request)
  const auth = { Authorization: `Bearer ${token}` }
  const titulo = `api ${Date.now()}` // único: no choca con otra corrida

  const alta = await request.post(`${API}/activities`, { headers: auth, data: actividad(titulo) })
  expect(alta.status()).toBe(201) // la API dice «creada»

  const creada = (await listar(request)).find(a => a.title === titulo) // y la BASE la devuelve
  expect(creada).toBeTruthy()
  expect(creada.day).toBe('Lun') // pasó por la columna ENUM de MySQL y volvió igual

  const borrado = await request.delete(`${API}/activities/${creada.id}`, { headers: auth })
  expect(borrado.status()).toBe(200)

  const despues = await listar(request) // y el borrado, comprobado
  expect(despues.some(a => a.title === titulo)).toBe(false)
})

test('un título vacío lo rechaza la API, y no crea nada', async ({ request }) => {
  const token = await loginAdmin(request)
  const antes = await listar(request)

  const alta = await request.post(`${API}/activities`, {
    headers: { Authorization: `Bearer ${token}` },
    data: actividad('   '),
  })
  expect(alta.status()).toBe(400) // el contrato del error
  expect((await alta.json()).error).toContain('título')

  const despues = await listar(request)
  expect(despues.length).toBe(antes.length) // y de verdad no se creó nada
})

test('el login valida contra la base: credenciales buenas entran, malas no', async ({ request }) => {
  const ok = await request.post(`${API}/users/login`, { data: { username: USUARIO, password: CLAVE } })
  expect(ok.status()).toBe(200)
  const cuerpo = await ok.json()
  expect(cuerpo.token).toBeTruthy()
  expect(cuerpo.rol).toBe('administrador') // el rol sale de la tabla usuarios

  const mal = await request.post(`${API}/users/login`, { data: { username: USUARIO, password: 'clave-incorrecta' } })
  expect(mal.status()).toBe(401)
})
