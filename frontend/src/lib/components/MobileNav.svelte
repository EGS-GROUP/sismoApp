<script lang="ts">
	import { List, MapPin, Info, BarChart3 } from '@lucide/svelte';

	let {
		mobileTab = 'feed',
		onTabChange
	}: {
		mobileTab?: string;
		onTabChange?: (tab: string) => void;
	} = $props();
</script>

<div class="mobile-nav">
	<button class="nav-btn {mobileTab === 'feed' ? 'active' : ''}" onclick={() => onTabChange?.('feed')}>
		<List size={20} />
		<span>Sismos</span>
	</button>
	<button class="nav-btn {mobileTab === 'map' ? 'active' : ''}" onclick={() => onTabChange?.('map')}>
		<MapPin size={20} />
		<span>Mapa</span>
	</button>
	<button class="nav-btn {mobileTab === 'info' ? 'active' : ''}" onclick={() => onTabChange?.('info')}>
		<Info size={20} />
		<span>Info</span>
	</button>
	<button class="nav-btn {mobileTab === 'stats' ? 'active' : ''}" onclick={() => onTabChange?.('stats')}>
		<BarChart3 size={20} />
		<span>Stats</span>
	</button>
</div>
