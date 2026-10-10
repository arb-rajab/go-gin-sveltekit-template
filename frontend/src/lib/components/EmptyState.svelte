<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		description,
		children
	}: { title: string; description?: string; children?: Snippet } = $props();
</script>

<div class="empty">
	<h2>{title}</h2>
	{#if description}<p>{description}</p>{/if}
	{#if children}<div class="action">{@render children()}</div>{/if}
</div>

<style>
	.empty {
		text-align: center;
		padding: 2rem 1rem;
		color: var(--color-muted);
	}
	h2 {
		margin: 0 0 0.5rem;
		color: var(--color-text);
		font-size: 1.1rem;
	}
	p {
		margin: 0;
	}
	.action {
		margin-top: 1rem;
	}
</style>
