import { describe, it, expect } from 'vitest';
import { findVolcanoAlert, findVolcanoFeatureByName, normalizeVolcanoName, volcanoAlertStyleMap } from './volcanoUtils';
import type { VolcanoAlert, VolcanoFeature, GeoJSONCollection } from '$lib/types';

const alert = (name: string, level: string): VolcanoAlert => ({
	name,
	lat: 10,
	lon: -70,
	alert_level: level,
	color_code: level,
	synopsis: '',
	threat: '',
	observatory: ''
});

describe('normalizeVolcanoName', () => {
	it('strips common prefixes and normalizes spacing', () => {
		expect(normalizeVolcanoName('Mount St. Helens')).toBe('st helens');
		expect(normalizeVolcanoName('Volcán Villarrica')).toBe('villarrica');
	});
});

describe('findVolcanoAlert', () => {
	it('finds exact match', () => {
		const alerts = [alert('Villarrica', 'ORANGE'), alert('Etna', 'YELLOW')];
		expect(findVolcanoAlert('Villarrica', alerts)?.name).toBe('Villarrica');
	});

	it('finds normalized match', () => {
		const alerts = [alert('Mount Villarrica', 'ORANGE')];
		expect(findVolcanoAlert('Volcán Villarrica', alerts)?.name).toBe('Mount Villarrica');
	});

	it('returns null when no match', () => {
		expect(findVolcanoAlert('Fuego', [])).toBeNull();
	});

	it('returns null for empty volcano name', () => {
		expect(findVolcanoAlert('', [alert('Etna', 'YELLOW')])).toBeNull();
	});
});

describe('findVolcanoFeatureByName', () => {
	const feature = (name: string): VolcanoFeature => ({
		type: 'Feature',
		properties: { name },
		geometry: { type: 'Point', coordinates: [0, 0] }
	});

	const collection: GeoJSONCollection = {
		type: 'FeatureCollection',
		features: [feature('Villarrica'), feature('Etna')]
	};

	it('finds feature by exact name', () => {
		expect(findVolcanoFeatureByName('Etna', collection)?.properties.name).toBe('Etna');
	});

	it('returns null when no features match', () => {
		expect(findVolcanoFeatureByName('Fuego', collection)).toBeNull();
	});

	it('returns null for null collection', () => {
		expect(findVolcanoFeatureByName('Etna', null)).toBeNull();
	});
});

describe('volcanoAlertStyleMap', () => {
	it('contains red for WARNING', () => {
		expect(volcanoAlertStyleMap.WARNING.color).toBe('#ff0000');
	});
});
