import { describe, expect, it, vi } from 'vitest';
import { actions } from './+page.server';
import { SESSION_COOKIE_NAME } from '$lib/server/backend';

function makeEvent(form: Record<string, string>, fetchMock: ReturnType<typeof vi.fn>) {
	const body = new FormData();
	for (const [k, v] of Object.entries(form)) body.set(k, v);
	const set = vi.fn();
	return {
		event: {
			request: new Request('http://localhost/login', { method: 'POST', body }),
			cookies: { set },
			fetch: fetchMock
		},
		set
	};
}

describe('login action', () => {
	it('rejects an empty submission without calling the backend', async () => {
		const fetchMock = vi.fn();
		const { event } = makeEvent({ email: '', password: '' }, fetchMock);

		const result = await actions.default(event as never);

		expect(result).toMatchObject({ status: 400 });
		expect(result).toMatchObject({
			data: { fieldErrors: { email: 'Enter your email.', password: 'Enter your password.' } }
		});
		expect(fetchMock).not.toHaveBeenCalled();
	});

	it('surfaces the backend error message on a failed login', async () => {
		const fetchMock = vi.fn(
			async () =>
				new Response(
					JSON.stringify({
						error: { code: 'unauthorized', message: 'invalid email or password', field: null }
					}),
					{ status: 401 }
				)
		);
		const { event, set } = makeEvent({ email: 'a@example.invalid', password: 'x' }, fetchMock);

		const result = await actions.default(event as never);

		expect(result).toMatchObject({ status: 401, data: { error: 'invalid email or password' } });
		expect(set).not.toHaveBeenCalled();
	});

	it('sets a frontend-origin session cookie and redirects on success', async () => {
		const headers = new Headers();
		headers.append('Set-Cookie', `${SESSION_COOKIE_NAME}=tok123; Path=/; Max-Age=600; HttpOnly`);
		const fetchMock = vi.fn(async () => new Response('{}', { status: 200, headers }));
		const { event, set } = makeEvent({ email: 'a@example.invalid', password: 'x' }, fetchMock);

		await expect(actions.default(event as never)).rejects.toMatchObject({
			status: 303,
			location: '/dashboard'
		});
		expect(set).toHaveBeenCalledWith(
			SESSION_COOKIE_NAME,
			'tok123',
			expect.objectContaining({ httpOnly: true, secure: true, sameSite: 'strict', maxAge: 600 })
		);
	});
});
