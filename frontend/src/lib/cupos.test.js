import { describe, expect, it } from 'vitest'
import { estadoCupo } from './cupos.js'

describe('estadoCupo', () => {
    // Un caso por cada camino que declara la función
    it.each([
        ['sin cupo definido', 0, 0, 'sin-cupo-definido'],
        ['completo', 20, 20, 'completo'],
        ['últimos lugares (quedan 3)', 17, 20, 'ultimos-lugares'],
        ['más de la mitad ocupado', 12, 20, 'mas-de-la-mitad-ocupado'],
        ['disponible', 5, 20, 'disponible'],
    ])('%s', (_caso, inscriptos, cupoMax, esperado) => {
        expect(estadoCupo(inscriptos, cupoMax)).toBe(esperado)
    })

    // Caso de error / borde: más inscriptos que cupo (dato inconsistente)
    it('si hay más inscriptos que cupo lo marca completo', () => {
        expect(estadoCupo(25, 20)).toBe('completo')
    })
})
