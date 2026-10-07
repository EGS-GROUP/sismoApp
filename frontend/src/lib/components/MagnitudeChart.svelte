<script lang="ts">
	import { onMount } from 'svelte';
	import type { Earthquake, VolcanoAlert, GeoJSONCollection } from '$lib/types';

	interface Props {
		innerWidth: number;
		mobileTab: string;
		mode: string;
		earthquakes: Earthquake[];
		volcanoAlerts: VolcanoAlert[];
		volcanoData: GeoJSONCollection | null;
		onModeChange?: (mode: 'sismos' | 'volcanes') => void;
		onSelectVolcano?: (volcano: VolcanoAlert) => void;
	}

	let { innerWidth, mobileTab, mode, earthquakes, volcanoAlerts, volcanoData, onModeChange, onSelectVolcano }: Props = $props();

	let chartCanvas: HTMLCanvasElement;
	let magChartInstance: any = null;
	let ChartClass: any = null;
	let chartReady = $state(false);
	let updateFrame: number;

	const SISMO_BINS: Record<string, number> = { '< 3.0': 0, '3.0 - 4.5': 0, '4.5 - 6.0': 0, '6.0+': 0 };
	const SISMO_COLORS = ['#10b981', '#eab308', '#ef4444', '#b91c1c'];
	const VOLCANO_BINS: Record<string, number> = { 'Advisory': 0, 'Watch': 0, 'Warning': 0 };
	const VOLCANO_COLORS = ['#eab308', '#f97316', '#ef4444'];

	// Precargar Chart.js al montar para que el cambio de pestañas sea inmediato
	// en lugar de cargar bajo demanda cuando se muestra el gráfico.
	onMount(async () => {
		const mod = await import('chart.js/auto');
		ChartClass = mod.default || mod;
		chartReady = true;
	});

	function getChartData() {
		if (mode === 'volcanes') {
			const bins = { ...VOLCANO_BINS };
			volcanoAlerts.forEach((a: any) => {
				const level = (a.alert_level || a.color_code || '').toUpperCase();
				if (level === 'WARNING' || level === 'RED') bins['Warning']++;
				else if (level === 'WATCH' || level === 'ORANGE') bins['Watch']++;
				else if (level === 'ADVISORY' || level === 'YELLOW') bins['Advisory']++;
			});
			const hasData = volcanoAlerts.length > 0;
			return {
				labels: hasData ? Object.keys(bins) : ['Sin datos'],
				datasets: [{
					label: 'Volcanes activos',
					data: hasData ? Object.values(bins) : [0],
					backgroundColor: hasData ? VOLCANO_COLORS : ['rgba(255,255,255,0.1)'],
					borderRadius: 6
				}]
			};
		}

		const bins = { ...SISMO_BINS };
		earthquakes.forEach(eq => {
			const mag = eq.magnitude;
			if (typeof mag !== 'number' || Number.isNaN(mag)) return;
			if (mag < 3.0) bins['< 3.0']++;
			else if (mag < 4.5) bins['3.0 - 4.5']++;
			else if (mag < 6.0) bins['4.5 - 6.0']++;
			else bins['6.0+']++;
		});
		return {
			labels: Object.keys(bins),
			datasets: [{
				label: 'Sismos',
				data: Object.values(bins),
				backgroundColor: SISMO_COLORS,
				borderRadius: 4
			}]
		};
	}

	function getChartOptions() {
		const baseOptions = {
			responsive: true,
			maintainAspectRatio: false,
			plugins: { legend: { display: false } },
			animation: { duration: 0 }
		};
		if (mode === 'volcanes') {
			return {
				...baseOptions,
				plugins: {
					legend: { display: false },
					tooltip: {
						callbacks: {
							label: (ctx: any) => `${ctx.raw} volcanes`
						}
					}
				},
				scales: {
					y: {
						beginAtZero: true,
						ticks: { color: '#94a3b8', stepSize: 1 },
						grid: { color: 'rgba(255,255,255,0.05)' }
					},
					x: {
						ticks: { color: '#94a3b8' },
						grid: { display: false }
					}
				}
			};
		}
		return {
			...baseOptions,
			scales: {
				y: {
					type: 'logarithmic',
					grid: { color: 'rgba(255,255,255,0.1)' },
					ticks: {
						color: '#94a3b8',
						callback: function (value: any) {
							if (value === 1 || value === 10 || value === 100 || value === 1000 || value === 10000) return value;
							return null;
						}
					}
				},
				x: { grid: { display: false }, ticks: { color: '#94a3b8' } }
			}
		};
	}

	function updateChart() {
		if (!chartCanvas || !ChartClass) return;

		cancelAnimationFrame(updateFrame);
		updateFrame = requestAnimationFrame(() => {
			const data = getChartData();

			if (magChartInstance) {
				const currentLabels = magChartInstance.data.labels || [];
				const sameLabels = currentLabels.length === data.labels.length &&
					currentLabels.every((l: string, i: number) => l === data.labels[i]);

				if (sameLabels) {
					magChartInstance.data.datasets[0].data = data.datasets[0].data;
					magChartInstance.data.datasets[0].backgroundColor = data.datasets[0].backgroundColor;
					magChartInstance.update('none');
					return;
				}
				magChartInstance.destroy();
			}

			magChartInstance = new ChartClass(chartCanvas, {
				type: 'bar',
				data,
				options: getChartOptions()
			});
		});
	}

	$effect(() => {
		if (chartCanvas && chartReady) {
			// Forzar lectura de dependencias para que el efecto se re-ejecute cuando cambian los datos o el modo
			void mode;
			void earthquakes.length;
			void volcanoAlerts.length;
			void volcanoData;
			updateChart();
		}
	});

	$effect(() => {
		return () => {
			cancelAnimationFrame(updateFrame);
			if (magChartInstance) {
				magChartInstance.destroy();
				magChartInstance = null;
			}
		};
	});
</script>

<div class="widget chart-widget" style="{innerWidth <= 768 && mobileTab !== 'info' ? 'display: none;' : ''}">
	<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px;">
		<h2 class="widget-title" style="margin: 0;">{mode === 'sismos' ? 'Distribución de Magnitudes' : 'Distribución de Alertas Volcánicas'}</h2>
		<div style="display: flex; gap: 6px;">
			<button class="tab-btn" class:active={mode === 'sismos'} onclick={() => onModeChange?.('sismos')} style="padding: 6px 12px; font-size: 12px;">Sismos</button>
			<button class="tab-btn" class:active={mode === 'volcanes'} onclick={() => onModeChange?.('volcanes')} style="padding: 6px 12px; font-size: 12px;">Volcanes</button>
		</div>
	</div>
	<div class="chart-container">
		<canvas bind:this={chartCanvas}></canvas>
	</div>
	{#if mode === 'volcanes'}
		<div style="margin-top: 16px; border-top: 1px solid var(--border-color); padding-top: 12px;">
			<h3 style="color: var(--text-primary); font-size: 13px; margin-bottom: 10px; font-weight: 600;">Volcanes con alerta activa</h3>
			{#if volcanoAlerts.length === 0}
				<p style="color: var(--text-secondary); font-size: 12px;">No hay alertas volcánicas activas reportadas por USGS.</p>
			{:else}
				<div style="display: flex; flex-direction: column; gap: 6px; max-height: 140px; overflow-y: auto;">
					{#each volcanoAlerts as v (v.name)}
						{@const level = (v.alert_level || v.color_code || '').toUpperCase()}
						{@const color = level === 'WARNING' || level === 'RED' ? '#ef4444' : level === 'WATCH' || level === 'ORANGE' ? '#f97316' : '#eab308'}
						<!-- svelte-ignore a11y_click_events_have_key_events -->
						<!-- svelte-ignore a11y_no_static_element_interactions -->
						<div style="display: flex; justify-content: space-between; align-items: center; padding: 8px 10px; background: var(--bg-input); border-radius: 6px; cursor: pointer;" onclick={() => onSelectVolcano?.(v)}>
							<span style="font-size: 12px; color: var(--text-primary);">{v.name}</span>
							<span style="font-size: 11px; color: {color}; font-weight: 700; text-transform: uppercase;">{v.alert_level || v.color_code || 'Alerta'}</span>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	{/if}
</div>
