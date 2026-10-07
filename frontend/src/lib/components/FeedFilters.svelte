<script lang="ts">
	interface Props {
		feedFilter: string;
		onChange?: (filter: string) => void;
	}

	let { feedFilter, onChange }: Props = $props();

	const filters = [
		{ value: 'all', label: 'Todos', title: 'Todos los orígenes' },
		{ value: 'FUNVISIS', label: 'FUNVISIS', title: 'Fundación Venezolana de Investigaciones Sismológicas' },
		{ value: 'USGS', label: 'USGS', title: 'Servicio Geológico de los Estados Unidos' },
		{ value: 'EMSC', label: 'EMSC', title: 'Centro Sismológico Euromediterráneo' },
		{ value: 'CSN', label: 'CSN', title: 'Centro Sismológico Nacional de Chile' },
		{ value: 'PRSN', label: 'PRSN', title: 'Puerto Rico Seismic Network' },
		{ value: 'UWI', label: 'UWI', title: 'UWI Seismic Research Centre (Caribe Oriental)' },
		{ value: 'IPGP', label: 'IPGP', title: 'Institut de Physique du Globe de Paris (Antillas Francesas)' },
		{ value: 'SGC', label: 'SGC', title: 'Servicio Geológico Colombiano' }
	];
</script>

<div class="filters" style="margin: 0 auto;">
	{#each filters as f (f.value)}
		<button
			class="filter-btn"
			class:active={feedFilter === f.value}
			title={f.title}
			onclick={() => onChange?.(f.value)}
		>
			{f.label}
		</button>
	{/each}
</div>
