// src/views/NuevaActividad.jsx
import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import api from '../api'
import './NuevaActividad.css'

const NuevaActividad = () => {
    const [titulo, setTitulo] = useState('')
    const [descripcion, setDescripcion] = useState('')
    const [cupoMax, setCupoMax] = useState(0)
    const [imagenUrl, setImagenUrl] = useState('')
    const [nombreCategoria, setNombreCategoria] = useState('')
    const [nombreInstructor, setNombreInstructor] = useState('')
    const [emailInstructor, setEmailInstructor] = useState('')
    const [especialidadInstructor, setEspecialidadInstructor] = useState('')
    const [activo, setActivo] = useState(true)

    // nuevos campos de horario
    const [diaSemana, setDiaSemana] = useState('Lun')
    const [horaInicio, setHoraInicio] = useState('08:00')
    const [duracionMin, setDuracionMin] = useState(60)

    const [error, setError] = useState(null)

    const navigate = useNavigate()

    const handleSubmit = async (e) => {
        e.preventDefault()
        const token = localStorage.getItem('token')
        try {
            await api.post('/activities', {
                title: titulo,
                descripcion,
                cupo_max: parseInt(cupoMax),
                imagen_url: imagenUrl,
                nombre_categoria: nombreCategoria,
                nombre_instructor: nombreInstructor,
                email_instructor: emailInstructor,
                especialidad_instructor: especialidadInstructor,
                activo,
                // añadimos los datos de horario
                horario: {
                    dia_semana: diaSemana,
                    hora_inicio: horaInicio,
                    duracion_min: parseInt(duracionMin)
                }
            }, {
                headers: {
                    Authorization: `Bearer ${token}`
                }
            })
            alert('¡Actividad creada con éxito!')
            navigate('/home')
        } catch (err) {
            console.error(err)
            setError(err.response?.data?.error || 'Error al crear la actividad')
        }
    }

    return (
        <div className="nueva-actividad-container">
            <h1>Crear Nueva Actividad</h1>
            {error && <p className="error" role="alert">{error}</p>}
            <form className="nueva-actividad-form" onSubmit={handleSubmit}>
                <label htmlFor="titulo">Título</label>
                <input
                    id="titulo"
                    type="text"
                    value={titulo}
                    onChange={e => setTitulo(e.target.value)}
                    required
                />

                <label htmlFor="descripcion">Descripción</label>
                <textarea
                    id="descripcion"
                    value={descripcion}
                    onChange={e => setDescripcion(e.target.value)}
                    required
                ></textarea>

                <label htmlFor="cupo">Cupo máximo</label>
                <input
                    id="cupo"
                    type="number"
                    value={cupoMax}
                    onChange={e => setCupoMax(e.target.value)}
                    required
                />

                <label htmlFor="imagen">Imagen (URL)</label>
                <input
                    id="imagen"
                    type="url"
                    value={imagenUrl}
                    onChange={e => setImagenUrl(e.target.value)}
                />

                <label htmlFor="categoria">Categoría</label>
                <input
                    id="categoria"
                    type="text"
                    value={nombreCategoria}
                    onChange={e => setNombreCategoria(e.target.value)}
                    required
                />

                <label htmlFor="instructor">Nombre del instructor</label>
                <input
                    id="instructor"
                    type="text"
                    value={nombreInstructor}
                    onChange={e => setNombreInstructor(e.target.value)}
                    required
                />

                <label htmlFor="email">Email del instructor</label>
                <input
                    id="email"
                    type="email"
                    value={emailInstructor}
                    onChange={e => setEmailInstructor(e.target.value)}
                />

                <label htmlFor="especialidad">Especialidad del instructor</label>
                <input
                    id="especialidad"
                    type="text"
                    value={especialidadInstructor}
                    onChange={e => setEspecialidadInstructor(e.target.value)}
                />

                <label htmlFor="dia">Día de la semana</label>
                <select
                    id="dia"
                    value={diaSemana}
                    onChange={e => setDiaSemana(e.target.value)}
                >
                    <option value="Lun">Lunes</option>
                    <option value="Mar">Martes</option>
                    <option value="Mie">Miércoles</option>
                    <option value="Jue">Jueves</option>
                    <option value="Vie">Viernes</option>
                    <option value="Sab">Sábado</option>
                    <option value="Dom">Domingo</option>
                </select>

                <label htmlFor="hora">Hora de inicio</label>
                <input
                    id="hora"
                    type="time"
                    value={horaInicio}
                    onChange={e => setHoraInicio(e.target.value)}
                    required
                />

                <label htmlFor="duracion">Duración (minutos)</label>
                <input
                    id="duracion"
                    type="number"
                    value={duracionMin}
                    onChange={e => setDuracionMin(e.target.value)}
                    required
                />

                <label>
                    Activa
                    <input
                        type="checkbox"
                        checked={activo}
                        onChange={e => setActivo(e.target.checked)}
                    />
                </label>

                <button type="submit">Guardar actividad</button>
            </form>
        </div>
    )
}

export default NuevaActividad
