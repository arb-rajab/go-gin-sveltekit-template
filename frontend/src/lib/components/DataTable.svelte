<script lang="ts" generics="Row extends Record<string, unknown>">
	import EmptyState from './EmptyState.svelte';
	import Spinner from './Spinner.svelte';

	type Column = { key: keyof Row & string; header: string; align?: 'left' | 'right' };

	type Props = {
		columns: Column[];
		rows: Row[];
		/** Accessible table name, shown to assistive technology. */
		caption: string;
		/** Unique value per row, used as the keyed-each identity. */
		rowKey: (row: Row) => string | number;
		loading?: boolean;
		emptyTitle?: string;
		emptyDescription?: string;
	};

	let {
		columns,
		rows,
		caption,
		rowKey,
		loading = false,
		emptyTitle = 'Nothing here yet',
		emptyDescription
	}: Props = $props();
</script>

{#if loading}
	<p class="loading"><Spinner label="Loading {caption}" /></p>
{:else if rows.length === 0}
	<EmptyState title={emptyTitle} description={emptyDescription} />
{:else}
	<table>
		<caption class="visually-hidden">{caption}</caption>
		<thead>
			<tr>
				{#each columns as column (column.key)}
					<th scope="col" class={column.align}>{column.header}</th>
				{/each}
			</tr>
		</thead>
		<tbody>
			{#each rows as row (rowKey(row))}
				<tr>
					{#each columns as column (column.key)}
						<td class={column.align}>{row[column.key] ?? ''}</td>
					{/each}
				</tr>
			{/each}
		</tbody>
	</table>
{/if}

<style>
	table {
		width: 100%;
		border-collapse: collapse;
	}
	th,
	td {
		padding: 0.5rem 0.75rem;
		border-bottom: 1px solid var(--color-border);
		text-align: left;
	}
	.right {
		text-align: right;
	}
	th {
		font-size: 0.85rem;
		color: var(--color-muted);
	}
	.loading {
		text-align: center;
	}
	.visually-hidden {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip-path: inset(50%);
	}
</style>
