# Plan de mejoras futuras — SismoMonitor

> Este documento resume la deuda técnica y mejoras identificadas durante la refactorización de `refactor/componentes-frontend`. Se atenderán progresivamente en iteraciones posteriores.

## 1. Tipado fuerte (alta prioridad)

- Reducir el uso de `any` en `+page.svelte`, componentes y utilidades.
- Tipar correctamente `map`, `L`, `markersLayer`, `volcanoesLayer` usando tipos de Leaflet (`L.Map`, `typeof L`, `L.GeoJSON`, etc.).
- Usar las interfaces definidas en `src/lib/types.ts` en lugar de objetos genéricos.
- Revisar y corregir los 57 warnings de ESLint, priorizando `@typescript-eslint/no-explicit-any`.

## 2. Cobertura de tests (alta prioridad)

- Agregar tests de componentes:
  - `Header`
  - `VolcanoDetails`
  - `EarthquakeDetails`
  - `RankingWidget`
  - `NewsFeed`
  - `RescueCarousel`
  - `VolcanoLayer` (con mocks de Leaflet)
- Mantener tests de utilidades (`geoUtils`, `mapUtils`, `volcanoUtils`) al día.
- Considerar tests de integración para flujos críticos: selección de sismo, carga de volcanes, SSE.

## 3. Estado del mapa centralizado (media prioridad)

- Evaluar crear un `MapContext` o un store (`mapStore`) para manejar:
  - `map`, `L`, `mapReady`
  - `mapTheme`, `showPlates`, `showVolcanoes`, `showDamage`
  - capas activas (`markersLayer`, `platesLayer`, `volcanoesLayer`, `damageLayer`)
- Esto reducirá el *props drilling* entre `+page.svelte` y los componentes de capas.

## 4. Carga de datos volcánicos (media prioridad)

- Mover `loadVolcanoData` del componente padre a `src/lib/api.ts` como `getVolcanoes()`.
- Que `VolcanoLayer` consuma directamente la API centralizada.
- Evitar pasar callbacks de carga como props.

## 5. URLs externas configurables (media prioridad)

- Extraer URLs de streams de TV en vivo de `LiveTV.svelte` a un archivo de configuración (`src/lib/config.ts`) o variables de entorno.
- Incluir:
  - Streams HLS por país.
  - IDs de DailyMotion.
  - URL del fallback.
- Esto facilita cambios sin tocar código.

## 6. Gestión de secrets y entorno (alta prioridad)

- Nunca compartir API keys en conversaciones ni commits.
- Usar `.env` para variables sensibles (VAPID keys, API keys de graphify, etc.).
- Asegurar que `.env` esté en `.gitignore`.
- Documentar variables requeridas en `README.md`.

## 7. Documentación de graphify (baja prioridad)

- Agregar al `README.md` una sección sobre cómo regenerar el grafo de conocimiento.
- Indicar que requiere una API key de Gemini/Google (u otra soportada).
- Incluir el comando: `GEMINI_API_KEY=... graphify /ruta --update`.

## 8. Smoke tests manuales en producción (siempre)

Tras cada merge a `main`, verificar:
- [ ] Feed carga sismos en vivo.
- [ ] Filtros de feed funcionan.
- [ ] Mapa renderiza y capas de placas/volcanes/daños se activan.
- [ ] Click en sismo/volcán abre detalles.
- [ ] TV en vivo carga o degrada correctamente.
- [ ] SSE mantiene conexión y actualiza datos.
- [ ] No hay errores rojos en consola del navegador.

## 9. Posibles optimizaciones futuras

- **Bundle size**: revisar si Leaflet y Chart.js se cargan de forma eficiente.
- **Lazy loading**: cargar componentes pesados (TV, daños) solo cuando se usan.
- **PWA**: revisar service worker y manifiesto si se quiere experiencia instalable.
- **Accesibilidad**: mejorar contrastes, labels y navegación por teclado.

## 10. Seguridad y escalabilidad

- Revisar CSP en producción tras cambios de fuentes externas (iframes, videos, scripts).
- Mantener rate limiting del backend y CORS configurados.
- Monitorear logs de nginx y backend para detectar errores tras despliegues.

---

*Última actualización: 2026-07-02*
