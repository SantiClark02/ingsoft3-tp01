# Postmortem: la API de PROD dejó de responder — 2026-10-21

> Incidente **simulado** (práctica del TP9 §3.7), sin culpables. La pregunta no es «¿quién fue?» sino «¿qué hizo posible que esto pasara y tardara en verse?».

## Resumen

El contenedor del backend de PROD (`gimnasio-prod-backend-1`) se detuvo y la API dejó de responder durante **unos 3 minutos** (21:45:27 a 21:48:19). Lo detectó el **monitor externo**, no un usuario. Se resolvió volviendo a levantar el contenedor, con la misma imagen y sin pérdida de datos.

## Impacto

- **Qué se cayó:** la API de PROD (`:8082`). El front de PROD (`:3002`) cargaba, pero sin datos y sin login: toda operación devolvía error.
- **Cuánto tiempo:** unos 3 minutos de caída real. El monitor la registró entre el DOWN (21:47) y el UP (21:52).
- **Datos:** sin pérdida. MySQL de PROD siguió corriendo, con su volumen intacto.
- **Usuarios afectados:** ninguno real, porque es un incidente simulado. En un sábado real habrían sido todos los socios que quisieran inscribirse.

## Timeline

| Hora | Qué pasó |
|---|---|
| 21:45:27 | Se detiene el backend de PROD (`docker stop`) desde la misma terminal con la que se trabaja en QA, preprod y desarrollo |
| 21:47 | **Detección:** llega el mail «Monitor is DOWN» de UptimeRobot (unos 2 min después) |
| 21:47–21:48 | **Diagnóstico:** `docker ps` muestra que falta `gimnasio-prod-backend-1`. La base sigue *healthy* y no hubo deploys en ese rato (historial de Actions) |
| 21:48:19 | **Mitigación:** `docker start gimnasio-prod-backend-1`, con la misma imagen `sha-<commit>` |
| 21:52 | **Resolución confirmada:** llega «Monitor is UP» |

## Causa raíz (del sistema, no de una persona)

1. **PROD comparte máquina y daemon de Docker con QA, preprod y desarrollo.** Cualquier comando manual de `docker` puede tocar PROD sin pasar por ningún gate. El pipeline protege *cómo se despliega*, pero nada protege *lo que ya está corriendo*.
2. La política `restart: unless-stopped` **a propósito no** levanta un contenedor detenido a mano: un stop manual queda caído hasta que alguien lo note.
3. Antes del TP9 **nada avisaba**: no había monitor externo y la caída sólo se habría visto por la queja de un usuario.

## Acciones

| Acción | Tipo | Dueño | Estado |
|---|---|---|---|
| Monitor externo sobre `/health` de PROD, con alerta por mail | Detección | Pato | Hecho (TP9), pausado fuera del ejercicio |
| Sentry en back y front de PROD, para ver errores y demoras que el monitor no ve | Detección | Pato | Hecho (TP9) |
| PROD en un host separado (o en el proveedor de nube), fuera de la máquina de desarrollo | Prevención | Pato | Pendiente: es lo que resuelve la causa 1 |
| Runbook «llegó un DOWN»: `docker compose -p gimnasio-prod ps` → `logs backend` → `up -d` | Recuperación | Pato | Pendiente (5 min de escribir) |
| Regla de alerta en Sentry para errores nuevos en `produccion` | Detección | Pato | Pendiente |

## Aprendizajes

- **Detectar tardó unos 2 minutos con el monitor.** Sin él, el tiempo de detección era «hasta que alguien se queje».
- **El UP tarda más que el DOWN en llegar** (unos 4 min): el monitor confirma en su próximo chequeo. El tiempo de recuperación real (21:48) no es el que dice el mail (21:52).
- **«Está viva» no es «anda bien»:** el monitor sólo mira `/health`. Un login roto con la API viva no lo detectaría; eso lo ve Sentry.
- La pregunta útil fue **qué permitió el error** (PROD al alcance de cualquier comando), no quién lo ejecutó.
