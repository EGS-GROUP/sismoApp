<script lang="ts">
	import { Layers } from '@lucide/svelte';

	interface Props {
		mapTheme?: string;
		showPlates?: boolean;
		showVolcanoes?: boolean;
		showDamage?: boolean;
		showHeatmap?: boolean;
		volcanoAlertCount?: number;
		damageLoading?: boolean;
		damageIndexProgress?: number;
		damageError?: string;
	}

	let {
		mapTheme = $bindable('dark'),
		showPlates = $bindable(false),
		showVolcanoes = $bindable(false),
		showDamage = $bindable(false),
		showHeatmap = $bindable(false),
		volcanoAlertCount = 0,
		damageLoading = false,
		damageIndexProgress = 0,
		damageError = ''
	}: Props = $props();
</script>

<section class="layer-controls" aria-label="Capas del mapa">
	<h3><Layers size={14} /> Capas del mapa</h3>
	<div class="layer-control-group">
		<span class="layer-control-label">Mapa base</span>
		<label>
			<input type="radio" name="mapTheme" value="dark" bind:group={mapTheme}>
			Oscuro
		</label>
		<label>
			<input type="radio" name="mapTheme" value="light" bind:group={mapTheme}>
			Claro
		</label>
		<label>
			<input type="radio" name="mapTheme" value="satellite" bind:group={mapTheme}>
			Satélite híbrido
		</label>
		<label>
			<input type="radio" name="mapTheme" value="satellite-pure" bind:group={mapTheme}>
			Satélite puro
		</label>
		<label title="Relieve, elevación y rutas al aire libre">
			<input type="radio" name="mapTheme" value="terrain" bind:group={mapTheme}>
			Terreno (relieve)
		</label>
		<label title="Mapa topográfico con batimetría — Esri">
			<input type="radio" name="mapTheme" value="topo" bind:group={mapTheme}>
			Topográfico
		</label>
		<label title="Mapa callejero estándar">
			<input type="radio" name="mapTheme" value="streets" bind:group={mapTheme}>
			Calles / Urbano
		</label>
	</div>
	<div class="layer-control-group">
		<span class="layer-control-label">Capas</span>
		<label>
			<input type="checkbox" bind:checked={showPlates}>
			Fallas tectónicas
		</label>
		<label title="1,571 volcanes activos del Holoceno — Smithsonian GVP">
			<input type="checkbox" bind:checked={showVolcanoes}>
			Volcanes activos{volcanoAlertCount > 0 ? ` (${volcanoAlertCount})` : ''}
		</label>
		<label title="Mapa de calor de actividad sísmica acumulada">
			<input type="checkbox" bind:checked={showHeatmap}>
			Mapa de calor
		</label>
		<label title="Daños estimados por Sentinel-1 InSAR - Venezuela junio 2026 (58,870 estructuras · datos locales)">
			<input type="checkbox" bind:checked={showDamage}>
			Daños Sentinel-1
		</label>
		{#if damageLoading}
			<div class="layer-control-status">
				{damageIndexProgress > 0 && damageIndexProgress < 100 ? `Indexando ${damageIndexProgress}%...` : 'Cargando...'}
			</div>
		{/if}
		{#if damageError}
			<div class="layer-control-error">{damageError}</div>
		{/if}
	</div>
</section>
