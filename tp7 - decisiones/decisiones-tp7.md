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
