<script lang="ts">
	import { volcanoAlertStyleMap, normalizeVolcanoName, findVolcanoAlert, findVolcanoFeatureByName } from '$lib/utils/volcanoUtils';
	import type { GeoJSONCollection, GeoJSONFeature, VolcanoAlert, VolcanoProperties } from '$lib/types';

	interface Props {
		map: any;
		L: any;
		volcanoData: GeoJSONCollection | null;
		volcanoAlerts: VolcanoAlert[];
		showVolcanoes: boolean;
		selectedVolcano: VolcanoProperties | null;
		onSelectVolcano?: (volcano: VolcanoProperties, coords: [number, number]) => void;
		onLoadVolcanoData?: () => Promise<void>;
	}

	let { map, L, volcanoData, volcanoAlerts, showVolcanoes, selectedVolcano, onSelectVolcano, onLoadVolcanoData }: Props = $props();

	let volcanoesLayer: any = null;
	let volcanoAlertMarkersLayer: any = null;
	let volcanoAlertRenderTimeout: ReturnType<typeof setTimeout> | null = null;

	async function loadVolcanoesLayer() {
		if (!map || !L) return;
		if (volcanoesLayer) { map.addLayer(volcanoesLayer); return; }
		if (!volcanoData && onLoadVolcanoData) {
			await onLoadVolcanoData();
		}
		renderVolcanoesLayer();
	}

	function renderVolcanoesLayer() {
		if (!map || !L || !volcanoData) return;
		if (volcanoesLayer) {
			map.removeLayer(volcanoesLayer);
			volcanoesLayer = null;
		}

		volcanoesLayer = L.geoJSON(volcanoData, {
			pointToLayer: (feature: GeoJSONFeature, latlng: any) => {
				const p = (feature.properties || {}) as VolcanoProperties;
				const alert = findVolcanoAlert(p.name || '', volcanoAlerts);
				const isSelected = selectedVolcano && selectedVolcano.name === p.name;
				let color = '#ff6347';
				let size = isSelected ? 28 : 20;
				let cssClass = '';
				let stroke = isSelected ? '#00e5ff' : '#000';
				let strokeWidth = isSelected ? 3 : 1;
				let markerLatLng = latlng;

				if (alert) {
					const alertKey = (alert.color_code || alert.alert_level || '').toUpperCase();
					const style = volcanoAlertStyleMap[alertKey] || volcanoAlertStyleMap['YELLOW'];
					color = style.color;
					size = isSelected ? style.radius * 2 + 8 : style.radius * 2;
					if (isSelected) cssClass += ' volcano-selected';
					if (alert.lat && alert.lon) {
						markerLatLng = [alert.lat, alert.lon];
					}
				} else if (isSelected) {
					cssClass += ' volcano-selected';
				}

				const icon = L.divIcon({
					className: 'volcano-triangle' + cssClass,
					html: `<svg width="${size}" height="${size}" viewBox="0 0 24 24" style="filter: drop-shadow(0 0 4px rgba(0,0,0,0.9));"><polygon points="12,2 22,22 2,22" fill="${color}" stroke="${stroke}" stroke-width="${strokeWidth}"/></svg>`,
					iconSize: [size, size],
					iconAnchor: [size / 2, size / 2]
				});

				const popupColor = alert ? (volcanoAlertStyleMap[(alert.alert_level || alert.color_code || '').toUpperCase()] || {}).color || '#ff8c00' : '#ff6347';
				const popupHtml = `
					<div style="min-width: 200px; font-family: 'Outfit', sans-serif;">
						<div style="font-size: 14px; font-weight: 800; color: var(--text-primary); margin-bottom: 6px; display: flex; align-items: center; gap: 6px;">
							<svg width="16" height="16" viewBox="0 0 24 24" fill="${popupColor}"><polygon points="12,2 22,22 2,22"/></svg>
							${p.name}
						</div>
						<div style="font-size: 11px; color: var(--text-secondary); margin-bottom: 8px;">
							${p.country || 'País desconocido'}${p.region ? ` — ${p.region}` : ''}
						</div>
						<div style="display: grid; grid-template-columns: 1fr 1fr; gap: 8px; font-size: 11px; margin-bottom: 8px;">
							<div style="background: var(--bg-input); padding: 6px; border-radius: 4px;">
								<div style="color: var(--text-secondary); text-transform: uppercase; font-size: 9px;">Tipo</div>
								<div style="color: var(--text-primary); font-weight: 600;">${p.type || 'N/D'}</div>
							</div>
							<div style="background: var(--bg-input); padding: 6px; border-radius: 4px;">
								<div style="color: var(--text-secondary); text-transform: uppercase; font-size: 9px;">Elevación</div>
								<div style="color: var(--text-primary); font-weight: 600;">${p.elevation ? p.elevation.toLocaleString() : 'N/D'} m</div>
							</div>
						</div>
						${alert ? `<div style="font-size: 11px; color: ${popupColor}; font-weight: 700;">Alerta: ${alert.alert_level || alert.color_code || 'N/A'}</div>` : ''}
					</div>
				`;

				const marker = L.marker(markerLatLng, { icon, zIndexOffset: 1000 });
				marker.bindPopup(popupHtml, { className: 'volcano-popup', closeButton: true, maxWidth: 260 });
				marker.on('click', () => {
					const selected = alert ? { ...p, ...alert, name: p.name } : p;
					const coords: [number, number] = [markerLatLng.lat || markerLatLng[0], markerLatLng.lng || markerLatLng[1]];
					onSelectVolcano?.(selected, coords);
				});
				return marker;
			},
			onEachFeature: (feature: GeoJSONFeature, layer: any) => {
				const p = (feature.properties || {}) as VolcanoProperties;
				const alert = findVolcanoAlert(p.name || '', volcanoAlerts);
				let tooltipHtml = `<b>${p.name || 'Volcán'}</b>`;
				if (alert) {
					const level = alert.alert_level || alert.color_code || 'N/A';
					tooltipHtml += `<br><span style="color:${(volcanoAlertStyleMap[level.toUpperCase()] || {}).color || '#ff8c00'};">Alerta: ${level}</span>`;
				}
				tooltipHtml += `<br><span style="font-size:11px;color: var(--text-secondary);">Clic para detalles</span>`;
				layer.bindTooltip(tooltipHtml, { sticky: true, className: 'volcano-tooltip' });
			}
		}).addTo(map);
		console.log(`Volcanes: ${volcanoData.features.length} volcanes del Holoceno cargados (${volcanoAlerts.length} con alerta)`);
		volcanoAlerts.forEach(a => {
			const found = volcanoData.features.find((f: any) => {
				const p = f.properties || {};
				return p.name?.toLowerCase() === a.name?.toLowerCase() ||
					normalizeVolcanoName(p.name || '') === normalizeVolcanoName(a.name || '');
			});
			console.log(`  Alerta: ${a.name} (${a.alert_level || a.color_code}) - ${found ? 'MAPEADO' : 'NO MAPEADO'} en volcanoes.json`);
		});
	}

	function renderVolcanoAlertMarkers() {
		if (!map || !L) return;
		if (volcanoAlertRenderTimeout) clearTimeout(volcanoAlertRenderTimeout);
		volcanoAlertRenderTimeout = setTimeout(() => {
			if (volcanoAlertMarkersLayer) {
				map.removeLayer(volcanoAlertMarkersLayer);
				volcanoAlertMarkersLayer = null;
			}
			if (volcanoAlerts.length === 0) return;

			volcanoAlertMarkersLayer = L.layerGroup();
			volcanoAlerts.forEach((alert: VolcanoAlert) => {
				if (!alert.lat || !alert.lon) return;
				const level = (alert.alert_level || alert.color_code || '').toUpperCase();
				const style = volcanoAlertStyleMap[level] || volcanoAlertStyleMap['YELLOW'];
				const isSelected = selectedVolcano && selectedVolcano.name === alert.name;
				const size = isSelected ? style.radius * 2 + 8 : style.radius * 2;
				const color = style.color;
				const stroke = isSelected ? '#00e5ff' : '#000';
				const strokeWidth = isSelected ? 3 : 1;
				const cssClass = 'volcano-alert-marker' + (isSelected ? ' volcano-selected' : '');

				const icon = L.divIcon({
					className: cssClass,
					html: `<svg width="${size}" height="${size}" viewBox="0 0 24 24" style="filter: drop-shadow(0 0 4px rgba(0,0,0,0.9));"><polygon points="12,2 22,22 2,22" fill="${color}" stroke="${stroke}" stroke-width="${strokeWidth}"/></svg>`,
					iconSize: [size, size],
					iconAnchor: [size / 2, size / 2]
				});

				const marker = L.marker([alert.lat, alert.lon], { icon, zIndexOffset: 2000 });
				const feature = findVolcanoFeatureByName(alert.name, volcanoData);
				const vData = feature?.properties || {};
				const popupHtml = `
					<div style="min-width: 220px; font-family: 'Outfit', sans-serif;">
						<div style="font-size: 14px; font-weight: 800; color: var(--text-primary); margin-bottom: 6px; display: flex; align-items: center; gap: 6px;">
							<svg width="16" height="16" viewBox="0 0 24 24" fill="${color}"><polygon points="12,2 22,22 2,22"/></svg>
							${alert.name}
						</div>
						<div style="font-size: 11px; color: var(--text-secondary); margin-bottom: 8px;">
							${vData.country || 'País desconocido'}${vData.region ? ` — ${vData.region}` : ''}
						</div>
						<div style="display: grid; grid-template-columns: 1fr 1fr; gap: 8px; font-size: 11px; margin-bottom: 10px;">
							<div style="background: var(--bg-input); padding: 6px; border-radius: 4px;">
								<div style="color: var(--text-secondary); text-transform: uppercase; font-size: 9px;">Tipo</div>
								<div style="color: var(--text-primary); font-weight: 600;">${vData.type || 'N/D'}</div>
							</div>
							<div style="background: var(--bg-input); padding: 6px; border-radius: 4px;">
								<div style="color: var(--text-secondary); text-transform: uppercase; font-size: 9px;">Elevación</div>
								<div style="color: var(--text-primary); font-weight: 600;">${vData.elevation ? vData.elevation.toLocaleString() : 'N/D'} m</div>
							</div>
						</div>
						<div style="font-size: 11px; color: ${color}; font-weight: 700; margin-bottom: 6px;">
							Alerta: ${alert.alert_level || alert.color_code || 'N/A'}
						</div>
						${alert.synopsis ? `<div style="font-size: 11px; color: var(--text-primary); line-height: 1.4; max-height: 80px; overflow-y: auto;">${alert.synopsis}</div>` : ''}
					</div>
				`;
				marker.bindPopup(popupHtml, { className: 'volcano-popup', closeButton: true, maxWidth: 280 });
				marker.bindTooltip(`<b>${alert.name}</b><br>Alerta: ${alert.alert_level || alert.color_code || 'N/A'}`, { sticky: true, className: 'volcano-alert-tooltip' });
				marker.on('click', () => {
					const selected = feature?.properties ? { ...feature.properties, ...alert, name: alert.name } as VolcanoProperties : alert as unknown as VolcanoProperties;
					onSelectVolcano?.(selected, [alert.lat, alert.lon]);
				});
				marker.addTo(volcanoAlertMarkersLayer);
			});
			volcanoAlertMarkersLayer.addTo(map);
		}, 100);
	}

	$effect(() => {
		if (!map || !L) return;
		if (showVolcanoes) {
			loadVolcanoesLayer();
			renderVolcanoAlertMarkers();
		} else {
			if (volcanoesLayer) {
				map.removeLayer(volcanoesLayer);
				volcanoesLayer = null;
			}
			if (volcanoAlertMarkersLayer) {
				map.removeLayer(volcanoAlertMarkersLayer);
				volcanoAlertMarkersLayer = null;
			}
		}
	});
</script>

<style>
	:global(.volcano-alert-tooltip) {
		max-width: 320px;
		line-height: 1.4;
	}

	:global(.volcano-popup .leaflet-popup-content-wrapper) {
		background: var(--bg-overlay);
		backdrop-filter: blur(8px);
		border: 1px solid var(--border-color);
		border-radius: 8px;
		color: var(--text-primary);
	}

	:global(.volcano-popup .leaflet-popup-tip) {
		background: var(--bg-overlay);
		border: 1px solid var(--border-color);
	}

	:global(.volcano-popup .leaflet-popup-content) {
		margin: 12px 14px;
	}

	:global(.volcano-popup .leaflet-popup-close-button) {
		color: var(--text-secondary);
		font-size: 18px;
		padding: 4px;
	}

	:global(.volcano-popup .leaflet-popup-close-button:hover) {
		color: var(--text-primary);
	}

	:global(.volcano-triangle),
	:global(.volcano-alert-marker) {
		background: none !important;
		border: none !important;
	}

	:global(.volcano-pulse-marker svg) {
		animation: volcanoPulse 2.5s ease-in-out infinite;
		transform-origin: center bottom;
	}

	:global(.volcano-selected svg) {
		filter: drop-shadow(0 0 5px #00e5ff);
	}

	@keyframes volcanoPulse {
		0% { transform: scale(1); opacity: 1; }
		50% { transform: scale(1.12); opacity: 0.85; }
		100% { transform: scale(1); opacity: 1; }
	}
</style>
