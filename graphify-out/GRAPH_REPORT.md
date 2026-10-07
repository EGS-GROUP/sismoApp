# Graph Report - /var/www/html/sismoApp  (2026-07-02)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 341 nodes · 503 edges · 25 communities (21 shown, 4 thin omitted)
- Extraction: 83% EXTRACTED · 17% INFERRED · 0% AMBIGUOUS · INFERRED: 87 edges (avg confidence: 0.82)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `9b528a50`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- [[_COMMUNITY_Frontend dev dependencies|Frontend dev dependencies]]
- [[_COMMUNITY_Frontend runtime dependencies|Frontend runtime dependencies]]
- [[_COMMUNITY_Backend API handlers|Backend API handlers]]
- [[_COMMUNITY_Backend scraper and SSE|Backend scraper and SSE]]
- [[_COMMUNITY_Frontend shell and PWA|Frontend shell and PWA]]
- [[_COMMUNITY_PWA manifest|PWA manifest]]
- [[_COMMUNITY_TypeScript configuration|TypeScript configuration]]
- [[_COMMUNITY_Live TV and HLS.js|Live TV and HLS.js]]
- [[_COMMUNITY_Push notifications|Push notifications]]
- [[_COMMUNITY_Favicon image assets|Favicon image assets]]
- [[_COMMUNITY_Documentation and deployment|Documentation and deployment]]
- [[_COMMUNITY_Database and schema|Database and schema]]
- [[_COMMUNITY_FUNVISIS data and fix script|FUNVISIS data and fix script]]
- [[_COMMUNITY_PWA manifest icons|PWA manifest icons]]
- [[_COMMUNITY_Favicon brand identity|Favicon brand identity]]
- [[_COMMUNITY_Svelte layout shell|Svelte layout shell]]
- [[_COMMUNITY_Community 16|Community 16]]
- [[_COMMUNITY_Community 17|Community 17]]
- [[_COMMUNITY_Community 18|Community 18]]
- [[_COMMUNITY_Community 19|Community 19]]
- [[_COMMUNITY_Community 22|Community 22]]

## God Nodes (most connected - your core abstractions)
1. `main()` - 15 edges
2. `scripts` - 14 edges
3. `SismoMonitor Dashboard Page` - 14 edges
4. `fetchEarthquakeData()` - 13 edges
5. `frontend/src/routes/+page.svelte` - 12 edges
6. `compilerOptions` - 11 edges
7. `sseHandler()` - 10 edges
8. `historicalHandler()` - 9 edges
9. `ResponseWriter` - 8 edges
10. `Request` - 8 edges

## Surprising Connections (you probably didn't know these)
- `SvelteKit Adapter Node` --conceptually_related_to--> `SvelteKit Frontend (port 3000, PM2)`  [INFERRED]
  frontend/vite.config.ts → README.md
- `HLS CORS-bypass Proxy` --conceptually_related_to--> `hls.min.js (HLS.js video library)`  [INFERRED]
  AGENTS.md → frontend/static/hls.min.js
- `FUNVISIS Unconventional Field Mapping Quirk` --conceptually_related_to--> `FUNVISIS Earthquake History Dataset`  [INFERRED]
  AGENTS.md → funvisis_history.json
- `FUNVISIS Unconventional Field Mapping Quirk` --conceptually_related_to--> `FUNVISIS Seismic Data Source`  [INFERRED]
  AGENTS.md → funvisis_history.json
- `HLS CORS-bypass Proxy` --semantically_similar_to--> `HLS.js Served Locally (avoids tracking prevention)`  [INFERRED] [semantically similar]
  AGENTS.md → frontend/src/app.html

## Import Cycles
- None detected.

## Communities (25 total, 4 thin omitted)

### Community 0 - "Frontend dev dependencies"
Cohesion: 0.10
Nodes (23): ApiError, getDamageData(), getNews(), getPlates(), getVolcanoAlerts(), getVolcanoes(), handleResponse(), Earthquake (+15 more)

### Community 1 - "Frontend runtime dependencies"
Cohesion: 0.13
Nodes (24): Earthquake, broadcastSSE(), getActiveCount(), Request, ResponseWriter, Server-Sent Events Streaming Architecture, sseHandler(), Earthquake Deduplication Strategy (+16 more)

### Community 2 - "Backend API handlers"
Cohesion: 0.09
Nodes (25): SvelteKit Frontend README (sv scaffold), SvelteKit Adapter Node, Svelte Runes Mode, initLiveTV, stopAllTV, triggerDegradation, AGENTS.md - SismoMonitor Architecture Reference, HLS CORS-bypass Proxy (+17 more)

### Community 3 - "Backend scraper and SSE"
Cohesion: 0.13
Nodes (6): active, $lib/api, $lib/types, $lib/utils/geoUtils, $lib/utils/mapUtils, $lib/utils/volcanoUtils

### Community 4 - "Frontend shell and PWA"
Cohesion: 0.16
Nodes (24): Request, ResponseWriter, healthHandler(), historicalHandler(), newsHandler(), proxyHandler(), statsHandler(), subscribeHandler() (+16 more)

### Community 5 - "PWA manifest"
Cohesion: 0.11
Nodes (20): TestMain(), getFunvisisHistory7Days(), getFunvisisHistoryByDate(), getVolcanoAlerts(), initDB(), migrateHistoryIfNeeded(), saveFunvisisHistory(), saveVolcanoAlerts() (+12 more)

### Community 6 - "TypeScript configuration"
Cohesion: 0.09
Nodes (23): devDependencies, eslint, eslint-config-prettier, @eslint/js, eslint-plugin-svelte, globals, jsdom, prettier (+15 more)

### Community 7 - "Live TV and HLS.js"
Cohesion: 0.12
Nodes (21): Svelte framework, Svelte logo favicon, Svelte logo favicon SVG, Android Chrome 192x192 Icon, PNG Image Format, App Favicon Icon Set, 192x192 Icon Size, Android Chrome 512x512 Icon (+13 more)

### Community 8 - "Push notifications"
Cohesion: 0.14
Nodes (17): allowedProxyDomain(), corsMiddleware(), getClientIP(), Request, Time, maxBodyMiddleware(), newRateLimiter(), rateLimitMiddleware() (+9 more)

### Community 9 - "Favicon image assets"
Cohesion: 0.11
Nodes (18): name, private, scripts, analyze, build, check, check:watch, dev (+10 more)

### Community 10 - "Documentation and deployment"
Cohesion: 0.13
Nodes (11): chart.js/auto, dependencies, chart.js, leaflet, @lucide/svelte, @types/leaflet, calculateRanking, extractRegion (+3 more)

### Community 11 - "Database and schema"
Cohesion: 0.13
Nodes (14): background_color, categories, description, dir, display, icons, id, lang (+6 more)

### Community 12 - "FUNVISIS data and fix script"
Cohesion: 0.22
Nodes (12): frontend/static/damage_sentinel1.json, frontend/src/lib/components/MagnitudeChart.svelte, UI Mockup Dashboard, frontend/src/lib/components/Modal.svelte, frontend/src/lib/components/NewsFeed.svelte, frontend/src/routes/+page.svelte, frontend/static/plates.json, frontend/src/lib/components/RescueCarousel.svelte (+4 more)

### Community 13 - "PWA manifest icons"
Cohesion: 0.15
Nodes (12): compilerOptions, allowJs, checkJs, esModuleInterop, forceConsistentCasingInFileNames, moduleResolution, resolveJsonModule, rewriteRelativeImportExtensions (+4 more)

### Community 14 - "Favicon brand identity"
Cohesion: 0.24
Nodes (9): getSubscriptions(), saveSubscription(), sendPushNotificationToAll(), PushSubscription, Web Push Notification Flow, subscribeToPush, toggleNotifications, Service Worker (+1 more)

### Community 15 - "Svelte layout shell"
Cohesion: 0.60
Nodes (3): extractRegion(), extractVenezuelaState(), usStates

### Community 17 - "Community 17"
Cohesion: 0.67
Nodes (3): SismoApp brand identity, SismoApp frontend web application, SismoApp favicon icon

## Knowledge Gaps
- **121 isolated node(s):** `fs`, `txt`, `badIdx`, `ResponseWriter`, `Request` (+116 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **4 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `AGENTS.md - SismoMonitor Architecture Reference` connect `Backend API handlers` to `PWA manifest`?**
  _High betweenness centrality (0.294) - this node is a cross-community bridge._
- **Why does `FUNVISIS Unconventional Field Mapping Quirk` connect `PWA manifest` to `Backend API handlers`?**
  _High betweenness centrality (0.286) - this node is a cross-community bridge._
- **Are the 8 inferred relationships involving `main()` (e.g. with `initDB()` and `corsMiddleware()`) actually correct?**
  _`main()` has 8 INFERRED edges - model-reasoned connections that need verification._
- **Are the 6 inferred relationships involving `SismoMonitor Dashboard Page` (e.g. with `Earthquake Deduplication Strategy` and `Frontend Package Config`) actually correct?**
  _`SismoMonitor Dashboard Page` has 6 INFERRED edges - model-reasoned connections that need verification._
- **Are the 5 inferred relationships involving `fetchEarthquakeData()` (e.g. with `historicalHandler()` and `getFunvisisHistory7Days()`) actually correct?**
  _`fetchEarthquakeData()` has 5 INFERRED edges - model-reasoned connections that need verification._
- **What connects `fs`, `txt`, `badIdx` to the rest of the system?**
  _123 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Frontend dev dependencies` be split into smaller, more focused modules?**
  _Cohesion score 0.10338680926916222 - nodes in this community are weakly interconnected._