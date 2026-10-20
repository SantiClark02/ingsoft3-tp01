// src/lib/cupos.js
// Estado que se muestra según cuántos lugares quedan en una actividad.
export function estadoCupo(inscriptos, cupoMax) {
    if (!cupoMax || cupoMax <= 0) {
        return 'sin-cupo-definido'
    }
    const libres = cupoMax - inscriptos
    if (libres <= 0) {
        return 'completo'
    }
    if (libres <= 3) {
        return 'ultimos-lugares'
    }
    if (libres / cupoMax < 0.5) {
        return 'mas-de-la-mitad-ocupado'
    }
    return 'disponible'
}
