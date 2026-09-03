# ingsoft3-tp01

Repositorio de la materia **Ingeniería de Software 3**.

Contiene la aplicación del semestre —un sistema de gestión de gimnasio— y los trabajos
prácticos que se desarrollan sobre ella.

- **TP1 — Git y GitHub colaborativo:** flujo de trabajo, protección de `main`, Pull Requests,
  resolución de conflictos, tags y releases.
- **TP2 — Contenedores y Docker:** contenerización completa de la aplicación y publicación de
  las imágenes en un registry público.

La documentación de cada entrega está en [`decisiones.md`](decisiones.md) (qué se decidió y por
qué) y [`evidencias.md`](evidencias.md) (pruebas de funcionamiento).

---

## La aplicación

Sistema de gestión de un gimnasio: catálogo de actividades con sus horarios, inscripción de
socios, y administración del catálogo por parte de un usuario administrador.

| Componente | Tecnología |
|---|---|
| Frontend | React 19 + Vite 6, servido por nginx |
| Backend | Go 1.21 (Gin + GORM) |
| Base de datos | MySQL 8.0 |

### Arquitectura

```
Navegador
    ↓  :5174
Frontend (nginx)
    ↓  /api  →  proxy interno
Backend (Go)
    ↓  red interna de Compose
MySQL
```

El navegador **no** pertenece a la red interna de Docker Compose, por lo que no puede resolver
nombres de servicio como `backend`. Por eso el frontend hace peticiones a rutas relativas
(`/api/...`) y nginx las reenvía internamente al backend. Esto además evita cualquier problema
de CORS.

---

## Levantar el sistema desde cero

### Requisitos

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) instalado y corriendo
- Git

No hace falta tener Go, Node ni MySQL instalados: todo corre en contenedores.

### Pasos

**1. Clonar el repositorio**

```bash
git clone https://github.com/SantiClark02/ingsoft3-tp01.git
cd ingsoft3-tp01
```

**2. Crear el archivo de variables de entorno**

```bash
cp .env.example .env
```

En PowerShell también funciona `cp`, o su forma completa `Copy-Item .env.example .env`.

Abrí el `.env` y reemplazá los valores de ejemplo por valores propios. Como mínimo:

```
MYSQL_ROOT_PASSWORD=una_password_root
MYSQL_DATABASE=backend
MYSQL_USER=gimnasio_user
MYSQL_PASSWORD=una_password_de_usuario
JWT_SECRET=un_secreto_largo_y_aleatorio
```

> ⚠️ **`MYSQL_DATABASE` debe ser `backend`.** Es el nombre de base que usa el script de
> inicialización `db.sql`. Si se cambia, la aplicación se conectará a una base vacía.

> ℹ️ **Por qué son dos pasos y no uno.** El archivo `.env` contiene secretos y por eso no está
> versionado; solo se versiona `.env.example`, que documenta qué variables hacen falta sin
> revelar ningún valor real. Ese es exactamente el punto: el secreto es lo único que no puede
> viajar dentro del repositorio.

**3. Levantar todo**

```bash
docker compose up -d
```

La primera vez tarda unos minutos: descarga las imágenes base y construye el backend y el
frontend.

**4. Verificar**

```bash
docker compose ps
```

Los tres servicios deben aparecer en estado `Up`, y `gimnasio_mysql` como `(healthy)`.

Abrí **http://localhost:5174** en el navegador.

### Credenciales de prueba

Los usuarios los siembra `db.sql` al inicializar la base por primera vez.

| Usuario | Contraseña | Rol |
|---|---|---|
| `admin` | `password123` | administrador |
| `emiliano` | `password123` | socio |
| `maria` | `password123` | socio |

> El login se hace con el **nombre de usuario**, no con el email.

### Puertos

| Servicio | URL / puerto |
|---|---|
| Frontend | http://localhost:5174 |
| Backend (API) | http://localhost:8080 |
| MySQL | `localhost:3307` |

MySQL se expone en el 3307 para no chocar con una instalación local de MySQL en el 3306.

---

## Levantar usando las imágenes publicadas

Existe una variante del Compose que **no construye nada**: descarga las imágenes ya publicadas
en GitHub Container Registry.

```bash
cp .env.example .env       # editar los valores
docker compose -f docker-compose.registry.yml up -d
```

Imágenes utilizadas, ambas públicas y descargables sin credenciales:

- `ghcr.io/santiclark02/gimnasio-backend:v0.1.0`
- `ghcr.io/santiclark02/gimnasio-frontend:v0.1.0`

---

## Comandos útiles

```bash
docker compose ps                    # estado de los servicios
docker compose logs backend          # logs de un servicio
docker compose logs -f               # seguir los logs en vivo
docker compose down                  # detener y eliminar contenedores (conserva los datos)
docker compose down -v               # además elimina el volumen: BORRA LOS DATOS
docker compose up -d --build         # reconstruir las imágenes y levantar
```

> ⚠️ `docker compose down -v` elimina el volumen `mysql_data` y con él **todos los datos de la
> base**. Al volver a levantar, MySQL se reinicializa desde `db.sql` y solo reaparecen los datos
> sembrados.

---

## Estructura del repositorio

```
.
├── backend/                      # API en Go
│   ├── Dockerfile                # multi-stage: golang:1.23-alpine → alpine:3.20
│   └── .dockerignore
├── frontend/                     # SPA en React + Vite
│   ├── Dockerfile                # multi-stage: node:20-alpine → nginx:1.27-alpine
│   ├── .dockerignore
│   └── nginx.conf                # sirve la SPA y hace proxy de /api al backend
├── img/                          # capturas de evidencia
├── db.sql                        # esquema y datos iniciales de MySQL
├── docker-compose.yml            # construye las imágenes localmente
├── docker-compose.registry.yml   # usa las imágenes publicadas en ghcr.io
├── .env.example                  # plantilla de variables (se versiona)
├── decisiones.md                 # decisiones técnicas de cada TP
└── evidencias.md                 # evidencias de funcionamiento
```

El archivo `.env` con los valores reales **no está versionado** y no debe subirse al
repositorio.
