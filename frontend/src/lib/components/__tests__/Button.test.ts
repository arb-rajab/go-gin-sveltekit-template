// @vitest-environment jsdom
import { fireEvent, render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import Button from '../Button.svelte';
import { textSnippet } from './snippet';

describe('Button', () => {
	it('renders its label and defaults to type=button', () => {
		render(Button, { children: textSnippet('Save') });
		const button = screen.getByRole('button', { name: 'Save' });
		expect(button).toHaveAttribute('type', 'button');
		expect(button).toBeEnabled();
	});

	it('calls onclick', async () => {
		const onclick = vi.fn();
		render(Button, { children: textSnippet('Go'), onclick });
		await fireEvent.click(screen.getByRole('button'));
		expect(onclick).toHaveBeenCalledOnce();
	});

	it('is disabled and shows a spinner while loading', () => {
		render(Button, { children: textSnippet('Save'), loading: true });
		const button = screen.getByRole('button');
		expect(button).toBeDisabled();
		expect(button).toHaveAttribute('aria-busy', 'true');
		expect(screen.getByRole('status', { name: 'Working' })).toBeInTheDocument();
	});

	it('does not fire onclick when disabled', async () => {
		const onclick = vi.fn();
		render(Button, { children: textSnippet('Go'), disabled: true, onclick });
		// user-event honours `disabled` the way a real browser does; fireEvent does not.
		await userEvent.click(screen.getByRole('button'));
		expect(onclick).not.toHaveBeenCalled();
	});
});
