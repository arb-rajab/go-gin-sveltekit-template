// @vitest-environment jsdom
import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import Alert from '../Alert.svelte';
import Spinner from '../Spinner.svelte';
import { textSnippet } from './snippet';

describe('Alert', () => {
	it('announces errors with role=alert', () => {
		render(Alert, { children: textSnippet('Bad things') });
		expect(screen.getByRole('alert')).toHaveTextContent('Bad things');
	});

	it('uses role=status for non-error variants', () => {
		render(Alert, { variant: 'success', children: textSnippet('Saved') });
		expect(screen.queryByRole('alert')).toBeNull();
		expect(screen.getByRole('status')).toHaveTextContent('Saved');
	});
});

describe('Spinner', () => {
	it('exposes an accessible label', () => {
		render(Spinner, { label: 'Loading users' });
		expect(screen.getByRole('status', { name: 'Loading users' })).toBeInTheDocument();
	});
});
