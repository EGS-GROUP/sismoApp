<script lang="ts">
	import { getDamageData } from '$lib/api';

	interface Props {
		map: any;
		L: any;
		showDamage?: boolean;
		damageLoading?: boolean;
		damageError?: string;
		damageIndexProgress?: number;
	}

	let {
		map,
		L,
		showDamage = $bindable(false),
		damageLoading = $bindable(false),
		damageError = $bindable(''),
		damageIndexProgress = $bindable(0)
	}: Props = $props();

	const DAMAGE_CELL_SIZE = 0.01; // degrees per grid cell
	const DAMAGE_INDEX_BUDGET = 8; // ms per frame while indexing

	let damageLayer: any = null;
	let allDamageFeatures: any[] | null = null;
	let damageSpatialIndex: Map<string, any[]> | null = null;

	function debounce<T extends (...args: any[]) => void>(fn: T, wait: number) {
		let timeout: ReturnType<typeof setTimeout> | null = null;
		return (...args: Parameters<T>) => {
			if (timeout) clearTimeout(timeout);
			timeout = setTimeout(() => fn(...args), wait);
		};
	}

	function getDamageFeatureCoords(f: any): { lat: number; lng: number } | null {
		const coords = f.geometry?.coordinates;
		if (!coords) return null;
		const ring = Array.isArray(coords[0]) && Array.isArray(coords[0][0]) ? coords[0] : coords;
		const first = Array.isArray(ring[0]) ? ring[0] : ring;
		const [lng, lat] = first;
		if (typeof lat !== 'number' || typeof lng !== 'number') return null;
		return { lat, lng };
	}

	function buildDamageSpatialIndexAsync(
		features: any[],
		index: Map<string, any[]>,
		onProgress: (pct: number) => void
	): Promise<Map<string, any[]>> {
		return new Promise((resolve) => {
			let start = 0;

			function processFrame() {
				const frameStart = performance.now();
				while (start < features.length) {
					const f = features[start];
					const c = getDamageFeatureCoords(f);
					if (c) {
						const cellKey = `${Math.floor(c.lat / DAMAGE_CELL_SIZE)},${Math.floor(c.lng / DAMAGE_CELL_SIZE)}`;
						if (!index.has(cellKey)) index.set(cellKey, []);
						index.get(cellKey)!.push(f);
					}
					start++;
					if (performance.now() - frameStart >= DAMAGE_INDEX_BUDGET) break;
				}

				const pct = Math.round((start / features.length) * 100);
				onProgress(pct);

				if (start < features.length) {
					requestAnimationFrame(processFrame);
				} else {
					resolve(index);
				}
			}

			requestAnimationFrame(processFrame);
		});
	}

	function renderDamageInViewport() {
		if (!map || !L || !allDamageFeatures) return;
		if (damageLayer) {
			map.removeLayer(damageLayer);
			damageLayer = null;
		}
		const bounds = map.getBounds();
		const zoom = map.getZoom();
		if (zoom < 8) {
			damageError = 'Zoom al menos a nivel 8 para ver estructuras';
			return;
		}
		damageError = '';

		let visibleFeatures: any[] = [];

		if (damageSpatialIndex) {
			const minLat = Math.floor(bounds.getSouth() / DAMAGE_CELL_SIZE);
			const maxLat = Math.floor(bounds.getNorth() / DAMAGE_CELL_SIZE);
			const minLng = Math.floor(bounds.getWest() / DAMAGE_CELL_SIZE);
			const maxLng = Math.floor(bounds.getEast() / DAMAGE_CELL_SIZE);
			for (let latCell = minLat; latCell <= maxLat; latCell++) {
				for (let lngCell = minLng; lngCell <= maxLng; lngCell++) {
					const cell = damageSpatialIndex.get(`${latCell},${lngCell}`);
					if (!cell) continue;
					for (const f of cell) {
						const c = getDamageFeatureCoords(f);
						if (c && bounds.contains([c.lat, c.lng])) visibleFeatures.push(f);
					}
				}
			}
		} else {
			visibleFeatures = allDamageFeatures.filter((f: any) => {
				const c = getDamageFeatureCoords(f);
				return c && bounds.contains([c.lat, c.lng]);
			});
		}

		if (visibleFeatures.length === 0) {
			damageError = 'No hay estructuras dañadas en esta vista';
			return;
		}

		const geojson = { type: 'FeatureCollection', features: visibleFeatures };
		damageLayer = L.geoJSON(geojson, {
			style: (feature: any) => {
				const p = feature.properties?.damage_probability ?? 0;
				let color = '#fbbf24'; // amarillo
				if (p >= 0.75) color = '#dc2626'; // rojo
				else if (p >= 0.5) color = '#f97316'; // naranja
				return {
					color: color,
					weight: 1,
					opacity: 0.8,
					fillColor: color,
					fillOpacity: 0.35
				};
			},
			onEachFeature: (feature: any, layer: any) => {
				const p = feature.properties?.damage_probability ?? 0;
				const pct = (p * 100).toFixed(1);
				layer.bindTooltip(`Probabilidad de daño: ${pct}%`, { sticky: true });
			}
		});
		map.addLayer(damageLayer);
	}

	async function loadDamageLayer() {
		if (!map || !L || !showDamage) return;
		damageLoading = true;
		damageError = '';
		damageIndexProgress = 0;
		try {
			if (allDamageFeatures) {
				renderDamageInViewport();
				damageLoading = false;
				return;
			}
			const data = await getDamageData();
			allDamageFeatures = data.features || [];
			if (allDamageFeatures.length === 0) {
				damageError = 'No hay datos de daño disponibles';
				damageLoading = false;
				return;
			}
			damageSpatialIndex = await buildDamageSpatialIndexAsync(
				allDamageFeatures,
				new Map<string, any[]>(),
				(pct) => { damageIndexProgress = pct; }
			);
			damageIndexProgress = 100;
			console.log(`Capa Sentinel-1: ${allDamageFeatures.length} estructuras dañadas cargadas localmente`);
			renderDamageInViewport();
		} catch (err: any) {
			console.error('Error cargando capa de daños local:', err);
			damageError = 'Error cargando capa de daños';
		} finally {
			damageLoading = false;
		}
	}

	const debouncedRenderDamage = debounce(() => renderDamageInViewport(), 300);

	$effect(() => {
		if (map && L) {
			if (showDamage) {
				loadDamageLayer();
				map.on('moveend', debouncedRenderDamage);
			} else {
				map.off('moveend', debouncedRenderDamage);
				if (damageLayer) {
					map.removeLayer(damageLayer);
					damageLayer = null;
					damageError = '';
				}
			}
			return () => {
				if (map) map.off('moveend', debouncedRenderDamage);
			};
		}
	});
</script>
