import type { RequestHandler } from './$types';
import { backendFetch, SESSION_COOKIE_NAME } from '$lib/server/backend';

// GET /events is the browser-facing half of the SSE proxy. The browser's
// EventSource only talks to this frontend's own origin; this route re-attaches
// the session cookie server-side, like every load function and form action,
// and streams the backend's GET /api/v1/events body through without buffering.
// It is a passthrough: the backend alone decides what an event is.
export const GET: RequestHandler = async ({ cookies, fetch, request }) => {
	const sessionToken = cookies.get(SESSION_COOKIE_NAME);
	if (!sessionToken) {
		return new Response('unauthorized', { status: 401 });
	}

	const backendRes = await backendFetch(fetch, '/api/v1/events', sessionToken, {
		signal: request.signal
	});
	if (!backendRes.ok || !backendRes.body) {
		return new Response('event stream unavailable', { status: backendRes.status || 502 });
	}

	return new Response(backendRes.body, {
		status: 200,
		headers: {
			'Content-Type': 'text/event-stream',
			'Cache-Control': 'no-cache',
			Connection: 'keep-alive',
			'X-Accel-Buffering': 'no'
		}
	});
};
