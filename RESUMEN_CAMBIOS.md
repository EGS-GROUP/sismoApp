# Resumen de cambios - SismoMonitor

Fecha: 2026-07-01

## 1. Capa de volcanes

### Funcionalidad
- Se agregó estado `selectedVolcano` para permitir seleccionar volcanes en el mapa.
- Al hacer clic en un volcán se muestra su ficha en la pestaña "Detalles", se limpia la selección de sismos y se hace zoom al volcán.
- Al hacer clic en un sismo se limpia la selección de volcanes.
- El volcán seleccionado se resalta con borde cyan y glow.

### Estilo
- Se quitó el emoji del título de la capa ("Volcanes Activos" sin emoji).
- Se reemplazó el emoji del título del panel por el icono `Mountain` de Lucide.
- Se quitó el emoji del tooltip de volcanes.
- Se cambió el borde blanco del triángulo por borde negro, igual que los sismos.
- Se redujo la animación de pulso para que sea más sutil.
- Se aumentó el tamaño de los iconos de volcanes.

### Datos
- Se reemplazó el campo confuso `last_eruption` (que venía con códigos tipo "D1", "U", "Q") por el campo `status` (por ejemplo: Historical, Holocene).
- Se usó `findVolcanoAlert` para mostrar la alerta USGS en el panel de detalles.

### Alertas reales
- Se implementaron alertas de verdad para volcanes:
  - Sonido `/alerta.mp3` al detectar nueva alerta o escalada.
  - Notificación push (si está habilitada y permitida).
  - Banner prominente con color según nivel (rojo/naranja/amarillo) y sinopsis.
- No spamea: la primera carga no alerta, solo los cambios posteriores.

## 2. Capa de daños Sentinel-1

### Rendimiento
- Se agregó un índice espacial de cuadrícula (`0.01° x 0.01°`) sobre los 58,870 polígonos.
- El filtrado por viewport ahora solo revisa las celdas que intersectan la vista, no todos los polígonos.
- Se mantiene el zoom mínimo 8 para evitar saturar la pantalla.
- Se mantiene el debounce de 300 ms en `moveend`.

### Indexación sin bloqueos
- La construcción del índice se hace de forma asíncrona usando `requestAnimationFrame` con un presupuesto de 8 ms por frame.
- Se muestra progreso en la leyenda: "Indexando 45%...".
- Esto elimina los congelamientos de ~1.4 segundos que causaban los `[Violation] 'setTimeout' handler took ...`.

## 3. Verificación

- `npm run check` pasa con 0 errores y 0 warnings.
- `npm run build` exitoso.
- Frontend reiniciado en puerto 3001.
- Backend `sismoserver` en puerto 8082.

## 4. Análisis de rendimiento identificado

### Recursos pesados
- `/damage_sentinel1.json`: 13.8 MB.
- `/volcanoes.json`: 364 KB.
- `/plates.json`: 222 KB.

### Tiempos de endpoints locales
- `/api/earthquakes`: 2.3 ms.
- `/api/news`: 470 ms (lento por fetch externo a Google News).
- `/volcanoes.json`: 2 ms.
- `/damage_sentinel1.json`: 22 ms.

### Posibles optimizaciones futuras (compatibles con tiempo real)
- Cargar Chart.js de forma lazy (no se usa inmediatamente).
- Reducir forced reflows causados por el marquee de rescates/noticias.
- Cache corto o SSE para noticias, según preferencia de tiempo real.
- Virtualizar o limitar polígonos de daños en móvil.
- Agregar índices en SQLite para consultas históricas.
- Reducir frecuencia de polling de Telegram.

## 5. Mejora de mantenibilidad: CSS por componente

### Cambios
- Se movieron los estilos específicos de componentes desde `app.css` a sus respectivos archivos `.svelte`:
  - `RescueCarousel.svelte`: todos los estilos del carrusel y las tarjetas de rescate.
  - `NewsFeed.svelte`: estilos del feed de noticias (ya estaban en el componente; se eliminaron duplicados de `app.css`).
  - `Modal.svelte`: estilos del modal, eliminando tres definiciones duplicadas en `app.css`.
  - `VolcanoLayer.svelte`: estilos globales de popups y tooltips de Leaflet con `:global()`.
- Se eliminaron estilos muertos y duplicados de `app.css` (clases de modal y rescue de versiones anteriores).
- Se redujo el CSS global de layout de ~20.7 kB a ~11.4 kB (compilado), delegando los estilos de componentes a cada componente.

## 6. Optimización de carga inicial

### Cambios
- Se evita la precarga de `volcanoes.json` solo por la llegada de alertas; ahora solo se carga cuando el usuario activa la capa de volcanes o el gráfico de volcanes.
- Se retrasa la carga de `plates.json` usando `requestIdleCallback` (o `setTimeout` fallback) para no competir con la pintura inicial.
- Se agregó `contain: layout paint` al contenedor del carrusel para reducir el alcance de los reflows.
- Se limitó el carrusel a un máximo de 15 alertas visibles.
- Se retrasó el inicio de la animación del carrusel 300 ms para no competir con el layout inicial.
- Se agregó `loading="lazy"` y `decoding="async"` a las imágenes del carrusel.
- Se agregó `will-change: transform` al track del carrusel.
- Se limitó a 10 el número de alertas renderizadas en el carrusel.
- Se agregó debounce de 500 ms en las actualizaciones del carrusel para evitar reflows en cada tick de SSE.
- Se limitó a 500 el número de sismos renderizados en el mapa de Leaflet (`mapEarthquakes`).
- Se reordenó `onMount` en `+page.svelte` para cargar secuencialmente: primero `/api/earthquakes`, luego Leaflet, luego SSE, y finalmente recursos no críticos vía `requestIdleCallback`.
- Se agregó caché en el Service Worker para `/api/earthquakes` con TTL de 24 horas, alineado con la ventana de datos en vivo.
- Se agregó flag `volcanoDataLoading` en `+page.svelte` para evitar cargas duplicadas de `volcanoes.json`.

## 7. Ajuste de velocidad del carrusel Centro de Apoyo y Rescate

### Cambio
- Se redujo la velocidad de la animación marquee de 40 s a 120 s.
- Se extrajo la duración a una variable CSS (`--marquee-duration`) en `.rescue-marquee-track` para facilitar ajustes futuros.
- Se movieron los estilos inline del carrusel a clases CSS dentro del componente.

## 8. Corrección de error `effect_update_depth_exceeded` tras refactorización

### Causa
- En `MagnitudeChart.svelte`, `magChartInstance` se declaró como `$state(null)`.
- El `$effect` que actualiza el gráfico llama `updateChart()`.
- Dentro de `updateChart()` se lee `magChartInstance` (`if (magChartInstance) magChartInstance.destroy();`) y luego se escribe (`magChartInstance = new Chart(...)`).
- En Svelte 5, leer y escribir el mismo estado dentro de un `$effect` crea un bucle infinito de reactividad (`Maximum update depth exceeded`).

### Corrección
- `MagnitudeChart.svelte`: `magChartInstance` pasa de `$state` a variable normal (`let magChartInstance: any = null;`), ya que no se usa en el template.
- `VolcanoLayer.svelte`: `volcanoesLayer`, `volcanoAlertMarkersLayer` y `volcanoAlertRenderTimeout` pasan de `$state` a variables normales, por la misma razón.
- `+page.svelte`: `damageLayer` pasa de `$state` a variable normal (`let damageLayer: any = null;`).

### Verificación
- Se reprodujo el error con Playwright en build de producción.
- Tras la corrección, `npm run check` pasa con 0 errores y 0 warnings.
- Build de producción exitoso.
- El error `effect_update_depth_exceeded` ya no aparece en la consola del navegador.

## 6. Archivos editados

- `/var/www/html/sismoApp/frontend/src/app.css`
- `/var/www/html/sismoApp/frontend/src/routes/+page.svelte`
- `/var/www/html/sismoApp/frontend/src/lib/components/MagnitudeChart.svelte`
- `/var/www/html/sismoApp/frontend/src/lib/components/VolcanoLayer.svelte`
- `/var/www/html/sismoApp/frontend/src/lib/components/RescueCarousel.svelte`
- `/var/www/html/sismoApp/frontend/src/lib/components/Modal.svelte`
- `/var/www/html/sismoApp/frontend/static/sw.js`
- `/var/www/html/sismoApp/RESUMEN_CAMBIOS.md`
