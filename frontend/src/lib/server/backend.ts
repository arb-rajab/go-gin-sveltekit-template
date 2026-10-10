import { env } from '$env/dynamic/private';

// Server-side URL of the backend API. Every backend call is made from
// +page.server.ts load functions, form actions and +server.ts routes, never
// from code that runs in the browser, so this value never reaches the client.
const BACKEND_URL = env.BACKEND_URL ?? 'http://localhost:8080';

// Must equal operatorauth.SessionCookieName in the Go backend
// (backend/internal/operatorauth/session.go). This frontend never mints or
// verifies a session; it only carries the backend's opaque token back and forth.
export const SESSION_COOKIE_NAME = 'starterapp_session';

// backendFetch is the one place a request leaves this frontend process for the
// API. The backend enables no CORS policy by design, so a direct
// browser-to-backend fetch would be blocked anyway. The session cookie is
// attached as a plain Cookie header: the browser only ever holds a cookie
// issued by this frontend's own origin (see routes/login/+page.server.ts).
export async function backendFetch(
	fetchFn: typeof fetch,
	path: string,
	sessionToken: string | undefined,
	init: RequestInit = {}
): Promise<Response> {
	const headers = new Headers(init.headers);
	if (sessionToken) {
		headers.set('Cookie', `${SESSION_COOKIE_NAME}=${sessionToken}`);
	}
	return fetchFn(`${BACKEND_URL}${path}`, { ...init, headers });
}

export interface BackendErrorEnvelope {
	error: { code: string; message: string; field: string | null };
}

// backendErrorMessage extracts a human-readable message from the backend's
// JSON error envelope without ever throwing on a body that does not match it.
export async function backendErrorMessage(res: Response, fallback: string): Promise<string> {
	try {
		const body = (await res.json()) as BackendErrorEnvelope;
		return body?.error?.message ?? fallback;
	} catch {
		return fallback;
	}
}

// parseSessionCookie reads the session token and its Max-Age out of
// POST /auth/login's Set-Cookie headers. They are not forwarded to the browser
// verbatim: the backend issued them for its own origin, and the browser needs a
// cookie scoped to this frontend's origin, carrying the same name, value and
// lifetime.
export function parseSessionCookie(
	setCookieHeaders: string[]
): { token: string; maxAge: number } | null {
	for (const header of setCookieHeaders) {
		const [nameValue, ...attrs] = header.split(';').map((part) => part.trim());
		const eq = nameValue.indexOf('=');
		if (eq === -1) continue;
		const name = nameValue.slice(0, eq);
		if (name !== SESSION_COOKIE_NAME) continue;
		const token = nameValue.slice(eq + 1);

		// 24h matches operatorauth.SessionTTL; real responses always carry Max-Age.
		let maxAge = 60 * 60 * 24;
		for (const attr of attrs) {
			const attrEq = attr.indexOf('=');
			if (attrEq === -1) continue;
			const key = attr.slice(0, attrEq).toLowerCase();
			const value = attr.slice(attrEq + 1);
			if (key === 'max-age') {
				maxAge = Number(value);
			}
		}
		return { token, maxAge };
	}
	return null;
}
