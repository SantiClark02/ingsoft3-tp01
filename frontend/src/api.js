// src/api.js
import axios from 'axios';

// Ruta relativa: el navegador pide al mismo origen que sirvió la página
// y nginx reenvía internamente al backend. Sin esto el front no sería
// portable entre entornos y habría que lidiar con CORS.
const api = axios.create({
    baseURL: '/api',
    headers: { 'Content-Type': 'application/json' },
});

api.interceptors.request.use(config => {
    const token = localStorage.getItem('token');
    if (token) config.headers.Authorization = `Bearer ${token}`;
    return config;
});

export default api;