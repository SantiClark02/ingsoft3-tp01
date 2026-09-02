// src/views/Home.jsx
import React, { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import api from '../api'
import './Home.css'

const Home = () => {
    const [allActivities, setAllActivities] = useState([])
    const [activities, setActivities] = useState([])
    const [error, setError] = useState(null)
    const [search, setSearch] = useState('')
    const [loadingId, setLoadingId] = useState(null)
    const userId = localStorage.getItem('user_id')
    const isAdmin = localStorage.getItem('is_admin') === 'true'
    const navigate = useNavigate()

    const handleLogout = () => {
        localStorage.removeItem('user_id')
        localStorage.removeItem('token')
        localStorage.removeItem('is_admin')
        navigate('/')
    }

    const goToMisActividades = () => {
        navigate('/misactividades')
    }

    const goToNuevaActividad = () => {
        navigate('/nuevaactividad')
    }

    const goToEditarEliminar = () => {
        navigate('/editarEliminarActividad')
    }

    const loadActivities = async () => {
        try {
            const params = {}
            if (userId) params.user_id = userId
            const resp = await api.get('/activities', { params })
            setAllActivities(resp.data)
            setActivities(resp.data)
        } catch {
            setError('No se pudieron cargar las actividades')
        }
    }

    useEffect(() => {
        loadActivities()
    }, [])

    const handleSearch = e => {
        e.preventDefault()
        const term = search.trim().toLowerCase()
        if (!term) {
            setActivities(allActivities)
            return
        }
        const filtered = allActivities.filter(act => {
            const title = (act.title || act.titulo || '').toLowerCase()
            const category = (act.category_name || act.nombre_categoria || '').toLowerCase()
            const day = (act.day || act.dia_semana || '').toLowerCase()
            const time = (act.start_time || act.hora_inicio || '').toLowerCase()
            return (
                title.includes(term) ||
                category.includes(term) ||
                day.includes(term) ||
                time.includes(term)
            )
        })
        setActivities(filtered)
    }

    const handleSubscribe = async activityId => {
        if (!userId) {
            alert('Debes iniciar sesión primero')
            return
        }
        setLoadingId(activityId)
        try {
            const { data } = await api.get(`/activities/${activityId}`)
            const horId = data.schedule?.[0]?.id
            if (!horId || horId <= 0) {
                alert('No hay horarios disponibles')
                return
            }
            await api.post('/inscripciones', { usuario_id: +userId, horario_id: horId })
            await loadActivities()
            alert('¡Inscripción exitosa!')
        } catch (err) {
            console.error(err)
            alert(err.response?.data?.error || 'Error al inscribirse')
        } finally {
            setLoadingId(null)
        }
    }

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
        <div className="home-container">
            <div className="content-wrapper">
                <div className="logout-bar" style={{ justifyContent: 'space-between', gap: '10px', flexWrap: 'wrap' }}>
                    <button className="logout-button" onClick={handleLogout}>
                        Cerrar sesión
                    </button>
                    <button className="my-activities-button" onClick={goToMisActividades}>
                        Mis actividades
                    </button>
                    {isAdmin && (
                        <>
                            <button className="create-activity-button" onClick={goToNuevaActividad}>
                                Crear actividad
                            </button>
                            <button className="edit-delete-button" onClick={goToEditarEliminar}>
                                Editar / Eliminar Actividades
                            </button>
                        </>
                    )}
                </div>

                <h1>Actividades Disponibles</h1>

                <form className="search-form" onSubmit={handleSearch}>
                    <input
                        type="text"
                        placeholder="Buscar por título, categoría, día u horario..."
                        value={search}
                        onChange={e => setSearch(e.target.value)}
                    />
                    <button type="submit">Buscar</button>
                </form>

                {activities.length === 0 && (
                    <p className="no-results">No se encontraron actividades para la búsqueda.</p>
                )}

                <ul className="activities-list">
                    {activities.map(act => {
                        const btnId = act.subscribed ? act.inscription_id : act.id
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

                                    <button
                                        onClick={() =>
                                            act.subscribed
                                                ? handleUnsubscribe(act.inscription_id)
                                                : handleSubscribe(act.id)
                                        }
                                        disabled={isLoading}
                                    >
                                        {isLoading
                                            ? 'Procesando...'
                                            : act.subscribed
                                                ? 'Desinscribirse'
                                                : 'Inscribirse'}
                                    </button>
                                </div>
                            </li>
                        )
                    })}
                </ul>
            </div>
        </div>
    )
}

export default Home
