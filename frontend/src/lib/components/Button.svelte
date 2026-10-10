<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLButtonAttributes } from 'svelte/elements';
	import Spinner from './Spinner.svelte';

	type Props = Omit<HTMLButtonAttributes, 'children'> & {
		variant?: 'primary' | 'secondary' | 'danger';
		/** Shows a spinner and disables the button while an action is in flight. */
		loading?: boolean;
		children: Snippet;
	};

	let {
		variant = 'primary',
		loading = false,
		type = 'button',
		disabled = false,
		children,
		...rest
	}: Props = $props();
</script>

<button {type} class={variant} disabled={disabled || loading} aria-busy={loading} {...rest}>
	{#if loading}<Spinner label="Working" size="1rem" />{/if}
	{@render children()}
</button>

<style>
	button {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		font: inherit;
		border-radius: 0.375rem;
		border: 1px solid transparent;
		cursor: pointer;
	}
	button:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}
	.primary {
		background: var(--color-primary);
		color: var(--color-on-primary);
	}
	.secondary {
		background: transparent;
		color: var(--color-text);
		border-color: var(--color-border);
	}
	.danger {
		background: var(--color-danger);
		color: var(--color-on-primary);
	}
</style>
