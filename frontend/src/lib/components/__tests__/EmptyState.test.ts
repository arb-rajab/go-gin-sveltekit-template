// @vitest-environment jsdom
import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import EmptyState from '../EmptyState.svelte';
import { textSnippet } from './snippet';

describe('EmptyState', () => {
	it('renders title and description', () => {
		render(EmptyState, { title: 'No items', description: 'Create one to get started.' });
		expect(screen.getByRole('heading', { name: 'No items' })).toBeInTheDocument();
		expect(screen.getByText('Create one to get started.')).toBeInTheDocument();
	});

	it('renders an action when given one', () => {
		render(EmptyState, { title: 'No items', children: textSnippet('Add item') });
		expect(screen.getByText('Add item')).toBeInTheDocument();
	});
});
