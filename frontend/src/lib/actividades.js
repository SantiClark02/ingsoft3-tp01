// src/lib/actividades.js
// Lógica pura del front, sacada de los componentes para poder testearla sin DOM.

// Filtra actividades por título, categoría, día u horario (sin distinguir mayúsculas).
export function filtrarActividades(actividades, termino) {
    const t = (termino || '').trim().toLowerCase()
    if (!t) return actividades
    return actividades.filter(act =>
        [act.title, act.category_name, act.day, act.start_time]
            .some(campo => (campo || '').toLowerCase().includes(t))
    )
}

// Inscribe al usuario en el primer horario de la actividad.
// El cliente HTTP entra por parámetro: la app le pasa axios, el test un doble.
export async function inscribirse(actividadId, usuarioId, cliente) {
    const { data } = await cliente.get(`/activities/${actividadId}`)
    const horarioId = data.schedule?.[0]?.id
    if (!horarioId || horarioId <= 0) {
        throw new Error('No hay horarios disponibles')
    }
    await cliente.post('/inscripciones', { usuario_id: Number(usuarioId), horario_id: horarioId })
    return horarioId
}
