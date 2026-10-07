/**
 * Utilidades geofísicas para la simulación de propagación de ondas sísmicas (P y S).
 * Modelo de corteza terrestre con velocidades medias estándar (IASPEI91 / AK135).
 */

export function parseDepthKm(depth: string | number | undefined | null): number {
	if (depth === undefined || depth === null) return 10;
	if (typeof depth === 'number') return Math.max(1, depth);
	const num = parseFloat(String(depth).replace(/[^0-9.]/g, ''));
	return isNaN(num) || num <= 0 ? 10 : num;
}

/**
 * Calcula el radio superficial (en km) alcanzado por la onda en la superficie terrestre
 * considerando la profundidad del hipocentro mediante el teorema de Pitágoras.
 * d_hipo = v * t
 * r_superficie = sqrt(max(0, d_hipo^2 - h^2))
 */
export function calculateEpicentralRadius(speedKmS: number, timeSec: number, depthKm: number): number {
	const hypocentralDist = speedKmS * timeSec;
	if (hypocentralDist < depthKm) return 0; // La onda aún no rompe la superficie
	return Math.sqrt(Math.max(0, Math.pow(hypocentralDist, 2) - Math.pow(depthKm, 2)));
}

/**
 * Radio estimado de sacudimiento severo / daño estructural potencial (MMI VI+).
 * Basado en relaciones empíricas de atenuación de magnitud y profundidad.
 */
export function calculateDamageRadius(magnitude: number, depthKm: number): number {
	if (magnitude < 4.5) return 0;
	// Relación empírica estándar
	const baseRadius = Math.pow(10, 0.42 * magnitude - 0.7);
	// Atenuación por profundidad focal
	const depthPenalty = Math.max(0, depthKm - 10) * 0.4;
	return Math.max(3, Math.round(baseRadius - depthPenalty));
}

/**
 * Tiempo de ventaja para Alerta Temprana Sísmica (EEW Lead Time).
 * Diferencia de tiempo de arribo entre la Onda P (detección) y la Onda S (destructiva).
 */
export function calculateLeadTime(distKm: number, pSpeed: number = 7.2, sSpeed: number = 3.8): number {
	if (distKm <= 0) return 0;
	const tP = distKm / pSpeed;
	const tS = distKm / sSpeed;
	return Math.max(0, Math.round(tS - tP));
}

/**
 * Estimación de Intensidad Mercalli Modificada (MMI) a una distancia dada.
 */
export function estimateMMI(magnitude: number, distKm: number, depthKm: number): {
	roman: string;
	intensity: number;
	label: string;
	color: string;
} {
	const hypoDist = Math.sqrt(distKm * distKm + depthKm * depthKm);
	// Fórmula empírica de atenuación de intensidad (tipo Atkinson & Wald)
	const rawMMI = 1.5 * magnitude - 2.8 * Math.log10(Math.max(10, hypoDist)) + 2.5;
	const mmi = Math.max(1, Math.min(10, Math.round(rawMMI)));

	const scale = [
		{ roman: 'I', label: 'Imperceptible', color: '#64748b' },
		{ roman: 'II', label: 'Muy Débil', color: '#38bdf8' },
		{ roman: 'III', label: 'Leve', color: '#22c55e' },
		{ roman: 'IV', label: 'Moderado', color: '#a3e635' },
		{ roman: 'V', label: 'Poco Fuerte', color: '#facc15' },
		{ roman: 'VI', label: 'Fuerte (Daño Leve)', color: '#fb923c' },
		{ roman: 'VII', label: 'Muy Fuerte (Daño Estructural)', color: '#f87171' },
		{ roman: 'VIII', label: 'Destructivo', color: '#ef4444' },
		{ roman: 'IX', label: 'Devastador', color: '#dc2626' },
		{ roman: 'X+', label: 'Catastrófico', color: '#7f1d1d' }
	];

	const entry = scale[Math.min(scale.length - 1, Math.max(0, mmi - 1))];
	return {
		roman: entry.roman,
		intensity: mmi,
		label: entry.label,
		color: entry.color
	};
}

export function formatTimeMMSS(totalSeconds: number): string {
	const s = Math.floor(Math.max(0, totalSeconds));
	const mins = Math.floor(s / 60);
	const secs = s % 60;
	return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
}
