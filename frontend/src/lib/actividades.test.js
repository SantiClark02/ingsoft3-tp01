import { describe, expect, it, vi } from 'vitest'
import { filtrarActividades, inscribirse } from './actividades.js'

const actividades = [
    { title: 'Zumba', category_name: 'Baile', day: 'Lun', start_time: '18:00' },
    { title: 'Spinning', category_name: 'Cardio', day: 'Mar', start_time: '07:30' },
]

describe('filtrarActividades', () => {
    // Parametrizado: el mismo comportamiento, buscando por cada campo
    it.each([
        ['título', 'zumba', 'Zumba'],
        ['categoría', 'CARDIO', 'Spinning'],
        ['día', 'lun', 'Zumba'],
        ['horario', '07:30', 'Spinning'],
    ])('encuentra por %s', (_campo, termino, esperado) => {
        const resultado = filtrarActividades(actividades, termino)

        expect(resultado.map(a => a.title)).toEqual([esperado])
    })

    it('con el término vacío devuelve todas', () => {
        expect(filtrarActividades(actividades, '   ')).toHaveLength(2)
    })

    it('si nada coincide devuelve una lista vacía', () => {
        expect(filtrarActividades(actividades, 'yoga')).toEqual([])
    })
})

describe('inscribirse', () => {
    it('inscribe en el primer horario de la actividad', async () => {
        // Arrange: el cliente HTTP es un doble, no sale a la red
        const cliente = {
            get: vi.fn().mockResolvedValue({ data: { schedule: [{ id: 5 }, { id: 9 }] } }),
            post: vi.fn().mockResolvedValue({}),
        }

        // Act
        await inscribirse(3, '7', cliente)

        // Assert: verifica la INTERACCIÓN (qué se le pidió a la API)
        expect(cliente.get).toHaveBeenCalledWith('/activities/3')
        expect(cliente.post).toHaveBeenCalledWith('/inscripciones', { usuario_id: 7, horario_id: 5 })
    })

    // Caso de error: actividad sin horarios
    it('sin horarios rechaza y no llama al POST', async () => {
        const cliente = {
            get: vi.fn().mockResolvedValue({ data: { schedule: [] } }),
            post: vi.fn(),
        }

        await expect(inscribirse(3, 7, cliente)).rejects.toThrow('No hay horarios disponibles')
        expect(cliente.post).not.toHaveBeenCalled()
    })
})
