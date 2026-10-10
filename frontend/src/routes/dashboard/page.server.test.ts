import { describe, expect, it, vi } from 'vitest';
import { load } from './+page.server';
import { SESSION_COOKIE_NAME } from '$lib/server/backend';

function cookieJar(token: string | undefined) {
	const store = new Map<string, string>();
	if (token) store.set(SESSION_COOKIE_NAME, token);
	return { get: (name: string) => store.get(name), delete: vi.fn() };
}

describe('dashboard load', () => {
	it('redirects to /login without a session cookie, without calling the backend', async () => {
		const fetchMock = vi.fn();
		await expect(
			load({ cookies: cookieJar(undefined), fetch: fetchMock } as never)
		).rejects.toMatchObject({ status: 303, location: '/login' });
		expect(fetchMock).not.toHaveBeenCalled();
	});

	it('clears the cookie and redirects when the backend rejects the session', async () => {
		const cookies = cookieJar('stale');
		const fetchMock = vi.fn(async () => new Response('{}', { status: 401 }));
		await expect(load({ cookies, fetch: fetchMock } as never)).rejects.toMatchObject({
			status: 303,
			location: '/login'
		});
		expect(cookies.delete).toHaveBeenCalledWith(SESSION_COOKIE_NAME, { path: '/' });
	});

	it('returns the signed-in operator on success', async () => {
		const fetchMock = vi.fn(
			async () =>
				new Response(JSON.stringify({ operator_id: 'op-1', email: 'a@example.invalid' }), {
					status: 200
				})
		);
		const result = await load({ cookies: cookieJar('good'), fetch: fetchMock } as never);
		expect(result).toMatchObject({ me: { email: 'a@example.invalid' }, loadError: null });
	});
});
