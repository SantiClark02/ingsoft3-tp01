---

# TP3 — Planificación y trazabilidad

Proyecto público: https://github.com/users/SantiClark02/projects/1

## 1. Jerarquía elegida

```
Épica: Automatizar la verificación y entrega de la aplicación
  └── Historia: Como desarrollador quiero que cada Pull Request se verifique
                automáticamente para no integrar código que no compila
        ├── Tarea: Escribir el workflow de CI que construye backend y frontend
        └── Tarea: Configurar el pipeline como requisito obligatorio de merge

Bug (al costado, sin colgar de nadie): El backend registra las contraseñas
     en texto plano en los logs
```

**Por qué el bug no cuelga de la historia.** Un bug es un defecto de algo que **ya existe** y
que se comporta distinto de lo esperado. No es trabajo planificado dentro de un objetivo: es
trabajo no previsto que aparece. Colgarlo de la historia distorsionaría el pronóstico del
sprint, porque haría parecer que ese esfuerzo estaba planificado cuando no lo estaba.

La forma de distinguirlo: si el comportamiento nunca existió, es trabajo que faltaba hacer
(una tarea). Si existía y dejó de funcionar como se esperaba, es un bug. En este caso el
backend funciona, pero escribe las contraseñas en texto plano en su salida — un
comportamiento existente e incorrecto.

Este bug es real: se detectó durante el TP2 al leer `docker compose logs backend`, y está
documentado en la sección de observaciones de ese práctico.

**Jerarquía vs etiquetas.** Los tipos de issue nativos de GitHub (Epic, Story, Task) solo
están disponibles en cuentas de organización. En una cuenta personal como la de este trabajo,
la convención es usar etiquetas: se crearon `epic`, `story` y `task`, y se usó la `bug` que
GitHub trae de fábrica. La jerarquía navegable se arma con sub-issues.

## 2. Criterios de aceptación: por qué son verificables

Los tres criterios de la historia están escritos de forma que cualquiera pueda comprobarlos sin
interpretar:

1. Al abrir un PR contra `main`, se ejecuta un pipeline que construye ambas imágenes.
2. Si algún build falla, el botón de merge queda bloqueado.
3. El estado de la última ejecución es visible en el README mediante un badge.

Cada uno se verifica mirando algo concreto: la pestaña de checks del PR, el estado del botón de
merge, y el README.

En cambio, un criterio como *"que el CI funcione bien"* no sirve, porque **"bien" no se puede
comprobar**: dos personas pueden mirar lo mismo y estar en desacuerdo sobre si se cumplió. Un
criterio de aceptación tiene que poder responderse con sí o no, sin debate.

## 3. Duración del sprint: 2 semanas

En la industria se usan sprints de una a cuatro semanas. Se eligieron **dos** por tres razones:

- **Feedback razonablemente rápido.** Si algo se planificó mal, el error se detecta en dos
  semanas y no en un mes.
- **Espacio para trabajo con sentido.** En una semana, el tiempo de planificar y revisar
  consume una proporción alta del sprint. En dos, la ceremonia pesa la mitad.
- **Se alinea con el calendario de entregas de la materia**, que es el ritmo real al que se
  está trabajando. Un sprint que no coincide con las fechas en las que efectivamente hay que
  entregar algo es una ficción.

Lo que el equipo se compromete a cumplir en un sprint es el **objetivo**, no la lista de items:
esa lista es un pronóstico y se renegocia si hace falta.

## 4. Límite de trabajo en progreso: 2

La regla de arranque es **cantidad de personas + 1**. Trabajando solo, da 2.

**Por qué el "+1":** es la válvula para cuando algo queda esperando —una revisión, una
respuesta, un pipeline corriendo— y hace falta avanzar en otra cosa sin quedarse bloqueado.
Pasarse de ahí hace que el límite deje de limitar: con un número alto nunca se alcanza y no
cumple ninguna función.

**Por qué existe el límite.** El trabajo empezado y no terminado no es productividad: es
inventario, y el inventario tiene costo. Más cambio de contexto, más ramas que envejecen, más
conflictos al integrar. La idea es empezar menos para terminar más.

**Qué hace y qué no.** GitHub pone el contador de la columna en rojo cuando se supera, pero
**no lo impide**. Es un acuerdo de equipo hecho visible, no un candado de la herramienta. Esa
distinción importa: el límite funciona porque el equipo decide respetarlo.

**Qué me haría subirlo:** que el trabajo quede bloqueado sistemáticamente por dependencias
externas (esperando revisiones ajenas, entornos, respuestas) y no por capacidad propia.
**Señal de que quedó demasiado alto:** que nunca se alcance. Un límite que nunca se toca no
está limitando nada.

## 5. Diagnóstico de la historia mal escrita

Ejemplo del anti-patrón: *"Como desarrollador quiero crear la tabla usuarios."*

**Por qué está mal escrita.** Tiene la forma de una historia pero es una tarea disfrazada.
Nadie *quiere* una tabla: una tabla es un medio, no un fin. Además carece de beneficio — no
dice para qué sirve — y el rol está mal elegido: describe a quien implementa, no a quien recibe
valor. Una historia debe expresar valor para un usuario, no un paso técnico de la solución.

**Cómo la reescribiría.** *"Como visitante quiero registrarme con mi email y contraseña para
poder inscribirme a las actividades del gimnasio."* Crear la tabla `usuarios` pasa a ser una de
las tareas que la implementan, que es su lugar correcto.

## 6. Automatización del tablero

Se activaron dos workflows del Project: mover a `Todo` al agregar un item, y mover a `Done` al
cerrarlo.

**Un detalle importante:** esos workflows actúan sobre el estado **propio** de cada item. No
cierran la historia porque se hayan cerrado sus tareas. La barra de progreso llega a 2/2 y la
historia sigue abierta hasta cerrarla explícitamente. La jerarquía informa, no decide.

## 7. Trazabilidad

El mecanismo es referenciar el issue desde la descripción del Pull Request con `Closes #N`. Al
mergear, GitHub cierra el issue y deja el enlace permanente en ambas direcciones: desde la
tarea se navega al PR y a sus commits, y desde ahí se sube a la historia y a la épica.

Responde dos preguntas espejo: *"¿por qué existe este cambio?"* (del código al requerimiento) y
*"¿este requerimiento ya está implementado, y en qué commit?"* (del requerimiento al código).
Parece un detalle de forma, pero es infraestructura de auditoría.

## 8. Qué aporta GitHub Projects que no aporta un Trello

Un tablero genérico gestiona tarjetas; GitHub Projects gestiona **el mismo objeto que el
código**. Las tarjetas *son* los issues del repositorio, así que el estado del trabajo y el
trabajo mismo no pueden desincronizarse. El vínculo issue ↔ PR ↔ commit es nativo y
bidireccional, y la automatización reacciona a eventos reales del repositorio: cerrar un PR
mueve la tarjeta sin que nadie la arrastre. En un tablero externo esa correspondencia depende
de que alguien la mantenga a mano, y por eso siempre termina desactualizada.

## 9. Problemas encontrados

**El primer Pull Request se mergeó sin `Closes #N` en la descripción.** Se creó el PR con el
título por defecto y sin cuerpo, así que al mergear el issue no se cerró automáticamente y no
quedó el vínculo.

*Cómo se resolvió:* como un PR mergeado no se puede editar para que dispare el cierre
automático, se vinculó manualmente comentando en el issue con una referencia al número del PR
—lo que crea el enlace cruzado en ambas direcciones— y se cerró el issue a mano. El resultado
navegable es el mismo; lo que se perdió es la automatización.

*Qué aprendí:* el `Closes #N` va en la **descripción** del PR, no en el título ni en el mensaje
del commit, y conviene escribirlo antes de crear el PR y no después.

---

# TP4 — Integración continua

## 1. Estructura del pipeline

Un workflow (`.github/workflows/ci.yml`) con **dos jobs independientes que corren en paralelo**:
`build-backend` y `build-frontend`.

**Por qué esos jobs.** Son las dos unidades desplegables de la aplicación. Cada una tiene su
propio Dockerfile y su propio contexto de build.

**Por qué en paralelo.** Son independientes: construir el backend no necesita nada del
frontend. Ejecutarlos en serie duplicaría el tiempo sin ganar nada.

**Qué NO comparten dos jobs.** Cada uno arranca en una **máquina virtual limpia y separada**.
No comparten sistema de archivos, ni memoria, ni variables de entorno, ni nada de lo que el
otro haya construido. Lo único común es el repositorio, que cada uno clona por su cuenta. Si
un job necesitara algo producido por otro, habría que pasarlo explícitamente como artefacto —
y eso, además, los volvería secuenciales.

**Disparadores elegidos:**

- `pull_request` hacia `main`: es el más importante. Verifica **antes** del merge, sobre el
  resultado propuesto, y es lo que alimenta el gate del PR.
- `push` a `main`: verifica lo que efectivamente quedó integrado. Con el gate activo rara vez
  sorprende —lo que iba a romper ya lo frenó el PR— pero es lo que mantiene actualizado el
  badge del README.

## 2. Qué cachea el pipeline

Como el build es un `docker build`, lo que se cachea son **las capas de la imagen**, usando
`cache-from` / `cache-to` con `type=gha` (el almacenamiento de cache de GitHub Actions). Cada
job usa un `scope` distinto para que no se pisen entre sí.

**Qué se reutiliza:** las capas que no cambiaron. En la práctica, la descarga de dependencias
(`go mod download` en el backend, `npm ci` en el frontend), que es la parte cara. Como los
Dockerfiles copian primero los archivos de dependencias y después el código, un cambio en el
código fuente no invalida esas capas. En el log de la corrida se ve `CACHED` en los pasos
reutilizados.

**Qué no se reutiliza:** las capas posteriores al primer cambio. Docker invalida en cascada:
si cambia una capa, todas las que vienen después se rehacen.

**Qué pasa si el cache desaparece.** El pipeline funciona igual, **solo más lento**. Y eso es
una propiedad deliberada, no un accidente: el cache es una optimización y puede desaparecer en
cualquier momento (expira, se limpia, cambia la clave). Un pipeline que **necesita** el cache
para funcionar está roto y todavía no lo sabe.

## 3. Por qué el pipeline construye con el Dockerfile

Porque así el pipeline verifica **exactamente el mismo artefacto que después se despliega**.

Si el CI compilara por su cuenta —con sus propios comandos de build, su propia versión de Go o
de Node— habría **dos definiciones distintas del build**: la del pipeline y la del Dockerfile.
Esas dos definiciones divergen con el tiempo sin que nadie se entere, y el resultado es el peor
escenario posible: un pipeline en verde y una imagen rota en producción. El pipeline estaría
certificando algo que no es lo que se despliega.

Usar el mismo Dockerfile elimina esa clase entera de problemas. Es el mismo principio que la
contenerización en general: una sola definición de cómo se construye la aplicación.

## 4. Qué produce el pipeline y dónde queda

Las imágenes que construye **nacen y mueren en el runner**: cada job arranca en una máquina
limpia y todo lo que queda adentro se pierde al terminar. Por eso los jobs usan `push: false`.

El pipeline no guarda nada a propósito, por dos razones: el lugar de una imagen es un
**registry**, no el almacén de artefactos de CI (publicar en un registry ya se hizo a mano en
el TP2); y conservar lo que produce una corrida recién tiene sentido cuando el pipeline
verifica algo más que la compilación — desde el TP5, con el reporte de tests.

**La salida real de este pipeline es otra, y no es menor: el check en verde que habilita el
merge.**

La distinción que va con esto: **artefacto** es lo que el build produce y se quiere conservar;
**cache** es lo que el build ya hizo una vez y no hace falta rehacer.

## 5. El pipeline como gate

Se activaron **required status checks** sobre `main` con los dos jobs, más
**Require branches to be up to date** (`strict: true`).

**Las tres condiciones que `main` exige hoy para aceptar un merge:**

1. Que el cambio venga por Pull Request (viene del TP1).
2. Que `build-backend` y `build-frontend` estén en verde.
3. Que la rama esté actualizada con `main`.

**Qué significa `strict: true`.** Obliga a que la rama incorpore los últimos cambios de `main`
antes de poder mergear. Sin eso, un check verde puede ser **viejo**: pasó contra un estado de
`main` que ya no existe. Dos cambios que funcionan por separado pueden romper al juntarse, y
`strict` fuerza a verificar la mezcla y no solo la rama aislada.

**Cero aprobaciones obligatorias**, igual que en el TP1: GitHub no permite aprobar el propio
PR y el trabajo es individual. **Lo que bloquea el merge en este TP no es una aprobación
humana: es el pipeline en verde.**

La configuración se hizo por la interfaz web y no con el comando `gh api --method PUT` que
ofrece la guía, deliberadamente: ese PUT **reescribe la protección entera** y todo campo
omitido vuelve a su valor por defecto, con lo que se perdería la configuración del TP1. La web
modifica solo lo que se toca.

## 6. Demostración del gate

Secuencia ejecutada y registrada en el historial de Pull Requests:

1. Rama `feature/demo-gate` con una llave de cierre eliminada de `backend/main.go`.
2. Pull Request abierto → `build-backend` **en rojo** → **botón de merge bloqueado**.
3. Commit de corrección restaurando la llave, sobre el mismo PR.
4. El pipeline vuelve a correr solo → **verde**.
5. Merge habilitado y ejecutado.

Se rompió la **compilación** y no un test porque el testing es materia del TP5. El mecanismo
del gate es idéntico: desde el TP5 el pipeline va a fallar también por un test en rojo, y solo
cambia qué lo hace fallar.

## 7. Integración continua: la práctica, no la herramienta

**¿Puede haber CI sin pipeline?** Sí. CI es la práctica de integrar cambios chicos y frecuentes
a una rama compartida verificándolos. Un equipo que integra varias veces al día y corre las
verificaciones a mano está haciendo CI.

**¿Puede haber pipeline sin CI?** También, y es el caso más común. Un pipeline que corre sobre
ramas que viven tres semanas y se integran una vez por mes automatiza la herramienta sin
cambiar la práctica. El pipeline no hace CI por sí solo: la hace posible.

## 8. Visibilidad

Badge del workflow en el README, enlazado al historial de corridas. Se escribió como enlace
(`[![CI](...badge.svg)](...ci.yml)`) y no solo como imagen: si se pone únicamente la imagen, el
badge se ve igual pero al hacerle clic se abre el SVG suelto, una página en blanco.

## 9. Versionado

Tag y release `v4.0.0`. La numeración sigue el número del práctico, según fija la consigna.

Vale explicitar la tensión: **por semver estricto esto sería `v1.1.0`**, porque agrega
funcionalidad compatible sin romper nada. Pero en este repositorio el tag numera **la entrega
académica**, no el software: sirve para que cada práctico quede congelado y navegable, y para
eso conviene que el número coincida con el TP. Es un caso donde la trazabilidad con el proceso
de evaluación pesa más que la pureza del estándar.

Se agregaron también los tags `v2.0.0` y `v3.0.0` sobre los commits que cerraron esos
prácticos, para completar la serie. Las consignas del TP2 y el TP3 no los pedían; se crearon al
llegar al TP4 y por eso su fecha de creación no coincide con la del commit que marcan.

## 10. Problemas encontrados

**Los checks no aparecían en el buscador de la protección de rama.** Al ir a configurar los
required status checks, el buscador solo ofrece checks que hayan corrido en los últimos siete
días. Antes de la primera corrida del workflow, `build-backend` y `build-frontend` no existen
para GitHub y no se pueden seleccionar.

*Cómo se resolvió:* respetando el orden — primero mergear el workflow y dejar que corra al
menos una vez, después configurar el gate. Un check obligatorio que nunca corre bloquea el Pull
Request para siempre, sin mensaje de error útil.

## 11. Declaración de uso de Inteligencia Artificial

Usé un asistente de IA (Claude) en los dos prácticos.

**En qué me asistió.** Explicación de los conceptos (jerarquía épica/historia/tarea, criterios
de aceptación verificables, sentido del límite de trabajo en progreso, diferencia entre
artefacto y cache, qué significa `strict`, por qué CI es una práctica y no una herramienta);
redacción del workflow `ci.yml` adaptado a mis dos Dockerfiles reales; redacción de los issues
y de este documento.

**Cómo lo verifiqué.**

- **Contra la ejecución real.** El workflow no se dio por bueno por parecer correcto: se
  verificó en la pestaña Actions que las dos corridas terminaran en verde, y se abrió el log
  del build para confirmar la presencia de `CACHED` en las capas reutilizadas.
- **El gate se probó rompiéndolo a propósito.** Un gate que nunca bloqueó nada no se sabe si
  funciona. Se rompió la compilación, se comprobó que el PR quedaba efectivamente bloqueado, y
  recién después se corrigió. Es el mismo criterio con el que se probó la protección de rama en
  el TP1.
- **Contra la consigna.** Se contrastaron las indicaciones con el enunciado antes de aplicarlas.
  En un caso concreto se decidió **apartarse** de la guía: configurar el gate por la web en
  lugar de usar el comando `gh api --method PUT`, porque ese comando reescribe la protección
  completa y habría borrado la configuración del TP1.
- **Cuando la IA omitió algo, la ejecución lo mostró.** En el primer PR faltó el `Closes #N` y
  el issue no se cerró solo; el problema se detectó al verificar el tablero y se resolvió a
  mano.

**Alcance.** Todas las acciones sobre el repositorio, el Project y la configuración las ejecuté
yo, verificando cada resultado. Comprendo cada decisión —por qué ese sprint, por qué ese
límite, por qué esos jobs en paralelo, qué exige `main` hoy para aceptar un merge— y puedo
explicarlas y demostrarlas navegando el repositorio y el tablero.
