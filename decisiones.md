# Decisiones — TP1

Repositorio: https://github.com/SantiClark02/ingsoft3-tp01

---

## 1. El conflicto de merge

### Qué modificaciones lo produjeron

Creé dos ramas, `feature/titulo-a` y `feature/titulo-b`, **ambas partiendo del mismo
commit de `main`**. Las dos modificaron exactamente la misma línea del `README.md` —
la primera, el título del proyecto — con contenido distinto:

- `feature/titulo-a` → `# Proyecto IngSoft3 - versión A`
- `feature/titulo-b` → `# Proyecto IngSoft3 - versión B`

Abrí los dos Pull Requests antes de mergear ninguno, y después mergeé primero el de A.
Entró limpio: nadie más había tocado esa línea todavía. Al intentar mergear el de B,
GitHub avisó que no podía hacerlo automáticamente.

### Por qué Git no pudo resolverlo solo

Git integra automáticamente sin problema cuando dos ramas tocan **partes distintas** de
un archivo: ahí no hay ambigüedad, cada cambio va a su lugar.

Acá no fue el caso. Git compara las dos puntas contra el **ancestro común** —el commit
de `main` del que salieron las dos ramas— y encuentra dos modificaciones incompatibles
sobre exactamente la misma línea. No tiene ningún criterio técnico para decidir cuál de
las dos es "la correcta": eso es una decisión de contenido, semántica, que depende de la
intención de quien escribió el cambio.

Frente a esa ambigüedad Git no adivina. Marca el archivo con `<<<<<<<`, `=======` y
`>>>>>>>` y delega la decisión a una persona. Resolver un conflicto es decidir qué queda,
no ejecutar un comando.

Un conflicto no es un error: es la consecuencia natural del trabajo en paralelo. Lo que
sí es evitable es el conflicto *grande* — ramas cortas e integración frecuente producen
conflictos chicos y triviales; ramas que viven semanas producen el problema opuesto.

### Qué tendría que haber pasado para que no apareciera

Cualquiera de estas tres condiciones habría evitado el conflicto:

1. **Que las dos ramas tocaran líneas distintas del archivo.** Ahí Git fusiona solo,
   porque no hay dos versiones compitiendo por el mismo lugar.
2. **Que `feature/titulo-b` hubiera nacido después de mergear A**, partiendo del `main`
   ya actualizado. En ese caso B habría visto el cambio de A y no habría divergencia:
   habría sido un avance sobre él, no una alternativa.
3. **Que B hubiera integrado `main` antes de pedir el merge**, resolviendo la diferencia
   en su propia rama en vez de al momento de integrar.

El patrón general: el conflicto nace de trabajar en paralelo sobre lo mismo sin integrar.

### Estrategia de merge elegida

Para todos los PRs de este práctico usé **Squash and merge**, siguiendo la regla de la
materia. Todos los commits de la rama se aplastan en uno solo sobre `main`. El historial
principal queda lineal y legible —un commit por Pull Request, fácil de leer y de
revertir—, a costa de perder el paso a paso interno de cómo se llegó a ese cambio.

Después de cada merge eliminé la rama: las ramas de feature son descartables, ya cumplieron
su función cuando su contenido entró a `main`.

---

## 2. Problemas encontrados

### 2.1 El `.gitignore` se commiteó vacío

Creé el archivo con `New-Item .gitignore` y lo abrí con el Bloc de notas, pero hice
`git add` antes de guardar el contenido. El commit entró con el archivo en cero bytes.

**Cómo lo detecté:** leyendo la salida del propio commit, que decía
`1 file changed, 0 insertions(+), 0 deletions(-)`. Cero inserciones sobre un archivo que
supuestamente tenía diez líneas.

**Cómo lo resolví:** guardé el contenido y usé `git commit --amend --no-edit`, que
reemplaza el último commit por uno nuevo con el contenido corregido. Fue seguro hacerlo
porque ese commit todavía no había salido de mi máquina —el push había fallado—, y la
regla es no reescribir historia que ya se compartió. Se nota que el commit fue reemplazado
y no editado porque cambió el hash: pasó de `ac110fe` a `1481b22`.

**Qué aprendí:** conviene leer la salida de los comandos en vez de asumir que funcionaron.
El error estaba escrito en pantalla desde el principio.

### 2.2 El primer `git push` falló con error de autenticación

El primer push devolvió `Invalid username or token. Password authentication is not
supported for Git operations`.

**Diagnóstico:** no era un problema de permisos del repositorio ni de la configuración de
la rama, aunque el mensaje puede confundir. Git Credential Manager estaba activo
(`git config credential.helper` devolvía `manager` a nivel *system*), pero tenía guardada
una credencial vieja e inválida para `github.com`. Como encontraba algo guardado, la
enviaba sin pedirme nada nuevo, y GitHub la rechazaba.

**Cómo lo resolví:** limpié la credencial almacenada desde el Administrador de
credenciales de Windows, para forzar a Git Credential Manager a reautenticar. En el
siguiente `push` se abrió el navegador, inicié sesión en GitHub y quedó resuelto de
forma permanente.

**Nota:** al configurar Git al principio consulté `git config --global credential.helper`
y salió vacío, lo que me hizo pensar que no tenía Credential Manager. En Windows el
instalador de Git lo configura a nivel *system*, no *global* — hay que consultar
`git config credential.helper` sin flag, que combina todos los niveles.

### 2.3 La rama del primer PR quedó sin el prefijo de la convención

Al crear el primer Pull Request desde la web, acepté el nombre de rama que GitHub propone
por defecto y quedó `seccion-instalacion`, sin el prefijo `feature/` que usa la convención
de la materia.

**Cómo lo resolví:** no lo corregí retroactivamente, porque la rama ya había cumplido su
función y renombrar historia habría sido un remedio peor que el problema. Escribí el
nombre completo a mano en las ramas siguientes: `feature/titulo-a` y `feature/titulo-b`
sí respetan la convención.

### 2.4 Caracteres acentuados mal mostrados en PowerShell

Al verificar el README con `Get-Content README.md -Head 1`, PowerShell mostraba
`versiÃ³n` en lugar de `versión`.

**Diagnóstico:** era un problema de visualización, no del archivo. El README se editó
desde la web de GitHub, que guarda en UTF-8, y PowerShell 5.1 lo lee asumiendo otra
codificación. Verificado abriendo el archivo en GitHub, donde se ve correctamente.
No requirió corrección.

---

## 3. Declaración de uso de Inteligencia Artificial

Usé un asistente de IA (Claude) durante todo el desarrollo de este trabajo práctico.

### En qué me asistió

- **Explicaciones conceptuales:** qué es realmente una rama para Git, por qué el staging
  area existe, la diferencia entre repositorio local y remoto, qué hace cada estrategia de
  merge, qué significan los marcadores de conflicto, y la diferencia entre un tag y una
  release.
- **Planificación:** organizar el trabajo en etapas verificables y en el orden correcto —
  por ejemplo, entender que el `.gitignore` tenía que entrar *antes* de proteger `main`,
  porque después ya no habría podido pushearlo directo.
- **Comandos y sintaxis:** los comandos de Git usados y qué hace cada flag.
- **Análisis de errores:** el diagnóstico de los problemas de la sección 2, especialmente
  el del `.gitignore` vacío y el de las credenciales.
- **Redacción:** la estructura y el texto de este archivo y de `evidencias.md`.

### Cómo verifiqué lo que me devolvió

- **Contra la guía de la cátedra:** contrasté las indicaciones del asistente con el
  enunciado del TP antes de aplicarlas, especialmente en las decisiones de configuración
  (protección de rama, cero aprobaciones, `Do not allow bypassing`).
- **Contra la salida real de mi terminal y de GitHub:** ninguna etapa se dio por buena sin
  ver el resultado. Trabajé de forma incremental, verificando cada checkpoint antes de
  avanzar, y pegando la salida real de cada comando. Cuando el asistente asumió algo que
  no era cierto —consultar `credential.helper` en el nivel equivocado—, la salida real lo
  desmintió y se corrigió el diagnóstico.
- **Probando las configuraciones en vez de confiar en ellas:** la protección de `main` no
  se dio por funcionando porque estuviera configurada, sino porque se probó con un push
  directo real que fue rechazado. Una protección que nunca rechazó nada no se sabe si
  funciona.

### Alcance

La IA me ayudó a entender y a documentar, pero las acciones sobre el repositorio las
ejecuté yo, verificando cada resultado. Comprendo cada decisión tomada, cada comando
utilizado y el porqué de cada configuración, y puedo explicarlos y demostrarlos navegando
el repositorio.
