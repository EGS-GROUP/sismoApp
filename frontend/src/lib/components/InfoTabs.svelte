<script lang="ts">
	import SismoTabContainer from './SismoTabContainer.svelte';
	import TabDetails from './TabDetails.svelte';
	import type { Earthquake, VolcanoAlert, VolcanoProperties } from '$lib/types';

	let {
		activeTab = 'details',
		onTabChange,
		selectedEarthquake = null,
		selectedVolcano = null,
		volcanoAlerts = [],
		onCloseVolcano
	}: {
		activeTab?: string;
		onTabChange?: (id: string) => void;
		selectedEarthquake?: Earthquake | null;
		selectedVolcano?: VolcanoProperties | null;
		volcanoAlerts?: VolcanoAlert[];
		onCloseVolcano?: () => void;
	} = $props();

	let tabs = $derived([
		{
			id: 'details',
			label: 'Detalles',
			component: TabDetails,
			props: { selectedEarthquake, selectedVolcano, volcanoAlerts, onCloseVolcano }
		}
	]);
</script>

<div style="display: flex; flex-direction: column; flex: 1; min-height: 0; overflow: hidden;">
	<SismoTabContainer {tabs} active={activeTab} onChange={onTabChange} />
</div>
