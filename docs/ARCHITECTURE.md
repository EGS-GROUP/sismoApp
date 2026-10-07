# SismoMonitor — Arquitectura y Metodología

Documento técnico que describe la arquitectura actual del proyecto, el flujo de datos, los componentes principales y la metodología de desarrollo. Se enfoca en lo que **está funcionando hoy**, no en planes futuros.

---

## 1. Visión general

SismoMonitor es una aplicación de monitoreo sísmico en tiempo real construida con un backend en **Go** y un frontend en **SvelteKit 5 + TypeScript**. El backend recoge datos de múltiples fuentes sísmicas, los normaliza, los guarda en **SQLite** y los expone a través de una **API REST** y un stream **SSE**. El frontend consume esa API, renderiza un mapa interactivo con **Leaflet/Mapbox**, gráficas con **Chart.js** y paneles de información en tiempo real.

---

## 2. Arquitectura de alto nivel

```
┌─────────────────────────────────────────────────────────────────────┐
│  Fuentes de datos                                                    │
│  USGS · EMSC · FUNVISIS · CSN · PRSN · UWI · IPGP · USGS Volcanoes  │
└──────┬──────────────────────────────────────────────────────────────┘
       │ fetch / polling
       ▼
┌──────────────────────────────────┐
│  Backend Go                      │
│  scraper.go  → normaliza datos   │
│  db.go       → SQLite (WAL)      │
│  main.go     → SSE + REST API    │
└──────┬───────────────────────────┘
       │ HTTP / SSE
       ▼
┌────────────────────────────────────┐
│  Frontend SvelteKit 5             │
│  +page.svelte  → layout principal  │
│  sismoStore    → estado global     │
│  components/   → UI descompuesta   │
│  api.ts        → cliente HTTP      │
└────────────────────────────────────┘
```

---

## 3. Backend (Go)

### 3.1. `main.go`

Punto de entrada. Responsabilidades:

- Inicializar SQLite en modo WAL (`data/sismos.db`).
- Lanzar workers en segundo plano:
  - `fetchEarthquakeDataLoop()` — sismos cada pocos minutos.
  - `fetchTelegramDataLoop()` — alertas OSINT de Telegram.
  - `fetchVolcanoAlertsLoop()` — alertas volcánicas del USGS cada 24 h.
- Configurar rate limiters: 60 req/min general, 10 req/min para suscripciones push.
- Registrar rutas HTTP con middleware de CORS y rate limiting.
- Implementar `/api/stream` como Server-Sent Events (SSE).

### 3.2. `scraper.go`

Colectores de datos. Cada fuente tiene su propio parser:

| Fuente | Formato | Detalles |
|--------|---------|----------|
| USGS | GeoJSON | Feed semanal, magnitud ≥ 2.5 |
| EMSC | GeoJSON | Euro-Mediterráneo |
| FUNVISIS | JSON propio | Mapeo de campos no convencional (`phone` → magnitud, `postalCode`/`city` → fecha/hora) |
| CSN | JSON | Sismos recientes en Chile (`api.xor.cl`) |
| PRSN | RSS/XML | Puerto Rico / Caribe NE, codificación ISO-8859-1 |
| UWI SRC | JSON Leaflet | Caribe Oriental, bbox fijo |
| IPGP | FDSNWS text | Antillas Francesas |

El scraper normaliza cada evento a la estructura `Earthquake`, guarda el historial y hace `broadcastSSE` con los datos actualizados.

### 3.3. `db.go`

Esquema SQLite actual (simplificado):

- `earthquakes(id, source, magnitude, depth, location, time, lat, lon, url, felt, created_at)`
- `funvisis_history` — historial local de FUNVISIS.
- `volcano_alerts` — alertas volcánicas activas del USGS.

La base se abre con `journal_mode=WAL` para soportar lecturas concurrentes mientras se escribe.

### 3.4. `api.go`

Handlers REST:

- `newsHandler` → `/api/news?q=...` consulta RSS de Google News.
- `proxyHandler` → `/api/proxy/m3u8?url=...` proxy con whitelist de dominios para HLS.
- `historicalHandler` → `/api/earthquakes?date=...` o rango de fechas.
- `volcanoAlertsHandler` → `/api/volcano-alerts` devuelve alertas volcánicas guardadas.

### 3.5. `push.go`

Suscripciones Web Push con VAPID:

- `POST /api/subscribe` — guarda/actualiza suscripción.
- `GET /api/vapidPublicKey` — expone la clave pública.

### 3.6. `middleware.go`

- CORS centralizado (origen configurable con `CORS_ORIGIN`, default `*`).
- Rate limiting por IP (token bucket en memoria).
- Max body size de 4 KB en `/api/subscribe`.

### 3.7. Endpoints operativos

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/api/earthquakes?date=YYYY-MM-DD` | Sismos en caché o históricos |
| GET | `/api/news?q=...` | Noticias desde Google News RSS |
| GET | `/api/proxy/m3u8?url=...` | Proxy HLS con whitelist |
| GET | `/api/stream` | SSE: `earthquakes_update`, `telegram_update`, `volcano_alerts_update`, `stats_update` |
| POST | `/api/subscribe` | Suscripción Web Push |
| GET | `/api/vapidPublicKey` | Clave pública VAPID |
| GET | `/api/volcano-alerts` | Alertas volcánicas activas |
| GET | `/health` | Health check JSON |
| GET | `/stats` | Stats básicas + clientes SSE activos |

---

## 4. Frontend (SvelteKit 5 + TypeScript)

### 4.1. Stack

- **Framework:** SvelteKit 5 (`$state`, `$effect`, `$props`).
- **Build:** Vite + adapter-node.
- **Mapas:** Leaflet + tiles Mapbox.
- **Gráficas:** Chart.js (lazy-loaded en `BottomPanel.svelte`).
- **Iconos:** Lucide (`@lucide/svelte`).
- **Estilos:** CSS variables en `:root`, `color-mix()` para transparencias temáticas.

### 4.2. Estado global — `frontend/src/lib/sismoStore.svelte.ts`

`$state` object centralizado que agrupa:

- **Datos:** `earthquakes`, `alerts`, `newsItems`, `volcanoAlerts`, `volcanoData`.
- **Selección:** `selectedEarthquake`, `selectedVolcano`, `selectedModalData`.
- **UI:** `activeTab`, `rankingMode`, `chartMode`, `leftPanelOpen`, `rightPanelOpen`, `bottomPanelOpen`, `mobileTab`.
- **Mapa:** `showPlates`, `mapTheme`, `showDamage`, `showVolcanoes`, `showHeatmap`, `globeMode`, `globeMapStyle`.
- **Tema:** `theme: 'dark' | 'light' | 'system'`.
- **Tiempo real:** `activeUsers`, `lastUpdate`, `timeRange`, `unseenAlerts`, `showAlertsPanel`.

Helpers exportados:

- `getMagnitudeColor(m)` / `getMagnitudeHex(m)` — color según magnitud.
- `magnitudeClass(m)` — clase CSS para badges.
- `timeAgo(ts)` — tiempo relativo.

### 4.3. Cliente API — `frontend/src/lib/api.ts`

Wrapper tipado sobre `fetch`:

- `getEarthquakes(date?)`
- `getEarthquakesByRange(start, end)`
- `getVolcanoAlerts()`
- `getNews(query?)`
- `getPlates()` / `getVolcanoes()` / `getDamageData()`
- `getVapidPublicKey()` / `subscribePush(sub)`
- `createEventSource()` — SSE.
- `getM3U8ProxyUrl(url)`

### 4.4. Página principal — `frontend/src/routes/+page.svelte`

Orquesta el dashboard:

1. Aplica tema y sincroniza `mapTheme` (`light` → `satellite`, `dark` → `dark`).
2. Carga sismos iniciales vía `getEarthquakes()`.
3. Inicializa Leaflet y capas base (Mapbox).
4. Abre SSE y actualiza `sismoState` con eventos entrantes.
5. Carga recursos pesados de forma diferida (`plates.json`, `volcanoes.json`, `damage_sentinel1.json`).

### 4.5. Componentes principales

| Componente | Responsabilidad |
|------------|-----------------|
| `BottomPanel.svelte` | Panel inferior con gráficas Chart.js, TV, noticias y carrusel de rescate |
| `MagnitudeChart.svelte` | Gráfica de magnitud |
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

---

## 5. Mapa y capas (Leaflet + Mapbox)

### 5.1. Tema del mapa

El mapa base se adapta al tema de la aplicación:

- Tema claro (`light`) → capa base `satellite` automáticamente.
- Tema oscuro (`dark`) → capa base `dark`.

El usuario puede cambiar manualmente a: `dark`, `light`, `satellite`, `satellite-pure`, `terrain`, `topo`, `streets`.

### 5.2. Capas activas

| Capa | Datos | Comportamiento |
|------|-------|----------------|
| Sismos | `cachedEarthquakes` / SSE | Marcadores con color por magnitud, pulse en selección |
| Placas tectónicas | `/plates.json` | Líneas rojas, carga diferida |
| Volcanes | `/volcanoes.json` | Triángulos con alerta USGS; carga bajo demanda |
| Daños Sentinel-1 | `/damage_sentinel1.json` | Polígonos con índice espacial de cuadrícula y filtrado por viewport |
| Heatmap | Sismos filtrados | `leaflet.heat` opcional |

### 5.3. Optimizaciones de mapa

- `plates.json` y `volcanoes.json` se cargan de forma diferida (`requestIdleCallback` / `setTimeout`).
- `damage_sentinel1.json` solo se descarga cuando el usuario activa la capa.
- Los polígonos de daño se indexan en una cuadrícula (`0.01° x 0.01°`) y se filtran por viewport para evitar recorrer ~58k polígonos.
- El renderizado del índice se hace por frames (`requestAnimationFrame` con presupuesto de 8 ms) para no bloquear el hilo principal.

---

## 6. Tiempo real y notificaciones

### 6.1. Server-Sent Events (`/api/stream`)

Al conectar, el backend envía el estado inicial:

- `earthquakes_update`
- `telegram_update`
- `volcano_alerts_update`

Después, los workers hacen `broadcastSSE` cuando hay nuevos datos. El frontend actualiza `sismoState` directamente.

### 6.2. Web Push

- `push.go` genera y sirve clave VAPID.
- El frontend suscribe usuarios y envía la suscripción a `POST /api/subscribe`.
- El backend puede enviar notificaciones push (por ejemplo, al detectar nueva alerta volcánica).

---

## 7. Fuentes de datos funcionales

| Fuente | Datos | Estado |
|--------|-------|--------|
| USGS | Sismos globales | Operativo |
| EMSC | Sismos Euro-Mediterráneo | Operativo |
| FUNVISIS | Sismos Venezuela | Operativo (puede fallar intermitentemente) |
| CSN | Sismos Chile | Operativo |
| PRSN | Sismos Caribe NE | Operativo |
| UWI SRC | Sismos Caribe Oriental | Operativo |
| IPGP | Sismos Antillas Francesas | Operativo |
| USGS Volcano Hazards | Alertas volcánicas | Operativo |
| Google News RSS | Noticias | Operativo (lento por fetch externo) |
| Telegram OSINT | Reportes de rescate/apoyo | Operativo vía SSE |

---

## 8. Metodología y convenciones

### 8.1. Desarrollo frontend

- **Componentes Svelte desacoplados:** cada componente encapsula su markup, lógica y estilos.
- **Svelte 5 runes:** `$state` para estado reactivo, `$effect` para side effects, `$props` para props.
- **Estado compartido:** todo estado global vive en `sismoStore.svelte.ts`; los componentes lo importan y mutan directamente.
- **CSS por componente:** los estilos específicos se mueven al `<style>` del componente; `app.css` solo conserva variables globales, layout base y utilidades.
- **CSS variables:** todas las decisiones de color pasan por `:root` y el atributo `data-theme` (`dark` / `light`).
- **`color-mix()` preferido:** se usa `color-mix(in srgb, var(--accent) X%, transparent)` en lugar de `rgba` fijos para respetar el tema.
- **Texto en español:** UI, comentarios y logs en español.

### 8.2. Rendimiento

- Carga diferida de recursos pesados (`volcanoes.json`, `plates.json`, `damage_sentinel1.json`).
- Índice espacial para polígonos de daños.
- Debounce en actualizaciones del carrusel de rescate (500 ms) y en eventos de mapa (300 ms).
- Límite de ~500 sismos renderizados en Leaflet.
- Caché del Service Worker para `/api/earthquakes` con TTL de 24 h.

### 8.3. Testing y calidad

```sh
# Frontend
cd frontend
npm run check   # TypeScript / Svelte diagnostics
npm run lint    # ESLint
npm run test    # Vitest
npm run build   # Producción

# Backend
cd backend
go vet ./...
go test ./...
go build -o sismoserver .
```

### 8.4. CI/CD

GitHub Actions en `.github/workflows/`:

- `frontend-ci.yml` → `check`, `lint`, `test`, `build`.
- `backend-ci.yml` → `go vet`, `go build`, `go test`.

### 8.5. Despliegue

- **Backend:** binario Go gestionado por PM2 en puerto `8082`.
- **Frontend:** servicio systemd `sismoapp.service` en puerto `3001` (el `3000` está ocupado por Grafana).
- **Proxy:** Nginx redirige `/api/*` al backend y el resto al frontend.

---

## 9. Funcionalidades operativas actualmente

- ✅ Monitoreo de sismos en tiempo real desde 7 fuentes.
- ✅ Mapa interactivo con capas de placas, volcanes, daños y heatmap.
- ✅ Selección de sismos/volcanes con paneles de detalle.
- ✅ Panel inferior con gráficas de magnitud, profundidad, línea temporal y energía acumulada.
- ✅ Noticias en tiempo real desde Google News.
- ✅ Carrusel de alertas OSINT de Telegram (Centro de Apoyo y Rescate).
- ✅ TV en vivo vía HLS con fallback a YouTube.
- ✅ Alertas volcánicas del USGS con banner y sonido.
- ✅ Notificaciones push (VAPID).
- ✅ Tema oscuro/claro con sincronización de capa base.
- ✅ Diseño responsive con navegación móvil.
- ✅ Caché local y Service Worker para sismos.

---

## 10. Archivos clave

| Archivo | Rol |
|---------|-----|
| `backend/main.go` | Inicialización, rutas, SSE |
| `backend/scraper.go` | Colectores de datos |
| `backend/db.go` | Esquema SQLite |
| `backend/api.go` | Handlers REST |
| `backend/push.go` | Web Push VAPID |
| `backend/middleware.go` | CORS, rate limiting |
| `frontend/src/lib/sismoStore.svelte.ts` | Estado global |
| `frontend/src/lib/api.ts` | Cliente API |
| `frontend/src/routes/+page.svelte` | Dashboard principal |
| `frontend/src/app.css` | Variables y estilos globales |
| `frontend/static/sw.js` | Service Worker |
