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
