import { describe, it, expect } from 'vitest';
import { extractRegion, extractVenezuelaState } from './geoUtils';

describe('extractRegion', () => {
	it('returns Venezuela when location contains venezuela', () => {
		expect(extractRegion('10 km SSE of Caracas, Venezuela')).toBe('Venezuela');
	});

	it('returns México for Mexico variants', () => {
		expect(extractRegion('Near coast of Chiapas, Mexico')).toBe('México');
		expect(extractRegion('Oaxaca, México')).toBe('México');
	});

	it('returns Estados Unidos for US state names', () => {
		expect(extractRegion('Los Angeles, California')).toBe('Estados Unidos');
		expect(extractRegion('Miami, FL')).toBe('Estados Unidos');
	});

	it('returns last comma segment for unknown locations', () => {
		expect(extractRegion('Southern Peru')).toBe('Perú');
		expect(extractRegion('Off the coast of Japan')).toBe('Japón');
	});

	it('returns Desconocido for empty string', () => {
		expect(extractRegion('')).toBe('Desconocido');
	});
});

describe('extractVenezuelaState', () => {
	it('detects Sucre', () => {
		expect(extractVenezuelaState('Carúpano, Venezuela')).toBe('Sucre');
	});

	it('detects Carabobo', () => {
		expect(extractVenezuelaState('Valencia, Venezuela')).toBe('Carabobo');
	});

	it('detects Distrito Capital', () => {
		expect(extractVenezuelaState('Caracas, Venezuela')).toBe('Distrito Capital');
	});

	it('returns Otras Regiones for unknown Venezuelan location', () => {
		expect(extractVenezuelaState('Someplace, Venezuela')).toBe('Otras Regiones');
	});

	it('returns Desconocido for empty string', () => {
		expect(extractVenezuelaState('')).toBe('Desconocido');
	});
});
