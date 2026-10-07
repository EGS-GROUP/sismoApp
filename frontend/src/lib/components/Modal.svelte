<script lang="ts">
	interface Props {
		data: any;
		canClose?: boolean;
		onClose?: () => void;
	}

	let { data, canClose = true, onClose }: Props = $props();
</script>

{#if data}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div class="modal-overlay show" onclick={(e) => { if (e.target === e.currentTarget && canClose) { onClose?.(); } }}>
		<div class="modal-content" onclick={(e) => e.stopPropagation()}>
			<div class="modal-header">
				<h3>{data.title}</h3>
				<button class="close-modal" onclick={(e) => { e.preventDefault(); e.stopPropagation(); if (canClose) { onClose?.(); } }}>&times;</button>
			</div>
			<div class="modal-body">
				{#if data.image}
					<img src={data.image} alt="Adjunto" class="modal-image">
				{/if}
				<div class="modal-body-text">
					{@html data.bodyHTML}
				</div>
			</div>
			{#if data.sourceLink}
				<div class="modal-footer">
					<a href={data.sourceLink} target="_blank" class="filter-btn active modal-source-link">{data.sourceText || 'Ver Fuente Original'}</a>
				</div>
			{/if}
		</div>
	</div>
{/if}

<style>
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: var(--bg-overlay);
		backdrop-filter: blur(8px);
		-webkit-backdrop-filter: blur(8px);
		z-index: 9999;
		display: flex;
		justify-content: center;
		align-items: center;
		padding: 20px;
		opacity: 0;
		pointer-events: none;
		transition: opacity 0.3s ease;
	}
	.modal-overlay.show {
		opacity: 1;
		pointer-events: auto;
	}

	.modal-content {
		background: var(--bg-darker);
		border: 1px solid var(--border-color);
		border-radius: 12px;
		width: 90vw;
		max-width: 600px;
		max-height: 90vh;
		display: flex;
		flex-direction: column;
		overflow: hidden;
		box-shadow: 0 10px 40px rgba(0, 0, 0, 0.8);
		animation: slideUp 0.3s ease;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 16px 24px;
		background: var(--bg-input);
		border-bottom: 1px solid var(--border-color);
	}
	.modal-header h3 {
		margin: 0;
		color: var(--text-primary);
		font-size: 16px;
	}

	.close-modal {
		background: var(--bg-input);
		border: none;
		color: var(--text-primary);
		width: 32px;
		height: 32px;
		border-radius: 50%;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 14px;
		transition: all 0.2s;
	}
	.close-modal:hover {
		background: #ff4444;
	}

	.modal-body {
		padding: 20px;
		overflow-y: auto;
		display: flex;
		flex-direction: column;
		gap: 16px;
		font-size: 14px;
		color: var(--text-primary);
	}

	.modal-image {
		width: 100%;
		max-height: 300px;
		object-fit: contain;
		border-radius: 8px;
		background: var(--bg-overlay);
	}

	.modal-footer {
		padding: 16px;
		border-top: 1px solid var(--border-color);
		text-align: center;
	}

	.modal-source-link {
		text-decoration: none;
		display: inline-block;
	}

	@keyframes slideUp {
		from { opacity: 0; transform: translateY(20px); }
		to { opacity: 1; transform: translateY(0); }
	}
</style>
