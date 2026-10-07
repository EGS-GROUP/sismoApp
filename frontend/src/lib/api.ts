import type {
	Earthquake,
	VolcanoAlert,
	TelegramAlert,
	NewsItem,
	PushSubscription,
	GeoJSONCollection
} from './types';

class ApiError extends Error {
	constructor(message: string, public status?: number) {
		super(message);
		this.name = 'ApiError';
	}
}

async function handleResponse<T>(res: Response): Promise<T> {
	if (!res.ok) {
		throw new ApiError(`HTTP ${res.status}: ${res.statusText}`, res.status);
	}
	return res.json() as Promise<T>;
}

export async function getEarthquakes(date?: string): Promise<Earthquake[]> {
	const url = date ? `/api/earthquakes?date=${encodeURIComponent(date)}` : '/api/earthquakes';
	const res = await fetch(url);
	const data = await handleResponse<{ earthquakes?: Earthquake[] }>(res);
	return data.earthquakes || [];
}

export async function getEarthquakesByRange(startTime: string, endTime: string): Promise<Earthquake[]> {
	const url = `/api/earthquakes?starttime=${encodeURIComponent(startTime)}&endtime=${encodeURIComponent(endTime)}`;
	const res = await fetch(url);
	const data = await handleResponse<{ earthquakes?: Earthquake[] }>(res);
	return data.earthquakes || [];
}

export async function getVolcanoAlerts(): Promise<VolcanoAlert[]> {
	const res = await fetch('/api/volcano-alerts');
	return handleResponse<VolcanoAlert[]>(res);
}

export async function getNews(query = 'sismo'): Promise<NewsItem[]> {
	const res = await fetch(`/api/news?q=${encodeURIComponent(query)}`);
	return handleResponse<NewsItem[]>(res);
}

export async function getPlates(): Promise<GeoJSONCollection> {
	const res = await fetch('/plates.json');
	return handleResponse<GeoJSONCollection>(res);
}

export async function getVolcanoes(): Promise<GeoJSONCollection> {
	const res = await fetch('/volcanoes.json');
	return handleResponse<GeoJSONCollection>(res);
}

export async function getDamageData(): Promise<GeoJSONCollection> {
	const res = await fetch('/damage_sentinel1.json');
	return handleResponse<GeoJSONCollection>(res);
}

export async function getVapidPublicKey(): Promise<string> {
	const res = await fetch('/api/vapidPublicKey');
	const data = await handleResponse<{ publicKey: string }>(res);
	return data.publicKey;
}

export async function subscribePush(subscription: PushSubscription): Promise<void> {
	const res = await fetch('/api/subscribe', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(subscription)
	});
	if (!res.ok) {
		throw new ApiError(`Error suscribiendo a push: ${res.status}`, res.status);
	}
}

export function createEventSource(): EventSource {
	return new EventSource('/api/stream');
}

export function getM3U8ProxyUrl(url: string): string {
	return `/api/proxy/m3u8?url=${encodeURIComponent(url)}`;
}

export { ApiError };
export type { TelegramAlert };
