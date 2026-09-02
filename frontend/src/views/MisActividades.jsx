// src/views/MisActividades.jsx
import React, { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import api from '../api'
import './MisActividades.css'

const MisActividades = () => {
    const [activities, setActivities] = useState([])
    const [error, setError] = useState(null)
    const [loadingId, setLoadingId] = useState(null)
    const userId = localStorage.getItem('user_id')
    const navigate = useNavigate()

    // Cierra sesión y vuelve al login
    const handleLogout = () => {
        localStorage.removeItem('user_id')
        navigate('/')
    }

    // Vuelve al Home
    const handleBack = () => {
        navigate('/home')
    }

    // Carga todas las actividades y filtra sólo las suscritas
    const loadActivities = async () => {
        try {
            const resp = await api.get('/activities', {
                params: { user_id: userId }
            })
            // dejamos sólo las activities con subscribed === true
            const subscribed = resp.data.filter(act => act.subscribed)
            setActivities(subscribed)
        } catch {
            setError('No se pudieron cargar tus actividades')
        }
    }

    useEffect(() => {
        loadActivities()
    }, [])

    // Desinscribir usando el inscription_id que viene en cada activity
    const handleUnsubscribe = async inscId => {
        setLoadingId(inscId)
        try {
            await api.delete(`/inscripciones/${inscId}`)
            await loadActivities()
            alert('Te has desinscrito correctamente')
        } catch (err) {
            console.error(err)
            alert(err.response?.data?.error || 'Error al desinscribirse')
        } finally {
            setLoadingId(null)
        }
    }

    if (error) return <p className="error">{error}</p>

    return (
        <div className="misactividades-container">
            <div className="content-wrapper">
                {/* Barra superior */}
                <div
                    className="logout-bar"
                    style={{ justifyContent: 'space-between', marginBottom: '1rem' }}
                >
                    <button className="logout-button" onClick={handleLogout}>
                        Cerrar sesión
                    </button>
                    <button className="my-activities-button" onClick={handleBack}>
                        Volver a Home
                    </button>
                </div>

                <h1>Mis Actividades</h1>

                {activities.length === 0 ? (
                    <p className="mis-no-activities">
                        No estás inscripto en ninguna actividad.
                    </p>
                ) : (
                    <ul className="activities-list">
                        {activities.map(act => {
                            const next = act.schedule?.[0] || {}
                            const btnId = act.inscription_id
                            const isLoading = loadingId === btnId

                            return (
                                <li key={act.id} className="activity-card">
                                    <div className="card-image">
                                        {act.image_url ? (
                                            <img src={act.image_url} alt={act.title} />
                                        ) : (
                                            <div className="no-image">Sin imagen</div>
                                        )}
                                    </div>
                                    <div className="card-content">
                                        <h2>{act.title}</h2>
                                        <p>
                                            <strong>Categoría:</strong> {act.category_name}
                                        </p>
                                        <p>
                                            <strong>Instructor:</strong> {act.instructor_name}
                                        </p>
                                        <p>
                                            <strong>Próximo turno:</strong>{' '}
                                            {(act.day || act.dia_semana) || '—'} a las{' '}
                                            {(act.start_time || act.hora_inicio)?.slice(0, 5) || '—'}
                                        </p>
                                        <p>
                                            <strong>Duración:</strong>{' '}
                                            {(act.duration != null
                                                ? act.duration
                                                : act.duracion_min) + ' minutos'}
                                        </p>
                                        <p>
                                            <strong>Cupo máximo:</strong> {act.capacity} personas
                                        </p>

                                        <button
                                            onClick={() => handleUnsubscribe(act.inscription_id)}
                                            disabled={isLoading}
                                        >
                                            {isLoading ? 'Procesando...' : 'Desinscribirse'}
                                        </button>
                                    </div>
                                </li>
                            )
                        })}
                    </ul>
                )}
            </div>
        </div>
    )
}

export default MisActividades
