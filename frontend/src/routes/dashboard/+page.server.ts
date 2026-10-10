import { redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { backendFetch, SESSION_COOKIE_NAME } from '$lib/server/backend';

// MeResponse mirrors operatorapi.loginResponse in the Go backend.
export interface MeResponse {
	operator_id: string;
	email: string;
}

// The gated read: the backend's 401, not cookie presence alone, is what
// actually decides whether this page may render. A missing or rejected session
// redirects to /login.
export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const sessionToken = cookies.get(SESSION_COOKIE_NAME);
	if (!sessionToken) {
		redirect(303, '/login');
	}

	const res = await backendFetch(fetch, '/api/v1/auth/me', sessionToken);
	if (res.status === 401) {
		cookies.delete(SESSION_COOKIE_NAME, { path: '/' });
		redirect(303, '/login');
	}
	if (!res.ok) {
		return { me: null as MeResponse | null, loadError: `Backend returned ${res.status}.` };
	}
	return { me: (await res.json()) as MeResponse, loadError: null as string | null };
};

export const actions: Actions = {
	logout: async ({ cookies, fetch }) => {
		const sessionToken = cookies.get(SESSION_COOKIE_NAME);
		await backendFetch(fetch, '/api/v1/auth/logout', sessionToken, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' }
		});
		cookies.delete(SESSION_COOKIE_NAME, { path: '/' });
		redirect(303, '/login');
	}
};
