<script lang="ts">
	import type { HTMLInputAttributes } from 'svelte/elements';

	type Props = Omit<HTMLInputAttributes, 'value'> & {
		label: string;
		name: string;
		value?: string;
		/** Field-level error(s), for example from the API's {error: {field, message}}. */
		error?: string | string[] | null;
		hint?: string;
	};

	let {
		label,
		name,
		value = $bindable(''),
		type = 'text',
		error = null,
		hint,
		id,
		...rest
	}: Props = $props();

	const fieldId = $derived(id ?? `field-${name}`);
	const messages = $derived(error ? (Array.isArray(error) ? error : [error]) : []);
	const hasError = $derived(messages.length > 0);
	const describedBy = $derived(
		[hasError ? `${fieldId}-error` : null, hint ? `${fieldId}-hint` : null]
			.filter(Boolean)
			.join(' ') || undefined
	);
</script>

<div class="field">
	<label for={fieldId}>{label}</label>
	<input
		id={fieldId}
		{name}
		{type}
		bind:value
		aria-invalid={hasError}
		aria-describedby={describedBy}
		{...rest}
	/>
	{#if hint}<small id="{fieldId}-hint" class="hint">{hint}</small>{/if}
	{#if hasError}
		<small id="{fieldId}-error" class="error">
			{#each messages as message (message)}<span>{message}</span>{/each}
		</small>
	{/if}
</div>

<style>
	.field {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}
	label {
		font-size: 0.9rem;
		font-weight: 500;
	}
	input {
		padding: 0.5rem;
		font: inherit;
		border: 1px solid var(--color-border);
		border-radius: 0.375rem;
		background: var(--color-bg);
		color: var(--color-text);
	}
	input[aria-invalid='true'] {
		border-color: var(--color-danger);
	}
	.hint {
		color: var(--color-muted);
	}
	.error {
		display: flex;
		flex-direction: column;
		color: var(--color-danger);
	}
</style>
