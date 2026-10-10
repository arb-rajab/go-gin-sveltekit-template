// @vitest-environment jsdom
import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';
import Modal from '../Modal.svelte';
import { textSnippet } from './snippet';

describe('Modal', () => {
	it('renders nothing while closed', () => {
		render(Modal, { open: false, title: 'Confirm', onclose: vi.fn(), children: textSnippet('x') });
		expect(screen.queryByRole('dialog')).toBeNull();
	});

	it('renders an accessible dialog and focuses it when open', () => {
		render(Modal, {
			open: true,
			title: 'Confirm',
			onclose: vi.fn(),
			children: textSnippet('Are you sure?')
		});
		const dialog = screen.getByRole('dialog', { name: 'Confirm' });
		expect(dialog).toHaveAttribute('aria-modal', 'true');
		expect(screen.getByText('Are you sure?')).toBeInTheDocument();
		expect(dialog).toHaveFocus();
	});

	it('closes on Escape', async () => {
		const onclose = vi.fn();
		render(Modal, { open: true, title: 'Confirm', onclose, children: textSnippet('x') });
		await fireEvent.keyDown(screen.getByRole('dialog'), { key: 'Escape' });
		expect(onclose).toHaveBeenCalledOnce();
	});

	it('closes on a backdrop click but not on a click inside the dialog', async () => {
		const onclose = vi.fn();
		render(Modal, { open: true, title: 'Confirm', onclose, children: textSnippet('x') });
		const dialog = screen.getByRole('dialog');
		await fireEvent.click(dialog);
		expect(onclose).not.toHaveBeenCalled();
		await fireEvent.click(dialog.parentElement!);
		expect(onclose).toHaveBeenCalledOnce();
	});

	it('renders a footer', () => {
		render(Modal, {
			open: true,
			title: 'Confirm',
			onclose: vi.fn(),
			children: textSnippet('x'),
			footer: textSnippet('Footer actions')
		});
		expect(screen.getByText('Footer actions')).toBeInTheDocument();
	});
});
