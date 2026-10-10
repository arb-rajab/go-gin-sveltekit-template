<script lang="ts">
	import type { Snippet } from 'svelte';

	type Props = {
		open: boolean;
		title: string;
		onclose: () => void;
		children: Snippet;
		footer?: Snippet;
	};

	let { open, title, onclose, children, footer }: Props = $props();

	let dialog = $state<HTMLDivElement>();
	const titleId = `modal-title-${Math.random().toString(36).slice(2, 10)}`;

	// Move focus into the dialog when it opens so keyboard and screen-reader
	// users land inside it.
	$effect(() => {
		if (open) dialog?.focus();
	});

	function onkeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			event.stopPropagation();
			onclose();
		}
	}
</script>

{#if open}
	<div class="backdrop" role="presentation" onclick={onclose}>
		<!-- stopPropagation keeps clicks inside the dialog from reaching the backdrop. -->
		<div
			bind:this={dialog}
			class="dialog"
			role="dialog"
			aria-modal="true"
			aria-labelledby={titleId}
			tabindex="-1"
			{onkeydown}
			onclick={(event) => event.stopPropagation()}
		>
			<h2 id={titleId}>{title}</h2>
			<div class="body">{@render children()}</div>
			{#if footer}<div class="footer">{@render footer()}</div>{/if}
		</div>
	</div>
{/if}

<style>
	.backdrop {
		position: fixed;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		background: rgb(0 0 0 / 0.5);
		z-index: 100;
	}
	.dialog {
		background: var(--color-bg);
		color: var(--color-text);
		border-radius: 0.5rem;
		padding: 1.5rem;
		width: min(32rem, calc(100vw - 2rem));
		max-height: calc(100vh - 2rem);
		overflow: auto;
	}
	h2 {
		margin-top: 0;
	}
	.footer {
		display: flex;
		justify-content: flex-end;
		gap: 0.5rem;
		margin-top: 1.5rem;
	}
</style>
