// src/main.jsx
import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Login from './views/Login.jsx';
import Home from './views/Home.jsx';
import MisActividades from './views/MisActividades.jsx';
import NuevaActividad from './views/NuevaActividad.jsx';
import EditarEliminarActividad from './views/EditarEliminarActividad.jsx';
import EditarActividad from './views/EditarActividad.jsx';
import './index.css';

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Login />} />
        <Route path="/home" element={<Home />} />
        <Route path="/misactividades" element={<MisActividades />} />
        <Route path="/nuevaactividad" element={<NuevaActividad />} />
        <Route path="/editarEliminarActividad" element={<EditarEliminarActividad />} />
        <Route path="/editaractividad/:id" element={<EditarActividad />} />
      </Routes>
    </BrowserRouter>
  </React.StrictMode>
);
