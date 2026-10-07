<script lang="ts">
	import type { NewsItem } from '$lib/types';

	interface Props {
		newsItems: NewsItem[];
		onSelectNews?: (news: NewsItem) => void;
	}

	let { newsItems, onSelectNews }: Props = $props();

	function handleClick(e: Event, news: NewsItem) {
		e.preventDefault();
		e.stopPropagation();
		onSelectNews?.(news);
	}
</script>

<div class="tab-content" id="tab-news" class:active={true}>
	{#if newsItems.length === 0}
		<div class="loading-spinner-container">
			<div class="loading-spinner"></div>
			Cargando noticias...
		</div>
	{:else}
		<div class="news-feed-container">
			{#each newsItems as news (news.link)}
				<!-- svelte-ignore a11y_click_events_have_key_events -->
				<!-- svelte-ignore a11y_no_static_element_interactions -->
				<div class="news-card" onclick={(e) => handleClick(e, news)}>
					<h4 class="news-title">{news.title}</h4>
					<div class="news-meta">
						<span class="news-source">{news.source}</span>
						<span class="news-date">{new Date(news.pubDate).toLocaleDateString()}</span>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
	#tab-news {
		display: flex;
		flex-direction: column;
		flex: 1;
		min-height: 0;
		overflow: hidden;
	}

	.news-feed-container {
		display: flex;
		flex-direction: column;
		flex: 1;
		min-height: 0;
		overflow-y: auto;
		padding-right: 8px;
		gap: 12px;
	}

	.news-card {
		background: var(--bg-input);
		border: 1px solid var(--border-color);
		padding: 12px;
		border-radius: 8px;
		cursor: pointer;
		transition: background 0.2s ease;
	}

	.news-card:hover {
		background: var(--bg-input);
	}

	.news-title {
		color: var(--text-primary);
		font-weight: 600;
		text-decoration: none;
		display: block;
		margin: 0 0 6px 0;
		font-size: 14px;
	}

	.news-meta {
		display: flex;
		justify-content: space-between;
		font-size: 11px;
		color: var(--text-secondary);
	}
</style>
