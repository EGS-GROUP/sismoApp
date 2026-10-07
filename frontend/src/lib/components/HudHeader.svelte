<script lang="ts">
	import { RefreshCw, Bell, Globe, X } from '@lucide/svelte';
	import { getMagnitudeColor, timeAgo } from '$lib/sismoStore.svelte';
	import type { Earthquake } from '$lib/types';

	let {
		sismosHoy = 0,
		maxMag = '--',
		crsRecibidos = 0,
		activeUsers = 0,
		lastUpdate = '',
		isHistorical = false,
		unseenAlerts = [] as Earthquake[],
		showAlertsPanel = false,
		onRefresh,
		onToggleAlertsPanel,
		onToggleGlobe,
		globeMode = false,
		onSelectEarthquake
	}: {
		sismosHoy?: number;
		maxMag?: string;
		crsRecibidos?: number;
		activeUsers?: number;
		lastUpdate?: string;
		isHistorical?: boolean;
		unseenAlerts?: Earthquake[];
		showAlertsPanel?: boolean;
		onRefresh?: () => void;
		onToggleAlertsPanel?: () => void;
		onToggleGlobe?: () => void;
		globeMode?: boolean;
		onSelectEarthquake?: (eq: Earthquake) => void;
	} = $props();

	let refreshing = $state(false);

	function handleRefresh() {
		refreshing = true;
		onRefresh?.();
		setTimeout(() => refreshing = false, 1000);
	}
</script>

<header class="hud-header glass">
	<div class="brand">
		Seismon<span class="version">V4.2</span>
	</div>

	<div class="header-controls">
		<button class="header-btn" onclick={handleRefresh} title="Actualizar">
			<RefreshCw size={16} class={refreshing ? 'spinning' : ''} />
		</button>

		<div class="bell-wrapper">
			<button
				class="header-btn"
				onclick={onToggleAlertsPanel}
				title="Alertas no vistas"
				style="{unseenAlerts.length > 0 ? 'color: var(--mag-high); border-color: var(--mag-high);' : ''}"
			>
				<Bell size={16} />
				{#if unseenAlerts.length > 0}
					<span class="badge">{unseenAlerts.length}</span>
				{/if}
			</button>
			{#if showAlertsPanel}
				<div class="alerts-dropdown glass">
					<div class="alerts-dropdown-header">
						<span>Alertas recientes ({unseenAlerts.length})</span>
						<button class="header-btn" onclick={onToggleAlertsPanel} style="width: 28px; height: 28px; padding: 0;">
							<X size={14} />
						</button>
					</div>
					{#if unseenAlerts.length === 0}
						<div class="alerts-empty">No hay alertas nuevas.</div>
					{:else}
						<div class="alerts-list">
							{#each unseenAlerts as eq (eq.id)}
								<button
									class="alert-item"
									onclick={() => { onSelectEarthquake?.(eq); onToggleAlertsPanel?.(); }}
								>
									<span class="alert-mag" style="color: {getMagnitudeColor(eq.magnitude)}">M{eq.magnitude.toFixed(1)}</span>
									<span class="alert-loc">{eq.location}</span>
									<span class="alert-time">{timeAgo(eq.time)}</span>
								</button>
							{/each}
						</div>
					{/if}
				</div>
			{/if}
		</div>

		<button class="header-btn" onclick={onToggleGlobe} title="Modo Globo 3D"
			style="{globeMode ? 'color: var(--accent); border-color: var(--accent);' : ''}">
			<Globe size={16} />
		</button>

	</div>

	<div class="header-meta">
		<div class="header-stats">
			<div class="stat-box">
				<span class="stat-label">Sismos</span>
				<span class="stat-value">{sismosHoy}</span>
			</div>
			<div class="stat-box">
				<span class="stat-label">Máx</span>
				<span class="stat-value">{maxMag}</span>
			</div>
			<div class="stat-box">
				<span class="stat-label">CRS</span>
				<span class="stat-value">{crsRecibidos}</span>
			</div>
			<div class="stat-box">
				<span class="stat-label">Usuarios</span>
				<span class="stat-value">{activeUsers}</span>
			</div>
		</div>
		<span class="datetime">{lastUpdate}</span>
		{#if isHistorical}
			<span style="color: var(--mag-medium); font-size: 0.75rem;">MODO HISTÓRICO</span>
		{/if}
	</div>
</header>
