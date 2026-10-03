# Revisión de seguridad, calidad y concurrencia — 2026-10-03

Base revisada: `601f7af032d9f126ca4f140005c736210c57c696` (`main`, versión 3.3.1).
Esta revisión combina lectura de código, reproducciones locales, tests de regresión y
análisis de dependencias. Los IDs R01–R16 identifican observaciones de la revisión;
no son IDs de avisos de GitHub ni equivalen todos a vulnerabilidades explotables.

## Comparación con GitHub

La consulta de issues y PRs públicos no encontró issues abiertos el 2026-10-03.
Los siguientes trabajos ya estaban integrados en la base revisada:

| Trabajo | Estado y relación con esta revisión |
| --- | --- |
| [Issue #9](https://github.com/rendis/statepro/issues/9), [PR #12](https://github.com/rendis/statepro/pull/12) | Actualizaciones de dependencias anteriores, ya cerradas/integradas. |
| [PR #18](https://github.com/rendis/statepro/pull/18) | Consolidó #14–#17: snapshots, tracking y correcciones de runtime. La restauración concurrente de metadata y el rechazo atómico de snapshots requieren las correcciones adicionales de este PR. |
| [PR #19](https://github.com/rendis/statepro/pull/19) | Actualizó dependencias y documentó que el test de autolayout se bloqueaba antes de esa actualización. Este PR corrige el ciclo de renders. |
| [PR #20](https://github.com/rendis/statepro/pull/20) | Corrigió ReDoS en identificadores y normalización de JSON Pointer, y un aviso de esbuild. La escritura segura de propiedades JSON abordada aquí es un problema distinto. |
| [PR #21](https://github.com/rendis/statepro/pull/21) | Publicó 3.3.1 con #19 y #20; no incorporó las correcciones nuevas de esta revisión. |

No se consultó el estado actual de los paneles privados de Code Scanning,
Dependabot o Security Advisories. Un issue cerrado o un PR integrado no demuestra
que esos paneles carezcan de alertas. No se creó otro issue público de
vulnerabilidad: el repositorio indica usar Security Advisories privados para esas
notificaciones.

## Disposición de las observaciones

| ID | Observación | Resultado de este PR |
| --- | --- | --- |
| R01 | Seguridad de propiedades y pointers en metadata | Corregida. Lectura sólo de propiedades propias; escritura y merge como propiedades de datos, conservando las claves JSON. Tests directos y de fusión de constantes de packs. |
| R02 | Fallback de observers desconocidos | Contrato existente documentado y probado en #14/#18: devuelve `true`. No se cambia el valor por defecto. Un modo estricto requiere una propuesta de compatibilidad y conocer el límite de confianza de las definiciones. |
| R03 | Pérdida de condiciones con igual `src` | Corregida en importación, exportación y validación. Se conservan argumentos, orden y multiplicidad, tal como ejecuta Go. Incluye exportación editada y snapshot importado sin editar. |
| R04 | Registro global de executors concurrente | Corregida con `RWMutex`. Test concurrente de los cuatro tipos de executor, ejecutado con `-race`. El mutex no se mantiene mientras se ejecuta una función registrada. |
| R05 | Restauración de metadata concurrente con invokes | Corregida usando el mutex compartido de metadata y conservando la identidad del mapa usado por los args existentes. Test concurrente y escritura posterior a la restauración. |
| R06 | Snapshot inválido modifica estado antes del error | Corregida mediante preparación/validación de todos los universos incluidos antes de aplicar. Se verifican flags esenciales, referencias de realities y entradas del acumulador. Los acumuladores vacíos omitidos se reconstruyen. |
| R07 | Ciclo de renders ante mediciones sin cambios | Corregida con bailouts por identidad en reducers y callback estable. Pasan los cinco tests de autolayout, incluido StrictMode, que antes no terminaban. |
| R08 | Reentrada de callbacks en métodos públicos del owner | Pendiente de diseño. Los callbacks síncronos ejecutan bajo el mutex de máquina; deben usar `args.GetSnapshot()` y `args.EmitEvent()` en lugar de llamar sincrónicamente al owner capturado. Esos args no deben retenerse para acceso asíncrono a estado no sincronizado. |
| R09 | Cancelación ignorada al enviar eventos | Corregida la admisión en `SendEvent`: se comprueba el contexto antes y después de adquirir el mutex. No añade rollback ni interrupción de callbacks ya en ejecución; los bucles del debugger bot y las demás operaciones requieren evaluación separada. |
| R10 | Caché de validadores sólo por ID editable | Corregida: clave por contenido del schema, compilador aislado por schema y LRU limitado a 128 entradas. Tests de edición y reutilización del mismo ID entre instancias, incluso con `$id` repetido. |
| R11 | Restauración retiene metadata/tracking/flags posteriores | Corregida para universos incluidos: metadata reemplazada bajo mutex, tracking ausente vaciado y flag final recalculado desde la reality concreta o anterior a superposición. |
| R12 | Conversión de snapshots oculta errores JSON y pierde precisión | Pendiente. Hace falta un API de snapshot con error y decidir una representación numérica compatible. La ruta actual mantiene sus firmas y conversión JSON. |
| R13 | Nil y eventos personalizados provocan panic | Corregidos `NewQuantumMachine` con modelo/universo nil, `NewExQuantumMachine` con modelo nil, `SendEvent` con nil tipado o no tipado y la acumulación de implementaciones personalizadas de `Event`. No se afirma tolerancia a cualquier modelo malformado: validar definiciones sigue siendo responsabilidad del caller; `NewExUniverse(nil)` no forma parte de esta corrección. |
| R14 | Invokes, acumuladores e historiales sin presupuesto | Pendiente de diseño. Los invokes que continúan tras abandonar una reality son un contrato explícito de #14/#18. Límites y cancelación al salir deben ser configurables; no se modifica su ciclo de vida en este PR. |
| R15 | Coste del historial, serialización y carga de ELK | Pendiente de optimización. El historial sigue comparando/guardando el grafo completo; no se declara resuelto con el bailout de mediciones. La carga del bundle conserva ELK en la entrada. |
| R16 | Typecheck roto y lint sin verificación efectiva | Corregido el typecheck del Web Component incluyendo la declaración SVG. La configuración de lint sigue siendo un placeholder y no se incorpora un workflow de CI en este PR. |

## Semántica y compatibilidad

`LoadSnapshot(nil, ...)` conserva su comportamiento de no-op. Los snapshots
parciales siguen permitidos: sólo se restauran universos incluidos y conocidos.
Para esos universos, metadata y tracking representan el estado restaurado, no una
fusión con el estado posterior. Un rechazo deja estado, tracking y contexto de
máquina sin modificar por la restauración. La validación no comprueba identidad
criptográfica ni versión de la definición.

Los invokes ya iniciados pueden escribir metadata después de restaurar; esta
corrección sincroniza el acceso, no cancela tareas ni hace transaccionales sus
efectos externos. Los valores de metadata deben ser compatibles con JSON para
la API actual de snapshots. No se deben compartir valores mutables anidados entre
goroutines sin sincronización propia.

El constructor que antes hacía panic con modelo nil ahora devuelve error. Los
callers deben manejar el error que ya forma parte de su firma. En Studio, repetir
un executor de condición deja de bloquear la exportación; repetirlo es válido para
el runtime, incluso cuando los argumentos coinciden.

## Validación y cobertura

Las regresiones de snapshots y Studio se ejecutaron contra la base sin corregir y
fallaron antes de aplicar los cambios. La detección de carreras original reprodujo
los accesos de registros y `LoadSnapshot`/invokes. Los tests permanentes verifican
el comportamiento corregido, no la presencia del fallo.

Comandos reproducibles desde la raíz, con un binario Go real en `PATH`:

```sh
go build ./...
go vet ./...
go test -race -cover ./...
GOMAXPROCS=2 make test-fuzz-smoke
```

En `studio/`, usando la versión de pnpm fijada por el repositorio:

```sh
npx --yes pnpm@10.11.0 install --frozen-lockfile
npx --yes pnpm@10.11.0 test
npx --yes pnpm@10.11.0 typecheck
npx --yes pnpm@10.11.0 build
npx --yes pnpm@10.11.0 audit --prod
npx --yes pnpm@10.11.0 audit
```

Resultados de la validación local del PR: build y vet de Go correctos;
`go test -race -cover ./...` correcto en los 11 paquetes (builtin y bot 100%,
runtime experimental ~92%, CLI 1,3%). Las tres sesiones de fuzz smoke de cinco
segundos terminaron correctamente. Studio completó 251 tests en 44 archivos,
typecheck y build de los tres paquetes. App y Web Component no tienen tests
propios: su script permite terminar sin archivos de test; no se contabilizan
como cobertura de integración.

`pnpm audit --prod` devuelve cero avisos. El audit completo bajó de 15 avisos
en la base revisada a tres: dos moderados (`vitest` y `@vitest/mocker`,
[GHSA-82fw-gwwq-j7x9](https://github.com/advisories/GHSA-82fw-gwwq-j7x9),
requieren >=4.1.11) y uno alto (`braces`,
[GHSA-vfj7-8cjw-p6xm](https://github.com/advisories/GHSA-vfj7-8cjw-p6xm),
sin versión corregida según la respuesta del registro). Todos aparecen en
dependencias de desarrollo. Los avisos del registro no prueban explotabilidad
en la aplicación desplegada. Migrar Vitest exige revisar sus dependencias y la
integración de Stryker; para braces hay que sustituir/actualizar su dependencia
padre cuando exista una solución, en lugar de forzar una versión inventada.

La entrada JavaScript de la app sigue siendo de aproximadamente 2,28 MB
minificada / 677 kB gzip; el build conserva la advertencia de tamaño. El audit
de dependencias y sus conteos dependen de la fecha y del registro consultado.

El análisis inicial cubrió 296 archivos versionados (~30.722 líneas de Go,
TS y TSX), con lectura dirigida a las rutas de mayor riesgo, no una lectura
exhaustiva de cada línea de interfaz. No se ejecutaron pruebas en navegador real,
producción ni mutation testing. El plugin Codex Security está instalado, pero
esta sesión no expone herramientas de escaneo; no se atribuyen estos resultados
a un scan oficial de ese plugin.

En la revisión base, `govulncheck` con Go 1.26.8 no encontró símbolos afectados;
con Go 1.25.0 identificó tres avisos de `net/url` por la ruta de inicialización de
jsonschema. No se demostró una entrada maliciosa en esa ruta. El resultado depende
del toolchain, no sólo de `go.mod`: usar una versión con parches de seguridad,
como Go 1.25.13 o posterior de esa rama. Dos avisos de módulos `x/text` y `x/sys`
no resultaron alcanzables en el análisis realizado. Estos resultados de la base
no son una garantía para otros sistemas operativos, tags de build o toolchains.
