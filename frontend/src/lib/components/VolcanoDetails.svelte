<script lang="ts">
	import { Mountain } from '@lucide/svelte';
	import { findVolcanoAlert, volcanoAlertStyleMap } from '$lib/utils/volcanoUtils';
	import type { VolcanoAlert, VolcanoProperties } from '$lib/types';

	interface Props {
		selectedVolcano: VolcanoProperties;
		volcanoAlerts: VolcanoAlert[];
		onClose?: () => void;
	}

	let { selectedVolcano, volcanoAlerts, onClose }: Props = $props();

	let volcanoAlert = $derived(findVolcanoAlert(selectedVolcano.name || '', volcanoAlerts));
	let alertColor = $derived((volcanoAlertStyleMap[(volcanoAlert?.alert_level || '').toUpperCase()] || volcanoAlertStyleMap[(volcanoAlert?.color_code || '').toUpperCase()] || {}).color || '#ff8c00');
</script>

<h3 style="color: var(--text-primary); margin-top: 0; margin-bottom: 16px; font-size: 16px; text-align: center; font-weight: 800; display: flex; align-items: center; justify-content: center; gap: 8px;">
	<Mountain size="20" color="var(--accent)" />
	{selectedVolcano.name}
</h3>
<div style="display: flex; justify-content: space-between; align-items: center; background: var(--bg-input); backdrop-filter: blur(10px); -webkit-backdrop-filter: blur(10px); border: 1px solid var(--border-color); border-radius: 8px; padding: 12px 16px; margin-bottom: 12px; box-shadow: 0 4px 12px rgba(0,0,0,0.1);">
	<div style="flex: 1; text-align: center;">
		<div style="font-size: 11px; color: var(--text-secondary); text-transform: uppercase; font-weight: 600; letter-spacing: 0.5px; margin-bottom: 2px;">Elevación</div>
		<div style="font-size: 22px; font-weight: 800; color: var(--accent);">{selectedVolcano.elevation ? selectedVolcano.elevation.toLocaleString() : 'N/D'} <span style="font-size: 14px; font-weight: 800;">m</span></div>
	</div>
	<div style="width: 1px; height: 36px; background: var(--border-color);"></div>
	<div style="flex: 1; text-align: center;">
		<div style="font-size: 11px; color: var(--text-secondary); text-transform: uppercase; font-weight: 600; letter-spacing: 0.5px; margin-bottom: 2px;">Tipo</div>
		<div style="font-size: 15px; font-weight: 800; color: var(--text-primary);">{selectedVolcano.type || 'N/D'}</div>
	</div>
</div>

<div style="text-align: center; background: var(--bg-input); backdrop-filter: blur(10px); -webkit-backdrop-filter: blur(10px); border: 1px solid var(--border-color); border-radius: 8px; padding: 12px 16px; margin-bottom: 12px; box-shadow: 0 4px 12px rgba(0,0,0,0.1);">
	<div style="font-size: 11px; color: var(--text-secondary); text-transform: uppercase; font-weight: 600; letter-spacing: 0.5px; margin-bottom: 4px;">País / Región</div>
	<div style="font-size: 15px; font-weight: 800; color: var(--accent); letter-spacing: 0.5px;">{selectedVolcano.country || 'N/D'}{selectedVolcano.region ? ` — ${selectedVolcano.region}` : ''}</div>
</div>

<div style="text-align: center; background: var(--bg-input); backdrop-filter: blur(10px); -webkit-backdrop-filter: blur(10px); border: 1px solid var(--border-color); border-radius: 8px; padding: 12px 16px; margin-bottom: 24px; box-shadow: 0 4px 12px rgba(0,0,0,0.1);">
	<div style="font-size: 11px; color: var(--text-secondary); text-transform: uppercase; font-weight: 600; letter-spacing: 0.5px; margin-bottom: 4px;">Estado</div>
	<div style="font-size: 15px; font-weight: 800; color: var(--accent); letter-spacing: 0.5px;">{selectedVolcano.status || 'N/D'}</div>
</div>

{#if volcanoAlert}
	<div style="background: rgba(255,140,0,0.08); border: 1px solid rgba(255,140,0,0.3); border-radius: 8px; padding: 12px 16px; margin-bottom: 16px;">
		<div style="font-size: 11px; color: var(--text-secondary); text-transform: uppercase; font-weight: 600; letter-spacing: 0.5px; margin-bottom: 6px;">Alerta USGS</div>
		<div style="font-size: 18px; font-weight: 800; color: {alertColor};">{volcanoAlert.alert_level || volcanoAlert.color_code || 'N/D'}</div>
		{#if volcanoAlert.synopsis}
			<p style="font-size: 12px; color: #ccc; margin-top: 8px; line-height: 1.4;">{volcanoAlert.synopsis}</p>
		{/if}
	</div>
{/if}

<div style="display: flex; gap: 8px; margin-top: 16px; align-items: center;">
	<a href="https://volcano.si.edu/volcano.cfm?vn={selectedVolcano.name ? encodeURIComponent(selectedVolcano.name) : ''}" target="_blank" style="flex: 1; text-align: center; text-decoration: none; font-size: 13px; font-weight: 600; padding: 10px 16px; border-radius: 4px; display: flex; align-items: center; justify-content: center; gap: 6px; background: var(--accent); color: #000; letter-spacing: 0.5px; transition: opacity 0.2s;">
		<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path><polyline points="15 3 21 3 21 9"></polyline><line x1="10" y1="14" x2="21" y2="3"></line></svg>
		Ficha Smithsonian
	</a>
	<button onclick={onClose} style="text-decoration: none; padding: 10px 14px; border-radius: 4px; display: flex; align-items: center; justify-content: center; background: var(--bg-input); border: 1px solid var(--border-color); color: var(--text-primary); cursor: pointer; font-size: 13px; font-weight: 600;">
		Cerrar
	</button>
</div>
