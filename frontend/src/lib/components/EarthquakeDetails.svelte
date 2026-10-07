<script lang="ts">
	import { getMagnitudeColor } from '$lib/utils/mapUtils';
	import type { Earthquake } from '$lib/types';

	interface Props {
		selectedEarthquake: Earthquake;
		onClose?: () => void;
	}

	let { selectedEarthquake, onClose }: Props = $props();
</script>

<h3 style="color: var(--text-primary); margin-top: 0; margin-bottom: 16px; font-size: 16px; text-align: center; font-weight: 800;">{selectedEarthquake.location}</h3>
<div style="display: flex; justify-content: space-between; align-items: center; background: var(--bg-input); backdrop-filter: blur(10px); -webkit-backdrop-filter: blur(10px); border: 1px solid var(--border-color); border-radius: 8px; padding: 12px 16px; margin-bottom: 12px; box-shadow: 0 4px 12px rgba(0,0,0,0.1);">
	<div style="flex: 1; text-align: center;">
		<div style="font-size: 11px; color: var(--text-secondary); text-transform: uppercase; font-weight: 600; letter-spacing: 0.5px; margin-bottom: 2px;">Magnitud</div>
		<div style="font-size: 22px; font-weight: 800; color: {getMagnitudeColor(selectedEarthquake.magnitude)}">{selectedEarthquake.magnitude !== null && selectedEarthquake.magnitude !== undefined ? selectedEarthquake.magnitude.toFixed(1) : 'N/D'}</div>
	</div>
	<div style="width: 1px; height: 36px; background: var(--border-color);"></div>
	<div style="flex: 1; text-align: center;">
		<div style="font-size: 11px; color: var(--text-secondary); text-transform: uppercase; font-weight: 600; letter-spacing: 0.5px; margin-bottom: 2px;">Profundidad</div>
		<div style="font-size: 22px; font-weight: 800; color: var(--accent);">{String(selectedEarthquake.depth).replace(' km', '')} <span style="font-size: 14px; font-weight: 800;">km</span></div>
	</div>
</div>

<div style="text-align: center; background: var(--bg-input); backdrop-filter: blur(10px); -webkit-backdrop-filter: blur(10px); border: 1px solid var(--border-color); border-radius: 8px; padding: 12px 16px; margin-bottom: 24px; box-shadow: 0 4px 12px rgba(0,0,0,0.1);">
	<div style="font-size: 11px; color: var(--text-secondary); text-transform: uppercase; font-weight: 600; letter-spacing: 0.5px; margin-bottom: 4px;">Hora (Local)</div>
	<div style="font-size: 15px; font-weight: 800; color: var(--accent); letter-spacing: 0.5px;">{new Date(selectedEarthquake.time).toLocaleString('es-ES', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', hour12: true })}</div>
</div>

<div style="display: flex; justify-content: space-around; margin-bottom: 12px; text-align: center;">
	<div>
		<div style="font-size: 11px; color: var(--text-secondary); text-transform: uppercase; font-weight: 600; letter-spacing: 0.5px; margin-bottom: 6px;">CRs (DYFI)</div>
		<div>
			{#if selectedEarthquake.source === 'FUNVISIS'}
				<a href="https://encuesta.funvisis.gob.ve" target="_blank" style="color:var(--accent); text-decoration:none; font-weight:800; font-size: 12px; padding:4px 12px; border:1px solid var(--accent); border-radius:4px; display:inline-block;">Encuesta Disponible</a>
			{:else}
				<span style="font-size: 13px; font-weight: 800; color: var(--text-primary);">{(selectedEarthquake.felt ?? 0) > 0 ? `${selectedEarthquake.felt!.toLocaleString()} reportes` : 'Sin reportes'}</span>
			{/if}
		</div>
	</div>
	<div>
		<div style="font-size: 11px; color: var(--text-secondary); text-transform: uppercase; font-weight: 600; letter-spacing: 0.5px; margin-bottom: 6px;">Fuente Oficial</div>
		<div>
			<span style="display:inline-block; padding:4px 12px; border-radius:4px; font-size:12px; font-weight:800; letter-spacing:0.5px; color:var(--accent); border:1px solid color-mix(in srgb, var(--accent) 30%, transparent);">{selectedEarthquake.source.toUpperCase()}</span>
		</div>
	</div>
</div>

<div style="display: flex; gap: 8px; margin-top: 16px; align-items: center;">
	<a href="{selectedEarthquake.url}" target="_blank" style="flex: 1; text-align: center; text-decoration: none; font-size: 13px; font-weight: 600; padding: 10px 16px; border-radius: 4px; display: flex; align-items: center; justify-content: center; gap: 6px; background: var(--accent); color: #000; letter-spacing: 0.5px; transition: opacity 0.2s;">
		<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path><polyline points="15 3 21 3 21 9"></polyline><line x1="10" y1="14" x2="21" y2="3"></line></svg>
		{selectedEarthquake.source === 'FUNVISIS' ? 'Ver Boletín Oficial' : 'Reporte Oficial'}
	</a>
	<a href="https://twitter.com/intent/tweet?text={encodeURIComponent(`M${selectedEarthquake.magnitude !== null ? selectedEarthquake.magnitude.toFixed(1) : 'N/D'} - ${selectedEarthquake.location} - SismoMonitor`)}&url={encodeURIComponent('https://sismo.jagmedia.com.ve')}" target="_blank" style="text-decoration: none; padding: 10px; border-radius: 4px; display: flex; align-items: center; justify-content: center; background: var(--bg-input); border: 1px solid var(--border-color); transition: background 0.2s; color: var(--text-primary);" title="Compartir en X">
		<svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M18.901 1.153h3.68l-8.04 9.19L24 22.846h-7.406l-5.8-7.584-6.638 7.584H.474l8.6-9.83L0 1.154h7.594l5.243 6.932ZM17.61 20.644h2.039L6.486 3.24H4.298Z"/></svg>
	</a>
	<a href="https://wa.me/?text={encodeURIComponent(`M${selectedEarthquake.magnitude !== null ? selectedEarthquake.magnitude.toFixed(1) : 'N/D'} - ${selectedEarthquake.location} https://sismo.jagmedia.com.ve`)}" target="_blank" style="text-decoration: none; padding: 10px; border-radius: 4px; display: flex; align-items: center; justify-content: center; background: var(--bg-input); border: 1px solid var(--border-color); transition: background 0.2s; color: var(--text-primary);" title="Compartir en WhatsApp">
		<svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51a12.8 12.8 0 0 0-.57-.01c-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 0 1-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 0 1-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 0 1 2.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0 0 12.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 0 0 5.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 0 0-3.48-8.413z"/></svg>
	</a>
	{#if onClose}
		<button onclick={onClose} style="text-decoration: none; padding: 10px 14px; border-radius: 4px; display: flex; align-items: center; justify-content: center; background: var(--bg-input); border: 1px solid var(--border-color); color: var(--text-primary); cursor: pointer; font-size: 13px; font-weight: 600;">
			Cerrar
		</button>
	{/if}
</div>
