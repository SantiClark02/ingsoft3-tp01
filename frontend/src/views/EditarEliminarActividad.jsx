// src/views/EditarEliminarActividad.jsx
import React, { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import api from '../api'
import './EditarEliminarActividad.css'

const EditarEliminarActividad = () => {
    const [activities, setActivities] = useState([])
    const [error, setError] = useState(null)
    const [loadingId, setLoadingId] = useState(null)
    const isAdmin = localStorage.getItem('is_admin') === 'true'
    const navigate = useNavigate()

    useEffect(() => {
        if (!isAdmin) {
            alert('Acceso restringido')
            navigate('/home')
            return
        }
        loadActivities()
    }, [])

    const loadActivities = async () => {
        try {
            const resp = await api.get('/activities')
            setActivities(resp.data)
        } catch {
            setError('No se pudieron cargar las actividades')
        }
    }

    const handleDelete = async activityId => {
        const token = localStorage.getItem('token')
        if (!window.confirm('¿Estás seguro de que quieres eliminar esta actividad?')) return
        try {
            await api.delete(`/activities/${activityId}`, {
                headers: { Authorization: `Bearer ${token}` }
            })
            alert('Actividad eliminada correctamente')
            loadActivities()
        } catch (err) {
            console.error(err)
            alert(err.response?.data?.error || 'Error al eliminar la actividad')
        }
    }

    const handleEdit = activityId => {
        navigate(`/editaractividad/${activityId}`)
    }

    const handleBack = () => {
        navigate('/home')
    }

    if (error) return <p className="error">{error}</p>

    return (
        <div className="home-container">
            <div className="content-wrapper">
                <div className="logout-bar" style={{ justifyContent: 'space-between', gap: '10px' }}>
                    <button className="back-button" onClick={handleBack}>
                        Volver al Home
                    </button>
                </div>

                <h1>Editar / Eliminar Actividades</h1>

                {activities.length === 0 && (
                    <p className="no-results">No hay actividades registradas.</p>
                )}

                <ul className="activities-list">
                    {activities.map(act => (
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
                                <p><strong>Categoría:</strong> {act.category_name}</p>
                                <p><strong>Instructor:</strong> {act.instructor_name}</p>
                                <p>
                                    <strong>Próximo turno:</strong>{' '}
                                    {act.day || act.dia_semana} a las {act.start_time || act.hora_inicio}
                                </p>
                                <p>
                                    <strong>Duración:</strong>{' '}
                                    {act.duration != null ? act.duration : act.duracion_min} minutos
                                </p>
                                <p>
                                    <strong>Cupo máximo:</strong> {act.capacity} personas
                                </p>

                                <div className="admin-buttons">
                                    <button
                                        className="edit-button"
                                        onClick={() => handleEdit(act.id)}
                                    >
                                        Editar
                                    </button>
                                    <button
                                        className="delete-button"
                                        onClick={() => handleDelete(act.id)}
                                    >
                                        Eliminar
                                    </button>
                                </div>
                            </div>
                        </li>
                    ))}
                </ul>
            </div>
        </div>
    )
}

export default EditarEliminarActividad
