<script lang="ts">
	import { RefreshCw, BookOpen, Bell, BellOff } from '@lucide/svelte';

	interface Props {
		innerWidth: number;
		mobileTab: string;
		sismosHoy: number;
		volcanoAlertCount: number;
		maxMag: string;
		crsRecibidos: number;
		activeUsers: number;
		selectedDate: string;
		lastUpdate: string;
		notificationsEnabled: boolean;
		onDateChange?: () => void;
		onRefresh?: () => void;
		onToggleNotifications?: () => void;
		onOpenManual?: () => void;
	}

	let {
		innerWidth,
		mobileTab,
		sismosHoy,
		volcanoAlertCount,
		maxMag,
		crsRecibidos,
		activeUsers,
		selectedDate = $bindable(),
		lastUpdate,
		notificationsEnabled,
		onDateChange,
		onRefresh,
		onToggleNotifications,
		onOpenManual
	}: Props = $props();
</script>

<header class="widget header-widget" style="{innerWidth <= 768 && mobileTab !== 'feed' ? 'display: none;' : ''}">
	<div class="header-titles">
		<h1><span class="logo-accent">Sismo</span>Monitor</h1>
		<p>Centro de Control Analítico (USGS, FUNVISIS & EMSC)</p>
		<div class="dev-credits">Desarrollado por <strong>JAG-MEDIA SERVICIOS, C.A.</strong></div>
	</div>
	<div class="header-stats">
		<div class="stat-box">
			<span class="stat-label">Sismos Hoy</span>
			<span class="stat-value">{sismosHoy}</span>
		</div>
		<div class="stat-box">
			<span class="stat-label">Volcanes Activos</span>
			<span class="stat-value" style="color: {volcanoAlertCount > 0 ? 'var(--mag-high)' : 'var(--text-secondary)'}">{volcanoAlertCount}</span>
		</div>
		<div class="stat-box">
			<span class="stat-label">Max Magnitud</span>
			<span class="stat-value">{maxMag}</span>
		</div>
		<div class="stat-box">
			<span class="stat-label">CRs Recibidos</span>
			<span class="stat-value">{crsRecibidos > 0 ? crsRecibidos.toLocaleString() : '--'}</span>
		</div>
		<div class="stat-box">
			<span class="stat-label">Usuarios Activos</span>
			<span class="stat-value">{activeUsers}</span>
		</div>
	</div>
	<div class="header-controls" style="display: flex; align-items: center; gap: 12px;">
		<div style="display: flex; flex-direction: column; align-items: flex-end; gap: 4px;">
			<input type="date" class="date-picker" bind:value={selectedDate} onchange={onDateChange}>
			<span id="update-time" style="font-size: 11px; color: var(--text-secondary); font-weight: 600; letter-spacing: 0.5px;">Actualizado: {lastUpdate}</span>
		</div>
		<button onclick={onRefresh} title="Actualizar Manualmente" style="display: flex; align-items: center; justify-content: center; background: var(--bg-input); border: 1px solid var(--border-color); border-radius: 6px; padding: 0 14px; align-self: stretch; color: var(--accent); cursor: pointer; transition: 0.2s;">
			<RefreshCw size="20" color="var(--accent)" />
		</button>
		<button onclick={onToggleNotifications} title="Activar/Desactivar Notificaciones" style="display: flex; align-items: center; justify-content: center; background: var(--bg-input); border: 1px solid var(--border-color); border-radius: 6px; padding: 0 14px; align-self: stretch; color: {notificationsEnabled ? 'var(--accent)' : 'var(--text-secondary)'}; cursor: pointer; transition: 0.2s;">
			{#if notificationsEnabled}
				<Bell size="20" color="var(--accent)" />
			{:else}
				<BellOff size="20" color="currentColor" />
			{/if}
		</button>
		<button onclick={onOpenManual} title="Manual del Sistema" style="display: flex; align-items: center; justify-content: center; background: var(--bg-input); border: 1px solid var(--border-color); border-radius: 6px; padding: 0 14px; align-self: stretch; color: var(--text-secondary); cursor: pointer; transition: 0.2s; text-decoration: none;">
			<BookOpen size="20" color="currentColor" />
		</button>
	</div>
</header>
