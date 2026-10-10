// Optional module: the browser half of the live-events feature. A thin wrapper
// around EventSource pointed at this frontend's own /events proxy (never the
// backend directly; see src/routes/(optional-sse)/events/+server.ts).
//
// EventSource reconnects natively, but only at a fixed server-suggested delay.
// This wrapper takes over reconnection: on every error it tears down the
// current EventSource and schedules a fresh one after a capped exponential
// backoff, resetting to the base delay when a connection opens. A page that
// silently stopped updating after a network blip would be worse than none.

export interface LiveEvent {
	type: string;
	subject: string;
	occurred_at: string;
	data?: Record<string, unknown>;
}

export interface LiveEventsHandle {
	close: () => void;
}

const baseReconnectDelayMS = 1000;
const maxReconnectDelayMS = 30_000;

// Add the event names your backend publishes (livefeed.Event.Type).
const eventTypes: string[] = ['item_created', 'item_updated'];

// connectLiveEvents opens the stream and calls onEvent for every event it
// receives. onStatusChange reports 'connected' | 'reconnecting' so a caller can
// show an indicator. Call close() from the component's teardown.
export function connectLiveEvents(
	onEvent: (event: LiveEvent) => void,
	onStatusChange?: (status: 'connected' | 'reconnecting') => void
): LiveEventsHandle {
	let source: EventSource | null = null;
	let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
	let reconnectDelay = baseReconnectDelayMS;
	let closed = false;

	function handleMessage(message: MessageEvent<string>) {
		try {
			onEvent(JSON.parse(message.data) as LiveEvent);
		} catch {
			// A malformed frame must never tear down a healthy connection.
		}
	}

	function scheduleReconnect() {
		if (closed || reconnectTimer) return;
		onStatusChange?.('reconnecting');
		reconnectTimer = setTimeout(() => {
			reconnectTimer = null;
			reconnectDelay = Math.min(reconnectDelay * 2, maxReconnectDelayMS);
			open();
		}, reconnectDelay);
	}

	function open() {
		if (closed) return;
		source = new EventSource('/events');
		for (const type of eventTypes) {
			source.addEventListener(type, handleMessage);
		}
		source.onopen = () => {
			reconnectDelay = baseReconnectDelayMS;
			onStatusChange?.('connected');
		};
		source.onerror = () => {
			source?.close();
			source = null;
			scheduleReconnect();
		};
	}

	open();

	return {
		close() {
			closed = true;
			if (reconnectTimer) clearTimeout(reconnectTimer);
			source?.close();
			source = null;
		}
	};
}
