<script lang="ts">
	import type { PageProps } from './$types';
	import Alert from '$lib/components/Alert.svelte';
	import Button from '$lib/components/Button.svelte';
	// Optional module: live updates over SSE. Delete the import, the
	// <LiveIndicator /> element, and src/lib/optional/ plus
	// src/routes/(optional-sse)/ to remove it.
	import LiveIndicator from '$lib/optional/LiveIndicator.svelte';

	let { data }: PageProps = $props();
</script>

<svelte:head>
	<title>Dashboard - starterapp</title>
</svelte:head>

<div class="heading">
	<h1>Dashboard</h1>
	<LiveIndicator onEvent={() => {}} />
	<form method="POST" action="?/logout">
		<Button type="submit" variant="secondary">Log out</Button>
	</form>
</div>

{#if data.loadError}
	<Alert>{data.loadError}</Alert>
{:else if data.me}
	<p>Signed in as <strong>{data.me.email}</strong>.</p>
	<p>Replace this page with your application.</p>
{/if}

<style>
	.heading {
		display: flex;
		align-items: center;
		gap: 1rem;
	}
	h1 {
		margin: 0;
		flex: 1;
	}
</style>
