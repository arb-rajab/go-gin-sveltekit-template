// @vitest-environment jsdom
import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import Pagination from '../Pagination.svelte';
import { pageWindow } from '../pageWindow';

const href = (p: number) => `/things?page=${p}`;

describe('pageWindow', () => {
	it('shows every page when there are few', () => {
		expect(pageWindow(2, 3)).toEqual([1, 2, 3]);
	});

	it('collapses the gaps on both sides of the current page', () => {
		expect(pageWindow(5, 10)).toEqual([1, '…', 4, 5, 6, '…', 10]);
	});

	it('collapses only the far side near an edge', () => {
		expect(pageWindow(1, 10)).toEqual([1, 2, '…', 10]);
		expect(pageWindow(10, 10)).toEqual([1, '…', 9, 10]);
	});

	it('does not insert a gap for a single hidden page', () => {
		expect(pageWindow(3, 5)).toEqual([1, 2, 3, 4, 5]);
	});

	it('clamps an out-of-range page and handles no pages', () => {
		expect(pageWindow(99, 3)).toEqual([1, 2, 3]);
		expect(pageWindow(1, 0)).toEqual([]);
	});
});

describe('Pagination', () => {
	it('renders nothing for a single page', () => {
		render(Pagination, { page: 1, totalPages: 1, href });
		expect(screen.queryByRole('navigation')).toBeNull();
	});

	it('marks the current page and links the others', () => {
		render(Pagination, { page: 2, totalPages: 3, href });
		expect(screen.getByText('2')).toHaveAttribute('aria-current', 'page');
		expect(screen.getByRole('link', { name: '3' })).toHaveAttribute('href', '/things?page=3');
		expect(screen.getByRole('link', { name: 'Previous' })).toHaveAttribute(
			'href',
			'/things?page=1'
		);
		expect(screen.getByRole('link', { name: 'Next' })).toHaveAttribute('href', '/things?page=3');
	});

	it('disables Previous on the first page and Next on the last', () => {
		const { unmount } = render(Pagination, { page: 1, totalPages: 3, href });
		expect(screen.queryByRole('link', { name: 'Previous' })).toBeNull();
		expect(screen.getByRole('link', { name: 'Next' })).toBeInTheDocument();
		unmount();

		render(Pagination, { page: 3, totalPages: 3, href });
		expect(screen.queryByRole('link', { name: 'Next' })).toBeNull();
		expect(screen.getByRole('link', { name: 'Previous' })).toBeInTheDocument();
	});
});
