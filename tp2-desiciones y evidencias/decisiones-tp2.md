---

# TP2 — Contenedores

## 1. Elección de la aplicación del semestre

### Las dos candidatas

Tenía dos aplicaciones web propias para elegir:

**Tienda online de indumentaria (descartada).** Es la más completa de las dos, pero está
construida con **arquitectura de microservicios**, con tres servicios independientes.
Contenerizarla habría implicado tres Dockerfiles, tres servicios en Compose y coordinar la
comunicación entre ellos. Esa complejidad de orquestación no aporta a lo que el TP2 evalúa —
que es entender imágenes, capas, redes, volúmenes y secretos — y multiplicaba las superficies
donde algo puede fallar.

**Aplicación de gestión de un gimnasio (elegida).** Arquitectura monolítica con tres piezas
claras: backend, frontend y base de datos.

### La app elegida

| Componente | Tecnología |
|---|---|
| Backend | Go 1.21, framework Gin, ORM GORM |
| Frontend | React 19 + Vite 6, cliente HTTP axios |
| Base de datos | MySQL 8.0 |

Funcionalidad: autenticación con JWT, roles de socio y administrador, catálogo de actividades
con horarios, e inscripciones de usuarios a esas actividades.

### Cómo cumple los criterios de la materia

- **Corre localmente:** sí, verificado antes de contenerizar nada.
- **La entiendo y la puedo modificar:** la escribí yo. Es requisito para la defensa en vivo.
- **Tiene backend, frontend y persistencia:** las tres capas, separadas.
- **Puede tener tests:** el backend en Go tiene una estructura por capas (controllers,
  services, dao) que permite testear los servicios de forma aislada. Es lo que voy a necesitar
  en TPs posteriores.
- **Tamaño manejable:** un dominio acotado (usuarios, actividades, horarios, inscripciones),
  lo bastante chico para defenderlo entero y lo bastante completo para cumplir los requisitos.
- **Es individual y distinta de la de mis compañeros.**

---

## 2. Decisiones de contenerización

### 2.1 Backend: multi-stage con Go

**Imagen de build:** `golang:1.23-alpine`
**Imagen final:** `alpine:3.20`

La etapa de build compila; la etapa final solo ejecuta. El binario resultante se copia con
`COPY --from=build` y el compilador, el módulo cache y el código fuente **no viajan** a la
imagen de producción.

Decisiones concretas:

- **`CGO_ENABLED=0`** produce un binario estático, sin dependencias de librerías del sistema.
  Es lo que permite que la imagen final sea Alpine y no una distribución completa.
- **`-ldflags="-s -w"`** quita los símbolos de debug y reduce el tamaño del binario.
- **Usuario sin privilegios** (`appuser`, UID 10001): el proceso no corre como root. Si alguien
  lograra ejecutar código dentro del contenedor, no tendría privilegios administrativos.
- **`ca-certificates`** instalado porque la app puede hacer llamadas HTTPS salientes.
- **`ENTRYPOINT` en lugar de `CMD`:** el contenedor tiene un único propósito, ejecutar el
  servidor. `ENTRYPOINT` fija ese comando y hace que cualquier argumento extra se le pase al
  binario en vez de reemplazarlo.

**Resultado medido: 370 MB (imagen de build) → 28.1 MB (imagen final). Una reducción del 92%.**

### 2.2 Frontend: build de Vite servido por nginx

**Imagen de build:** `node:20-alpine`
**Imagen final:** `nginx:1.27-alpine`

Vite compila la SPA a archivos estáticos. Node solo hace falta para *construir* esos archivos,
no para servirlos: la imagen final solo tiene nginx y el contenido de `dist/`.

- **`npm ci` en lugar de `npm install`:** `ci` instala exactamente las versiones fijadas en
  `package-lock.json`. Es reproducible; `install` puede resolver versiones distintas según
  cuándo se ejecute, lo que rompe la garantía de que la imagen se construya siempre igual.

**Resultado medido: 193 MB → 92.4 MB.** La reducción es menor que en el backend porque las
imágenes de fondo de la aplicación pesan unos 8 MB y forman parte del contenido servido.

### 2.3 Orden de las instrucciones y cache de capas

En los dos Dockerfiles se copian **primero los archivos de dependencias** (`go.mod` y `go.sum`
en el backend; `package.json` y `package-lock.json` en el frontend), se instalan las
dependencias, y **recién después** se copia el resto del código.

Docker cachea cada instrucción como una capa y solo reconstruye desde el punto donde algo
cambió. Con este orden, editar un archivo `.go` o `.jsx` no obliga a volver a descargar todas
las dependencias. Con el orden inverso, cada cambio de código invalidaría el cache de
instalación y cada build tardaría lo mismo que el primero.

### 2.4 `.dockerignore`

Se creó uno por cada contexto de build (`backend/` y `frontend/`).

En el frontend es especialmente importante excluir `node_modules`: sin eso, `COPY . .` copiaría
al contenedor las dependencias instaladas en Windows, que además de agrandar el contexto pueden
contener binarios compilados para otro sistema operativo e incompatibles con Linux.

Ambos excluyen también `.git`, `.env` y los artefactos de compilación.

### 2.5 Comunicación entre servicios

Docker Compose crea una red interna donde cada servicio es alcanzable por su nombre. El backend
se conecta a la base de datos usando el host `mysql`, que es el nombre del servicio en el
`docker-compose.yml`.

No se usan IPs fijas porque Docker las asigna dinámicamente y pueden cambiar entre reinicios.
El nombre del servicio es estable.

### 2.6 Frontend, nginx y el problema del navegador

**Esto exigió modificar código de la aplicación.**

El frontend original llamaba a la API con una URL absoluta:

```javascript
baseURL: 'http://localhost:8080'
```

Eso funciona cuando todo corre en la misma máquina, pero rompe la portabilidad: **el navegador
del usuario no pertenece a la red interna de Docker Compose**. El navegador no puede resolver
`http://backend:8080`, porque ese nombre solo existe dentro de la red de contenedores.

La solución adoptada es la que recomienda la consigna:

1. El frontend usa **rutas relativas**: `baseURL: '/api'`.
2. **nginx** recibe esas peticiones y las reenvía internamente al backend
   (`proxy_pass http://backend:8080/`). La barra final hace que `/api/activities` llegue al
   backend como `/activities`.

Ventajas: el frontend queda portable entre entornos (no tiene ninguna URL específica del
entorno en su código), y desaparece el problema de CORS, porque para el navegador todas las
peticiones van al mismo origen que sirvió la página.

También se agregó `try_files $uri $uri/ /index.html` en la configuración de nginx: es necesario
en una SPA para que las rutas manejadas por React Router funcionen al recargar la página.

En `vite.config.js` se configuró un proxy equivalente para el servidor de desarrollo, de modo
que `/api` funcione igual con Docker y sin Docker.

### 2.7 Base de datos, persistencia y healthcheck

**Volumen nombrado** (`mysql_data` montado en `/var/lib/mysql`), no bind mount. Los volúmenes
nombrados los gestiona Docker, son portables entre sistemas operativos y no dependen de una
ruta del host. Para bases de datos son la opción correcta.

**Qué persiste:** todo lo que se escribe en la base de datos (usuarios, actividades, horarios,
inscripciones).
**Qué es efímero:** el sistema de archivos de los contenedores de backend y frontend. Se
reconstruyen desde la imagen en cada arranque y no guardan estado.

**Healthcheck:** `mysqladmin ping`, con `depends_on: condition: service_healthy` en el backend.

Esto se decidió reemplazando un mecanismo anterior (ver problema 4.2).

### 2.8 Manejo de secretos

- **`.env.example`** se versiona: documenta qué variables necesita el proyecto, con valores de
  ejemplo.
- **`.env`** contiene los valores reales y está en `.gitignore`.

El README indica `cp .env.example .env` como primer paso de configuración. Esto significa que
levantar el sistema requiere **dos comandos, no uno** — y eso es intencional: el secreto es
justamente lo único que no puede viajar en un repositorio público.

Las variables `DB_*` que consume el backend se derivan en el `docker-compose.yml` a partir de
las `MYSQL_*`, en lugar de definirse por separado en el `.env`. Cada valor se define en un solo
lugar, lo que hace imposible que se desincronicen (ver problema 4.4).

### 2.9 Puertos

| Servicio | Puerto host | Puerto contenedor |
|---|---|---|
| frontend (nginx) | 5174 | 80 |
| backend | 8080 | 8080 |
| mysql | 3307 | 3306 |

MySQL se expone en el 3307 y no en el 3306 para no chocar con la instancia de MySQL instalada
directamente en el sistema operativo. Las dos pueden convivir.

### 2.10 Registry

Se eligió **GitHub Container Registry (ghcr.io)** porque el código ya vive en GitHub y no
requiere una cuenta adicional.

Imágenes publicadas con tag semántico `v0.1.0` y visibilidad pública:

- `ghcr.io/santiclark02/gimnasio-backend:v0.1.0`
- `ghcr.io/santiclark02/gimnasio-frontend:v0.1.0`

El `docker-compose.registry.yml` reemplaza las directivas `build:` por `image:` apuntando a
esas imágenes. El resto de la configuración (red, volumen, healthcheck, variables) se mantiene
idéntica.

La relación conceptual es: **el código va a GitHub, las imágenes van al registry.** Son dos
artefactos distintos con dos destinos distintos.

---

## 3. Cambios realizados sobre la aplicación original

La aplicación ya tenía Dockerfiles y un `docker-compose.yml` escritos previamente, pero no
cumplían los requisitos de la consigna. Se auditaron y modificaron:

| Archivo | Acción |
|---|---|
| `backend/Dockerfile` | Reescrito: no era multi-stage |
| `frontend/Dockerfile` | Ajustado: `npm ci`, nginx.conf activado |
| `docker-compose.yml` | Ajustado: `start_period`, `image:` con nombre fijo, montaje `:ro` |
| `frontend/src/api.js` | Modificado: URL absoluta → ruta relativa |
| `frontend/vite.config.js` | Agregado proxy de desarrollo |
| `backend/wait-for-it.sh` | **Eliminado** |
| `backend/.dockerignore` | Creado |
| `frontend/.dockerignore` | Creado |
| `frontend/nginx.conf` | Creado |
| `.env.example` | Creado |
| `docker-compose.registry.yml` | Creado |

---

## 4. Problemas encontrados

### 4.1 El Dockerfile del backend no era multi-stage

**Qué ocurrió.** El Dockerfile existente usaba `FROM golang:1.21` como única etapa: compilaba
y ejecutaba en la misma imagen.

**Por qué.** Se había escrito con el objetivo de que la aplicación funcionara, no de producir
una imagen apta para producción.

**Consecuencia.** La imagen final incluía el compilador de Go, el módulo cache y el código
fuente completo para ejecutar un binario de 28 MB. Además de un tamaño desproporcionado, eso
amplía la superficie de ataque: cada herramienta presente en una imagen de producción es una
herramienta disponible para quien logre entrar.

**Cómo se resolvió.** Reescritura completa como multi-stage, con etapa de build sobre
`golang:1.23-alpine` y etapa final sobre `alpine:3.20`.

**Cómo se verificó.** Comparando tamaños con `docker images`: 370 MB de la imagen de build
contra 28.1 MB de la final (evidencia 2).

### 4.2 `wait-for-it.sh`: mecanismo redundante y frágil

**Qué ocurrió.** El backend usaba un script `wait-for-it.sh` en su `CMD` para esperar a que
MySQL aceptara conexiones antes de arrancar. Además, al hacer `git add` apareció el aviso
`LF will be replaced by CRLF`.

**Por qué era un problema.** Tres razones acumuladas:

1. **Era redundante.** El `docker-compose.yml` ya resolvía la espera con un healthcheck de
   MySQL y `depends_on: condition: service_healthy`, que es el mecanismo que la propia
   plataforma provee.
2. **Impedía una imagen mínima.** El script necesita bash, lo que obliga a incluir un shell
   completo en la imagen final.
3. **Riesgo de fin de línea.** Un script con terminaciones CRLF de Windows falla dentro de un
   contenedor Linux con un error críptico (`bad interpreter`).

**Cómo se resolvió.** Se eliminó el script y se dejó el healthcheck como único mecanismo de
sincronización.

**Cómo se verificó.** En la salida de `docker compose up` se observa que MySQL demora entre 13
y 23 segundos en reportarse `Healthy`, y que el backend arranca recién después de eso. Sin ese
mecanismo, el backend arrancaría en el primer segundo y fallaría contra una base que todavía no
acepta conexiones.

Esto ilustra la diferencia central entre **"el contenedor arrancó"** y **"el servicio está
listo para aceptar conexiones"**. `depends_on` por sí solo solo garantiza lo primero.

### 4.3 El frontend usaba una URL absoluta hacia el backend

Descrito en detalle en la sección 2.6. Fue el único cambio que requirió modificar código de la
aplicación, y es el problema que la arquitectura de contenedores hace visible: algo que
funcionaba en una máquina deja de funcionar apenas el navegador y el backend dejan de compartir
el mismo host.

### 4.4 Desajustes en las variables de entorno

**Qué ocurrió.** El `.env` heredado del proyecto tenía tres inconsistencias:

1. `MYSQL_PASSWORD` estaba vacío. MySQL 8 no crea el usuario sin contraseña y el contenedor
   falla al inicializar.
2. `MYSQL_DATABASE=backend` pero `DB_NAME=gimnasio`. El script `db.sql` siembra los datos en
   una base llamada `backend`; el backend habría buscado en otra base, vacía.
3. `DB_USER`/`DB_PASSWORD` no coincidían con `MYSQL_USER`/`MYSQL_PASSWORD`.

**Por qué.** El archivo definía dos veces la misma información: una vez para el contenedor de
MySQL y otra para el backend. Duplicar una fuente de verdad garantiza que tarde o temprano se
desincronice.

**Cómo se resolvió.** El `.env` define únicamente las variables `MYSQL_*`, y el
`docker-compose.yml` deriva de ellas las `DB_*` que consume el backend. Cada valor se define
en un solo lugar.

**Cómo se verificó.** El sistema levanta y responde con datos reales desde la base
(evidencias 5 y 7).

### 4.5 El build parecía colgado y era saturación de recursos

**Qué ocurrió.** El primer `docker compose up -d --build` quedó sin avanzar, con la terminal
aparentemente congelada y sin mensajes de error.

**Cómo se investigó.** Revisando Docker Desktop se detectó que otro proyecto de una materia
anterior seguía corriendo en segundo plano, con contenedores de memcached, solr y rabbitmq,
consumiendo **661% de CPU sobre 400% disponible** y 1.66 GB de los 3.62 GB asignados a Docker.

**Por qué.** El build no estaba trabado: estaba esperando CPU. Compilar Go y construir una
aplicación de Node son tareas intensivas y no quedaba procesador libre.

**Cómo se resolvió.** Deteniendo el proyecto ajeno desde Docker Desktop.

**Cómo se verificó.** `docker ps` y `docker stats` quedaron vacíos, y el build completo terminó
en 98.5 segundos.

Es un buen ejemplo de un problema que parece de configuración y es de entorno.

### 4.6 Login fallido por credenciales incorrectas

**Qué ocurrió.** Al probar el sistema levantado, el login devolvía "credenciales inválidas".

**Cómo se investigó.** Con `docker compose logs backend`. Los logs mostraban dos cosas: que el
backend busca usuarios por el campo `nombre` y no por email (por eso `admin@gimnasio.com`
devolvía "usuario no encontrado"), y que con el usuario `admin` la comparación de bcrypt
fallaba.

**Cómo se resolvió.** La contraseña correcta de los usuarios sembrados por `db.sql` es
`password123`, no `admin123`. No era un problema de contenerización.

**Aprendizaje aplicable.** Los logs del contenedor son la primera herramienta de diagnóstico:
`docker compose logs <servicio>` mostró exactamente qué consulta se ejecutaba y por qué fallaba.

---

## 5. Observaciones detectadas y no corregidas

Se documentan por transparencia. Están fuera del alcance del TP2, que evalúa contenerización y
no calidad del código de la aplicación, pero conviene tenerlas identificadas:

- **El backend registra contraseñas en texto plano en sus logs.** Se observó
  `Contraseña ingresada: admin123` en la salida de `docker compose logs`. En un entorno real
  esto es un problema de seguridad serio: las credenciales terminan almacenadas en los sistemas
  de logging.
- **Existen dos conexiones a MySQL, no una.** `db/basedatos.go` abre una conexión en
  `InitDB()` y `clients/mysql_client.go` abre otra en su `func init()`. Como los `init()` de
  paquete se ejecutan antes de `main()`, la aplicación se conecta y ejecuta `AutoMigrate` dos
  veces al arrancar. Funciona, pero es una duplicación innecesaria.
- **`db.sql` y `AutoMigrate` crean las mismas tablas.** MySQL ejecuta el script al inicializar
  el volumen y GORM aplica sus migraciones después. Conviven porque `AutoMigrate` es aditivo,
  pero son dos fuentes de verdad para el mismo esquema.

---

## 6. Declaración de uso de Inteligencia Artificial

Usé un asistente de IA (Claude) durante todo el desarrollo del TP2.

### En qué me asistió

- **Auditoría de los Dockerfiles existentes**, identificando que el del backend no era
  multi-stage y qué implicaba eso.
- **Escritura de los Dockerfiles, el `nginx.conf`, el `docker-compose.yml` y el
  `docker-compose.registry.yml`**, adaptados a mi stack real (Go 1.21, Vite 6, MySQL 8) y a mis
  rutas, puertos y nombres de proyecto concretos.
- **Explicaciones conceptuales:** multi-stage, cache de capas, la diferencia entre `CMD` y
  `ENTRYPOINT`, volúmenes nombrados vs bind mounts, por qué el navegador no puede resolver
  nombres de la red interna de Compose, y la diferencia entre un contenedor iniciado y un
  servicio listo.
- **Diagnóstico de problemas:** el análisis de los desajustes del `.env`, la identificación de
  la saturación de CPU a partir de las capturas de Docker Desktop, y la lectura de los logs del
  backend para entender el fallo de login.
- **Redacción** de este documento y de `evidencias.md`.

### Cómo verifiqué lo que me devolvió

- **Contra la ejecución real, no contra la teoría.** Ninguna configuración se dio por buena por
  parecer correcta. Cada cambio se probó levantando el sistema y verificando el resultado:
  `docker compose ps`, peticiones reales a la API, y el sistema funcionando en el navegador.
- **Los tamaños se midieron, no se estimaron.** Los 370 MB y los 28.1 MB salen de
  `docker images`, no de una expectativa.
- **La persistencia se probó en las dos direcciones**, comprobando que `down` conserva los
  datos y que `down -v` los elimina.
- **La publicación en el registry se verificó de la forma más exigente posible:** cerrando
  sesión con `docker logout`, borrando las imágenes locales con `docker rmi`, y levantando el
  sistema desde cero con el Compose de registry. Ver la imagen publicada en una página web no
  demuestra que sea descargable; esto sí.
- **Cuando el asistente asumió algo incorrecto, la salida real lo desmintió** y se corrigió el
  diagnóstico. Ocurrió por ejemplo con las imágenes de build, que no aparecían en
  `docker images` por quedar en la caché de BuildKit.

### Alcance

La IA me ayudó a entender, a configurar y a documentar, pero todas las acciones sobre el
repositorio y sobre Docker las ejecuté yo, verificando cada resultado. Comprendo cada decisión
tomada, cada instrucción de los Dockerfiles y el porqué de cada configuración del Compose, y
puedo explicarlas y demostrarlas levantando el sistema.
