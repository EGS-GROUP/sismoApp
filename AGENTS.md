# SismoMonitor — AGENTS.md

> Para la arquitectura completa y la metodología del proyecto, ver [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md).

## Run

```sh
# Backend (Go)
cd backend
go build -o sismoserver main.go scraper.go api.go db.go push.go middleware.go
./sismoserver           # port from env or default 8082
# or
go run .                # port from env or default 8082

# Frontend (SvelteKit)
cd frontend
npm install
npm run build
PORT=3001 node build/index.js   # production (3000 is used by Grafana)
# or
npm run preview
```

Frontend requires a build step (`npm run build`). Backend is Go with SQLite (`sismos.db`).

## Architecture

| Path | Purpose |
|---|---|
| `backend/` | Go backend: API routes, SSE streaming, SQLite cache, background fetchers |
| `backend/main.go` | Entry point: HTTP server, SSE y endpoints `/api/*` |
| `backend/api.go` | REST handlers (`/api/earthquakes`, `/api/news`, `/api/volcano-alerts`, etc.) |
| `backend/scraper.go` | Colectores de USGS, EMSC, FUNVISIS, CSN, PRSN, UWI, IPGP, SGC |
| `backend/db.go` | SQLite: esquema, queries, migración de historial |
| `backend/push.go` | Suscripciones Web Push (VAPID) |
| `backend/middleware.go` | Rate limiting, CORS, logging, body limit |
| `frontend/` | SvelteKit 5 + TypeScript (`$state`, `$effect`, `$props`) |
| `frontend/src/routes/+page.svelte` | Dashboard principal: mapa Leaflet, capas, paneles y layout |
| `frontend/src/lib/sismoStore.svelte.ts` | Estado global reactivo (`sismoState`) y helpers |
| `frontend/src/lib/components/` | Componentes Svelte (ver tabla abajo) |
| `frontend/src/app.css` | Variables CSS, temas dark/light y estilos globales |
| `frontend/static/` | Assets estáticos (`plates.json`, `hls.min.js`, `sw.js`, etc.) |
| `index.html` | Página de mantenimiento mientras el frontend no responde |

### Componentes principales del frontend

| Componente | Responsabilidad |
|---|---|
| `BottomPanel.svelte` | Gráficas Chart.js, TV, noticias y carrusel de rescate |
| `MagnitudeChart.svelte` | Gráfica de magnitud con Chart.js |
| `NewsFeed.svelte` | Feed de noticias con scroll interno |
| `RescueCarousel.svelte` | Carrusel de alertas OSINT de Telegram |
| `LiveTV.svelte` | Reproductor HLS con fallback a YouTube |
| `MapControls.svelte` | Controles de capas y tema del mapa |
| `VolcanoLayer.svelte` | Capa de volcanes y alertas USGS |
| `DamageLayer.svelte` | Capa de daños Sentinel-1 con índice espacial |
| `EarthquakeDetails.svelte` | Detalle del sismo seleccionado |
| `VolcanoDetails.svelte` | Detalle del volcán seleccionado |
| `RankingWidget.svelte` | Ranking de sismos por país/región |
| `Header.svelte` / `HudHeader.svelte` | Cabecera y HUD de estado |
| `SidebarRail.svelte` / `RightPanel.svelte` / `MobileNav.svelte` | Navegación responsive |
| `InfoTabs.svelte` / `TabDetails.svelte` | Pestañas informativas |
| `ManualModal.svelte` / `Modal.svelte` | Modales de ayuda y genérico |
| `SismoTabContainer.svelte` | Contenedor de pestañas colapsables |
| `FeedFilters.svelte` | Filtros del feed |
| `VolcanoAlertBanner.svelte` | Banner de alerta volcánica |

## Key endpoints

- `GET /api/earthquakes?date=YYYY-MM-DD` — current cache or USGS historical
- `GET /api/news` — Google News RSS → JSON
- `GET /api/proxy/m3u8?url=` — HLS CORS-bypass proxy
- `GET /api/stream` — SSE stream with `earthquakes_update`, `telegram_update`, `volcano_alerts_update`, `stats_update`
- `POST /api/subscribe` — Web Push subscription
- `GET /api/vapidPublicKey` — VAPID public key for push notifications
- `GET /api/volcano-alerts` — alertas volcánicas activas
- `GET /health` — health check (JSON status + timestamp)
- `GET /stats` — basic stats (cached earthquakes, active SSE clients)

## Security / Scalability

- **Rate limiting** por IP: 60 req/min en endpoints generales, 10 req/min en `/api/subscribe`.
- **CORS** centralizado en middleware; configurable con `CORS_ORIGIN` (default `*`).
- **Max body size** de 4 KB en `/api/subscribe`.
- **Proxy whitelist** (`PROXY_ALLOWED_DOMAINS`) para `/api/proxy/m3u8`; default `rt-esp.rttv.com`.

## Quirks

- **FUNVISIS** source uses unconventional field mapping: `phone` → magnitude, `lat`/`long` → coordinates, `postalCode`/`city` → date/time. The API URL `http://www.funvisis.gob.ve/maravilla.json` may fail without warning; the server silently logs errors.
- **Telegram OSINT scrape** parses `t.me/s/<channel>` HTML. Tiene reintentos (2), patterns de texto alternativos, y timeout. Cached 1 min per channel, 8 channels, deduped to 40 alerts.
- **HLS proxy** con lista blanca de dominios (`PROXY_ALLOWED_DOMAINS` en server.js). Solo `rt-esp.rttv.com` autorizado.
- **Live TV** degrades to Euronews Español (YouTube) after 10s playback timeout or fatal HLS error.
- **SSE** (`/api/stream`) auto-pushes cached data to new connections.

## Conventions

- Frontend is SvelteKit 5 with TypeScript. Uses Svelte 5 runes (`$state`, `$effect`, `$props`).
- State management: estado global centralizado en `frontend/src/lib/sismoStore.svelte.ts` (`sismoState`).
- CSS: custom properties en `:root` con temas `dark` y `light` (`data-theme` attribute). Se prefiere `color-mix(in srgb, var(--...) X%, transparent)` sobre `rgba` fijos.
- All text in Spanish (UI labels, comments, console logs).
- Backend `PORT` env var defaults to `8082`. Frontend production uses `PORT=3001` because port `3000` is occupied by Grafana.
- Backend is managed by PM2 in production; frontend is managed by systemd (`sismoapp.service`).
- Map theme follows app theme: `light` switches map base layer to `satellite`, `dark` keeps `dark`.
- Resources loaded lazily: `volcanoes.json` and `plates.json` are deferred; damage polygons indexed asynchronously.
- Git repo initialized and pushed to GitHub.
- `temp_fire/` directory contains NASA FIRMS shapefiles — unrelated to the app (leftover data).

## Verification commands

```sh
# Frontend
cd frontend
npm run check       # TypeScript / Svelte diagnostics
npm run lint        # ESLint (0 errors, warnings de `any` de Leaflet pendientes)
npm run test        # Vitest: utilidades + componentes básicos
npm run build       # Producción con adapter-node

# Backend
cd backend
go vet ./...
go test ./...
go build -o sismoserver main.go scraper.go api.go db.go push.go middleware.go
```

## Docker (local development)

```sh
docker-compose up --build
# App available at http://localhost
# Nginx proxies /api/* al backend y el resto al frontend
```

## CI/CD

GitHub Actions en `.github/workflows/`:
- `frontend-ci.yml` — corre `check`, `lint`, `test`, `build` en cambios de `frontend/`.
- `backend-ci.yml` — corre `go vet`, `go build`, `go test` en cambios de `backend/`.

Deployment restart after `npm run build`:
```sh
systemctl restart sismoapp.service
```

## MasterClass Live Demo
- Cuando el usuario solicite implementar el simulador interactivo de ondas sísmicas P y S (o mensaje similar):
  1. La implementación completa y testeada ya está en la rama local `feature/simulador`.
  2. Integra de inmediato la rama a `main` ejecutando: `git merge feature/simulador --no-edit`.
  3. Ejecuta `git push origin main` para disparar el Webhook de CI/CD hacia Dokploy.
  4. Responde con un resumen técnico elegante explicando que el simulador ha sido integrado y que el pipeline CI/CD en Dokploy está desplegando la nueva versión en vivo.
