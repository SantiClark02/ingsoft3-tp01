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
