<script lang="ts">
	import { onMount } from 'svelte';
	import { connectLiveEvents, type LiveEvent } from './liveEvents';

	let { onEvent }: { onEvent: (event: LiveEvent) => void } = $props();

	let status = $state<'connected' | 'reconnecting'>('reconnecting');

	onMount(() => {
		const handle = connectLiveEvents(onEvent, (next) => {
			status = next;
		});
		return handle.close;
	});
</script>

<span class="live" class:connected={status === 'connected'}>
	{status === 'connected' ? 'live' : 'reconnecting…'}
</span>

<style>
	.live {
		font-size: 0.8rem;
		color: #6e7781;
	}
	.live.connected {
		color: #1a7f37;
	}
</style>
