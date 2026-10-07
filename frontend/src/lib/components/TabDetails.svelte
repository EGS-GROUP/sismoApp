<script lang="ts">
	import EarthquakeDetails from './EarthquakeDetails.svelte';
	import VolcanoDetails from './VolcanoDetails.svelte';
	import type { Earthquake, VolcanoProperties, VolcanoAlert } from '$lib/types';

	interface Props {
		selectedEarthquake: Earthquake | null;
		selectedVolcano: VolcanoProperties | null;
		volcanoAlerts: VolcanoAlert[];
		onCloseVolcano?: () => void;
	}

	let { selectedEarthquake, selectedVolcano, volcanoAlerts, onCloseVolcano }: Props = $props();
</script>

<div class="tab-content active" id="tab-details" style="padding-right: 8px;">
	{#if selectedEarthquake}
		<EarthquakeDetails {selectedEarthquake} />
	{:else if selectedVolcano}
		<VolcanoDetails {selectedVolcano} {volcanoAlerts} onClose={onCloseVolcano} />
	{:else}
		<div class="placeholder-text" style="text-align: center;">
			Selecciona un sismo o volcán en el mapa para ver detalles.
		</div>
	{/if}
</div>
