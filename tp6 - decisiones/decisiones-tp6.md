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
