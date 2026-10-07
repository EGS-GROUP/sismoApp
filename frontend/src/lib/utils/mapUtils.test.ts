import { describe, it, expect } from 'vitest';
import { getMagnitudeColor } from './mapUtils';

describe('getMagnitudeColor', () => {
	it('returns low color for magnitude below 3.0', () => {
		expect(getMagnitudeColor(1.5)).toBe('var(--mag-low)');
		expect(getMagnitudeColor(2.9)).toBe('var(--mag-low)');
	});

	it('returns medium color for magnitude between 3.0 and 4.5', () => {
		expect(getMagnitudeColor(3.0)).toBe('var(--mag-medium)');
		expect(getMagnitudeColor(4.4)).toBe('var(--mag-medium)');
	});

	it('returns high color for magnitude between 4.5 and 6.0', () => {
		expect(getMagnitudeColor(4.5)).toBe('var(--mag-high)');
		expect(getMagnitudeColor(5.9)).toBe('var(--mag-high)');
	});

	it('returns extreme color for magnitude 6.0 or greater', () => {
		expect(getMagnitudeColor(6.0)).toBe('var(--mag-extreme)');
		expect(getMagnitudeColor(7.8)).toBe('var(--mag-extreme)');
	});
});
