export const volcanoAlertStyleMap: Record<string, { radius: number; color: string; pulse: boolean }> = {
	'RED':     { radius: 18, color: '#ff0000', pulse: true },
	'WARNING': { radius: 18, color: '#ff0000', pulse: true },
	'ORANGE':  { radius: 16, color: '#ff8c00', pulse: true },
	'WATCH':   { radius: 16, color: '#ff8c00', pulse: true },
	'YELLOW':  { radius: 14, color: '#ffd700', pulse: false },
	'ADVISORY':{ radius: 14, color: '#ffd700', pulse: false },
	'GREEN':   { radius: 12, color: '#00ff00', pulse: false },
	'NORMAL':  { radius: 12, color: '#00ff00', pulse: false },
	'UNASSIGNED': { radius: 10, color: '#cccccc', pulse: false }
};

export function normalizeVolcanoName(name: string): string {
	return name.toLowerCase()
		.replace(/\b(mount|mt|mt\.|volcano|volcán|cerro|nevado|pico|santa?\s)/g, ' ')
		.replace(/[^a-z0-9]/g, ' ')
		.replace(/\s+/g, ' ')
		.trim();
}

import type { VolcanoAlert, VolcanoFeature, GeoJSONCollection } from '$lib/types';

export function findVolcanoAlert(volcanoName: string, volcanoAlerts: VolcanoAlert[]): VolcanoAlert | null {
	if (!volcanoName || volcanoAlerts.length === 0) return null;
	const nameLower = volcanoName.toLowerCase();
	const nameNorm = normalizeVolcanoName(volcanoName);
	const nameWords = nameNorm.split(' ').filter(w => w.length > 2);

	for (const a of volcanoAlerts) {
		if (a.name && a.name.toLowerCase() === nameLower) return a;
	}

	for (const a of volcanoAlerts) {
		if (a.name && normalizeVolcanoName(a.name) === nameNorm) return a;
	}

	for (const a of volcanoAlerts) {
		if (!a.name) continue;
		const alertLower = a.name.toLowerCase();
		if (alertLower.includes(nameLower) || nameLower.includes(alertLower)) return a;
	}

	let bestMatch: VolcanoAlert | null = null;
	let bestOverlap = 0;
	for (const a of volcanoAlerts) {
		if (!a.name) continue;
		const alertWords = normalizeVolcanoName(a.name).split(' ').filter(w => w.length > 2);
		const overlap = nameWords.filter(w => alertWords.includes(w)).length;
		if (overlap > bestOverlap) {
			bestOverlap = overlap;
			bestMatch = a;
		}
	}
	if (bestOverlap >= 2) return bestMatch;

	return null;
}

export function findVolcanoFeatureByName(name: string, volcanoData: GeoJSONCollection | null): VolcanoFeature | null {
	if (!volcanoData?.features) return null;
	const nameLower = name.toLowerCase();
	const nameNorm = normalizeVolcanoName(name);
	for (const f of volcanoData.features) {
		const pName = (f.properties?.name || '') as string;
		if (pName.toLowerCase() === nameLower) return f as VolcanoFeature;
		if (normalizeVolcanoName(pName) === nameNorm) return f as VolcanoFeature;
		const pNameLower = pName.toLowerCase();
		if (pNameLower.includes(nameLower) || nameLower.includes(pNameLower)) return f as VolcanoFeature;
	}
	return null;
}
