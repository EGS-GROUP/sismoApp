<script lang="ts">
	import { sismoState } from '$lib/sismoStore.svelte';
	import { parseDepthKm, calculateEpicentralRadius, calculateDamageRadius } from '$lib/utils/seismicWaves';

	interface Props {
		map: any;
		L: any;
	}

	let { map, L }: Props = $props();

	let simLayerGroup: any = null;
	let pCircle: any = null;
	let sCircle: any = null;
	let damageCircle: any = null;
	let epicenterMarker: any = null;
	let lastSimulatedId: string | null = null;

	$effect(() => {
		const sim = sismoState.waveSimulation;

		if (!map || !L) return;

		if (!sim.active || !sim.earthquake) {
			cleanLayer();
			lastSimulatedId = null;
			return;
		}

		const eq = sim.earthquake;
		const coords = eq.coordinates;
		if (!coords || coords.length < 2) return;

		const [lat, lon] = coords;
		const depth = parseDepthKm(eq.depth);
		const mag = eq.magnitude || 5.0;

		// Si es la primera vez que se activa para este sismo, centramos el mapa
		if (lastSimulatedId !== eq.id) {
			lastSimulatedId = eq.id;
			initSimulationLayers(lat, lon, mag, depth);
			try {
				map.flyTo([lat, lon], 7, { duration: 1.0 });
			} catch (_) {}
		}

		// Actualizar radios con el tiempo actual
		updateWaveRadii(lat, lon, depth, sim.pSpeed, sim.sSpeed, sim.timeSec);
	});

	function initSimulationLayers(lat: number, lon: number, mag: number, depth: number) {
		cleanLayer();
		simLayerGroup = L.layerGroup().addTo(map);

		// 1. Epicenter Custom DivIcon con efecto de ondas pulsantes
		const epicenterIcon = L.divIcon({
			className: 'seismic-epicenter-marker',
			html: `
				<div class="epicenter-pulse-container">
					<div class="pulse-ring ring-p"></div>
					<div class="pulse-ring ring-s"></div>
					<div class="epicenter-dot"></div>
				</div>
			`,
			iconSize: [40, 40],
			iconAnchor: [20, 20]
		});

		epicenterMarker = L.marker([lat, lon], { icon: epicenterIcon, zIndexOffset: 2000 })
			.bindTooltip(`<b>Epicentro</b>: M${mag.toFixed(1)}<br>Profundidad: ${depth} km`, {
				direction: 'top',
				className: 'seismic-hud-tooltip'
			})
			.addTo(simLayerGroup);

		// 2. Zona de Daño Estimado (Mercalli VI+)
		const damageRadiusKm = calculateDamageRadius(mag, depth);
		if (damageRadiusKm > 0) {
			damageCircle = L.circle([lat, lon], {
				radius: damageRadiusKm * 1000,
				color: '#f59e0b',
				weight: 1.5,
				dashArray: '5, 8',
				fillColor: '#f59e0b',
				fillOpacity: 0.08,
				interactive: true
			})
				.bindTooltip(`Zona Crítica Mercalli VI+ (${damageRadiusKm} km)`, {
					direction: 'bottom',
					className: 'seismic-hud-tooltip'
				})
				.addTo(simLayerGroup);
		}

		// 3. Onda P (Cyan / Primaria)
		pCircle = L.circle([lat, lon], {
			radius: 10,
			color: '#00f0ff',
			weight: 2.5,
			dashArray: '8, 6',
			fillColor: '#00f0ff',
			fillOpacity: 0.08,
			interactive: true
		})
			.bindTooltip('Frente de Onda P (Compresional)', { direction: 'top', className: 'seismic-hud-tooltip' })
			.addTo(simLayerGroup);

		// 4. Onda S (Rojo / Destructiva)
		sCircle = L.circle([lat, lon], {
			radius: 5,
			color: '#ef4444',
			weight: 3.5,
			fillColor: '#ef4444',
			fillOpacity: 0.16,
			interactive: true
		})
			.bindTooltip('Frente de Onda S (Cizalla / Destructiva)', { direction: 'top', className: 'seismic-hud-tooltip' })
			.addTo(simLayerGroup);
	}

	function updateWaveRadii(lat: number, lon: number, depth: number, pSpeed: number, sSpeed: number, timeSec: number) {
		if (!simLayerGroup) return;

		const rPKm = calculateEpicentralRadius(pSpeed, timeSec, depth);
		const rSKm = calculateEpicentralRadius(sSpeed, timeSec, depth);

		if (pCircle) {
			if (rPKm > 0) {
				pCircle.setRadius(rPKm * 1000);
				pCircle.setStyle({ opacity: 1, fillOpacity: 0.08 });
				pCircle.setTooltipContent(`<b>Onda P:</b> ${rPKm.toFixed(0)} km (v=${pSpeed} km/s)`);
			} else {
				pCircle.setStyle({ opacity: 0, fillOpacity: 0 });
			}
		}

		if (sCircle) {
			if (rSKm > 0) {
				sCircle.setRadius(rSKm * 1000);
				sCircle.setStyle({ opacity: 1, fillOpacity: 0.16 });
				sCircle.setTooltipContent(`<b>Onda S:</b> ${rSKm.toFixed(0)} km (v=${sSpeed} km/s)`);
			} else {
				sCircle.setStyle({ opacity: 0, fillOpacity: 0 });
			}
		}
	}

	function cleanLayer() {
		if (simLayerGroup && map) {
			map.removeLayer(simLayerGroup);
			simLayerGroup = null;
			pCircle = null;
			sCircle = null;
			damageCircle = null;
			epicenterMarker = null;
		}
	}
</script>

<style>
	:global(.seismic-epicenter-marker) {
		background: transparent;
		border: none;
	}

	:global(.epicenter-pulse-container) {
		position: relative;
		width: 40px;
		height: 40px;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	:global(.epicenter-dot) {
		width: 14px;
		height: 14px;
		background: #ffffff;
		border: 3px solid #ef4444;
		border-radius: 50%;
		box-shadow: 0 0 16px #ef4444, inset 0 0 6px #ef4444;
		z-index: 5;
	}

	:global(.pulse-ring) {
		position: absolute;
		border-radius: 50%;
		animation: pulse-expand 2s cubic-bezier(0.2, 0.8, 0.2, 1) infinite;
	}

	:global(.ring-p) {
		width: 38px;
		height: 38px;
		border: 2px solid #00f0ff;
		box-shadow: 0 0 10px #00f0ff;
		animation-delay: 0s;
	}

	:global(.ring-s) {
		width: 38px;
		height: 38px;
		border: 2px solid #ef4444;
		box-shadow: 0 0 10px #ef4444;
		animation-delay: 0.7s;
	}

	@keyframes pulse-expand {
		0% {
			transform: scale(0.3);
			opacity: 1;
		}
		100% {
			transform: scale(2.2);
			opacity: 0;
		}
	}

	:global(.seismic-hud-tooltip) {
		background: rgba(10, 15, 29, 0.92) !important;
		backdrop-filter: blur(8px);
		border: 1px solid rgba(0, 240, 255, 0.4) !important;
		color: #f8fafc !important;
		border-radius: 6px !important;
		font-size: 11px !important;
		padding: 4px 8px !important;
		box-shadow: 0 4px 14px rgba(0, 0, 0, 0.5) !important;
	}
</style>
