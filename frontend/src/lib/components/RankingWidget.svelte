<script lang="ts">
	import { getMagnitudeColor } from '$lib/utils/mapUtils';
	import type { RankingData } from '$lib/types';

	interface Props {
		rankingData: RankingData;
		rankingMode: 'global' | 'venezuela';
		onModeChange?: (mode: 'global' | 'venezuela') => void;
	}

	let { rankingData, rankingMode, onModeChange }: Props = $props();
</script>

<div class="tab-content active" style="overflow-y: auto; flex: 1; min-height: 0; padding-right: 8px;">
	<div style="display: flex; justify-content: flex-end; align-items: center; margin-bottom: 16px;">
		<div style="display: flex; gap: 4px;">
			<button class="ranking-btn" class:active={rankingMode === 'global'} onclick={() => onModeChange?.('global')}>Global</button>
			<button class="ranking-btn" class:active={rankingMode === 'venezuela'} onclick={() => onModeChange?.('venezuela')}>Venezuela</button>
		</div>
	</div>
	
	{#if rankingData.length === 0}
		<div class="placeholder-text" style="text-align: center;">No hay sismos registrados en este rango de tiempo.</div>
	{:else}
		<table style="width: 100%; text-align: left; border-collapse: collapse; font-size: 13px;">
			<thead>
				<tr style="border-bottom: 1px solid var(--border-color); color: var(--text-secondary);">
					<th style="padding: 8px;">#</th>
					<th style="padding: 8px;">{rankingMode === 'global' ? 'Región / País' : 'Estado (VZLA)'}</th>
					<th style="padding: 8px; text-align: center;">Max Mag</th>
					<th style="padding: 8px; text-align: right;">Sismos</th>
				</tr>
			</thead>
			<tbody>
				{#each rankingData as [country, data], i (country)}
					<tr style="border-bottom: 1px solid rgba(255,255,255,0.05);">
						<td style="padding: 12px 8px; color: var(--accent); font-weight: bold;">{i + 1}</td>
						<td style="padding: 12px 8px;">{country}</td>
						<td style="padding: 12px 8px; text-align: center; color: {getMagnitudeColor(data.maxMag)}">{data.maxMag > 0 ? data.maxMag.toFixed(1) : 'N/D'}</td>
						<td style="padding: 12px 8px; text-align: right; color: var(--text-primary); font-weight: bold;">{data.count}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{/if}
</div>
