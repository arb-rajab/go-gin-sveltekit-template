<script lang="ts">
	import { pageWindow } from './pageWindow';

	type Props = {
		page: number;
		totalPages: number;
		/** Builds the link for a page, e.g. (p) => `/things?page=${p}`. */
		href: (page: number) => string;
	};

	let { page, totalPages, href }: Props = $props();

	const items = $derived(pageWindow(page, totalPages));
</script>

{#if totalPages > 1}
	<nav aria-label="Pagination">
		<ul>
			<li>
				{#if page > 1}
					<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- href() is built by the caller from resolve() -->
					<a href={href(page - 1)} rel="prev">Previous</a>
				{:else}
					<span aria-disabled="true">Previous</span>
				{/if}
			</li>
			{#each items as item, i (i)}
				<li>
					{#if item === '…'}
						<span aria-hidden="true">…</span>
					{:else if item === page}
						<span aria-current="page" class="current">{item}</span>
					{:else}
						<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- href() is built by the caller from resolve() -->
						<a href={href(item)}>{item}</a>
					{/if}
				</li>
			{/each}
			<li>
				{#if page < totalPages}
					<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- href() is built by the caller from resolve() -->
					<a href={href(page + 1)} rel="next">Next</a>
				{:else}
					<span aria-disabled="true">Next</span>
				{/if}
			</li>
		</ul>
	</nav>
{/if}

<style>
	ul {
		display: flex;
		gap: 0.25rem;
		list-style: none;
		padding: 0;
		margin: 1rem 0;
		flex-wrap: wrap;
	}
	a,
	span {
		display: inline-block;
		padding: 0.25rem 0.6rem;
		border: 1px solid var(--color-border);
		border-radius: 0.375rem;
		text-decoration: none;
	}
	span[aria-disabled='true'] {
		color: var(--color-muted);
	}
	span[aria-hidden='true'] {
		border-color: transparent;
	}
	.current {
		background: var(--color-primary);
		color: var(--color-on-primary);
		border-color: var(--color-primary);
	}
</style>
