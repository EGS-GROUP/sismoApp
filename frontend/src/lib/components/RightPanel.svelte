<script lang="ts">
	import { ChevronDown, Layers, AlertTriangle } from '@lucide/svelte';
	import type { Earthquake } from '$lib/types';
	import { getMagnitudeColor, timeAgo } from '$lib/sismoStore.svelte';

	let {
		earthquakes = [],
		selectedEarthquake = null,
		onSelectEarthquake
	}: {
		earthquakes?: Earthquake[];
		selectedEarthquake?: Earthquake | null;
		onSelectEarthquake?: (earthquake: Earthquake) => void;
	} = $props();

	let collapsed = $state(false);

	function toggle() {
		collapsed = !collapsed;
	}

	function isCritical(eq: Earthquake): boolean {
		return eq.magnitude >= 4.5;
	}

	let recentEvents = $derived(earthquakes.slice(0, 30));
</script>

<div style="display: flex; flex-direction: column; flex: 0 1 40%; min-height: 0; max-height: 40%; overflow: hidden;">
	<div class="panel-body">
		<div class="collapsible {collapsed ? 'collapsed' : ''}" style="flex: 1; min-height: 0;">
			<button class="collapsible-header" aria-expanded={!collapsed} onclick={toggle}>
				<h4><Layers size={14} /> Eventos Recientes</h4>
				<span class="chevron"><ChevronDown size={16} /></span>
			</button>
			<div class="collapsible-content" style="overflow-y: auto; flex: 1; min-height: 0;">
				{#if recentEvents.length === 0}
					<div class="placeholder-text" style="text-align: center; padding: 16px;">No hay eventos para los filtros seleccionados.</div>
				{:else}
					{#each recentEvents as eq (eq.id)}
					<!-- svelte-ignore a11y_click_events_have_key_events -->
					<!-- svelte-ignore a11y_no_static_element_interactions -->
					<div
						class="eq-card {selectedEarthquake?.id === eq.id ? 'selected' : ''} {isCritical(eq) ? 'eq-critical' : ''}"
						onclick={() => onSelectEarthquake?.(eq)}
					>
						<div class="eq-mag" style="background: {getMagnitudeColor(eq.magnitude)};">
							{eq.magnitude.toFixed(1)}
						</div>
						<div class="eq-details">
							<div class="eq-loc">{eq.location}</div>
							<div class="eq-meta">
								<span>{new Date(eq.time).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'})}</span>
								<span>{timeAgo(new Date(eq.time).getTime())}</span>
								{#if isCritical(eq)}
									<span class="eq-critical-tag"><AlertTriangle size={9} /> Crítico</span>
								{/if}
							</div>
						</div>
					</div>
					{/each}
				{/if}
			</div>
		</div>
	</div>
</div>
