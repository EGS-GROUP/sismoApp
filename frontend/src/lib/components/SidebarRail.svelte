<script lang="ts">
	import { Activity, Layers, BarChart3, Trophy, Settings, Info } from '@lucide/svelte';

	let {
		sidebarPage = 'monitor',
		leftPanelOpen = true,
		onPageChange,
		onToggleLeft,
		onOpenInfo
	}: {
		sidebarPage?: string;
		leftPanelOpen?: boolean;
		onPageChange?: (page: string) => void;
		onToggleLeft?: () => void;
		onOpenInfo?: () => void;
	} = $props();

	const links = [
		{ id: 'monitor', icon: Activity, label: 'Monitor' },
		{ id: 'layers', icon: Layers, label: 'Capas' },
		{ id: 'analysis', icon: BarChart3, label: 'Análisis' },
		{ id: 'reports', icon: Trophy, label: 'Ranking' },
		{ id: 'settings', icon: Settings, label: 'Configuración' },
	];

	function handleNav(linkId: string) {
		if (sidebarPage === linkId) {
			onToggleLeft?.();
		} else {
			onPageChange?.(linkId);
			if (!leftPanelOpen) {
				onToggleLeft?.();
			}
		}
	}
</script>

<div class="sidebar-rail">
	{#each links as link (link.id)}
		<button
			class="rail-link {sidebarPage === link.id ? 'active' : ''}"
			onclick={() => handleNav(link.id)}
			title={link.label}
		>
			<link.icon size={20} />
		</button>
	{/each}

	<div class="rail-spacer"></div>

	<button class="rail-link" title="Información" onclick={onOpenInfo}>
		<Info size={20} />
	</button>
</div>
