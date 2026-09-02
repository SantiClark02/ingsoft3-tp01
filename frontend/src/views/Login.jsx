// src/views/Login.jsx
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../api';
import './Login.css';

const Login = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const navigate = useNavigate();

    const handleLogin = async (e) => {
        e.preventDefault();
        try {
            const resp = await api.post('/users/login', {
                username,
                password
            });

            // Guardar token, user_id y rol
            localStorage.setItem('token', resp.data.token);
            localStorage.setItem('user_id', resp.data.user_id);

            // Ajustamos para cubrir cualquier combinación de mayúsculas/minúsculas
            if (resp.data.rol && resp.data.rol.toLowerCase() === 'administrador') {
                localStorage.setItem('is_admin', 'true');
            } else {
                localStorage.setItem('is_admin', 'false');
            }

            navigate('/home');
        } catch (err) {
            console.error(err);
            alert(err.response?.data?.error || 'Error al iniciar sesión');
        }
    };

    return (
        <div className="login-container">
            <form className="login-form" onSubmit={handleLogin}>
                <h2>Iniciar Sesión</h2>
                <input
                    type="text"
                    placeholder="Usuario"
                    value={username}
                    onChange={e => setUsername(e.target.value)}
                />
                <input
                    type="password"
                    placeholder="Contraseña"
                    value={password}
                    onChange={e => setPassword(e.target.value)}
                />
                <button type="submit">Ingresar</button>
            </form>
        </div>
    );
};

export default Login;