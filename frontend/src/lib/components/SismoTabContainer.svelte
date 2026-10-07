<script lang="ts">
	import { ChevronDown, FileText, Newspaper } from '@lucide/svelte';

	interface Tab {
		id: string;
		label: string;
		component: any;
		props?: Record<string, unknown>;
	}

	interface Props {
		tabs: Tab[];
		active?: string;
		onChange?: (id: string) => void;
	}

	let { tabs, active = tabs[0]?.id ?? '', onChange }: Props = $props();

	const iconMap: Record<string, any> = {
		details: FileText,
		news: Newspaper
	};

	let collapsedSet = $state<Set<string>>(new Set());

	function toggle(id: string) {
		const next = new Set(collapsedSet);
		if (next.has(id)) next.delete(id);
		else next.add(id);
		collapsedSet = next;
		onChange?.(id);
	}
</script>

<div class="sismo-tab-container" style="display: flex; flex-direction: column; flex: 1; min-height: 0; overflow: hidden; gap: 0.4rem;">
	{#each tabs as tab (tab.id)}
		{@const Icon = iconMap[tab.id] || FileText}
		{@const isCollapsed = collapsedSet.has(tab.id) && active !== tab.id}
		<div class="collapsible {isCollapsed ? 'collapsed' : ''}" style={tab.id === active ? 'flex: 1; min-height: 0;' : ''}>
			<button class="collapsible-header" class:active={tab.id === active} aria-expanded={!isCollapsed} onclick={() => toggle(tab.id)}>
				<h4><Icon size={14} /> {tab.label}</h4>
				<span class="chevron"><ChevronDown size={16} /></span>
			</button>
			{#if tab.id === active}
				{@const ActiveComponent = tab.component}
				<div class="collapsible-content" style="flex: 1; min-height: 0; overflow-y: auto;">
					{#if ActiveComponent}
						<ActiveComponent {...(tab.props ?? {})} activeTab={active} />
					{:else}
						<div class="placeholder-text" style="text-align: center; padding: 24px;">
							Cargando contenido...
						</div>
					{/if}
				</div>
			{/if}
		</div>
	{/each}
</div>
