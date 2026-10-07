<script lang="ts">
	import { Mountain } from '@lucide/svelte';
	import { volcanoAlertStyleMap } from '$lib/utils/volcanoUtils';
	import type { VolcanoAlertBannerData } from '$lib/types';

	interface Props {
		banner: VolcanoAlertBannerData;
		onDismiss?: () => void;
	}

	let { banner, onDismiss }: Props = $props();
	let bannerColor = $derived((volcanoAlertStyleMap[(banner.level || '').toUpperCase()] || {}).color || '#ff8c00');
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="widget" style="grid-column: 1 / -1; background: {bannerColor}15; border: 1px solid {bannerColor}60; border-radius: 8px; padding: 10px 16px; display: flex; align-items: center; justify-content: space-between; gap: 12px;" onclick={() => { /* banner click no action */ }}>
	<div style="display: flex; align-items: center; gap: 10px;">
		<Mountain size="20" color={bannerColor} />
		<div>
			<div style="font-size: 13px; font-weight: 700; color: {bannerColor};">
				{banner.type === 'escalated' ? 'ALERTA VOLCÁNICA ESCALADA' : 'ALERTA VOLCÁNICA ACTIVA'}: {banner.level}
			</div>
			<div style="font-size: 12px; color: var(--text-primary);">
				{banner.name}{banner.previousLevel ? ` (subió de ${banner.previousLevel})` : ''}
				{banner.synopsis ? ` — ${banner.synopsis.substring(0, 90)}${banner.synopsis.length > 90 ? '…' : ''}` : ''}
			</div>
		</div>
	</div>
	<button onclick={onDismiss} style="background: transparent; border: none; color: var(--text-secondary); cursor: pointer; font-size: 18px; line-height: 1; padding: 4px;">&times;</button>
</div>
