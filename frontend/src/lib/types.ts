export interface Earthquake {
	id: string;
	source: string;
	magnitude: number;
	depth: string;
	location: string;
	time: number;
	coordinates: [number, number]; // [lat, lon]
	url: string;
	felt?: number | null;
	cdi?: number | null;
	mmi?: number | null;
	alert?: string;
}

export interface VolcanoAlert {
	name: string;
	lat: number;
	lon: number;
	alert_level: string;
	color_code: string;
	synopsis: string;
	threat: string;
	observatory: string;
}

export interface TelegramAlert {
	group?: string;
	link?: string;
	text?: string;
	fullText?: string;
	type?: string;
	image?: string;
	image_url?: string;
	video?: string;
	date?: string;
	channel_id?: string;
}

export interface NewsItem {
	title: string;
	link: string;
	pubDate: string;
	source: string;
}

export interface PushSubscription {
	endpoint: string;
	keys: {
		p256dh: string;
		auth: string;
	};
}

export interface GeoJSONFeature {
	type: string;
	properties: Record<string, unknown>;
	geometry: {
		type: string;
		coordinates: unknown;
	};
}

export interface GeoJSONCollection {
	type: 'FeatureCollection';
	features: GeoJSONFeature[];
}

export interface VolcanoProperties {
	name?: string;
	elevation?: number;
	country?: string;
	region?: string;
	status?: string;
	type?: string;
	lat?: number;
	lon?: number;
	[key: string]: unknown;
}

export interface VolcanoFeature extends GeoJSONFeature {
	properties: VolcanoProperties;
}

export type RankingData = [string, { count: number; maxMag: number }][];

export interface VolcanoAlertBannerData {
	name: string;
	level: string;
	synopsis: string;
	type: string;
	previousLevel: string | null;
}

export interface ModalData {
	title: string;
	image?: string;
	bodyHTML: string;
	sourceLink?: string;
	sourceText?: string;
}

export interface WaveSimulationState {
	active: boolean;
	earthquake: Earthquake | null;
	timeSec: number;
	isPlaying: boolean;
	pSpeed: number; // km/s
	sSpeed: number; // km/s
	maxTimeSec: number; // max duration in seconds
	speedMultiplier: number; // 1x, 2x, 5x
}
