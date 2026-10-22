# decisiones.md — acumulativo (TP1 → TP9)

Repositorio: https://github.com/SantiClark02/ingsoft3-tp01

Éste es el `decisiones.md` **acumulativo** que piden las guías: cada TP agrega su sección abajo. Las secciones de **TP5 a TP9 están completas acá**, y cada carpeta `tpN - decisiones/` guarda la misma sección del práctico. Los TP1 a TP4 (entrega 1) viven en sus carpetas, junto con sus evidencias e imágenes.

| TP | Tema | Dónde | Release |
|---|---|---|---|
| TP1 | Repositorio, ramas, PRs y protección de `main` | [tp1-desiciones y evidencia/decisiones.md](tp1-desiciones%20y%20evidencia/decisiones.md) | `v1.0.0` |
| TP2 | Contenedores y Docker Compose | [tp2-desiciones y evidencias/decisiones-tp2.md](tp2-desiciones%20y%20evidencias/decisiones-tp2.md) | — |
| TP3 y TP4 | Planificación y trazabilidad · Integración continua | [tp3 y 4 - desiciones y evidencia/decisiones-tp3-tp4.md](tp3%20y%204%20-%20desiciones%20y%20evidencia/decisiones-tp3-tp4.md) | `v4.0.0` |
| TP5 | Testing, cobertura y el umbral que frena el merge | [más abajo](#tp5) | `v5.0.0` |
| TP6 | CD: environments QA/PROD, aprobación, rechazo y rollback. **Por qué el fallback local y no Render + Neon** | [más abajo](#tp6) | `v6.0.0` |
| TP7 | Build once, deploy many + integración y e2e como gate | [más abajo](#tp7) | `v7.0.0` |
| TP8 | Infraestructura como código: preprod con Terraform | [más abajo](#tp8) | `v8.0.0` |
| TP9 | DevSecOps + observabilidad (+ [`postmortem.md`](postmortem.md)) | [más abajo](#tp9) | `v9.0.0` |


---

<a id="tp5"></a>

# TP5 — Calidad automatizada: tests, coverage y el umbral que frena un merge

Sección del TP5 del registro de decisiones (sigue la misma convención de carpeta por TP que los
prácticos anteriores). Como pide la consigna, acá no hay `evidencias.md`: cada decisión lleva al
lado el link que la prueba.

## Links de evidencia

| Qué prueba | Link |
|---|---|
| PR de la suite + coverage + umbral (verde, mergeado) | https://github.com/SantiClark02/ingsoft3-tp01/pull/20 |
| Corrida verde con el resumen de coverage (Summary) y los reportes descargables (Artifacts) | https://github.com/SantiClark02/ingsoft3-tp01/actions/runs/39506491353 |
| PR con la secuencia rojo → tests → verde → merge | https://github.com/SantiClark02/ingsoft3-tp01/pull/21 |
| La corrida **roja por umbral** de ese PR | https://github.com/SantiClark02/ingsoft3-tp01/actions/runs/39511661429 |
| La corrida verde después de agregar los tests | https://github.com/SantiClark02/ingsoft3-tp01/actions/runs/39512189750 |
| PR que **queda abierto y en rojo hasta la defensa** (el freno vigente) | https://github.com/SantiClark02/ingsoft3-tp01/pull/22 |
| Su corrida roja (backend bajo el umbral) | https://github.com/SantiClark02/ingsoft3-tp01/actions/runs/39512659964 |

---

## 1. Qué lógica elegí testear y por qué esa

El criterio fue **dónde duele un bug** en un sistema de gimnasio:

| Regla | Dónde vive | Por qué duele si se rompe |
|---|---|---|
| Validación de una inscripción (actividad activa, cupo, no duplicarse) | `backend/reglas/inscripcion.go` → `ValidarInscripcion` | Sobrevender una clase o anotar gente en una actividad dada de baja |
| Servicio de inscripción (sólo crea si se cumplen las reglas) | `backend/reglas/inscripcion.go` → `Inscribir` | Que se guarde una inscripción inválida o que se guarde dos veces |
| Autenticación (JWT) | `backend/utils/jwt.go` | Que un token falsificado o basura sea aceptado |
| Autorización (sólo administradores crean/editan/borran actividades) | `backend/utils/jwt.go` → `AdminMiddleware` | Que un socio pueda modificar el catálogo |
| Búsqueda de actividades | `frontend/src/lib/actividades.js` → `filtrarActividades` | El socio no encuentra la clase que busca |
| Inscripción desde el front (qué horario se manda a la API) | `frontend/src/lib/actividades.js` → `inscribirse` | Se inscribe en un horario equivocado o se llama a la API sin horario |
| Estado del cupo (agregado en la demo del gate) | `frontend/src/lib/cupos.js` → `estadoCupo` | Mostrar "disponible" una clase llena |

**Suite resultante:**

- **Backend — 9 funciones de test sobre 4 reglas** (`reglas/inscripcion_test.go`, `utils/jwt_test.go`):
  - Parametrizados (table-driven + `t.Run`): `TestValidarInscripcion_CasosQueSeRechazan`, `TestAdminMiddleware_SoloDejaPasarAdministradores`.
  - Casos de error / borde: cupo justo lleno (10 de 10), último lugar disponible (9 de 10), horario inexistente, token firmado con otro secreto, token mal formado.
  - Con mock: `TestInscribir_TodoValido_CreaUnaSolaVez` y `TestInscribir_SinCupo_NoCreaLaInscripcion`.
- **Frontend — 4 + 2 bloques de test** (`src/lib/actividades.test.js`, `src/lib/cupos.test.js`):
  - Parametrizados con `it.each` (búsqueda por título, categoría, día y horario; un caso por cada camino de `estadoCupo`).
  - Caso de error: actividad sin horarios → rechaza y no llama al POST.
  - Con mock: el cliente HTTP es un `vi.fn()`.

**Criterio de que un test vale:** si invierto el `>=` del cupo por `>`, se pone rojo "cupo justo lleno (borde)"; si el middleware dejara pasar a un socio, se pone rojo el caso `socio`; si el front mandara el segundo horario en vez del primero, se pone rojo el `toHaveBeenCalledWith`.

## 2. El refactor para poder mockear

**Backend.** Antes, `services.CrearInscripcion` hacía todas las validaciones adentro y le hablaba directo a `clients.DB` (una variable global). Además el paquete `clients` se conecta a MySQL en su `init()`, o sea **al importarlo**: cualquier test de `services` intenta conectarse a la base y explota antes de correr. No era difícil de testear: era imposible sin una base levantada.

Lo que cambié:

1. Saqué las reglas a un paquete nuevo, `backend/reglas`, que **no importa nada de la base**.
2. `ValidarInscripcion` quedó como función pura (recibe datos, devuelve error o nil): se testea sin dobles.
3. `Inscribir` recibe la dependencia **desde afuera**: una interfaz `Repositorio` (`DatosHorario`, `ContarInscriptos`, `YaInscripto`, `Crear`).
4. En `services/inscripciones_servicies.go` quedó `repoGorm`, la implementación real con GORM, y `CrearInscripcion` mantiene su firma: el controller no cambió. Ese es el lugar donde se "cablea" la dependencia real (el equivalente al registro en `Program.cs` de la guía).

Las reglas siguen siendo las mismas y en el mismo orden (inactiva → sin cupo → ya inscripto); sólo cambió de dónde viene la base.

**Frontend.** La lógica estaba adentro de `Home.jsx` y usaba el `api` (axios) importado directamente. La saqué a `src/lib/actividades.js`, y `inscribirse(actividadId, usuarioId, cliente)` recibe el cliente HTTP por parámetro: la app le pasa `api`, el test un doble. `Home.jsx` ahora llama a esas funciones.

**Mock vs stub en mi suite:** `repoMock` (Go) y los `vi.fn()` contestan datos fijos (eso es stub), pero los tests además **verifican la interacción**: cuántas veces se llamó a `Crear` (1 o 0) y con qué argumentos se llamó al `post`. Esa verificación es lo que los vuelve mocks.

## 3. Mi stack no es el de la cátedra: qué usé en cada fila

| Lo que se pide | Backend (Go) | Frontend (React + Vite) |
|---|---|---|
| Dónde viven los tests | `*_test.go` al lado del código | `*.test.js` al lado del código |
| Parametrizado | tabla de casos + `t.Run` (table-driven) | `it.each` |
| Que la dependencia entre desde afuera | interfaz + parámetro de la función | parámetro de la función |
| Fabricar el doble | struct escrito a mano que implementa la interfaz y cuenta llamadas (en Go las interfaces son implícitas; no hizo falta ninguna librería) | `vi.fn()` de vitest |
| Medir la cobertura | `go test -coverpkg=./... -coverprofile` | vitest + `@vitest/coverage-v8` (3.2.7) |
| Umbral que rompe el build | Go no tiene flag: `backend/scripts/test-coverage.sh` lee el total de `go tool cover -func` y sale con error si no llega | `coverage.thresholds` en `vite.config.js` |
| Qué entra en la cuenta | se **excluyen** líneas del profile con `grep -v` | `include: ['src/lib/**']` |
| Reporte legible | `go tool cover -html` | reporter `html` + `json-summary` |
| Herramientas de test en la etapa de tests del Dockerfile | `FROM build AS test`: ya tiene Go y los módulos | `FROM build AS test`: `npm ci` sin `--omit=dev`, ya trae vitest |

**Cómo corre en el pipeline:** cada Dockerfile tiene una etapa nueva `test` en el medio (entre `build` y la final). El job la construye con `target: test` y `load: true`, la corre con `docker run` montando una carpeta del runner (`-v`) para sacar los reportes, publica el resumen en el Summary y sube el reporte como artefacto (`coverage-backend`, `coverage-frontend`). El pipeline sigue sin saber cómo se compila o se testea la app: se lo pide al Dockerfile. La imagen final no se lleva ni las herramientas ni los tests.

## 4. Mis umbrales

### Backend: 15 % de sentencias (medido hoy: 15,9 %)

- **Por qué 15:** antes del TP5 el backend no tenía ni un test. Medido honestamente (incluyendo controllers y servicios que hablan con la base), hoy da 15,9 %. Puse el umbral apenas debajo para que me **frene si baja**: cualquier código nuevo sin tests lo hace caer. Exigir 70 % el primer día sobre un repo con esa deuda sería condenar el gate a que alguien lo apague.
- **Métrica:** sentencias, porque **Go no mide cobertura de ramas** (`-cover` cuenta sentencias). Lo reporto como "no disponible" en el Summary en vez de inventarlo.
- **Para subirlo a ~40 %** habría que sacar las reglas de `activities_servicies.go` a `reglas/` como hice con inscripciones y testear los handlers con `httptest`. Si mañana lo subo 10 puntos (a 25), el build queda rojo hoy mismo.
- **Cómo lo evalúa:** sobre el total de todo lo medido (no por paquete).

### Frontend: 80 % de líneas **y** 80 % de ramas (medido hoy: 100 % líneas, 90 % ramas)

- Puse umbral en las dos métricas porque con vitest 3 una función que ningún test llama sólo baja las líneas, y la de ramas es la más honesta.
- 80 y no 100: la lógica del front es chica, y un 100 se rompería por una rama defensiva sin valor (ver el ejercicio de la sección 6).

## 5. Qué dejé afuera de la cuenta y por qué

**Backend** (excluido en `scripts/test-coverage.sh`):

- `main.go` — el arranque: cablea rutas y middlewares, no tiene reglas. Si está mal, la app no levanta.
- `db/` y `clients/` — conexión a MySQL, migraciones y seed: arranque e infraestructura, no lógica.
- `dao/` y `domain/` — structs de datos sin comportamiento.

**Lo que sí cuenta, aunque hoy esté en 0 %:** `controllers/` y `services/`. Tienen comportamiento (validaciones de entrada, mapeos), así que dejarlos afuera sería esconder lo que no testeé. Uso **exclusión** y no inclusión a propósito: un paquete nuevo entra solo a la cuenta, y si no tiene tests el número baja.

**Frontend** (`include: ['src/lib/**']`): cuenta sólo la lógica pura. Los componentes React (`src/views`) y el pegamento de UI (`main.jsx`, `App.jsx`, `api.js`) se verifican end-to-end en el TP7. Testear componentes requiere jsdom + Testing Library y la consigna lo deja como opcional.

## 6. Ejercicio del camino sin cubrir

En el reporte del frontend, `actividades.js` marca ramas sin cubrir en las líneas 6 y 10. Elegí la línea 6:

```js
const t = (termino || '').trim().toLowerCase()
```

1. **Qué línea es:** la rama derecha del `||`: el caso en que `termino` es `null` o `undefined`. Ningún test la recorre: todos pasan un string.
2. **Qué entrada la recorrería:** `filtrarActividades(actividades, null)` → debería devolver todas las actividades.
3. **Qué decidí: no agregarlo.** En la app el término viene de un `useState('')`, que nunca es `null`; la rama es defensiva. Prefiero no inflar la suite con un test que protege un caso que hoy no puede pasar. Si el término empezara a venir de la URL (donde sí puede faltar), ahí agrego el test.

(En el backend, el HTML de Go muestra en rojo los `return 0, err` de `Inscribir` cuando el repositorio falla al contar o al buscar duplicados. Se recorrerían con un doble que devuelva error en `ContarInscriptos`. Tampoco lo agregué: es propagación del error, no una regla de negocio.)

## 7. Por qué coverage alto no garantiza calidad (con mi ejemplo)

Este test dejaría `inscribirse` con cobertura total y **no verifica nada**:

```js
it('inscribe', async () => {
    const cliente = { get: vi.fn().mockResolvedValue({ data: { schedule: [{ id: 5 }] } }), post: vi.fn() }
    await inscribirse(3, 7, cliente)   // se ejecutó todo… y no hay ningún expect
})
```

Si alguien cambiara el código para mandar `horario_id: 9`, ese test seguiría verde. El mío no, porque verifica con `toHaveBeenCalledWith` qué se le mandó a la API. Por eso la cobertura sirve como **detector de agujeros** (lo que está en 0 % seguro no está verificado), no como medida de calidad. Lo que contesta "¿mis tests verifican?" son los mutantes: cambiar `>=` por `>` en el cupo y ver si algo se pone rojo. En mi suite, el caso "cupo justo lleno (borde)" lo mata.

## 8. El Pull Request bloqueado

**PR #21 (la historia completa).** Agregué `estadoCupo` en `frontend/src/lib/cupos.js` (18 líneas, 5 caminos) sin tests. Compilaba y todos los tests pasaban, pero el check **`build-frontend` (Required)** se puso rojo en el paso "Correr tests del frontend con coverage y umbral":

```
ERROR: Coverage for lines (55%) does not meet global threshold (80%)
ERROR: Coverage for branches (75%) does not meet global threshold (80%)
```

Frenó **por las dos métricas** (vitest 3.2.7), y el merge quedó bloqueado. Lo arreglé con `cupos.test.js`: un caso por cada camino del código (`it.each` con sin cupo definido, completo, últimos lugares, más de la mitad ocupado, disponible) más un caso de borde (más inscriptos que cupo). Volvió a verde (100 % líneas, 90 % ramas) y lo mergeé.

**PR #22 (el que queda abierto).** Agrega `CuotaMensual` en `backend/reglas/cuota.go` sin tests. El check **`build-backend` (Required)** queda rojo:

```
ERROR: Coverage (statements) 14.9% does not meet threshold 15%
```

Queda abierto a propósito hasta la defensa, para que se pueda comprobar que el freno existe sin depender de una captura. Entre los dos PRs, el freno queda demostrado **de los dos lados**.

**Por qué este freno es distinto del del TP4:** el del TP4 frenaba si algo no compilaba. Este frena código que compila perfecto y pasa todos los tests, por un número que elegí yo. Ahora `main` tiene tres guardianes: el PR obligatorio (TP1), el build verde (TP4) y los tests con su umbral (TP5).

**Qué deja pasar igual:** tests sin asserts que suben el número, y sobre todo un requisito mal entendido. Si leí mal la regla del cupo, el test la congela mal y queda verde para siempre. Eso sólo lo detecta alguien que sepa qué se quiso pedir, mirando el PR.

## 9. Problemas encontrados y cómo los resolví

| Problema | Solución |
|---|---|
| Go no está en la tabla de la cátedra y no tiene un flag de umbral | Script `test-coverage.sh`: lee el total de `go tool cover -func` y sale con `exit 1` si no llega. Verificado localmente: con umbral 16 falla, con 15 pasa |
| Go no mide ramas | Lo documenté y lo reporto como "no disponible"; la métrica de rama la reporto desde el frontend |
| `clients` se conecta a MySQL en `init()`, y cualquier test de `services` explotaba | Las reglas pasaron a `reglas/`, un paquete que no importa la base |
| Con `-coverpkg=./...` los paquetes sin tests cuentan en 0 % y el número da bajo | Es lo honesto: lo dejé así y elegí el umbral contra esa medición |
| Riesgo de finales de línea CRLF en el `.sh` editado desde Windows (ya nos pasó en el TP2 con `wait-for-it.sh`) | La etapa de tests corre `sed -i 's/\r$//'` sobre el script antes de ejecutarlo |
| `${COVERAGE_DIR:-coverage}` no se expande en `cmd.exe` | En Windows corro `npm test -- --run --coverage`; `test:ci` es el del contenedor (Linux) |
| Durante la demo, dos commits quedaron en mi `main` local y el push fue rechazado (`GH006: Protected branch update failed … Changes must be made through a pull request`) | Moví los commits a la rama `demo/tp5-gate-frontend` (`git branch -f` + `git reset --hard origin/main`). De paso confirmó que la protección del TP1 sigue funcionando |

## 10. Declaración de uso de IA

Usé **Claude (Anthropic)** para analizar la consigna, adaptar cada fila de la tabla de la cátedra a Go, escribir el refactor de inscripciones, los tests, el script de cobertura, las etapas de test de los Dockerfiles, los pasos del `ci.yml` y el borrador de este documento.

**Cómo lo verifiqué:**

- Los tests se corrieron localmente y en el pipeline (PR #20 verde).
- El gate se probó en los dos sentidos: el script falla con umbral 16 y pasa con 15, y los PRs #21 y #22 se pusieron rojos por cobertura.
- El refactor no cambió el contrato: el backend y el frontend compilan en el pipeline, `CrearInscripcion` mantiene su firma (el controller no se tocó) y `Home.jsx` sólo pasó a llamar a las funciones de `src/lib`.

**Qué verifica cada assert, y qué NO está cubierto:**

- `TestValidarInscripcion_*`: el error exacto que devuelve cada regla (`errors.Is`). **No cubierto:** usuario o horario con ID 0 (lo filtra el controller, que no tiene tests).
- `TestInscribir_*`: que `Crear` se llame 1 vez cuando todo es válido y 0 veces sin cupo, y el error de horario inexistente. **No cubierto:** que el repositorio falle al contar o al buscar duplicados.
- `TestJWT_*` / `TestValidarJWT_*`: que el token generado devuelva el mismo usuario y rol, y que se rechace uno de otro secreto o mal formado. **No cubierto:** el token vencido (habría que controlar el reloj).
- `TestAdminMiddleware_*`: el código HTTP (401/403/200) y si la cadena se aborta. **No cubierto:** `AuthMiddleware`, que hoy no usa ninguna ruta.
- Front: qué actividades devuelve el filtro, qué ruta y qué payload recibe la API (`toHaveBeenCalledWith`), que sin horarios no se llama al POST, y el estado de cada camino de `estadoCupo`. **No cubierto:** los componentes React y la conexión real con la API (eso es e2e, TP7).

---

<a id="tp6"></a>

# TP6 — CD: environments, aprobaciones y deployment patterns

Sección del TP6 del registro de decisiones. Sigue la convención de carpeta por TP de los prácticos anteriores. Como pide la consigna, no hay `evidencias.md`: cada decisión lleva al lado el link que la prueba.

## Enlaces de este TP

### (a) El artefacto: los dos paquetes públicos del registry

| Paquete | Link |
|---|---|
| Backend | https://github.com/users/SantiClark02/packages/container/package/ingsoft3-tp01-backend |
| Frontend | https://github.com/users/SantiClark02/packages/container/package/ingsoft3-tp01-frontend |

Cada imagen se etiqueta con el commit del merge que la produjo, con los 40 caracteres (`sha-<commit>`). Se baja sin credenciales desde cualquier máquina:

```
docker pull ghcr.io/santiclark02/ingsoft3-tp01-backend:sha-a3ec3646baa7962ddce35c5aeb8309241176a24f
docker pull ghcr.io/santiclark02/ingsoft3-tp01-frontend:sha-a3ec3646baa7962ddce35c5aeb8309241176a24f
```

### (b) La cadena que hace confiable al registry

| Eslabón | Link |
|---|---|
| Un PR con los tests en verde **no publica**: «Entrar al registry» aparece salteado | https://github.com/SantiClark02/ingsoft3-tp01/actions/runs/39545483115/job/119042588312 |
| En `main`, «Construir y publicar la imagen» es el **último paso** del job que corrió los tests | https://github.com/SantiClark02/ingsoft3-tp01/actions/runs/39552568767/job/119064858722 |

### (c) Los entornos

Uso el plan C de la cátedra (§3.6): los entornos corren en mi máquina, que es el runner self-hosted. Están vivos mientras el runner y Docker Desktop estén prendidos, y en la defensa se muestran en vivo.

| Entorno | Front | API | Base de datos |
|---|---|---|---|
| QA | http://localhost:3001 | http://localhost:8081 (`/health`) | MySQL propio, proyecto `gimnasio-qa`, volumen `mysql_data_qa` |
| PROD | http://localhost:3002 | http://localhost:8082 (`/health`) | MySQL propio, proyecto `gimnasio-prod`, volumen `mysql_data_prod` |

### Corridas y PRs que prueban el resto

| Qué prueba | Link |
|---|---|
| PR que agrega el CD (publicar al final, jobs `deploy-qa` / `deploy-prod`, compose por entorno) | https://github.com/SantiClark02/ingsoft3-tp01/pull/24 |
| Primer intento: el entorno **reconstruía** en mi máquina y quedó colgado más de 20 min (cancelado). Ver §5 | https://github.com/SantiClark02/ingsoft3-tp01/actions/runs/39546095950 |
| PR que cambia los entornos a **ejecutar la imagen publicada** | https://github.com/SantiClark02/ingsoft3-tp01/pull/25 |
| Flujo completo: QA automático → *Waiting for review* → aprobado → PROD con smoke verde | https://github.com/SantiClark02/ingsoft3-tp01/actions/runs/39552568767 |
| Aprobación con criterios escritos (PR #27: título de Login) | https://github.com/SantiClark02/ingsoft3-tp01/actions/runs/39637261606 |
| **Rechazo con motivo** (PR #28: subtítulo con un horario no confirmado) | https://github.com/SantiClark02/ingsoft3-tp01/actions/runs/39639882324 |
| **Rollback cronometrado** (re-run de `deploy-prod` hacia `dc0803b`) | https://github.com/SantiClark02/ingsoft3-tp01/actions/runs/39635626663/attempts/2 |
| Historial de deploys por entorno | https://github.com/SantiClark02/ingsoft3-tp01/deployments |

---

## 1. El artefacto se publica sólo si la verificación pasó

Cada job de build (`build-backend`, `build-frontend`) termina con dos pasos: «Entrar al registry» y «Construir y publicar la imagen». **Van últimos a propósito.** Los pasos de un job se cortan al primer error, así que si los tests de la etapa `test` fallan, el job muere antes y nunca se llega a publicar. La condición «si pasaron los tests» no está escrita en un `if`: está en el orden de los pasos.

La garantía sale de encadenar tres eslabones que ya existían:

1. **Nada entra a `main` sin verde.** Lo garantizan el PR obligatorio (TP1), el build (TP4) y los tests con su umbral de cobertura (TP5), todos como required checks.
2. **Sólo `main` publica.** `push: ${{ github.event_name == 'push' && github.ref == 'refs/heads/main' }}`. Un PR construye y verifica, pero no sube nada (link 1 de la tabla (b)).
3. **Publicar es el último paso del mismo job que testea** (link 2 de la tabla (b)). Si publicara un job aparte, tendría que volver a construir, y lo publicado sería otra construcción, no lo verificado.

**Si se publicara igual con los tests en rojo**, estar en el registry dejaría de significar «esto pasó la verificación». El registry pasaría a ser un depósito de cosas que alguien construyó alguna vez, y desplegar desde ahí sería apostar.

**Lo que esto NO garantiza:** la cadena asegura lo que el *pipeline* publica, no lo que físicamente puede entrar. Alguien con permiso de escritura podría hacer un `docker push` a mano y pisar una etiqueta `sha-…`, porque los tags son mutables. Lo único inmutable es el **digest** (`imagen@sha256:…`), que identifica el contenido exacto. Promover por digest en vez de por tag es la mejora siguiente.

**Permisos:** `packages: write` está declarado **por job** (sólo en los dos jobs que publican), no a nivel del workflow: es el mínimo privilegio. Para publicar no hace falta ningún secret mío: el login usa el `GITHUB_TOKEN` que GitHub crea para cada corrida y que muere con ella.

## 2. Continuous Delivery, no Continuous Deployment

- **CI** (TP4/TP5): cada cambio se integra y se verifica. Termina en un artefacto verificado.
- **Continuous Delivery** (lo que implementé): cada cambio verificado llega solo a QA y queda listo para PROD, pero el último paso lo autoriza una persona.
- **Continuous Deployment**: todo lo que pasa las verificaciones llega solo a PROD, sin humano.

**Por qué Delivery:** mi red de seguridad automática todavía no alcanza para sacar al humano.

- Los tests son unitarios (TP5).
- El smoke sólo pregunta «¿respondés?».
- Recién en el TP7 llegan la integración y las e2e.
- No tengo monitoreo ni alertas (TP9).

Con esa red, el Deployment automatizaría la propagación de errores. Para la tercera me faltaría: e2e como gate, métricas de error y latencia en PROD con alertas, y un rollback automático ante una alerta.

**Deploy ≠ release:** desplegar es poner una versión nueva a correr; release es exponer la funcionalidad a los usuarios. Con un feature flag podría desplegar el subtítulo de horarios del PR #28 apagado, y prenderlo recién cuando el gimnasio confirme el horario, sin otro deploy.

## 3. El diseño de la cadena y el alcance de cada secret

```
build-backend ─┐
               ├─▶ deploy-qa ─▶ [aprobación] ─▶ deploy-prod
build-frontend ┘   (automático)  (environment     (misma imagen)
                                  production)
```

| Pieza | Qué hace |
|---|---|
| `needs: [build-backend, build-frontend]` en `deploy-qa` | Sin CI verde (e imagen publicada) no hay deploy |
| `if: github.ref == 'refs/heads/main'` en `deploy-qa` | Los PRs verifican, sólo `main` despliega. `deploy-prod` no repite el `if`: depende de `deploy-qa`, así que la condición se hereda por la cadena |
| `environment: qa` | Le da al job los secrets de QA y suma el deploy al historial. Sin reglas: QA es automático a propósito |
| `needs: deploy-qa` + `environment: production` | PROD sólo después de QA vivo. El job **se pausa** («Waiting for review») hasta que el required reviewer (yo, con *Prevent self-review* desactivado) aprueba |
| `concurrency: { group: deploy-prod, cancel-in-progress: false }` | Evita dos deploys a PROD pisándose. **No** ordena la cola de aprobaciones: si dos corridas esperan, aprobar la vieja después de la nueva haría retroceder a PROD. La regla es rechazar la vieja a mano, con su motivo |
| La imagen `sha-${GITHUB_SHA}` | Cada corrida despliega **la imagen de su propio commit**, no «lo último de main». Aunque entre la espera y la aprobación entre otro merge, se aprueba y se despliega exactamente lo que QA probó |

**Secrets por environment, no en el repo.**

| Environment | Secrets |
|---|---|
| `qa` | `DB_PASSWORD`, `JWT_SECRET` |
| `production` | `DB_PASSWORD`, `JWT_SECRET` (valores distintos de QA) |

Un job sólo ve los secrets del environment que declara. Los de producción **sólo los lee un job aprobado hacia `production`**. Si estuvieran en el repo, cualquier job de cualquier rama (también el de un PR) podría leer la contraseña de la base de PROD.

En la guía los secrets son los *deploy hooks* de Render. En mi plan C son las credenciales de la base y el secreto de los JWT de cada entorno: quien los tenga puede operar ese entorno.

## 4. Qué mira el aprobador antes de aprobar (los criterios del gate)

No apruebo por reflejo. Antes de tocar *Approve* miro:

1. **QA verde de punta a punta en esa corrida:** `deploy-qa` en verde y el smoke diciendo «corre el commit <sha>» con el mismo commit de la corrida.
2. **Que el cambio se ve y funciona en QA** (http://localhost:3001): login con un usuario real, el listado carga y el cambio del PR aparece como se esperaba.
3. **Qué cambia este deploy:** el diff del PR. Si toca configuración, la base o algo que no se puede deshacer (datos), lo reviso con más tiempo.
4. **Que no haya otra corrida esperando** con un commit más nuevo. Si la hay, rechazo la vieja.
5. **Que el contenido esté validado:** un dato para los socios (un horario, un precio) no sale a PROD sin confirmar.

**Evidencia de los dos caminos:**

- **Aprobación** (run 39637261606): «QA OK: smoke verde con el commit correcto; revisé en localhost:3001 los dos títulos nuevos (Login y Home), login y listado funcionan. Apruebo.»
- **Rechazo** (run 39639882324, PR #28): «Rechazo: el horario del subtítulo (7 a 23 hs) no está confirmado por el gimnasio; no publico en PROD un dato para los socios sin validar. Queda en QA.» `deploy-prod` quedó en rojo y PROD no recibió el cambio. El subtítulo se sacó en el PR siguiente (el de este documento), así que ese contenido nunca llegó a PROD.

**Lo que mi aprobador NO puede ver:** errores que no aparecen en una prueba manual de dos minutos, carga real, ni el comportamiento con datos de producción. Para decidir mejor le faltan e2e automáticas contra QA (TP7) y métricas o logs de PROD (TP9). Un gate que nunca rechaza es decoración: por eso quedó la evidencia del rechazo.

## 5. Dónde corren los entornos: plan C y qué garantía conservo

**Por qué no Render + Neon:** Neon es PostgreSQL y mi app es **MySQL** (usa columnas `ENUM` y `FIELD()` en las consultas). Pasar a Postgres era reescribir el acceso a datos, y otros MySQL gratuitos en la nube piden tarjeta o tienen límites que no cubren la cursada. La cátedra acepta el **fallback local del §3.6** como entrega válida:

- El runner self-hosted es mi máquina.
- QA y PROD son **dos proyectos de Docker Compose separados** (`compose.qa.yml` / `compose.prod.yml`), cada uno con su `name:`, sus puertos y **su propia base** (volumen propio, contraseñas distintas).
- **Los levanta siempre una corrida de Actions, nunca yo a mano:** se ve en el log de `deploy-qa` / `deploy-prod`, sobre el runner `self-hosted`.

**Cómo cumplo los cinco puntos del contrato (§3):**

1. Dos entornos separados, QA y PROD, cada uno con su URL (localhost, por puerto).
2. Una base por entorno (dos contenedores MySQL con volúmenes distintos).
3. Front y back como contenedores, a partir de mis Dockerfiles: la imagen que construye el pipeline.
4. El deploy lo dispara el pipeline con el commit verificado; no hay auto-deploy.
5. PROD detrás de la aprobación del environment `production`.

**Qué pierdo respecto de la nube:**

- Las URLs no son públicas.
- El entorno depende de mi máquina: si se apaga, QA y PROD se caen juntos.
- No hay *cold start* real.
- Un runner self-hosted en un repo público es un riesgo: un fork malicioso podría intentar correr código en mi PC. Lo mitigo así:
  - **«Require approval for all external contributors»** en *Settings → Actions → General*.
  - Soy el único contributor.
  - Apago el runner cuando no lo uso.

**Qué garantía conservo.** La guía hace que Render **reconstruya** el commit en cada entorno. Así, lo que corre no es la imagen que aprobaron los tests, sino otra construcción del mismo código: puede salir distinta si cambió una dependencia o una imagen base.

Mi primer intento hacía lo mismo (`docker compose up -d --build`, run 39546095950): reconstruía en mi máquina, tardaba más de 20 minutos y se colgó. En el **PR #25** lo cambié por lo correcto: **los entornos bajan y ejecutan la imagen `sha-<commit>` que el pipeline verificó y publicó** (`docker compose pull` + `up -d`), sin construir nada. QA y PROD corren la **misma imagen** (*build once, deploy many*): se promueve lo mismo que se verificó, imagen y no sólo commit. Esto adelanta la Tarea 1 del TP7.

**La letra chica de mi «free tier»:**

- Los minutos de GitHub Actions son gratis (repo público) y el runner self-hosted no consume minutos.
- El límite real es mi máquina:
  - **Disco:** el primer deploy falló con 7 MB libres.
  - **Docker Desktop:** se colgó una vez y hubo que reiniciarlo.
  - Tienen que estar prendidos Docker Desktop y `run.cmd`.
- El pipeline lo maneja con un smoke con reintentos (30 × 10 s, para esperar a que MySQL arranque) y con un `exit /b 1` si el `pull` falla, para no levantar una versión vieja sin avisar.

## 6. Qué quedó por variable y qué quedó adentro de la imagen

**Front.** El `nginx.conf` del TP2 tenía escrito adentro `http://backend:8080`. Ahora es una **plantilla** (`frontend/default.conf.template`, copiada a `/etc/nginx/templates/`). Al arrancar, nginx reemplaza `${BACKEND_URL}` y `${DNS_RESOLVER}` con las variables del entorno.

- **Por variable:** la dirección del backend y el DNS.
- **Adentro de la imagen:** el `dist/` de Vite y la plantilla.

Los valores por defecto son los de compose, así que el `docker-compose.yml` del TP2 sigue andando sin tocarlo. **La misma imagen del front corre en QA y en PROD**, y cada una habla con su api.

**Back.**

- **Por variable:** `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `JWT_SECRET` y `APP_COMMIT` (el commit desplegado, que devuelve `/health`).
- **Adentro:** el binario compilado.

**Prueba de que cada entorno usa su propia base.** Inserté una actividad «SOY PROD» sólo en la base de PROD (`docker exec gimnasio-prod-mysql-1 mysql … INSERT …`). Aparece en http://localhost:3002 y **no** en http://localhost:3001.

## 7. Qué prueba el smoke test y qué no

Después de cada deploy, el job reintenta hasta 30 veces cada 10 segundos y verifica cuatro cosas:

1. `GET /health` responde: el proceso de la api está vivo.
2. `GET /activities` responde: la api **lee de su base**. Un `/health` solo daría verde con la base caída.
3. `GET /` del front responde: nginx sirve la app.
4. El `commit` que devuelve `/health` es **igual al de la corrida** (`GITHUB_SHA`); si no, falla. Es la mitigación que la guía propone para no dar verde contra la versión vieja.

**Lo que NO prueba:**

- Que la app **funcione** de punta a punta (login, inscribirse): eso es e2e, TP7.
- Que el front llegue a la api a través de su proxy `/api`: el smoke le habla a la api directo.
- Que la imagen sea bit a bit la publicada. `APP_COMMIT` llega por variable de entorno en el deploy, no horneada en la imagen. Prueba que el contenedor se recreó con la configuración de esta corrida; la identidad fuerte sería un `ARG` horneado en el build o el digest.

## 8. Deployment pattern para una producción real

**Elegiría blue-green.** Es un gimnasio con una sola instancia, tráfico chico y horarios pico conocidos (mañana temprano y después de las 18).

- **Cómo:** dos entornos PROD completos (azul y verde). Despliego la versión nueva en el que no recibe tráfico, la verifico con el smoke y las e2e, y el router cambia de uno al otro.
- **Costo:** 2× infraestructura, aceptable a esta escala.
- **Riesgo:** casi sin downtime; el switch es instantáneo.
- **Rollback:** volver el switch, en segundos.
- **Cuidado:** durante el cambio la base es compartida, así que los cambios de esquema tienen que ser compatibles hacia atrás (*expand/contract*).

**Por qué no canary:** necesita repartir un porcentaje de tráfico real y **métricas que digan «el canario sangra»** (errores, latencia por versión). Hoy no tengo observabilidad (llega en el TP9), así que un canary sería ruleta. Con pocos usuarios, además, un 5 % de tráfico es casi nadie.

**Feature flags, combinados:** para contenido que depende del negocio (como el horario del PR #28) desplegaría el código apagado y lo prendería al confirmarse. Los flags viejos hay que borrarlos, o son deuda técnica.

**Qué observabilidad me falta:** tasa de errores y latencia por endpoint, logs centralizados de PROD y alertas. Sin eso no puedo decidir automáticamente si una versión nueva está sana.

## 9. Plan de rollback (probado y cronometrado)

**Si un deploy aprobado sale mal:**

1. **Identifico la última versión buena.** Está en *Deployments → production* (qué commit estuvo en PROD antes) o en la release anterior. La imagen `sha-<ese commit>` sigue en el registry.
2. **Abro la corrida de ese commit** en Actions y re-corro sólo el job **`deploy-prod`**, no *Re-run all jobs*, que también redesplegaría QA.
3. **Apruebo**, porque el gate actúa igual: es un deploy a producción. El job baja la imagen vieja y recrea los contenedores.
4. **El smoke lo verifica solo:** `/health` tiene que devolver el commit viejo.

**Medido de verdad** (run 39635626663, attempt 2):

| Momento | Hora |
|---|---|
| Aprobé el rollback | 14:07 hora local (17:07 UTC) |
| El runner arranca `deploy-prod` | 17:07:59Z |
| `deploy-prod` termina con smoke verde | 17:08:34Z |

El job tardó **35 segundos** y el rollback completo, desde la aprobación, **menos de 1 minuto**. PROD volvió de `a3ec364` a `dc0803b`:

```
curl.exe -s http://localhost:8082/health
{"commit":"dc0803bec0fabcd816d0539ff7c531bbdfe325b7","status":"ok"}
```

Es rápido porque no se construye nada: la imagen anterior ya está en el registry. Esto ejercita la 4.ª métrica DORA, **tiempo de recuperación** (MTTR).

**Plan B si GitHub no está disponible** (*break-glass*): desde la máquina del runner, con las variables del entorno de PROD cargadas, correr `docker compose -f compose.prod.yml pull` y `up -d` con `GITHUB_SHA=<sha anterior>`. Es la misma operación, sin el gate. Por eso es sólo para emergencias y se documenta después.

**Lo que el rollback de código NO deshace: los datos.** La base de PROD vive en su volumen y no vuelve atrás con la imagen. Si una versión hubiera borrado o transformado datos (una migración destructiva), volver al binario anterior no los recupera. Para eso hacen falta backups de la base (un `mysqldump` antes de cada deploy a PROD) y migraciones compatibles hacia atrás. GORM `AutoMigrate` agrega columnas pero no las borra, lo que ayuda, pero no reemplaza un backup.

## 10. Versión de entrega

El práctico cierra con el tag **`v6.0.0`**, puesto sobre **el commit que está corriendo en PROD**, no sobre la punta de `main`. Se lo pregunto a PROD mismo (`/health` → `commit`) y lo confirmo en *Deployments → production*.

`v6.0.0` es la marca de cierre del TP6 (el número lo fija el práctico). No es una versión **SemVer** de producto: en SemVer, MAJOR rompe compatibilidad, MINOR agrega funcionalidad sin romper y PATCH arregla. Los títulos nuevos serían un MINOR (`v1.1.0`), no un MAJOR.

## 11. Problemas encontrados y cómo los resolví

| Problema | Solución |
|---|---|
| El primer `deploy-qa` falló: el disco C: tenía **7 MB libres** (aviso del runner) | Liberé espacio (caché de Docker, temporales). El deploy necesita varios GB para imágenes y volúmenes |
| Con `up -d --build` el entorno **reconstruía** todo en mi máquina; quedó colgado más de 20 minutos (run 39546095950) | PR #25: los entornos **ejecutan la imagen publicada** (`pull` + `up -d`). El deploy pasó a menos de 1 minuto, y se promueve exactamente lo verificado |
| Docker Desktop se colgó (`docker ps` no respondía y la app no abría) | Maté los procesos de Docker, `wsl --shutdown`, reabrí Docker Desktop y re-corrí el job fallido |
| Los paquetes de ghcr.io nacen privados, y el runner no los podía bajar sin login | Los pasé a **Public** (no tiene vuelta atrás). Ahora cualquiera los baja sin credenciales |
| En la corrida del PR #26 apreté **«Approve and deploy» con un texto de rechazo**; GitHub lo registró como aprobación (run 39635626663) | Lo detecté al revisar la corrida. Repetí el rechazo de verdad en el PR #28 (run 39639882324). Lección: el texto del comentario no decide nada, decide el botón, y después hay que verificar el resultado en la corrida |
| Los acentos de los datos semilla se ven mal («BÃ¡sica») | Problema previo del `db.sql`, que se carga con otra codificación. No afecta al TP6; queda anotado para corregir (`SET NAMES utf8mb4;` al inicio del script) |

## 12. Declaración de uso de IA

Usé **Claude (Anthropic)** para:

- analizar la consigna y elegir el plan C;
- escribir los jobs `deploy-qa` / `deploy-prod`, los `compose.qa.yml` / `compose.prod.yml`, la plantilla de nginx y el endpoint `/health`;
- diagnosticar los problemas (disco, deploy colgado, Docker Desktop);
- redactar el borrador de este documento.

**Cómo lo verifiqué:**

- Cada cambio pasó por PR con los checks en verde.
- QA y PROD se probaron a mano: login y listado en 3001 y 3002.
- La separación de bases se probó con el dato «SOY PROD».
- El gate se probó en los dos sentidos: aprobación y rechazo.
- El rollback se ejecutó y se cronometró de verdad.

**Qué pasa exactamente entre el merge y PROD:**

1. El merge dispara la corrida en `main`.
2. `build-backend` / `build-frontend` construyen la etapa de tests, corren tests + cobertura + umbral y, sólo si todo pasó, publican `sha-<commit>` en ghcr.io (último paso).
3. `deploy-qa` corre en mi runner con los secrets de `qa`: hace checkout de ese commit, `docker compose pull` de esa imagen, `up -d`, y el smoke reintenta hasta confirmar API + base + front y el commit correcto.
4. `deploy-prod` queda en *Waiting for review*: no corre ni lee sus secrets hasta que apruebo.
5. Al aprobar, hace lo mismo con `compose.prod.yml` y la **misma imagen**, y el smoke de PROD lo confirma.
6. Si rechazo, el job queda en rojo y PROD no se toca.

---

<a id="tp7"></a>

# TP7 — Contenedores en el pipeline + integración y e2e

Sección del TP7 del registro de decisiones. Sigue la convención de carpeta por TP de los prácticos anteriores. Como pide la consigna, no hay `evidencias.md`: cada decisión lleva al lado el link que la prueba.

## Enlaces del TP7

### (a) Los dos paquetes públicos del registry

| Paquete | Link |
|---|---|
| Backend | https://github.com/users/SantiClark02/packages/container/package/ingsoft3-tp01-backend |
| Frontend | https://github.com/users/SantiClark02/packages/container/package/ingsoft3-tp01-frontend |

Cada imagen tiene una sola etiqueta: `sha-<commit completo>`. No publico `latest`. Ejemplo con la imagen del fix (PR #33), que llegó a PROD:

```
docker pull ghcr.io/santiclark02/ingsoft3-tp01-backend:sha-234b445219b7890460930b1f4424071832016ab8
docker pull ghcr.io/santiclark02/ingsoft3-tp01-frontend:sha-234b445219b7890460930b1f4424071832016ab8
```

### (b) La corrida con la integración VERDE y la e2e ROJA (el gate frenando una promoción)

| Qué | Link |
|---|---|
| Commit que rompió la app (PR #32: el front manda `title` en vez de `titulo`) | https://github.com/SantiClark02/ingsoft3-tp01/commit/9e553360c3573aa8d44ff7514e48e9cc008cbed8 |
| La corrida: smoke verde, `integracion` verde, `e2e` roja, `deploy-prod` sin arrancar | https://github.com/SantiClark02/ingsoft3-tp01/actions/runs/39660829531 |
| Artefacto `playwright-report-integracion` (verde) | https://github.com/SantiClark02/ingsoft3-tp01/actions/runs/39660829531#artifacts |
| Artefacto `playwright-report-e2e` (rojo, con screenshot y traza) | https://github.com/SantiClark02/ingsoft3-tp01/actions/runs/39660829531#artifacts |

### (c) La corrida completa en verde posterior, hasta PROD

| Qué | Link |
|---|---|
| Fix (PR #33): build → QA → integración 3/3 → e2e 3/3 → aprobación → PROD con la misma imagen | https://github.com/SantiClark02/ingsoft3-tp01/actions/runs/39663244466 |
| Primera corrida completa del TP7 en verde (PR #31) | https://github.com/SantiClark02/ingsoft3-tp01/actions/runs/39653636939 |

### (d) Los entornos (plan C: runner self-hosted, vivos en la defensa)

| Entorno | Front | API |
|---|---|---|
| QA | http://localhost:3001 | http://localhost:8081 |
| PROD | http://localhost:3002 | http://localhost:8082 |

Uso el **plan C** de la cátedra (TP6 §3.6, TP7 §3.6). Por eso las dos suites corren **en el runner self-hosted**, que es donde vive QA, contra `localhost`. Las tres condiciones del plan C se leen en el repo:

- Los jobs `integracion` y `e2e` tienen `runs-on: self-hosted`.
- `compose.qa.yml` y `compose.prod.yml` usan `image:` del registry, no `build:`.
- El log de `deploy-qa` / `deploy-prod` muestra la imagen `sha-<commit>` de esa corrida: el `echo` del paso, el `pull` y `docker compose images`.

---

## 1. Build once, deploy many: qué problema del TP6 resuelve

Una reconstrucción no da lo mismo, aunque el código sea idéntico:

- Las imágenes base (`golang:1.23-alpine`, `node:20-alpine`, `nginx:1.27-alpine`) son **etiquetas que se mueven**: un build de mañana puede bajar otro parche.
- Cada build es otra hora y otro resultado posible.

Si PROD reconstruye, lo que llega a producción no es bit a bit lo que probaron QA y las pruebas.

**Mi ejemplo real** (TP6, corrida 39546095950): el primer deploy hacía `docker compose up -d --build` en cada entorno. Reconstruía todo en mi máquina, tardaba más de 20 minutos y se colgó. Desde el PR #25 los dos entornos **bajan y ejecutan la imagen `sha-<commit>`** que el pipeline construyó **una sola vez**, testeó y publicó. QA y PROD corren el mismo binario, y "promover" pasa a ser "ejecutá en PROD la imagen X que ya pasó por QA".

En la corrida del fix (39663244466) el log de `deploy-qa` y el de `deploy-prod` nombran **la misma imagen**, `sha-234b445…`.

El TP7 no le agregó nada al paso que publica: sigue siendo el último paso del job que testea (TP6 §3.0), así que lo publicado es lo verificado.

## 2. Estrategia de etiquetas

| Etiqueta | Dónde | Qué garantiza | Qué no |
|---|---|---|---|
| `sha-<commit>` (40 caracteres) | Registry (ghcr.io) | De qué commit salió la imagen. La usa cada deploy para decir **qué** imagen ejecutar | Es mutable: re-correr el build de ese commit la volvería a publicar, quizá con otro contenido |
| `v7.0.0` | Git (tag + release) | Nombra la versión que **anduvo en PROD**: el punto de cierre del práctico | No es una imagen: se llega a ella en un paso (ver §3) |
| digest `@sha256:…` | Registry | El contenido exacto: si cambia un byte, cambia | Hoy no lo uso (habría que pasarlo entre jobs); es la mejora siguiente |

**Por qué mi pipeline no publica `latest`:** es un puntero que se mueve con cada merge. Un deploy que lo nombrara ejecutaría "lo último que se publicó", haya pasado las e2e o no, y después no podría decir qué desplegué. Con una sola etiqueta por imagen no hay forma de desplegar algo sin nombrarlo exactamente.

**Con qué imagen están configurados los entornos:** en el plan C no hay una imagen "configurada" que quede vieja, como pasa en Render. Los compose dicen `image: …:sha-${GITHUB_SHA}` y **cada corrida** completa la etiqueta con su propio commit. No creé servicios nuevos: son los mismos dos proyectos de Compose (`gimnasio-qa`, `gimnasio-prod`) del TP6.

**Nunca vuelvo a correr un build viejo para desplegar**, porque republicaría la etiqueta. El rollback del TP6 re-corre **sólo** `deploy-prod`, que baja la imagen ya publicada.

## 3. De la release a la imagen, en un paso

```
git rev-list -n1 v7.0.0                 # → <commit>
ghcr.io/santiclark02/ingsoft3-tp01-backend:sha-<commit>
ghcr.io/santiclark02/ingsoft3-tp01-frontend:sha-<commit>
```

La release apunta al commit, y el pipeline publicó las dos imágenes de ese commit con ese nombre. Volver a una versión es ejecutar esa imagen.

## 4. Cómo se comprueba desde afuera que el entorno ejecuta mi imagen

**Lo que lo prueba:**

- El log de cada deploy imprime la imagen (`Imagenes a desplegar en QA: …:sha-<commit>`), hace el `pull` de esa etiqueta y lista con `docker compose images` el repositorio y el tag que quedaron corriendo. No hay `build` en ningún paso de deploy.
- En la máquina (en vivo en la defensa) se ve el tag de cada contenedor con `docker compose -p gimnasio-qa images` o `docker compose -p gimnasio-prod images`. Es el equivalente de los *Events* de Render.

**Lo que el smoke NO alcanza a probar:** el smoke pregunta "¿respondés?". También compara el `commit` de `/health` con el de la corrida, pero ese valor llega por variable de entorno en el deploy, no horneado en la imagen. Si el `up` fallara a mitad de camino, podría seguir respondiendo la versión anterior.

El cierre completo (concepto del §2.4, no implementado): pasar el commit como `ARG` en el build, que la API lo devuelva y que el smoke lo compare. Así, un deploy que no ocurrió pondría la corrida en rojo sin que nadie mire.

## 5. Las dos suites: qué elegí y por qué

**Integración** (`frontend/e2e/api.spec.js`, Playwright con el fixture `request`, sin navegador), contra la **API de QA** (`API_BASE_URL`):

| Prueba | Qué verifica |
|---|---|
| (a) Alta → GET la encuentra → DELETE → GET ya no está | Que la API y **su MySQL de verdad** se entienden. Además, que `day` vuelve como `Lun` después de pasar por la columna `ENUM` |
| (b) Título vacío → 400 y no se creó nada | El contrato del error, y que de verdad no escribió en la base |
| (c) Login: credenciales buenas → 200 + token + rol `administrador`; malas → 401 | Lo elegí yo. Es la puerta de todo: si se rompe, nadie entra y **me escriben todos**. Toca la base (tabla `usuarios` + bcrypt) |

**E2E** (`frontend/e2e/actividades.spec.js`, Chromium real), contra el **front de QA** (`E2E_BASE_URL`):

| Flujo | Tipo |
|---|---|
| Crear actividad → aparece en el listado → eliminarla → ya no está | Creación |
| Cupo 0 → el usuario ve el error (`role="alert"`) → no aparece en el listado | Error / validación |
| Un socio se inscribe → la ve en "Mis actividades" → se desinscribe → ya no está | Lo elegí yo: es **lo que el socio hace todos los días**. Si mañana no anda, me escriben los socios. Toca la base (inscripciones) |

**Cómo están escritas:**

- Cada prueba que crea datos usa un nombre único (`e2e <hora>`, `api <hora>`) y **borra lo que creó, comprobándolo**: QA es uno solo y lo comparten todas las corridas.
- Sin mocks (`page.route`, `fulfill`), sin `skip`.
- Buscan como una persona: `getByLabel`, `getByRole`.

**Qué tuve que agregar a la app para que las pruebas pudieran existir:**

- **Accesibilidad:** el formulario no tenía `label` asociados y los botones de cada fila no tenían un nombre propio. Agregué `htmlFor`/`id`, `aria-label` y `role="alert"`. El test estaba señalando un problema real de accesibilidad.
- **Una regla de negocio:** la API aceptaba actividades con título vacío o cupo 0. Agregué `reglas.ValidarActividad` (devuelve 400) con sus tests unitarios. La cobertura del backend subió a 16,8%.

**Qué NO puse en cada suite (pirámide):**

- **En e2e** no pruebo cada combinación de datos inválidos, ni el JWT, ni el cálculo de cupos: eso lo cubren los unitarios del TP5 en milisegundos. Tampoco el login fallido, que lo ve la integración sin abrir un navegador. Tampoco editar actividad: es menos crítico, y cada e2e cuesta segundos.
- **En integración** no pruebo la pantalla: no puede ver si el front usa bien la API.
- **En unitarios** no pruebo lo que contesta MySQL: el doble contesta lo que yo le digo.

## 6. Integración vs e2e: el par verde/rojo que diagnosticó mi rotura

**La rotura** (PR #32, cambio de una línea en `NuevaActividad.jsx`): el front pasó a mandar `title` en lugar de `titulo`. Compilaba, los unitarios del TP5 seguían en verde y el PR se mergeó.

**La corrida** (39660829531):

- smoke **verde**: el sistema respondía;
- `integracion` **verde**;
- `e2e` **roja**;
- `deploy-prod` **ni pidió aprobación**: su `needs: e2e` no se cumplió.

**Leído contra la tabla del §2.5, fila 1:** la integración le habla a la API directo con el campo correcto, y la API y su base guardaron y borraron sin problema, así que **la API y la base están sanas**. Quien no usa bien la API es **el front**. No hizo falta abrir el código.

**Cómo lo diagnostiqué con el reporte** (`playwright-report-e2e` de esa corrida):

- El *page snapshot* del fallo muestra el formulario con el campo **Título lleno** (`e2e 1792610860326 inscripcion`) y, arriba, el **alert "el título es obligatorio"**. Contradicción: el usuario completó el título y la app dice que falta.
- La traza (pestaña Network) muestra `POST /api/activities` con body `{"title":"e2e … cupo cero", …}` y respuesta **400 `{"error":"el título es obligatorio"}`**. El front manda el campo con otro nombre.

**Arreglo:** PR #33 vuelve a mandar `titulo` → cadena completa verde → aprobación → PROD (corrida 39663244466).

**Si hubiera roto la API**, la integración habría quedado roja y la e2e **no habría corrido** (`needs: integracion`): fila 2 de la tabla. No se gasta un navegador en confirmar algo que ya se sabe.

**Lo que sólo la integración atrapa:** si alguien cambiara en el backend el día a `"Lunes"` (texto completo), compilaría y los unitarios seguirían verdes, porque el doble acepta cualquier string. Pero la columna `ENUM('Lun',…)` de MySQL lo rechaza: el alta contesta 500 y la prueba de integración (a) se pone roja.

## 7. Por qué mi integración es la amplia (contra QA desplegado) y no la estrecha

- **Estrecha:** la API armada dentro del pipeline (`httptest`) con un MySQL descartable como `services:` del job. Llega antes (no necesita deploy) y está más aislada, pero es bastante más configuración y otra base que mantener.
- **Amplia (la mía):** le habla a la API de QA, que ya ejecuta la imagen exacta de esta corrida.

| | Amplia |
|---|---|
| **Gana** | Prueba además que el deploy quedó bien. Reutiliza QA. Es una décima parte de la configuración |
| **Pierde** | El feedback llega después del deploy. Depende de que QA y el runner estén vivos. Comparte la base con la e2e y con otras corridas (por eso cada prueba limpia lo suyo) |

## 8. Timeouts, reintentos y tests flaky

**Un test flaky** pasa a veces y falla otras con el mismo código. Es peor que no tener el test: enseña al equipo a ignorar el rojo.

**Mi configuración** (`playwright.config.js`):

| Ajuste | Para qué |
|---|---|
| `timeout: 60s` por test | Margen generoso: QA puede estar arrancando |
| `expect.timeout: 15s` por aserción | El default de 5 s es corto para un entorno real |
| *Auto-waiting* de Playwright, **sin `sleep`** | Espera lo que busca en vez de adivinar tiempos |
| `retries: 1` | Absorbe una demora suelta. Si un test pasa recién en el reintento, el reporte lo marca **flaky** y lo miro aunque la corrida esté verde |
| `workers: 1` | QA es uno solo: no corro dos pruebas a la vez sobre los mismos datos |
| `trace: on-first-retry` + `screenshot: only-on-failure` | Para diagnosticar un rojo remoto |

En el plan C no hay *cold start* del proveedor. El equivalente es MySQL arrancando después de un deploy, y lo absorben el smoke con reintentos (30 × 10 s), que corre antes que las suites, y estos timeouts.

## 9. La misma imagen del front en QA y en PROD

La dirección del backend **no está en la imagen**: nginx la lee de `BACKEND_URL` al arrancar (plantilla `default.conf.template`, TP6).

- **Adentro de la imagen:** el `dist/` de Vite y la plantilla.
- **Por variable:** `BACKEND_URL` y `DNS_RESOLVER`.

En el plan C cada entorno es un proyecto de Compose con su propia red, así que `http://backend:8080` resuelve al backend **de ese entorno**. Lo que importa es que la imagen es una sola (`…-frontend:sha-<commit>`) y corre igual en los dos.

## 10. Límites conocidos de mi cadena

- **QA compartido entre dos merges seguidos.** Si la corrida B redespliega QA mientras corren las suites de la A, la A podría dar rojo por algo que no es su código, o verde probando en parte la imagen de B.
  - **Cómo lo reconocería:** comparo la hora del `deploy-qa` de B con la de la e2e de A, y miro con `docker compose -p gimnasio-qa images` qué tag corre.
  - **Qué hago:** rechazo la A con motivo y me quedo con la B. Regla: **un merge por vez** mientras la cadena corre.
  - **La solución de un equipo real:** un entorno efímero por corrida.
- **Lo que mi gate no ataja:** flujos que no recorren mis pruebas (por ejemplo, editar actividad), problemas visuales o de rendimiento, y lo que sólo pasa con datos o carga de producción. También el rol socio puro: las e2e entran como `admin` (las credenciales de prueba del README).

## 11. Problemas encontrados y cómo los resolví

| Problema | Solución |
|---|---|
| QA vive en `localhost` de mi PC: un runner de GitHub no lo alcanza | Los jobs `integracion` y `e2e` corren `runs-on: self-hosted`, como pide el plan C |
| Primera corrida (#34 del workflow, PR #30): la e2e falló con `Executable doesn't exist … headless_shell.exe`. **Fue infraestructura, no la app** | En `cmd`, `npm` y `npx` son `.cmd`: llamarlos sin `CALL` termina el script, así que `npx playwright install` nunca corría. PR #31: la instalación del navegador va en su propio paso. Esa corrida **no** cuenta como evidencia del gate |
| La API aceptaba título vacío (la prueba (b) exige 400) | Regla `ValidarActividad` + tests unitarios (TP5 sigue verde, cobertura 16,8%) |
| El formulario no tenía labels asociados (`getByLabel` no encontraba nada) | `htmlFor`/`id`, `aria-label` en botones por fila y `role="alert"` en el error |
| vitest toma por defecto los `*.spec.js` y rompería `build-frontend` | `exclude: [...configDefaults.exclude, 'e2e/**']` en `vite.config.js` |
| Avisos `Failed to save … tar.exe exit code 2` en los jobs del runner | Es la caché de npm de `setup-node` en Windows: no bloquea, sólo hace `npm ci` más lento |

## 12. Declaración de uso de IA

Usé **Claude (Anthropic)** para:

- analizar la consigna;
- escribir las dos suites de Playwright, la config, los jobs `integracion` / `e2e` y el cambio de `needs` del gate;
- escribir la regla `ValidarActividad` con sus tests y los cambios de accesibilidad del front;
- diagnosticar el fallo de instalación del navegador y redactar el borrador de este documento.

**Cómo lo verifiqué:**

- Las 6 pruebas pasaron contra una copia local de la app antes de subirlas, y después en el pipeline contra QA (corridas 39653636939 y 39663244466).
- El gate se probó rompiendo la app de verdad (PR #32) y viendo la integración verde y la e2e roja.
- El diagnóstico salió del reporte y de la traza, no del código.

**Qué verifica cada test, contra qué entorno y qué pasa si falla:**

- **Integración:** contra la API de QA (`localhost:8081`) y su MySQL. Si falla, la e2e no corre y PROD no se pide.
- **E2E:** contra el front de QA (`localhost:3001`), pasando por nginx → API → MySQL. Si falla, `deploy-prod` no pide aprobación.
- El reporte de cada suite se publica como artefacto aunque falle, y el Summary muestra `passed`/`failed`/`flaky`/`skipped`.

---

<a id="tp8"></a>

# TP8 — Infraestructura como Código

Sección corta, como pide la consigna: riel y versiones, problemas, y uso de IA. La prueba del práctico no es un texto: es `preprod` naciendo con `terraform apply` en la defensa.

## 1. Riel elegido y versiones fijadas

**Riel: plan C de la cátedra (§3.7), Terraform + provider Docker, 100 % local.** Es coherente con el TP6 y el TP7, donde QA y PROD ya corren como proyectos de Docker Compose en mi máquina. Además, mi app usa **MySQL**: el riel de Neon es Postgres.

`infra/` declara un entorno **nuevo**, `preprod`, que no existía. QA (`compose.qa.yml`) y PROD (`compose.prod.yml`) no se tocan, no se importan y no se migran. Los tres conviven en la misma máquina:

| Entorno | Front | API | Quién lo crea |
|---|---|---|---|
| QA | 3001 | 8081 | el pipeline (`deploy-qa`) |
| PROD | 3002 | 8082 | el pipeline (`deploy-prod`, con aprobación) |
| **preprod** | **3003** | **8083** | **`terraform apply` desde `infra/`** |

| Pieza | Versión | Cómo se fija |
|---|---|---|
| Terraform | v1.16.6 (`required_version = ">= 1.5"`) | instalado con `winget install Hashicorp.Terraform` |
| `kreuzwerker/docker` | **v4.6.0** | `version = "~> 4.0"` + `.terraform.lock.hcl` (commiteado) |
| `hashicorp/random` | **v3.6.3** | `version = "3.6.3"` + `.terraform.lock.hcl` |

**Qué declara `infra/main.tf`** (10 recursos): red `gimnasio-preprod`, volumen `gimnasio-preprod-mysql`, las tres imágenes (`mysql:8.0` y la API y el front del TP7 por su `sha-<commit>`), los contenedores `db`, `api` y `front`, y dos `random_password`.

**Decisiones de diseño:**

- **Ninguna credencial escrita.** La contraseña de MySQL y el secreto JWT los **genera** Terraform (`random_password`): no hay tfvars ni variables de entorno que cargar. Viven sólo en el estado, que es exactamente por qué el estado no se commitea. Lo comprobé abriéndolo: el `"result"` está en texto plano.
- **Todo se deriva.** `DB_HOST` = el nombre del contenedor de **su** base, y `BACKEND_URL` = el nombre del contenedor de **su** API. No se copia ni pega nada, que fue el error más caro del TP6.
- **La misma imagen del TP7.** No se construye nada: preprod hace pull de la imagen `sha-<commit>` publicada por el pipeline. `APP_COMMIT` se deriva del tag, así que `/health` dice qué imagen corre.
- **"Arrancó" no es "está lista".** La base tiene un `healthcheck` por TCP y `wait = true`, así que la API se crea recién cuando MySQL acepta conexiones. Es el equivalente del `service_healthy` de compose (TP2).
- **`keep_locally = true` en las imágenes.** `destroy` no borra `mysql:8.0` ni las imágenes que QA y PROD también usan.
- **Estado local.** Sin bloque `backend`: `infra/terraform.tfstate` vive sólo en mi máquina y el `.gitignore` (`*.tfstate`, `*.tfstate.*`, `.terraform/`, `*.tfvars`) lo deja afuera desde antes del primer `apply`. Si lo pierdo, no pierdo preprod: pierdo la prueba de que es mío, y el próximo plan querría crear todo de nuevo. Si tuviera que aplicar otra persona o el pipeline, el estado tendría que mudarse a un backend remoto con bloqueo.

**Qué comprobé en mi máquina:**

- `plan`: 10 to add. `apply`: la base tardó 46 s en estar *healthy* y después se crearon la API y el front.
- `/health` devolvió el commit de la imagen, `/activities` respondió desde **su** base y el front cargó en 3003 contra **su** API.
- **Idempotencia:** el segundo `plan` dio «No changes».
- **Drift:** borré a mano `gimnasio-preprod-api`. El `plan` lo detectó (`docker_container.api will be created · 1 to add`) y el `apply` lo reparó.
- **Lo no declarado es invisible:** creé a mano el volumen `base_intrusa` y el `plan` siguió en «No changes». Lo borré con la misma mano. Lo mismo pasa con QA y PROD: existen, corren y ningún `plan` habla de ellos.
- **Ciclo de vida:** `terraform destroy` borró preprod entero (sin tocar los contenedores de QA, PROD ni las imágenes compartidas) y `terraform apply` lo devolvió idéntico.

## 2. Problemas y cómo los resolví

| Problema | Solución |
|---|---|
| **Preprod encontró un bug que QA no veía.** Con la base nueva, la API devolvía `null` en vez de `[]` en `/activities`, y el front quedaba en blanco al hacer `.length` sobre `null`. QA y PROD siempre tienen datos de semilla, así que nunca se vio | Arreglo en el backend (`out := []domain.ActivityListItem{}`) por PR, con la cadena completa del TP7. Después actualicé `image_tag` en `infra/variables.tf` al `sha-` de ese merge. El `plan` mostró el **reemplazo** de las imágenes y los contenedores de API y front (en el plan C se puede modificar, a diferencia de Render) y el `apply` lo aplicó. Es el argumento de preprod: un entorno que nace vacío prueba cosas que un QA sucio no |
| El provider de Docker, por defecto, busca un socket Unix; Docker Desktop en Windows escucha en un *named pipe* | `host = var.docker_host`, con default `npipe:////./pipe/docker_engine` |
| Una API que arranca antes que MySQL se cae (el clásico del TP2) | `healthcheck` por TCP en la base + `wait = true` / `wait_timeout = 240` |
| `destroy` podría borrar imágenes compartidas con QA y PROD | `keep_locally = true` en los `docker_image` |
| El estado tiene secretos | El `.gitignore` se actualizó antes del primer `apply`. Verifiqué con `git status` que sólo viajan los `.tf`, el `README` y el `.terraform.lock.hcl` |

**Lo que sigue siendo artesanal en mi cuenta:** QA y PROD siguen viviendo en `compose.*.yml` levantados por el pipeline, no en Terraform. El runner self-hosted y los environments de GitHub se configuraron a mano. El siguiente paso sería **importarlos** (bloques `import`) y extraer un **módulo** "entorno" que se instancie tres veces.

## 3. Declaración de uso de IA

Usé **Claude (Anthropic)** para:

- analizar la consigna y elegir el plan C;
- escribir `infra/main.tf`, `variables.tf`, `outputs.tf` y el `README` de `infra/`;
- diagnosticar el `null` de `/activities` que dejaba el front en blanco;
- redactar el borrador de esta sección.

**Cómo lo verifiqué:** corrí yo `init`, `fmt`, `validate`, `plan` y `apply`, y leí los planes antes de confirmar. Probé `/health`, `/activities` y el front de preprod. Hice los experimentos de idempotencia, estado, drift y recurso no declarado, y el ciclo `destroy` → `apply`, comprobando con `docker ps` que QA y PROD seguían corriendo.

**Puedo leer el plan en voz alta:**

- `+` es crear y `(known after apply)` es un valor que asigna Docker al crear (ids, red).
- La última línea (`N to add, 0 to change, 0 to destroy`) es lo primero que se lee.
- El orden de creación lo deduce Terraform del grafo de referencias: red, volumen e imágenes → base → API → front.
- El `destroy` lo recorre al revés: front → API → base → volumen y red.

---

<a id="tp9"></a>

# TP9 — DevSecOps + Observabilidad + Continuous Feedback

Sección corta, como pide la consigna: el PR de Dependabot, el monitor y Sentry, y los problemas con el uso de IA. Lo que se evalúa se ve navegando: la pestaña **Security**, los pull requests, el historial del monitor y Sentry.

## 1. El PR de Dependabot que mergeé y por qué me animé

**Configuración.** `.github/dependabot.yml` vigila tres ecosistemas con chequeo semanal: `gomod` (`/backend`), `npm` (`/frontend`) y `github-actions` (`/`). Llegó a `main` por el PR #37, y en Settings → Advanced Security prendí *Dependabot alerts* y *Dependabot security updates*. A los minutos del merge Dependabot abrió **20 PRs** (#38 a #57), y la pestaña Security mostró **104 alertas** abiertas.

**Mergeé el #46:** `github.com/golang-jwt/jwt/v5` de **5.2.2 a 5.3.1**, en el backend.

- **SemVer:** es **minor**. Trae funciones nuevas y correcciones, y el changelog no anuncia cambios que rompan nada. La 5.3.0 declara **Go 1.21** como mínimo, que es la versión de mi `go.mod`.
- **Por qué me animé: lo que toca está probado.**
  - `utils/jwt_test.go` (TP5) prueba generar y validar el token y el `AdminMiddleware`.
  - La suite de integración y las e2e (TP7) hacen **login real con JWT** contra QA.
  - Si la librería cambiaba algo, alguna de esas pruebas quedaba roja.
- **La evidencia es la cadena entera en verde** ([run 39700946767](https://github.com/SantiClark02/ingsoft3-tp01/actions/runs/39700946767)): build, tests y cobertura → deploy-qa → integración → e2e → aprobación → PROD.

**Qué descarté y por qué:**

- **Los major** (vitest 5, eslint 10, @vitejs/plugin-react 6, `actions/checkout` 7, `setup-node` 7…): pueden traer cambios que rompen, y quedan para revisar uno por uno.
- **`sentry-go` 0.50:** pide una versión de Go más nueva que la que compila mi imagen.

**Sobre las 104 alertas.** Varias no son explotables en mi PROD:

- `vite` es la herramienta de desarrollo, y en PROD nginx sirve el build estático.
- `golang.org/x/crypto` reporta una falla en su paquete SSH, que no uso.

Que no sean explotables no las vuelve irrelevantes: son el backlog de actualización.

## 2. El monitor y lo que vi en Sentry

**Monitor externo.** Uso UptimeRobot, tipo HTTP(s), con intervalo de 5 minutos y aviso por mail, sobre el `/health` de PROD.

En el plan C, PROD está en `localhost`, así que la expuse con un **túnel de Cloudflare** (`cloudflared tunnel --url http://localhost:8082`). El túnel expone **sólo la API de PROD** y sólo mientras dura el ejercicio.

| Evento | Hora |
|---|---|
| `docker stop gimnasio-prod-backend-1` (caída provocada) | 21:45:27 |
| Mail **DOWN** | 21:47 |
| `docker start gimnasio-prod-backend-1` | 21:48:19 |
| Mail **UP** | 21:52 |

**El monitor tardó unos 2 minutos en avisarme de la caída.** El peor caso es el intervalo, unos 5 minutos. Sin el monitor me habría enterado cuando un usuario se quejara. Al terminar **pausé el monitor** y cerré el túnel.

**Sentry** está en el back y en el front de PROD, con entorno `produccion`:

- **Error:** «Falla de prueba del TP9» en `GET /falla`, **2 eventos**, con archivo y línea. El endpoint era temporal y se saca en el PR de cierre.
- **Operación más usada:** `GET /activities`, con **0,0187 TPM**, p50 21,67 ms y p95 82,99 ms.
- **Operación más lenta:** `POST /users/login`, con **p95 396,59 ms** y p50 376 ms. La causa es **bcrypt**: es lento a propósito, para que la fuerza bruta sea cara. No se «optimiza» bajando el costo.

Con ese dato abrí la **tarjeta en el tablero** del TP3 (*IngSoft3 - Gimnasio*): revisar el login con su p95 medido, confirmar que el tiempo es de bcrypt y evaluar *rate limiting* en `/users/login`.

## 3. Problemas que tuve y cómo los resolví

| Problema | Solución |
|---|---|
| PROD es local y un monitor externo no llega a `localhost` | Túnel `cloudflared` sin cuenta, sólo hacia la API de PROD (8082), abierto mientras dura la prueba |
| UptimeRobot gratis pregunta con `HEAD` y Gin contestaba **404** | `router.HEAD("/health", …)` con el mismo handler. `curl -I` da 200 |
| La última versión del SDK de Sentry para Go pide un Go más nuevo | Fijé `sentry-go` **v0.33.0**, la última compatible con Go 1.21. Generé el `go.sum` con un contenedor `golang:1.23`, sin instalar Go |
| El DSN del back no tiene que estar en el código, y QA no tiene que ensuciar Sentry | *Environment secret* `SENTRY_DSN` **sólo en `production`**. Sin DSN (local, tests, QA, preprod) el SDK queda apagado. El DSN del front sí va en el código, porque no es secreto: sirve para enviar, no para leer |
| Los tres entornos son `localhost` | El front informa el entorno según el puerto: 3001 qa, 3002 produccion, 3003 preprod |
| **CodeQL**: *Workflow does not contain permissions* (4 alertas en `ci.yml`, categoría OWASP *Broken Access Control* / mínimo privilegio) | `permissions: contents: read` global. Los builds siguen pidiendo `packages: write` en su bloque. Las 4 quedaron **cerradas** con el #37 |
| **CodeQL**: *Clear-text logging of sensitive information* (High, `servicies_login.go:27`). El login imprimía **la contraseña ingresada y el hash**. Era el bug #14 del TP3, que seguía vivo. **Es explotable:** cualquiera con acceso a `docker logs` o al runner veía contraseñas en texto plano (OWASP A09, *Security Logging and Alerting Failures*) | Saqué esas dos líneas en el PR de cierre |
| Strict mode: cada merge deja desactualizados los PRs de Dependabot | `@dependabot rebase`, y un merge por vez (lección del TP7) |

## 4. Declaración de uso de IA

Usé **Claude (Anthropic)** para:

- analizar la consigna y armar el plan;
- escribir `dependabot.yml`, la integración de Sentry (back y front), el `HEAD /health` y los permisos del workflow;
- elegir qué PR de Dependabot mergear;
- analizar los hallazgos de CodeQL;
- redactar esta sección y el `postmortem.md`.

**Cómo lo verifiqué:**

- Cada cambio pasó por PR con la cadena completa en verde y aprobación de PROD.
- `curl -I` sobre `/health` devolvió 200.
- En Sentry vi el error y las transacciones con sus tiempos.
- Provoqué la caída y recibí el DOWN y el UP, con sus horas.
- Leí el changelog del #46 antes de mergearlo.
- Comprobé que CodeQL cerró las alertas de permisos.
