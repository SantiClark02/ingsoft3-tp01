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
