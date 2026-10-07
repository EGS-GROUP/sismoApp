export function getMagnitudeColor(mag: number) {
	if (mag >= 6.0) return 'var(--mag-extreme)';
	if (mag >= 4.5) return 'var(--mag-high)';
	if (mag >= 3.0) return 'var(--mag-medium)';
	return 'var(--mag-low)';
}
