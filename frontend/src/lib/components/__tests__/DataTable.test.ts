// @vitest-environment jsdom
import { render, screen, within } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import DataTable from '../DataTable.svelte';

type Row = { id: number; name: string; role: string };
const columns = [
	{ key: 'name', header: 'Name' },
	{ key: 'role', header: 'Role' }
] as const;
const rowKey = (row: Row) => row.id;

describe('DataTable', () => {
	it('renders headers and rows', () => {
		render(DataTable<Row>, {
			columns: [...columns],
			rows: [
				{ id: 1, name: 'Ada', role: 'admin' },
				{ id: 2, name: 'Grace', role: 'user' }
			],
			caption: 'Users',
			rowKey
		});
		expect(screen.getByRole('table', { name: 'Users' })).toBeInTheDocument();
		expect(screen.getByRole('columnheader', { name: 'Name' })).toBeInTheDocument();
		const rows = screen.getAllByRole('row');
		expect(rows).toHaveLength(3); // header + 2 data rows
		expect(within(rows[1]).getByText('Ada')).toBeInTheDocument();
		expect(within(rows[2]).getByText('user')).toBeInTheDocument();
	});

	it('shows the empty state when there are no rows', () => {
		render(DataTable<Row>, {
			columns: [...columns],
			rows: [],
			caption: 'Users',
			rowKey,
			emptyTitle: 'No users'
		});
		expect(screen.queryByRole('table')).toBeNull();
		expect(screen.getByRole('heading', { name: 'No users' })).toBeInTheDocument();
	});

	it('shows a spinner while loading, even if rows exist', () => {
		render(DataTable<Row>, {
			columns: [...columns],
			rows: [{ id: 1, name: 'Ada', role: 'admin' }],
			caption: 'Users',
			rowKey,
			loading: true
		});
		expect(screen.queryByRole('table')).toBeNull();
		expect(screen.getByRole('status', { name: 'Loading Users' })).toBeInTheDocument();
	});
});
