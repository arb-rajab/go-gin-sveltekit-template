import { describe, expect, it, vi } from 'vitest';
import {
	backendErrorMessage,
	backendFetch,
	parseSessionCookie,
	SESSION_COOKIE_NAME
} from './backend';

describe('parseSessionCookie', () => {
	it('extracts the token and Max-Age of the session cookie', () => {
		const header = `${SESSION_COOKIE_NAME}=abc.def; Path=/; Max-Age=3600; HttpOnly; Secure; SameSite=Strict`;
		expect(parseSessionCookie(['other=1; Path=/', header])).toEqual({
			token: 'abc.def',
			maxAge: 3600
		});
	});

	it('falls back to 24h when Max-Age is absent', () => {
		expect(parseSessionCookie([`${SESSION_COOKIE_NAME}=t; Path=/`])).toEqual({
			token: 't',
			maxAge: 86400
		});
	});

	it('returns null when no session cookie is present', () => {
		expect(parseSessionCookie(['other=1; Path=/'])).toBeNull();
		expect(parseSessionCookie([])).toBeNull();
	});
});

describe('backendFetch', () => {
	it('attaches the session token as a Cookie header', async () => {
		const fetchMock = vi.fn(async () => new Response('{}'));
		await backendFetch(fetchMock as unknown as typeof fetch, '/api/v1/auth/me', 'tok');

		const [url, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit];
		expect(url).toContain('/api/v1/auth/me');
		expect(new Headers(init.headers).get('Cookie')).toBe(`${SESSION_COOKIE_NAME}=tok`);
	});

	it('sends no Cookie header without a token', async () => {
		const fetchMock = vi.fn(async () => new Response('{}'));
		await backendFetch(fetchMock as unknown as typeof fetch, '/api/v1/auth/login', undefined);

		const [, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit];
		expect(new Headers(init.headers).has('Cookie')).toBe(false);
	});
});

describe('backendErrorMessage', () => {
	it('reads the error envelope message', async () => {
		const res = new Response(
			JSON.stringify({ error: { code: 'x', message: 'nope', field: null } })
		);
		expect(await backendErrorMessage(res, 'fallback')).toBe('nope');
	});

	it('falls back for a body that is not the envelope', async () => {
		expect(await backendErrorMessage(new Response('not json'), 'fallback')).toBe('fallback');
	});
});
