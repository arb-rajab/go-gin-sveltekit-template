<script lang="ts">
	let { label = 'Loading', size = '1.25rem' }: { label?: string; size?: string } = $props();
</script>

<span class="spinner" role="status" aria-label={label} style:--size={size}></span>

<style>
	.spinner {
		display: inline-block;
		width: var(--size);
		height: var(--size);
		border: 2px solid var(--color-border);
		border-top-color: var(--color-primary);
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
		vertical-align: middle;
	}
	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.spinner {
			animation-duration: 3s;
		}
	}
</style>
