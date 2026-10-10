<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		variant = 'error',
		children
	}: { variant?: 'error' | 'success' | 'info'; children: Snippet } = $props();
</script>

<!-- Errors interrupt screen readers (role=alert); success and info do not. -->
<div class="alert {variant}" role={variant === 'error' ? 'alert' : 'status'}>
	{@render children()}
</div>

<style>
	.alert {
		padding: 0.75rem 1rem;
		border-radius: 0.375rem;
		border: 1px solid;
	}
	.error {
		color: var(--color-danger);
		border-color: var(--color-danger);
		background: var(--color-danger-bg);
	}
	.success {
		color: var(--color-success);
		border-color: var(--color-success);
		background: var(--color-success-bg);
	}
	.info {
		color: var(--color-text);
		border-color: var(--color-border);
		background: var(--color-surface);
	}
</style>
