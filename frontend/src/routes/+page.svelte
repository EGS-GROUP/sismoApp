<script lang="ts">
	import { onMount } from 'svelte';
	import { Moon, Sun, Monitor, ChevronDown, Clock, Filter, List, BarChart3, Settings, SlidersHorizontal, Trophy } from '@lucide/svelte';
	import { findVolcanoAlert, volcanoAlertStyleMap } from '$lib/utils/volcanoUtils';
	import { extractRegion, extractVenezuelaState } from '$lib/utils/geoUtils';
	import { getMagnitudeColor } from '$lib/utils/mapUtils';
	import {
		getEarthquakes,
		getEarthquakesByRange,
		getVolcanoAlerts,
		getNews,
		getPlates,
		getVolcanoes,
		getVapidPublicKey,
		subscribePush,
		createEventSource
	} from '$lib/api';
	import type { Earthquake, VolcanoAlert, TelegramAlert, NewsItem, PushSubscription, GeoJSONCollection, VolcanoProperties } from '$lib/types';
	import { sismoState, timeAgo } from '$lib/sismoStore.svelte';

	import HudHeader from '$lib/components/HudHeader.svelte';
	import SidebarRail from '$lib/components/SidebarRail.svelte';
	import RightPanel from '$lib/components/RightPanel.svelte';
	import BottomPanel from '$lib/components/BottomPanel.svelte';
	import MobileNav from '$lib/components/MobileNav.svelte';
	import ManualModal from '$lib/components/ManualModal.svelte';
	import InfoTabs from '$lib/components/InfoTabs.svelte';
	import VolcanoAlertBanner from '$lib/components/VolcanoAlertBanner.svelte';
	import MapControls from '$lib/components/MapControls.svelte';
	import VolcanoLayer from '$lib/components/VolcanoLayer.svelte';
	import DamageLayer from '$lib/components/DamageLayer.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import RankingWidget from '$lib/components/RankingWidget.svelte';

	// @ts-ignore
	import mapboxgl from 'mapbox-gl/dist/mapbox-gl-csp.js';
	import 'mapbox-gl/dist/mapbox-gl.css';
	// @ts-ignore
	import workerUrl from 'mapbox-gl/dist/mapbox-gl-csp-worker.js?url';
	mapboxgl.workerUrl = workerUrl;

	let L: any = $state();
	let map: any = $state();
	let markersLayer: any;
	let heatmapLayer: any;
	let mapReady = $state(false);
	let mapElement: HTMLElement | undefined = $state();
	let cleanupEventSource: (() => void) | undefined;
	let markerRenderFrame: number;
	let alertSound: HTMLAudioElement;

	let LiveTVComponent: any = $state(null);

	let globeElement: HTMLElement | undefined = $state();
	let globeMap: mapboxgl.Map | undefined = $state();
	let globeStyleLoaded = $state(false);
	let globePulseInterval: ReturnType<typeof setInterval> | null = null;
	let globePlatesData: GeoJSONCollection | null = null;
	let platesLayer: any = $state();
	let darkLayer: any = $state();
	let satelliteLayer: any = $state();
	let terrainLayer: any = $state();
	let lightLayer: any = $state();
	let topoLayer: any = $state();
	let streetsLayer: any = $state();
	let satellitePureLayer: any = $state();

	let previousVolcanoAlerts: VolcanoAlert[] = [];
	let initialVolcanoAlertReceived = false;
	let volcanoDataLoading = $state(false);

	let monitorCollapsed = $state({ time: false, filters: false, events: false });

	$effect(() => {
		if (!LiveTVComponent) {
			import('$lib/components/LiveTV.svelte').then(m => {
				LiveTVComponent = m.default;
			});
		}
	});

	const slowSources = ['USP', 'SGC', 'IPGP'];
	let displayedEarthquakes = $derived([...sismoState.earthquakes].filter(eq => {
		if (sismoState.isHistorical) return true;
		const now = Date.now();
		const eqTime = typeof eq.time === 'number' ? eq.time : new Date(eq.time).getTime();
		const isSlowSource = slowSources.some(s => eq.source.includes(s));
		const windowMs = (sismoState.feedFilter !== 'all' && slowSources.includes(sismoState.feedFilter)) || isSlowSource
			? 7 * 24 * 60 * 60 * 1000
			: 24 * 60 * 60 * 1000;
		return now - eqTime <= windowMs;
	}).sort((a, b) => new Date(b.time).getTime() - new Date(a.time).getTime()));

	let sortedAlerts = $derived([...sismoState.alerts].reverse());

	let filteredEarthquakes = $derived(displayedEarthquakes.filter(eq => {
		if (sismoState.feedFilter !== 'all' && !eq.source.includes(sismoState.feedFilter)) return false;
		if (eq.magnitude < sismoState.minMagnitude) return false;
		if (sismoState.rankingMode === 'venezuela' && extractRegion(eq.location) !== 'Venezuela') return false;
		return true;
	}));

	let sismosHoy = $derived(filteredEarthquakes.length);
	let maxMag = $derived(filteredEarthquakes.length > 0 ? Math.max(...filteredEarthquakes.map(e => e.magnitude)).toFixed(1) : '--');
	let crsRecibidos = $derived(filteredEarthquakes.reduce((sum, eq) => sum + (typeof eq.felt === 'number' ? eq.felt : 0), 0));
	let volcanoAlertCount = $derived(sismoState.volcanoAlerts.length);

	const MAX_MAP_EARTHQUAKES = 500;
	let mapEarthquakes = $derived(filteredEarthquakes.slice(0, MAX_MAP_EARTHQUAKES));

	async function subscribeToPush() {
		try {
			if (!('serviceWorker' in navigator)) return;
			const swReg = await navigator.serviceWorker.ready;
			const pubKey = await getVapidPublicKey();
			const padding = '='.repeat((4 - pubKey.length % 4) % 4);
			const base64 = (pubKey + padding).replace(/-/g, '+').replace(/_/g, '/');
			const rawData = window.atob(base64);
			const applicationServerKey = new Uint8Array(rawData.length);
			for (let i = 0; i < rawData.length; ++i) {
				applicationServerKey[i] = rawData.charCodeAt(i);
			}
			const subscription = await swReg.pushManager.subscribe({
				userVisibleOnly: true,
				applicationServerKey: applicationServerKey
			});
			await subscribePush(subscription.toJSON() as PushSubscription);
			console.log('Suscrito a Push remotas exitosamente');
		} catch (e) {
			console.error('Error suscribiendo a push:', e);
		}
	}

	function toggleNotifications() {
		if (!('Notification' in window)) {
			alert('Este navegador no soporta notificaciones de escritorio.');
			return;
		}
		if (Notification.permission === 'granted') {
			sismoState.notificationsEnabled = !sismoState.notificationsEnabled;
			localStorage.setItem('sismo_notifications', sismoState.notificationsEnabled.toString());
			if (sismoState.notificationsEnabled) {
				subscribeToPush();
			}
		} else if (Notification.permission !== 'denied') {
			Notification.requestPermission().then(permission => {
				if (permission === 'granted') {
					sismoState.notificationsEnabled = true;
					localStorage.setItem('sismo_notifications', 'true');
					subscribeToPush();
					if (alertSound) alertSound.play().catch(e => console.log('Audio autoplay prevented:', e));
					new Notification('SismoMonitor', {
						body: 'Notificaciones activadas. Recibirás alertas incluso si cierras la app.',
						icon: '/favicon/android-chrome-192x192.png'
					});
				}
			});
		} else {
			alert('Permiso de notificaciones denegado. Debes habilitarlo en los ajustes de tu navegador.');
		}
	}

	function applyTheme(theme: 'dark' | 'light' | 'system') {
		let effective = theme;
		if (theme === 'system') {
			effective = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
		}
		document.documentElement.setAttribute('data-theme', effective);
		sismoState.mapTheme = effective === 'light' ? 'satellite' : 'dark';
	}

	function setTheme(theme: 'dark' | 'light' | 'system') {
		sismoState.theme = theme;
		localStorage.setItem('sismo_theme', theme);
		applyTheme(theme);
	}

	function toggleAlertsPanel() {
		sismoState.showAlertsPanel = !sismoState.showAlertsPanel;
		if (sismoState.showAlertsPanel) {
			sismoState.unseenAlerts = [];
		}
	}

	function formatDate(d: Date): string {
		return d.toISOString().split('T')[0];
	}

	async function handleTimeRangeChange(range: 'live' | '24h' | '7d' | 'custom') {
		sismoState.timeRange = range;
		try {
			if (range === 'live') {
				sismoState.isHistorical = false;
				sismoState.selectedDate = '';
				sismoState.earthquakes = await getEarthquakes();
			} else if (range === '24h') {
				sismoState.isHistorical = true;
				const end = new Date();
				const start = new Date(end.getTime() - 24 * 60 * 60 * 1000);
				sismoState.dateStart = formatDate(start);
				sismoState.dateEnd = formatDate(end);
				sismoState.earthquakes = await getEarthquakesByRange(sismoState.dateStart, sismoState.dateEnd);
			} else if (range === '7d') {
				sismoState.isHistorical = true;
				const end = new Date();
				const start = new Date(end.getTime() - 7 * 24 * 60 * 60 * 1000);
				sismoState.dateStart = formatDate(start);
				sismoState.dateEnd = formatDate(end);
				sismoState.earthquakes = await getEarthquakesByRange(sismoState.dateStart, sismoState.dateEnd);
			} else if (range === 'custom') {
				sismoState.isHistorical = true;
				if (sismoState.dateStart && sismoState.dateEnd) {
					sismoState.earthquakes = await getEarthquakesByRange(sismoState.dateStart, sismoState.dateEnd);
				}
			}
			sismoState.lastUpdate = new Date().toLocaleTimeString('es-ES', { hour: 'numeric', minute: '2-digit', second: '2-digit', hour12: true });
		} catch (err) {
			console.log('Error cambiando rango temporal:', err);
		}
	}

	async function handleRefresh() {
		try {
			if (sismoState.timeRange === '24h' || sismoState.timeRange === '7d' || (sismoState.timeRange === 'custom' && sismoState.dateStart && sismoState.dateEnd)) {
				await handleTimeRangeChange(sismoState.timeRange);
			} else {
				sismoState.earthquakes = await getEarthquakes();
				sismoState.lastUpdate = new Date().toLocaleTimeString('es-ES', { hour: 'numeric', minute: '2-digit', second: '2-digit', hour12: true });
			}
		} catch (err) {
			console.log('Error refrescando sismos:', err);
		}
	}

	function handleSelectEarthquake(eq: Earthquake) {
		sismoState.selectedEarthquake = eq;
		sismoState.selectedVolcano = null;
		sismoState.activeTab = 'details';
		if (map) map.flyTo([eq.coordinates[0], eq.coordinates[1]], 8);
	}

	function handleSelectVolcanoFromLayer(volcano: VolcanoProperties, coords: [number, number]) {
		if (sismoState.selectedVolcano?.name === volcano.name && sismoState.activeTab === 'details') return;
		sismoState.selectedVolcano = volcano;
		sismoState.selectedEarthquake = null;
		sismoState.activeTab = 'details';
		sismoState.showVolcanoes = true;
		if (map && coords) { map.flyTo(coords, 10); }
	}

	async function loadVolcanoData() {
		if (sismoState.volcanoData || volcanoDataLoading) return;
		volcanoDataLoading = true;
		try {
			const data = await getVolcanoes();
			sismoState.volcanoData = data;
			console.log(`Datos de volcanes precargados: ${data.features.length} volcanes`);
		} catch (err) {
			console.error('Error cargando volcanes:', err);
		} finally {
			volcanoDataLoading = false;
		}
	}

	function handleSelectAlert(alert: TelegramAlert) {
		sismoState.modalCanClose = false;
		setTimeout(() => sismoState.modalCanClose = true, 300);
		const rescueInfo = alert.type === 'missing' ? { text: 'Personas Extraviadas', class: 'tag-missing' }
			: alert.type === 'supplies' ? { text: 'Insumos / Ayuda', class: 'tag-supplies' }
			: { text: 'Reporte General', class: 'tag-general' };
		sismoState.selectedModalData = {
			title: rescueInfo.text,
			image: alert.image_url,
			bodyHTML: `<p style="font-size: 14px; line-height: 1.5;">${(alert.fullText || alert.text || '').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/\n/g, '<br>')}</p><br><p style="color: var(--text-secondary); font-size: 11px;">Extraído de: ${alert.group || alert.channel_id || ''}</p>`,
			sourceLink: alert.link || `https://t.me/${(alert.channel_id || '').replace('@', '')}`,
			sourceText: 'Ver Mensaje Original'
		};
	}

	function handleSelectNews(news: NewsItem) {
		sismoState.modalCanClose = false;
		setTimeout(() => sismoState.modalCanClose = true, 300);
		sismoState.selectedModalData = {
			title: news.title,
			bodyHTML: `<p style="color:var(--text-secondary); margin-bottom:12px; font-size: 12px;">Publicado: ${new Date(news.pubDate).toLocaleString()}</p><p style="font-size: 14px;">Haz clic en "Ver Fuente Original" para leer el artículo completo.</p>`,
			sourceLink: news.link,
			sourceText: 'Ver Fuente Original'
		};
	}

	function closeVolcano() {
		sismoState.selectedVolcano = null;
	}

	function setRankingMode(mode: 'global' | 'venezuela') {
		sismoState.rankingMode = mode;
	}

	function calculateRanking() {
		const counts: Record<string, { count: number, maxMag: number }> = {};
		filteredEarthquakes.forEach(eq => {
			let region = '';
			if (sismoState.rankingMode === 'venezuela') {
				if (extractRegion(eq.location) !== 'Venezuela') return;
				region = extractVenezuelaState(eq.location);
				if (region === 'Otras Regiones') return;
			} else {
				region = extractRegion(eq.location);
			}
			if (!counts[region]) {
				counts[region] = { count: 0, maxMag: 0 };
			}
			counts[region].count++;
			if (eq.magnitude && eq.magnitude > counts[region].maxMag) {
				counts[region].maxMag = eq.magnitude;
			}
		});
		sismoState.rankingData = Object.entries(counts).sort((a, b) => b[1].count - a[1].count).slice(0, 10);
	}

	const volcanoAlertPriority: Record<string, number> = {
		'ADVISORY': 1, 'YELLOW': 1, 'WATCH': 2, 'ORANGE': 2, 'WARNING': 3, 'RED': 3
	};

	function getVolcanoAlertChanges(current: any[], previous: any[]): any[] {
		const changes: any[] = [];
		const prevMap = new Map<string, any>();
		for (const a of previous) {
			if (a.name) prevMap.set(a.name.toLowerCase(), a);
		}
		for (const a of current) {
			if (!a.name) continue;
			const key = a.name.toLowerCase();
			const prev = prevMap.get(key);
			if (!prev) {
				changes.push({ type: 'new', alert: a });
			} else {
				const prevLevel = prev.alert_level || prev.color_code || '';
				const currLevel = a.alert_level || a.color_code || '';
				const prevPri = volcanoAlertPriority[prevLevel.toUpperCase()] || 0;
				const currPri = volcanoAlertPriority[currLevel.toUpperCase()] || 0;
				if (currPri > prevPri) {
					changes.push({ type: 'escalated', alert: a, previousLevel: prevLevel });
				}
			}
		}
		return changes;
	}

	// ===== GLOBE MAP =====
	function resetGlobeMap() {
		if (globePulseInterval) {
			clearInterval(globePulseInterval);
			globePulseInterval = null;
		}
		if (globeMap) {
			globeMap.remove();
			globeMap = undefined;
			globeStyleLoaded = false;
			volcanoIconsPromise = null;
		}
	}

	function initGlobeMap() {
		if (globeMap || !globeElement) return;
		const MAPBOX_TOKEN = 'pk.eyJ1IjoidHIzdzAxIiwiYSI6ImNscWZmOGJraTAwY28ycm1nNGRpOGI2azkifQ.btAAsZ-1rY5o7pf1cLNo5g';
		mapboxgl.accessToken = MAPBOX_TOKEN;
		const styleMap: Record<string, string> = {
			dark: 'mapbox://styles/mapbox/dark-v11',
			satellite: 'mapbox://styles/mapbox/satellite-streets-v12',
			terrain: 'mapbox://styles/mapbox/outdoors-v12',
			light: 'mapbox://styles/mapbox/light-v11',
			topo: 'mapbox://styles/mapbox/outdoors-v12',
			streets: 'mapbox://styles/mapbox/streets-v12',
			'satellite-pure': 'mapbox://styles/mapbox/satellite-v9'
		};
		globeMap = new mapboxgl.Map({
			container: globeElement,
			style: styleMap[sismoState.globeMapStyle] || styleMap['dark'],
			center: [-66.0, 7.5],
			zoom: 2.8,
			projection: { name: 'globe' }
		});
		globeMap.on('load', () => {
			globeStyleLoaded = true;
			addGlobeMarkerSource();
			updateGlobeMarkers();
			updateGlobePlates();
			updateGlobeVolcanoes();
		});
		globeMap.on('click', 'eq-circles', (e: any) => {
			const feature = e.features?.[0];
			if (!feature) return;
			const eq = mapEarthquakes.find((x: Earthquake) => x.id === feature.properties.id);
			if (!eq) return;
			sismoState.selectedEarthquake = eq;
			sismoState.selectedVolcano = null;
			sismoState.activeTab = 'details';
			globeMap?.flyTo({ center: [eq.coordinates[1], eq.coordinates[0]], zoom: 8, essential: true });
		});
		globeMap.on('mouseenter', 'eq-circles', () => { if (globeMap) globeMap.getCanvas().style.cursor = 'pointer'; });
		globeMap.on('mouseleave', 'eq-circles', () => { if (globeMap) globeMap.getCanvas().style.cursor = ''; });
		globeMap.on('click', 'volcanoes', (e: any) => {
			const feature = e.features?.[0];
			if (!feature) return;
			const p = feature.properties;
			const alert = findVolcanoAlert(p.name, sismoState.volcanoAlerts);
			sismoState.selectedVolcano = alert ? { ...p, ...alert, name: p.name } as VolcanoProperties : p as VolcanoProperties;
			sismoState.selectedEarthquake = null;
			sismoState.activeTab = 'details';
			const coords = feature.geometry?.coordinates || [0, 0];
			globeMap?.flyTo({ center: [coords[0], coords[1]], zoom: 8, essential: true });
		});
		globeMap.on('mouseenter', 'volcanoes', () => { if (globeMap) globeMap.getCanvas().style.cursor = 'pointer'; });
		globeMap.on('mouseleave', 'volcanoes', () => { if (globeMap) globeMap.getCanvas().style.cursor = ''; });
	}

	function addGlobeMarkerSource() {
		if (!globeMap || !globeStyleLoaded || globeMap.getSource('earthquakes')) return;
		globeMap.addSource('earthquakes', { type: 'geojson', data: { type: 'FeatureCollection', features: [] } });
		// Seismic wave layer (expanding rings for recent events)
		globeMap.addLayer({
			id: 'eq-wave', type: 'circle', source: 'earthquakes',
			filter: ['==', ['get', 'recent'], true],
			paint: {
				'circle-radius': ['max', ['*', ['get', 'magnitude'], 2], 5],
				'circle-color': ['to-color', ['get', 'color']],
				'circle-opacity': 0.4,
				'circle-stroke-width': 1.5,
				'circle-stroke-color': ['to-color', ['get', 'color']],
				'circle-stroke-opacity': 0.5
			}
		});
		// Shadow/halo layer for depth perception
		globeMap.addLayer({
			id: 'eq-shadow', type: 'circle', source: 'earthquakes',
			paint: {
				'circle-radius': ['+', ['max', ['*', ['get', 'magnitude'], 2], 5], 1.5],
				'circle-color': '#000000',
				'circle-opacity': 0.25,
				'circle-blur': 1,
				'circle-stroke-width': 0
			}
		});
		// Main earthquake markers
		globeMap.addLayer({
			id: 'eq-circles', type: 'circle', source: 'earthquakes',
			paint: {
				'circle-radius': ['max', ['*', ['get', 'magnitude'], 2], 5],
				'circle-color': ['to-color', ['get', 'color']],
				'circle-opacity': 0.85,
				'circle-stroke-width': 1,
				'circle-stroke-color': ['to-color', ['get', 'strokeColor']],
				'circle-stroke-opacity': 1
			}
		});
		// Animate wave layer — expanding ring like seismic waves
		let wavePhase = 0;
		globePulseInterval = setInterval(() => {
			wavePhase = (wavePhase + 1) % 50;
			const t = wavePhase / 50;
			const scale = 1 + t * 2;
			const strokeOpacity = Math.max(0, 0.6 * (1 - t));
			if (globeMap?.getLayer('eq-wave')) {
				globeMap.setPaintProperty('eq-wave', 'circle-radius',
					['*', ['max', ['*', ['get', 'magnitude'], 2], 5], scale]);
				globeMap.setPaintProperty('eq-wave', 'circle-opacity', 0);
				globeMap.setPaintProperty('eq-wave', 'circle-stroke-opacity',
					['*', strokeOpacity, ['get', 'intensity']]);
			}
		}, 50);
	}

	function getMagnitudeHex(mag: number) {
		const styles = getComputedStyle(document.documentElement);
		if (mag >= 6.0) return styles.getPropertyValue('--mag-extreme').trim() || '#b91c1c';
		if (mag >= 4.5) return styles.getPropertyValue('--mag-high').trim() || '#ef4444';
		if (mag >= 3.0) return styles.getPropertyValue('--mag-medium').trim() || '#eab308';
		return styles.getPropertyValue('--mag-low').trim() || '#10b981';
	}

	function darkenColor(hex: string, amount = 0.5): string {
		const h = hex.replace('#', '');
		const r = Math.max(0, Math.round(parseInt(h.slice(0, 2), 16) * (1 - amount)));
		const g = Math.max(0, Math.round(parseInt(h.slice(2, 4), 16) * (1 - amount)));
		const b = Math.max(0, Math.round(parseInt(h.slice(4, 6), 16) * (1 - amount)));
		return `#${r.toString(16).padStart(2, '0')}${g.toString(16).padStart(2, '0')}${b.toString(16).padStart(2, '0')}`;
	}

	function updateGlobeMarkers() {
		if (!globeMap || !globeStyleLoaded) return;
		addGlobeMarkerSource();
		const now = Date.now();
		const features = mapEarthquakes.map(eq => {
			const depthNum = parseFloat(eq.depth) || 50;
			const magFactor = Math.max(eq.magnitude / 6, 0.5);
			const depthFactor = Math.max(1 - depthNum / 300, 0.3);
			return {
				type: 'Feature' as const,
				properties: {
					id: eq.id,
					magnitude: eq.magnitude,
					color: getMagnitudeHex(eq.magnitude),
					strokeColor: darkenColor(getMagnitudeHex(eq.magnitude)),
					recent: (now - new Date(eq.time).getTime()) < 3600000,
					intensity: magFactor * depthFactor
				},
				geometry: { type: 'Point' as const, coordinates: [eq.coordinates[1], eq.coordinates[0]] }
			};
		});
		const source = globeMap.getSource('earthquakes') as mapboxgl.GeoJSONSource | undefined;
		if (source) source.setData({ type: 'FeatureCollection', features });
	}

	function updateGlobePlates() {
		if (!globeMap || !globeStyleLoaded) return;
		if (globeMap.getLayer('plates')) {
			globeMap.setLayoutProperty('plates', 'visibility', sismoState.showPlates ? 'visible' : 'none');
			return;
		}
		if (!sismoState.showPlates) return;
		if (globePlatesData) {
			globeMap.addSource('plates', { type: 'geojson', data: globePlatesData as any });
			globeMap.addLayer({ id: 'plates', type: 'line', source: 'plates', paint: { 'line-color': getComputedStyle(document.documentElement).getPropertyValue('--plate-color').trim() || '#ff4444', 'line-width': 1.5 } });
		} else {
			getPlates().then(data => {
				globePlatesData = data;
				if (!globeMap || !globeStyleLoaded) return;
				if (globeMap.getSource('plates')) return;
				globeMap.addSource('plates', { type: 'geojson', data: data as any });
				globeMap.addLayer({ id: 'plates', type: 'line', source: 'plates', paint: { 'line-color': getComputedStyle(document.documentElement).getPropertyValue('--plate-color').trim() || '#ff4444', 'line-width': 1.5 } });
			});
		}
	}

	const VOLCANO_ICON_COLORS: Record<string, string> = {
		'volcano-red': '#ff0000', 'volcano-orange': '#ff8c00', 'volcano-yellow': '#ffd700',
		'volcano-green': '#00ff00', 'volcano-default': '#ff6347'
	};
	let volcanoIconsPromise: Promise<void> | null = null;

	function getVolcanoSvg(color: string): string {
		return `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><polygon points="12,2 22,22 2,22" fill="${color}" stroke="none"/></svg>`;
	}

	function ensureVolcanoIcons(map: mapboxgl.Map): Promise<void> {
		if (map.hasImage('volcano-red')) return Promise.resolve();
		if (volcanoIconsPromise) return volcanoIconsPromise;
		volcanoIconsPromise = Promise.all(
			Object.entries(VOLCANO_ICON_COLORS).map(([key, color]) => new Promise<void>((resolve) => {
				const svg = getVolcanoSvg(color);
				const img = new Image();
				img.onload = () => { map.addImage(key, img); resolve(); };
				img.onerror = () => { console.error('Error cargando icono de volcán', key); resolve(); };
				img.src = 'data:image/svg+xml;base64,' + btoa(svg);
			}))
		).then(() => {});
		return volcanoIconsPromise;
	}

	function getGlobeVolcanoFeatures() {
		if (!sismoState.volcanoData) return [];
		const iconMap: Record<string, string> = {
			'RED': 'volcano-red', 'WARNING': 'volcano-red', 'ORANGE': 'volcano-orange',
			'WATCH': 'volcano-orange', 'YELLOW': 'volcano-yellow', 'ADVISORY': 'volcano-yellow',
			'GREEN': 'volcano-green', 'NORMAL': 'volcano-green'
		};
		return sismoState.volcanoData.features.map((f: any) => {
			const p = f.properties || {};
			const alert = findVolcanoAlert(p.name || '', sismoState.volcanoAlerts);
			const key = (alert?.alert_level || alert?.color_code || '').toUpperCase();
			const style = volcanoAlertStyleMap[key];
			const radius = style ? style.radius : 10;
			return {
				type: 'Feature' as const,
				properties: { ...p, name: p.name, icon: iconMap[key] || 'volcano-default', iconSize: radius / 12 },
				geometry: f.geometry
			};
		});
	}

	async function updateGlobeVolcanoes() {
		if (!globeMap || !globeStyleLoaded) return;
		if (globeMap.getLayer('volcanoes')) {
			globeMap.setLayoutProperty('volcanoes', 'visibility', sismoState.showVolcanoes ? 'visible' : 'none');
			if (sismoState.showVolcanoes && sismoState.volcanoData) {
				const source = globeMap.getSource('volcanoes') as mapboxgl.GeoJSONSource | undefined;
				source?.setData({ type: 'FeatureCollection', features: getGlobeVolcanoFeatures() });
			}
			return;
		}
		if (!sismoState.showVolcanoes) return;
		if (!sismoState.volcanoData) {
			if (!volcanoDataLoading) loadVolcanoData();
			return;
		}
		await ensureVolcanoIcons(globeMap);
		if (!globeMap || !globeStyleLoaded || globeMap.getLayer('volcanoes')) return;
		if (!sismoState.showVolcanoes) return;
		globeMap.addSource('volcanoes', { type: 'geojson', data: { type: 'FeatureCollection', features: getGlobeVolcanoFeatures() } });
		globeMap.addLayer({
			id: 'volcanoes', type: 'symbol', source: 'volcanoes',
			layout: { 'icon-image': ['get', 'icon'], 'icon-size': ['get', 'iconSize'], 'icon-allow-overlap': true, 'icon-ignore-placement': true }
		});
	}

	// ===== EFFECTS =====
	$effect(() => {
		if (sismoState.globeMode && globeElement && !globeMap) initGlobeMap();
		if (!sismoState.globeMode && globeMap) resetGlobeMap();
		requestAnimationFrame(() => {
			if (!sismoState.globeMode && map) map.invalidateSize();
			if (sismoState.globeMode && globeMap) globeMap.resize();
		});
	});

	// Recalculate map size when panels toggle
	$effect(() => {
		// Track panel states to trigger this effect
		void sismoState.leftPanelOpen;
		void sismoState.rightPanelOpen;
		void sismoState.bottomPanelOpen;
		// Call invalidateSize at multiple points during the 350ms CSS transition
		const timeouts = [50, 200, 400];
		timeouts.forEach(ms => {
			setTimeout(() => {
				if (!sismoState.globeMode && map) map.invalidateSize();
				if (sismoState.globeMode && globeMap) globeMap.resize();
			}, ms);
		});
	});

	$effect(() => {
		if (sismoState.globeMode && globeMap && globeStyleLoaded && sismoState.earthquakes.length) updateGlobeMarkers();
	});

	$effect(() => {
		if (sismoState.globeMode && globeMap && globeStyleLoaded) updateGlobePlates();
	});

	// Heatmap layer
	$effect(() => {
		if (!mapReady || !L || !map) return;
		// Track dependencies
		void sismoState.showHeatmap;
		const eqs = mapEarthquakes;

		if (heatmapLayer) {
			map.removeLayer(heatmapLayer);
			heatmapLayer = null;
		}
		if (sismoState.showHeatmap && eqs.length > 0) {
			const heatPoints = eqs.map(eq => [
				eq.coordinates[0],
				eq.coordinates[1],
				Math.max(eq.magnitude, 1)
			]);
			heatmapLayer = L.heatLayer(heatPoints, {
				radius: 35,
				blur: 25,
				maxZoom: 10,
				max: 8,
				gradient: { 0.2: '#22c55e', 0.4: '#eab308', 0.6: '#ef4444', 0.8: '#b91c1c', 1.0: '#7f1d1d' }
			});
			heatmapLayer.addTo(map);
		}
	});

	$effect(() => {
		if (sismoState.globeMode && globeMap && sismoState.mapTheme !== sismoState.globeMapStyle) {
			sismoState.globeMapStyle = sismoState.mapTheme;
			resetGlobeMap();
		}
	});

	$effect(() => {
		if (sismoState.globeMode && globeMap && globeStyleLoaded && sismoState.selectedEarthquake) {
			globeMap.flyTo({ center: [sismoState.selectedEarthquake.coordinates[1], sismoState.selectedEarthquake.coordinates[0]], zoom: 8, essential: true });
		}
	});

	$effect(() => {
		if ((sismoState.chartMode === 'volcanes' || sismoState.showVolcanoes) && !sismoState.volcanoData) {
			loadVolcanoData();
		}
	});

	$effect(() => {
		if (filteredEarthquakes && sismoState.rankingMode) {
			calculateRanking();
		}
	});

	$effect(() => {
		if (mapReady && L && markersLayer) {
			const currentEarthquakes = mapEarthquakes;
			const currentSelected = sismoState.selectedEarthquake;
			const currentMap = map;
			cancelAnimationFrame(markerRenderFrame);
			markerRenderFrame = requestAnimationFrame(() => {
				markersLayer.clearLayers();
				const now = Date.now();
				currentEarthquakes.forEach(eq => {
					const isSelected = currentSelected && currentSelected.id === eq.id;
					const color = getMagnitudeHex(eq.magnitude);
					const ageMs = now - new Date(eq.time).getTime();
					const isRecent = ageMs < 3600000;

					if (isRecent && !isSelected) {
						const depthNum = parseFloat(eq.depth) || 50;
						const magFactor = Math.max(eq.magnitude / 6, 0.5);
						const depthFactor = Math.max(1 - depthNum / 300, 0.3);
						const intensity = magFactor * depthFactor;
						const baseRadius = Math.max(eq.magnitude * 2, 5);
						const maxRings = Math.min(Math.ceil(eq.magnitude / 2), 4);
						for (let r = 0; r < maxRings; r++) {
							const ringOpacity = 0.5 * intensity * (1 - r / maxRings);
							const ring = L.circle([eq.coordinates[0], eq.coordinates[1]], {
								radius: (baseRadius + 4 + r * 6) * 1000,
								fillColor: color,
								color: color,
								weight: 1.5,
								opacity: ringOpacity,
								fillOpacity: 0,
								className: `eq-wave eq-wave-${r}`
							});
							ring.addTo(markersLayer);
						}
					}

					const marker = L.circleMarker([eq.coordinates[0], eq.coordinates[1]], {
						radius: isSelected ? Math.max(eq.magnitude * 2, 5) + 4 : Math.max(eq.magnitude * 2, 5),
						fillColor: color,
						color: isSelected ? (getComputedStyle(document.documentElement).getPropertyValue('--accent').trim() || '#00e5ff') : darkenColor(color),
						weight: isSelected ? 3 : 1,
						opacity: 1,
						fillOpacity: isSelected ? 1 : 0.85
					});
					marker.on('click', () => {
						sismoState.selectedEarthquake = eq;
						sismoState.selectedVolcano = null;
						sismoState.activeTab = 'details';
						currentMap.flyTo([eq.coordinates[0], eq.coordinates[1]], 8);
					});
					marker.addTo(markersLayer);
				});
			});
		}
	});

	$effect(() => {
		if (mapReady && L && map && platesLayer) {
			if (sismoState.showPlates) map.addLayer(platesLayer);
			else map.removeLayer(platesLayer);
		}
	});

	$effect(() => {
		if (mapReady && map && darkLayer && satelliteLayer && terrainLayer && lightLayer && topoLayer && streetsLayer && satellitePureLayer) {
			const layers = [darkLayer, satelliteLayer, terrainLayer, lightLayer, topoLayer, streetsLayer, satellitePureLayer];
			layers.forEach(l => map.removeLayer(l));
			const layerMap: Record<string, any> = {
				dark: darkLayer, satellite: satelliteLayer, terrain: terrainLayer,
				light: lightLayer, topo: topoLayer, streets: streetsLayer, 'satellite-pure': satellitePureLayer
			};
			const active = layerMap[sismoState.mapTheme] || darkLayer;
			map.addLayer(active);
		}
	});

	$effect(() => {
		// Cargar noticias al iniciar si no hay (ahora se muestran en BottomPanel)
		if (sismoState.newsItems.length === 0) {
			getNews('sismo').then(data => sismoState.newsItems = data).catch(() => {});
		}
		if (sismoState.activeTab === 'ranking') {
			calculateRanking();
		}
	});

	// ===== LAYOUT HELPERS =====
	let appClasses = $derived([
		'app',
		!sismoState.leftPanelOpen ? 'left-closed' : '',
		!sismoState.rightPanelOpen ? 'right-closed' : '',
		!sismoState.bottomPanelOpen ? 'bottom-closed' : '',
		sismoState.innerWidth <= 768 ? `mobile-${sismoState.mobileTab}` : ''
	].filter(Boolean).join(' '));

	function handleMobileTabChange(tab: string) {
		sismoState.mobileTab = tab as any;
		if (tab === 'map') {
			setTimeout(() => { if (map) map.invalidateSize(); }, 100);
		}
	}

	onMount(() => {
		alertSound = new Audio('/alerta.mp3');

		try {
			const savedTheme = localStorage.getItem('sismo_theme') as 'dark' | 'light' | 'system' | null;
			if (savedTheme) {
				sismoState.theme = savedTheme;
			}
			applyTheme(sismoState.theme);
		} catch (e) {
			console.warn('Error reading theme from localStorage', e);
		}

		try {
			const savedNotifs = localStorage.getItem('sismo_notifications');
			if (savedNotifs === 'true') {
				sismoState.notificationsEnabled = true;
				subscribeToPush();
			}
		} catch (e) {
			console.warn('Error reading localStorage for notifications', e);
		}

		(async () => {
			try {
				const [eqs, alerts] = await Promise.all([
					getEarthquakes(),
					getVolcanoAlerts()
				]);
				sismoState.earthquakes = eqs;
				sismoState.lastUpdate = new Date().toLocaleTimeString();
				sismoState.volcanoAlerts = alerts;
			} catch (err) {
				console.log('Error fetching initial data:', err);
			}

			await new Promise(resolve => requestAnimationFrame(resolve));
			L = (await import('leaflet')).default;
			map = L.map(mapElement, { zoomControl: false }).setView([7.5, -66.0], 5);
			L.control.zoom({ position: 'topright' }).addTo(map);

			const MAPBOX_TOKEN = 'pk.eyJ1IjoidHIzdzAxIiwiYSI6ImNscWZmOGJraTAwY28ycm1nNGRpOGI2azkifQ.btAAsZ-1rY5o7pf1cLNo5g';

			darkLayer = L.tileLayer('https://api.mapbox.com/styles/v1/mapbox/dark-v11/tiles/{z}/{x}/{y}?access_token=' + MAPBOX_TOKEN, {
				attribution: '&copy; Mapbox &copy; OpenStreetMap',
				maxZoom: 20
			});

			satelliteLayer = L.tileLayer('https://api.mapbox.com/styles/v1/mapbox/satellite-streets-v12/tiles/{z}/{x}/{y}?access_token=' + MAPBOX_TOKEN, {
				attribution: '&copy; Mapbox &copy; OpenStreetMap', maxZoom: 20
			});
			terrainLayer = L.tileLayer('https://api.mapbox.com/styles/v1/mapbox/outdoors-v12/tiles/{z}/{x}/{y}?access_token=' + MAPBOX_TOKEN, {
				attribution: '&copy; Mapbox &copy; OpenStreetMap', maxZoom: 20
			});
			lightLayer = L.tileLayer('https://api.mapbox.com/styles/v1/mapbox/light-v11/tiles/{z}/{x}/{y}?access_token=' + MAPBOX_TOKEN, {
				attribution: '&copy; Mapbox &copy; OpenStreetMap',
				maxZoom: 20
			});
			topoLayer = L.tileLayer('https://server.arcgisonline.com/ArcGIS/rest/services/World_Topo_Map/MapServer/tile/{z}/{y}/{x}', {
				attribution: '&copy; Esri, USGS, NOAA', maxZoom: 19
			});
			streetsLayer = L.tileLayer('https://api.mapbox.com/styles/v1/mapbox/streets-v12/tiles/{z}/{x}/{y}?access_token=' + MAPBOX_TOKEN, {
				attribution: '&copy; Mapbox &copy; OpenStreetMap', maxZoom: 20
			});
			satellitePureLayer = L.tileLayer('https://api.mapbox.com/styles/v1/mapbox/satellite-v9/tiles/{z}/{x}/{y}?access_token=' + MAPBOX_TOKEN, {
				attribution: '&copy; Mapbox &copy; OpenStreetMap', maxZoom: 20
			});

			markersLayer = L.layerGroup().addTo(map);
			mapReady = true;

			requestAnimationFrame(() => map?.invalidateSize());

			const eventSource = createEventSource();

			eventSource.addEventListener('earthquakes_update', (e) => {
				if (sismoState.isHistorical) return;
				const data = JSON.parse(e.data);
				const newEarthquakes = data.earthquakes || [];
				if (sismoState.earthquakes.length > 0 && newEarthquakes.length > 0) {
					const oldLatest = sismoState.earthquakes[0];
					const newLatest = newEarthquakes[0];
					if (newLatest.id !== oldLatest.id) {
						const newOnes = newEarthquakes.filter((ne: Earthquake) => !sismoState.earthquakes.some((oe: Earthquake) => oe.id === ne.id));
						sismoState.unseenAlerts = [...newOnes, ...sismoState.unseenAlerts].slice(0, 50);
						if (sismoState.notificationsEnabled) {
							if (alertSound) alertSound.play().catch(e => console.log('Audio play error:', e));
							if ('Notification' in window && Notification.permission === 'granted') {
								new Notification('¡Nuevo Sismo Detectado!', {
									body: `Mag ${newLatest.magnitude} en ${newLatest.location}`,
									icon: '/favicon/android-chrome-192x192.png',
									// @ts-ignore
									vibrate: [200, 100, 200]
								});
							}
						}
					}
				}
				sismoState.earthquakes = newEarthquakes;
				sismoState.lastUpdate = new Date().toLocaleTimeString();
			});

			eventSource.addEventListener('telegram_update', (e) => {
				const data = JSON.parse(e.data);
				sismoState.alerts = data.alerts || [];
			});

			eventSource.addEventListener('stats_update', (e) => {
				const data = JSON.parse(e.data);
				sismoState.activeUsers = data.active_users;
			});

			eventSource.addEventListener('volcano_alerts_update', (e) => {
				const data = JSON.parse(e.data);
				const newAlerts = data.volcano_alerts || [];
				if (initialVolcanoAlertReceived) {
					const changes = getVolcanoAlertChanges(newAlerts, previousVolcanoAlerts);
					for (const change of changes) {
						const a = change.alert;
						const level = a.alert_level || a.color_code || 'ALERTA';
						const title = change.type === 'new' ? `Nueva Alerta Volcánica: ${level}` : `Alerta Volcánica Escalada: ${level}`;
						const body = `${a.name}\n${a.synopsis ? a.synopsis.substring(0, 100) : 'Actividad volcánica detectada'}`;
						if (alertSound) alertSound.play().catch(err => console.log('Audio play error:', err));
						if ('Notification' in window && Notification.permission === 'granted' && sismoState.notificationsEnabled) {
							new Notification(title, { body, icon: '/favicon/android-chrome-192x192.png' });
						}
						sismoState.activeVolcanoAlertBanner = {
							name: a.name, level, synopsis: a.synopsis || '', type: change.type, previousLevel: change.previousLevel || null
						};
					}
				} else {
					initialVolcanoAlertReceived = true;
				}
				previousVolcanoAlerts = newAlerts;
				sismoState.volcanoAlerts = newAlerts;
			});

			const loadPlates = () => getPlates()
				.then(data => {
					const plateColor = getComputedStyle(document.documentElement).getPropertyValue('--plate-color').trim() || '#ff4444';
					platesLayer = L.geoJSON(data, {
						style: { color: plateColor, weight: 1, opacity: 0.5, dashArray: '5, 5' }
					});
				})
				.catch(() => console.log('Plates not found, ignoring.'));
			if ('requestIdleCallback' in window) {
				window.requestIdleCallback(loadPlates, { timeout: 2000 });
			} else {
				setTimeout(loadPlates, 500);
			}

			cleanupEventSource = () => eventSource.close();
		})();

		return () => {
			cancelAnimationFrame(markerRenderFrame);
			cleanupEventSource?.();
			if (map) map.remove();
		};
	});
</script>

<svelte:window bind:innerWidth={sismoState.innerWidth} />

<div class={appClasses}>
	<HudHeader
		sismosHoy={sismosHoy}
		maxMag={maxMag}
		crsRecibidos={crsRecibidos}
		activeUsers={sismoState.activeUsers}
		lastUpdate={sismoState.lastUpdate}
		isHistorical={sismoState.isHistorical}
		globeMode={sismoState.globeMode}
		unseenAlerts={sismoState.unseenAlerts}
		showAlertsPanel={sismoState.showAlertsPanel}
		onRefresh={handleRefresh}
		onToggleAlertsPanel={toggleAlertsPanel}
		onToggleGlobe={() => sismoState.globeMode = !sismoState.globeMode}
		onSelectEarthquake={handleSelectEarthquake}
	/>

	<!-- LEFT PANEL: Sidebar Rail + Feed -->
	<div class="panel-left glass" style:display={sismoState.innerWidth <= 768 ? (sismoState.mobileTab === 'feed' ? 'flex' : 'none') : undefined}>
		<SidebarRail
			sidebarPage={sismoState.sidebarPage}
			leftPanelOpen={sismoState.leftPanelOpen}
			onPageChange={(p) => sismoState.sidebarPage = p}
			onToggleLeft={() => sismoState.leftPanelOpen = !sismoState.leftPanelOpen}
			onOpenInfo={() => sismoState.showManualModal = true}
		/>
		<div class="panel-body">
			{#if sismoState.sidebarPage === 'monitor'}
				<div class="panel-page active">
					<section class="monitor-controls collapsible {monitorCollapsed.time ? 'collapsed' : ''}">
						<button class="collapsible-header" aria-expanded={!monitorCollapsed.time} onclick={() => monitorCollapsed.time = !monitorCollapsed.time}>
							<h3><Clock size={14} /> Rango temporal</h3>
							<span class="chevron"><ChevronDown size={16} /></span>
						</button>
						<div class="collapsible-content">
							<div class="time-range-buttons">
								<button
									class="time-btn {sismoState.timeRange === 'live' ? 'active' : ''}"
									onclick={() => handleTimeRangeChange('live')}
								>En vivo</button>
								<button
									class="time-btn {sismoState.timeRange === '24h' ? 'active' : ''}"
									onclick={() => handleTimeRangeChange('24h')}
								>24h</button>
								<button
									class="time-btn {sismoState.timeRange === '7d' ? 'active' : ''}"
									onclick={() => handleTimeRangeChange('7d')}
								>7d</button>
								<button
									class="time-btn {sismoState.timeRange === 'custom' ? 'active' : ''}"
									onclick={() => sismoState.timeRange = 'custom'}
								>Personalizado</button>
							</div>
							{#if sismoState.timeRange === 'custom'}
								<div class="custom-date-range">
									<label>
										<span>Desde</span>
										<input type="date" bind:value={sismoState.dateStart} />
									</label>
									<label>
										<span>Hasta</span>
										<input type="date" bind:value={sismoState.dateEnd} />
									</label>
									<button
										class="time-btn apply-btn"
										onclick={() => handleTimeRangeChange('custom')}
										disabled={!sismoState.dateStart || !sismoState.dateEnd}
									>Aplicar rango</button>
								</div>
							{/if}
						</div>
					</section>
					<section class="monitor-controls collapsible {monitorCollapsed.filters ? 'collapsed' : ''}">
						<button class="collapsible-header" aria-expanded={!monitorCollapsed.filters} onclick={() => monitorCollapsed.filters = !monitorCollapsed.filters}>
							<h3><Filter size={14} /> Filtro de eventos</h3>
							<span class="chevron"><ChevronDown size={16} /></span>
						</button>
						<div class="collapsible-content">
							<label class="agency-select-label">
								<span>Agencia</span>
								<select class="agency-select" value={sismoState.feedFilter} onchange={(e) => sismoState.feedFilter = e.currentTarget.value}>
									<option value="all">Todas las agencias</option>
									<option value="FUNVISIS">FUNVISIS</option>
									<option value="USGS">USGS</option>
									<option value="EMSC">EMSC</option>
									<option value="CSN">CSN</option>
									<option value="PRSN">PRSN</option>
									<option value="UWI">UWI</option>
									<option value="IPGP">IPGP</option>
									<option value="SGC">SGC</option>
								</select>
							</label>
							<label class="magnitude-filter">
								<span>Intensidad mínima <strong>M {sismoState.minMagnitude.toFixed(1)}</strong></span>
								<input
									type="range"
									min="0"
									max="8"
									step="0.5"
									bind:value={sismoState.minMagnitude}
								>
							</label>
						</div>
					</section>
					<section class="monitor-controls collapsible {monitorCollapsed.events ? 'collapsed' : ''}" style="flex: 1; min-height: 0; overflow: hidden;">
						<button class="collapsible-header" aria-expanded={!monitorCollapsed.events} onclick={() => monitorCollapsed.events = !monitorCollapsed.events}>
							<h3><List size={14} /> Eventos ({filteredEarthquakes.length})</h3>
							<span class="chevron"><ChevronDown size={16} /></span>
						</button>
						<div class="collapsible-content" style="flex: 1; min-height: 0; overflow-y: auto;">
							<div class="sidebar-event-list">
								{#if filteredEarthquakes.length === 0}
									<div class="placeholder-text" style="text-align: center; padding: 16px;">No hay eventos para los filtros seleccionados.</div>
								{:else}
									{#each filteredEarthquakes.slice(0, 50) as eq (eq.id)}
										<!-- svelte-ignore a11y_click_events_have_key_events -->
										<!-- svelte-ignore a11y_no_static_element_interactions -->
										<div
											class="eq-card {sismoState.selectedEarthquake?.id === eq.id ? 'selected' : ''}"
											onclick={() => handleSelectEarthquake(eq)}
										>
											<div class="eq-mag" style="background: {getMagnitudeColor(eq.magnitude)};">
												{eq.magnitude.toFixed(1)}
											</div>
											<div class="eq-details">
												<div class="eq-loc">{eq.location}</div>
												<div class="eq-meta">
													<span>{new Date(eq.time).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'})}</span>
													<span>{timeAgo(new Date(eq.time).getTime())}</span>
												</div>
											</div>
										</div>
									{/each}
								{/if}
							</div>
						</div>
					</section>
				</div>
			{:else if sismoState.sidebarPage === 'layers'}
				<div class="panel-page active">
					<MapControls
						bind:mapTheme={sismoState.mapTheme}
						bind:showPlates={sismoState.showPlates}
						bind:showVolcanoes={sismoState.showVolcanoes}
						bind:showDamage={sismoState.showDamage}
						bind:showHeatmap={sismoState.showHeatmap}
						{volcanoAlertCount}
						damageLoading={sismoState.damageLoading}
						damageIndexProgress={sismoState.damageIndexProgress}
						damageError={sismoState.damageError}
					/>
				</div>
			{:else if sismoState.sidebarPage === 'analysis'}
				<div class="panel-page active">
					<section class="analysis-stats">
						<h3><BarChart3 size={14} /> Estadísticas</h3>
						<div class="analysis-stat-grid">
							<div><span>Sismos hoy</span><strong>{sismosHoy}</strong></div>
							<div><span>Mag. máxima</span><strong>{maxMag}</strong></div>
							<div><span>CRS recibidos</span><strong>{crsRecibidos}</strong></div>
							<div><span>Usuarios</span><strong>{sismoState.activeUsers}</strong></div>
						</div>
					</section>
				</div>
			{:else if sismoState.sidebarPage === 'settings'}
				<div class="panel-page active">
					<section class="settings-section">
						<h3><SlidersHorizontal size={14} /> Tema</h3>
						<div class="theme-buttons">
							<button class="theme-btn {sismoState.theme === 'dark' ? 'active' : ''}" onclick={() => setTheme('dark')}>
								<Moon size={18} />
								<span>Oscuro</span>
							</button>
							<button class="theme-btn {sismoState.theme === 'light' ? 'active' : ''}" onclick={() => setTheme('light')}>
								<Sun size={18} />
								<span>Claro</span>
							</button>
							<button class="theme-btn {sismoState.theme === 'system' ? 'active' : ''}" onclick={() => setTheme('system')}>
								<Monitor size={18} />
								<span>Sistema</span>
							</button>
						</div>
					</section>
					<section class="settings-section">
						<h3><Settings size={14} /> Notificaciones</h3>
						<div class="settings-row">
							<span>Activar alertas de escritorio</span>
							<label class="toggle-switch">
								<input type="checkbox" checked={sismoState.notificationsEnabled} onchange={toggleNotifications}>
								<span class="toggle-slider"></span>
							</label>
						</div>
					</section>
				</div>
			{:else if sismoState.sidebarPage === 'reports'}
				<div class="panel-page active">
					<h3><Trophy size={14} /> Ranking de Países</h3>
					<RankingWidget
						rankingData={sismoState.rankingData}
						rankingMode={sismoState.rankingMode}
						onModeChange={setRankingMode}
					/>
				</div>
			{:else}
				<div class="panel-page active">
					<div style="text-align: center; padding: 40px 20px; color: var(--text-secondary);">
						<h3 style="margin-bottom: 8px; color: var(--text-primary); text-transform: capitalize;">{sismoState.sidebarPage}</h3>
						<p style="font-size: 13px;">Sección en desarrollo.</p>
					</div>
				</div>
			{/if}
		</div>
	</div>

	<!-- MAIN AREA: Map + Bottom -->
	<div class="main-area" style:display={sismoState.innerWidth <= 768 ? (sismoState.mobileTab === 'map' || sismoState.mobileTab === 'stats' ? 'flex' : 'none') : undefined}>
	<!-- MAIN: Map -->
	<div class="main glass" style:display={sismoState.innerWidth <= 768 ? (sismoState.mobileTab === 'map' ? 'flex' : 'none') : undefined}>
		<div class="widget-header-flex" style="padding: 8px 12px; position: absolute; top: 0; left: 0; right: 0; z-index: 1000; background: linear-gradient({['dark', 'satellite', 'satellite-pure'].includes(sismoState.mapTheme) ? 'rgba(0,0,0,0.7)' : 'rgba(255,255,255,0.7)'}, transparent); pointer-events: none; border-bottom: none;">
			<h2 class="widget-title" style="pointer-events: auto; display: flex; align-items: center; gap: 8px; color: {['dark', 'satellite', 'satellite-pure'].includes(sismoState.mapTheme) ? '#fff' : '#1e293b'};">
				Actividad Global
				<div style="display: flex; align-items: center; gap: 4px; background: color-mix(in srgb, var(--live-color) 10%, transparent); padding: 2px 6px; border-radius: 4px; border: 1px solid color-mix(in srgb, var(--live-color) 20%, transparent); font-size: 9px; color: var(--live-color); letter-spacing: 0.5px; margin-top: -2px;">
					<span class="live-indicator" style="width: 6px; height: 6px; margin: 0;"></span> EN VIVO
				</div>
			</h2>
		</div>


		<div bind:this={mapElement} class="map-container" style="display: {sismoState.globeMode ? 'none' : 'flex'}; flex: 1; width: 100%; border-radius: 0; z-index: 1;"></div>
		<div bind:this={globeElement} class="map-container" style="display: {!sismoState.globeMode ? 'none' : 'flex'}; flex: 1; width: 100%; border-radius: 0; z-index: 1; background: var(--bg-darker);"></div>

		<!-- Magnitud legend -->
		<div class="map-legend">
			<div class="map-legend-title">Magnitud</div>
			<div class="map-legend-item"><span class="map-legend-dot" style="background: var(--mag-low); box-shadow: 0 0 6px var(--mag-low);"></span>&lt;3.0</div>
			<div class="map-legend-item"><span class="map-legend-dot" style="background: var(--mag-medium); box-shadow: 0 0 6px var(--mag-medium);"></span>3.0–4.5</div>
			<div class="map-legend-item"><span class="map-legend-dot" style="background: var(--mag-high); box-shadow: 0 0 6px var(--mag-high);"></span>4.5–6.0</div>
			<div class="map-legend-item"><span class="map-legend-dot" style="background: var(--mag-extreme); box-shadow: 0 0 6px var(--mag-extreme);"></span>6.0+</div>
		</div>
	</div>

	<VolcanoLayer
		{map}
		{L}
		volcanoData={sismoState.volcanoData}
		volcanoAlerts={sismoState.volcanoAlerts}
		showVolcanoes={sismoState.showVolcanoes}
		selectedVolcano={sismoState.selectedVolcano}
		onSelectVolcano={handleSelectVolcanoFromLayer}
		onLoadVolcanoData={loadVolcanoData}
	/>

	<DamageLayer
		{map}
		{L}
		bind:showDamage={sismoState.showDamage}
		bind:damageLoading={sismoState.damageLoading}
		bind:damageError={sismoState.damageError}
		bind:damageIndexProgress={sismoState.damageIndexProgress}
	/>

	<!-- BOTTOM PANEL: Charts + TV (inside main-area, grid row 2) -->
	<BottomPanel
		earthquakes={filteredEarthquakes}
		alerts={sortedAlerts}
		location={sismoState.selectedEarthquake ? sismoState.selectedEarthquake.location : 'GLOBAL'}
		paused={sismoState.selectedModalData !== null}
		onSelectAlert={handleSelectAlert}
		bottomPanelOpen={sismoState.bottomPanelOpen}
		onToggle={() => sismoState.bottomPanelOpen = !sismoState.bottomPanelOpen}
		{LiveTVComponent}
		selectedEarthquake={sismoState.selectedEarthquake}
		newsItems={sismoState.newsItems}
		onSelectNews={handleSelectNews}
	/>
	</div><!-- /main-area -->

	<!-- RIGHT PANEL: Stats + Info Tabs (outside main-area, floats right) -->
	<div class="panel-right glass" style:display={sismoState.innerWidth <= 768 ? (sismoState.mobileTab === 'info' ? 'flex' : 'none') : undefined}>
		<RightPanel
			earthquakes={filteredEarthquakes}
			selectedEarthquake={sismoState.selectedEarthquake}
			onSelectEarthquake={handleSelectEarthquake}
		/>
		<InfoTabs
			activeTab={sismoState.activeTab}
			onTabChange={(id) => sismoState.activeTab = id}
			selectedEarthquake={sismoState.selectedEarthquake}
			selectedVolcano={sismoState.selectedVolcano}
			volcanoAlerts={sismoState.volcanoAlerts}
			onCloseVolcano={closeVolcano}
		/>
	</div>
</div>

{#if sismoState.activeVolcanoAlertBanner}
	<div style="position: fixed; top: 60px; left: 50%; transform: translateX(-50%); z-index: 9999; max-width: 90vw;">
		<VolcanoAlertBanner banner={sismoState.activeVolcanoAlertBanner} onDismiss={() => sismoState.activeVolcanoAlertBanner = null} />
	</div>
{/if}

<Modal data={sismoState.selectedModalData} canClose={sismoState.modalCanClose} onClose={() => sismoState.selectedModalData = null} />

<ManualModal show={sismoState.showManualModal} onClose={() => sismoState.showManualModal = false} />

{#if sismoState.innerWidth <= 768}
	<MobileNav mobileTab={sismoState.mobileTab} onTabChange={handleMobileTabChange} />
{/if}

