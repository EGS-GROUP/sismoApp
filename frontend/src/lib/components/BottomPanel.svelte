<script lang="ts">
	import { onDestroy } from 'svelte';
	import { ChevronDown, ChevronUp, BarChart3, Radio, Tv, Newspaper } from '@lucide/svelte';
	import type { Earthquake, TelegramAlert } from '$lib/types';
	import { getMagnitudeHex, sismoState } from '$lib/sismoStore.svelte';
	import RescueCarousel from './RescueCarousel.svelte';
	import NewsFeed from './NewsFeed.svelte';
	import Chart from 'chart.js/auto';

	let {
		earthquakes = [],
		alerts = [],
		location = 'GLOBAL',
		paused = false,
		onSelectAlert,
		bottomPanelOpen = true,
		onToggle,
		LiveTVComponent = null,
		selectedEarthquake = null,
		newsItems = [],
		onSelectNews
	}: {
		earthquakes?: Earthquake[];
		alerts?: TelegramAlert[];
		location?: string;
		paused?: boolean;
		onSelectAlert?: (alert: TelegramAlert) => void;
		bottomPanelOpen?: boolean;
		onToggle?: () => void;
		LiveTVComponent?: any;
		selectedEarthquake?: Earthquake | null;
		newsItems?: any[];
		onSelectNews?: (n: any) => void;
	} = $props();

	let activeBottomTab = $state<'charts' | 'support' | 'tv' | 'news'>('charts');

	let timelineCanvas: HTMLCanvasElement = $state()!;
	let magnitudeCanvas: HTMLCanvasElement = $state()!;
	let depthCanvas: HTMLCanvasElement = $state()!;
	let energyCanvas: HTMLCanvasElement = $state()!;
	let timelineChart: Chart | null = null;
	let magnitudeChart: Chart | null = null;
	let depthChart: Chart | null = null;
	let energyChart: Chart | null = null;


	function destroyCharts() {
		timelineChart?.destroy(); timelineChart = null;
		magnitudeChart?.destroy(); magnitudeChart = null;
		depthChart?.destroy(); depthChart = null;
		energyChart?.destroy(); energyChart = null;
	}

	function buildCharts() {
		if (!timelineCanvas || !magnitudeCanvas || !depthCanvas || !energyCanvas) return;
		destroyCharts();

	const rootStyles = getComputedStyle(document.documentElement);
		const chartTextColor = rootStyles.getPropertyValue('--chart-text-color').trim() || '#94a3b8';
		const chartGridColor = rootStyles.getPropertyValue('--chart-grid-color').trim() || 'rgba(255,255,255,0.05)';
		const accentColor = rootStyles.getPropertyValue('--accent').trim() || '#00e5ff';
		function withAlpha(color: string, alpha: number): string {
			const m = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(color);
			if (m) {
				return `rgba(${parseInt(m[1], 16)}, ${parseInt(m[2], 16)}, ${parseInt(m[3], 16)}, ${alpha})`;
			}
			const rgb = color.match(/rgba?\s*\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)/);
			if (rgb) return `rgba(${rgb[1]}, ${rgb[2]}, ${rgb[3]}, ${alpha})`;
			return color;
		}
		const chartTooltipBg = rootStyles.getPropertyValue('--bg-widget').trim() || 'rgba(0,0,0,0.7)';
		const chartTooltipBorder = rootStyles.getPropertyValue('--border-color').trim() || 'rgba(255,255,255,0.1)';
		const tooltipOptions = { backgroundColor: chartTooltipBg, titleColor: chartTextColor, bodyColor: chartTextColor, borderColor: chartTooltipBorder, borderWidth: 1, padding: 8, cornerRadius: 4, displayColors: false };

		const eqs = [...earthquakes].sort((a, b) => new Date(a.time).getTime() - new Date(b.time).getTime());
		const selectedId = selectedEarthquake?.id ?? null;

		// 1. Timeline scatter
		timelineChart = new Chart(timelineCanvas, {
			type: 'scatter',
			data: {
				datasets: [{
					label: 'Sismos',
					data: eqs.map(eq => ({
						x: new Date(eq.time).getTime(),
						y: eq.magnitude,
						id: eq.id
					})),
					backgroundColor: eqs.map(eq => getMagnitudeHex(eq.magnitude)),
					pointRadius: eqs.map(eq => eq.id === selectedId ? 9 : Math.max(eq.magnitude * 0.5, 2)),
					pointHoverRadius: eqs.map(eq => eq.id === selectedId ? 12 : 6),
					borderWidth: eqs.map(eq => eq.id === selectedId ? 2 : 0),
					borderColor: eqs.map(eq => eq.id === selectedId ? '#ffffff' : 'transparent'),
				}]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				plugins: {
					legend: { display: false },
					tooltip: {
						...tooltipOptions,
						callbacks: {
							label: (ctx: any) => {
								const d = new Date(ctx.parsed.x);
								const time = d.toLocaleDateString('es-VE', { day: '2-digit', month: '2-digit' }) + ' ' + d.toLocaleTimeString('es-VE', { hour: '2-digit', minute: '2-digit' });
								return 'Mag ' + ctx.parsed.y + ' | ' + time;
							}
						}
					}
				},
				scales: {
					x: {
						type: 'linear',
						ticks: {
							color: chartTextColor,
							font: { size: 9 },
							maxTicksLimit: 6,
							callback: (v: any) => {
								const d = new Date(v);
								return d.toLocaleDateString('es-VE', { day: '2-digit', month: '2-digit' }) + ' ' + d.toLocaleTimeString('es-VE', { hour: '2-digit', minute: '2-digit' });
							}
						},
						grid: { color: chartGridColor }
					},
					y: {
						ticks: { color: chartTextColor, font: { size: 9 } },
						grid: { color: chartGridColor },
						title: { display: true, text: 'Mag', color: chartTextColor, font: { size: 9 } }
					}
				}
			}
		});

		// 2. Magnitude distribution
		const magBuckets = { '<3': 0, '3-4': 0, '4-5': 0, '5-6': 0, '6+': 0 };
		eqs.forEach(eq => {
			if (eq.magnitude < 3) magBuckets['<3']++;
			else if (eq.magnitude < 4) magBuckets['3-4']++;
			else if (eq.magnitude < 5) magBuckets['4-5']++;
			else if (eq.magnitude < 6) magBuckets['5-6']++;
			else magBuckets['6+']++;
		});
		magnitudeChart = new Chart(magnitudeCanvas, {
			type: 'bar',
			data: {
				labels: Object.keys(magBuckets),
				datasets: [{
					data: Object.values(magBuckets),
					backgroundColor: [getMagnitudeHex(2), getMagnitudeHex(3.5), getMagnitudeHex(4.5), getMagnitudeHex(5.5), getMagnitudeHex(6.5)],
					borderRadius: 4
				}]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				plugins: { legend: { display: false }, tooltip: tooltipOptions },
				scales: {
					x: { ticks: { color: chartTextColor, font: { size: 9 } }, grid: { display: false } },
					y: { ticks: { color: chartTextColor, font: { size: 9 } }, grid: { color: chartGridColor } }
				}
			}
		});

		// 3. Depth profile
		depthChart = new Chart(depthCanvas, {
			type: 'scatter',
			data: {
				datasets: [{
					label: 'Profundidad',
					data: eqs.map(eq => ({
						x: eq.magnitude,
						y: parseFloat(eq.depth) || 0
					})),
					backgroundColor: eqs.map(eq => getMagnitudeHex(eq.magnitude)),
					pointRadius: eqs.map(eq => eq.id === selectedId ? 9 : 2),
					pointHoverRadius: eqs.map(eq => eq.id === selectedId ? 12 : 6),
					borderWidth: eqs.map(eq => eq.id === selectedId ? 2 : 0),
					borderColor: eqs.map(eq => eq.id === selectedId ? '#ffffff' : 'transparent'),
				}]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				plugins: {
					legend: { display: false },
					tooltip: {
						...tooltipOptions,
						callbacks: {
							label: (ctx: any) => 'Mag ' + ctx.parsed.x + ' | Prof: ' + ctx.parsed.y + ' km'
						}
					}
				},
				scales: {
					x: {
						ticks: { color: chartTextColor, font: { size: 9 } },
						grid: { color: chartGridColor },
						title: { display: true, text: 'Mag', color: chartTextColor, font: { size: 9 } }
					},
					y: {
						reverse: true,
						ticks: { color: chartTextColor, font: { size: 9 } },
						grid: { color: chartGridColor },
						title: { display: true, text: 'km', color: chartTextColor, font: { size: 9 } }
					}
				}
			}
		});

		// 4. Cumulative energy
		let cumulative = 0;
		const energyData = eqs.map(eq => {
			const energy = Math.pow(10, 1.5 * eq.magnitude);
			cumulative += energy;
			return { x: new Date(eq.time).getTime(), y: cumulative };
		});
		energyChart = new Chart(energyCanvas, {
			type: 'line',
			data: {
				datasets: [{
					label: 'Energía Acumulada',
					data: energyData,
					borderColor: accentColor,
					backgroundColor: withAlpha(accentColor, 0.1),
					fill: true,
					pointRadius: 0,
					borderWidth: 2
				}]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				plugins: {
					legend: { display: false },
					tooltip: {
						...tooltipOptions,
						callbacks: {
							label: (ctx: any) => {
								const d = new Date(ctx.parsed.x);
								const time = d.toLocaleDateString('es-VE', { day: '2-digit', month: '2-digit' }) + ' ' + d.toLocaleTimeString('es-VE', { hour: '2-digit', minute: '2-digit' });
								return 'Energía: ' + Number(ctx.parsed.y).toExponential(1) + ' | ' + time;
							}
						}
					}
				},
				scales: {
					x: {
						type: 'linear',
						ticks: {
							color: chartTextColor,
							font: { size: 9 },
							maxTicksLimit: 6,
							callback: (v: any) => {
								const d = new Date(v);
								return d.toLocaleDateString('es-VE', { day: '2-digit', month: '2-digit' }) + ' ' + d.toLocaleTimeString('es-VE', { hour: '2-digit', minute: '2-digit' });
							}
						},
						grid: { color: chartGridColor }
					},
					y: {
						ticks: { color: chartTextColor, font: { size: 9 }, callback: (v: any) => Number(v).toExponential(1) },
						grid: { color: chartGridColor }
					}
				}
			}
		});
	}

	$effect(() => {
		if (bottomPanelOpen && activeBottomTab === 'charts' && earthquakes.length > 0) {
			// selectedEarthquake y tema se referencian para disparar el rebuild
			void selectedEarthquake?.id;
			void sismoState.theme;
			requestAnimationFrame(() => buildCharts());
		}
	});

	onDestroy(() => destroyCharts());
</script>

<div class="bottom-panel glass" class:open={bottomPanelOpen}>
	<button class="bottom-toggle" onclick={onToggle} aria-label={bottomPanelOpen ? 'Cerrar panel inferior' : 'Abrir panel inferior'}>
		<span class="bottom-handle"></span>
		{#if bottomPanelOpen}
			<ChevronDown size={18} />
		{:else}
			<ChevronUp size={18} />
		{/if}
	</button>
	{#if bottomPanelOpen}
		<div class="bottom-tabs" role="tablist" aria-label="Panel inferior">
			<button
				class:active={activeBottomTab === 'charts'}
				onclick={() => activeBottomTab = 'charts'}
				role="tab"
				aria-selected={activeBottomTab === 'charts'}
			>
				<BarChart3 size={14} /> Estadísticas
			</button>
			<button
				class:active={activeBottomTab === 'tv'}
				onclick={() => activeBottomTab = 'tv'}
				role="tab"
				aria-selected={activeBottomTab === 'tv'}
			>
				<Tv size={14} /> TV en Vivo
			</button>
			<button
				class:active={activeBottomTab === 'support'}
				onclick={() => activeBottomTab = 'support'}
				role="tab"
				aria-selected={activeBottomTab === 'support'}
			>
				<Radio size={14} /> Centro de Apoyo
			</button>
			<button
				class:active={activeBottomTab === 'news'}
				onclick={() => activeBottomTab = 'news'}
				role="tab"
				aria-selected={activeBottomTab === 'news'}
			>
				<Newspaper size={14} /> Noticias
			</button>
		</div>

		{#if activeBottomTab === 'charts'}
			<div class="bottom-body">
				<div class="bottom-module">
					<h4>Línea Temporal</h4>
					<div class="chart-box"><canvas bind:this={timelineCanvas}></canvas></div>
				</div>
				<div class="bottom-module">
					<h4>Distribución de Magnitud</h4>
					<div class="chart-box"><canvas bind:this={magnitudeCanvas}></canvas></div>
				</div>
				<div class="bottom-module">
					<h4>Perfil de Profundidad</h4>
					<div class="chart-box"><canvas bind:this={depthCanvas}></canvas></div>
				</div>
				<div class="bottom-module">
					<h4>Energía Acumulada</h4>
					<div class="chart-box"><canvas bind:this={energyCanvas}></canvas></div>
				</div>
			</div>
		{:else if activeBottomTab === 'tv'}
			<div class="bottom-tv">
				{#if LiveTVComponent}
					<LiveTVComponent {selectedEarthquake} activeTab="tv" />
				{:else}
					<div class="placeholder-text" style="text-align: center; padding: 24px;">Cargando TV...</div>
				{/if}
			</div>
		{:else if activeBottomTab === 'news'}
			<div class="bottom-support">
				<NewsFeed {newsItems} {onSelectNews} />
			</div>
		{:else}
			<div class="bottom-support">
				<RescueCarousel {alerts} {location} {paused} {onSelectAlert} />
			</div>
		{/if}
	{/if}
</div>
