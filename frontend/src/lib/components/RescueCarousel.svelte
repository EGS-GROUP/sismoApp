<script lang="ts">
	import type { TelegramAlert } from '$lib/types';

	interface Props {
		alerts: TelegramAlert[];
		location?: string;
		paused?: boolean;
		onSelectAlert?: (alert: TelegramAlert) => void;
	}

	let { alerts, location = 'GLOBAL', paused = false, onSelectAlert }: Props = $props();

	const MAX_VISIBLE_ALERTS = 10;
	let visibleAlerts = $state<TelegramAlert[]>([]);
	let animating = $state(false);
	let debounceTimer: ReturnType<typeof setTimeout> | undefined = undefined;

	$effect(() => {
		// Debounce updates to avoid reflows on every SSE poll tick
		const input = alerts;
		clearTimeout(debounceTimer);
		debounceTimer = setTimeout(() => {
			visibleAlerts = input.slice(0, MAX_VISIBLE_ALERTS);
		}, 500);
		return () => clearTimeout(debounceTimer);
	});

	$effect(() => {
		// Retrasar la animación hasta que el layout inicial esté estable
		const id = setTimeout(() => animating = true, 300);
		return () => clearTimeout(id);
	});

	function getRescueType(type: string | undefined) {
		if (type === 'missing') return { text: 'Personas Extraviadas', class: 'tag-missing' };
		if (type === 'supplies') return { text: 'Insumos / Ayuda', class: 'tag-supplies' };
		return { text: 'Reporte General', class: 'tag-general' };
	}

	function handleClick(e: Event, alert: TelegramAlert) {
		e.preventDefault();
		e.stopPropagation();
		onSelectAlert?.(alert);
	}

</script>

<div class="widget rescue-widget">
	<h2 class="widget-title">Centro de Apoyo y Rescate (RRSS) <span class="rescue-location">{location}</span></h2>
	<div class="rescue-feed-container">
		{#if alerts.length === 0}
			<div class="placeholder-text" style="text-align: center; width: 100%; color: var(--text-secondary);">
				No hay reportes de inteligencia (OSINT) recientes en Telegram.
			</div>
		{:else}
			<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
			<div class="rescue-marquee" class:paused>
				<div class="rescue-marquee-track" class:animating>
					<div class="rescue-track-group">
					{#each visibleAlerts as alert, i (i)}
						{@const rescueInfo = getRescueType(alert.type)}
						<!-- svelte-ignore a11y_click_events_have_key_events -->
						<!-- svelte-ignore a11y_no_static_element_interactions -->
						<div class="rescue-card {rescueInfo.class}" onclick={(e) => handleClick(e, alert)}>
							<div class="rescue-card-header">
								<span class="rescue-type">{rescueInfo.text}</span>
								<span class="rescue-time">{alert.date ? new Date(alert.date).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'}) : ''}</span>
							</div>
							
							<div class="rescue-content">
								{#if alert.image_url}
									<!-- svelte-ignore a11y_missing_attribute -->
									<img src={alert.image_url} class="rescue-image" loading="lazy" decoding="async" />
								{/if}
								<p>{alert.text}</p>
							</div>

							<div class="rescue-actions">
								<button class="btn-action">Difundir</button>
								<a href="https://t.me/{(alert.channel_id || '').replace('@', '')}" target="_blank" class="btn-primary">Contactar Grupo</a>
							</div>
							
							<div class="rescue-source">
								➤ Fuente: {alert.channel_id || 'Desconocido'}
							</div>
						</div>
					{/each}
					</div>
					<div class="rescue-track-group">
					{#each visibleAlerts as alert, i (i)}
						{@const rescueInfo = getRescueType(alert.type)}
						<!-- svelte-ignore a11y_click_events_have_key_events -->
						<!-- svelte-ignore a11y_no_static_element_interactions -->
						<div class="rescue-card {rescueInfo.class}" onclick={(e) => handleClick(e, alert)}>
							<div class="rescue-card-header">
								<span class="rescue-type">{rescueInfo.text}</span>
								<span class="rescue-time">{alert.date ? new Date(alert.date).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'}) : ''}</span>
							</div>
							
							<div class="rescue-content">
								{#if alert.image_url}
									<!-- svelte-ignore a11y_missing_attribute -->
									<img src={alert.image_url} class="rescue-image" loading="lazy" decoding="async" />
								{/if}
								<p>{alert.text}</p>
							</div>

							<div class="rescue-actions">
								<button class="btn-action">Difundir</button>
								<a href="https://t.me/{(alert.channel_id || '').replace('@', '')}" target="_blank" class="btn-primary">Contactar Grupo</a>
							</div>
							
							<div class="rescue-source">
								➤ Fuente: {alert.channel_id || 'Desconocido'}
							</div>
						</div>
					{/each}
					</div>
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	.rescue-widget {
		display: flex;
		flex-direction: column;
		flex: 1;
		min-height: 0;
	}

	.rescue-location {
		color: var(--accent);
		font-weight: normal;
		font-size: 11px;
		margin-left: 8px;
	}

	.rescue-feed-container {
		flex: 1;
		overflow: hidden;
		padding: 12px;
		display: flex;
		align-items: center;
		scrollbar-width: none;
		-ms-overflow-style: none;
	}
	.rescue-feed-container::-webkit-scrollbar {
		display: none;
	}

	.rescue-marquee {
		overflow: hidden;
		flex: 1;
		display: flex;
		align-items: center;
		mask-image: linear-gradient(to right, transparent, black 5%, black 95%, transparent);
		-webkit-mask-image: linear-gradient(to right, transparent, black 5%, black 95%, transparent);
		contain: layout paint;
	}

	.rescue-marquee.paused .rescue-marquee-track.animating {
		animation-play-state: paused;
	}

	.rescue-track-group {
		display: flex;
		gap: 12px;
		flex-shrink: 0;
	}

	.rescue-marquee-track {
		--marquee-duration: 120s;
		display: flex;
		gap: 12px;
		width: max-content;
		will-change: transform;
	}

	.rescue-marquee-track.animating {
		animation: marquee var(--marquee-duration) linear infinite;
	}

	.rescue-marquee-track.animating:hover {
		animation-play-state: paused;
	}

	@keyframes marquee {
		0% { transform: translateX(0); }
		100% { transform: translateX(-50%); }
	}

	.rescue-card {
		display: inline-flex;
		flex-direction: column;
		white-space: normal;
		background: var(--bg-overlay);
		border: 1px solid var(--border-color);
		border-radius: 8px;
		padding: 12px;
		width: 300px;
		flex-shrink: 0;
		text-align: left;
		cursor: pointer;
	}

	.rescue-card-header {
		display: flex;
		justify-content: space-between;
		margin-bottom: 8px;
	}

	.rescue-type {
		font-size: 10px;
		font-weight: bold;
		padding: 2px 6px;
		border-radius: 4px;
		border: 1px solid currentColor;
	}
	.rescue-card.tag-missing .rescue-type { color: var(--mag-high); }
	.rescue-card.tag-supplies .rescue-type { color: var(--mag-medium); }
	.rescue-card.tag-general .rescue-type { color: var(--source-usgs); }

	.rescue-time {
		font-size: 10px;
		color: var(--text-secondary);
	}

	.rescue-content {
		flex: 1;
		overflow: hidden;
		margin-bottom: 8px;
	}

	.rescue-content p {
		font-size: 12px;
		color: var(--text-primary);
		margin: 0;
		display: -webkit-box;
		-webkit-line-clamp: 4;
		line-clamp: 4;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	.rescue-image {
		width: 100%;
		height: 120px;
		object-fit: cover;
		border-radius: 4px;
		margin-bottom: 8px;
	}

	.rescue-actions {
		display: flex;
		gap: 8px;
		margin-bottom: 8px;
	}

	.rescue-actions .btn-action,
	.rescue-actions .btn-primary {
		flex: 1;
		text-align: center;
		padding: 6px;
		font-size: 11px;
		border-radius: 4px;
		text-decoration: none;
		cursor: pointer;
		border: none;
	}

	.rescue-actions .btn-action {
		background: var(--bg-input);
		color: var(--text-primary);
	}

	.rescue-actions .btn-primary {
		background: color-mix(in srgb, var(--accent) 15%, transparent);
		color: var(--accent);
	}

	.rescue-source {
		font-size: 10px;
		color: var(--text-secondary);
	}
</style>
