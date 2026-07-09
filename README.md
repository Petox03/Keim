# Keim (🌱)

> Keim (del alemán: germen, brote o semilla) es una CLI que automatiza el scaffolding de un entorno de desarrollo Go reproducible, aislado en Docker y sin requerir Go instalado en el host.

---

## Qué es

Keim elimina la fricción inicial al comenzar proyectos en Go dentro de entornos contenerizados. Genera un entorno de desarrollo reproducible, eficiente y moderno basado en Docker, listo para programar con un solo comando.

## Filosofía

- **Cero Configuración Manual:** elimina el boilerplate repetitivo al iniciar un proyecto Go en Docker.
- **Aislamiento Total:** el proyecto se ejecuta, compila y descarga dependencias exclusivamente dentro de Docker. No necesitas Go en tu máquina.
- **Persistencia de Caché:** las cachés de Go (`GOCACHE`, `GOMODCACHE`) se persisten en volúmenes nombrados para evitar descargas y recompilaciones redundantes.
- **Principio de No Maleficencia:** Keim jamás destruye, sobreescribe o altera archivos preexistentes en el directorio de trabajo.

## Usuario objetivo

Un desarrollador que está aprendiendo Go y quiere experimentar sin instalar Go en su máquina. Trabaja cómodo con Docker porque es el estándar que usa con su entorno social/corporativo.

## Qué genera

Keim inyecta 6 archivos en el directorio objetivo:

```
[Directorio del Proyecto]
 ├── .dockerignore
 ├── .gitignore
 ├── compose.yml
 ├── Dockerfile
 ├── go.mod
 └── main.go
```

- `go.mod` — inicializa el módulo de Go con el nombre de la carpeta y la versión detectada.
- `main.go` — punto de entrada básico en la raíz del módulo.
- `Dockerfile` — imagen `golang:alpine` con cachés redirigidas a volúmenes persistibles.
- `compose.yml` — servicio `app` dormido (`sleep infinity`), bind mount para código, volúmenes nombrados para caché.
- `.gitignore` — excluye binarios temporales y caché local.
- `.dockerignore` — excluye archivos irrelevantes del contexto de build de Docker.

## Cómo se usa

```
keim init                                # Siembra en el directorio actual, cascada por defecto
keim init mi-proyecto                    # Crea carpeta y siembra allí
keim init --detect host,internet         # Cascada personalizada
keim init --detect manual=1.26           # Versión fija explícita
keim init --detect manual                # Manual sin versión (Keim pregunta por stdin)
```

Después de generar, el flujo de desarrollo es:

```
cd mi-proyecto
docker compose up -d
docker compose exec app go run .
```

## Documentación

- [`docs/decisions.md`](.devin/docs/decisions.md) — Decisiones de diseño con contexto, alternativas y trade-offs (ADRs).
- [`docs/architecture.md`](.devin/docs/architecture.md) — Arquitectura técnica: paquetes, responsabilidades, flujo, dependencias.
- [`docs/roadmap.md`](.devin/docs/roadmap.md) — Deudas técnicas, planes futuros y lo que NO está en el MVP.

## Estado

MVP en desarrollo. Iteración 1 (walking skeleton) en progreso: `project`, `validator`, `templates` y `generator` implementados. La documentación en `.devin/docs/` es la base de lo que se va a implementar.
