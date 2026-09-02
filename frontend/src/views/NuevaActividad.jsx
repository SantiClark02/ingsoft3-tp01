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
                titulo,
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
            {error && <p className="error">{error}</p>}
            <form className="nueva-actividad-form" onSubmit={handleSubmit}>
                <label>Título</label>
                <input
                    type="text"
                    value={titulo}
                    onChange={e => setTitulo(e.target.value)}
                    required
                />

                <label>Descripción</label>
                <textarea
                    value={descripcion}
                    onChange={e => setDescripcion(e.target.value)}
                    required
                ></textarea>

                <label>Cupo máximo</label>
                <input
                    type="number"
                    value={cupoMax}
                    onChange={e => setCupoMax(e.target.value)}
                    required
                />

                <label>Imagen (URL)</label>
                <input
                    type="url"
                    value={imagenUrl}
                    onChange={e => setImagenUrl(e.target.value)}
                />

                <label>Categoría</label>
                <input
                    type="text"
                    value={nombreCategoria}
                    onChange={e => setNombreCategoria(e.target.value)}
                    required
                />

                <label>Nombre del instructor</label>
                <input
                    type="text"
                    value={nombreInstructor}
                    onChange={e => setNombreInstructor(e.target.value)}
                    required
                />

                <label>Email del instructor</label>
                <input
                    type="email"
                    value={emailInstructor}
                    onChange={e => setEmailInstructor(e.target.value)}
                />

                <label>Especialidad del instructor</label>
                <input
                    type="text"
                    value={especialidadInstructor}
                    onChange={e => setEspecialidadInstructor(e.target.value)}
                />

                <label>Día de la semana</label>
                <select
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

                <label>Hora de inicio</label>
                <input
                    type="time"
                    value={horaInicio}
                    onChange={e => setHoraInicio(e.target.value)}
                    required
                />

                <label>Duración (minutos)</label>
                <input
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
