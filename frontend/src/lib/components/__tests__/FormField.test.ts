// @vitest-environment jsdom
import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import FormField from '../FormField.svelte';

describe('FormField', () => {
	it('links the label to the input', () => {
		render(FormField, { label: 'Email', name: 'email' });
		expect(screen.getByLabelText('Email')).toHaveAttribute('name', 'email');
	});

	it('is valid with no error', () => {
		render(FormField, { label: 'Email', name: 'email' });
		expect(screen.getByLabelText('Email')).toHaveAttribute('aria-invalid', 'false');
	});

	it('shows a field error and ties it to the input', () => {
		render(FormField, { label: 'Email', name: 'email', error: 'Enter a valid email.' });
		const input = screen.getByLabelText('Email');
		expect(input).toHaveAttribute('aria-invalid', 'true');
		const message = screen.getByText('Enter a valid email.');
		expect(input.getAttribute('aria-describedby')).toContain(message.parentElement!.id);
	});

	it('shows every message when given several', () => {
		render(FormField, { label: 'Password', name: 'password', error: ['Too short', 'Too common'] });
		expect(screen.getByText('Too short')).toBeInTheDocument();
		expect(screen.getByText('Too common')).toBeInTheDocument();
	});

	it('shows a hint', () => {
		render(FormField, { label: 'Name', name: 'name', hint: 'As on your ID' });
		const input = screen.getByLabelText('Name');
		expect(screen.getByText('As on your ID').id).toBe(input.getAttribute('aria-describedby'));
	});

	it('updates its value as the user types', async () => {
		render(FormField, { label: 'Name', name: 'name' });
		const input = screen.getByLabelText<HTMLInputElement>('Name');
		await fireEvent.input(input, { target: { value: 'Ada' } });
		expect(input.value).toBe('Ada');
	});
});
