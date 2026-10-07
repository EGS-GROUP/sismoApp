import type {
	Earthquake,
	VolcanoAlert,
	TelegramAlert,
	NewsItem,
	GeoJSONCollection,
	RankingData,
	VolcanoAlertBannerData,
	VolcanoProperties,
	ModalData
} from './types';

export const sismoState = $state({
	// Data
	earthquakes: [] as Earthquake[],
	alerts: [] as TelegramAlert[],
	newsItems: [] as NewsItem[],
	volcanoAlerts: [] as VolcanoAlert[],
	volcanoData: null as GeoJSONCollection | null,

	// Selection
	selectedEarthquake: null as Earthquake | null,
	selectedVolcano: null as VolcanoProperties | null,
	selectedModalData: null as ModalData | null,
	modalCanClose: false,

	// UI State
	feedFilter: 'all' as string,
	minMagnitude: 0,
	activeTab: 'details' as string,
	rankingMode: 'global' as 'global' | 'venezuela',
	chartMode: 'sismos' as 'sismos' | 'volcanes',

	// Map state
	showPlates: true,
	mapTheme: 'dark' as 'dark' | 'satellite',
	showDamage: false,
	damageLoading: false,
	damageError: '',
	damageIndexProgress: 0,
	showVolcanoes: false,
	showHeatmap: false,

	// Globe state
	globeMode: false,
	globeMapStyle: 'dark' as string,

	// Layout
	innerWidth: 1024,
	mobileTab: 'feed' as 'feed' | 'map' | 'info' | 'stats',
	leftPanelOpen: true,
	rightPanelOpen: true,
	bottomPanelOpen: true,
	sidebarPage: 'monitor' as string,

	// Modal
	showManualModal: false,
	manualTab: 'intro' as string,

	// Stats
	activeUsers: 0,
	lastUpdate: new Date().toLocaleTimeString(),
	isHistorical: false,
	selectedDate: '',
	timeRange: 'live' as 'live' | '24h' | '7d' | 'custom',
	dateStart: '',
	dateEnd: '',
	notificationsEnabled: false,

	// Theme
	theme: 'dark' as 'dark' | 'light' | 'system',

	// Unseen alerts (events missed while away)
	unseenAlerts: [] as Earthquake[],
	showAlertsPanel: false,

	// Ranking
	rankingData: [] as RankingData,

	// Volcano alerts
	activeVolcanoAlertBanner: null as VolcanoAlertBannerData | null,
});

export function getMagnitudeColor(m: number): string {
	if (m >= 6.0) return 'var(--mag-extreme)';
	if (m >= 4.5) return 'var(--mag-high)';
	if (m >= 3.0) return 'var(--mag-medium)';
	return 'var(--mag-low)';
}

export function getMagnitudeHex(m: number): string {
	const s = getComputedStyle(document.documentElement);
	if (m >= 6.0) return s.getPropertyValue('--mag-extreme').trim() || '#b91c1c';
	if (m >= 4.5) return s.getPropertyValue('--mag-high').trim() || '#ef4444';
	if (m >= 3.0) return s.getPropertyValue('--mag-medium').trim() || '#eab308';
	return s.getPropertyValue('--mag-low').trim() || '#10b981';
}

export function magnitudeClass(m: number): string {
	if (m >= 6) return 'critical';
	if (m >= 5) return 'warning';
	return 'info';
}

export function timeAgo(timeMs: number): string {
	const diff = Math.floor((Date.now() - timeMs) / 1000);
	if (diff < 0) return 'Hace instantes';
	if (diff < 60) return `${diff}s`;
	const m = Math.floor(diff / 60);
	if (m < 60) return `(Hace ${m}m)`;
	const h = Math.floor(m / 60);
	if (h < 24) return `(Hace ${h}h)`;
	return `(Hace ${Math.floor(h / 24)}d)`;
}
